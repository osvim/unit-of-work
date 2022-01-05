package unit_of_work_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	uow "github.com/osvim/unit-of-work"
)

func TestFlush(t *testing.T) {
	tests := map[string]struct {
		rollbackLogger  *rollbackLogger
		flushed         error
		err             error
		commitErr       error
		rollbackErr     error
		commitErrWrap   string
		rollbackErrWrap string
	}{
		"no error": {},
		"commit error": {
			commitErr:     errors.New("failed to commit unit-of-work"),
			commitErrWrap: "failed to save user",
			flushed:       errors.New("failed to save user: failed to commit unit-of-work"),
		},
		"rollback no error": {
			err:     errors.New("failed to save user"),
			flushed: errors.New("failed to save user"),
		},
		"rollback error": {
			err:             errors.New("failed to save user"),
			rollbackErr:     errors.New("failed to rollback unit-of-work"),
			rollbackErrWrap: "failed to save user",
			rollbackLogger:  &rollbackLogger{},
			flushed:         errors.New("failed to save user"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var opts []uow.Option
			if test.commitErrWrap != "" {
				opts = append(opts, uow.WithCommitErrorWrap(test.commitErrWrap))
			}
			if test.rollbackErrWrap != "" {
				opts = append(opts, uow.WithRollbackErrorWrap(test.rollbackErrWrap))
			}
			if test.rollbackLogger != nil {
				opts = append(opts, uow.WithRollbackErrorLogger(test.rollbackLogger.log))
			}

			ctx := &nilUnitOfWork{Context: context.TODO(), commit: test.commitErr, rollback: test.rollbackErr}
			uow.Flush(ctx, &test.err, opts...)

			if test.flushed == nil {
				if test.err != nil {
					t.Errorf("expected nil error, got: '%s'", test.err)
				}
			} else {
				if test.err == nil {
					t.Errorf("nil error, expected: '%s'", test.flushed)
				}

				if test.flushed.Error() != test.err.Error() {
					t.Errorf("expected error: '%s', got: '%s'", test.flushed, test.err)
				}
			}

			if test.rollbackLogger != nil && test.rollbackErr != nil {
				expected := test.rollbackErr
				if test.rollbackErrWrap != "" {
					expected = fmt.Errorf("%s: %w", test.rollbackErrWrap, test.rollbackErr)
				}

				if test.rollbackLogger.err == nil {
					t.Error("expected to log rollback error")
				} else {
					logged := test.rollbackLogger.err.Error()
					if logged != expected.Error() {
						t.Errorf("expected to log: '%s', got: '%s'", expected, logged)
					}
				}
			}
		})
	}
}

type nilUnitOfWork struct {
	context.Context
	commit   error
	rollback error
}

func (w *nilUnitOfWork) Commit() error {
	return w.commit
}

func (w *nilUnitOfWork) Rollback() error {
	return w.rollback
}

type rollbackLogger struct {
	err error
}

func (l *rollbackLogger) log(err error) {
	l.err = err
}
