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

func (m *Query[T]) Join(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return m.join(JoinTypeInner, table, where, alias...)
}
func (m *Query[T]) LeftJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return m.join(JoinTypeLeft, table, where, alias...)
}
func (m *Query[T]) RightJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return m.join(JoinTypeRight, table, where, alias...)
}
func (m *Query[T]) OuterJoin(table Table, where func(join *Query[T]) *Query[T], alias ...string) *Query[T] {
    return m.join(JoinTypeOuter, table, where, alias...)
}

func (m *Query[T]) join(joinType JoinType, table Table, wheref func(*Query[T]) *Query[T], alias ...string) *Query[T] {
    newTable, err := m.parseTable(table)
    if err != nil {
        return m.setErr(err)
    }

    if len(alias) > 0 {
        newTable.alias = alias[0]
    } else if newTable.rawSql != "" {
        newTable.alias = subqueryDefaultName
    }

    newTable.joinType = joinType
    m.tables = append(m.tables, newTable)
    m.tables[len(m.tables)-1].joinCondition, err = m.generateWhereGroup(wheref)
    return m.setErr(err)
}

func (m *Query[T]) generateTableAndJoinStr(tables []*queryTable, bindings *[]interface{}) string {
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
                    whereStr := m.generateWhereStr(v.joinCondition.SubWheres, bindings)
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
                    whereStr := m.generateWhereStr(v.joinCondition.SubWheres, bindings)
                    tempStr += " on " + whereStr
                }
            }
        }

        tableStrs = append(tableStrs, tempStr)
    }

    return strings.Join(tableStrs, " ")
}
