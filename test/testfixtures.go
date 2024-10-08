package test

import (
	"context"
	"database/sql"
	"github.com/anmho/idempotent-rides/api"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v79"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}

	stripeKey := os.Getenv("STRIPE_KEY")
	stripe.Key = stripeKey
}

func MakeTestServer(t *testing.T) *httptest.Server {
	db := MakePostgres(t)
	rocketRides := api.MakeServer(db)
	srv := httptest.NewServer(rocketRides)
	t.Cleanup(func() {
		srv.Close()
	})

	return srv
}
func MakeTx(t *testing.T, ctx context.Context, db *sql.DB) *sql.Tx {
	require.NotNil(t, db)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	require.NoError(t, err)
	require.NotNil(t, tx)
	return tx
}

// MakePostgres returns server dependencies and a cleanup function for tests.
func MakePostgres(t *testing.T) *sql.DB {
	ctx := context.Background()
	pgContainer := createPostgres(ctx, t)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		shutdownPostgres(ctx, t, pgContainer)
	})

	return db
}

func createPostgres(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	pgContainer, err := postgres.Run(ctx,
		"postgres",
		postgres.WithDatabase("rocket_rides"),
		postgres.WithUsername("admin"),
		postgres.WithPassword("admin"),
		postgres.WithInitScripts("../sql/1-schema.sql", "../sql/2-data.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
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
