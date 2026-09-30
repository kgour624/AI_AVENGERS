// Package storage — Tx handling (Ultimate Go §7).
// App begins tx, puts Tx in context, Business substitutes pool. If no Tx in ctx, fail — don't fallback.
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// txKey is context key for Tx (unexported to avoid collisions).
type txKey struct{}

// Tx wraps pgx.Tx for Business use.
type Tx interface {
	Exec(ctx context.Context, sql string, args ...any) (int64, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// pgxTxAdapter adapts pgx.Tx to Tx.
type pgxTxAdapter struct{ tx pgx.Tx }

func (a pgxTxAdapter) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tag, err := a.tx.Exec(ctx, sql, args...)
	return tag.RowsAffected(), err
}
func (a pgxTxAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return a.tx.QueryRow(ctx, sql, args...)
}
func (a pgxTxAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return a.tx.Query(ctx, sql, args...)
}

// TxManager begins transactions at App layer ( §7 — initiated at App).
type TxManager struct{ pool *pgxpool.Pool }

func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

// Begin starts tx and returns new context with Tx. Caller must Commit or Rollback.
func (m *TxManager) Begin(ctx context.Context) (context.Context, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("begin tx: %w", err)
	}
	return context.WithValue(ctx, txKey{}, pgxTxAdapter{tx: tx}), nil
}

// GetTx extracts Tx from context. If missing, returns error — don't fallback to pool ( §7).
func GetTx(ctx context.Context) (Tx, error) {
	v := ctx.Value(txKey{})
	if v == nil {
		return nil, fmt.Errorf("no tx in context — fail closed")
	}
	tx, ok := v.(Tx)
	if !ok {
		return nil, fmt.Errorf("invalid tx type")
	}
	return tx, nil
}

// Commit commits Tx in context.
func Commit(ctx context.Context) error {
	v := ctx.Value(txKey{})
	if v == nil {
		return fmt.Errorf("no tx to commit")
	}
	// Need underlying pgx.Tx to commit — store raw Tx as well
	// For simplicity, we store pgx.Tx directly via private wrapper
	if adapter, ok := v.(pgxTxAdapter); ok {
		return adapter.tx.Commit(ctx)
	}
	return fmt.Errorf("cannot commit — unknown tx")
}

// Rollback rollbacks Tx in context.
func Rollback(ctx context.Context) error {
	v := ctx.Value(txKey{})
	if v == nil {
		return nil
	}
	if adapter, ok := v.(pgxTxAdapter); ok {
		return adapter.tx.Rollback(ctx)
	}
	return nil
}