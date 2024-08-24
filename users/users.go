package users

var (
	TestUser1ID   = GetPtr(123)
	TestUser2ID   = GetPtr(456)
	TestUserEmail = "awesome-user@email.com"
)

func GetPtr[T any](t T) *T {
	return &t
}

type User struct {
}
