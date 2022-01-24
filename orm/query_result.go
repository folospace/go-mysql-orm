package orm

import (
    "fmt"
    "strings"
)

type QueryResult struct {
    PrepareSql   string
    Bindings     []interface{}
    LastInsertId int64
    RowsAffected int64
    Err          error
}

func (q *QueryResult) Sql() string {
    return fmt.Sprintf(strings.Replace(q.PrepareSql, "?", "'%+v'", len(q.Bindings)), q.Bindings...)
}
