package orm

func (q *Query[T]) Raw(prepareSql string, bindings ...interface{}) *Query[T] {
    q.prepareSql = prepareSql
    q.bindings = bindings
    return q
}
