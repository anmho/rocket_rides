package idempotency

import (
	"context"
	"database/sql"
	"github.com/anmho/idempotent-rides/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

const (
	TestUserID = 123
)

var (
	TestKeyStarted = Key{
		ID:            736,
		CreatedAt:     time.Time{},
		Key:           "testKeyStarted",
		LastRunAt:     time.Time{},
		LockedAt:      sql.Null[time.Time]{},
		RequestMethod: http.MethodPost,
		RequestParams: []byte("{}"),
		RequestPath:   "/charges",
		ResponseCode:  sql.Null[int]{},
		ResponseBody:  sql.Null[[]byte]{},
		RecoveryPoint: StartedRecoveryPoint,
		UserID:        TestUserID,
	}
	TestKeyRideCreated = Key{
		ID:            737,
		CreatedAt:     time.Time{},
		Key:           "testKeyRideCreated",
		LastRunAt:     time.Time{},
		LockedAt:      sql.Null[time.Time]{},
		RequestMethod: http.MethodPost,
		RequestParams: []byte("{}"),
		RequestPath:   "/rides",
		ResponseCode:  sql.Null[int]{},
		ResponseBody:  sql.Null[[]byte]{},
		RecoveryPoint: RideCreatedRecoveryPoint,
		UserID:        123,
	}
	TestKeyRideChargeCreated = Key{
		ID:            737,
		CreatedAt:     time.Time{},
		Key:           "testKeyChargeCreated",
		LastRunAt:     time.Time{},
		LockedAt:      sql.Null[time.Time]{},
		RequestMethod: http.MethodPost,
		RequestParams: []byte("{}"),
		RequestPath:   "/rides",
		ResponseCode:  sql.Null[int]{},
		ResponseBody:  sql.Null[[]byte]{},
		RecoveryPoint: "charge_created",
		UserID:        123,
	}
	TestKeyFinished = Key{
		ID:            738,
		CreatedAt:     time.Time{},
		Key:           "testKeyFinished",
		LastRunAt:     time.Time{},
		LockedAt:      sql.Null[time.Time]{},
		RequestMethod: http.MethodPost,
		RequestParams: []byte("{}"),
		RequestPath:   "/rides",
		ResponseCode:  sql.Null[int]{V: 201, Valid: true},
		ResponseBody:  sql.Null[[]byte]{V: []byte("{}"), Valid: true},
		RecoveryPoint: "finished",
		UserID:        123,
	}
)

func assertEqualIdempotencyKey(t *testing.T, expectedIdempotencyKey, idempotencyKey *Key) {
	assert.Equal(t, expectedIdempotencyKey.ID, idempotencyKey.ID, "key id")
	assert.Equal(t, expectedIdempotencyKey.Key, idempotencyKey.Key, "key strings")
	assert.Equal(t, expectedIdempotencyKey.UserID, idempotencyKey.UserID, "UserID")

	assert.Equal(t, expectedIdempotencyKey.RequestMethod, idempotencyKey.RequestMethod, "http method")
	assert.Equal(t, expectedIdempotencyKey.RequestPath, idempotencyKey.RequestPath, "request path")
	assert.Equal(t, expectedIdempotencyKey.RequestParams, idempotencyKey.RequestParams, "request params")

	assert.Equal(t, expectedIdempotencyKey.ResponseCode, idempotencyKey.ResponseCode, "send code")
	assert.Equal(t, expectedIdempotencyKey.ResponseBody, idempotencyKey.ResponseBody, "send body")
	assert.Equal(t, expectedIdempotencyKey.RecoveryPoint, idempotencyKey.RecoveryPoint, "recovery point")
}

func Test_GetIdempotencyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		userID int
		key    string

		expectedErr            bool
		expectedIdempotencyKey *Key
	}{
		{
			name:   "happy path: full idempotency key is present",
			userID: TestUserID,
			key:    "testKeyFinished",

			expectedErr: false,
			expectedIdempotencyKey: &Key{
				ID:            739,
				Key:           "testKeyFinished",
				RequestMethod: http.MethodPost,
				RequestParams: []byte("{}"),
				RequestPath:   "/rides",
				ResponseCode: sql.Null[int]{
					V:     201,
					Valid: true,
				},
				ResponseBody: sql.Null[[]byte]{
					V:     []byte("{}"),
					Valid: true,
				},
				RecoveryPoint: FinishedRecoveryPoint,
				UserID:        TestUserID,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db := test.MakePostgres(t)

			tx, err := db.BeginTx(ctx, nil)
			require.NoError(t, err)

			idempotencyKey, err := FindKey(ctx, tx, tc.userID, tc.key)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assertEqualIdempotencyKey(t, tc.expectedIdempotencyKey, idempotencyKey)
			}
		})
	}
}

