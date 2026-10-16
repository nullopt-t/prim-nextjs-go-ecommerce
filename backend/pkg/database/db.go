// Package database provides PostgreSQL connection management, connection pooling,
// transaction abstraction, and query execution interfaces using pgx/v5.
package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)



// QueryExecutor abstracts SQL query execution across both raw connection pools (*DB)
// and active transactional contexts (*Tx). This allows repository methods to execute
// queries seamlessly inside or outside a database transaction.
type QueryExecutor interface {
	// Exec executes a SQL command that does not return rows (e.g. INSERT, UPDATE, DELETE).
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)

	// Query executes a SQL query returning multiple rows (e.g. SELECT).
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)

	// QueryRow executes a SQL query expected to return at most one row.
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}

// DB wraps a pgxpool.Pool instance to manage pooled PostgreSQL connections
// and provide thread-safe database access for concurrent requests.
type DB struct {
	// Pool is the underlying pgx connection pool.
	Pool *pgxpool.Pool
}

// ConnectDB parses configuration and initializes a pgxpool.Pool connection pool.
// It verifies connectivity to the PostgreSQL cluster using the supplied context.
func ConnectDB(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:   net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Path:   cfg.DBName,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	poolConfig, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse database config : %w", err)
	}

	// Configure connection pool sizing and lifecycle tunings to avoid exhaustion and stale connections
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("database pool initialize : %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close gracefully closes all idle and active connections in the pool.
func (db *DB) Close() {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping verifies that a connection to the database is still alive, establishing a connection if necessary.
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Exec executes a SQL statement using a pooled connection without returning rows.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.Pool.Exec(ctx, sql, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// The caller must call Scan on the returned Row to release the connection.
func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns rows, typically a SELECT.
// The caller is responsible for closing the returned Rows to return the connection to the pool.
func (db *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.Pool.Query(ctx, sql, args...)
}

// Tx wraps an active pgx.Tx transaction, implementing QueryExecutor.
type Tx struct {
	tx pgx.Tx
}

// NewTx creates a new Tx wrapper around an active pgx.Tx.
func NewTx(tx pgx.Tx) *Tx {
	return &Tx{tx: tx}
}

// Exec executes a query within the active transaction.
func (t *Tx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}

// QueryRow executes a single-row query within the active transaction.
func (t *Tx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

// Query executes a multi-row query within the active transaction.
func (t *Tx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

// Commit commits the active transaction.
func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the active transaction.
func (t *Tx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

