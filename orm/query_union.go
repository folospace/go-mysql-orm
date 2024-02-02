package orm

func (q *Query[T]) Union(subquery *SubQuery) *Query[T] {
    return q.union(false, subquery)
}
func (q *Query[T]) UnionAll(subquery *SubQuery) *Query[T] {
    return q.union(true, subquery)
}
func (q *Query[T]) union(isAll bool, subquery *SubQuery) *Query[T] {
    subquery.unionAll = isAll
    q.setErr(subquery.err)
    q.unions = append(q.unions, subquery)
    return q
}
