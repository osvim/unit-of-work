package unit_of_work

import (
	"context"
	"fmt"
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
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if *err != nil {
		e := ctx.Rollback()
		if e != nil && options.logRollbackError != nil {
			if options.wrapRollbackError != "" {
				e = fmt.Errorf("%s: %w", options.wrapRollbackError, e)
			}
			options.logRollbackError(e)
		}
		return
	}

	*err = ctx.Commit()
	if options.wrapCommitError != "" {
		*err = fmt.Errorf("%s: %w", options.wrapCommitError, *err)
	}
}

// Option configures Flush
type Option func(*Options)

// WithRollbackErrorLogger logs error on rollback failure.
func WithRollbackErrorLogger(log func(err error)) Option {
	return func(options *Options) {
		options.logRollbackError = log
	}
}

// WithRollbackErrorWrap wraps error on rollback failure.
// It makes sense to use this option along with rollback logger.
func WithRollbackErrorWrap(msg string) Option {
	return func(options *Options) {
		options.wrapRollbackError = msg
	}
}

// WithCommitErrorWrap wraps error on commit failure.
func WithCommitErrorWrap(msg string) Option {
	return func(options *Options) {
		options.wrapCommitError = msg
	}
}

type Options struct {
	logRollbackError  func(err error)
	wrapRollbackError string
	wrapCommitError   string
}
