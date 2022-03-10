package unit_of_work

import (
	"context"
)

// New is a UnitOfWork constructor.
type New func(context.Context) UnitOfWork

// UnitOfWork allows the domain to define execution context of repository methods.
type UnitOfWork interface {
	context.Context
	Commit() error
	Rollback() error
}

// Flush commits transaction or rollbacks on error.
func Flush(ctx UnitOfWork, err *error, opts ...Option) {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	if *err != nil {
		e := ctx.Rollback()
		if e != nil && options.rollbackErrorHandlerFunc != nil {
			options.rollbackErrorHandlerFunc(e)
		}
		return
	}

	*err = ctx.Commit()
}

// WithRollbackErrorHandler configures transaction rollback error handler
func WithRollbackErrorHandler(handle func(err error)) Option {
	return func(opts *options) { opts.rollbackErrorHandlerFunc = handle }
}

// Option configures Flush
type Option func(*options)

type options struct {
	rollbackErrorHandlerFunc func(err error)
}
