package api

import (
	"database/sql"
	"github.com/anmho/idempotent-rides/audit"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/users"
	"net/http"
)

func registerRoutes(
	mux *http.ServeMux,
	db *sql.DB,
	rideService rides.Service,
	auditService audit.Service,
	userService users.Service) {

	mux.HandleFunc("POST /rides", MakeHandlerFunc(handleRideReservation(db, rideService, auditService, userService)))

}
