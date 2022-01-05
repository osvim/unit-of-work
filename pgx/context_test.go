package pgx_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	uow "github.com/osvim/unit-of-work"
	pg "github.com/osvim/unit-of-work/pgx"
)

func TestQuerier(t *testing.T) {
	res, log, _ := setup()
	ctx := context.TODO()
	pgQuerier, _ := pg.Querier(ctx, res)
	_, _ = pgQuerier.Exec(ctx, "SELECT 1")
	assertQueries(t, log, "SELECT 1")
}

func TestTxCtx_Commit(t *testing.T) {
	res, log, err := setup()
	// create unit of work context twice,
	// but begin unit-of-work once
	unitOfWork := pg.New(res,
		pg.WithIsoLevel(pgx.Serializable),
		pg.WithDeferrableMode(pgx.Deferrable),
		pg.WithAccessMode(pgx.ReadWrite))(context.TODO())
	unitOfWork = pg.New(res)(unitOfWork)
	pgQuerier, _ := pg.Querier(unitOfWork, res)
	_, _ = pgQuerier.Exec(unitOfWork, "SELECT 1")
	uow.Flush(unitOfWork, &err)
	assertQueries(t, log, "BEGIN\nSELECT 1\nCOMMIT")
}

func TestTxCtx_Commit_Lazy(t *testing.T) {
	res, log, err := setup()
	unitOfWork := pg.New(res)(context.TODO())
	uow.Flush(unitOfWork, &err)
	// no queries = no begin
	assertQueries(t, log, "")
}

func TestTxCtx_Rollback(t *testing.T) {
	res, log, err := setup()
	unitOfWork := pg.New(res)(context.TODO())
	pgQuerier, _ := pg.Querier(unitOfWork, res)
	_, _ = pgQuerier.Exec(unitOfWork, "SELECT 1")
	err = errors.New("fail")
	uow.Flush(unitOfWork, &err)
	if err == nil {
		t.Errorf("unexpected nil error")
	}
	assertQueries(t, log, "BEGIN\nSELECT 1\nROLLBACK")
}

func TestTxCtx_Rollback_Lazy(t *testing.T) {
	res, log, err := setup()
	unitOfWork := pg.New(res)(context.TODO())
	err = errors.New("fail")
	uow.Flush(unitOfWork, &err)
	if err == nil {
		t.Errorf("unexpected nil error")
	}
	// no queries = no begin
	assertQueries(t, log, "")
}

func setup() (res *resource, log *logger, err error) {
	log = new(logger)
	res = &resource{&querier{log}}
	return
}

type resource struct{ *querier }

func (r *resource) BeginTx(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	r.log("BEGIN")
	return &tx{r.querier}, nil
}

type querier struct{ *logger }

func (r *querier) Exec(_ context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	r.log(sql)
	return
}

func (r *querier) Query(_ context.Context, sql string, args ...interface{}) (rows pgx.Rows, err error) {
	r.log(sql)
	return
}

func (r *querier) QueryRow(ctx context.Context, sql string, args ...interface{}) (row pgx.Row) {
	r.log(sql)
	return
}

type tx struct{ *querier }

func (m *tx) Commit(ctx context.Context) (err error) {
	m.log("COMMIT")
	return
}

func (m *tx) Rollback(ctx context.Context) (err error) {
	m.log("ROLLBACK")
	return
}

func (m *tx) Begin(_ context.Context) (tx pgx.Tx, err error) { return }

func (m *tx) BeginFunc(_ context.Context, _ func(pgx.Tx) error) (err error)  { return }
func (m *tx) Conn() (conn *pgx.Conn)                                         { return }
func (m *tx) LargeObjects() (o pgx.LargeObjects)                             { return }
func (m *tx) SendBatch(_ context.Context, _ *pgx.Batch) (r pgx.BatchResults) { return }
func (m *tx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (i int64, err error) {
	return
}
func (m *tx) Prepare(_ context.Context, _, _ string) (desc *pgconn.StatementDescription, err error) {
	return
}
func (m *tx) QueryFunc(
	_ context.Context,
	_ string,
	_ []interface{},
	_ []interface{},
	_ func(pgx.QueryFuncRow) error,
) (tag pgconn.CommandTag, err error) {
	return
}

type logger struct{ queries []string }

func (l *logger) log(query string) { l.queries = append(l.queries, query) }

func assertQueries(t *testing.T, l *logger, expected string) {
	got := strings.Join(l.queries, "\n")
	if got != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s\n", expected, got)
	}
}
