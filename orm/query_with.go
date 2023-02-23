package orm

import "strings"

func (m Query[T]) WithParentsOnColumn(pidColumn interface{}) Query[T] {
    tempName := m.TableInterface().TableName() + "_cte"

    col, err := m.parseColumn(pidColumn)
    if err != nil {
        return m.setErr(err)
    }
    coln := strings.Split(col, ".")
    newcol := strings.Trim(coln[len(coln)-1], "`")

    cte := NewQueryRaw(tempName, m.writeAndReadDbs...)

    appendQuery := NewQuery(*m.T, m.writeAndReadDbs...)
    appendQuery = appendQuery.Join(cte.T, func(query Query[T]) Query[T] {
        return query.Where(appendQuery.tables[0].tableStruct.Field(0).Addr().Interface(), Raw(tempName+"."+newcol))
    }).Select(appendQuery.AllCols())

    m.self = &cte
    return m.UnionAll(appendQuery.SubQuery())
}

func (m Query[T]) WithChildrenOnColumn(pidColumn interface{}) Query[T] {
    tempName := m.TableInterface().TableName() + "_cte"

    pcol, err := m.parseColumn(pidColumn)
    if err != nil {
        return m.setErr(err)
    }
    if strings.Contains(pcol, ".") == false {
        pcol = m.TableInterface().TableName() + "." + pcol
    }
    col, err := m.parseColumn(m.tables[0].tableStruct.Field(0).Addr().Interface())
    if err != nil {
        return m.setErr(err)
    }
    coln := strings.Split(col, ".")
    newcol := strings.Trim(coln[len(coln)-1], "`")

    cte := NewQueryRaw(tempName, m.writeAndReadDbs...)

    appendQuery := NewQuery(*m.T, m.writeAndReadDbs...)
    appendQuery = appendQuery.Join(cte.T, func(query Query[T]) Query[T] {
        return query.Where(pcol, Raw(tempName+"."+newcol))
    }).Select(appendQuery.AllCols())

    m.self = &cte
    return m.UnionAll(appendQuery.SubQuery())
}

func (m Query[T]) WithCte(subquery SubQuery, cteName string, columns ...string) Query[T] {
    return m.withCte(subquery, cteName, false, columns...)
}

func (m Query[T]) WithRecursiveCte(subquery SubQuery, cteName string, columns ...string) Query[T] {
    return m.withCte(subquery, cteName, true, columns...)
}

func (m Query[T]) withCte(subquery SubQuery, cteName string, recursive bool, columns ...string) Query[T] {
    subquery.tableName = cteName
    subquery.recursive = recursive
    subquery.columns = columns
    m.withCtes = append(m.withCtes, subquery)
    return m
}
