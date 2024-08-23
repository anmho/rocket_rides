package api

import (
	"database/sql"
	"net/http"
)

func registerRoutes(
	mux *http.ServeMux,
	db *sql.DB,
) {
	mux.HandleFunc("POST /rides", HandleRideReservation(db))
}
