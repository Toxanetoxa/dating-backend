package pg

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type ConnPool interface {
	Queryer

	Begin(ctx context.Context) (Tx, error)
	BeginTx(ctx context.Context, opts pgx.TxOptions) (Tx, error)
	Close()
}

type Tx interface {
	Queryer
	TxController
}

type Queryer interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type TxController interface {
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction if this is a real transaction or rolls back to the savepoint if this is a
	// pseudo nested transaction. Rollback will return ErrTxClosed if the Tx is already closed, but is otherwise safe to
	// call multiple times. Hence, a defer tx.Rollback() is safe even if tx.Commit() will be called first in a non-error
	// condition. Any other failure of a real transaction will result in the connection being closed. That includes
	// context.Canceled - i.e. connection is closed even if a context is dead.
	Rollback(ctx context.Context) error
}
