package users

import (
	"context"
	"database/sql"
)

type Service interface {
	GetUser(ctx context.Context, tx *sql.Tx, userID int) (*User, error)
	CreateUser(ctx context.Context, tx *sql.Tx, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, tx *sql.Tx, user *User) (bool, error)
}

type service struct {
}

func MakeService() Service {
	return &service{}
}

func (s *service) GetUser(context.Context, *sql.Tx, int) (*User, error) {
	//TODO implement me
	panic("implement me")
	return nil, nil
}

func (s *service) CreateUser(context.Context, *sql.Tx, *User) (*User, error) {
	//TODO implement me
	panic("implement me")
	return nil, nil
}

func (s *service) UpdateUser(context.Context, *User) (*User, error) {
	//TODO implement me
	panic("implement me")
	return nil, nil
}

func (s *service) DeleteUser(context.Context, *sql.Tx, *User) (bool, error) {
	//TODO implement me
	panic("implement me")
	return false, nil
}
