package orm

import (
    "database/sql"
)

//select from raw sql
func (m Query[T]) SelectRaw(dest interface{}, prepareSql string, bindings ...interface{}) QueryResult {
    m.result.PrepareSql = prepareSql
    m.result.Bindings = bindings

    if m.result.Err != nil {
        return m.result
    }

    var rows *sql.Rows
    var err error
    if m.dbTx() != nil {
        rows, err = m.dbTx().Query(prepareSql, bindings...)
    } else {
        rows, err = m.DB().Query(prepareSql, bindings...)
    }

    defer func() {
        if rows != nil {
            _ = rows.Close()
        }
    }()

    if err != nil {
        m.result.Err = err
        return m.result
    }

    m.result.Err = m.scanRows(dest, rows)
    return m.result
}
