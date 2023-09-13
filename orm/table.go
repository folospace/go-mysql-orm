package orm

import "database/sql"

type Table interface {
    Connection() []*sql.DB
    DatabaseName() string
    TableName() string
}
