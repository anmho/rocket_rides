package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/anmho/idempotent-rides/api"
	"github.com/caarlos0/env/v11"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v79"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const (
	port = 8080
)

func MakeConnString(
	user, pass, host, port, name string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, name, port)
}

type config struct {
	DBUser string `env:"DB_USER"`
	DBPass string `env:"DB_PASS"`
	DBHost string `env:"DB_HOST"`
	DBPort string `env:"DB_PORT"`
	DBName string `env:"DB_NAME"`

	StripeKey string `env:"STRIPE_KEY"`
}

func main() {
	if os.Getenv("STAGE") == "" || os.Getenv("STAGE") == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalln("Error loading .env file", err)
		}
	}

	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln("error parsing config")
	}
	fmt.Printf("%+v\n", cfg)
	dbURL := MakeConnString(cfg.DBUser, cfg.DBPass, cfg.DBPort, cfg.DBHost, cfg.DBName)

	stripe.Key = cfg.StripeKey
	db, err := sql.Open("pgx", dbURL)
	mux := api.MakeServer(db)

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
