package orm

import (
    "strings"
)

func (q *Query[T]) SubQuery() *SubQuery {
    if q.self != nil {
        cte := q.self
        q.self = nil

        mt := cte.WithRecursiveCte(q.SubQuery(), cte.T.TableName())
        tempTable := mt.generateSelectQuery(mt.columns...)

        tempTable.dbs = mt.DBs()
        tempTable.tx = mt.tx
        tempTable.dbName = mt.tables[0].table.DatabaseName()
        if mt.result.Err != nil {
            tempTable.err = mt.result.Err
        }

        return tempTable
    } else {
        mt := q
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

func (q *Query[T]) generateSelectQuery(columns ...interface{}) *SubQuery {
    var ret SubQuery
    if q.prepareSql != "" {
        ret.raw = q.prepareSql
        ret.bindings = q.bindings
    } else {
        var rawSql string
        bindings := make([]interface{}, 0)

        if len(q.withCtes) > 0 {
            var raws []string
            for _, v := range q.withCtes {
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

        selectStr, err := q.generateSelectColumns(columns...)
        if err != nil {
            ret.err = err
        }

        tableStr := q.generateTableAndJoinStr(q.tables, &bindings)

        whereStr := q.generateWhereStr(q.wheres, &bindings)

        var groupBy string
        if len(q.groupBy) > 0 {
            groupBy, err = q.generateSelectColumns(q.groupBy...)
            if err != nil {
                ret.err = err
            }
        }
        var having string
        if len(q.having) > 0 {
            having = q.generateWhereStr(q.having, &bindings)
        }

        orderLimitOffsetStr := q.getOrderAndLimitSqlStr()

        var selectKeyword = "select"
        if q.selectTimeout != "" {
            selectKeyword += " " + q.selectTimeout
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

        if q.forUpdate != "" {
            rawSql += " " + string(q.forUpdate)
        }

        if len(q.unions) > 0 {
            for _, v := range q.unions {
                prefix := "\nunion"
                if v.unionAll {
                    prefix += " all"
                }
                prefix += " \n" + v.raw
                rawSql += prefix
                bindings = append(bindings, v.bindings...)
            }
        }

        if len(q.windows) > 0 {
            var raws []string
            for _, v := range q.windows {
                var raw = v.tableName + " as (" + v.raw + ")"
                raws = append(raws, raw)
            }
            rawSql += "\nwindow " + strings.Join(raws, ",\n")
        }

        ret.raw = rawSql
        ret.bindings = bindings
    }
    return &ret
}

func (q *Query[T]) generateSelectColumns(columns ...interface{}) (string, error) {
    var outColumns []string
    for _, v := range columns {
        column, err := q.parseColumn(v)

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
