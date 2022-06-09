package orm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	escape  = `'`
	nullStr = "NULL"
)

func varToString(i interface{}) string {
	switch v := i.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%.6f", v)
	case bool:
		return strconv.FormatBool(v)
	case string:
		return escape + strings.ReplaceAll(v, escape, "\\"+escape) + escape
	case []byte:
		if s := string(v); stringIsPrintable(s) {
			return escape + strings.ReplaceAll(s, escape, "\\"+escape) + escape
		} else {
			return escape + "<binary>" + escape
		}
	case time.Time:
		if v.IsZero() {
			return escape + "0000-00-00 00:00:00" + escape
		} else {
			return escape + v.Format("2006-01-02 15:04:05.999") + escape
		}
	case *time.Time:
		if v != nil {
			if v.IsZero() {
				return escape + "0000-00-00 00:00:00" + escape
			} else {
				return escape + v.Format("2006-01-02 15:04:05.999") + escape
			}
		} else {
			return nullStr
		}
	case driver.Valuer:
		reflectValue := reflect.ValueOf(v)
		if v != nil && reflectValue.IsValid() && ((reflectValue.Kind() == reflect.Ptr && !reflectValue.IsNil()) || reflectValue.Kind() != reflect.Ptr) {
			r, _ := v.Value()
			return varToString(r)
		} else {
			return nullStr
		}
	case fmt.Stringer:
		reflectValue := reflect.ValueOf(v)
		switch reflectValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fmt.Sprintf("%d", reflectValue.Interface())
		case reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%.6f", reflectValue.Interface())
		case reflect.Bool:
			return fmt.Sprintf("%t", reflectValue.Interface())
		case reflect.String:
			return escape + strings.ReplaceAll(fmt.Sprintf("%v", v), escape, "\\"+escape) + escape
		default:
			if v != nil && reflectValue.IsValid() && ((reflectValue.Kind() == reflect.Ptr && !reflectValue.IsNil()) || reflectValue.Kind() != reflect.Ptr) {
				return escape + strings.ReplaceAll(fmt.Sprintf("%v", v), escape, "\\"+escape) + escape
			} else {
				return nullStr
			}
		}
	default:
		rv := reflect.ValueOf(v)
		if v == nil || !rv.IsValid() || rv.Kind() == reflect.Ptr && rv.IsNil() {
			return nullStr
		} else if valuer, ok := v.(driver.Valuer); ok {
			v, _ = valuer.Value()
			return varToString(v)
		} else if rv.Kind() == reflect.Ptr && !rv.IsZero() {
			return varToString(reflect.Indirect(rv).Interface())
		} else {
			for _, t := range []reflect.Type{reflect.TypeOf(time.Time{}), reflect.TypeOf(false), reflect.TypeOf([]byte{})} {
				if rv.Type().ConvertibleTo(t) {
					return varToString(rv.Convert(t).Interface())
				}
			}
			return escape + strings.ReplaceAll(fmt.Sprint(v), escape, "\\"+escape) + escape
		}
	}
}

func stringIsPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
