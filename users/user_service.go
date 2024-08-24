package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	ID               int
	Email            string
	StripeCustomerID string
}

type Service interface {
	GetUser(ctx context.Context, tx *sql.Tx, userID int) (*User, error)
	CreateUser(ctx context.Context, tx *sql.Tx, user *User) (*User, error)
	UpdateUser(ctx context.Context, tx *sql.Tx, user *User) (*User, error)
	DeleteUser(ctx context.Context, tx *sql.Tx, userID int) (bool, error)
}

type service struct {
}

func MakeService() Service {
	return &service{}
}

func (s *service) GetUser(ctx context.Context, tx *sql.Tx, userID int) (*User, error) {
	query := `
	SELECT 
	    id, email, stripe_customer_id 
	FROM rocket_rides.public.users 
	WHERE id = $1
	;
	`

	row := tx.QueryRowContext(ctx, query, userID)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.StripeCustomerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (s *service) CreateUser(ctx context.Context, tx *sql.Tx, user *User) (*User, error) {
	query := `
	INSERT INTO rocket_rides.public.users (
		email, stripe_customer_id
	) VALUES (
		$1, $2
	) RETURNING id
	;
	`
	err := tx.QueryRowContext(ctx, query, user.Email, user.StripeCustomerID).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, tx *sql.Tx, user *User) (*User, error) {
	query := `
	UPDATE rocket_rides.public.users 
	SET 
	    email = $2, 
		stripe_customer_id = $3 
	WHERE id = $1
	;
	`

	result, err := tx.ExecContext(ctx, query, user.ID, user.Email, user.StripeCustomerID)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New(fmt.Sprintf("no user with id %d", user.ID))
	}

	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, tx *sql.Tx, userID int) (bool, error) {
	query := `
	DELETE FROM rocket_rides.public.users 
	WHERE id = $1
	;
	`
	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
