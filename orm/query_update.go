package orm

import (
    "reflect"
    "strings"
)

func (q *Query[T]) Update(column interface{}, val interface{}) QueryResult {
    return q.Updates(UpdateColumn{
        Column: column,
        Val:    val,
    })
}

func (q *Query[T]) Updates(updates ...UpdateColumn) QueryResult {
    bindings := make([]interface{}, 0)

    tableStr := q.generateTableAndJoinStr(q.tables, &bindings)

    updateStr := q.generateUpdateStr(updates, &bindings)

    whereStr := q.generateWhereStr(q.wheres, &bindings)

    orderAndLimitStr := q.getOrderAndLimitSqlStr()

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

    q.prepareSql = rawSql
    q.bindings = bindings

    return q.Execute()
}

func (q *Query[T]) generateUpdateStr(updates []UpdateColumn, bindings *[]interface{}) string {
    var updateStrs []string
    for _, v := range updates {
        var temp string
        column, err := q.parseColumn(v.Column)
        if err != nil {
            q.setErr(err)
            return ""
        }

        val, ok := q.isRaw(v.Val)
        if ok {
            temp = column + " = " + val
        } else if reflect.ValueOf(v.Val).Kind() == reflect.Ptr {
            if v.Val == v.Column {
                dotIndex := strings.LastIndex(column, ".")
                temp = column + " = values(" + strings.Trim(column[dotIndex+1:], "`") + ")"
            } else {
                targetColumn, err := q.parseColumn(v.Val)
                if err != nil {
                    q.setErr(err)
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
