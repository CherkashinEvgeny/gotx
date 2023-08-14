package sql

import (
	"context"
	"github.com/go-pg/pg/v10"

	"github.com/CherkashinEvgeny/gotx/base"
)

type baseTx struct {
	tx        *pg.Tx
	isolation Isolation
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
