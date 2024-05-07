package orm

import (
    "errors"
    "reflect"
    "sort"
    "strings"
)

//insert and set primary for first T
func (q *Query[T]) Insert(data ...T) QueryResult {
    return q.insert(data)
}

func (q *Query[T]) InsertSubquery(data *SubQuery) QueryResult {
    return q.insert(data)
}

func (q *Query[T]) OnConflictUpdate(column any, val any, columnVars ...any) *Query[T] {
    q.insertIgnore = true
    q.conflictUpdates = []updateColumn{{col: column, val: val}}
    if len(columnVars) > 0 {
        for k := range columnVars {
            if k%2 != 0 {
                q.conflictUpdates = append(q.conflictUpdates, updateColumn{col: columnVars[k-1], val: columnVars[k]})
            }
        }
    }
    return q
}

func (q *Query[T]) OnConflictUpdates(columnVars map[any]any) *Query[T] {
    q.insertIgnore = true

    if len(columnVars) > 0 {
        q.conflictUpdates = make([]updateColumn, len(columnVars))
        var i = 0
        for k := range columnVars {
            q.conflictUpdates[i] = updateColumn{col: k, val: columnVars[k]}
            i++
        }
    }

    return q
}

func (q *Query[T]) gennerateInsertSql(InsertColumns []string, rowCount int) string {
    columnRawStr := ""
    valRawStr := ""

    if len(InsertColumns) > 0 {
        for _, v := range InsertColumns {
            columnRawStr += "`" + v + "`,"
            valRawStr += "?,"
        }
        columnRawStr = "(" + strings.TrimRight(columnRawStr, ",") + ")"
        valRawStr = "(" + strings.TrimRight(valRawStr, ",") + ")"
    }

    if rowCount > 0 {
        rows := make([]string, rowCount)
        for k := range rows {
            rows[k] = valRawStr
        }
        return columnRawStr + " values " + strings.Join(rows, ",")
    } else {
        return columnRawStr
    }
}

func (q *Query[T]) getInsertBindings(val reflect.Value, validFieldIndex map[int]struct{}, defaults map[int]any) []any {
    var bindings []any
    for i := 0; i < val.Len(); i++ {
        for k := 0; k < val.Index(i).Elem().NumField(); k++ {
            if _, ok := validFieldIndex[k]; ok {
                if defaults[k] != nil && val.Index(i).Elem().Field(k).IsZero() {
                    bindings = append(bindings, defaults[k])
                } else {
                    bindings = append(bindings, val.Index(i).Elem().Field(k).Interface())
                }
            }
        }
    }
    return bindings
}

func (q *Query[T]) insert(data any) QueryResult {
    var err error
    var acceptFields = q.columns

    var updates = q.conflictUpdates
    var ignore = q.insertIgnore

    val := reflect.ValueOf(data)
    isSubQuery := false
    var subq *SubQuery
    rowCount := 1
    var structFields []string
    var structDefaults map[int]any
    if val.Kind() == reflect.Slice {
        rowCount = val.Len()
        if rowCount == 0 {
            q.setErr(errors.New("slice is empty"))
            return q.result
        }
        if val.Index(0).Type().Elem() != q.tables[0].tableStructType {
            q.setErr(errors.New("slice elem must be T"))
        } else {
            structFields, err = getStructFieldNameSlice(val.Index(0).Elem().Interface())
            q.setErr(err)
            structDefaults, err = getStructFieldWithDefaultTime(val.Index(0).Elem().Interface())
            q.setErr(err)
        }
    } else if val.Kind() == reflect.Ptr {
        sub, ok := data.(*SubQuery)
        if ok {
            isSubQuery = true
            subq = sub
        } else {
            q.setErr(ErrInsertPtrNotAllowed)
        }
    } else {
        q.setErr(errors.New("data must be subquery or slice of T"))
    }

    if q.result.Err != nil {
        if errorLogger != nil {
            errorLogger.Error(q.result.Sql(), q.result.Error())
        }
        return q.result
    } else if infoLogger != nil {
        infoLogger.Info(q.result.Sql(), q.result.Error())
    }

    var validFieldNameMap = make(map[string]int)
    var validFieldNames = make([]string, 0)
    var validFieldIndex = make(map[int]struct{})
    var InsertColumns []string //actually insert columns
    var allowFields = make(map[any]int)

    for k, v := range acceptFields {
        allowFields[v] = k
    }
    for k, v := range q.tables[0].ormFields {
        pos, ok := allowFields[k]
        if ok || (len(allowFields) == 0 && isSubQuery == false) {
            validFieldNameMap[v] = pos
            validFieldNames = append(validFieldNames, v)
        }
    }

    var insertSql, updateStr string
    var bindings []any
    if isSubQuery {
        sort.SliceStable(validFieldNames, func(i, j int) bool {
            return validFieldNameMap[validFieldNames[i]] < validFieldNameMap[validFieldNames[j]]
        })
        insertSql = q.gennerateInsertSql(validFieldNames, 0)
        if insertSql != "" {
            insertSql += " "
        }
        insertSql += subq.raw
        bindings = subq.bindings
    } else {
        for k, v := range structFields {
            if _, ok := validFieldNameMap[v]; ok {
                validFieldIndex[k] = struct{}{}
                InsertColumns = append(InsertColumns, v)
            }
        }
        insertSql = q.gennerateInsertSql(InsertColumns, rowCount)
        bindings = q.getInsertBindings(val, validFieldIndex, structDefaults)
    }

    updateStr = q.generateUpdateStr(updates, &bindings)

    rawSql := "insert"
    if ignore {
        rawSql += " ignore"
    }

    rawSql += " into " + q.tables[0].getTableName() + " " + insertSql

    if updateStr != "" {
        rawSql += " on duplicate key update " + updateStr
    }

    rawSql += ";"

    q.prepareSql = rawSql
    q.bindings = bindings

    res := q.Execute()

    //set first element's first field on condition
    if isSubQuery == false && res.Err == nil && res.LastInsertId > 0 && (val.Len() == 1 || q.insertIgnore == false) {
        val.Index(0).Elem().Field(0).Set(reflect.ValueOf(res.LastInsertId).Convert(val.Index(0).Elem().Field(0).Type()))
    }
    return res
}
