package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Key struct {
	ID        int
	CreatedAt time.Time
	Key       string
	LastRunAt time.Time
	LockedAt  sql.Null[time.Time]
	// Request metadata
	RequestMethod RequestMethod
	RequestParams []byte
	RequestPath   string
	// Response metadata
	ResponseCode sql.Null[int]
	ResponseBody sql.Null[[]byte]

	RecoveryPoint RecoveryPointEnum
	UserID        int
}

type KeyParams struct {
	Key           string
	RequestMethod RequestMethod
	RequestParams []byte
	RequestPath   string
	UserID        int
}

func scanAllKeyFields(row *sql.Row, key *Key) error {
	return row.Scan(
		&key.ID, &key.CreatedAt, &key.Key, &key.LastRunAt, &key.LockedAt,
		&key.RequestMethod, &key.RequestParams, &key.RequestPath,
		&key.ResponseCode, &key.ResponseBody,
		&key.RecoveryPoint, &key.UserID,
	)
}

func FindKey(
	ctx context.Context,
	tx *sql.Tx,
	userID int,
	key string,
) (*Key, error) {
	stmt, err := tx.PrepareContext(ctx,
		`
		SELECT
		    id, created_at, idempotency_key, last_run_at, locked_at,
		    request_method, request_params, request_path,
		    response_code, response_body,
		    recovery_point, user_id
		FROM idempotency_keys
		WHERE 
			user_id = $1 AND idempotency_key = $2;`,
	)
	if err != nil {
		return nil, fmt.Errorf("preparing context: %w", err)
	}
	defer stmt.Close()

	var iKey Key
	row := stmt.QueryRowContext(ctx, userID, key)
	err = scanAllKeyFields(row, &iKey)

	if err != nil {
		return nil, fmt.Errorf("querying row: %w", err)
	}

	return &iKey, nil
}

func InsertKey(
	ctx context.Context,
	tx *sql.Tx,
	params KeyParams,
) (*Key, error) {
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO rocket_rides.public.idempotency_keys (
		idempotency_key, 
		request_method,  
		request_params,
		request_path,
		recovery_point,
		user_id
		) VALUES (
		  $1, $2, $3, $4, $5, $6
		) RETURNING 
		    id, created_at, idempotency_key, last_run_at, locked_at, 
		 	request_method, request_params, request_path,
			response_body, response_code,
			recovery_point, user_id
		;`,
	)
	defer stmt.Close()

	if err != nil {
		return nil, fmt.Errorf("preparing statement: %w", err)
	}

	row := stmt.QueryRow(
		params.Key,
		params.RequestMethod,
		params.RequestParams,
		params.RequestPath,
		StartedRecoveryPoint,
		params.UserID,
	)

	var key Key
	err = scanAllKeyFields(row, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func UpdateKey(ctx context.Context, tx *sql.Tx, key *Key) (*Key, error) {
	if key == nil {
		return nil, errors.New("key must not be nil")
	}
	stmt, err := tx.PrepareContext(ctx,
		`
		UPDATE idempotency_keys
		SET 
			created_at = $2,
			idempotency_key = $3,
			last_run_at = $4,
			locked_at = $5,
			request_method = $6,
			request_params = $7,
			request_path = $8,
			response_code = $9,
			response_body = $10,
			recovery_point = $11,
			user_id = $12
		WHERE id = $1
		RETURNING 
			id, created_at, idempotency_key, last_run_at, locked_at, 
			request_method, request_params, request_path,
			response_code, response_body, 
			recovery_point, user_id
		;
	`)

	if err != nil {
		return nil, err
	}

	var updatedKey Key
	err = stmt.QueryRowContext(ctx,
		key.ID, key.CreatedAt, key.Key, key.LastRunAt, key.LockedAt,
		key.RequestMethod, key.RequestParams, key.RequestPath,
		key.ResponseCode, key.ResponseBody,
		key.RecoveryPoint, key.UserID).Scan(
		&updatedKey.ID, &updatedKey.CreatedAt, &updatedKey.Key, &updatedKey.LastRunAt, &updatedKey.LockedAt,
		&updatedKey.RequestMethod, &updatedKey.RequestParams, &updatedKey.RequestPath,
		&updatedKey.ResponseCode, &updatedKey.ResponseBody,
		&updatedKey.RecoveryPoint, &updatedKey.UserID,
	)
	//err = scanAllKeyFields(row, &updatedKey)
	if err != nil {
		return nil, err
	}
	return &updatedKey, nil
}

func DeleteIdempotencyKey(tx *sql.Tx, key *Key) error {
	return nil
}
