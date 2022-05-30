package orm

type Table interface {
    //Query() *Query
    TableName() string
    DatabaseName() string
}
