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

func (q *Query[T]) SelectRank(column interface{}, alias string) *Query[T] {
    return q.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderBy(column)
    }, alias)
}

func (q *Query[T]) SelectRankDesc(column interface{}, alias string) *Query[T] {
    return q.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderByDesc(column)
    }, alias)
}

func (q *Query[T]) SelectOver(windowFunc string, f func(query *Query[T]) *Query[T], alias string) *Query[T] {
    partitionStart := len(q.partitionbys)
    orderStart := len(q.orderbys)
    nq := *q
    f(&nq)
    partitions := nq.partitionbys[partitionStart:]
    orders := nq.orderbys[orderStart:]

    q.setErr(nq.result.Err)

    newSelect := windowFunc + " over ("
    if len(partitions) > 0 {
        newSelect += "partition by " + strings.Join(partitions, ",") + " "
    }
    if len(orders) > 0 {
        newSelect += "order by " + strings.Join(orders, ",")
    }
    newSelect += ")"

    newSelect += " as " + alias

    q.columns = append(q.columns, newSelect)
    return q
}

func (q *Query[T]) SelectOverRaw(windowFunc string, windowName string, alias string) *Query[T] {
    newSelect := windowFunc + " over " + windowName + " as " + alias
    q.columns = append(q.columns, newSelect)
    return q
}

func (q *Query[T]) Select(columns ...interface{}) *Query[T] {
    q.columns = append(q.columns, columns...)
    return q
}

func (q *Query[T]) ForUpdate(forUpdateType ...SelectForUpdateType) *Query[T] {
    if len(forUpdateType) == 0 {
        q.forUpdate = SelectForUpdateTypeDefault
    } else {
        q.forUpdate = forUpdateType[0]
    }
    return q
}

func (q *Query[T]) SelectWithTimeout(duration time.Duration) *Query[T] {
    ms := duration.Milliseconds()
    if ms > 0 {
        q.selectTimeout = "/*+ MAX_EXECUTION_TIME(+" + strconv.FormatInt(ms, 10) + "+) */"
    }
    return q
}
