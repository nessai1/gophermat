package intransaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Transaction interface {
	InTransaction(context.Context, func(innerCtx context.Context) error) error
}

type PGXTransaction struct {
	db *sql.DB
}

func NewPGXTransaction(db *sql.DB) Transaction {
	return &PGXTransaction{db: db}
}

func (transaction *PGXTransaction) InTransaction(ctx context.Context, consistent func(innerCtx context.Context) error) error {
	tx, err := transaction.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error while start transaction: %w", err)
	}

	err = consistent(ctx)
	if err != nil {
		rollErr := tx.Rollback()
		if rollErr != nil {
			return fmt.Errorf("error while rollback unsuccessful transaction: %w", errors.Join(err, rollErr))
		}

		return fmt.Errorf("rollback unsuccessful transaction: %w", err)
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("error while commit successful transaction: %w", commitErr)
	}

	return nil
}
