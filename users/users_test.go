package users_test

import (
	"context"
	"github.com/anmho/idempotent-rides/test"
	"github.com/anmho/idempotent-rides/users"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_GetUser(t *testing.T) {
	tests := []struct {
		desc   string
		userID int

		expectedErr  bool
		expectedUser *users.User
	}{
		{
			desc:   "happy path: get user that exists in db",
			userID: users.TestUser1.ID,

			expectedUser: users.TestUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)
			ctx := context.Background()
			tx := test.MakeTx(t, ctx, db)

			userService := users.MakeService()
			user, err := userService.GetUser(ctx, tx, tc.userID)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.Equal(t, tc.expectedUser, user)
			}
		})
	}
}

func TestService_CreateUser(t *testing.T) {
	tests := []struct {
		desc string
		user *users.User

		expectedErr  bool
		expectedUser *users.User
	}{
		{
			desc:         "happy path: attempt to create a new user",
			user:         users.NewTestUserNotInDB,
			expectedUser: users.NewTestUserNotInDB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)

			ctx := context.Background()
			tx := test.MakeTx(t, ctx, db)

			userService := users.MakeService()
			user, err := userService.CreateUser(ctx, tx, tc.user)

			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.user, user)
			}
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	tests := []struct {
		desc string
		user *users.User

		expectedErr  bool
		expectedUser *users.User
	}{
		{
			desc: "happy path: update user that exists",
			user: &users.User{
				ID:               users.TestUser1.ID,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account",
			},
			expectedUser: &users.User{
				ID:               users.TestUser1.ID,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account",
			},
		},
		{
			desc: "error path: update user that doesn't exist",
			user: &users.User{
				ID:               759123,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account",
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)

			ctx := context.Background()

			tx := test.MakeTx(t, ctx, db)
			test.MakeTx(t, ctx, db)

			userService := users.MakeService()
			user, err := userService.UpdateUser(ctx, tx, tc.user)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
		})
	}

}

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		desc   string
		userID int

		expectedAffectedRow bool
		expectedErr         bool
	}{
		{
			desc:                "happy path: delete user that exists",
			userID:              users.TestUser1.ID,
			expectedAffectedRow: true,
			expectedErr:         false,
		},
		{
			desc:   "happy path: delete user that was already deleted or does not exist",
			userID: 9123120,

			expectedAffectedRow: false,
			expectedErr:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := test.MakePostgres(t)
			ctx := context.Background()
			userService := users.MakeService()

			tx := test.MakeTx(t, ctx, db)
			affectedRow, err := userService.DeleteUser(ctx, tx, tc.userID)
			if tc.expectedAffectedRow {
				assert.Error(t, err)
				assert.False(t, affectedRow)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, affectedRow, tc.expectedAffectedRow)
			}
		})
	}

}
