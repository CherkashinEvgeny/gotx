package sql

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"io"
)

type Executor interface {
	Prepare(q string) (*pg.Stmt, error)

	Exec(query interface{}, params ...interface{}) (res pg.Result, err error)
	ExecContext(c context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	ExecOne(query interface{}, params ...interface{}) (pg.Result, error)
	ExecOneContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error)

	Query(model interface{}, query interface{}, params ...interface{}) (res pg.Result, err error)
	QueryContext(c context.Context, model interface{}, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOne(model interface{}, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOneContext(ctx context.Context, model interface{}, query interface{}, params ...interface{}) (pg.Result, error)

	CopyTo(w io.Writer, query interface{}, params ...interface{}) (res pg.Result, err error)
	CopyFrom(r io.Reader, query interface{}, params ...interface{}) (res pg.Result, err error)

	Model(model ...interface{}) *pg.Query
	ModelContext(c context.Context, model ...interface{}) *pg.Query

	Formatter() orm.QueryFormatter
}

var _ Executor = (*pg.DB)(nil)

var _ Executor = (*pg.Tx)(nil)
