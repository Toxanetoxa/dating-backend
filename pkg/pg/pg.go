// Package pg implements postgres connection.
package pg

import (
	"context"
	"fmt"
	"log"
	"time"

	"log/slog"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres - contains connection parameters and connection pool.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Pool ConnPool
}

// New - creates new connection.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	var pool *pgxpool.Pool
	for pg.connAttempts > 0 {
		pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}
	pg.Pool = wrapPool(pool)

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close - closes connection.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func NewSqlTransaction(ctx context.Context, db ConnPool, l *slog.Logger) (result *sqlTransaction, err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	result = &sqlTransaction{tx: tx, l: l}

	return
}

func GetWithTx(ctx context.Context, db ConnPool, withTx Transaction, l *slog.Logger) (tx Transaction, err error) {
	if withTx == nil || !withTx.IsInitialized() {
		tx, err = NewSqlTransaction(ctx, db, l)
		if err != nil {
			return
		}
	} else {
		tx = withTx.WithLogger(l)
	}

	return
}

func EmptyTransaction() Transaction {
	return &sqlTransaction{}
}
