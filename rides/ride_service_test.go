package rides

import (
	"context"
	"database/sql"
	"github.com/anmho/idempotent-rides/testfixtures"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	TestRide = &Ride{
		ID: 1337,
		IdempotencyKeyID: sql.Null[int]{
			V:     738,
			Valid: true,
		},
		Origin: Coordinate{
			Lat:  1,
			Long: 2,
		},
		Target: Coordinate{
			Lat:  3,
			Long: 4,
		},
		StripeChargeID: sql.Null[string]{},
		UserID:         *users.TestUserID,
	}
)

func AssertEqualRide(t *testing.T, expected, ride *Ride) {
	assert.Equal(t, expected.ID, ride.ID)
	assert.Equal(t, expected.UserID, ride.UserID)
	assert.Equal(t, expected.StripeChargeID, ride.StripeChargeID)
	assert.Equal(t, expected.Origin, ride.Origin)
	assert.Equal(t, expected.Target, ride.Target)
}

func TestRideService_GetRide(t *testing.T) {
	tests := []struct {
		desc   string
		rideID int

		expectedRide *Ride
		expectedErr  bool
	}{
		{
			desc:   "happy path: get a ride that exists in the database",
			rideID: TestRide.ID,

			expectedErr:  false,
			expectedRide: TestRide,
		},
		{
			desc:   "error path: get a ride that doesn't exist int the database",
			rideID: 7258,

			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			db, cleanup := testfixtures.MakePostgres(t)
			t.Cleanup(func() {
				cleanup()
			})
			tx, err := db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
				ReadOnly:  false,
			})
			if err != nil {
				return
			}
			rideService := NewService()
			ride, err := rideService.GetRide(ctx, tx, tc.rideID)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ride)
				AssertEqualRide(t, tc.expectedRide, ride)
			}
		})
	}
}

func TestRideService_CreateRide(t *testing.T) {

}

func TestRideService_UpdateRide(t *testing.T) {

}

func TestRideService_DeleteRide(t *testing.T) {

}
