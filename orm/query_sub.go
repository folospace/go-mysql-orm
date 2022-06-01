package orm

import (
    "strings"
)

func (m Query[T]) SubQuery() SubQuery {
    tempTable := m.generateSelectQuery(m.columns...)

    tempTable.db = m.db
    tempTable.tx = m.tx
    tempTable.dbName = m.tables[0].table.DatabaseName()
    tempTable.err = m.result.Err

    return tempTable
}

func (m Query[T]) generateSelectQuery(columns ...interface{}) SubQuery {
    var ret SubQuery
    if m.prepareSql != "" {
        ret.raw = m.prepareSql
        ret.bindings = m.bindings
    } else {
        bindings := make([]interface{}, 0)

        selectStr, err := m.generateSelectColumns(columns...)
        if err != nil {
            ret.err = err
        }

        tableStr := m.generateTableAndJoinStr(m.tables, &bindings)

        whereStr := m.generateWhereStr(m.wheres, &bindings)

        var groupBy string
        if len(m.groupBy) > 0 {
            groupBy, err = m.generateSelectColumns(m.groupBy...)
            if err != nil {
                ret.err = err
            }
        }
        var having string
        if len(m.having) > 0 {
            having = m.generateWhereStr(m.having, &bindings)
        }

        orderLimitOffsetStr := m.getOrderAndLimitSqlStr()

        rawSql := "select " + selectStr
        if m.forUpdate {
            rawSql += " for update"
        }
        if tableStr != "" {
            rawSql += " from " + tableStr
            if whereStr != "" {
                rawSql += " where " + whereStr
            }
        }

        if groupBy != "" && groupBy != "*" {
            rawSql += " group by " + groupBy
            if having != "" {
                rawSql += " having " + having
            }
        }

        if orderLimitOffsetStr != "" {
            rawSql += " " + orderLimitOffsetStr
        }

        ret.raw = rawSql
        ret.bindings = bindings
    }
    return ret
}

func (m Query[T]) generateSelectColumns(columns ...interface{}) (string, error) {
    var outColumns []string
    for _, v := range columns {
        column, err := m.parseColumn(v)

        if err != nil {
            return "", err
        }
        outColumns = append(outColumns, column) //column string name
    }

    if len(outColumns) == 0 {
        return "*", nil
    } else {
        return strings.Join(outColumns, ","), nil
    }
}
