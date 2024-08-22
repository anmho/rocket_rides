package main

import (
	"database/sql"
	"errors"
	"fmt"
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

var (
	dbURL = makeConnString(dbUser, dbPass, dbPort, dbHost, dbName)
)

func makeConnString(
	user, pass, host, port, name string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, name, port)
}

func main() {
	db, err := sql.Open("pgx", dbURL)
	mux := NewServer(db)

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err != nil {
		log.Fatalln(err)
	}

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server %s", slog.String("error", err.Error()))
		}
	}
}
