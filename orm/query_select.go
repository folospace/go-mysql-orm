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

func (q *Query[T]) SelectRank(column any, alias string) *Query[T] {
    return q.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderBy(column)
    }, alias)
}

func (q *Query[T]) SelectRankDesc(column any, alias string) *Query[T] {
    return q.SelectOver("rank()", func(query *Query[T]) *Query[T] {
        return query.OrderByDesc(column)
    }, alias)
}

func (q *Query[T]) SelectRowNumber(column any, alias string) *Query[T] {
    return q.SelectOver("row_number()", func(query *Query[T]) *Query[T] {
        return query.OrderBy(column)
    }, alias)
}

func (q *Query[T]) SelectRowNumberDesc(column any, alias string) *Query[T] {
    return q.SelectOver("row_number()", func(query *Query[T]) *Query[T] {
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

func (q *Query[T]) Select(columns ...any) *Query[T] {
    q.columns = append(q.columns, columns...)
    return q
}

func (q *Query[T]) SelectExclude(exceptColumns ...any) *Query[T] {
    q.columns = nil

    for i := 0; i < q.tables[0].tableStruct.NumField(); i++ {
        field := q.tables[0].tableStruct.Field(i)
        if field.CanAddr() == false || field.Addr().CanInterface() == false {
            continue
        }
        addr := field.Addr().Interface()

        if v, ok := q.tables[0].ormFields[addr]; ok {
            except := false
            for _, v2 := range exceptColumns {
                if v2 == v || v2 == addr {
                    except = true
                    break
                }
            }
            if except == false {
                q.columns = append(q.columns, addr)
            }
        }
    }

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
        q.selectTimeout = "/*+ MAX_EXECUTION_TIME(" + strconv.FormatInt(ms, 10) + ") */"
    }
    return q
}
