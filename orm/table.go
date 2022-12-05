package orm

type Table interface {
    TableName() string
    DatabaseName() string
}
