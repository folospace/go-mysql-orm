package orm

import (
    "database/sql"
    "strings"
)

const subqueryDefaultName = "sub"

type SubQuery struct {
    raw       string
    bindings  []interface{}
    recursive bool
    dbName    string
    tableName string
    columns   []string
    dbs       []*sql.DB
    tx        *sql.Tx
    err       error
    unionAll  bool
}

func NewSubQuery(prepareSql string, bindings ...interface{}) SubQuery {
    return SubQuery{raw: prepareSql, bindings: bindings}
}

func (m SubQuery) TableName() string {
    if m.tableName != "" {
        return m.tableName
    }
    if m.raw != "" {
        return subqueryDefaultName
    }
    return ""
}

func (m SubQuery) DatabaseName() string {
    return m.dbName
}

func (m SubQuery) Error() error {
    return m.err
}

func (m SubQuery) Sql() string {
    params := make([]string, len(m.bindings))
    for k, v := range m.bindings {
        params[k] = varToString(v)
    }

    var sqlb strings.Builder
    var index = 0

    for _, v := range []byte(m.raw) {
        if v == '?' {
            if len(params) > index {
                sqlb.WriteString(params[index])
                index++
            } else {
                break
            }
        } else {
            sqlb.WriteByte(v)
        }
    }

    return sqlb.String()
}
