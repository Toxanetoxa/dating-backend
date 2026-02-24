package pg

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var _ ConnPool = (*poolWrapper)(nil)

// poolWrapper wraps the connection pool for tracing and logging manipulations
// TODO: add tracing and logging to wrappers
type poolWrapper struct {
	p *pgxpool.Pool
	queryerWrapper
}

func wrapPool(pgxPool *pgxpool.Pool) *poolWrapper {
	return &poolWrapper{
		p:              pgxPool,
		queryerWrapper: wrapTraceQueryer(pgxPool),
	}
}

func (w poolWrapper) Begin(ctx context.Context) (Tx, error) {
	return w.p.BeginTx(ctx, pgx.TxOptions{})
}

func (w poolWrapper) BeginTx(ctx context.Context, opts pgx.TxOptions) (Tx, error) {
	return w.p.BeginTx(ctx, opts)
}

func (w *poolWrapper) Close() {
	if w.p != nil {
		w.p.Close()
	}
}

type queryerWrapper struct {
	p Queryer
}

func wrapTraceQueryer(q Queryer) queryerWrapper {
	return queryerWrapper{p: q}
}

func (w queryerWrapper) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return w.p.Exec(ctx, query, args...)
}

func (w queryerWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return w.p.QueryRow(ctx, sql, args...)
}

func (w queryerWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return w.p.Query(ctx, sql, args...)
}

func (w queryerWrapper) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return w.p.SendBatch(ctx, b)
}
