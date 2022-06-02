package orm

import (
    "context"
    "database/sql"
)

func (m Query[T]) Transaction(q func(tx *sql.Tx) error) error {
    tx, err := m.DB().BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    m.tx = tx

    err = q(tx)

    if err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
