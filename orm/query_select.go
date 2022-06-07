package orm

type SelectForUpdateType string

const (
    SelectForUpdateTypeDefault    SelectForUpdateType = "for update"
    SelectForUpdateTypeNowait     SelectForUpdateType = "for update nowait"
    SelectForUpdateTypeSkipLocked SelectForUpdateType = "for update skip locked"
)

func (m Query[T]) Select(columns ...interface{}) Query[T] {
    m.columns = append(m.columns, columns...)
    return m
}

func (m Query[T]) ForUpdate(forUpdateType ...SelectForUpdateType) Query[T] {
    if len(forUpdateType) == 0 {
        m.forUpdate = SelectForUpdateTypeDefault
    } else {
        m.forUpdate = forUpdateType[0]
    }
    return m
}
