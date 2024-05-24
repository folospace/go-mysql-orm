package orm

func (q *Query[T]) Delete(primaryIds ...any) QueryResult {
    if len(q.tables) == 0 {
        q.setErr(ErrTableNotSelected)
        return q.result
    }

    if len(primaryIds) == 1 {
        return q.WherePrimary(primaryIds[0]).delete()
    } else {
        return q.WherePrimary(primaryIds).delete()
    }
}

func (q *Query[T]) delete() QueryResult {
    bindings := make([]any, 0)

    if len(q.wheres) == 0 && len(q.tables) <= 1 && q.limit == 0 {
        q.setErr(ErrDeleteWithoutCondition)
    }

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
