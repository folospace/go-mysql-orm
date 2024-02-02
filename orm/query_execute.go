package orm

import (
    "database/sql"
)

//excute raw
func (q *Query[T]) Execute() QueryResult {
    if q.prepareSql == "" {
        q.setErr(ErrRawSqlRequired)
    }

    q.result.PrepareSql = q.prepareSql
    q.result.Bindings = q.bindings

    if q.result.Err != nil {
        if errorLogger != nil {
            errorLogger.Error(q.result.Sql(), q.result.Error())
        }
        return q.result
    } else if infoLogger != nil {
        infoLogger.Info(q.result.Sql(), q.result.Error())
    }

    var res sql.Result
    var err error
    if q.Tx() != nil {
        if q.ctx != nil {
            res, err = q.Tx().ExecContext(*q.ctx, q.prepareSql, q.bindings...)
        } else {
            res, err = q.Tx().Exec(q.prepareSql, q.bindings...)
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
        if errorLogger != nil {
            errorLogger.Error(q.result.Sql(), q.result.Error())
        }
    } else if res != nil {
        q.result.LastInsertId, q.result.Err = res.LastInsertId()
        q.result.RowsAffected, q.result.Err = res.RowsAffected()
    }
    return q.result
}
