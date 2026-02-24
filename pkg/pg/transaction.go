package pg

import (
	"context"
	stderr "errors"
	"fmt"

	"github.com/toxanetoxa/dating-backend/pkg/errors"
	"log/slog"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Transaction interface {
	WithLogger(l *slog.Logger) Transaction

	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Commit(ctx context.Context) error
	MustRollback(ctx context.Context, entailedError error)
	RollbackIfNotDone(ctx context.Context)
	IsInitialized() bool
}

var _ Transaction = (*sqlTransaction)(nil)

type sqlTransaction struct {
	l  *slog.Logger
	tx Tx
}

func (t sqlTransaction) WithLogger(l *slog.Logger) Transaction {
	t.l = l
	return &t
}

func (t *sqlTransaction) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, query, args...)
}

func (t *sqlTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return t.tx.QueryRow(ctx, query, args...)
}

func (t *sqlTransaction) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return t.tx.Query(ctx, query, args...)
}

func (t *sqlTransaction) Commit(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return nil
	}

	if t.tx == nil {
		return errors.New("Nil tx commit")
	}

	return t.tx.Commit(ctx)
}

func (t *sqlTransaction) MustRollback(ctx context.Context, entailedError error) {
	if t == nil || t.tx == nil {
		return
	}

	if err := t.tx.Rollback(ctx); err != nil {
		errMsg := fmt.Sprintf("entailed rollback error: %s, rollback err: %s", entailedError, err.Error())
		err = errors.New(errMsg)
		panic(err.Error())
	}
}

func (t *sqlTransaction) RollbackIfNotDone(ctx context.Context) {
	if t == nil || t.tx == nil {
		return
	}

	err := t.tx.Rollback(ctx)
	if err != nil {
		if stderr.Is(err, pgx.ErrTxClosed) {
			return
		}

		panic(err.Error())
	}
}

func (t sqlTransaction) IsInitialized() bool {
	return t.tx != nil
}
