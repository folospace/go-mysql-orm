package orm

import (
    "reflect"
)

func SliceContain(array any, val any) (index int) {
    index = -1
    switch reflect.TypeOf(array).Kind() {
    case reflect.Slice:
        {
            s := reflect.ValueOf(array)
            for i := 0; i < s.Len(); i++ {
                if reflect.DeepEqual(val, s.Index(i).Interface()) {
                    index = i
                    return
                }
            }
        }
    }
    return
}
