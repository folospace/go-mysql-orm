package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func castFieldsToStrSlice(tableObjAddr interface{}, tableColumnAddrs ...interface{}) ([]string, error) {
	if len(tableColumnAddrs) == 0 {
		return nil, nil
	}

	tableStructAddr := reflect.ValueOf(tableObjAddr)
	if tableStructAddr.Kind() != reflect.Ptr {
		return nil, errors.New("params must be address of variable")
	}

	tableStruct := tableStructAddr.Elem()

	if tableStruct.Kind() != reflect.Struct {
		return nil, errors.New("obj must be struct")
	}

	tableStructType := reflect.TypeOf(tableObjAddr).Elem()

	var columns []string
	for k, v := range tableColumnAddrs {
		columnVar := reflect.ValueOf(v)
		if columnVar.Kind() != reflect.Ptr {
			return nil, errors.New("params must be address of variable")
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
		fmt.Println(tableStructAddr.Kind())
		return nil, errors.New("params must be address of variable")
	}

	tableStruct := tableStructAddr.Elem()
	if tableStruct.Kind() != reflect.Struct {
		return nil, errors.New("obj must be struct")
	}

	tableStructType := reflect.TypeOf(objAddr).Elem()

	ret := make(map[string]interface{})

	for i := 0; i < tableStruct.NumField(); i++ {
		valueField := tableStruct.Field(i)

		name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
		if name == "-" {
			name = ""
		}
		if name != "" {
			ret[name] = valueField.Addr().Interface()
		}
	}
	return ret, nil
}

func getStructAddrFieldMap(objAddr interface{}) (map[interface{}]string, error) {
	tableStructAddr := reflect.ValueOf(objAddr)
	if tableStructAddr.Kind() != reflect.Ptr {
		return nil, errors.New("params must be address of variable")
	}

	tableStruct := tableStructAddr.Elem()
	if tableStruct.Kind() != reflect.Struct {
		return nil, errors.New("obj must be struct")
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
	tableStruct := reflect.ValueOf(obj)
	if tableStruct.Kind() != reflect.Struct {
		return nil, errors.New("obj must be struct")
	}

	tableStructType := reflect.TypeOf(obj)

	ret := make([]string, tableStruct.NumField())

	for i := 0; i < tableStruct.NumField(); i++ {
		name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
		if name == "-" {
			name = ""
		}
		ret[i] = name
	}
	return ret, nil
}
