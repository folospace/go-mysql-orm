package orm

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
