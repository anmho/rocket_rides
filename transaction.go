package main

import (
	"context"
	"database/sql"
)

type AtomicPhaseResultType int

const (
	NoOpResult AtomicPhaseResultType = iota
	RecoveryPointResult
	ResponseResult
)

type AtomicPhaseResult struct {
	Type AtomicPhaseResultType
}
type BlockFunc func(tx *sql.Tx) (AtomicPhaseResult, error)

func AtomicPhase(ctx context.Context, key *IdempotencyKey, db *sql.DB, block BlockFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	result, err := block(tx)
	if err != nil {
		return err
	}
	err = tx.Rollback()
	if err != nil {
		return err
	}

	if err == nil {
		switch result.Type {
		case NoOpResult:
		case RecoveryPointResult:
		case ResponseResult:
		}
	} else {
	}
	return nil
}
