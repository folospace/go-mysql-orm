package orm

func NewCte(cteName string, recursive bool, columns ...string) Query[SubQuery] {
    sq := SubQuery{
        tableName: cteName,
        recursive: recursive,
        columns:   columns,
    }
    return NewQuery(sq, nil)
}

func (m Query[T]) CteAs(subquery SubQuery) Query[T] {
    return m.union(false, subquery)
}

func (m Query[T]) With(subquery SubQuery) Query[T] {
    return m.union(false, subquery)
}
