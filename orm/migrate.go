package orm

import (
    "errors"
    "reflect"
    "strings"
)

type dBColumn struct {
    Name          string   // `id`
    Type          string   //bigint //varchar(255)
    Null          bool     //null //not null
    AutoIncrement bool     //auto_increment
    Default       string   //default ''
    Comment       string   //comment ''
    Primary       bool     //primary key
    Index         []string //key
    Unique        []string //unique key
}

func (m *Query) Migrate() error {
    db := m.DB()
    if db == nil {
        return errors.New("db not exist")
    }

    if len(m.tables) == 0 || len(m.tables[0].jsonFields) == 0 {
        return errors.New("table not exist")
    }

    createTableSql := ""

    table := m.tables[0]
    for i := 0; i < table.tableStruct.NumField(); i++ {
        varField := table.tableStruct.Field(i)
        fieldIsPtr := false

        if varField.CanSet() == false {
            continue
        }
        if varField.Kind() == reflect.Ptr {
            fieldIsPtr = true
            if varField.Elem().Kind() == reflect.Ptr {
                continue
            }
        }

        column := dBColumn{}

        column.Name = table.getTags(i, "json")[0]
        column.Type = table.getTags(i, "type")[0]
    }

    _, err := db.Exec(createTableSql)
    return err
}
