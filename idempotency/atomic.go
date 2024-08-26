package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anmho/idempotent-rides/scope"
	"log/slog"
	"time"
)

type AtomicPhaseResult interface {
	UpdateKeyForNextPhase(ctx context.Context, tx *sql.Tx, key *Key) (*Key, error)
}

var _ AtomicPhaseResult = (*NoOpResult)(nil)

// TODO Update these models to return the relevant data from the previous phase

type NoOpResult struct{}

func (r *NoOpResult) UpdateKeyForNextPhase(ctx context.Context, tx *sql.Tx, key *Key) (*Key, error) {
	return nil, nil
}

var _ AtomicPhaseResult = (*RecoveryPointResult)(nil)

// TODO Update these models to return the relevant data from the previous phase

// RecoveryPointResult represents an action to set a new recovery point. One possible option for a
// return from a #atomic_phase block.
type RecoveryPointResult struct {
	RecoveryPoint RecoveryPointEnum
}

func NewRecoveryPointResult(recoveryPoint RecoveryPointEnum) *RecoveryPointResult {
	return &RecoveryPointResult{
		RecoveryPoint: recoveryPoint,
	}
}

func (r *RecoveryPointResult) UpdateKeyForNextPhase(ctx context.Context, tx *sql.Tx, key *Key) (*Key, error) {
	if key == nil {
		return nil, errors.New("nil key in update")
	}
	var newKey = new(Key)
	*newKey = *key
	newKey.RecoveryPoint = r.RecoveryPoint

	scope.GetLogger().Error("UpdateKeyForNextPhase")
	return newKey, nil
}

var _ AtomicPhaseResult = (*ResponseResult)(nil)

// TODO Update these models to return the relevant data from the previous phase

type ResponseResult struct {
	Status int
	Data   any
}

func NewResponseResult(status int, data any) *ResponseResult {
	return &ResponseResult{Status: status, Data: data}
}

func (r *ResponseResult) UpdateKeyForNextPhase(ctx context.Context, tx *sql.Tx, key *Key) (*Key, error) {
	newKey := new(Key)
	*newKey = *key
	newKey.LockedAt = sql.Null[time.Time]{
		Valid: false,
	}
	newKey.RecoveryPoint = FinishedRecoveryPoint
	newKey.ResponseCode = sql.Null[int]{
		V:     r.Status,
		Valid: true,
	}
	if r.Data != nil {
		b, err := json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
		newKey.ResponseBody = sql.Null[[]byte]{
			V:     b,
			Valid: true,
		}
	}

	return UpdateKey(ctx, tx, newKey)
}

type BlockFunc func(tx *sql.Tx) (AtomicPhaseResult, error)

func AtomicPhase(ctx context.Context, key *Key, db *sql.DB, block BlockFunc) (*Key, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		scope.GetLogger().Error("failed to rollback", slog.Any("cause", err))
	}(tx)

	result, err := block(tx)
	var updatedKey *Key
	if err == nil {
		switch result.(type) {
		case *NoOpResult, *RecoveryPointResult, *ResponseResult:
			updatedKey, err = result.UpdateKeyForNextPhase(ctx, tx, key)
		default:
			err = errors.New("invalid atomic result type")
		}
		err = tx.Commit()
	}

	if err != nil {
		scope.GetLogger().Error("error during transaction", slog.Any("cause", err))
		if updatedKey != nil {
			// If we're leaving under an error condition, try to unlock the idempotency
			// key right away so that another request can try again.
			// release the idempotency key lock
			updatedKey.LockedAt = sql.Null[time.Time]{}
			_, err = UpdateKey(ctx, tx, updatedKey)
			if err != nil {
				return nil, err
			}
			// this needs to be committed
		}
		return nil, err
	}
	return updatedKey, nil
}
