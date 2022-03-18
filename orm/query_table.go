package orm

import (
    "reflect"
    "strings"
)

type JoinType string

const (
    JoinTypeInner JoinType = "join"
    JoinTypeLeft  JoinType = "left join"
    JoinTypeRight JoinType = "right join"
)

type queryTable struct {
    table           Table
    tableStruct     reflect.Value
    tableStructType reflect.Type
    jsonFields      map[interface{}]string
    joinType        JoinType //(left|right) join
    joinCondition   where
    alias           string
    rawSql          string
    bindings        []interface{}
}

func (q queryTable) getAlias() string {
    return q.alias
}

func (q queryTable) getAliasOrTableName() string {
    if q.alias != "" {
        return q.alias
    }
    return q.getTableName()
}

func (q queryTable) getTableNameAndAlias() string {
    var strs []string
    temp := q.getTableName()
    if temp != "" {
        strs = append(strs, temp)
    }
    temp = q.getAlias()
    if temp != "" {
        strs = append(strs, temp)
    }
    return strings.Join(strs, " ")
}

func (q queryTable) getTableName() string {
    if q.table.TableName() != "" {
        if q.table.DatabaseName() != "" {
            return q.table.DatabaseName() + "." + q.table.TableName()
        } else {
            return q.table.TableName()
        }
    }
    return ""
}

func (q queryTable) getTags(index int, tagName string) []string {
    tags := strings.Split(q.tableStructType.Field(index).Tag.Get(tagName), ",")
    return tags
}
