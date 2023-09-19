package orm

import "database/sql"

type Table interface {
    Connections() []*sql.DB
    DatabaseName() string
    TableName() string
}
