package orm

import (
    "strings"
)

func (m *Query) Join(table Table, where func(join *Query), alias ...string) *Query {
    return m.join(JoinTypeInner, table, where, alias...)
}
func (m *Query) LeftJoin(table Table, where func(join *Query), alias ...string) *Query {
    return m.join(JoinTypeLeft, table, where, alias...)
}
func (m *Query) RightJoin(table Table, where func(join *Query), alias ...string) *Query {
    return m.join(JoinTypeRight, table, where, alias...)
}

func (m *Query) join(joinType JoinType, table Table, wheref func(*Query), alias ...string) *Query {
    newTable, err := m.parseTable(table)
    if err != nil {
        return m.setErr(err)
    }

    if len(alias) > 0 {
        newTable.alias = alias[0]
    } else if newTable.rawSql != "" {
        newTable.alias = "sub"
    }

    newTable.joinType = joinType
    m.tables = append(m.tables, newTable)
    newTable.joinCondition = m.generateWhereGroup(wheref)
    return m
}

func (m *Query) generateTableAndJoinStr(tables []*queryTable, bindings *[]interface{}) string {
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
