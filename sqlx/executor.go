package sql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Executor interface {
	Prepare(query string) (stmt *sql.Stmt, err error)
	PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
	Preparex(query string) (stmt *sqlx.Stmt, err error)
	PreparexContext(ctx context.Context, query string) (stmt *sqlx.Stmt, err error)
	PrepareNamed(query string) (stmt *sqlx.NamedStmt, err error)
	PrepareNamedContext(ctx context.Context, query string) (stmt *sqlx.NamedStmt, err error)

	Exec(query string, args ...any) (result sql.Result, err error)
	ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error)
	NamedExec(query string, arg interface{}) (result sql.Result, err error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (result sql.Result, err error)
	MustExec(query string, args ...interface{}) (result sql.Result)
	MustExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result)

	Query(query string, args ...any) (rows *sql.Rows, err error)
	QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error)
	Queryx(query string, args ...interface{}) (rows *sqlx.Rows, err error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (rows *sqlx.Rows, err error)
	NamedQuery(query string, arg interface{}) (rows *sqlx.Rows, err error)
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (rows *sqlx.Rows, err error)

	QueryRow(query string, args ...any) (row *sql.Row)
	QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row)
	QueryRowx(query string, args ...interface{}) (row *sqlx.Row)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) (row *sqlx.Row)

	Get(dest interface{}, query string, args ...interface{}) (err error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error)
	Select(dest interface{}, query string, args ...interface{}) (err error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error)
}

var _ Executor = (*sqlx.DB)(nil)

var _ Executor = (*Tx)(nil)
