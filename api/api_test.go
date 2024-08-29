package api_test

import (
	"encoding/json"
	"github.com/anmho/idempotent-rides/api"
	"github.com/anmho/idempotent-rides/idempotency"
	"github.com/anmho/idempotent-rides/rides"
	"github.com/anmho/idempotent-rides/send"
	"github.com/anmho/idempotent-rides/test"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v79/customer"
	"net/http"
	"strings"
	"testing"
)

const (
	emptyIdempotencyKey = ""
	dbIdempotencyKey    = "testKey"
	newIdempotencyKey   = "newKey"
)

var (
	emptyRequestBody = api.RideReservationParams{}
	JoshTestUser     = &users.User{
		ID:               1337,
		Email:            "jgoon@uiuc.edu",
		StripeCustomerID: "cus_Qjlq6Bl2Bb2nTq",
	}
)

func TestServer_handleRideReservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc           string
		idempotencyKey string
		method         idempotency.RequestMethod
		params         api.RideReservationParams

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
			params: api.RideReservationParams{
				UserID: &JoshTestUser.ID,
				Origin: &rides.Coordinate{},
				Target: &rides.Coordinate{},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			desc:           "POST /rides: valid idempotency key that is in the database",
			idempotencyKey: dbIdempotencyKey,
			method:         http.MethodPost,
			params: api.RideReservationParams{
				UserID: &JoshTestUser.ID,
				Origin: &rides.Coordinate{},
				Target: &rides.Coordinate{},
			},

			expectedStatus: http.StatusCreated,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			srv := test.MakeTestServer(t)

			body := strings.NewReader(string(must(json.Marshal(tc.params))))
			client := srv.Client()
			req := must(http.NewRequest(tc.method.String(), srv.URL+"/rides", body))
			req.Header.Set(idempotency.HeaderKey, tc.idempotencyKey)
			resp := must(client.Do(req))
			require.NotNil(t, resp)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestServer_handleRegisterUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc   string
		params api.RegisterUserParams

		expectedStatus int
		expectedUser   *users.User
	}{
		{
			desc: "happy path: registering valid user with available email",
			params: api.RegisterUserParams{
				Email: "testuser@uiuc.edu",
			},
			expectedStatus: http.StatusCreated,
			expectedUser: &users.User{
				ID:    1,
				Email: "testuser@uiuc.edu",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			srv := test.MakeTestServer(t)

			body := strings.NewReader(string(must(json.Marshal(tc.params))))
			client := srv.Client()
			req := must(http.NewRequest(http.MethodPost, srv.URL+"/users", body))

			resp := must(client.Do(req))
			require.NotNil(t, resp)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			user, err := send.Read[*users.User](resp.Body)
			require.NoError(t, err)
			assert.NotNil(t, user)

			assert.Equal(t, tc.expectedUser.Email, user.Email)

			t.Cleanup(func() {
				_, err := customer.Del(
					user.StripeCustomerID, nil,
				)
				assert.NoError(t, err)
			})
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
