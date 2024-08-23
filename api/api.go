package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anmho/idempotent-rides/idempotency"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/scope"
	"github.com/anmho/idempotent-rides/send"
	"log"
	"log/slog"
	"net/http"
)

func NewServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, db)

	// register middlewares

	return mux
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

func HandleRideReservation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		keyVal := r.Header.Get(idempotency.HeaderKey)
		scope.GetLogger().Info("HandleRideReservation", slog.String("keyVal", keyVal))
		if !validateIdempotencyKey(keyVal) {
			send.Error(w, send.HTTPError{
				Message: "idempotency key required",
				Status:  http.StatusBadRequest,
			})
			return
		}
		params, err := send.Read[RideReservationParams](r)
		if err != nil {
			send.Error(w, send.HTTPError{
				Cause:   err,
				Message: "bad request",
				Status:  http.StatusBadRequest,
			})
			return
		}
		// TODO: add actual validation logic
		if !validateReservationParams(params) {
			send.Error(w, send.HTTPError{
				Message: "invalid request params",
				Status:  http.StatusBadRequest,
			})
			return
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
				scope.GetLogger().Error("1 key should not be nil at this point", slog.Any("key", key), slog.Any("newKey", newKey))
				key = newKey

				result := idempotency.NewRecoveryPointResult(idempotency.StartedRecoveryPoint)
				return result, nil
			},
		)
		if err != nil {
			send.Error(w, send.NewErrInternal(fmt.Errorf("failed to upsert idempotency key: %w", err)))
			return
		}

		//key = updatedKey
		scope.GetLogger().Error("2 key should not be nil at this point", slog.Any("key", key))

		var ride rides.Ride
		_ = ride
		var updatedKey *idempotency.Key
		_ = updatedKey

		// Once we have the key, we'll continue the work and verify it is not completed.

		for {
			scope.GetLogger().Error("hello", slog.Any("key stage", key))
			switch key.RecoveryPoint {
			case idempotency.StartedRecoveryPoint:
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 2: ride_created
						//	Create ride
						//	Create ride audit record
						log.Println("StartedRecoveryPoint")

						return idempotency.NewRecoveryPointResult(idempotency.RideCreatedRecoveryPoint), nil
					},
				)
			case idempotency.RideCreatedRecoveryPoint:
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 3: Create ride audit record
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
				send.Error(w, send.NewErrInternal(errors.New("unknown recovery point"+key.RecoveryPoint.String())))
				return

			}

			//scope.GetLogger().Error("rest of phases", slog.Any("updatedKey", updatedKey), slog.Any("error", err))
			if err != nil {
				send.Error(w, send.NewErrInternal(err))
				return
			}
			key = updatedKey
		}
	loop:

		scope.GetLogger().Error("end loop", slog.Any("key", key))
		w.WriteHeader(key.ResponseCode.V)
		w.Write(key.ResponseBody.V)
	}
}
