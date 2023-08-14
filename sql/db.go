package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/CherkashinEvgeny/gotx/base"
)

type baseDb struct {
	db *sql.DB
}

func (d *baseDb) Executor() (executor any) {
	return d.db
}

func (d *baseDb) Tx(ctx context.Context, tx base.Tx, options base.Options) (newTx base.Tx, err error) {
	err = d.checkIsolation(tx, options)
	if err != nil {
		return nil, err
	}
	return d.propagation(options)(ctx, tx, options)
}

func (d *baseDb) checkIsolation(oldTx base.Tx, options base.Options) (err error) {
	var txLevel sql.IsolationLevel
	tx := oldTx
	for tx != nil {
		base, ok := tx.(baseTx)
		if ok {
			txLevel = base.options.Isolation
			break
		}
		tx = tx.Parent()
	}
	level := getIsolationLevel(options)
	if txLevel < level {
		if txLevel == sql.LevelDefault {
			// sql.LevelDefault is implementation specific, so skip it
			return nil
		}
		return fmt.Errorf("isolation level %s is to low to handle transaction", txLevel)
	}
	return nil
}

func (d *baseDb) propagation(options []base.Option) (factory func(ctx context.Context, tx base.Tx, options base.Options) (newTx base.Tx, err error)) {
	switch getPropagation(options) {
	case Never:
		return d.never
	case Supports:
		return d.supports
	case Required:
		return d.required
	case Nested:
		return d.nested
	case Mandatory:
		return d.mandatory
	default:
		panic("illegal propagation")
	}
}

func (d *baseDb) never(_ context.Context, tx base.Tx, _ base.Options) (newTx base.Tx, err error) {
	if tx != nil {
		return nil, TransactionMissingError
	}
	return nil, nil
}

var TransactionMissingError = errors.New("transaction is missing")

func (d *baseDb) supports(_ context.Context, tx base.Tx, _ base.Options) (newTx base.Tx, err error) {
	return tx, nil
}

func (d *baseDb) required(ctx context.Context, tx base.Tx, options base.Options) (newTx base.Tx, err error) {
	if tx == nil {
		return d.tx(ctx, options)
	}
	return nop(tx), nil
}

func (d *baseDb) nested(ctx context.Context, oldTx base.Tx, options base.Options) (newTx base.Tx, err error) {
	if oldTx == nil {
		return d.tx(ctx, options)
	}
	id := 0
	tx := oldTx
	for tx != nil {
		nested, ok := tx.(nestedPropagationTx)
		if ok {
			id = nested.id + 1
			break
		}
		tx = tx.Parent()
	}
	_, err = oldTx.Executor().(*sql.Tx).ExecContext(ctx, "SAVEPOINT $1", id)
	if err != nil {
		return nil, err
	}
	return nestedPropagationTx{tx, id}, nil
}

type nestedPropagationTx struct {
	parent base.Tx
	id     int
}

func (n nestedPropagationTx) Executor() (executor any) {
	return n.parent.Executor()
}

func (n nestedPropagationTx) Parent() base.Tx {
	return n.parent
}

func (n nestedPropagationTx) Commit(ctx context.Context) (err error) {
	_, err = n.Executor().(*sql.Tx).ExecContext(ctx, "RELEASE SAVEPOINT $1", n.id)
	return err
}

func (n nestedPropagationTx) Rollback(ctx context.Context) (err error) {
	_, err = n.Executor().(*sql.Tx).ExecContext(ctx, "ROLLBACK TO SAVEPOINT $1", n.id)
	return err
}

func (d *baseDb) mandatory(_ context.Context, tx base.Tx, _ base.Options) (newTx base.Tx, err error) {
	if tx == nil {
		return nil, TransactionRequiredError
	}
	return nop(tx), nil
}

var TransactionRequiredError = errors.New("transaction is required")

func (d *baseDb) tx(ctx context.Context, options base.Options) (tx base.Tx, err error) {
	sqlOptions := &sql.TxOptions{
		Isolation: getIsolationLevel(options),
		ReadOnly:  getReadonly(options),
	}
	sqlTx, err := d.db.BeginTx(ctx, sqlOptions)
	if err != nil {
		return nil, err
	}
	return baseTx{sqlTx, sqlOptions}, nil
}
