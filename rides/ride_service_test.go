package rides_test

import (
	"context"
	"database/sql"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/test"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	TestExistingRide = &rides.Ride{
		ID: 1442,
		IdempotencyKeyID: sql.Null[int]{
			V:     738,
			Valid: true,
		},
		Origin: rides.Coordinate{
			Lat:  72,
			Long: 72,
		},
		Target: rides.Coordinate{
			Lat:  72,
			Long: 72,
		},
		StripeChargeID: sql.Null[string]{
			V:     "ch_456",
			Valid: true,
		},
		UserID: *users.TestUser2ID,
	}
	TestNewRide = &rides.Ride{
		ID:               1442,
		IdempotencyKeyID: sql.Null[int]{},
		Origin: rides.Coordinate{
			Lat:  72,
			Long: 72,
		},
		Target: rides.Coordinate{
			Lat:  72,
			Long: 72,
		},
		StripeChargeID: sql.Null[string]{},
		UserID:         *users.TestUser2ID,
	}
)

func AssertEqualRide(t *testing.T, expected, ride *rides.Ride) {
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

		expectedRide *rides.Ride
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
			db := test.MakePostgres(t)

			tx, err := db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
				ReadOnly:  false,
			})
			if err != nil {
				return
			}
			rideService := rides.MakeService()
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
		ride *rides.Ride

		expectedErr  bool
		expectedRide *rides.Ride
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
			db := test.MakePostgres(t)

			ctx := context.Background()
			tx, err := db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
				ReadOnly:  false,
			})
			require.NoError(t, err)
			require.NotNil(t, tx)

			rideService := rides.MakeService()

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
	updatedRide := &rides.Ride{
		ID: 1442,
		IdempotencyKeyID: sql.Null[int]{
			V:     738,
			Valid: true,
		},
		Origin: rides.Coordinate{
			Lat:  100,
			Long: 100,
		},
		Target: rides.Coordinate{
			Lat:  100,
			Long: 100,
		},
		UserID: 456,
	}
	tests := []struct {
		desc string
		ride *rides.Ride

		expectedErr  bool
		expectedRide *rides.Ride
	}{
		{
			desc: "happy path: update an existing ride. change the target",
			ride: updatedRide,

			expectedRide: updatedRide,
		},
		{
			desc: "error path: update a non-existent ride. ride id does not exist",
			ride: &rides.Ride{
				ID:               5823,
				IdempotencyKeyID: TestExistingRide.IdempotencyKeyID,
				Origin:           TestExistingRide.Origin,
				Target: rides.Coordinate{
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
			db := test.MakePostgres(t)
			ctx := context.Background()

			tx := test.MakeTx(t, ctx, db)

			rideService := rides.MakeService()

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
			rideService := rides.MakeService()
			ctx := context.Background()

			db := test.MakePostgres(t)

			tx := test.MakeTx(t, ctx, db)
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
