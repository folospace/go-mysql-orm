package orm

import (
    "reflect"
    "strings"
)

type queryTable struct {
    table             Table
    tableStruct       reflect.Value
    tableStructType   reflect.Type
    ormFields         map[any]string
    joinType          JoinType //(left|right) join
    joinCondition     where
    alias             string
    overrideTableName string //override table name
    partition         string
    rawSql            string
    bindings          []any
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
    temp := q.getTableNamePartition()
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
    if q.overrideTableName != "" {
        return q.overrideTableName
    }
    if q.table.TableName() != "" {
        if q.table.DatabaseName() != "" {
            return q.table.DatabaseName() + "." + q.table.TableName()
        } else {
            return q.table.TableName()
        }
    }
    return ""
}

func (q queryTable) getTableNamePartition() string {
    if q.overrideTableName != "" {
        return q.overrideTableName
    }
    if q.table.TableName() != "" {
        if q.table.DatabaseName() != "" {
            return q.table.DatabaseName() + "." + q.table.TableName() + q.getPartition()
        } else {
            return q.table.TableName() + q.getPartition()
        }
    }
    return ""
}

func (q queryTable) getPartition() string {
    if q.partition == "" {
        return ""
    }
    return " partition(" + q.partition + ")"
}

func (q queryTable) getTags(index int, tagName string) []string {
    tags := strings.Split(q.tableStructType.Field(index).Tag.Get(tagName), ",")
    return tags
}
func (q queryTable) getTag(index int, tagName string) string {
    return q.tableStructType.Field(index).Tag.Get(tagName)
}
