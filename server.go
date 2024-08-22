package main

import (
	"database/sql"
	"net/http"
)

type paymentServer struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, db)

	// register middlewares

	return mux
}

func handleCharge(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Each atomic phase will be wrapped in a transaction.
		// Checkpoint 1: started
		//	Create idempotency key

		// Checkpoint 2: ride_created
		//	Create ride
		//	Create ride audit record
		// Checkpoint 3: Create ride audit record
		// 	Charge user via Stripe
		//	Update ride
		// Checkpoint 4: Charge user via Stripe
		//	Stage send receipt job
		// Checkpoint 5: Update ride
		// 	Update idempotency key
	}
}
