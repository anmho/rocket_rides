package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anmho/idempotent-rides/audit"
	"github.com/anmho/idempotent-rides/idempotency"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/scope"
	"github.com/anmho/idempotent-rides/send"
	"github.com/anmho/idempotent-rides/users"
	"log"
	"log/slog"
	"net/http"
	"time"
)

func MakeServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	rideService := rides.MakeService()
	auditService := audit.MakeService()
	userService := users.MakeService()

	// register middlewares
	registerRoutes(mux, db, rideService, auditService, userService)

	return mux
}

func handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, &send.HTTPError{}) {
		send.Error(w, err.(send.HTTPError))
	} else {
		send.Error(w, send.NewErrInternal(err))
	}
}

type RouteHandler = func(w http.ResponseWriter, r *http.Request) error

func MakeHandlerFunc(f RouteHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			handleError(w, err)
		}
	}
}

const (
	MinIdempotencyKeyLength = 2
	IdempotencyKeyLockTimeout
)

func validateIdempotencyKey(key string) bool {
	return len(key) >= MinIdempotencyKeyLength
}

type RideReservationParams struct {
	UserID *int
	Origin rides.Coordinate
	Target rides.Coordinate
}

func validateReservationParams(params RideReservationParams) bool {
	return params.UserID != nil && params.Origin.IsValid() && params.Target.IsValid()
}

func handleRideReservation(db *sql.DB, rideService rides.Service, auditService audit.Service, userService users.Service) RouteHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		keyVal := r.Header.Get(idempotency.HeaderKey)
		scope.GetLogger().Info("handleRideReservation", slog.String("keyVal", keyVal))
		if !validateIdempotencyKey(keyVal) {
			return send.HTTPError{
				Message: "idempotency key required",
				Status:  http.StatusBadRequest,
			}
		}
		params, err := send.Read[RideReservationParams](r)
		if err != nil {
			return send.HTTPError{
				Cause:   err,
				Message: "bad request",
				Status:  http.StatusBadRequest,
			}
		}

		if !validateReservationParams(params) {
			return send.HTTPError{
				Message: "invalid request params",
				Status:  http.StatusBadRequest,
			}
		}

		// if there's an idempotency key we should retrieve it and check the status.
		// Each atomic phase will be wrapped in a transaction.

		userID := *params.UserID
		var key *idempotency.Key
		// Checkpoint 1: Started
		_, err = idempotency.AtomicPhase(ctx, key, db,
			func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
				//	Create or get key if they supplied the header
				key, err = idempotency.FindKey(ctx, tx, userID, keyVal)
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf("error finding key: %w", err)
				}

				// need to marshal into binary
				bytes, err := json.Marshal(params)
				if err != nil {
					return nil, fmt.Errorf("marshaling params: %w", err)
				}

				var newKey *idempotency.Key
				newKey, err = idempotency.InsertKey(ctx, tx, idempotency.KeyParams{
					Key:           keyVal,
					RequestMethod: idempotency.RequestMethod(r.Method),
					RequestParams: bytes,
					RequestPath:   r.URL.Path,
					UserID:        userID,
				})

				if err != nil {
					return nil, fmt.Errorf("failed to add new key: %w", err)
				}
				key = newKey

				result := idempotency.NewRecoveryPointResult(idempotency.StartedRecoveryPoint)
				return result, nil
			},
		)
		if err != nil {
			return fmt.Errorf("failed to upsert idempotency key: %w", err)
		}

		var ride *rides.Ride
		var updatedKey *idempotency.Key

		// Once we have the key, we'll continue the work and verify it is not completed.
		for {
			if key == nil {
				return errors.New("nil key when executing phases")
			}
			scope.GetLogger().Error("hello", slog.Any("key stage", key))
			switch key.RecoveryPoint {
			case idempotency.StartedRecoveryPoint:
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 2: ride_created
						//	Create ride
						ride, err = rides.New(key.ID, params.Origin, params.Origin)
						if err != nil {
							return nil, send.HTTPError{
								Cause:   nil,
								Message: "bad request for ride",
								Status:  http.StatusBadRequest,
							}
						}

						//	Create ride audit record
						log.Println("StartedRecoveryPoint")

						data, err := json.Marshal(params)
						if err != nil {
							return nil, send.HTTPError{
								Cause:   fmt.Errorf("marshalling params: %w - %+v", err, params),
								Message: "marshaling params",
								Status:  http.StatusInternalServerError,
							}
						}

						record := audit.NewRecord("create_ride", data, r.RemoteAddr, audit.Resource{
							ID:   ride.ID,
							Type: "ride",
						}, *params.UserID)
						record, err = auditService.CreateRecord(ctx, tx, record)
						if err != nil {
							return nil, err
						}

						return idempotency.NewRecoveryPointResult(idempotency.RideCreatedRecoveryPoint), nil
					},
				)
			case idempotency.RideCreatedRecoveryPoint:
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 3:
						//	Create ride audit record
						bytes, err := json.Marshal(params)
						if err != nil {
							return nil, fmt.Errorf("creating audit record: %w", err)
						}
						record, err := auditService.CreateRecord(ctx, tx, &audit.Record{
							Action:    "created",
							CreatedAt: time.Now(),
							Data:      bytes,
							OriginIP:  r.RemoteAddr,
							Resource: audit.Resource{
								ID:   ride.ID,
								Type: "ride",
							},
							UserID: userID,
						})

						if err != nil {
							return nil, err
						}
						scope.GetLogger().Info(
							"record created",
							slog.Any("record", record))

						// 	Charge user via Stripe
						//	Update ride
						scope.GetLogger().Info("RideCreatedRecoveryPoint")
						return idempotency.NewRecoveryPointResult(idempotency.ChargeCreatedRecoveryPoint), nil
					},
				)
			case idempotency.ChargeCreatedRecoveryPoint:
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 4: Charge user via Stripe
						//	Stage send receipt job
						scope.GetLogger().Info("ChargeCreatedRecoveryPoint")
						// need to get the ride id
						return idempotency.NewResponseResult(http.StatusCreated, map[string]any{"ride_id": "hello"}), nil
					},
				)
			case idempotency.FinishedRecoveryPoint:
				goto loop
			default:
				return errors.New("unknown recovery point" + key.RecoveryPoint.String())
			}

			// should switch right here
			if err != nil {
				return err
			}
			key = updatedKey
		}
	loop:
		send.WriteJSON(w, key.ResponseCode.V, key.ResponseBody.V)
		return nil
	}
}
