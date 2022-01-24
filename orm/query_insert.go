package orm

import (
    "database/sql"
    "errors"
    "reflect"
    "strings"
)

//tableFieldAddrs: allow insert table columns
func (m *Query) Insert(data interface{}, tableFieldAddrs ...interface{}) QueryResult {
    return m.insert(false, data, tableFieldAddrs, nil)
}

//insert ignore ... // on duplicate key update ...
func (m *Query) InsertIgnore(data interface{}, tableFieldAddrs []interface{}, updates ...UpdateColumn) QueryResult {
    return m.insert(true, data, tableFieldAddrs, updates)
}

func (m *Query) gennerateInsertSql(InsertColumns []string, rowCount int) string {
    columnRawStr := ""
    valRawStr := ""
    for _, v := range InsertColumns {
        columnRawStr += "`" + v + "`,"
        valRawStr += "?,"
    }
    columnRawStr = "(" + strings.TrimRight(columnRawStr, ",") + ")"
    valRawStr = "(" + strings.TrimRight(valRawStr, ",") + ")"

    rows := make([]string, rowCount)
    for k := range rows {
        rows[k] = valRawStr
    }

    return columnRawStr + " values " + strings.Join(rows, ",")
}

func (m *Query) getInsertBindings(val reflect.Value, isSlice bool, validFieldIndex map[int]struct{}) []interface{} {
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

func (m *Query) insert(ignore bool, data interface{}, tableFieldAddrs []interface{}, updates []UpdateColumn) QueryResult {
    val := reflect.ValueOf(data)
    var err error
    isSlice := false
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
    } else if val.Kind() != reflect.Struct {
        m.setErr(errors.New("data must be struct or slice of struct"))
    } else {
        structFields, err = getStructFieldNameSlice(data)
        m.setErr(err)
    }

    if m.result.Err != nil {
        return m.result
    }

    var validFieldName = make(map[string]struct{})
    var validFieldIndex = make(map[int]struct{})
    var InsertColumns []string
    var allowFields = make(map[interface{}]struct{})

    for _, v := range tableFieldAddrs {
        allowFields[v] = struct{}{}
    }
    for k, v := range m.tables[0].jsonFields {
        _, ok := allowFields[k]
        if ok || len(allowFields) == 0 {
            validFieldName[v] = struct{}{}
        }
    }

    for k, v := range structFields {
        if _, ok := validFieldName[v]; ok {
            validFieldIndex[k] = struct{}{}
            InsertColumns = append(InsertColumns, v)
        }
    }

    var insertSql = m.gennerateInsertSql(InsertColumns, rowCount)
    var bindings = m.getInsertBindings(val, isSlice, validFieldIndex)

    updateStr := m.generateUpdateStr(updates, &bindings)

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
