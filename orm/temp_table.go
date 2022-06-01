package orm

import "database/sql"

type SubQuery struct {
    raw      string
    bindings []interface{}
    dbName   string
    db       *sql.DB
    tx       *sql.Tx
    err      error
}

func (SubQuery) TableName() string {
    return "sub"
}

func (m SubQuery) DatabaseName() string {
    return m.dbName
}

func (m SubQuery) Error() error {
    return m.err
}
