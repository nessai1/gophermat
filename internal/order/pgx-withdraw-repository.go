package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PGXWithdrawRepository struct {
	db *sql.DB
}

func NewPGXWithdrawRepository(db *sql.DB) *PGXWithdrawRepository {
	return &PGXWithdrawRepository{db: db}
}

func (repository *PGXWithdrawRepository) AddWithdraw(ctx context.Context, userID int, orderID string, sum int64) (*Withdraw, error) {
	now := time.Now()
	_, err := repository.db.ExecContext(ctx, "INSERT INTO withdraw_order (order_id, user_id, sum, processed_at) VALUES ($1, $2, $3, $4)", orderID, userID, sum, now)

	if err != nil {
		return nil, fmt.Errorf("error while creating user withdraw: %w", err)
	}

	withdraw := Withdraw{
		OrderID:     orderID,
		Sum:         sum,
		ProcessedAt: now,
	}

	return &withdraw, nil
}

func (repository *PGXWithdrawRepository) GetWithdrawListByUserID(ctx context.Context, userID int) ([]*Withdraw, error) {
	rows, err := repository.db.QueryContext(ctx, "SELECT sum, order_id, processed_at FROM withdraw_order WHERE user_id = $1 ORDER BY processed_at ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawList := make([]*Withdraw, 0)

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}

		withdraw := Withdraw{}
		err = rows.Scan(&withdraw.Sum, &withdraw.OrderID, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawList = append(withdrawList, &withdraw)
	}

	return withdrawList, nil
}

func (repository *PGXWithdrawRepository) GetWithdrawSumByUserID(ctx context.Context, userID int) (int64, error) {
	row := repository.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(sum),0) FROM withdraw_order WHERE user_id = $1", userID)

	if row.Err() != nil {
		return 0, fmt.Errorf("error while query for get user withdraw sum by id: %w", row.Err())
	}

	var sum int64
	if err := row.Scan(&sum); err != nil {
		return 0, fmt.Errorf("error while scan sum after query of user withdraw sum: %w", err)
	}

	return sum, nil
}
