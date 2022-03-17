package orm

import (
    "database/sql"
    "errors"
    "reflect"
    "strings"
)

func (m *Query) SelectCount(dest interface{}) QueryResult {
    return m.Select(dest, "count(*)")
}
func (m *Query) SelectValueOfFirstCell(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectSliceOfColumn1(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectStructOfRow1(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectSliceOfStruct(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectMapOfStructKeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectMapOfStructSliceKeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectMapOfColumn2KeyByColumn1(dest interface{}, columns ...interface{}) QueryResult {
    return m.Select(dest, columns...)
}
func (m *Query) SelectSub(columns ...interface{}) *tempTable {
    tempTable := m.generateSelectQuery(columns...)

    tempTable.db = m.db
    tempTable.tx = m.tx
    tempTable.dbName = m.tables[0].table.DatabaseName()
    tempTable.err = m.result.Err

    return &tempTable
}

func (m *Query) SelectForUpdate(dest interface{}, columns ...interface{}) QueryResult {
    m.forUpdate = true
    return m.Select(dest, columns...)
}

func (m *Query) Select(dest interface{}, columns ...interface{}) QueryResult {
    tempTable := m.generateSelectQuery(columns...)

    m.result.PrepareSql = tempTable.raw
    m.result.Bindings = tempTable.bindings

    if m.result.Err != nil {
        return m.result
    }

    var rows *sql.Rows
    var err error
    if m.dbTx() != nil {
        rows, err = m.dbTx().Query(tempTable.raw, tempTable.bindings...)
    } else {
        rows, err = m.DB().Query(tempTable.raw, tempTable.bindings...)
    }

    defer func() {
        if rows != nil {
            _ = rows.Close()
        }
    }()

    if err != nil {
        m.result.Err = err
        return m.result
    }

    m.result.Err = m.scanRows(dest, rows)
    return m.result
}
func (m *Query) generateSelectColumns(columns ...interface{}) string {
    var outColumns []string
    for _, v := range columns {
        column, err := m.parseColumn(v)
        if err != nil {
            m.result.Err = err
            return ""
        }
        outColumns = append(outColumns, column) //column string name
    }

    if len(outColumns) == 0 {
        return "*"
    } else {
        return strings.Join(outColumns, ",")
    }
}

func (m *Query) generateSelectQuery(columns ...interface{}) tempTable {
    bindings := make([]interface{}, 0)

    selectStr := m.generateSelectColumns(columns...)

    tableStr := m.generateTableAndJoinStr(m.tables, &bindings)

    whereStr := m.generateWhereStr(m.wheres, &bindings)

    orderLimitOffsetStr := m.getOrderAndLimitSqlStr()

    rawSql := "select " + selectStr
    if m.forUpdate {
        rawSql += " for update"
    }
    if tableStr != "" {
        rawSql += " from " + tableStr
        if whereStr != "" {
            rawSql += " where " + whereStr
        }
    }

    if orderLimitOffsetStr != "" {
        rawSql += " " + orderLimitOffsetStr
    }

    var ret tempTable
    ret.raw = rawSql
    ret.bindings = bindings
    return ret
}

func (m *Query) scanValues(baseAddrs []interface{}, rowColumns []string, rows *sql.Rows, setVal func(), tryOnce bool) error {
    var err error
    var tempAddrs = make([]interface{}, len(rowColumns))
    for k := range rowColumns {
        var temp interface{}
        tempAddrs[k] = &temp
    }

    finalAddrs := make([]interface{}, len(rowColumns))

    for rows.Next() {
        err = rows.Scan(tempAddrs...)
        if err != nil {
            return err
        }

        for k, v := range tempAddrs {
            if reflect.ValueOf(v).Elem().IsNil() {
                felement := reflect.ValueOf(baseAddrs[k]).Elem()
                felement.Set(reflect.Zero(felement.Type()))
                finalAddrs[k] = v
            } else {
                finalAddrs[k] = baseAddrs[k]
            }
        }

        err = rows.Scan(finalAddrs...)
        if setVal != nil {
            setVal()
        }
        if tryOnce {
            break
        }
    }
    if err == nil {
        err = rows.Err()
    }
    return err
}

func (m *Query) scanRows(dest interface{}, rows *sql.Rows) error {
    rowColumns, err := rows.Columns()
    if err != nil {
        return err
    }
    base := reflect.ValueOf(dest)
    if base.Kind() != reflect.Ptr {
        return errors.New("select dest must be ptr")
    }
    val := base.Elem()
    if val.Kind() == reflect.Ptr {
        return errors.New("select dest must be ptr")
    }

    switch val.Kind() {
    case reflect.Map:
        ele := reflect.TypeOf(dest).Elem().Elem()
        if ele.Kind() == reflect.Ptr {
            return errors.New("select dest slice element must not be ptr")
        }

        newVal := reflect.MakeMap(reflect.TypeOf(dest).Elem())
        switch ele.Kind() {
        case reflect.Struct:
            structAddr := reflect.New(ele).Interface()
            structAddrMap, err := getStructFieldAddrMap(structAddr)
            if err != nil {
                return err
            }
            var baseAddrs = make([]interface{}, len(rowColumns))

            for k, v := range rowColumns {
                baseAddrs[k] = structAddrMap[v]
                if baseAddrs[k] == nil {
                    var temp interface{}
                    baseAddrs[k] = &temp
                }
            }
            err = m.scanValues(baseAddrs, rowColumns, rows, func() {
                newVal.SetMapIndex(reflect.ValueOf(baseAddrs[0]).Elem(), reflect.ValueOf(structAddr).Elem())
            }, false)
            base.Elem().Set(newVal)
        default:
            keyType := reflect.TypeOf(dest).Elem().Key()

            keyAddr := reflect.New(keyType).Interface()
            tempAddr := reflect.New(ele).Interface()

            var baseAddrs = make([]interface{}, len(rowColumns))

            for k := 0; k < len(rowColumns); k++ {
                if k == 0 {
                    baseAddrs[k] = keyAddr
                } else if k == 1 {
                    baseAddrs[k] = tempAddr
                } else {
                    var temp interface{}
                    baseAddrs[k] = &temp
                }
            }
            err = m.scanValues(baseAddrs, rowColumns, rows, func() {
                newVal.SetMapIndex(reflect.ValueOf(keyAddr).Elem(), reflect.ValueOf(tempAddr).Elem())
            }, false)

            base.Elem().Set(newVal)
        }
    case reflect.Struct:
        structAddr := dest
        structAddrMap, err := getStructFieldAddrMap(structAddr)
        if err != nil {
            return err
        }
        var baseAddrs = make([]interface{}, len(rowColumns))

        for k, v := range rowColumns {
            baseAddrs[k] = structAddrMap[v]
            if baseAddrs[k] == nil {
                var temp interface{}
                baseAddrs[k] = &temp
            }
        }
        err = m.scanValues(baseAddrs, rowColumns, rows, nil, true)
    case reflect.Slice:
        ele := reflect.TypeOf(dest).Elem().Elem()
        if ele.Kind() == reflect.Ptr {
            return errors.New("select dest slice element must not be ptr")
        }

        switch ele.Kind() {
        case reflect.Struct:
            structAddr := reflect.New(ele).Interface()
            structAddrMap, err := getStructFieldAddrMap(structAddr)
            if err != nil {
                return err
            }
            var baseAddrs = make([]interface{}, len(rowColumns))

            for k, v := range rowColumns {
                baseAddrs[k] = structAddrMap[v]
                if baseAddrs[k] == nil {
                    var temp interface{}
                    baseAddrs[k] = &temp
                }
            }

            err = m.scanValues(baseAddrs, rowColumns, rows, func() {
                val = reflect.Append(val, reflect.ValueOf(structAddr).Elem())
            }, false)

            base.Elem().Set(val)
        default:
            tempAddr := reflect.New(ele).Interface()

            var baseAddrs = make([]interface{}, len(rowColumns))

            for k := 0; k < len(rowColumns); k++ {
                if k == 0 {
                    baseAddrs[k] = tempAddr
                } else {
                    var temp interface{}
                    baseAddrs[k] = &temp
                }
            }

            err = m.scanValues(baseAddrs, rowColumns, rows, func() {
                val = reflect.Append(val, reflect.ValueOf(tempAddr).Elem())
            }, false)

            base.Elem().Set(val)
        }
    default:
        var baseAddrs = make([]interface{}, len(rowColumns))
        for k := 0; k < len(rowColumns); k++ {
            if k == 0 {
                baseAddrs[k] = dest
            } else {
                var temp interface{}
                baseAddrs[k] = &temp
            }
        }
        err = m.scanValues(baseAddrs, rowColumns, rows, nil, true)
    }
    return err
}
