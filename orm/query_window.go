package orm

func (m Query[T]) WithWindow(subquery SubQuery, windowName string) Query[T] {
    subquery.tableName = windowName
    m.windows = append(m.windows, subquery)
    return m
}
