package main

import (
    "database/sql"
    "github.com/folospace/go-mysql-orm/orm"
    _ "github.com/go-sql-driver/mysql"
    "testing"
)

var tdb, _ = sql.Open("mysql", "rfamro@tcp(mysql-rfam-public.ebi.ac.uk:4497)/Rfam?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

func TestSelect(t *testing.T) {
    t.Run("query_raw", func(t *testing.T) {

        var data map[string][]string
        query := orm.NewQueryRaw("family", tdb).Select("type", "author").Limit(5).GetTo(&data)
        t.Log(data)
        t.Log(query.Sql())
        t.Log(query.Error())
    })
}
