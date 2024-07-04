package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models/register"
)

type Store interface {
	SaveUser(ctx context.Context, user *register.Request) (string, error)
	GetUser(ctx context.Context, login string) (UserRow, error)
}

type DB struct {
	*DBPool
	cfg *config.Config
	log *logger.Log
}

func LoadStorage(ctx context.Context, cfg *config.Config, log *logger.Log) (Store, error) {
	pool, err := NewPool(ctx, cfg.Service.DatabaseDSN, log)
	if err != nil {
		return nil, fmt.Errorf("failed acquire new db pool: %w", err)
	}
	return &DB{pool, cfg, log}, nil

}

func (d *DB) SaveUser(ctx context.Context, user *register.Request) (string, error) {
	const (
		insertStmt        = "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"
		loginShouldBeUniq = "idx_login_is_unique"
	)

	var userID string
	row := d.pool.QueryRow(ctx, insertStmt, user.Login, user.Password)
	err := row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == loginShouldBeUniq {
				return "", ErrUserExists
			}
		}
		return "", fmt.Errorf("failed query row request: %w", err)
	}

	return userID, nil
}

type UserRow struct {
	ID       string `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
	//CreatedAt string `db:"created_at"` // TODO: пока не уверен что нужно поле
}

func (d *DB) GetUser(ctx context.Context, login string) (UserRow, error) {
	const getStmt = "SELECT id, login, password FROM users WHERE login = $1"
	row := d.pool.QueryRow(ctx, getStmt, login)
	var (
		userRow   UserRow
		passBytes []byte
	)
	if err := row.Scan(&userRow.ID, &userRow.Login, &passBytes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserRow{}, ErrUserNotExists
		}
		return UserRow{}, fmt.Errorf("failed scan row: %w", err)
	}
	userRow.Password = string(passBytes)

	return userRow, nil
}

var (
	ErrUserExists        = errors.New("user already exists")
	ErrIncorrectPassword = errors.New("invalid password")
	ErrUserNotExists     = errors.New("user not exists")
)
