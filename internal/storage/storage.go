package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/RIBorisov/gophermart/internal/models/orders"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models"
	"github.com/RIBorisov/gophermart/internal/models/register"
)

type Store interface {
	SaveUser(ctx context.Context, user *register.Request) (string, error)
	GetUser(ctx context.Context, login string) (UserRow, error)
	SaveOrder(ctx context.Context, orderNo string) error
	GetOrders(ctx context.Context) ([]orderEntity, error)
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
		insertStmt        = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id`
		loginShouldBeUniq = "idx_login_is_unique"
	)

	var userID string
	row := d.pool.QueryRow(ctx, insertStmt, user.Login, user.Password)
	err := row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == loginShouldBeUniq {
				return "", errs.ErrUserExists
			}
		}
		return "", fmt.Errorf("failed query row request: %w", err)
	}

	return userID, nil
}

type UserRow struct {
	ID       string `db:"user_id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

func (d *DB) GetUser(ctx context.Context, login string) (UserRow, error) {
	const getStmt = `SELECT user_id, login, password FROM users WHERE login = $1`
	row := d.pool.QueryRow(ctx, getStmt, login)
	var (
		userRow   UserRow
		passBytes []byte
	)
	if err := row.Scan(&userRow.ID, &userRow.Login, &passBytes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserRow{}, errs.ErrUserNotExists
		}
		return UserRow{}, fmt.Errorf("failed scan row: %w", err)
	}
	userRow.Password = string(passBytes)

	return userRow, nil
}

// SaveOrder checks if order already registered and returns corresponding error
// otherwise saving the new order
func (d *DB) SaveOrder(ctx context.Context, orderNo string) error {
	const (
		insertStmt = `INSERT INTO orders (order_id, user_id) VALUES ($1, $2)`
		selectStmt = `SELECT order_id, user_id FROM orders WHERE order_id = $1`
	)
	var existedOrderNo, userID string
	err := d.pool.QueryRow(ctx, selectStmt, orderNo).Scan(&existedOrderNo, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			d.log.Debug("order number not found in db and can be stored", "orderNo", orderNo)
		} else {
			return fmt.Errorf("failed scan order from db: %w", err)
		}
	}

	ctxUserID, ok := ctx.Value(models.CtxUserIDKey).(string)
	if !ok {
		return errs.ErrGetUserFromContext
	}

	// orderNo already in db, therefore we should return
	// one of these errors: ErrOrderCreatedAlready, ErrAnotherUserOrderCreated
	if existedOrderNo != "" {
		if userID == ctxUserID {
			return errs.ErrOrderCreatedAlready
		}

		return errs.ErrAnotherUserOrderCreated
	}

	_, err = d.pool.Exec(ctx, insertStmt, orderNo, ctxUserID)
	if err != nil {
		return fmt.Errorf("failed execute statement: %w", err)
	}

	return nil
}

type orderEntity struct {
	OrderID    string        `db:"order_id"`
	UserID     string        `db:"user_id"`
	Status     orders.Status `db:"status"`
	Bonus      int           `db:"bonus"`
	UploadedAt time.Time     `db:"uploaded_at"`
}

func (d *DB) GetOrders(ctx context.Context) ([]orderEntity, error) {
	const stmt = `SELECT * FROM orders WHERE user_id = $1` // ORDER BY uploaded_at DESC`
	var orderList []orderEntity
	fmt.Println(orderList)
	ctxUserID, ok := ctx.Value(models.CtxUserIDKey).(string)
	if !ok {
		return nil, errs.ErrGetUserFromContext
	}
	rows, err := d.pool.Query(ctx, stmt, ctxUserID)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	for rows.Next() {
		var o orderEntity
		if err = rows.Scan(&o.OrderID, &o.UserID, &o.Status, &o.Bonus, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed scan into order entity: %w", err)
		}
		orderList = append(orderList, o)
	}

	return orderList, nil
}
