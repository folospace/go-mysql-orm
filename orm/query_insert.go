package orm

import (
    "database/sql"
    "errors"
    "reflect"
    "strings"
)

//tableFieldAddrs: allow insert table columns
func (m Query[T]) Insert(data interface{}, tableFieldAddrs ...interface{}) QueryResult {
    return m.insert(false, data, tableFieldAddrs, nil)
}

//insert ignore ... // on duplicate key update ...
func (m Query[T]) InsertIgnore(data interface{}, tableFieldAddrs []interface{}, updates ...UpdateColumn) QueryResult {
    return m.insert(true, data, tableFieldAddrs, updates)
}

func (m Query[T]) gennerateInsertSql(InsertColumns []string, rowCount int) string {
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

func (m Query[T]) getInsertBindings(val reflect.Value, isSlice bool, validFieldIndex map[int]struct{}) []interface{} {
    var bindings []interface{}
    if isSlice == false {
        for i := 0; i < val.NumField(); i++ {
            if _, ok := validFieldIndex[i]; ok {
                bindings = append(bindings, val.Field(i).Interface())
            }
        }
    } else {
        for i := 0; i < val.Len(); i++ {
            for k := 0; k < val.Index(i).NumField(); k++ {
                if _, ok := validFieldIndex[k]; ok {
                    bindings = append(bindings, val.Index(i).Field(k).Interface())
                }
            }
        }
    }
    return bindings
}

func (m Query[T]) insert(ignore bool, data interface{}, tableFieldAddrs []interface{}, updates []UpdateColumn) QueryResult {
    val := reflect.ValueOf(data)
    var err error
    isSlice := false
    isSubQuery := false
    var subq *SubQuery
    rowCount := 1
    var structFields []string
    if val.Kind() == reflect.Slice {
        rowCount = val.Len()
        if rowCount == 0 {
            m.setErr(errors.New("slice is empty"))
            return m.result
        }
        if val.Index(0).Kind() != reflect.Struct {
            m.setErr(errors.New("slice element must be struct"))
        } else {
            isSlice = true
            structFields, err = getStructFieldNameSlice(val.Index(0).Interface())
            m.setErr(err)
        }
    } else if val.Kind() == reflect.Struct {
        structFields, err = getStructFieldNameSlice(data)
        m.setErr(err)
    } else if val.Kind() == reflect.Ptr {
        sub, ok := data.(*SubQuery)
        if ok == false {
            m.setErr(errors.New("data is not a subquery"))
        } else {
            isSubQuery = true
            subq = sub
        }
    } else {
        m.setErr(errors.New("data must be struct or slice of struct"))
    }

    if m.result.Err != nil {
        return m.result
    }

    var validFieldNameMap = make(map[string]struct{})
    var validFieldNames = make([]string, 0)
    var validFieldIndex = make(map[int]struct{})
    var InsertColumns []string //actually insert columns
    var allowFields = make(map[interface{}]struct{})

    for _, v := range tableFieldAddrs {
        allowFields[v] = struct{}{}
    }
    for k, v := range m.tables[0].ormFields {
        _, ok := allowFields[k]
        if ok || (len(allowFields) == 0 && isSubQuery == false) {
            validFieldNameMap[v] = struct{}{}
            validFieldNames = append(validFieldNames, v)
        }
    }

    var insertSql, updateStr string
    var bindings []interface{}
    if isSubQuery && subq != nil {
        insertSql = m.gennerateInsertSql(validFieldNames, 0)
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
        insertSql = m.gennerateInsertSql(InsertColumns, rowCount)
        bindings = m.getInsertBindings(val, isSlice, validFieldIndex)
    }

    updateStr = m.generateUpdateStr(updates, &bindings)

    rawSql := "insert"
    if ignore {
        rawSql += " ignore"
    }

    rawSql += " into " + m.tables[0].getTableName() + " " + insertSql

    if updateStr != "" {
        rawSql += " on duplicate key update " + updateStr
    }

    rawSql += ";"
    m.result.PrepareSql = rawSql
    m.result.Bindings = bindings

    var insertRes sql.Result
    if m.dbTx() != nil {
        insertRes, err = m.dbTx().Exec(rawSql, bindings...)
    } else {
        insertRes, err = m.DB().Exec(rawSql, bindings...)
    }

    m.setErr(err)
    if insertRes != nil {
        m.result.LastInsertId, err = insertRes.LastInsertId()
        m.setErr(err)
        m.result.RowsAffected, err = insertRes.RowsAffected()
        m.setErr(err)
    }
    return m.result
}
