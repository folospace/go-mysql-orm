package orm

import (
	"database/sql"
)

//excute raw
func (m Query[T]) Execute() QueryResult {
	if m.prepareSql == "" {
		m.setErr(ErrRawSqlRequired)
	}
	if m.result.Err != nil {
		return m.result
	}

	m.result.PrepareSql = m.prepareSql
	m.result.Bindings = m.bindings

	var res sql.Result
	var err error
	if m.dbTx() != nil {
		res, err = m.dbTx().Exec(m.prepareSql, m.bindings...)
	} else {
		res, err = m.DB().Exec(m.prepareSql, m.bindings...)
	}

	if err != nil {
		m.result.Err = err
	} else if res != nil {
		m.result.LastInsertId, m.result.Err = res.LastInsertId()
		m.result.RowsAffected, m.result.Err = res.RowsAffected()
	}

	return m.result
}
