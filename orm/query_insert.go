package orm

import (
    "errors"
    "reflect"
    "sort"
    "strings"
)

//tableFieldPtrs: allow insert table columns
func (q *Query[T]) Insert(data T, acceptFieldPtrs ...interface{}) QueryResult {
    return q.insert(false, []T{data}, acceptFieldPtrs, nil)
}

//tableFieldPtrs: allow insert table columns
func (q *Query[T]) InsertUpdate(data T, updates []UpdateColumn, acceptFieldPtrs ...interface{}) QueryResult {
    return q.insert(true, []T{data}, acceptFieldPtrs, updates)
}

//tableFieldPtrs: allow insert table columns
func (q *Query[T]) Inserts(data []T, acceptFieldPtrs ...interface{}) QueryResult {
    return q.insert(false, data, acceptFieldPtrs, nil)
}

//insert ignore ... // on duplicate key update ...
func (q *Query[T]) InsertUpdates(data []T, updates []UpdateColumn, acceptFieldPtrs ...interface{}) QueryResult {
    return q.insert(true, data, acceptFieldPtrs, updates)
}

func (q *Query[T]) InsertSubquery(data *SubQuery, updates []UpdateColumn, acceptFieldPtrs ...interface{}) QueryResult {
    return q.insert(true, data, acceptFieldPtrs, updates)
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

func (q *Query[T]) getInsertBindings(val reflect.Value, isSlice bool, validFieldIndex map[int]struct{}, defaults map[int]interface{}) []interface{} {
    var bindings []interface{}
    if isSlice == false {
        for i := 0; i < val.NumField(); i++ {
            if _, ok := validFieldIndex[i]; ok {
                if defaults[i] != nil && val.Field(i).IsZero() {
                    bindings = append(bindings, defaults[i])
                } else {
                    bindings = append(bindings, val.Field(i).Interface())
                }
            }
        }
    } else {
        for i := 0; i < val.Len(); i++ {
            for k := 0; k < val.Index(i).NumField(); k++ {
                if _, ok := validFieldIndex[k]; ok {
                    if defaults[k] != nil && val.Index(i).Field(k).IsZero() {
                        bindings = append(bindings, defaults[k])
                    } else {
                        bindings = append(bindings, val.Index(i).Field(k).Interface())
                    }
                }
            }
        }
    }
    return bindings
}

func (q *Query[T]) insert(ignore bool, data interface{}, tableFieldPtrs []interface{}, updates []UpdateColumn) QueryResult {
    val := reflect.ValueOf(data)
    var err error
    isSlice := false
    isSubQuery := false
    var subq *SubQuery
    rowCount := 1
    var structFields []string
    var structDefaults map[int]interface{}
    if val.Kind() == reflect.Slice {
        rowCount = val.Len()
        if rowCount == 0 {
            q.setErr(errors.New("slice is empty"))
            return q.result
        }
        if val.Index(0).Kind() != reflect.Struct {
            q.setErr(errors.New("slice elem must be struct"))
        } else {
            isSlice = true
            structFields, err = getStructFieldNameSlice(val.Index(0).Interface())
            q.setErr(err)
            structDefaults, err = getStructFieldWithDefaultTime(val.Index(0).Interface())
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
    } else if val.Kind() == reflect.Struct {
        structFields, err = getStructFieldNameSlice(data)
        q.setErr(err)
        structDefaults, err = getStructFieldWithDefaultTime(val.Index(0).Interface())
        q.setErr(err)
    } else {
        q.setErr(errors.New("data must be struct or slice of struct"))
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
    var allowFields = make(map[interface{}]int)

    for k, v := range tableFieldPtrs {
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
    var bindings []interface{}
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
        bindings = q.getInsertBindings(val, isSlice, validFieldIndex, structDefaults)
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

    return q.Execute()
}
