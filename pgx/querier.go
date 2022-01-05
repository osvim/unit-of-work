package pgx

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
)

// Querier returns a pgx.Tx from UnitOfWork context or default pgxtype.Querier (*pgx.Conn or *pgxpool.Pool).
func Querier(ctx context.Context, defaultQuerier pgxtype.Querier) (pgxtype.Querier, error) {
	value, ok := ctx.Value(txKey{}).(*txValue)
	if !ok || value == nil {
		return defaultQuerier, nil
	}
	return value.beginTxOnce(ctx)
}
