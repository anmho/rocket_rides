package idempotency

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"testing"
	"time"
)

const (
	TestUserID = 123
)

func createPostgres(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	pgContainer, err := postgres.Run(ctx,
		"postgres",
		postgres.WithDatabase("rocket_rides"),
		postgres.WithUsername("admin"),
		postgres.WithPassword("admin"),
		postgres.WithInitScripts("./sql/1-schema.sql", "./sql/2-data.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, pgContainer)
	require.True(t, pgContainer.IsRunning())

	return pgContainer
}

func shutdownPostgres(ctx context.Context, t *testing.T, pgContainer *postgres.PostgresContainer) {
	if err := pgContainer.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate pgContainer: %s", err)
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func assertEqualIdempotencyKey(t *testing.T, expectedIdempotencyKey, idempotencyKey *Key) {
	assert.Equal(t, expectedIdempotencyKey.ID, idempotencyKey.ID, "key id")
	assert.Equal(t, expectedIdempotencyKey.Key, idempotencyKey.Key, "key strings")
	assert.Equal(t, expectedIdempotencyKey.UserID, idempotencyKey.UserID, "UserID")

	assert.Equal(t, expectedIdempotencyKey.RequestMethod, idempotencyKey.RequestMethod, "http method")
	assert.Equal(t, expectedIdempotencyKey.RequestPath, idempotencyKey.RequestPath, "request path")
	assert.Equal(t, expectedIdempotencyKey.RequestParams, idempotencyKey.RequestParams, "request params")

	assert.Equal(t, expectedIdempotencyKey.ResponseCode, idempotencyKey.ResponseCode, "response code")
	assert.Equal(t, expectedIdempotencyKey.ResponseBody, idempotencyKey.ResponseBody, "response body")
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
			key:    "testKey",

			expectedErr: false,
			expectedIdempotencyKey: &Key{
				ID:            738,
				Key:           "testKey",
				RequestMethod: http.MethodPost,
				RequestParams: []byte("{}"),
				RequestPath:   "/charges",
				ResponseCode: sql.Null[int]{
					V:     200,
					Valid: true,
				},
				ResponseBody: sql.Null[[]byte]{
					V:     []byte("{}"),
					Valid: true,
				},
				RecoveryPoint: RecoveryPointFinished,
				UserID:        TestUserID,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pgContainer := createPostgres(ctx, t)
			t.Cleanup(func() {
				shutdownPostgres(ctx, t, pgContainer)
			})

			connString := must(pgContainer.ConnectionString(ctx))
			db := must(sql.Open("pgx", connString))
			tx := must(db.BeginTx(ctx, nil))

			idempotencyKey, err := GetIdempotencyKey(ctx, tx, tc.userID, tc.key)
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
				RecoveryPoint: RecoveryPointStarted,
				UserID:        u1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pgContainer := createPostgres(ctx, t)

			t.Cleanup(func() {
				shutdownPostgres(ctx, t, pgContainer)
			})

			connStr := must(pgContainer.ConnectionString(ctx, "sslmode=disable"))
			db := must(sql.Open("pgx", connStr))
			tx := must(db.BeginTx(ctx, nil))

			idempotencyKey, err := InsertIdempotencyKey(ctx, tx, tc.params)
			require.NoError(t, err)
			require.NoError(t, err)
			assert.NotNil(t, idempotencyKey, "idempotency not nil")

			// skip timestamps since that would be difficult to mock
			assertEqualIdempotencyKey(t, tc.expectedIdempotencyKey, idempotencyKey)

		})
	}
}
