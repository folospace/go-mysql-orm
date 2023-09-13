package orm

func (q *Query[T]) WithWindow(subquery SubQuery, windowName string) *Query[T] {
    subquery.tableName = windowName
    q.windows = append(q.windows, subquery)
    return q
}
