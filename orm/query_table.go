package orm

import "reflect"

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
    if q.rawSql != "" {
        if q.alias != "" {
            return q.alias
        }
        return q.table.TableName()
    } else {
        return q.alias
    }
}

func (q queryTable) getTableName() string {
    if q.alias != "" {
        return q.alias
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
