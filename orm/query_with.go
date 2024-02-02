package orm

import (
    "context"
    "strings"
)

func (q *Query[T]) WithParentsOnColumn(pidColumn interface{}) *Query[T] {
    tempName := q.TableInterface().TableName() + "_cte"

    col, err := q.parseColumn(pidColumn)
    if err != nil {
        return q.setErr(err)
    }
    coln := strings.Split(col, ".")
    newcol := strings.Trim(coln[len(coln)-1], "`")

    cte := newQueryRaw(tempName, q.DBs()...)

    appendQuery := NewQuery(q.T, q.DBs()...)
    appendQuery = appendQuery.Join(cte.T, func(query *Query[T]) *Query[T] {
        return query.Where(appendQuery.tables[0].tableStruct.Field(0).Addr().Interface(), Raw(tempName+"."+newcol))
    }).Select(appendQuery.AllCols())

    q.self = cte
    return q.UnionAll(appendQuery.SubQuery())
}

func (q *Query[T]) WithChildrenOnColumn(pidColumn interface{}) *Query[T] {
    tempName := q.TableInterface().TableName() + "_cte"

    pcol, err := q.parseColumn(pidColumn)
    if err != nil {
        return q.setErr(err)
    }
    if strings.Contains(pcol, ".") == false {
        pcol = q.TableInterface().TableName() + "." + pcol
    }
    col, err := q.parseColumn(q.tables[0].tableStruct.Field(0).Addr().Interface())
    if err != nil {
        return q.setErr(err)
    }
    coln := strings.Split(col, ".")
    newcol := strings.Trim(coln[len(coln)-1], "`")

    cte := newQueryRaw(tempName, q.DBs()...)

    appendQuery := NewQuery(q.T, q.DBs()...)
    appendQuery = appendQuery.Join(cte.T, func(query *Query[T]) *Query[T] {
        return query.Where(pcol, Raw(tempName+"."+newcol))
    }).Select(appendQuery.AllCols())

    q.self = cte
    return q.UnionAll(appendQuery.SubQuery())
}

func (q *Query[T]) WithCte(subquery *SubQuery, cteName string, columns ...string) *Query[T] {
    return q.withCte(subquery, cteName, false, columns...)
}

func (q *Query[T]) WithRecursiveCte(subquery *SubQuery, cteName string, columns ...string) *Query[T] {
    return q.withCte(subquery, cteName, true, columns...)
}

func (q *Query[T]) withCte(subquery *SubQuery, cteName string, recursive bool, columns ...string) *Query[T] {
    subquery.tableName = cteName
    subquery.recursive = recursive
    subquery.columns = columns
    q.setErr(subquery.err)
    q.withCtes = append(q.withCtes, subquery)
    return q
}

func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
    q.ctx = &ctx
    return q
}
