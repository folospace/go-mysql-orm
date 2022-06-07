package orm

import (
    "database/sql"
    "errors"
    "reflect"
    "runtime"
    "strconv"
    "strings"
)

type Raw string

type Query[T Table] struct {
    db          *sql.DB
    tx          *sql.Tx
    tables      []*queryTable
    wheres      []where
    result      QueryResult
    limit       int
    offset      int
    orderbys    []string
    forUpdate   SelectForUpdateType
    T           *T
    columns     []interface{}
    prepareSql  string
    bindings    []interface{}
    groupBy     []interface{}
    having      []where
    curFileName string
}

func NewQuery[T Table](t T, db *sql.DB) Query[T] {
    q := Query[T]{T: &t, db: db}
    q.curFileName = q.currentFilename()
    return q.FromTable(q.TableInterface())
}

func (m Query[T]) TableInterface() Table {
    return interface{}(m.T).(Table)
}

func (m Query[T]) AllCols() string {
    return m.tables[0].getAliasOrTableName() + ".*"
}

func (m Query[T]) UseDB(db *sql.DB) Query[T] {
    m.db = db
    return m
}
func (m Query[T]) UseTx(tx *sql.Tx) Query[T] {
    m.tx = tx
    return m
}

func (m Query[T]) DB() *sql.DB {
    return m.db
}
func (m Query[T]) dbTx() *sql.Tx {
    return m.tx
}

func (m Query[T]) FromTable(table Table, alias ...string) Query[T] {
    m.tables = nil
    m.wheres = nil
    m.orderbys = nil
    m.columns = nil
    m.prepareSql = ""
    m.bindings = nil
    m.groupBy = nil
    m.having = nil
    m.limit, m.offset = 0, 0
    m.result = QueryResult{}

    newTable, err := m.parseTable(table)
    if err != nil {
        return m.setErr(err)
    }

    if len(alias) > 0 {
        newTable.alias = alias[0]
    } else if newTable.rawSql != "" {
        newTable.alias = "sub"
    }
    m.tables = append(m.tables, newTable)
    return m
}

func (m Query[T]) parseTable(table Table) (*queryTable, error) {
    cached := getTableFromCache(table)
    if cached != nil {
        return cached, nil
    }
    tableStructAddr := reflect.ValueOf(table)
    if tableStructAddr.Kind() != reflect.Ptr {
        return nil, errors.New("params must be address of variable")
    }
    //reset query vars
    tableStruct := tableStructAddr.Elem()
    if tableStruct.Kind() != reflect.Struct {
        return nil, errors.New("obj must be struct")
    }

    temp, ok := table.(*SubQuery)
    var newTable *queryTable
    if ok {
        newTable = &queryTable{
            table:    table,
            rawSql:   temp.raw,
            bindings: temp.bindings,
        }
    } else {
        tableStructType := reflect.TypeOf(table).Elem()
        ormFields := make(map[interface{}]string)

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
    }
    return newTable, nil
}

func (m Query[T]) Alias(alias string) Query[T] {
    m.tables[0].alias = alias
    return m
}

func (m Query[T]) isRaw(v interface{}) (string, bool) {
    val, ok := v.(Raw)
    return string(val), ok
}

func (m Query[T]) isOperator(v interface{}) (string, bool) {
    val, ok := v.(WhereOperator)
    return string(val), ok
}

func (m Query[T]) isStringOrRaw(v interface{}) (string, bool) {
    val := reflect.ValueOf(v)

    if val.Kind() == reflect.String {
        return val.String(), true
    } else {
        return "", false
    }
}

func (m Query[T]) parseColumn(v interface{}) (string, error) {
    columnVar := reflect.ValueOf(v)
    if columnVar.Kind() == reflect.String {
        ret := columnVar.String()
        if ret == "*" && len(m.tables) > 0 {
            prefix := m.tables[0].getAliasOrTableName()
            if prefix != "" {
                prefix += "."
            }
            return prefix + ret, nil
        } else {
            return ret, nil
        }
    } else if columnVar.Kind() == reflect.Ptr {
        table, column := m.getTableColumn(columnVar)
        if table == nil {
            return "", errors.New("column is not exist in table")
        }
        if column == "" {
            return "", errors.New("column is not exist in table " + table.table.TableName())
        }
        prefix := table.getAliasOrTableName()
        if prefix != "" {
            prefix += "."
        }
        return prefix + "`" + column + "`", nil
    } else {
        return "", errors.New("column should be either string or address of field of table")
    }
}

func (m Query[T]) getTableColumn(i reflect.Value) (*queryTable, string) {
    for _, t := range m.tables {
        if s, exist := t.ormFields[i.Elem().Addr().Interface()]; exist {
            return t, s
        }
    }
    return nil, ""
}

func (m *Query[T]) setErr(err error) Query[T] {
    if err != nil {
        m.result.Err = err
    }
    return *m
}

func (m Query[T]) Limit(limit int) Query[T] {
    m.limit = limit
    return m
}

func (m Query[T]) Offset(offset int) Query[T] {
    m.offset = offset
    return m
}

//should not use group by after order by
func (m Query[T]) GroupBy(columns ...interface{}) Query[T] {
    m.groupBy = append(m.groupBy, columns...)
    return m
}

func (m Query[T]) Having(column interface{}, vals ...interface{}) Query[T] {
    oldWheres := m.wheres

    newQuery := m.where(false, column, vals...)

    newWheres := newQuery.wheres[len(oldWheres):]
    if len(newWheres) > 0 {
        newQuery.having = append(newQuery.having, newWheres...)
        newQuery.wheres = oldWheres
    }
    return newQuery
}

func (m Query[T]) orHaving(column interface{}, vals ...interface{}) Query[T] {
    oldWheres := m.wheres

    newQuery := m.where(true, column, vals...)

    newWheres := newQuery.wheres[len(oldWheres):]
    if len(newWheres) > 0 {
        newQuery.having = append(newQuery.having, newWheres...)
        newQuery.wheres = oldWheres
    }
    return newQuery
}

func (m Query[T]) OrderBy(column interface{}) Query[T] {
    val, err := m.parseColumn(column)
    if err != nil {
        return m.setErr(err)
    }
    m.orderbys = append(m.orderbys, val)
    return m
}
func (m Query[T]) OrderByDesc(column interface{}) Query[T] {
    val, err := m.parseColumn(column)
    if err != nil {
        return m.setErr(err)
    }
    m.orderbys = append(m.orderbys, val+" desc")
    return m
}

func (m Query[T]) getOrderAndLimitSqlStr() string {
    var orderStr string
    if len(m.orderbys) > 0 {
        orderStr = "order by " + strings.Join(m.orderbys, ",")
    }
    var limitStr string
    if m.limit > 0 {
        limitStr = "limit " + strconv.Itoa(m.limit)
    }
    var offsetStr string
    if m.offset > 0 {
        offsetStr = "offset " + strconv.Itoa(m.offset)
    }

    var ret = orderStr
    if limitStr != "" {
        ret += " " + limitStr
        if offsetStr != "" {
            ret += " " + offsetStr
        }
    }

    return ret
}

func (m Query[T]) currentFilename() string {
    _, fs, _, _ := runtime.Caller(2)
    return fs
}
