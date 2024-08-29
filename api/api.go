package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	audit "github.com/anmho/idempotent-rides/audit"
	"github.com/anmho/idempotent-rides/idempotency"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/scope"
	"github.com/anmho/idempotent-rides/send"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/paymentintent"
	"log/slog"
	"net/http"
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
	if errors.As(err, &send.HTTPError{}) {
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

type RegisterUserParams struct {
	Email string
}

func handleRegisterUser(db *sql.DB, userService users.Service) RouteHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		params, err := send.Read[RegisterUserParams](r.Body)
		if err != nil {
			return send.HTTPError{
				Cause:   err,
				Message: "bad request - invalid params for register user",
				Status:  http.StatusBadRequest,
			}
		}

		customerParams := &stripe.CustomerParams{
			Name:  stripe.String("test customer"),
			Email: stripe.String(params.Email),
		}

		result, err := customer.New(customerParams)
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return err
		}
		defer func(tx *sql.Tx) {
			err := tx.Rollback()
			scope.GetLogger().Error("error transaction", slog.Any("err", err))
		}(tx)

		user := users.New(params.Email, result.ID)
		createdUser, err := userService.CreateUser(ctx, tx, user)
		if err != nil {
			return err
		}

		user = createdUser

		scope.GetLogger().Info(
			"creating stripe customer",
			slog.Any("stripeCustomerID", user.StripeCustomerID),
		)

		err = tx.Commit()
		if err != nil {
			return err
		}

		return send.WriteJSON[*users.User](w, http.StatusCreated, user)
	}
}

type RideReservationParams struct {
	UserID *int              `json:"user_id"`
	Origin *rides.Coordinate `json:"origin"`
	Target *rides.Coordinate `json:"target"`
}

type RideReservationResponse struct {
	RideID int `json:"ride_id"`
}

func validateReservationParams(params RideReservationParams) error {
	if params.UserID == nil {
		return errors.New("must provide valid userID")
	}

	if params.Origin == nil || !params.Origin.IsValid() {
		return errors.New("must provide valid origin")
	}

	if params.Target == nil || !params.Origin.IsValid() {
		return errors.New("must provide valid target")
	}
	return nil
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
		params, err := send.Read[RideReservationParams](r.Body)
		if err != nil {
			return send.HTTPError{
				Cause:   err,
				Message: "bad request",
				Status:  http.StatusBadRequest,
			}
		}

		if err = validateReservationParams(params); err != nil {
			scope.GetLogger().Error("invalid request params")
			return send.HTTPError{
				Cause:   err,
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
				if err != nil {
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
				}

				result := idempotency.NewRecoveryPointResult(idempotency.StartedRecoveryPoint)
				return result, nil
			},
		)
		if err != nil {
			return fmt.Errorf("failed to upsert idempotency key: %w", err)
		}

		var ride *rides.Ride
		var updatedKey *idempotency.Key
		user, err := userService.GetUser(ctx, db, userID)
		if err != nil {
			return err
		}

		// Once we have the key, we'll continue the work and verify it is not completed.
		for {
			if key == nil {
				return errors.New("nil key when executing phases")
			}
			switch key.RecoveryPoint {
			case idempotency.StartedRecoveryPoint:
				scope.GetLogger().Info("StartedRecoveryPoint")
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 2: ride_created
						//	Create ride
						origin := *params.Origin
						target := *params.Target
						ride, err = rides.New(key.ID, origin, target, userID)
						if err != nil {
							return nil, send.HTTPError{
								Cause:   nil,
								Message: "bad request for ride",
								Status:  http.StatusBadRequest,
							}
						}

						ride, err = rideService.CreateRide(ctx, tx, ride)
						if err != nil {
							return nil, err
						}

						// Create ride audit record
						return idempotency.NewRecoveryPointResult(idempotency.RideCreatedRecoveryPoint), nil
					},
				)
			case idempotency.RideCreatedRecoveryPoint:
				scope.GetLogger().Info("RideCreatedRecoveryPoint")
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 3:
						//	Charge user via Stripe
						//	Create ride payment charged audit record

						// 	Charge user via Stripe
						paymentParams := &stripe.PaymentIntentParams{
							Amount:       stripe.Int64(500),
							Currency:     stripe.String(string(stripe.CurrencyUSD)),
							Customer:     stripe.String(user.StripeCustomerID),
							ReceiptEmail: stripe.String(user.Email),
						}
						paymentIntent, err := paymentintent.New(paymentParams)
						if err != nil {
							return nil, err
						}

						ride.StripeChargeID = sql.Null[string]{
							V:     paymentIntent.ID,
							Valid: true,
						}
						//	Update ride
						updatedRide, err := rideService.UpdateRide(ctx, tx, ride)
						if err != nil {
							return nil, err
						}
						ride = updatedRide
						return idempotency.NewRecoveryPointResult(idempotency.ChargeCreatedRecoveryPoint), nil
					},
				)
			case idempotency.ChargeCreatedRecoveryPoint:
				scope.GetLogger().Info("ChargeCreatedRecoveryPoint")
				updatedKey, err = idempotency.AtomicPhase(ctx, key, db,
					func(tx *sql.Tx) (idempotency.AtomicPhaseResult, error) {
						// Checkpoint 4:
						//	Stage send receipt job
						// need to get the ride id
						return idempotency.NewResponseResult(http.StatusCreated, map[string]any{"ride_id": ride.ID}), nil
					},
				)
			case idempotency.FinishedRecoveryPoint:
				scope.GetLogger().Info("FinishedRecoveryPoint")
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
		var response RideReservationResponse
		err = json.Unmarshal(key.ResponseBody.V, &response)
		if err != nil {
			return err
		}
		return send.WriteJSON(w, key.ResponseCode.V, response)
	}
}
