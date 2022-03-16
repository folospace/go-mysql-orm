package orm

import (
    "errors"
    "fmt"
    "reflect"
    "strings"
)

const primaryKeyPrefix = "keyp"
const uniqueKeyPrefix = "keyu"
const keyPrefix = "key"
const nullPrefix = "null"
const autoIncrementPrefix = "ai"
const createdAtColumn = "created_at"
const updatedAtColumn = "updated_at"
const deletedAtColumn = "deleted_at"

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

func (m *Query) Migrate() (string, error) {
    db := m.DB()
    if db == nil {
        return "", errors.New("no db exist")
    }

    if len(m.tables) == 0 || len(m.tables[0].jsonFields) == 0 ||
        m.tables[0].table == nil || m.tables[0].table.TableName() == "" {
        return "", errors.New("no table exist")
    }

    dbColums := m.getMigrateColumns(m.tables[0])
    if len(dbColums) == 0 {
        return "", errors.New("no column exist")
    }

    dbColumnStrs := m.generateColumnStrings(dbColums)

    createTableSql := fmt.Sprintf("create table IF NOT EXISTS  `%s` (%s)",
        m.tables[0].table.TableName(),
        strings.Join(dbColumnStrs, ","))

    _, err := db.Exec(createTableSql)
    return createTableSql, err
}

func (m *Query) generateColumnStrings(dbColums []dBColumn) []string {
    var ret []string
    var primaryStr string
    var uniqueColumns []string
    var indexColumns []string
    var uniqueComps = make(map[string][]string)
    var indexComps = make(map[string][]string)

    for _, v := range dbColums {
        var words []string
        //add column name
        words = append(words, "`"+v.Name+"`")
        //add type
        words = append(words, v.Type)

        //add null
        if v.Null {
            words = append(words, "null")
        } else {
            words = append(words, "not null")
        }

        //add default
        if v.AutoIncrement {
            words = append(words, "auto_increment")
        } else if v.Default != "" {
            words = append(words, "default "+"'"+v.Default+"'")
        }

        //add comment
        if v.Comment != "" {
            words = append(words, "comment "+"'"+v.Comment+"'")
        }

        if v.Primary {
            primaryStr = fmt.Sprintf("primary key (%s)", "`"+v.Name+"`")
        } else if v.Unique {
            uniqueColumns = append(uniqueColumns, fmt.Sprintf("unique key `%s` (`%s`)", v.Name, v.Name))
        } else if v.Index {
            indexColumns = append(indexColumns, fmt.Sprintf("key `%s` (`%s`)", v.Name, v.Name))
        }

        if len(v.Uniques) > 0 {
            for _, v2 := range v.Uniques {
                uniqueComps[v2] = append(uniqueComps[v2], "`"+v.Name+"`")
            }
        }

        if len(v.Indexs) > 0 {
            for _, v2 := range v.Indexs {
                indexComps[v2] = append(indexComps[v2], "`"+v.Name+"`")
            }
        }
        ret = append(ret, strings.Join(words, " "))
    }
    if primaryStr != "" {
        ret = append(ret, primaryStr)
    }
    for k, v := range uniqueComps {
        ret = append(ret, fmt.Sprintf("unique key `%s` (%s)", k, strings.Join(v, ",")))
    }
    for k, v := range indexComps {
        ret = append(ret, fmt.Sprintf("key `%s` (%s)", k, strings.Join(v, ",")))
    }
    return ret
}

func (m *Query) getMigrateColumns(table *queryTable) []dBColumn {
    var ret []dBColumn
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

        column.Comment = table.getTags(i, "comment")[0]
        column.Default = table.getTags(i, "default")[0]

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
            if column.Name == createdAtColumn {
                column.Type = "timestamp"
                if column.Default == "" {
                    column.Default = "CURRENT_TIMESTAMP"
                }
            } else if column.Name == updatedAtColumn {
                column.Type = "timestamp"
                if column.Default == "" {
                    column.Default = "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
                }
            } else if column.Name == deletedAtColumn {
                column.Type = "timestamp"
                column.Null = true
                column.Default = "Null"
            } else {
                switch columnKind {
                case reflect.Bool, reflect.Int8:
                    column.Type = "tinyint"
                case reflect.Int16:
                    column.Type = "smallinit"
                case reflect.Int, reflect.Int32:
                    column.Type = "int"
                case reflect.Int64:
                    column.Type = "bigint"
                case reflect.Uint8:
                    column.Type = "tinyint unsigned"
                case reflect.Uint16:
                    column.Type = "smallinit unsigned"
                case reflect.Uint, reflect.Uint32:
                    column.Type = "int unsigned"
                case reflect.Uint64:
                    column.Type = "bigint unsigned"
                case reflect.String:
                    column.Type = "varchar(255)"
                default:
                    column.Type = "varchar(255)"
                }
            }
        }
        ret = append(ret, column)
    }

    return ret
}
