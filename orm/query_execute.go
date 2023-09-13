package orm

import (
    "database/sql"
)

//excute raw
func (q *Query[T]) Execute() QueryResult {
    if q.prepareSql == "" {
        q.setErr(ErrRawSqlRequired)
    }
    if q.result.Err != nil {
        return q.result
    }

    q.result.PrepareSql = q.prepareSql
    q.result.Bindings = q.bindings

    var res sql.Result
    var err error
    if q.dbTx() != nil {
        if q.ctx != nil {
            res, err = q.dbTx().ExecContext(*q.ctx, q.prepareSql, q.bindings...)
        } else {
            res, err = q.dbTx().Exec(q.prepareSql, q.bindings...)
        }
    } else {
        if q.ctx != nil {
            res, err = q.DB().ExecContext(*q.ctx, q.prepareSql, q.bindings...)
        } else {
            res, err = q.DB().Exec(q.prepareSql, q.bindings...)
        }
    }

    if err != nil {
        q.result.Err = err
    } else if res != nil {
        q.result.LastInsertId, q.result.Err = res.LastInsertId()
        q.result.RowsAffected, q.result.Err = res.RowsAffected()
    }
    return q.result
}
