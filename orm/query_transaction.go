package orm

import (
    "context"
)

func (q *Query[T]) Transaction(f func(query *Query[T]) error) error {
    tx, err := q.DB().BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    q.tx = tx

    err = f(q)

    if err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
