package orm

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

const primaryKeyPrefix = "primary"
const uniqueKeyPrefix = "unique"
const keyPrefix = "index"
const nullPrefix = "null"
const autoIncrementPrefix = "auto_increment"
const createdAtColumn = "created_at"
const updatedAtColumn = "updated_at"
const deletedAtColumn = "deleted_at"

var definedDefault = []string{"null", "current_timestamp", "current_timestamp on update current_timestamp"}

var tagSplitRegex = regexp.MustCompile(`[^\s"]+|"([^"]*)"`)

//a := r.FindAllString(s, -1)

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

func CreateTableFromStruct[T Table](query Query[T]) (string, error) {
	originQuery := query
	db := originQuery.DB()
	if db == nil {
		return "", errors.New("no db exist")
	}

	if len(originQuery.tables) == 0 || len(originQuery.tables[0].ormFields) == 0 ||
		originQuery.tables[0].table == nil || originQuery.tables[0].table.TableName() == "" {
		return "", errors.New("no table exist")
	}

	dbColums := getMigrateColumns(originQuery.tables[0])
	if len(dbColums) == 0 {
		return "", errors.New("no column exist")
	}

	dbColumnStrs := generateColumnStrings(dbColums)

	originColumnStrs, _ := getSqlSegments(query)

	if len(originColumnStrs) > 0 {
		extraStrs := getTableNewColumns(originColumnStrs, dbColumnStrs)
		retSql := ""
		var err error
		for _, v := range extraStrs {
			tempSql := "ALTER TABLE " + query.TableInterface().TableName() + " ADD " + v

			retSql += tempSql + "\n"
			_, err = db.Exec(tempSql)
			if err != nil {
				break
			}
		}
		return retSql, err
	} else {
		createTableSql := fmt.Sprintf("create table IF NOT EXISTS `%s` (%s)",
			originQuery.tables[0].table.TableName(),
			strings.Join(dbColumnStrs, ","))

		_, err := db.Exec(createTableSql)
		return createTableSql, err

	}
}

func getTableNewColumns(origin, current []string) []string {
	var exist = make(map[string]bool)
	for _, v := range origin {
		v := strings.Trim(v, " ")
		if strings.HasPrefix(v, "`") {
			tempStrs := strings.SplitN(v, " ", 2)
			exist[strings.ToLower(tempStrs[0])] = true
		} else {
			tempStrs := strings.SplitN(v, " (", 2)
			exist[strings.ToLower(tempStrs[0])] = true
		}
	}

	var ret []string
	for _, v := range current {
		if strings.HasPrefix(v, "`") {
			tempStrs := strings.SplitN(v, " ", 2)
			if exist[strings.ToLower(tempStrs[0])] == false {
				ret = append(ret, v)
			}
		} else {
			tempStrs := strings.SplitN(v, " (", 2)
			if exist[strings.ToLower(tempStrs[0])] == false {
				ret = append(ret, v)
			}
		}
	}

	return ret
}

func generateColumnStrings(dbColums []dBColumn) []string {
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
			words = append(words, "default "+v.Default)
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
				uniqueComps[v2] = append(uniqueComps[v2], v.Name)
			}
		}

		if len(v.Indexs) > 0 {
			for _, v2 := range v.Indexs {
				indexComps[v2] = append(indexComps[v2], v.Name)
			}
		}
		ret = append(ret, strings.Join(words, " "))
	}
	if primaryStr != "" {
		ret = append(ret, primaryStr)
	}
	for _, v := range uniqueColumns {
		ret = append(ret, v)
	}

	for _, v := range indexColumns {
		ret = append(ret, v)
	}
	for k, v := range uniqueComps {
		sort.SliceStable(v, func(i, j int) bool {
			return strings.Index(k, v[i]) < strings.Index(k, v[j])
		})
		ret = append(ret, fmt.Sprintf("unique key `%s` (%s)", k, "`"+strings.Join(v, "`,`")+"`"))
	}
	for k, v := range indexComps {
		sort.SliceStable(v, func(i, j int) bool {
			return strings.Index(k, v[i]) < strings.Index(k, v[j])
		})
		ret = append(ret, fmt.Sprintf("key `%s` (%s)", k, "`"+strings.Join(v, "`,`")+"`"))
	}
	return ret
}

