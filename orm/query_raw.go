package orm

func (q *Query[T]) Raw(prepareSql string, bindings ...any) *Query[T] {
    q.prepareSql = prepareSql
    q.bindings = bindings
    return q
}
