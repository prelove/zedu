// Package repository provides shared, HTTP-agnostic database execution
// abstractions used by M2 application services. It defines the minimal
// Executor/Tx/DB interfaces so that repository code can run against either a
// *sql.DB or a *sql.Tx within the same service transaction, and so that tests
// can wrap transactions to inject deterministic failures.
//
// Application services are the only multi-table transaction boundary: they
// call DB.BeginTx to obtain a Tx, pass that Tx to every repository write and
// to InsertAuditLog, then Commit or Rollback. Repository code never depends
// on net/http and never produces HTTP responses.
package repository

import (
	"context"
	"database/sql"
)

// Executor is the minimal read/write interface satisfied by both *sql.DB and
// *sql.Tx. Repository methods accept Executor so the same code path runs
// inside or outside a transaction.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Tx is the minimal transaction interface used by application services.
// *sql.Tx satisfies this interface directly; test wrappers may implement it to
// inject failures on specific query types within a transaction.
type Tx interface {
	Executor
	Commit() error
	Rollback() error
}

// DB is the minimal database interface used by application services. *sql.DB
// does NOT directly satisfy this interface (BeginTx returns *sql.Tx, not Tx);
// use NewDB which wraps *sql.DB in a dbAdapter. Test wrappers that implement DB
// directly are also accepted for fault injection.
type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Executor
}

// dbAdapter wraps *sql.DB so its BeginTx returns the Tx interface.
type dbAdapter struct {
	*sql.DB
}

// NewDB wraps a *sql.DB so it satisfies the DB interface. Pass the returned DB
// to application services; pass the original *sql.DB to HTTP auth middleware.
func NewDB(db *sql.DB) DB {
	return &dbAdapter{db}
}

func (a *dbAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// AsDB accepts either *sql.DB or a ready DB and returns a DB. This lets
// handlers accept a concrete *sql.DB in production while tests can inject a
// custom DB wrapper for fault injection.
func AsDB(v any) DB {
	switch d := v.(type) {
	case DB:
		return d
	case *sql.DB:
		return NewDB(d)
	default:
		panic("repository.AsDB: unsupported db type")
	}
}
