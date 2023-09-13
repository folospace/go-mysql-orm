package orm

func (q *Query[T]) Delete(primaryIds ...interface{}) QueryResult {
    if len(q.tables) == 0 {
        q.setErr(ErrTableNotSelected)
        return q.result
    }
    if len(primaryIds) > 0 {
        return q.WherePrimary(primaryIds).delete()
    } else {
        return q.delete()
    }
}

func (q *Query[T]) delete() QueryResult {
    bindings := make([]interface{}, 0)

    tableStr := q.generateTableAndJoinStr(q.tables, &bindings)

    whereStr := q.generateWhereStr(q.wheres, &bindings)

    orderLimitOffsetStr := q.getOrderAndLimitSqlStr()

    rawSql := "delete"
    if orderLimitOffsetStr == "" {
        rawSql += " " + q.tables[0].getTableName()
    }
    rawSql += " from " + tableStr

    if whereStr != "" {
        rawSql += " where " + whereStr
    }
    if orderLimitOffsetStr != "" {
        rawSql += " " + orderLimitOffsetStr
    }

    q.prepareSql = rawSql
    q.bindings = bindings

    return q.Execute()
}
