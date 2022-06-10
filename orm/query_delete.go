package orm

import (
	"database/sql"
	"errors"
)

func (m Query[T]) Delete(primaryValues ...interface{}) QueryResult {
	if len(m.tables) == 0 {
		m.setErr(errors.New("delete table not selected"))
		return m.result
	}
	if len(primaryValues) > 0 {
		return m.Where(m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, primaryValues).delete()
	} else {
		return m.delete()
	}
}

func (m Query[T]) delete() QueryResult {
	bindings := make([]interface{}, 0)

	tableStr := m.generateTableAndJoinStr(m.tables, &bindings)

	whereStr := m.generateWhereStr(m.wheres, &bindings)

	orderLimitOffsetStr := m.getOrderAndLimitSqlStr()

	rawSql := "delete"
	if orderLimitOffsetStr == "" {
		rawSql += " " + m.tables[0].getTableName()
	}
	rawSql += " from " + tableStr

	if whereStr != "" {
		rawSql += " where " + whereStr
	}
	if orderLimitOffsetStr != "" {
		rawSql += " " + orderLimitOffsetStr
	}

	m.result.PrepareSql = rawSql
	m.result.Bindings = bindings

	if m.result.Err != nil {
		return m.result
	}

	var res sql.Result
	var err error
	if m.dbTx() != nil {
		res, err = m.dbTx().Exec(rawSql, bindings...)
	} else {
		res, err = m.DB().Exec(rawSql, bindings...)
	}

	if err != nil {
		m.result.Err = err
	} else if res != nil {
		m.result.LastInsertId, m.result.Err = res.LastInsertId()
		m.result.RowsAffected, m.result.Err = res.RowsAffected()
	}

	return m.result
}
