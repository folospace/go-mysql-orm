package orm

import (
	"database/sql"
	"reflect"
	"strings"
)

//get first T
func (m Query[T]) Get(primaryValue ...interface{}) (T, QueryResult) {
	var ret T
	if len(primaryValue) > 0 {
		res := m.Where(m.tables[0].tableStruct.Field(0).Addr().Interface(), primaryValue[0]).Limit(1).GetTo(&ret)
		return ret, res
	} else {
		res := m.Limit(1).GetTo(&ret)
		return ret, res
	}
}

//get slice T
func (m Query[T]) Gets(primaryIds ...interface{}) ([]T, QueryResult) {
	var ret []T
	if len(primaryIds) > 0 {
		res := m.Where(m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, primaryIds).GetTo(&ret)
		return ret, res
	} else {
		res := m.GetTo(&ret)
		return ret, res
	}
}

//get first row
func (m Query[T]) GetRow() (map[string]interface{}, QueryResult) {
	var ret map[string]interface{}
	res := m.Limit(1).GetTo(&ret)
	return ret, res
}

//get slice row
func (m Query[T]) GetRows() ([]map[string]interface{}, QueryResult) {
	var ret []map[string]interface{}
	res := m.GetTo(&ret)
	return ret, res
}

//get count T
func (m Query[T]) GetCount() (int64, QueryResult) {
	var ret int64
	if len(m.groupBy) == 0 {
		if len(m.columns) == 0 {
			res := m.Select("count(*)").GetTo(&ret)
			return ret, res
		} else {
			c, err := m.parseColumn(m.columns[0])
			m.columns = nil
			if err == nil {
				cl := strings.ToLower(c)
				if strings.HasPrefix(cl, "count(") == false || strings.Contains(cl, ")") == false {
					c = "count(" + c + ")"
				}
			}
			res := m.setErr(err).Select(c).GetTo(&ret)
			return ret, res
		}
	} else {
		tempTable := m.SubQuery()

		newQuery := NewQuery(tempTable, tempTable.db)

		res := newQuery.Select("count(*)").GetTo(&ret)
		return ret, res
	}
}

//destPtr: *int | *int64 |  *string | ...
//destPtr: *[]int | *[]string | ...
//destPtr: *struct | *[]struct
//destPtr: *map [int | string | ...] int | string ...
//destPtr: *map [int | string | ...] struct
//destPtr: *map [int | string | ...] []struct
func (m Query[T]) GetTo(destPtr interface{}) QueryResult {
	tempTable := m.SubQuery()

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

	m.result.Err = m.scanRows(destPtr, rows)
	return m.result
}

func (m Query[T]) scanValues(basePtrs []interface{}, rowColumns []string, rows *sql.Rows, setVal func(), tryOnce bool) error {
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
			if reflect.ValueOf(v).Elem().IsNil() {
				felement := reflect.ValueOf(basePtrs[k]).Elem()
				felement.Set(reflect.Zero(felement.Type()))
				finalPtrs[k] = v
			} else {
				finalPtrs[k] = basePtrs[k]
			}
		}

		err = rows.Scan(finalPtrs...)
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

func (m Query[T]) scanRows(dest interface{}, rows *sql.Rows) error {
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
		ele := reflect.TypeOf(dest).Elem().Elem()
		if ele.Kind() == reflect.Ptr {
			return ErrDestOfGetToSliceElemMustNotBePtr
		}

		newVal := reflect.MakeMap(reflect.TypeOf(dest).Elem())
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
			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
				newVal.SetMapIndex(reflect.ValueOf(basePtrs[0]).Elem(), reflect.ValueOf(structAddr).Elem())
			}, false)
			base.Elem().Set(newVal)
		case reflect.Slice:
			if ele.Elem().Kind() != reflect.Struct {
				return ErrDestOfGetToSliceElemMustBeStruct
			}
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
			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
				index := reflect.ValueOf(basePtrs[0]).Elem()
				tempSlice := newVal.MapIndex(index)
				if tempSlice.IsValid() == false {
					tempSlice = reflect.MakeSlice(ele, 0, 0)
				}
				newVal.SetMapIndex(index, reflect.Append(tempSlice, reflect.ValueOf(structAddr).Elem()))
			}, false)
			base.Elem().Set(newVal)

		case reflect.Interface:
			if reflect.TypeOf(dest).Elem().Key().Kind() == reflect.String {

				var basePtrs = make([]interface{}, len(rowColumns))
				for k := range basePtrs {
					var temp interface{}
					basePtrs[k] = &temp
				}

				gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
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
			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
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
		gerr = m.scanValues(basePtrs, rowColumns, rows, nil, true)
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

			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
				val = reflect.Append(val, reflect.ValueOf(structAddr).Elem())
			}, false)

			base.Elem().Set(val)
		case reflect.Map:
			var basePtrs = make([]interface{}, len(rowColumns))

			for k := range basePtrs {
				var temp interface{}
				basePtrs[k] = &temp
			}
			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
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

			gerr = m.scanValues(basePtrs, rowColumns, rows, func() {
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
		gerr = m.scanValues(basePtrs, rowColumns, rows, nil, true)
	}
	return gerr
}
