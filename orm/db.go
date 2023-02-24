package orm

import (
    "database/sql"
    sqldriver "database/sql/driver"
    _ "github.com/go-sql-driver/mysql"
)

func OpenMysql(dataSourceName string) (*sql.DB, error) {
    return sql.Open("mysql", dataSourceName)
}

func Open(driverName, dataSourceName string) (*sql.DB, error) {
    return sql.Open(driverName, dataSourceName)
}

func OpenDB(driver sqldriver.Connector) *sql.DB {
    return sql.OpenDB(driver)
}

func Register(name string, drvier sqldriver.Driver) {
    sql.Register(name, drvier)
}
