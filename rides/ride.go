package rides

import (
	"database/sql"
	"time"
)

type Coordinate struct {
	Lat  float64
	Long float64
}

func (c Coordinate) IsValid() bool {
	return (c.Lat >= -90.0 && c.Lat <= 90.0) && (c.Long >= -180.0 && c.Long <= 180.0)
}

type Ride struct {
	ID        int
	CreatedAt time.Time
	// Store a reference to the idempotency key so that we can recover an
	// already-created ride. Not that idempotency keys are not stored
	// permanently, so make sure to SET NULL when a referenced key is being reaped.
	IdempotencyKeyID sql.Null[int]
	// origin and destination latitudes and longitudes
	Origin Coordinate
	Target Coordinate
	// ID of Stripe charge like ch_123; NULL until we have one
	StripeChargeID sql.Null[string]
	UserID         int
}

func New() *Ride {
	return &Ride{}
}
