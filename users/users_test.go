package users

import (
	"context"
	"github.com/anmho/idempotent-rides/testfixtures"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	TestUser1 = &User{
		ID:               *TestUser1ID,
		Email:            "awesome-user@email.com",
		StripeCustomerID: "sk_123",
	}
	NewTestUser = &User{
		ID:               999,
		Email:            "new-test-user@email.com",
		StripeCustomerID: "sk_999",
	}
)

func TestService_GetUser(t *testing.T) {
	tests := []struct {
		desc   string
		userID int

		expectedErr  bool
		expectedUser *User
	}{
		{
			desc:   "happy path: get user that exists in db",
			userID: TestUser1.ID,

			expectedUser: TestUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)

			t.Cleanup(func() {
				cleanup()
			})

			ctx := context.Background()
			tx := testfixtures.MakeTx(t, ctx, db)

			userService := MakeService()
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
		user *User

		expectedErr  bool
		expectedUser *User
	}{
		{
			desc:         "happy path: attempt to create a new user",
			user:         NewTestUser,
			expectedUser: NewTestUser,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)
			t.Cleanup(func() {
				cleanup()
			})

			ctx := context.Background()
			tx := testfixtures.MakeTx(t, ctx, db)

			userService := MakeService()
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
		user *User

		expectedErr  bool
		expectedUser *User
	}{
		{
			desc: "happy path: update user that exists",
			user: &User{
				ID:               TestUser1.ID,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account", // we make this stripe account on behalf of the user
			},
			expectedUser: &User{
				ID:               TestUser1.ID,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account",
			},
		},
		{
			desc: "error path: update user that doesn't exist",
			user: &User{
				ID:               759123,
				Email:            "updated-test-user-email@xxx.com",
				StripeCustomerID: "sk_new-stripe-user-account", // we make this stripe account on behalf of the user
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db, cleanup := testfixtures.MakePostgres(t)
			t.Cleanup(func() {
				cleanup()
			})
			ctx := context.Background()

			tx := testfixtures.MakeTx(t, ctx, db)
			testfixtures.MakeTx(t, ctx, db)

			userService := MakeService()
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
			userID:              TestUser1.ID,
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
			db, cleanup := testfixtures.MakePostgres(t)

			ctx := context.Background()
			t.Cleanup(func() {
				cleanup()
			})
			userService := MakeService()

			tx := testfixtures.MakeTx(t, ctx, db)
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
