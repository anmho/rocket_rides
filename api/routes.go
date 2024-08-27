package api

import (
	"database/sql"
	audit "github.com/anmho/idempotent-rides/audit_test"
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
	mux.HandleFunc("POST /users", MakeHandlerFunc(handleRegisterUser(db, userService)))

}
