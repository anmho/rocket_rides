package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/anmho/idempotent-rides/api"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"log/slog"
	"net/http"
)

const (
	dbUser = "admin"
	dbPass = "admin"
	dbPort = "5433"
	dbHost = "localhost"
	dbName = "rocket_rides"
)

const (
	port = 8080
)

var (
	dbURL = MakeConnString(dbUser, dbPass, dbPort, dbHost, dbName)
)

func MakeConnString(
	user, pass, host, port, name string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, name, port)
}

func main() {
	db, err := sql.Open("pgx", dbURL)
	mux := api.NewServer(db)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	if err != nil {
		log.Fatalln(err)
	}

	slog.Info("server starting", slog.Int("port", port))
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting api %s", slog.String("error", err.Error()))
		}
	}
}
