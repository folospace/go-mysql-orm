package orm

import (
    "database/sql"
    "reflect"
    "strings"
)

func (m Query[T]) Update(column interface{}, val interface{}) QueryResult {
    return m.Updates(UpdateColumn{
        Column: column,
        Val:    val,
    })
}

func (m Query[T]) Updates(updates ...UpdateColumn) QueryResult {
    bindings := make([]interface{}, 0)

    tableStr := m.generateTableAndJoinStr(m.tables, &bindings)

    updateStr := m.generateUpdateStr(updates, &bindings)

    whereStr := m.generateWhereStr(m.wheres, &bindings)

    orderAndLimitStr := m.getOrderAndLimitSqlStr()

    rawSql := "update " + tableStr
    if updateStr != "" {
        rawSql += " set " + updateStr
        if whereStr != "" {
            rawSql += " where " + whereStr
        }
    }

    if orderAndLimitStr != "" {
        rawSql += " " + orderAndLimitStr
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

func (m Query[T]) generateUpdateStr(updates []UpdateColumn, bindings *[]interface{}) string {
    var updateStrs []string
    for _, v := range updates {
        var temp string
        column, err := m.parseColumn(v.Column)
        if err != nil {
            m.setErr(err)
            return ""
        }

        val, ok := m.isRaw(v.Val)
        if ok {
            temp = column + " = " + val
        } else if reflect.ValueOf(v.Val).Kind() == reflect.Ptr {
            targetColumn, err := m.parseColumn(v.Val)
            if err != nil {
                m.setErr(err)
                return ""
            }
            temp = column + " = " + targetColumn
        } else {
            temp = column + " = ?"
            *bindings = append(*bindings, v.Val)
        }
        updateStrs = append(updateStrs, temp)
    }
    return strings.Join(updateStrs, ",")
}
