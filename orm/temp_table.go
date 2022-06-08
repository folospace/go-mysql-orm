package orm

import "database/sql"

type SubQuery struct {
    raw       string
    bindings  []interface{}
    recursive bool
    dbName    string
    tableName string
    columns   []string
    db        *sql.DB
    tx        *sql.Tx
    err       error
}

func (m SubQuery) TableName() string {
    if m.tableName != "" {
        return m.tableName
    }
    return "sub"
}

func (m SubQuery) DatabaseName() string {
    return m.dbName
}

func (m SubQuery) Error() error {
    return m.err
}
