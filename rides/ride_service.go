package rides

import (
	"context"
	"database/sql"
)

type Service interface {
	GetRide(ctx context.Context, tx *sql.Tx, rideID int) (*Ride, error)
	CreateRide(ctx context.Context, tx *sql.Tx, ride *Ride) (*Ride, error)
	UpdateRide(ctx context.Context, tx *sql.Tx, ride *Ride) (*Ride, error)
	DeleteRide(ctx context.Context, tx *sql.Tx, rideID int) (bool, error)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (rs *service) GetRide(ctx context.Context, tx *sql.Tx, rideID int) (*Ride, error) {
	stmt, err := tx.PrepareContext(ctx,
		`
	SELECT 
		id, created_at, idempotency_key_id, 
		origin_lat, origin_lon, 
		target_lat, target_lon, 
		stripe_charge_id, user_id
	FROM rocket_rides.public.rides
	WHERE id = $1
	;
	`,
	)
	if err != nil {
		return nil, err
	}

	var ride Ride
	err = stmt.QueryRowContext(ctx, rideID).Scan(
		&ride.ID, &ride.CreatedAt, &ride.IdempotencyKeyID,
		&ride.Origin.Lat, &ride.Origin.Long,
		&ride.Target.Lat, &ride.Target.Long,
		&ride.StripeChargeID, &ride.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &ride, nil
}

func (rs *service) CreateRide(ctx context.Context, tx *sql.Tx, ride *Ride) (*Ride, error) {

	stmt, err := tx.PrepareContext(ctx,
		`
	INSERT INTO rocket_rides.public.rides (
		idempotency_key_id, 
		origin_lat, origin_lon, 
		target_lat, target_lon, 
		stripe_charge_id, user_id
	) VALUES (
		$1, 
		$2, $3, 
		$4, $5,
		$6, $7
	)
	RETURNING 
	    id, created_at, idempotency_key_id, 
	    origin_lat, origin_lon, 
	    target_lat, target_lon, 
	    stripe_charge_id, user_id 
	; 
	`,
	)

	if err != nil {
		return nil, err
	}

	var newRide Ride
	err = stmt.QueryRowContext(ctx,
		&ride.IdempotencyKeyID,
		&ride.Origin.Lat, &ride.Origin.Long,
		&ride.Target.Lat, &ride.Target.Long,
		&ride.StripeChargeID, &ride.UserID,
	).Scan(
		&newRide.ID, &newRide.CreatedAt, &newRide.IdempotencyKeyID,
		&newRide.Origin.Lat, &newRide.Origin.Long,
		&newRide.Target.Lat, &newRide.Target.Long,
		&newRide.StripeChargeID, &newRide.UserID,
	)
	if err != nil {
		return nil, err
	}

	return &newRide, nil
}
func (rs *service) UpdateRide(ctx context.Context, tx *sql.Tx, ride *Ride) (*Ride, error) {
	stmt, err := tx.PrepareContext(ctx,
		`
	UPDATE rocket_rides.public.rides
	SET 
	    idempotency_key_id = $2,
		origin_lat = $3,
 		origin_lon = $4,
		target_lat = $5,
		target_lon = $6,
		stripe_charge_id = $7,
		user_id = $8
	WHERE id = $1
	RETURNING 
	    id, created_at, idempotency_key_id, 
	    origin_lat, origin_lon, 
	    target_lat, target_lon, 
	    stripe_charge_id, user_id 
	`,
	)

	if err != nil {
		return nil, err
	}

	var updatedRide Ride
	err = stmt.QueryRowContext(ctx,
		ride.ID,                           // $1
		ride.IdempotencyKeyID,             // $2
		ride.Origin.Lat, ride.Origin.Long, // $3, $4
		ride.Target.Lat, ride.Target.Long, // $5, $6
		ride.StripeChargeID, // $7
		ride.UserID,         // $8
	).Scan(
		&updatedRide.ID,
		&updatedRide.CreatedAt,
		&updatedRide.IdempotencyKeyID,
		&updatedRide.Origin.Lat,
		&updatedRide.Origin.Long,
		&updatedRide.Target.Lat,
		&updatedRide.Target.Long,
		&updatedRide.StripeChargeID,
		&updatedRide.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &updatedRide, nil
}
func (rs *service) DeleteRide(ctx context.Context, tx *sql.Tx, rideID int) (bool, error) {

	stmt, err := tx.PrepareContext(ctx,
		`
	DELETE FROM rocket_rides.public.rides
	WHERE id = $1
	`)

	if err != nil {
		return false, err
	}

	result, err := stmt.Exec(rideID)
	if err != nil {
		return false, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
