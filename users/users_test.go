package users

import "testing"

func TestService_GetUser(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{
			desc: "happy path: get user that exists in db",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

		})
	}
}

func TestService_CreateUser(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{
			desc: "happy path: attempt to create valid user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{
			desc: "happy path: update user that exists",
		},
		{
			desc: "error path: update user that doesn't exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

		})
	}

}

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{
			desc: "happy path: delete user that exists",
		},
		{
			desc: "happy path: delete user that was already deleted",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

		})
	}

}