func getMigrateColumns(table *queryTable) []dBColumn {
	var ret []dBColumn
	for i := 0; i < table.tableStruct.NumField(); i++ {
		varField := table.tableStruct.Field(i)

		if varField.CanSet() == false {
			continue
		}

		column := dBColumn{}

		ormTags := stringSplitEscapeParentheses(table.getTag(i, "orm"), ",")
		if ormTags[0] != "" {
			column.Name = ormTags[0]
		} else {
			column.Name = table.getTags(i, "json")[0]
		}

		if column.Name == "" || column.Name == "-" {
			continue
		}

		kind := varField.Kind()
		if varField.Kind() == reflect.Ptr {
			kind = varField.Elem().Kind()
			if varField.Elem().Kind() == reflect.Ptr {
				continue
			}
			column.Null = true
		}

		column.Type, column.Default = getTypeAndDefault(varField)

		if i == 0 {
			column.Primary = true
			if column.Default == "0" {
				column.AutoIncrement = true
			}
		}

		if column.Name == createdAtColumn {
			column.Type = "timestamp"
			column.Default = "CURRENT_TIMESTAMP"
		} else if column.Name == updatedAtColumn {
			column.Type = "timestamp"
			column.Default = "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
		} else if column.Name == deletedAtColumn {
			column.Null = true
			column.Type = "timestamp"
			column.Default = "Null"
		}

		column.Comment = table.getTags(i, "comment")[0]
		customDefault := table.getTags(i, "default")[0]
		if customDefault != "" {
			column.Default = customDefault
			if kind == reflect.Bool {
				if strings.ToLower(customDefault) == "true" {
					column.Default = "1"
				} else if strings.ToLower(customDefault) == "false" {
					column.Default = "0"
				}
			}
		}

		if ormTags[0] != "" {
			overideColumn := dBColumn{}

			for k, v := range ormTags {
				if k == 0 {
					continue
				}
				if v == nullPrefix {
					overideColumn.Null = true
				} else if v == autoIncrementPrefix {
					overideColumn.AutoIncrement = true
				} else if strings.HasPrefix(v, primaryKeyPrefix) {
					overideColumn.Primary = true
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
					overideColumn.Type = v
				}
			}

			column.Null = overideColumn.Null
			column.AutoIncrement = overideColumn.AutoIncrement
			column.Primary = overideColumn.Primary
			if overideColumn.Type != "" {
				column.Type = overideColumn.Type
			}
		}

		if column.Null {
			if customDefault == "" {
				column.Default = "null"
			}
		}

		if column.Default == "" || SliceContain(definedDefault, strings.ToLower(column.Default)) < 0 {
			column.Default = "'" + column.Default + "'"
		}

		if column.Null == false && strings.ToLower(customDefault) == "null" {
			column.Default = ""
		}

		ret = append(ret, column)
	}

	return ret
}

func getTypeAndDefault(val reflect.Value) (string, string) {
	var types, defaults string
	kind := val.Kind()
	if kind == reflect.Ptr {
		kind = val.Elem().Kind()
	}
	switch kind {
	case reflect.Bool, reflect.Int8:
		types = "tinyint"
		defaults = "0"
	case reflect.Int16:
		types = "smallint"
		defaults = "0"
	case reflect.Int, reflect.Int32:
		types = "int"
		defaults = "0"
	case reflect.Int64:
		types = "bigint"
		defaults = "0"
	case reflect.Uint8:
		types = "tinyint unsigned"
		defaults = "0"
	case reflect.Uint16:
		types = "smallint unsigned"
		defaults = "0"
	case reflect.Uint, reflect.Uint32:
		types = "int unsigned"
		defaults = "0"
	case reflect.Uint64:
		types = "bigint unsigned"
		defaults = "0"
	case reflect.String:
		types = "varchar(255)"
	default:
		if _, ok := val.Interface().(*time.Time); ok {
			types = "timestamp"
		} else if _, ok := val.Interface().(time.Time); ok {
			types = "timestamp"
		} else {
			types = "varchar(255)"
		}
	}
	return types, defaults
}

func stringSplitEscapeParentheses(s string, seperator string) []string {
	var splits []string
	var start = "("
	var end = ")"

	var openP int
	var before string
	for i, v := range s {
		temp := string(v)
		if temp == seperator && (openP == 0 || strings.Contains(s[i:], end) == false) {
			if before != "" {
				splits = append(splits, before)
			}
			before = ""
		} else {
			if temp == start {
				openP += 1
			} else if temp == end {
				if openP > 0 {
					openP -= 1
				}
			}
			before += temp
		}
	}
	if before != "" {
		splits = append(splits, before)
	}
	if len(splits) == 0 {
		splits = append(splits, "")
	}
	return splits
}
