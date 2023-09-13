package orm

import (
    "strings"
)

func (m *Query[T]) SubQuery() SubQuery {
    if m.self != nil {
        cte := m.self
        m.self = nil

        mt := cte.WithRecursiveCte(m.SubQuery(), cte.T.TableName())
        tempTable := mt.generateSelectQuery(mt.columns...)

        tempTable.dbs = mt.DBs()
        tempTable.tx = mt.tx
        tempTable.dbName = mt.tables[0].table.DatabaseName()
        if mt.result.Err != nil {
            tempTable.err = mt.result.Err
        }

        return tempTable
    } else {
        mt := m
        tempTable := mt.generateSelectQuery(mt.columns...)

        tempTable.dbs = mt.DBs()
        tempTable.tx = mt.tx
        tempTable.dbName = mt.tables[0].table.DatabaseName()
        if mt.result.Err != nil {
            tempTable.err = mt.result.Err
        }

        return tempTable
    }
}

func (m *Query[T]) generateSelectQuery(columns ...interface{}) SubQuery {
    var ret SubQuery
    if m.prepareSql != "" {
        ret.raw = m.prepareSql
        ret.bindings = m.bindings
    } else {
        var rawSql string
        bindings := make([]interface{}, 0)

        if len(m.withCtes) > 0 {
            var raws []string
            for _, v := range m.withCtes {
                var raw string
                if v.recursive {
                    raw += "recursive "
                }
                raw += v.tableName
                if len(v.columns) > 0 {
                    raw += "(" + strings.Join(v.columns, ",") + ")"
                }
                raw += " as ("
                raw += v.raw
                raw += ")"
                raws = append(raws, raw)
                bindings = append(bindings, v.bindings...)
            }
            rawSql += "with " + strings.Join(raws, ",\n") + "\n"
        }

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

        var selectKeyword = "select"
        if m.selectTimeout != "" {
            selectKeyword += " " + m.selectTimeout
        }

        rawSql += selectKeyword + " " + selectStr

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

        if m.forUpdate != "" {
            rawSql += " " + string(m.forUpdate)
        }

        if len(m.unions) > 0 {
            for _, v := range m.unions {
                prefix := "\nunion"
                if v.unionAll {
                    prefix += " all"
                }
                prefix += " \n" + v.raw
                rawSql += prefix
                bindings = append(bindings, v.bindings...)
            }
        }

        if len(m.windows) > 0 {
            var raws []string
            for _, v := range m.windows {
                var raw = v.tableName + " as (" + v.raw + ")"
                raws = append(raws, raw)
            }
            rawSql += "\nwindow " + strings.Join(raws, ",\n")
        }

        ret.raw = rawSql
        ret.bindings = bindings
    }
    return ret
}

func (m *Query[T]) generateSelectColumns(columns ...interface{}) (string, error) {
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
