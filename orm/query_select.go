package orm

func (m Query[T]) SelectForUpdate(columns ...interface{}) Query[T] {
    m.columns = append(m.columns, columns...)
    m.forUpdate = true
    return m
}

func (m Query[T]) Select(columns ...interface{}) Query[T] {
    m.columns = append(m.columns, columns...)
    return m
}
