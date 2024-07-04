package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Store interface {
	Register(ctx context.Context, user *models.RegisterRequest) (string, error)
}

type DB struct {
	*DBPool
	cfg *config.Config
	log *logger.Log
}

func (d *DB) Register(ctx context.Context, user *models.RegisterRequest) (string, error) {
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

func LoadStorage(ctx context.Context, cfg *config.Config, log *logger.Log) (Store, error) {
	pool, err := NewPool(ctx, cfg.Service.DatabaseDSN, log)
	if err != nil {
		return nil, fmt.Errorf("failed acquire new db pool: %w", err)
	}
	return &DB{pool, cfg, log}, nil

}

var (
	ErrUserExists        = errors.New("user already exists")
	ErrIncorrectPassword = errors.New("invalid password")
)
