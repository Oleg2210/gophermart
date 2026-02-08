package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Oleg2210/gophermart/internal/domain"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	retryCount int           = 3
	retryTime  time.Duration = 50 * time.Millisecond
)

type PgxTxManager struct {
	db *sql.DB
}

func applyMigrations(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrations up: %w", err)
	}

	return nil
}

func NewPgxTxManager(dsn string) (*PgxTxManager, error) {
	err := applyMigrations(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PgxTxManager{db: db}, nil
}

func (m *PgxTxManager) WithTx(ctx context.Context, fn func(tx domain.Tx) error) error {
	return retry(ctx, retryCount, retryTime, func() error {
		sqlTx, err := m.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return err
		}

		tx := &PgxTx{tx: sqlTx}

		if err := fn(tx); err != nil {
			_ = sqlTx.Rollback()
			return err
		}

		err = sqlTx.Commit()
		return err
	})
}

type PgxTx struct {
	tx *sql.Tx
}

func (tx *PgxTx) User() domain.UserRepository         { return &PgxUserRepository{tx} }
func (tx *PgxTx) Order() domain.OrderRepository       { return &PgxOrderRepository{tx} }
func (tx *PgxTx) Withdraw() domain.WithdrawRepository { return &PgxWithdrawRepository{tx} }

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	var err error

	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}
	}

	return err
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001": // serialization_failure
		case "40P01": // deadlock_detected
		case "08006": // connection failure
			return true
		}
	}
	return false
}
