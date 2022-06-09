package orm

import (
	"strings"
)

type QueryResult struct {
	PrepareSql   string
	Bindings     []interface{}
	LastInsertId int64
	RowsAffected int64
	Err          error
}

func (q QueryResult) Error() error {
	return q.Err
}

func (q QueryResult) Sql() string {
	params := make([]string, len(q.Bindings))
	for k, v := range q.Bindings {
		params[k] = varToString(v)
	}

	var sql strings.Builder
	var index int = 0

	for _, v := range []byte(q.PrepareSql) {
		if v == '?' {
			if len(params) > index {
				sql.WriteString(params[index])
				index++
			} else {
				break
			}
		} else {
			sql.WriteByte(v)
		}
	}

	return sql.String()
}
