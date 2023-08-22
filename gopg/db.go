package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/CherkashinEvgeny/gotx/base"
	"github.com/go-pg/pg/v10"
)

type baseDb struct {
	db *pg.DB
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
	var txLevel Isolation
	tx := oldTx
	for tx != nil {
		base, ok := tx.(baseTx)
		if ok {
			txLevel = base.isolation
			break
		}
		tx = tx.Parent()
	}
	level := getIsolationLevel(options)
	if txLevel < level {
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
	sqlTx, err := d.db.BeginContext(ctx)
	if err != nil {
		return nil, err
	}
	isolation := getIsolationLevel(options)
	if isolation != ReadCommitted {
		err = d.setIsolationLevel(ctx, sqlTx, isolation)
		if err != nil {
			_ = sqlTx.Rollback()
			return nil, err
		}
	}
	return baseTx{sqlTx, isolation}, nil
}

func (d *baseDb) setIsolationLevel(ctx context.Context, tx *pg.Tx, level Isolation) (err error) {
	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", mapIsolation(level)))
	return err
}

func mapIsolation(level Isolation) (str string) {
	switch level {
	case ReadCommitted:
		return "READ COMMITTED"
	case RepeatableRead:
		return "REPEATABLE READ"
	case Serializable:
		return "SERIALIZABLE"
	default:
		panic("unknown isolation level")
	}
}
