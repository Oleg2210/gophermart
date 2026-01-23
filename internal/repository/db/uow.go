package db

import (
	"context"
	"database/sql"
	"fmt"

	domainrepository "github.com/Oleg2210/gophermart/internal/domain/domain_repository"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func (m *PgxTxManager) WithTx(ctx context.Context, fn func(tx domainrepository.Tx) error) error {
	sqlTx, err := m.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	tx := &PgxTx{tx: sqlTx}

	if err := fn(tx); err != nil {
		_ = sqlTx.Rollback()
		return err
	}

	return sqlTx.Commit()
}

type PgxTx struct {
	tx *sql.Tx
}

func (tx *PgxTx) User() domainrepository.UserRepository         { return &PgxUserRepository{tx} }
func (tx *PgxTx) Order() domainrepository.OrderRepository       { return &PgxOrderRepository{tx} }
func (tx *PgxTx) Withdraw() domainrepository.WithdrawRepository { return &PgxWithdrawRepository{tx} }
