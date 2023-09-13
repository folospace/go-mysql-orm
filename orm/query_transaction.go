package orm

import (
    "context"
    "database/sql"
)

func (q *Query[T]) Transaction(f func(tx *sql.Tx) error) error {
    tx, err := q.DB().BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    q.tx = tx

    err = f(tx)

    if err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
