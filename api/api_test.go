package api

import (
	"encoding/json"
	"github.com/anmho/idempotent-rides/idempotency"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/testfixtures"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	emptyIdempotencyKey = ""
	dbIdempotencyKey    = "testKey"
	newIdempotencyKey   = "newKey"
)

var (
	emptyRequestBody = RideReservationParams{}
)

func TestServer_handleRideReservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc           string
		idempotencyKey string
		method         idempotency.RequestMethod
		params         RideReservationParams

		expectedStatus int
	}{
		{
			desc:           "POST /rides: idempotency key is empty. should return 400 bad request",
			idempotencyKey: emptyIdempotencyKey,
			method:         http.MethodPost,
			params:         emptyRequestBody,

			expectedStatus: http.StatusBadRequest,
		},
		{
			desc:           "DELETE /rides: unsupported method. should return 405",
			idempotencyKey: dbIdempotencyKey,
			method:         http.MethodDelete,
			params:         emptyRequestBody,

			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			desc:           "POST /rides: valid idempotency key that is not in the database",
			idempotencyKey: newIdempotencyKey,
			method:         http.MethodPost,
			params: RideReservationParams{
				UserID: users.TestUserID,
				Origin: rides.Coordinate{},
				Target: rides.Coordinate{},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			desc:           "POST /rides: valid idempotency key that is in the database",
			idempotencyKey: dbIdempotencyKey,
			method:         http.MethodPost,
			params: RideReservationParams{
				UserID: users.TestUserID,
				Origin: rides.Coordinate{},
				Target: rides.Coordinate{},
			},

			expectedStatus: http.StatusCreated,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)
			t.Cleanup(func() {
				cleanup()
			})

			rocketRides := NewServer(db)
			srv := httptest.NewServer(rocketRides)
			t.Cleanup(func() {
				srv.Close()
			})

			body := strings.NewReader(string(must(json.Marshal(tc.params))))
			client := srv.Client()
			req := must(http.NewRequest(tc.method.String(), srv.URL+"/rides", body))
			req.Header.Set(idempotency.HeaderKey, tc.idempotencyKey)
			res := must(client.Do(req))
			require.NotNil(t, res)
			assert.Equal(t, tc.expectedStatus, res.StatusCode)
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
