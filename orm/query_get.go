package orm

import (
    "database/sql"
    "reflect"
    "strings"
)

//get first T
func (q *Query[T]) Get(primaryIds ...any) (T, QueryResult) {
    ret := reflect.New(q.tables[0].tableStructType).Interface()
    var res QueryResult

    if len(primaryIds) == 1 {
        res = q.WherePrimary(primaryIds[0]).Limit(1).GetTo(ret)
    } else {
        res = q.WherePrimary(primaryIds).Limit(1).GetTo(ret)
    }

    return ret.(T), res
}

//get slice T
func (q *Query[T]) Gets(primaryIds ...any) ([]T, QueryResult) {
    var ret []T
    var res QueryResult
    if len(primaryIds) == 1 {
        res = q.WherePrimary(primaryIds[0]).GetTo(&ret)
    } else {
        res = q.WherePrimary(primaryIds).GetTo(&ret)
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
func (q *Query[T]) GetUint() (uint64, QueryResult) {
    var ret uint64
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

//get first string
func (q *Query[T]) GetString() (string, QueryResult) {
    var ret string
    res := q.Limit(1).GetTo(&ret)
    return ret, res
}

// ↓↓ more Get examples ↓↓
func (q *Query[T]) GetSliceInt() ([]int64, QueryResult) {
    var ret []int64
    res := q.GetTo(&ret)
    return ret, res
}
func (q *Query[T]) GetSliceUint() ([]uint64, QueryResult) {
    var ret []uint64
    res := q.GetTo(&ret)
    return ret, res
}
func (q *Query[T]) GetSliceString() ([]string, QueryResult) {
    var ret []string
    res := q.GetTo(&ret)
    return ret, res
}
func (q *Query[T]) GetMapString() (map[string]string, QueryResult) {
    var ret map[string]string
    res := q.GetTo(&ret)
    return ret, res
}
func (q *Query[T]) GetMapSliceString() (map[string][]string, QueryResult) {
    var ret map[string][]string
    res := q.GetTo(&ret)
    return ret, res
}

//get count T
func (q *Query[T]) GetCount() (int64, QueryResult) {
    if len(q.groupBy) == 0 {
        if len(q.columns) == 0 {
            return q.Select("count(*)").GetInt()
        } else {
            c, err := q.parseColumn(q.columns[0])
            q.columns = nil
            if err == nil {
                cl := strings.ToLower(c)
                if strings.HasPrefix(cl, "count(") == false || strings.Contains(cl, ")") == false {
                    c = "count(" + c + ")"
                }
            }
            return q.setErr(err).Select(c).GetInt()
        }
    } else {
        tempTable := q.SubQuery()

        newQuery := NewQuery(tempTable, tempTable.dbs...)

        return newQuery.setErr(tempTable.err).Select("count(*)").GetInt()
    }
}

/*
   destPtr, pointer of any value like:
   value, []value, map[key]value, map[key][]value
*/
func (q *Query[T]) GetTo(destPtr any) QueryResult {
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

func (q *Query[T]) scanRows(dest any, rows *sql.Rows) error {
    rowColumns, gerr := rows.Columns()
    if gerr != nil {
        return gerr
    }
    destValue := reflect.ValueOf(dest)
    if destValue.Kind() != reflect.Ptr {
        return ErrDestOfGetToMustBePtr
    }
    destValueValue := destValue.Elem()
    if destValueValue.Kind() == reflect.Ptr {
        return ErrDestOfGetToMustBePtr
    }

    if reflectValueIsOrmField(destValue) == true {
        var basePtrs = make([]any, len(rowColumns))
        for k := 0; k < len(rowColumns); k++ {
            if k == 0 {
                basePtrs[k] = dest
            } else {
                var temp any
                basePtrs[k] = &temp
            }
        }
        gerr = q.scanValues(basePtrs, rowColumns, rows, nil, true)
    } else {
        switch destValueValue.Kind() {
        case reflect.Map:
            reflectMap := reflect.TypeOf(dest).Elem()

            mapKeyType := reflectMap.Key()
            mapValueType := reflectMap.Elem()
            if mapValueType.Kind() == reflect.Ptr && mapValueType.Elem() != q.tables[0].tableStructType {
                return ErrDestOfGetToMapElemMustNotBePtr
            }
            newVal := reflect.MakeMap(reflectMap)
            switch mapValueType.Kind() {
            case reflect.Ptr:
                structAddr := reflect.New(q.tables[0].tableStructType).Interface()

                structAddrMap, err := getStructFieldAddrMap(structAddr)
                if err != nil {
                    return err
                }
                var basePtrs = make([]any, len(rowColumns))

                keyType := reflectMap.Key()
                keyAddr := reflect.New(keyType).Interface()

                structVal := reflect.ValueOf(structAddr).Elem()

                for k, v := range rowColumns {
                    basePtrs[k] = structAddrMap[v]
                    if basePtrs[k] == nil {
                        if k == 0 {
                            basePtrs[k] = keyAddr
                        } else {
                            var temp any
                            basePtrs[k] = &temp
                        }
                    }
                }

                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    tmp := reflect.New(q.tables[0].tableStructType)
                    tmp.Elem().Set(structVal)
                    newVal.SetMapIndex(reflect.ValueOf(basePtrs[0]).Elem(), tmp)
                }, false)
                destValue.Elem().Set(newVal)
            case reflect.Slice: //group by results
                switch mapValueType.Elem().Kind() {
                case reflect.Ptr:
                    if mapValueType.Elem().Elem() != q.tables[0].tableStructType {
                        return ErrDestOfGetToSliceElemMustNotBePtr
                    }
                    keyAddr := reflect.New(reflectMap.Key()).Interface()

                    structAddr := reflect.New(q.tables[0].tableStructType).Interface()

                    structAddrMap, err := getStructFieldAddrMap(structAddr)
                    if err != nil {
                        return err
                    }
                    var basePtrs = make([]any, len(rowColumns))

                    structVal := reflect.ValueOf(structAddr).Elem()

                    for k, v := range rowColumns {
                        basePtrs[k] = structAddrMap[v]
                        if basePtrs[k] == nil {
                            var temp any
                            basePtrs[k] = &temp
                        }
                    }
                    basePtrs[0] = keyAddr

                    gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                        tmp := reflect.New(q.tables[0].tableStructType)
                        tmp.Elem().Set(structVal)

                        index := reflect.ValueOf(basePtrs[0]).Elem()
                        tempSlice := newVal.MapIndex(index)
                        if tempSlice.IsValid() == false {
                            tempSlice = reflect.MakeSlice(mapValueType, 0, 0)
                        }

                        newVal.SetMapIndex(index, reflect.Append(tempSlice, tmp))
                    }, false)

                    destValue.Elem().Set(newVal)
                case reflect.Struct:
                    newStructVal := reflect.New(mapValueType.Elem())
                    if reflectValueIsOrmField(newStructVal) == false {
                        keyAddr := reflect.New(reflectMap.Key()).Interface()
                        structAddr := newStructVal.Interface()
                        structAddrMap, err := getStructFieldAddrMap(structAddr)
                        if err != nil {
                            return err
                        }
                        var basePtrs = make([]any, len(rowColumns))
                        structVal := newStructVal.Elem()

                        for k, v := range rowColumns {
                            basePtrs[k] = structAddrMap[v]
                            if basePtrs[k] == nil {
                                var temp any
                                basePtrs[k] = &temp
                            }
                        }
                        basePtrs[0] = keyAddr
                        gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                            index := reflect.ValueOf(basePtrs[0]).Elem()
                            tempSlice := newVal.MapIndex(index)
                            if tempSlice.IsValid() == false {
                                tempSlice = reflect.MakeSlice(mapValueType, 0, 0)
                            }
                            newVal.SetMapIndex(index, reflect.Append(tempSlice, structVal))
                        }, false)
                        destValue.Elem().Set(newVal)
                        break
                    }
                    fallthrough
                default:
                    newKeyVal := reflect.New(reflectMap.Key())
                    var basePtrs = make([]any, len(rowColumns))
                    for k := range basePtrs {
                        if k == 0 {
                            basePtrs[0] = newKeyVal.Interface()
                        } else {
                            basePtrs[k] = reflect.New(mapValueType.Elem()).Interface()
                        }
                    }
                    gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                        index := newKeyVal.Elem()
                        tempSlice := newVal.MapIndex(index)
                        if tempSlice.IsValid() == false {
                            tempSlice = reflect.MakeSlice(mapValueType, 0, 0)
                        }
                        for k := range basePtrs {
                            if k > 0 {
                                tempSlice = reflect.Append(tempSlice, reflect.ValueOf(basePtrs[k]).Elem())
                            }
                        }
                        newVal.SetMapIndex(index, tempSlice)
                    }, false)
                    destValue.Elem().Set(newVal)
                }
            case reflect.Struct:
                newValue := reflect.New(mapValueType)
                if reflectValueIsOrmField(newValue) == false {
                    structAddr := newValue.Interface()
                    structAddrMap, err := getStructFieldAddrMap(structAddr)
                    if err != nil {
                        return err
                    }
                    var basePtrs = make([]any, len(rowColumns))

                    keyType := reflectMap.Key()
                    keyAddr := reflect.New(keyType).Interface()

                    structVal := newValue.Elem()

                    for k, v := range rowColumns {
                        basePtrs[k] = structAddrMap[v]
                        if k == 0 {
                            basePtrs[k] = keyAddr
                        } else if basePtrs[k] == nil {
                            var temp any
                            basePtrs[k] = &temp
                        }
                    }
                    gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                        newVal.SetMapIndex(reflect.ValueOf(basePtrs[0]).Elem(), structVal)
                    }, false)
                    destValue.Elem().Set(newVal)
                    break
                }
                fallthrough
            default:
                newKeyValue := reflect.New(mapKeyType)
                newValValue := reflect.New(mapValueType)

                var basePtrs = make([]any, len(rowColumns))

                for k := 0; k < len(rowColumns); k++ {
                    if k == 0 {
                        basePtrs[k] = newKeyValue.Interface()
                    } else if k == 1 {
                        basePtrs[k] = newValValue.Interface()
                    } else {
                        var temp any
                        basePtrs[k] = &temp
                    }
                }
                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    newVal.SetMapIndex(newKeyValue.Elem(), newValValue.Elem())
                }, false)

                destValue.Elem().Set(newVal)
            }
        case reflect.Slice:
            ele := reflect.TypeOf(dest).Elem().Elem()
            if ele.Kind() == reflect.Ptr && ele.Elem() != q.tables[0].tableStructType {
                return ErrDestOfGetToSliceElemMustNotBePtr
            }

            switch ele.Kind() {
            case reflect.Ptr:
                structAddr := reflect.New(q.tables[0].tableStructType).Interface()

                structAddrMap, err := getStructFieldAddrMap(structAddr)
                if err != nil {
                    return err
                }
                var basePtrs = make([]any, len(rowColumns))

                structVal := reflect.ValueOf(structAddr).Elem()

                for k, v := range rowColumns {
                    basePtrs[k] = structAddrMap[v]
                    if basePtrs[k] == nil {
                        var temp any
                        basePtrs[k] = &temp
                    }
                }

                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    tmp := reflect.New(q.tables[0].tableStructType)
                    tmp.Elem().Set(structVal)
                    destValueValue = reflect.Append(destValueValue, tmp)
                }, false)

                destValue.Elem().Set(destValueValue)
            case reflect.Struct:
                eleNew := reflect.New(ele)
                if reflectValueIsOrmField(eleNew) == false {
                    structAddr := eleNew.Interface()
                    structVal := eleNew.Elem()
                    structAddrMap, err := getStructFieldAddrMap(structAddr)
                    if err != nil {
                        return err
                    }
                    var basePtrs = make([]any, len(rowColumns))

                    for k, v := range rowColumns {
                        basePtrs[k] = structAddrMap[v]
                        if basePtrs[k] == nil {
                            var temp any
                            basePtrs[k] = &temp
                        }
                    }

                    gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                        destValueValue = reflect.Append(destValueValue, structVal)
                    }, false)

                    destValue.Elem().Set(destValueValue)
                    break
                }
                fallthrough
            default:
                var basePtrs = make([]any, len(rowColumns))

                for k := 0; k < len(rowColumns); k++ {
                    basePtrs[k] = reflect.New(ele).Interface()
                }

                gerr = q.scanValues(basePtrs, rowColumns, rows, func() {
                    for _, v := range basePtrs {
                        destValueValue = reflect.Append(destValueValue, reflect.ValueOf(v).Elem())
                    }
                }, false)

                destValue.Elem().Set(destValueValue)
            }
        case reflect.Struct:
            if reflectValueIsOrmField(destValue) == false {
                structAddr := dest
                structAddrMap, err := getStructFieldAddrMap(structAddr)
                if err != nil {
                    return err
                }
                var basePtrs = make([]any, len(rowColumns))

                for k, v := range rowColumns {
                    basePtrs[k] = structAddrMap[v]
                    if basePtrs[k] == nil {
                        var temp any
                        basePtrs[k] = &temp
                    }
                }
                gerr = q.scanValues(basePtrs, rowColumns, rows, nil, true)
                break
            }
            fallthrough
        default:
            var basePtrs = make([]any, len(rowColumns))
            for k := 0; k < len(rowColumns); k++ {
                if k == 0 {
                    basePtrs[k] = dest
                } else {
                    var temp any
                    basePtrs[k] = &temp
                }
            }
            gerr = q.scanValues(basePtrs, rowColumns, rows, nil, true)
        }
    }
    return gerr
}

func (q *Query[T]) scanValues(basePtrs []any, rowColumns []string, rows *sql.Rows, setVal func(), tryOnce bool) error {
    var err error
    var tempPtrs = make([]any, len(rowColumns))
    for k := range rowColumns {
        var temp any
        tempPtrs[k] = &temp
    }

    finalPtrs := make([]any, len(rowColumns))

    for rows.Next() {
        err = rows.Scan(tempPtrs...)
        if err != nil {
            return err
        }

        for k, v := range tempPtrs {
            if *v.(*any) == nil {
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
