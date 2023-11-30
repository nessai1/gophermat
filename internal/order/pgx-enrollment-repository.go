package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nessai1/gophermat/internal/postgrescodes"
)

type PGXEnrollmentRepository struct {
	db *sql.DB
}

func CreatePGXEnrollmentRepository(db *sql.DB) *PGXEnrollmentRepository {
	return &PGXEnrollmentRepository{db: db}
}

func (repository *PGXEnrollmentRepository) GetByID(ctx context.Context, orderID string) (*Enrollment, error) {
	row := repository.db.QueryRowContext(ctx, "SELECT user_id, order_id, status, accrual FROM enrollment_order WHERE order_id = $1", orderID)

	if row.Err() != nil {
		return nil, fmt.Errorf("error while query for get user by login: %w", row.Err())
	}

	var enrollment Enrollment

	err := row.Scan(&enrollment.UserID, &enrollment.OrderID, &enrollment.Status, &enrollment.Accrual)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEnrollmentNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error while scan row for get enrollment by login: %w", err)
	}

	return &enrollment, nil
}

func (repository *PGXEnrollmentRepository) CreateNewOrder(ctx context.Context, orderID string, ownerID int) (*Enrollment, error) {
	_, err := repository.db.ExecContext(ctx, "INSERT INTO enrollment_order (order_id, user_id, status) VALUES ($1, $2, $3)", orderID, ownerID, EnrollmentStatusNew)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == postgrescodes.PostgresErrCodeUniqueViolation {
				return nil, ErrEnrollmentAlreadyExists
			}
		}

		return nil, fmt.Errorf("error while creating user: %w", err)
	}

	enrollment := Enrollment{
		UserID:  ownerID,
		OrderID: orderID,
		Status:  EnrollmentStatusNew,
		Accrual: 0,
	}

	return &enrollment, nil
}

func (repository *PGXEnrollmentRepository) ChangeStatus(ctx context.Context, orderID, status string) error {
	_, err := repository.db.ExecContext(ctx, "UPDATE enrollment_order SET status = $1 WHERE order_id = $2", status, orderID)
	return err
}
