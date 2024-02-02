package orm

import (
    "database/sql"
    "reflect"
    "strings"
)

//get first T
func (q *Query[T]) Get(primaryIds ...interface{}) (T, QueryResult) {
    var ret T
    var res QueryResult
    if len(primaryIds) == 1 {
        res = q.WherePrimary(primaryIds[0]).Limit(1).GetTo(&ret)
    } else {
        res = q.WherePrimary(primaryIds).Limit(1).GetTo(&ret)
    }
    return ret, res
}

//get first bool
func (q *Query[T]) GetBool() (bool, QueryResult) {
    var ret bool
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

//get first int
func (q *Query[T]) GetInt() (int64, QueryResult) {
    var ret int64
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

//get first string
func (q *Query[T]) GetString() (string, QueryResult) {
    var ret string
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

//get slice T
func (q *Query[T]) Gets(primaryIds ...interface{}) ([]T, QueryResult) {
    var ret []T
    var res QueryResult
    if len(primaryIds) == 1 {
        res = q.WherePrimary(primaryIds[0]).GetTo(&ret)
    } else {
        res = q.WherePrimary(primaryIds).GetTo(&ret)
    }
    return ret, res
}

//get first row
func (q *Query[T]) GetRow() (map[string]interface{}, QueryResult) {
    var ret map[string]interface{}
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

//get slice row
func (q *Query[T]) GetRows() ([]map[string]interface{}, QueryResult) {
    var ret []map[string]interface{}
    res := q.GetTo(&ret)
    return ret, res
}

//get count T
func (q *Query[T]) GetCount() (int64, QueryResult) {
    var ret int64
    if len(q.groupBy) == 0 {
        if len(q.columns) == 0 {
            res := q.Select("count(*)").GetTo(&ret)
            return ret, res
        } else {
            c, err := q.parseColumn(q.columns[0])
            q.columns = nil
            if err == nil {
                cl := strings.ToLower(c)
                if strings.HasPrefix(cl, "count(") == false || strings.Contains(cl, ")") == false {
                    c = "count(" + c + ")"
                }
            }
            res := q.setErr(err).Select(c).GetTo(&ret)
            return ret, res
        }
    } else {
        tempTable := q.SubQuery()

        newQuery := NewQuery(tempTable, tempTable.dbs...)

        res := newQuery.setErr(tempTable.err).Select("count(*)").GetTo(&ret)
        return ret, res
    }
}

//destPtr: *int | *int64 |  *string | ...
//destPtr: *[]int | *[]string | ...
//destPtr: *struct | *[]struct
//destPtr: *map [int | string | ...] int | string ...
//destPtr: *map [int | string | ...] struct
//destPtr: *map [int | string | ...] []struct
func (q *Query[T]) GetTo(destPtr interface{}) QueryResult {
    tempTable := q.SubQuery()

    q.result.PrepareSql = tempTable.raw
    q.result.Bindings = tempTable.bindings
    if tempTable.err != nil {
        q.result.Err = tempTable.err
    }

    if q.result.Err != nil {
        if errorLogger != nil {
            errorLogger.Error(q.result.Sql(), q.result.Error())
        }
        return q.result
    } else if infoLogger != nil {
        infoLogger.Info(q.result.Sql(), q.result.Error())
    }

    var rows *sql.Rows
    var err error
    if q.Tx() != nil {
        if q.ctx != nil {
            rows, err = q.Tx().QueryContext(*q.ctx, tempTable.raw, tempTable.bindings...)
        } else {
            rows, err = q.Tx().Query(tempTable.raw, tempTable.bindings...)
        }
    } else {
        if q.ctx != nil {
            rows, err = q.readDB().QueryContext(*q.ctx, tempTable.raw, tempTable.bindings...)
        } else {
            rows, err = q.readDB().Query(tempTable.raw, tempTable.bindings...)
        }
    }

    defer func() {
        if rows != nil {
            _ = rows.Close()
        }
    }()

    if err != nil {
        q.result.Err = err
        if errorLogger != nil {
            errorLogger.Error(q.result.Sql(), q.result.Error())
        }
        return q.result
    }

    q.result.Err = q.scanRows(destPtr, rows)
    return q.result
}

func (q *Query[T]) scanValues(basePtrs []interface{}, rowColumns []string, rows *sql.Rows, setVal func(), tryOnce bool) error {
    var err error
    var tempPtrs = make([]interface{}, len(rowColumns))
    for k := range rowColumns {
        var temp interface{}
        tempPtrs[k] = &temp
    }

    finalPtrs := make([]interface{}, len(rowColumns))

    for rows.Next() {
        err = rows.Scan(tempPtrs...)
        if err != nil {
            return err
        }

        for k, v := range tempPtrs {
            if *v.(*interface{}) == nil {
                felement := reflect.ValueOf(basePtrs[k]).Elem()
                felement.Set(reflect.Zero(felement.Type()))
                finalPtrs[k] = v
            } else {
                finalPtrs[k] = basePtrs[k]
            }
        }

        err = rows.Scan(finalPtrs...)
        q.result.RowsAffected += 1

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

func (q *Query[T]) scanRows(dest interface{}, rows *sql.Rows) error {
    rowColumns, gerr := rows.Columns()
    if gerr != nil {
        return gerr
    }
    base := reflect.ValueOf(dest)
    if base.Kind() != reflect.Ptr {
        return ErrDestOfGetToMustBePtr
    }
    val := base.Elem()
    if val.Kind() == reflect.Ptr {
        return ErrDestOfGetToMustBePtr
    }

    switch val.Kind() {
    case reflect.Map:
        reflectMap := reflect.TypeOf(dest).Elem()

        ele := reflectMap.Elem()
        if ele.Kind() == reflect.Ptr {
            return ErrDestOfGetToSliceElemMustNotBePtr
        }
        newVal := reflect.MakeMap(reflectMap)
        switch ele.Kind() {
        case reflect.Struct:
            structAddr := reflect.New(ele).Interface()
            structAddrMap, err := getStructFieldAddrMap(structAddr)
            if err != nil {
                return err
            }
            var basePtrs = make([]interface{}, len(rowColumns))

            for k, v := range rowColumns {
                basePtrs[k] = structAddrMap[v]
                if basePtrs[k] == nil {
                    var temp interface{}
                    basePtrs[k] = &temp
                }
            }
            gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                newVal.SetMapIndex(reflect.ValueOf(basePtrs[0]).Elem(), reflect.ValueOf(structAddr).Elem())
            }, false)
            base.Elem().Set(newVal)
        case reflect.Slice:
            switch ele.Elem().Kind() {
            case reflect.Struct:
                keyAddr := reflect.New(reflectMap.Key()).Interface()
                structAddr := reflect.New(ele.Elem()).Interface()
                structAddrMap, err := getStructFieldAddrMap(structAddr)
                if err != nil {
                    return err
                }
                var basePtrs = make([]interface{}, len(rowColumns))

                for k, v := range rowColumns {
                    basePtrs[k] = structAddrMap[v]
                    if basePtrs[k] == nil {
                        var temp interface{}
                        basePtrs[k] = &temp
                    }
                }
                basePtrs[0] = keyAddr
                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    index := reflect.ValueOf(basePtrs[0]).Elem()
                    tempSlice := newVal.MapIndex(index)
                    if tempSlice.IsValid() == false {
                        tempSlice = reflect.MakeSlice(ele, 0, 0)
                    }
                    newVal.SetMapIndex(index, reflect.Append(tempSlice, reflect.ValueOf(structAddr).Elem()))
                }, false)
                base.Elem().Set(newVal)
            default:
                keyAddr := reflect.New(reflectMap.Key()).Interface()
                valAddr := reflect.New(ele.Elem()).Interface()

                var basePtrs = make([]interface{}, 2)
                basePtrs[0] = keyAddr
                basePtrs[1] = valAddr

                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    index := reflect.ValueOf(basePtrs[0]).Elem()
                    tempSlice := newVal.MapIndex(index)
                    if tempSlice.IsValid() == false {
                        tempSlice = reflect.MakeSlice(ele, 0, 0)
                    }
                    newVal.SetMapIndex(index, reflect.Append(tempSlice, reflect.ValueOf(valAddr).Elem()))
                }, false)
                base.Elem().Set(newVal)
            }

        case reflect.Interface:
            if reflect.TypeOf(dest).Elem().Key().Kind() == reflect.String {

                var basePtrs = make([]interface{}, len(rowColumns))
                for k := range basePtrs {
                    var temp interface{}
                    basePtrs[k] = &temp
                }

                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    for k, v := range rowColumns {
                        newVal.SetMapIndex(reflect.ValueOf(v), reflect.ValueOf(basePtrs[k]).Elem())
                    }
                }, true)

                base.Elem().Set(newVal)
                return gerr
            }
            fallthrough
        default:
            keyType := reflect.TypeOf(dest).Elem().Key()

            keyAddr := reflect.New(keyType).Interface()
            tempAddr := reflect.New(ele).Interface()

            var basePtrs = make([]interface{}, len(rowColumns))

            for k := 0; k < len(rowColumns); k++ {
                if k == 0 {
                    basePtrs[k] = keyAddr
                } else if k == 1 {
                    basePtrs[k] = tempAddr
                } else {
                    var temp interface{}
                    basePtrs[k] = &temp
                }
            }
            gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
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
        var basePtrs = make([]interface{}, len(rowColumns))

        for k, v := range rowColumns {
            basePtrs[k] = structAddrMap[v]
            if basePtrs[k] == nil {
                var temp interface{}
                basePtrs[k] = &temp
            }
        }
        gerr = q.scanValues(basePtrs, rowColumns, rows, nil, true)
    case reflect.Slice:
        ele := reflect.TypeOf(dest).Elem().Elem()
        if ele.Kind() == reflect.Ptr {
            return ErrDestOfGetToSliceElemMustNotBePtr
        }

        switch ele.Kind() {
        case reflect.Struct:
            structAddr := reflect.New(ele).Interface()
            structAddrMap, err := getStructFieldAddrMap(structAddr)
            if err != nil {
                return err
            }
            var basePtrs = make([]interface{}, len(rowColumns))

            for k, v := range rowColumns {
                basePtrs[k] = structAddrMap[v]
                if basePtrs[k] == nil {
                    var temp interface{}
                    basePtrs[k] = &temp
                }
            }

            gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                val = reflect.Append(val, reflect.ValueOf(structAddr).Elem())
            }, false)

            base.Elem().Set(val)
        case reflect.Map:
            var basePtrs = make([]interface{}, len(rowColumns))

            valEle := ele.Elem()

            for k := range basePtrs {
                //var temp interface{}
                basePtrs[k] = reflect.New(valEle).Interface()
            }
            gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                newVal := reflect.MakeMap(ele)
                for k, v := range rowColumns {
                    newVal.SetMapIndex(reflect.ValueOf(v), reflect.ValueOf(basePtrs[k]).Elem())
                }
                val = reflect.Append(val, newVal)
            }, false)

            base.Elem().Set(val)
        default:
            tempAddr := reflect.New(ele).Interface()

            var basePtrs = make([]interface{}, len(rowColumns))

            for k := 0; k < len(rowColumns); k++ {
                if k == 0 {
                    basePtrs[k] = tempAddr
                } else {
                    var temp interface{}
                    basePtrs[k] = &temp
                }
            }

            gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                val = reflect.Append(val, reflect.ValueOf(tempAddr).Elem())
            }, false)

            base.Elem().Set(val)
        }
    default:
        var basePtrs = make([]interface{}, len(rowColumns))
        for k := 0; k < len(rowColumns); k++ {
            if k == 0 {
                basePtrs[k] = dest
            } else {
                var temp interface{}
                basePtrs[k] = &temp
            }
        }
        gerr = q.scanValues(basePtrs, rowColumns, rows, nil, true)
    }
    return gerr
}
