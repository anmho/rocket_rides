package rides

import (
	"context"
	"database/sql"
	"github.com/anmho/idempotent-rides/testfixtures"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	TestExistingRide = &Ride{
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
		UserID:         *users.TestUser1ID,
	}
	TestNewRide = &Ride{
		ID:               1442,
		IdempotencyKeyID: sql.Null[int]{},
		Origin: Coordinate{
			Lat:  72,
			Long: 72,
		},
		Target: Coordinate{
			Lat:  72,
			Long: 72,
		},
		StripeChargeID: sql.Null[string]{},
		UserID:         *users.TestUser2ID,
	}
)

func AssertEqualRide(t *testing.T, expected, ride *Ride) {
	assert.Equal(t, expected.UserID, ride.UserID)
	assert.Equal(t, expected.StripeChargeID, ride.StripeChargeID)
	assert.Equal(t, expected.Origin, ride.Origin)
	assert.Equal(t, expected.Target, ride.Target)
}

func TestRideService_GetRide(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc   string
		rideID int

		expectedRide *Ride
		expectedErr  bool
	}{
		{
			desc:   "happy path: get a ride that exists in the database",
			rideID: TestExistingRide.ID,

			expectedErr:  false,
			expectedRide: TestExistingRide,
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
			rideService := MakeService()
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
	t.Parallel()
	tests := []struct {
		desc string
		ride *Ride

		expectedErr  bool
		expectedRide *Ride
	}{
		{
			desc:         "happy path: create new ride",
			ride:         TestNewRide,
			expectedErr:  false,
			expectedRide: TestNewRide,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)

			t.Cleanup(func() {
				cleanup()
			})

			ctx := context.Background()
			tx, err := db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
				ReadOnly:  false,
			})
			require.NoError(t, err)
			require.NotNil(t, tx)

			rideService := MakeService()

			newRide, err := rideService.CreateRide(ctx, tx, tc.ride)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, newRide)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newRide)
				AssertEqualRide(t, tc.expectedRide, newRide)
			}
		})
	}
}

func TestRideService_UpdateRide(t *testing.T) {
	t.Parallel()
	updatedRide := &Ride{
		ID:               TestExistingRide.ID,
		IdempotencyKeyID: TestExistingRide.IdempotencyKeyID,
		Origin:           TestExistingRide.Origin,
		Target: Coordinate{
			Lat:  TestExistingRide.Target.Long + 10.0,
			Long: TestExistingRide.Target.Long + 10.0,
		},
		StripeChargeID: sql.Null[string]{},
		UserID:         TestExistingRide.UserID,
	}
	tests := []struct {
		desc string
		ride *Ride

		expectedErr  bool
		expectedRide *Ride
	}{
		{
			desc: "happy path: update an existing ride. change the target",
			ride: updatedRide,

			expectedRide: updatedRide,
		},
		{
			desc: "error path: update a non-existent ride. ride id does not exist",
			ride: &Ride{
				ID:               5823,
				IdempotencyKeyID: TestExistingRide.IdempotencyKeyID,
				Origin:           TestExistingRide.Origin,
				Target: Coordinate{
					Lat:  TestExistingRide.Target.Long + 10.0,
					Long: TestExistingRide.Target.Long + 10.0,
				},
				StripeChargeID: sql.Null[string]{},
				UserID:         TestExistingRide.UserID,
			},

			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)
			ctx := context.Background()
			t.Cleanup(func() {
				cleanup()
			})
			tx := testfixtures.MakeTx(t, ctx, db)

			rideService := MakeService()

			updatedRide, err := rideService.UpdateRide(ctx, tx, tc.ride)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, updatedRide)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedRide)
				AssertEqualRide(t, tc.expectedRide, updatedRide)
			}
		})
	}
}

func TestRideService_DeleteRide(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc   string
		rideID int

		expectedErr         bool
		expectedAffectedRow bool
	}{
		{
			desc:                "happy path: delete ride that exists",
			rideID:              TestExistingRide.ID,
			expectedAffectedRow: true,
		},
		{
			desc:                "error path: delete ride that does not exist. no error but should not return ok",
			rideID:              999,
			expectedAffectedRow: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			rideService := MakeService()
			ctx := context.Background()

			db, cleanup := testfixtures.MakePostgres(t)
			t.Cleanup(func() {
				cleanup()
			})
			tx := testfixtures.MakeTx(t, ctx, db)
			affectedRow, err := rideService.DeleteRide(ctx, tx, tc.rideID)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.False(t, affectedRow)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, affectedRow, tc.expectedAffectedRow)
			}
		})
	}
}
