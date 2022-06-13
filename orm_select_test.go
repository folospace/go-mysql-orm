package main

import (
	"database/sql"
	"github.com/folospace/go-mysql-orm/orm"
	"testing"
)

var tdb, _ = sql.Open("mysql", "rfamro@tcp(mysql-rfam-public.ebi.ac.uk:4497)/Rfam?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

func TestSelect(t *testing.T) {
	t.Run("query_raw", func(t *testing.T) {
		data, query := orm.NewQueryRaw(tdb, "family").Limit(5).GetRows()
		t.Log(data)
		t.Log(query.Sql())
		t.Log(query.Error())
	})
}
