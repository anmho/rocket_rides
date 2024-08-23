package users

var (
	TestUserID    = GetPtr(123)
	TestUserEmail = "awesome-user@email.com"
)

func GetPtr[T any](t T) *T {
	return &t
}

type User struct {
}

func GetUser() {

}

func InsertUser() {

}
