package orm

import "context"

func (m Query[T]) Transaction(query func(db *Query[T]) error) error {
    tx, err := m.DB().BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    m.tx = tx

    err = query(&m)

    if err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
