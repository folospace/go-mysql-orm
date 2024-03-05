package orm

import (
    "sync"
)

var tableCache sync.Map

func getTableFromCache(key any) *queryTable {
    res, ok := tableCache.Load(key)
    if ok {
        ret, ok := res.(*queryTable)
        if ok {
            return ret
        }
    }
    return nil
}

func cacheTable(key any, val *queryTable) {
    tableCache.Store(key, val)
}
