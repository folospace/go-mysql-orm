package orm

import "strings"

type SelectForUpdateType string

const (
	SelectForUpdateTypeDefault    SelectForUpdateType = "for update"
	SelectForUpdateTypeNowait     SelectForUpdateType = "for update nowait"
	SelectForUpdateTypeSkipLocked SelectForUpdateType = "for update skip locked"
)

func (m Query[T]) SelectRank(f func(query Query[T]) Query[T], alias string) Query[T] {
	windowFunc := "rank()"
	partitionStart := len(m.partitionbys)
	orderStart := len(m.orderbys)
	nq := f(m)
	partitions := nq.partitionbys[partitionStart:]
	orders := nq.orderbys[orderStart:]

	m.setErr(nq.result.Err)

	newSelect := windowFunc + " over ("
	if len(partitions) > 0 {
		newSelect += "partition by " + strings.Join(partitions, ",") + " "
	}
	if len(orders) > 0 {
		newSelect += "order by " + strings.Join(orders, ",")
	}
	newSelect += ")"

	newSelect += " as " + alias

	m.columns = append(m.columns, newSelect)
	return m
}

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
