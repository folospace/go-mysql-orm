package orm

import (
    "encoding/json"
    "reflect"
)

func MakeAllNilSliceEmpty(inter interface{}) interface{} {
    // original input that can't be modified
    val := reflect.ValueOf(inter)

    switch val.Kind() {
    case reflect.Slice:
        newSlice := reflect.MakeSlice(val.Type(), 0, val.Len())
        if !val.IsZero() {
            // iterate over each element in slice
            for j := 0; j < val.Len(); j++ {
                item := val.Index(j)

                var newItem reflect.Value
                switch item.Kind() {
                case reflect.Struct:
                    // recursively handle nested struct
                    newItem = reflect.Indirect(reflect.ValueOf(MakeAllNilSliceEmpty(item.Interface())))
                default:
                    newItem = item
                }

                newSlice = reflect.Append(newSlice, newItem)
            }

        }
        return newSlice.Interface()
    case reflect.Struct:
        // new struct that will be returned
        newStruct := reflect.New(reflect.TypeOf(inter))
        newVal := newStruct.Elem()
        // iterate over input's fields
        for i := 0; i < val.NumField(); i++ {
            newValField := newVal.Field(i)
            valField := val.Field(i)
            if newValField.CanSet() == false {
                continue
            }
            switch valField.Kind() {
            case reflect.Slice:
                // recursively handle nested slice
                newValField.Set(reflect.Indirect(reflect.ValueOf(MakeAllNilSliceEmpty(valField.Interface()))))
            case reflect.Struct:
                // recursively handle nested struct
                if _, ok := valField.Interface().(json.Marshaler); ok == false {
                    newValField.Set(reflect.Indirect(reflect.ValueOf(MakeAllNilSliceEmpty(valField.Interface()))))
                } else {
                    newValField.Set(valField)
                }
            default:
                newValField.Set(valField)
            }
        }

        return newStruct.Interface()
    case reflect.Map:
        // new map to be returned
        newMap := reflect.MakeMap(reflect.TypeOf(inter))
        // iterate over every key value pair in input map
        iter := val.MapRange()
        for iter.Next() {
            k := iter.Key()
            v := iter.Value()
            // recursively handle nested value
            newV := reflect.Indirect(reflect.ValueOf(MakeAllNilSliceEmpty(v.Interface())))
            newMap.SetMapIndex(k, newV)
        }
        return newMap.Interface()
    case reflect.Ptr:
        // dereference pointer
        if val.IsNil() {
            temp := reflect.TypeOf(inter).Elem()
            if temp.Kind() == reflect.Slice {
                return reflect.MakeSlice(temp, 0, 0).Interface()
            } else {
                return inter
            }
        }
        return MakeAllNilSliceEmpty(val.Elem().Interface())
    default:
        return inter
    }
}
