package sql

import (
	"context"
	"database/sql"

	"github.com/CherkashinEvgeny/gotx/base"
	"github.com/jmoiron/sqlx"
)

type baseTx struct {
	tx      Tx
	options *sql.TxOptions
}

func (t baseTx) Executor() (executor any) {
	return t.tx
}

func (t baseTx) Parent() (db base.Tx) {
	return nil
}

func (t baseTx) Commit(_ context.Context) (err error) {
	return t.tx.Commit()
}

func (t baseTx) Rollback(_ context.Context) (err error) {
	return t.tx.Rollback()
}

func nop(tx base.Tx) (newTx base.Nop) {
	return base.Nop{Tx: tx}
}

type Tx struct {
	*sqlx.Tx
}

func (t Tx) NamedQueryContext(ctx context.Context, query string, arg interface{}) (rows *sqlx.Rows, err error) {
	return sqlx.NamedQueryContext(ctx, t.Tx, query, arg)
}
