package orm

import (
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

    m.prepareSql = rawSql
    m.bindings = bindings

    return m.Execute()
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
            if v.Val == v.Column {
                dotIndex := strings.LastIndex(column, ".")
                temp = column + " = values(" + strings.Trim(column[dotIndex+1:], "`") + ")"
            } else {
                targetColumn, err := m.parseColumn(v.Val)
                if err != nil {
                    m.setErr(err)
                    return ""
                }
                temp = column + " = " + targetColumn
            }
        } else {
            temp = column + " = ?"
            *bindings = append(*bindings, v.Val)
        }
        updateStrs = append(updateStrs, temp)
    }
    return strings.Join(updateStrs, ",")
}
