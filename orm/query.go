package orm

import (
    "context"
    "database/sql"
    "github.com/mcuadros/go-defaults"
    "math/rand"
    "reflect"
    "strconv"
    "strings"
)

type Raw string

type Query[T Table] struct {
    writeAndReadDbs []*sql.DB //first element as write db, rest as read dbs
    tx              *sql.Tx
    ctx             *context.Context
    tables          []*queryTable
    wheres          []where
    result          QueryResult
    limit           int
    offset          int
    partitionbys    []string
    orderbys        []string
    forUpdate       SelectForUpdateType
    T               T
    columns         []any
    insertIgnore    bool
    conflictUpdates []updateColumn
    prepareSql      string
    bindings        []any
    groupBy         []any
    having          []where
    unions          []*SubQuery
    withCtes        []*SubQuery
    windows         []*SubQuery
    self            *Query[*SubQuery]
    selectTimeout   string
}

//query table[struct] generics
func NewQuery[T Table](t T, writeAndReadDbs ...*sql.DB) *Query[T] {
    q := Query[T]{T: t, writeAndReadDbs: writeAndReadDbs}
    return q.fromTable(q.T)
}

func (q *Query[T]) NewDefault() T {
    ret, _ := reflect.New(q.tables[0].tableStructType).Interface().(T)
    defaults.SetDefaults(ret)
    return ret
}

func (q *Query[T]) Clone() *Query[T] {
    var clone = *q
    return &clone
}

func (q *Query[T]) UseDB(db ...*sql.DB) *Query[T] {
    q.writeAndReadDbs = db
    return q
}

func (q *Query[T]) UseFirstDB() *Query[T] {
    q.writeAndReadDbs = q.tables[0].table.Connections()[:1]
    return q
}

func (q *Query[T]) UseTx(tx *sql.Tx) *Query[T] {
    q.tx = tx
    return q
}

func (q *Query[T]) DB() *sql.DB {
    return q.writeDB()
}

func (q *Query[T]) DBs() []*sql.DB {
    if len(q.writeAndReadDbs) == 0 && len(q.tables) > 0 {
        q.writeAndReadDbs = q.tables[0].table.Connections()
    }
    return q.writeAndReadDbs
}

func (q *Query[T]) Tx() *sql.Tx {
    return q.tx
}

func (q *Query[T]) Alias(alias string) *Query[T] {
    q.tables[0].alias = alias
    return q
}

//query raw, tablename can be empty
func newQueryRaw(tableName string, writeAndReadDbs ...*sql.DB) *Query[*SubQuery] {
    sq := &SubQuery{}
    if tableName != "" {
        sq.tableName = tableName
    }
    return NewQuery(sq, writeAndReadDbs...)
}

func (q *Query[T]) writeDB() *sql.DB {
    dbs := q.DBs()
    if len(dbs) > 0 {
        return dbs[0]
    }
    return nil
}
func (q *Query[T]) readDB() *sql.DB {
    dbs := q.DBs()
    if len(dbs) > 1 {
        return dbs[rand.Intn(len(dbs)-1)+1] //rand get db
    } else {
        return q.writeDB()
    }
}

func (q *Query[T]) tableInterface() Table {
    return any(q.T).(Table)
}

func (q *Query[T]) allCols() string {
    return q.tables[0].getAliasOrTableName() + ".*"
}

func (q *Query[T]) fromTable(table Table) *Query[T] {
    newTable, err := q.parseTable(table)
    if err != nil {
        return q.setErr(err)
    }

    if newTable.rawSql != "" {
        newTable.alias = subqueryDefaultName
    }
    q.tables = append(q.tables, newTable)
    return q
}

func (q *Query[T]) parseTable(table Table) (*queryTable, error) {
    var newTable *queryTable

    if temp, ok := table.(SubQuery); ok {
        newTable = &queryTable{
            table:    table,
            rawSql:   temp.raw,
            bindings: temp.bindings,
        }
        return newTable, nil
    } else if temp, ok := table.(*SubQuery); ok {
        newTable = &queryTable{
            table:    table,
            rawSql:   temp.raw,
            bindings: temp.bindings,
        }
        return newTable, nil
    } else {
        cached := getTableFromCache(table)
        if cached != nil {
            tmp := *cached
            return &tmp, nil
        }
        tableStructAddr := reflect.ValueOf(table)
        if tableStructAddr.Kind() != reflect.Ptr {
            return nil, ErrParamMustBePtr
        }
        //reset query vars
        tableStruct := tableStructAddr.Elem()
        if tableStruct.Kind() != reflect.Struct {
            return nil, ErrParamElemKindMustBeStruct
        }

        tableStructType := reflect.TypeOf(table).Elem()
        ormFields := make(map[any]string)

        for i := 0; i < tableStruct.NumField(); i++ {
            valueField := tableStruct.Field(i)

            ormTag := strings.Split(tableStructType.Field(i).Tag.Get("orm"), ",")[0]
            if ormTag == "-" {
                continue
            }
            if ormTag != "" {
                ormFields[valueField.Addr().Interface()] = ormTag
                continue
            }

            name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
            if name == "-" {
                continue
            }
            if name != "" {
                ormFields[valueField.Addr().Interface()] = name
            }
        }
        newTable = &queryTable{
            table:           table,
            tableStruct:     tableStruct,
            tableStructType: reflect.TypeOf(table).Elem(),
            ormFields:       ormFields,
        }
        cacheTable(table, newTable)

        tmp := *newTable
        return &tmp, nil
    }
}

