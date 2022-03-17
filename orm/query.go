package orm

import (
    "database/sql"
    "errors"
    "reflect"
    "strconv"
    "strings"
)

type Raw string

var AllCols = "*"

type Query struct {
    db        *sql.DB
    tx        *sql.Tx
    tables    []*queryTable
    wheres    []where
    result    QueryResult
    limit     int
    offset    int
    orderbys  []string
    forUpdate bool
}

func (m *Query) UseDB(db *sql.DB) *Query {
    m.db = db
    return m
}
func (m *Query) UseTx(tx *sql.Tx) *Query {
    m.tx = tx
    return m
}

func (m *Query) DB() *sql.DB {
    return m.db
}
func (m *Query) dbTx() *sql.Tx {
    return m.tx
}

func (m *Query) FromTable(table Table, alias ...string) *Query {
    m.tables = nil
    m.wheres = nil
    m.orderbys = nil
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

func (m *Query) parseTable(table Table) (*queryTable, error) {
    tableStructAddr := reflect.ValueOf(table)
    if tableStructAddr.Kind() != reflect.Ptr {
        return nil, errors.New("params must be address of variable")
    }
    //reset query vars
    tableStruct := tableStructAddr.Elem()
    if tableStruct.Kind() != reflect.Struct {
        return nil, errors.New("obj must be struct")
    }

    temp, ok := table.(*tempTable)
    var newTable *queryTable
    if ok {
        newTable = &queryTable{
            table:    table,
            rawSql:   temp.raw,
            bindings: temp.bindings,
        }
    } else {
        tableStructType := reflect.TypeOf(table).Elem()
        jsonFields := make(map[interface{}]string)

        for i := 0; i < tableStruct.NumField(); i++ {
            valueField := tableStruct.Field(i)

            name := strings.Split(tableStructType.Field(i).Tag.Get("json"), ",")[0]
            if name == "-" {
                name = ""
            }
            jsonFields[valueField.Addr().Interface()] = name
        }
        newTable = &queryTable{
            table:           table,
            tableStruct:     tableStruct,
            tableStructType: reflect.TypeOf(table).Elem(),
            jsonFields:      jsonFields,
        }
    }
    return newTable, nil
}

func (m *Query) AliasTable(alias string) *Query {
    m.tables[0].alias = alias
    return m
}

func (m *Query) isRaw(v interface{}) (string, bool) {
    val, ok := v.(Raw)
    return string(val), ok
}

func (m *Query) isOperator(v interface{}) (string, bool) {
    val, ok := v.(WhereOperator)
    return string(val), ok
}

func (m *Query) isStringOrRaw(v interface{}) (string, bool) {
    val := reflect.ValueOf(v)

    if val.Kind() == reflect.String {
        return val.String(), true
    } else {
        return "", false
    }
}

func (m *Query) parseColumn(v interface{}) (string, error) {
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

func (m *Query) getTableColumn(i reflect.Value) (*queryTable, string) {
    for _, t := range m.tables {
        if s, exist := t.jsonFields[i.Elem().Addr().Interface()]; exist {
            return t, s
        }
    }
    return nil, ""
}

func (m *Query) setErr(err error) *Query {
    if err != nil {
        m.result.Err = err
    }
    return m
}

func (m *Query) Limit(limit int) *Query {
    m.limit = limit
    return m
}

func (m *Query) Offset(offset int) *Query {
    m.offset = offset
    return m
}

//should not use group by after order by
func (m *Query) GroupBy(columns ...interface{}) *Query {
    return m
}

func (m *Query) Having(where func(*Query)) *Query {
    return m
}

func (m *Query) OrderBy(column interface{}) *Query {
    val, err := m.parseColumn(column)
    if err != nil {
        return m.setErr(err)
    }
    m.orderbys = append(m.orderbys, val)
    return m
}
func (m *Query) OrderByDesc(column interface{}) *Query {
    val, err := m.parseColumn(column)
    if err != nil {
        return m.setErr(err)
    }
    m.orderbys = append(m.orderbys, val+" desc")
    return m
}

func (m *Query) getOrderAndLimitSqlStr() string {
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
