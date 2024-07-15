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
	GetUserOrders(ctx context.Context) ([]OrderEntity, error)
	GetBalance(ctx context.Context) (*BalanceEntity, error)
	BalanceWithdraw(ctx context.Context, req balance.WithdrawRequest) error
	GetWithdrawals(ctx context.Context) ([]WithdrawalsEntity, error)
	GetOrdersList(ctx context.Context) ([]string, error)
	UpdateOrder(ctx context.Context, data *orders.UpdateOrder) error
	ClosePool() error
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

var errGetUserFromContext = errors.New("failed get userID from context")

func getCtxUserID(ctx context.Context) (string, error) {
	ctxUserID, ok := ctx.Value(models.CtxUserIDKey).(string)
	if !ok {
		return "", errGetUserFromContext
	}
	return ctxUserID, nil
}

func (d *DB) ClosePool() error {
	d.pool.Close()
	return nil
}

func (d *DB) SaveUser(ctx context.Context, user *register.Request) (string, error) {
	const (
		insertStmt        = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id`
		loginShouldBeUniq = "idx_login_is_unique"
		insertBalanceStmt = `INSERT INTO balance(user_id) VALUES ($1)`
	)

	var userID string

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: "read committed"})
	if err != nil {
		return "", fmt.Errorf("failed begin tx: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			d.log.Warn("failed rollback transaction", "txError", err)
		}
	}()

	err = tx.QueryRow(ctx, insertStmt, user.Login, user.Password).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == loginShouldBeUniq {
				return "", ErrUserExists
			}
		}
		return "", fmt.Errorf("failed query row request: %w", err)
	}

	if _, err = tx.Exec(ctx, insertBalanceStmt, userID); err != nil {
		return "", fmt.Errorf("failed insert balance row: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed commit tx: %w", err)
	}

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
		uRow UserRow
		pass []byte
	)
	if err := row.Scan(&uRow.ID, &uRow.Login, &pass); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotExists
		}
		return nil, fmt.Errorf("failed scan row: %w", err)
	}
	uRow.Password = string(pass)

	return &uRow, nil
}

// SaveOrder checks if order already registered and returns corresponding error
// otherwise saving the new order.
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
			return ErrOrderCreatedAlready
		}

		return ErrAnotherUserOrderCreated
	}

	_, err = d.pool.Exec(ctx, insertStmt, orderNo, userID)
	if err != nil {
		return fmt.Errorf("failed execute statement: %w", err)
	}

	return nil
}

type OrderEntity struct {
	Status     orders.Status `db:"status"`
	UploadedAt time.Time     `db:"uploaded_at"`
	OrderID    string        `db:"order_id"`
	UserID     string        `db:"user_id"`
	Bonus      float32       `db:"bonus"`
}

func (d *DB) GetUserOrders(ctx context.Context) ([]OrderEntity, error) {
	const stmt = `SELECT * FROM orders WHERE user_id = $1`
	var oList []OrderEntity

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.pool.Query(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	for rows.Next() {
		var o OrderEntity
		if err = rows.Scan(&o.OrderID, &o.UserID, &o.Status, &o.Bonus, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed scan into order entity: %w", err)
		}
		oList = append(oList, o)
	}

	return oList, nil
}

type BalanceEntity struct {
	UpdatedAt time.Time `db:"updated_at"`
	UserID    string    `db:"user_id"`
	Current   float32   `db:"current"`
	Withdrawn float32   `db:"withdrawn"`
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
			return nil, ErrUserNotExists
		}
		return nil, fmt.Errorf("failed query row: %w", err)
	}
	return &b, nil
}

func (d *DB) BalanceWithdraw(ctx context.Context, req balance.WithdrawRequest) error {
	const (
		selectStmt = `SELECT current FROM balance WHERE user_id = $1`
		updateStmt = `UPDATE balance 
					  SET current = current - @sum, withdrawn = withdrawn + @sum, updated_at = NOW()
					  WHERE user_id = @userID`
		insertWithdrawalsStmt = `INSERT INTO withdrawals (user_id, order_id, amount) VALUES (@userID, @orderID, @sum)`
	)

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return err
	}

	var current float32

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: "read committed"})
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			d.log.Warn("failed rollback transaction", "txErr", err)
		}
	}()

	if err = tx.QueryRow(ctx, selectStmt, userID).Scan(&current); err != nil {
		return fmt.Errorf("failed query row: %w", err)
	}

	if current < req.Sum {
		return ErrInsufficientFunds
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

type WithdrawalsEntity struct {
	ProcessedAt time.Time `db:"processed_at"`
	UserID      string    `db:"user_id"`
	OrderID     string    `db:"order_id"`
	Amount      float32   `db:"amount"`
}

func (d *DB) GetWithdrawals(ctx context.Context) ([]WithdrawalsEntity, error) {
	const stmt = `SELECT order_id, amount, processed_at::timestamptz FROM withdrawals WHERE user_id = $1`

	userID, err := getCtxUserID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.pool.Query(ctx, stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed query withdrawals: %w", err)
	}
	var wList []WithdrawalsEntity
	for rows.Next() {
		var row WithdrawalsEntity

		if err = rows.Scan(&row.OrderID, &row.Amount, &row.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed scan withdrawals row: %w", err)
		}

		wList = append(wList, row)
	}

	return wList, nil
}

func (d *DB) GetOrdersList(ctx context.Context) ([]string, error) {
	const stmt = `SELECT order_id FROM orders WHERE status IN ('NEW', 'PROCESSING')`
	rows, err := d.pool.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed query rows: %w", err)
	}

	oList := make([]string, 0)
	for rows.Next() {
		var row string
		if err = rows.Scan(&row); err != nil {
			return nil, fmt.Errorf("failed scan row: %w", err)
		}
		oList = append(oList, row)
	}
	return oList, nil
}

// UpdateOrder updates order with new status then updates user balance using transaction.
func (d *DB) UpdateOrder(ctx context.Context, data *orders.UpdateOrder) error {
	const (
		updOrdersStmt  = `UPDATE orders SET status = $1, bonus = $2 WHERE order_id = $3 RETURNING user_id`
		updBalanceStmt = `UPDATE balance SET current = current + $1 WHERE user_id = $2`
	)

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: "read committed"})
	if err != nil {
		return fmt.Errorf("failed begin tx: %w", err)
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil {
			d.log.Warn("failed rollback transaction", "txError", err)
		}
	}()

	var userID string
	if err = tx.QueryRow(ctx, updOrdersStmt, data.Status, data.Accrual, data.Number).Scan(&userID); err != nil {
		return fmt.Errorf("failed execute order stmt: %w", err)
	}

	if _, err = tx.Exec(ctx, updBalanceStmt, data.Accrual, userID); err != nil {
		return fmt.Errorf("failed execute balance stmt: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed commit tx: %w", err)
	}

	return nil
}

var (
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrOrderCreatedAlready     = errors.New("order number already created by this user")
	ErrUserExists              = errors.New("user already exists")
	ErrUserNotExists           = errors.New("user not exists")
	ErrAnotherUserOrderCreated = errors.New("order number already created by another user")
)
