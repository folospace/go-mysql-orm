package orm

func (m Query[T]) Raw(prepareSql string, bindings ...interface{}) Query[T] {
    m.prepareSql = prepareSql
    m.bindings = bindings
    return m
}
