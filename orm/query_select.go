package orm

import (
    "strconv"
    "strings"
    "time"
)

type SelectForUpdateType string

const (
    SelectForUpdateTypeDefault    SelectForUpdateType = "for update"
    SelectForUpdateTypeNowait     SelectForUpdateType = "for update nowait"
    SelectForUpdateTypeSkipLocked SelectForUpdateType = "for update skip locked"
)

func (m *Query[T]) SelectRank(column interface{}, alias string) *Query[T] {
    return m.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderBy(column)
    }, alias)
}

func (m *Query[T]) SelectRankDesc(column interface{}, alias string) *Query[T] {
    return m.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderByDesc(column)
    }, alias)
}

func (m *Query[T]) SelectOver(windowFunc string, f func(query *Query[T]) *Query[T], alias string) *Query[T] {
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

func (m *Query[T]) SelectOverRaw(windowFunc string, windowName string, alias string) *Query[T] {
    newSelect := windowFunc + " over " + windowName + " as " + alias
    m.columns = append(m.columns, newSelect)
    return m
}

func (m *Query[T]) Select(columns ...interface{}) *Query[T] {
    m.columns = append(m.columns, columns...)
    return m
}

func (m *Query[T]) ForUpdate(forUpdateType ...SelectForUpdateType) *Query[T] {
    if len(forUpdateType) == 0 {
        m.forUpdate = SelectForUpdateTypeDefault
    } else {
        m.forUpdate = forUpdateType[0]
    }
    return m
}

func (m *Query[T]) SelectWithTimeout(duration time.Duration) *Query[T] {
    ms := duration.Milliseconds()
    if ms > 0 {
        m.selectTimeout = "/*+ MAX_EXECUTION_TIME(+" + strconv.FormatInt(ms, 10) + "+) */"
    }
    return m
}
