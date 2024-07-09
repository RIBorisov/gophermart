package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models"
	"github.com/RIBorisov/gophermart/internal/models/balance"
	"github.com/RIBorisov/gophermart/internal/models/orders"
	"github.com/RIBorisov/gophermart/internal/models/register"
)

type Store interface {
	SaveUser(ctx context.Context, user *register.Request) (string, error)
	GetUser(ctx context.Context, login string) (*UserRow, error)
	SaveOrder(ctx context.Context, orderNo string) error
	GetOrders(ctx context.Context) ([]orderEntity, error)
	GetBalance(ctx context.Context) (*BalanceEntity, error)
	BalanceWithdraw(ctx context.Context, req balance.WithdrawRequest) error
	GetWithdrawals(ctx context.Context) ([]withdrawalsEntity, error)
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

func getCtxUserID(ctx context.Context) (string, error) {
	ctxUserID, ok := ctx.Value(models.CtxUserIDKey).(string)
	if !ok {
		return "", errs.ErrGetUserFromContext
	}
	return ctxUserID, nil
}

func (d *DB) SaveUser(ctx context.Context, user *register.Request) (string, error) {
	const (
		insertStmt        = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id`
		loginShouldBeUniq = "idx_login_is_unique"
		insertBalanceStmt = `INSERT INTO balance(user_id) VALUES ($1)`
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

	err = d.pool.QueryRow(ctx, insertBalanceStmt, userID).Scan(&userID)

	return userID, nil
}

type UserRow struct {
	ID       string `db:"user_id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

func (d *DB) GetUser(ctx context.Context, login string) (*UserRow, error) {
	const getStmt = `SELECT user_id, login, password FROM users WHERE login = $1`
	row := d.pool.QueryRow(ctx, getStmt, login)
	var (
		uRow      UserRow
		passBytes []byte
	)
	if err := row.Scan(&uRow.ID, &uRow.Login, &passBytes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotExists
		}
		return nil, fmt.Errorf("failed scan row: %w", err)
	}
	uRow.Password = string(passBytes)

	return &uRow, nil
}

// SaveOrder checks if order already registered and returns corresponding error
// otherwise saving the new order
func (d *DB) SaveOrder(ctx context.Context, orderNo string) error {
	const (
		insertStmt = `INSERT INTO orders (order_id, user_id) VALUES ($1, $2)`
		selectStmt = `SELECT order_id, user_id FROM orders WHERE order_id = $1`
	)
	var existedOrderNo, existedUserID string
	err := d.pool.QueryRow(ctx, selectStmt, orderNo).Scan(&existedOrderNo, &existedUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			d.log.Debug("order number not found in db and can be stored", "orderNo", orderNo)
		} else {
			return fmt.Errorf("failed scan order from db: %w", err)
		}
	}

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return err
	}

	// orderNo already in db, therefore we should return
	// one of these errors: ErrOrderCreatedAlready, ErrAnotherUserOrderCreated
	if existedOrderNo != "" {
		if userID == existedUserID {
			return errs.ErrOrderCreatedAlready
		}

		return errs.ErrAnotherUserOrderCreated
	}

	_, err = d.pool.Exec(ctx, insertStmt, orderNo, userID)
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
	const stmt = `SELECT * FROM orders WHERE user_id = $1`
	var oList []orderEntity

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.pool.Query(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	for rows.Next() {
		var o orderEntity
		if err = rows.Scan(&o.OrderID, &o.UserID, &o.Status, &o.Bonus, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed scan into order entity: %w", err)
		}
		oList = append(oList, o)
	}

	return oList, nil
}

type BalanceEntity struct {
	UserID    string    `db:"user_id"`
	Current   float64   `db:"current"`
	Withdrawn float64   `db:"withdrawn"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *DB) GetBalance(ctx context.Context) (*BalanceEntity, error) {
	const stmt = `SELECT current, withdrawn FROM balance WHERE user_id = $1`
	userID, err := getCtxUserID(ctx)
	if err != nil {
		return nil, err
	}

	var b BalanceEntity
	err = d.pool.QueryRow(ctx, stmt, userID).Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotExists
		}
		return nil, fmt.Errorf("failed query row: %w", err)
	}
	return &b, nil
}

func (d *DB) BalanceWithdraw(ctx context.Context, req balance.WithdrawRequest) error {
	const (
		selectStmt = `SELECT current FROM balance WHERE user_id = $1`
		updateStmt = `UPDATE balance 
					  SET current = current - @sum, withdrawn = withdrawn + @sum
					  WHERE user_id = @userID`
		insertWithdrawalsStmt = `INSERT INTO withdrawals (user_id, order_id, amount) VALUES (@userID, @orderID, @sum)`
	)

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return err
	}

	var current float64

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: "read committed"})
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			d.log.Debug("failed rollback transaction", err)
		}
	}()

	if err = tx.QueryRow(ctx, selectStmt, userID).Scan(&current); err != nil {
		return fmt.Errorf("failed query row: %w", err)
	}

	if current < req.Sum {
		return errs.ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, updateStmt, pgx.NamedArgs{"sum": req.Sum, "userID": userID})
	if err != nil {
		return fmt.Errorf("failed execute update balance stmt: %w", err)
	}

	_, err = tx.Exec(ctx, insertWithdrawalsStmt, pgx.NamedArgs{"userID": userID, "orderID": req.Order, "sum": req.Sum})
	if err != nil {
		return fmt.Errorf("failed execute withdrawal request stmt: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed commit tx: %w", err)
	}

	return nil
}

type withdrawalsEntity struct {
	UserID      string    `db:"user_id"`
	OrderID     string    `db:"order_id"`
	Amount      float64   `db:"amount"`
	ProcessedAt time.Time `db:"processed_at"`
}

func (d *DB) GetWithdrawals(ctx context.Context) ([]withdrawalsEntity, error) {
	const stmt = `SELECT order_id, amount, processed_at::timestamptz FROM withdrawals WHERE user_id = $1`

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.pool.Query(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed query withdrawals: %w", err)
	}
	var wList []withdrawalsEntity
	for rows.Next() {
		var row withdrawalsEntity

		if err = rows.Scan(&row.OrderID, &row.Amount, &row.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed scan withdrawals row: %w", err)
		}

		wList = append(wList, row)
	}

	return wList, nil
}
