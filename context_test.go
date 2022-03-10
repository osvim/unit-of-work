package unit_of_work_test

import (
	"context"
	"errors"
	"testing"

	uow "github.com/osvim/unit-of-work"
)

func TestFlush(t *testing.T) {
	tests := map[string]struct {
		rollbackErrHandler *rollbackErrorHandler
		flushed            error
		err                error
		commitErr          error
		rollbackErr        error
	}{
		"no error": {},
		"commit error": {
			commitErr: errors.New("failed to commit unit-of-work"),
			flushed:   errors.New("failed to commit unit-of-work"),
		},
		"rollback no error": {
			err:     errors.New("failed to save user"),
			flushed: errors.New("failed to save user"),
		},
		"rollback error": {
			err:                errors.New("failed to save user"),
			rollbackErr:        errors.New("failed to rollback unit-of-work"),
			rollbackErrHandler: &rollbackErrorHandler{},
			flushed:            errors.New("failed to save user"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var opts []uow.Option
			if test.rollbackErrHandler != nil {
				opts = append(opts, uow.WithRollbackErrorHandler(test.rollbackErrHandler.handle))
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

			if test.rollbackErrHandler != nil && test.rollbackErr != nil {
				if test.rollbackErr != test.rollbackErrHandler.err {
					t.Errorf("expected rollback error: '%s', got: '%s'", test.rollbackErr,
						test.rollbackErrHandler.err)
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

type rollbackErrorHandler struct {
	err error
}

func (h *rollbackErrorHandler) handle(err error) {
	h.err = err
}
