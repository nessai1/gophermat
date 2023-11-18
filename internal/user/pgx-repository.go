package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type PGXRepository struct {
	db *sql.DB
}

const PostgresErrCodeUniqueViolation = "23505"

func CreatePGXRepository(db *sql.DB) *PGXRepository {
	return &PGXRepository{db: db}
}

func (repository *PGXRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	row := repository.db.QueryRowContext(ctx, "SELECT login, password, balance FROM user WHERE login = $1", login)

	if row.Err() != nil && errors.Is(row.Err(), sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if row.Err() != nil {
		return nil, fmt.Errorf("error while query for get user by login: %w", row.Err())
	}

	var user User

	err := row.Scan(&user.Login, &user.password, &user.Balance)
	if err != nil {
		return nil, fmt.Errorf("error while scan row for get user by login: %w", err)
	}

	return &user, nil
}

func (repository *PGXRepository) CreateUser(ctx context.Context, user *User) error {
	_, err := repository.db.ExecContext(ctx, "INSERT INTO user(login, password, balance) VALUES ($1, $2, $3)", user.Login, user.password, user.Balance)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == PostgresErrCodeUniqueViolation {
				return ErrLoginAlreadyExists
			}
		}

		return fmt.Errorf("error while creating user: %w", err)
	}

	return nil
}
