package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type RecoveryPointEnum string

const (
	RecoveryPointStarted       RecoveryPointEnum = "started"
	RecoveryPointRideCreated                     = "ride_created"
	RecoveryPointChargeStarted                   = "charge_started"
	RecoveryPointFinished                        = "finished"
)

func (rp RecoveryPointEnum) IsValid() bool {
	switch rp {
	case RecoveryPointStarted, RecoveryPointRideCreated,
		RecoveryPointChargeStarted,
		RecoveryPointFinished:
		return true
	default:
		return false
	}
}

type RequestMethod string

func (m RequestMethod) IsValid() bool {
	switch m {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
		return true
	default:
		return false
	}
}

type IdempotencyKey struct {
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

type IdempotencyKeyParams struct {
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
) (*IdempotencyKey, error) {
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

	var iKey IdempotencyKey
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
	params IdempotencyKeyParams,
) (*IdempotencyKey, error) {
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

	var key IdempotencyKey

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

func UpdateIdempotencyKey(tx *sql.Tx, key *IdempotencyKey) (*IdempotencyKey, error) {
	return nil, nil
}

func DeleteIdempotencyKey(tx *sql.Tx, key *IdempotencyKey) error {
	return nil
}
