package orm

import "database/sql"

func (m *Query) ExecuteRaw(prepareSql string, bindings ...interface{}) QueryResult {
    m.result.PrepareSql = prepareSql
    m.result.Bindings = bindings

    var res sql.Result
    var err error
    if m.dbTx() != nil {
        res, err = m.dbTx().Exec(prepareSql, bindings...)
    } else {
        res, err = m.DB().Exec(prepareSql, bindings...)
    }

    if err != nil {
        m.result.Err = err
    } else if res != nil {
        m.result.LastInsertId, m.result.Err = res.LastInsertId()
        m.result.RowsAffected, m.result.Err = res.RowsAffected()
    }

    return m.result
}