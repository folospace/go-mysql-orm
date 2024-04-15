package orm

import (
    "database/sql"
    "reflect"
    "time"
)

func reflectValueIsOrmField(v reflect.Value) bool {
    if v.CanInterface() == false {
        return false
    }

    if _, ok := v.Interface().(*time.Time); ok {
        return true
    }
    if _, ok := v.Interface().(**time.Time); ok {
        return true
    }
    if _, ok := v.Interface().(sql.Scanner); ok {
        return true
    }

    vv := reflect.Indirect(v)

    if vv.CanInterface() {
        if _, ok := vv.Interface().(sql.Scanner); ok {
            return true
        }
    }

    return false
}
