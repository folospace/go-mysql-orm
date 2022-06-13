package orm

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var structFieldsCache sync.Map

func getFieldsCache(key string) []string {
	intef, ok := structFieldsCache.Load(key)
	if ok {
		ret, ok := intef.([]string)
		if ok {
			return ret
		}
	}
	return nil
}
func setFieldsCache(key string, val []string) {
	structFieldsCache.Store(key, val)
}

func castFieldsToStrSlice(tableObjAddr interface{}, tableColumnPtrs ...interface{}) ([]string, error) {
	if len(tableColumnPtrs) == 0 {
		return nil, nil
	}

	tableStructAddr := reflect.ValueOf(tableObjAddr)
	if tableStructAddr.Kind() != reflect.Ptr {
		return nil, ErrParamMustBePtr
	}

	tableStruct := tableStructAddr.Elem()

	if tableStruct.Kind() != reflect.Struct {
		return nil, ErrParamElemKindMustBeStruct
	}

	tableStructType := reflect.TypeOf(tableObjAddr).Elem()

	var columns []string
	for k, v := range tableColumnPtrs {
		columnVar := reflect.ValueOf(v)
		if columnVar.Kind() != reflect.Ptr {
			return nil, ErrParamMustBePtr
		}

		for i := 0; i < tableStruct.NumField(); i++ {
			valueField := tableStruct.Field(i)
			if valueField.Addr().Interface() == columnVar.Elem().Addr().Interface() {
				name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
				if name != "" && name != "-" {
					columns = append(columns, name)
				}
				break
			} else if i == tableStruct.NumField()-1 {
				return columns, errors.New("param " + strconv.Itoa(k+2) + " is not a field of first obj")
			}
		}
	}

	return columns, nil
}

func getStructFieldAddrMap(objAddr interface{}) (map[string]interface{}, error) {
	tableStructAddr := reflect.ValueOf(objAddr)
	if tableStructAddr.Kind() != reflect.Ptr {
		return nil, ErrParamMustBePtr
	}

	tableStruct := tableStructAddr.Elem()
	if tableStruct.Kind() != reflect.Struct {
		return nil, ErrParamElemKindMustBeStruct
	}

	tableStructType := reflect.TypeOf(objAddr).Elem()

	ret := make(map[string]interface{})

	fields, err := getStructFieldNameSlice(tableStruct.Interface())
	if err != nil {
		return nil, err
	}

	for i := 0; i < tableStruct.NumField(); i++ {
		if tableStruct.Field(i).Kind() == reflect.Struct && tableStructType.Field(i).Anonymous {
			innerMap, err := getStructFieldAddrMap(tableStruct.Field(i).Addr().Interface())
			if err != nil {
				return ret, err
			}
			for k, v := range innerMap {
				ret[k] = v
			}
		} else {
			valueField := tableStruct.Field(i)

			name := fields[i]
			if name != "" {
				ret[name] = valueField.Addr().Interface()
			}

		}
	}

	return ret, nil
}

func getStructAddrFieldMap(objAddr interface{}) (map[interface{}]string, error) {
	tableStructAddr := reflect.ValueOf(objAddr)
	if tableStructAddr.Kind() != reflect.Ptr {
		return nil, ErrParamMustBePtr
	}

	tableStruct := tableStructAddr.Elem()
	if tableStruct.Kind() != reflect.Struct {
		return nil, ErrParamElemKindMustBeStruct
	}

	tableStructType := reflect.TypeOf(objAddr).Elem()

	ret := make(map[interface{}]string)

	for i := 0; i < tableStruct.NumField(); i++ {
		valueField := tableStruct.Field(i)

		name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
		if name == "-" {
			name = ""
		}
		if name != "" {
			ret[valueField.Addr().Interface()] = name
		}
	}
	return ret, nil
}

func getStructFieldNameSlice(obj interface{}) ([]string, error) {
	tableStructType := reflect.TypeOf(obj)

	fieldsCache := getFieldsCache(tableStructType.String())
	if fieldsCache != nil {
		return fieldsCache, nil
	}

	tableStruct := reflect.ValueOf(obj)
	if tableStruct.Kind() != reflect.Struct {
		return nil, ErrParamElemKindMustBeStruct
	}
	var ret = make([]string, tableStruct.NumField())

	for i := 0; i < tableStruct.NumField(); i++ {
		//if tableStruct.Field(i).Kind() == reflect.Struct && tableStructType.Field(i).Anonymous {
		//	innerFields, err := getStructFieldNameSlice(tableStruct.Field(i).Interface())
		//	if err != nil {
		//		return ret, err
		//	}
		//	ret = append(ret, innerFields...)
		//}
		ormTag := strings.Split(tableStructType.Field(i).Tag.Get("orm"), ",")[0]
		if ormTag == "-" {
			ormTag = ""
		}

		if ormTag != "" {
			ret[i] = ormTag
			continue
		}

		ormTag = strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
		if ormTag == "-" {
			ormTag = ""
		}
		if ormTag != "" {
			ret[i] = ormTag
		}
	}

	setFieldsCache(tableStructType.String(), ret)
	return ret, nil
}

func getStructFieldWithDefaultTime(obj interface{}) (map[int]interface{}, error) {
	tableStructType := reflect.TypeOf(obj)

	tableStruct := reflect.ValueOf(obj)
	if tableStruct.Kind() != reflect.Struct {
		return nil, ErrParamElemKindMustBeStruct
	}
	ret := make(map[int]interface{})

	for i := 0; i < tableStruct.NumField(); i++ {
		defaultVar := tableStructType.Field(i).Tag.Get("default")
		if defaultVar == "" {
			continue
		}

		v := tableStruct.Field(i)
		if v.CanInterface() {
			if _, ok := v.Interface().(time.Time); ok {
				if strings.Contains(strings.ToLower(defaultVar), "current_timestamp") {
					ret[i] = time.Now()
				}
			}
		}
	}
	return ret, nil
}