func Test_InsertIdempotencyKey(t *testing.T) {
	t.Parallel()

	u1 := TestUserID
	tests := []struct {
		name   string
		params KeyParams

		expectedIdempotencyKey *Key
	}{
		{
			name: "happy path: insert new idempotency key with valid fields and empty body",
			params: KeyParams{
				Key:           "awesomeKey",
				RequestMethod: http.MethodPost,
				RequestParams: []byte("{}"),
				RequestPath:   "/charges",
				UserID:        u1,
			},

			// We will assume timestamps will work since they are harder to mock but we should find a way.
			expectedIdempotencyKey: &Key{
				ID:            1,
				Key:           "awesomeKey",
				RequestMethod: http.MethodPost,
				RequestParams: []byte("{}"),
				RequestPath:   "/charges",
				ResponseBody:  sql.Null[[]byte]{},
				ResponseCode:  sql.Null[int]{},
				RecoveryPoint: StartedRecoveryPoint,
				UserID:        u1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := test.MakePostgres(t)

			ctx := context.Background()
			tx, err := db.BeginTx(ctx, nil)
			require.NoError(t, err)
			require.NotNil(t, tx)

			idempotencyKey, err := InsertKey(ctx, tx, tc.params)
			require.NoError(t, err)
			require.NoError(t, err)
			assert.NotNil(t, idempotencyKey, "idempotency not nil")

			// skip timestamps since that would be difficult to mock
			assertEqualIdempotencyKey(t, tc.expectedIdempotencyKey, idempotencyKey)

		})
	}
}

func Test_UpdateIdempotencyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		key  *Key

		expectedErr bool
		expectedKey *Key
	}{
		{
			desc: "happy path: update ride created key that exists in the database to be charge created ",
			key: &Key{
				ID:            TestKeyRideCreated.ID,
				CreatedAt:     TestKeyRideCreated.CreatedAt,
				Key:           TestKeyRideCreated.Key,
				LastRunAt:     TestKeyRideCreated.LastRunAt,
				LockedAt:      TestKeyRideCreated.LockedAt,
				RequestMethod: TestKeyRideCreated.RequestMethod,
				RequestParams: TestKeyRideCreated.RequestParams, // update to
				RequestPath:   TestKeyRideCreated.RequestPath,
				ResponseCode:  TestKeyRideCreated.ResponseCode,
				ResponseBody:  TestKeyRideCreated.ResponseBody,
				RecoveryPoint: ChargeCreatedRecoveryPoint,
				UserID:        TestKeyRideCreated.UserID,
			},

			expectedKey: &Key{
				ID:            TestKeyRideCreated.ID,
				CreatedAt:     TestKeyRideCreated.CreatedAt,
				Key:           TestKeyRideCreated.Key,
				LastRunAt:     TestKeyRideCreated.LastRunAt,
				LockedAt:      TestKeyRideCreated.LockedAt,
				RequestMethod: TestKeyRideCreated.RequestMethod,
				RequestParams: TestKeyRideCreated.RequestParams, // update to
				RequestPath:   TestKeyRideCreated.RequestPath,
				ResponseCode:  TestKeyRideCreated.ResponseCode,
				ResponseBody:  TestKeyRideCreated.ResponseBody,
				RecoveryPoint: ChargeCreatedRecoveryPoint,
				UserID:        TestKeyRideCreated.UserID,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)

			ctx := context.Background()
			tx := must(db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
				ReadOnly:  false,
			}))

			updatedKey, err := UpdateKey(ctx, tx, tc.key)
			require.NotNil(t, updatedKey)
			require.NoError(t, err)

			assertEqualIdempotencyKey(t, updatedKey, tc.expectedKey)
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v

}
