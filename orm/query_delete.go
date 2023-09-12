package orm

func (m Query[T]) Delete(primaryIds ...interface{}) QueryResult {
    if len(m.tables) == 0 {
        m.setErr(ErrTableNotSelected)
        return m.result
    }
    if len(primaryIds) > 0 {
        return m.WherePrimary(primaryIds).delete()
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

    m.prepareSql = rawSql
    m.bindings = bindings

    return m.Execute()
}
