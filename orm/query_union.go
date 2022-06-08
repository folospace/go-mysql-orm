package orm

func (m Query[T]) Union(subquery SubQuery) Query[T] {
    return m.union(false, subquery)
}
func (m Query[T]) UnionAll(subquery SubQuery) Query[T] {
    return m.union(true, subquery)
}
func (m Query[T]) union(isAll bool, subquery SubQuery) Query[T] {
    return m
}
