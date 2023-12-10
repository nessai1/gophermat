package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/nessai1/gophermat/internal/postgrescodes"

	"github.com/jackc/pgx/v5/pgconn"
)

type PGXRepository struct {
	db *sql.DB
}

func CreatePGXRepository(db *sql.DB) *PGXRepository {
	return &PGXRepository{db: db}
}

func (repository *PGXRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	row := repository.db.QueryRowContext(ctx, "SELECT id, login, password, balance FROM \"user\" WHERE login = $1", login)

	if row.Err() != nil {
		return nil, fmt.Errorf("error while query for get user by login: %w", row.Err())
	}

	var user User

	err := row.Scan(&user.ID, &user.Login, &user.password, &user.Balance)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error while scan row for get user by login: %w", err)
	}

	return &user, nil
}

func (repository *PGXRepository) CreateUser(ctx context.Context, user *User) error {
	_, err := repository.db.ExecContext(ctx, "INSERT INTO \"user\" (login, password, balance) VALUES ($1, $2, $3)", user.Login, user.password, user.Balance)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == postgrescodes.PostgresErrCodeUniqueViolation {
				return ErrLoginAlreadyExists
			}
		}

		return fmt.Errorf("error while creating user: %w", err)
	}

	return nil
}

func (repository *PGXRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	row := repository.db.QueryRowContext(ctx, "SELECT id, login, password, balance FROM \"user\" WHERE id = $1", id)

	if row.Err() != nil {
		return nil, fmt.Errorf("error while query for get user by id: %w", row.Err())
	}

	var user User

	err := row.Scan(&user.ID, &user.Login, &user.password, &user.Balance)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error while scan row for get user by id: %w", err)
	}

	return &user, nil
}

func (repository *PGXRepository) SetUserBalanceByID(ctx context.Context, userID int, balance int64) error {
	_, err := repository.db.ExecContext(ctx, "UPDATE \"user\" SET balance = $1 WHERE id = $2", balance, userID)
	return err
}
