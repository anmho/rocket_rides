package users

var (
	TestUser1ID   = GetPtr(123)
	TestUser2ID   = GetPtr(456)
	TestUserEmail = "awesome-user@email.com"
)

var (
	TestUser1 = &User{
		ID:               *TestUser1ID,
		Email:            "awesome-user@email.com",
		StripeCustomerID: "sk_123",
	}

	NewTestUserNotInDB = &User{
		ID:               999,
		Email:            "new-test-user@email.com",
		StripeCustomerID: "sk_999",
	}
)

type User struct {
	ID               int
	Email            string
	StripeCustomerID string
}

func New(email string, customerID string) *User {

	return &User{
		Email:            email,
		StripeCustomerID: customerID,
	}
}

func GetPtr[T any](t T) *T {
	return &t
}