func (q *Query[T]) isRaw(v any) (string, bool) {
    val, ok := v.(Raw)
    return string(val), ok
}

func (q *Query[T]) isOperator(v any) (string, bool) {
    val, ok := v.(WhereOperator)
    return string(val), ok
}

func (q *Query[T]) isStringOrRaw(v any) (string, bool) {
    val := reflect.ValueOf(v)

    if val.Kind() == reflect.String {
        return val.String(), true
    } else {
        return "", false
    }
}

func (q *Query[T]) parseColumn(v any) (string, error) {
    columnVar := reflect.ValueOf(v)
    if columnVar.Kind() == reflect.String {
        ret := columnVar.String()
        if ret == "*" && len(q.tables) > 0 {
            prefix := q.tables[0].getAliasOrTableName()
            if prefix != "" {
                prefix += "."
            }
            return prefix + ret, nil
        } else if ret == "" {
            return "", ErrColumnShouldBeStringOrPtr
        } else {
            return ret, nil
        }
    } else if columnVar.Kind() == reflect.Ptr && columnVar.Elem().CanAddr() {
        table, column := q.getTableColumn(columnVar)
        if table == nil {
            return "", ErrColumnNotExisted
        }
        prefix := table.getAliasOrTableName()
        if prefix != "" {
            prefix += "."
        }
        if column == "" {
            return prefix + "*", nil
        }

        return prefix + "`" + column + "`", nil
    } else {
        return "", ErrColumnShouldBeStringOrPtr
    }
}

func (q *Query[T]) getTableColumn(i reflect.Value) (*queryTable, string) {
    for _, t := range q.tables {
        if i.Interface() == t.table || (i.Elem().CanInterface() && i.Elem().Interface() == t.table) {
            return t, ""
        }
        if s, exist := t.ormFields[i.Elem().Addr().Interface()]; exist {
            return t, s
        }
    }
    return nil, ""
}

func (q *Query[T]) setErr(err error) *Query[T] {
    if err != nil {
        q.result.Err = err
    }
    return q
}

func (q *Query[T]) Limit(limit int) *Query[T] {
    q.limit = limit
    return q
}

func (q *Query[T]) Offset(offset int) *Query[T] {
    q.offset = offset
    return q
}

//should not use group by after order by
func (q *Query[T]) GroupBy(columns ...any) *Query[T] {
    q.groupBy = append(q.groupBy, columns...)
    return q
}

func (q *Query[T]) Having(column any, vals ...any) *Query[T] {
    oldWheres := q.wheres

    newQuery := q.where(false, column, vals...)

    newWheres := newQuery.wheres[len(oldWheres):]
    if len(newWheres) > 0 {
        newQuery.having = append(newQuery.having, newWheres...)
        newQuery.wheres = oldWheres
    }
    return newQuery
}

func (q *Query[T]) OrHaving(column any, vals ...any) *Query[T] {
    oldWheres := q.wheres

    newQuery := q.where(true, column, vals...)

    newWheres := newQuery.wheres[len(oldWheres):]
    if len(newWheres) > 0 {
        newQuery.having = append(newQuery.having, newWheres...)
        newQuery.wheres = oldWheres
    }
    return newQuery
}

func (q *Query[T]) PartitionBy(column any) *Query[T] {
    val, err := q.parseColumn(column)
    if err != nil {
        return q.setErr(err)
    }
    q.partitionbys = append(q.partitionbys, val)
    return q
}
func (q *Query[T]) OrderBy(column any) *Query[T] {
    val, err := q.parseColumn(column)
    if err != nil {
        return q.setErr(err)
    }
    q.orderbys = append(q.orderbys, val)
    return q
}
func (q *Query[T]) OrderByDesc(column any) *Query[T] {
    val, err := q.parseColumn(column)
    if err != nil {
        return q.setErr(err)
    }
    q.orderbys = append(q.orderbys, val+" desc")
    return q
}

func (q *Query[T]) getOrderAndLimitSqlStr() string {
    var ret []string
    if len(q.orderbys) > 0 {
        orderStr := "order by " + strings.Join(q.orderbys, ",")
        ret = append(ret, orderStr)
    }
    if q.limit > 0 {
        limitStr := "limit " + strconv.Itoa(q.limit)
        ret = append(ret, limitStr)
    }
    if q.offset > 0 {
        offsetStr := "offset " + strconv.Itoa(q.offset)
        ret = append(ret, offsetStr)
    }

    return strings.Join(ret, " ")
}
