package pgx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	uow "github.com/osvim/unit-of-work"
)

// New returns UnitOfWork constructor
func New(resource Resource, opts ...TxOption) uow.New {
	return func(ctx context.Context) uow.UnitOfWork {
		if value := ctx.Value(txKey{}); value != nil {
			// UnitOfWork context already exists
			return &txCtx{ctx}
		}

		// begin transaction in a lazy way
		begin := func(ctx context.Context) (pgx.Tx, error) {
			txOptions := pgx.TxOptions{}
			for _, opt := range opts {
				opt(&txOptions)
			}
			return resource.BeginTx(ctx, txOptions)
		}

		return &txCtx{context.WithValue(ctx, txKey{}, &txValue{begin: begin})}
	}
}

// Resource is something that begins transactions.
type Resource interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// TxOption makes unit-of-work configuration more user-friendly
type TxOption func(options *pgx.TxOptions)

func WithIsoLevel(level pgx.TxIsoLevel) TxOption {
	return func(options *pgx.TxOptions) {
		options.IsoLevel = level
	}
}

func WithAccessMode(mode pgx.TxAccessMode) TxOption {
	return func(options *pgx.TxOptions) {
		options.AccessMode = mode
	}
}

func WithDeferrableMode(mode pgx.TxDeferrableMode) TxOption {
	return func(options *pgx.TxOptions) {
		options.DeferrableMode = mode
	}
}

type txKey struct{}

type txValue struct {
	tx    pgx.Tx
	err   error
	begin func(context.Context) (pgx.Tx, error)
}

func (v *txValue) beginTxOnce(ctx context.Context) (pgx.Tx, error) {
	if v.tx == nil && v.err == nil {
		v.tx, v.err = v.begin(ctx)
	}
	return v.tx, v.err
}

var _ uow.UnitOfWork = (*txCtx)(nil)

type txCtx struct {
	context.Context
}

func (c *txCtx) Commit() error {
	if tx, ok := txFromCtx(c.Context); ok {
		return tx.Commit(c.Context)
	}
	return nil
}

func (c *txCtx) Rollback() error {
	tx, ok := txFromCtx(c.Context)
	if !ok {
		return nil
	}

	err := tx.Rollback(c.Context)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}

func txFromCtx(ctx context.Context) (tx pgx.Tx, exists bool) {
	var value *txValue
	if value, exists = ctx.Value(txKey{}).(*txValue); exists {
		tx = value.tx
		exists = value.tx != nil
	}
	return
}
