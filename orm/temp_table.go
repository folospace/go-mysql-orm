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

func (m *tempTable) Query() *Query {
    return new(Query).UseDB(m.db).UseTx(m.tx).FromTable(m)
}

func (*tempTable) TableName() string {
    return "temp"
}

func (m *tempTable) DatabaseName() string {
    return m.dbName
}

func (m *tempTable) Error() error {
    return m.err
}
