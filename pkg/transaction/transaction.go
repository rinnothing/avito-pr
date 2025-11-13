package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

type Transactor interface {
	DoAtomically(context.Context, func(context.Context) error) error
}

var _ Transactor = &impl{}

type impl struct {
	db *pgxpool.Pool
}

func NewTransactor(db *pgxpool.Pool) *impl {
	return &impl{db: db}
}

func (t *impl) DoAtomically(ctx context.Context, f func(context.Context) error) (txError error) {
	txCtx, tx, err := injectTx(ctx, t.db)
	if err != nil {
		return fmt.Errorf("cannot inject transaction: %w", err)
	}

	defer func() {
		if txError != nil {
			if err = tx.Rollback(txCtx); err != nil {
				logger.ErrorCtx(ctx, "cannot rollback transaction", zap.Error(err))
			}
			return
		}

		if insErr := tx.Commit(txCtx); insErr != nil {
			logger.ErrorCtx(ctx, "cannot commit transaction", zap.Error(insErr))
		}
	}()

	err = f(txCtx)
	if err != nil {
		return err
	}

	return nil
}

type keyType struct{}

var ErrTxNotFound = errors.New("tx not found in context")

func extractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(keyType{}).(pgx.Tx)
	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}

func injectTx(ctx context.Context, pool *pgxpool.Pool) (context.Context, pgx.Tx, error) {
	if tx, err := extractTx(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, keyType{}, tx), tx, nil
}
