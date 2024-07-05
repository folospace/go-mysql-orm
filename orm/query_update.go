package orm

import (
    "errors"
    "reflect"
    "strings"
)

func (q *Query[T]) Update(column any, val any, columnVars ...any) QueryResult {

    var updates = []updateColumn{{col: column, val: val}}

    if len(columnVars) > 0 {
        for k := range columnVars {
            if k%2 != 0 {
                updates = append(updates, updateColumn{col: columnVars[k-1], val: columnVars[k]})
            }
        }
    }

    return q.updates(updates...)
}

func (q *Query[T]) Updates(columnVars map[any]any) QueryResult {
    var updates = make([]updateColumn, len(columnVars))
    var i = 0
    for k := range columnVars {
        updates[i] = updateColumn{col: k, val: columnVars[k]}
        i++
    }
    return q.updates(updates...)
}

func (q *Query[T]) genUpdates(vals ...any) ([]updateColumn, error) {
    if len(vals)%2 != 0 {
        return nil, errors.New("update column and val should be pairs")
    }

    var updates []updateColumn

    for k := range vals {
        if k%2 != 0 {
            continue
        }
        updates = append(updates, updateColumn{col: vals[k], val: vals[k+1]})
    }

    return updates, nil
}

func (q *Query[T]) updates(updates ...updateColumn) QueryResult {
    bindings := make([]any, 0)

    if len(q.wheres) == 0 && len(q.tables) <= 1 && q.limit == 0 {
        q.setErr(ErrUpdateWithoutCondition)
    }

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

func (q *Query[T]) generateUpdateStr(updates []updateColumn, bindings *[]any) string {
    var updateStrs []string
    for _, v := range updates {
        var temp string
        column, err := q.parseColumn(v.col)
        if err != nil {
            q.setErr(err)
            return ""
        }

        val, ok := q.isRaw(v.val)
        if ok {
            temp = column + " = " + val
        } else if reflect.ValueOf(v.val).Kind() == reflect.Ptr {
            if v.val == v.col {
                dotIndex := strings.LastIndex(column, ".")
                temp = column + " = values(`" + strings.Trim(column[dotIndex+1:], "`") + "`)"
            } else {
                targetColumn, err := q.parseColumn(v.val)
                if err == nil {
                    temp = column + " = " + targetColumn
                } else {
                    //q.setErr(err)
                    //return ""
                    temp = column + " = ?"
                    *bindings = append(*bindings, v.val)
                }
            }
        } else {
            temp = column + " = ?"
            *bindings = append(*bindings, v.val)
        }
        updateStrs = append(updateStrs, temp)
    }
    return strings.Join(updateStrs, ",")
}
