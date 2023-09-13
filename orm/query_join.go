package orm

import (
    "strings"
)

type JoinType string

const (
    JoinTypeInner JoinType = "inner join"
    JoinTypeLeft  JoinType = "left join"
    JoinTypeRight JoinType = "right join"
    JoinTypeOuter JoinType = "outer join"
)

func (q *Query[T]) Join(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return q.join(JoinTypeInner, table, where, alias...)
}
func (q *Query[T]) LeftJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return q.join(JoinTypeLeft, table, where, alias...)
}
func (q *Query[T]) RightJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return q.join(JoinTypeRight, table, where, alias...)
}
func (q *Query[T]) OuterJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return q.join(JoinTypeOuter, table, where, alias...)
}

func (q *Query[T]) join(joinType JoinType, table Table, wheref func(*Query[T]) *Query[T], alias ...string) *Query[T] {
    newTable, err := q.parseTable(table)
    if err != nil {
        return q.setErr(err)
    }

    if len(alias) > 0 {
        newTable.alias = alias[0]
    } else if newTable.rawSql != "" {
        newTable.alias = subqueryDefaultName
    }

    newTable.joinType = joinType
    q.tables = append(q.tables, newTable)
    q.tables[len(q.tables)-1].joinCondition, err = q.generateWhereGroup(wheref)
    return q.setErr(err)
}

func (q *Query[T]) generateTableAndJoinStr(tables []*queryTable, bindings *[]interface{}) string {
    if len(tables) == 0 {
        return ""
    }
    var tableStrs []string
    for k, v := range tables {
        tempStr := ""
        if v.rawSql == "" {
            if k == 0 {
                tempStr = v.getTableNameAndAlias()
            } else {
                tempStr = string(v.joinType)
                tempStr += " " + v.getTableNameAndAlias()
                if len(v.joinCondition.SubWheres) > 0 {
                    whereStr := q.generateWhereStr(v.joinCondition.SubWheres, bindings)
                    tempStr += " on " + whereStr
                }
            }
        } else {
            if k == 0 {
                tempStr = "(" + v.rawSql + ")"
                if v.getAlias() != "" {
                    tempStr += " " + v.getAlias()
                }
                *bindings = append(*bindings, v.bindings...)
            } else {
                tempStr = string(v.joinType)
                tempStr += " (" + v.rawSql + ")"
                if v.getAlias() != "" {
                    tempStr += " " + v.getAlias()
                }
                *bindings = append(*bindings, v.bindings...)
                if len(v.joinCondition.SubWheres) > 0 {
                    whereStr := q.generateWhereStr(v.joinCondition.SubWheres, bindings)
                    tempStr += " on " + whereStr
                }
            }
        }

        tableStrs = append(tableStrs, tempStr)
    }

    return strings.Join(tableStrs, " ")
}
