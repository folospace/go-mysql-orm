package orm

import "database/sql"

type tempTable struct {
    raw      string
    bindings []interface{}
    dbName   string
    db       *sql.DB
    tx       *sql.Tx
    err      error
}

func (tempTable) TableName() string {
    return "sub"
}

func (m tempTable) DatabaseName() string {
    return m.dbName
}

func (m tempTable) Error() error {
    return m.err
}
