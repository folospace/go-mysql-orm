package orm

import (
    "errors"
    "reflect"
    "strings"
)

const primaryKeyPrefix = "keyp"
const uniqueKeyPrefix = "keyu"
const keyPrefix = "key"
const nullPrefix = "null"
const autoIncrementPrefix = "ai"

type dBColumn struct {
    Name          string // `id`
    Type          string //bigint //varchar(255)
    Null          bool   //null //not null
    AutoIncrement bool   //auto_increment
    Primary       bool
    Unique        bool
    Index         bool

    Default string   //default ''
    Comment string   //comment ''
    Indexs  []string //composite index names
    Uniques []string //composite unique index names
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

        if varField.CanSet() == false {
            continue
        }

        column := dBColumn{}

        column.Name = table.getTags(i, "json")[0]
        if column.Name == "" {
            continue
        }

        columnKind := varField.Kind()
        if varField.Kind() == reflect.Ptr {
            column.Null = true
            columnKind = varField.Elem().Kind()
            if varField.Elem().Kind() == reflect.Ptr {
                continue
            }
        }

        if i == 0 {
            column.AutoIncrement = true
            column.Primary = true
        }

        column.Default = table.getTags(i, "default")[0]
        column.Comment = table.getTags(i, "comment")[0]

        ormTags := table.getTags(i, "orm")
        if ormTags[0] != "" {
            for _, v := range ormTags {
                if v == nullPrefix {
                    column.Null = true
                } else if v == autoIncrementPrefix {
                    column.AutoIncrement = true
                } else if strings.HasPrefix(v, primaryKeyPrefix) {
                    column.Primary = true
                } else if strings.HasPrefix(v, uniqueKeyPrefix) {
                    if v == uniqueKeyPrefix {
                        column.Unique = true
                    } else {
                        column.Uniques = append(column.Uniques, v)
                    }
                } else if strings.HasPrefix(v, keyPrefix) {
                    if v == keyPrefix {
                        column.Index = true
                    } else {
                        column.Indexs = append(column.Indexs, v)
                    }
                } else {
                    column.Type = v
                }
            }
        }
        if column.Type == "" {
            switch columnKind {
            case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                column.Type = "bigint"
            case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
                column.Type = "bigint unsigned"
            case reflect.String:
                column.Type = "varchar(255)"
            default:
                column.Type = "varchar(255)"
            }
        }
    }

    _, err := db.Exec(createTableSql)
    return err
}
