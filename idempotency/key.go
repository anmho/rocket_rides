package idempotency

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Key struct {
	ID        int
	CreatedAt time.Time
	Key       string
	LastRunAt time.Time
	LockedAt  time.Time
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

func GetIdempotencyKey(
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
	err = stmt.QueryRowContext(ctx, userID, key).Scan(
		&iKey.ID, &iKey.CreatedAt, &iKey.Key, &iKey.LastRunAt, &iKey.LockedAt,
		&iKey.RequestMethod, &iKey.RequestParams, &iKey.RequestPath,
		&iKey.ResponseCode, &iKey.ResponseBody,
		&iKey.RecoveryPoint, &iKey.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying row: %w", err)
	}

	return &iKey, nil
}

func InsertIdempotencyKey(
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

	var key Key

	row := stmt.QueryRow(
		params.Key,
		params.RequestMethod,
		params.RequestParams,
		params.RequestPath,
		RecoveryPointStarted,
		params.UserID,
	)
	err = row.Scan(
		&key.ID, &key.CreatedAt, &key.Key, &key.LastRunAt, &key.LockedAt,
		&key.RequestMethod, &key.RequestParams, &key.RequestPath,
		&key.ResponseBody, &key.ResponseCode,
		&key.RecoveryPoint, &key.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func UpdateIdempotencyKey(tx *sql.Tx, key *Key) (*Key, error) {
	return nil, nil
}

func DeleteIdempotencyKey(tx *sql.Tx, key *Key) error {
	return nil
}
