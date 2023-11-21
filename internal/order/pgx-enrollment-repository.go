package order

import (
	"context"
	"database/sql"
)

type PGXEnrollmentRepository struct {
	db *sql.DB
}

func CreatePGXEnrollmentRepository(db *sql.DB) *PGXEnrollmentRepository {
	return &PGXEnrollmentRepository{db: db}
}

func (P PGXEnrollmentRepository) GetByID(ctx context.Context, orderID string) (*Enrollment, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGXEnrollmentRepository) CreateNewOrder(ctx context.Context, orderID string) (*Enrollment, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGXEnrollmentRepository) ChangeStatus(ctx context.Context, orderID, status string) error {
	//TODO implement me
	panic("implement me")
}
