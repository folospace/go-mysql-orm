package orm

import (
    "errors"
    "reflect"
    "strings"
)

type WhereOperator Raw

const (
    WhereEqual          WhereOperator = "="
    WhereNotEqual       WhereOperator = "!="
    WhereGreatThan      WhereOperator = ">"
    WhereGreaterOrEqual WhereOperator = ">="
    WhereLessThan       WhereOperator = "<"
    WhereLessOrEqual    WhereOperator = "<="
    WhereIn             WhereOperator = "in"
    WhereNotIn          WhereOperator = "not in"
    WhereLike           WhereOperator = "like"
    WhereNotLike        WhereOperator = "not like"
    WhereRlike          WhereOperator = "rlike"
    WhereNotRlike       WhereOperator = "not rlike"
    WhereIsNull         WhereOperator = "is null"
    WhereIsNotNull      WhereOperator = "is not null"
)

func (m Query[T]) where(isOr bool, column interface{}, vals ...interface{}) Query[T] {
    if len(vals) > 2 {
        return m.setErr(errors.New("two many where-params"))
    }

    if len(vals) == 0 {
        c, ok := m.isStringOrRaw(column)
        if ok == false {
            return m.setErr(errors.New("where-param should be string while only 1 param exist"))
        }
        if c != "" {
            m.wheres = append(m.wheres, where{Raw: c, IsOr: isOr})
        } else {
            return m.setErr(errors.New("where-param should not be empty string"))
        }
    } else {
        c, err := m.parseColumn(column)
        if err != nil {
            return m.setErr(err)
        }
        operator := "="
        var val interface{}
        if len(vals) == 2 {
            operator2, ok := m.isStringOrRaw(vals[0])
            if ok == false {
                return m.setErr(errors.New("the second where-param should be operator as string"))
            }
            operator = operator2
            val = vals[1]
        } else {
            if vals[0] == nil {
                vals[0] = WhereIsNull
            }
            tempVal, ok := m.isOperator(vals[0])
            if ok {
                if tempVal != string(WhereIsNull) && tempVal != string(WhereIsNotNull) {
                    return m.setErr(errors.New("operator \"" + tempVal + "\" must have params"))
                }
                operator = ""
                val = Raw(tempVal)
            } else {
                val = vals[0]
            }
        }

        value, ok := m.isRaw(val)
        raw := ""
        var rawBindings []interface{}
        if ok {
            if operator != "" {
                operator += " "
            }
            raw = c + " " + operator + value
        } else {
            tempTable, ok := val.(SubQuery)
            if ok {
                if operator != "" {
                    operator += " "
                }
                raw = c + " " + operator + "(" + tempTable.raw + ")"
                rawBindings = append(rawBindings, tempTable.bindings...)
            } else {
                temp := reflect.ValueOf(val)
                if temp.Kind() == reflect.Slice && temp.Len() > 0 {
                    rawBindings = make([]interface{}, temp.Len())
                    rawCells := make([]string, temp.Len())

                    for i := 0; i < temp.Len(); i++ {
                        rawCells[i] = "?"
                        rawBindings[i] = temp.Index(i).Interface()
                    }

                    raw = c + " " + operator + " " + "(" + strings.Join(rawCells, ",") + ")"
                } else if temp.Kind() == reflect.Ptr {
                    rawColumn, err := m.parseColumn(val)
                    if err == nil {
                        raw = c + " " + operator + " " + rawColumn
                    } else {
                        return m.setErr(errors.New("Error where " + c + " " + operator + " ? val is invalid"))
                    }
                }
            }
        }
        m.wheres = append(m.wheres, where{Raw: raw, Column: c, Val: val, Operator: operator, IsOr: isOr, RawBindings: rawBindings})
    }
    return m
}

func (m Query[T]) generateWhereStr(wheres []where, bindings *[]interface{}) string {
    var whereStr []string
    for k, v := range wheres {
        tempStr := ""
        if k > 0 {
            if v.IsOr {
                tempStr = "or "
            } else {
                tempStr = "and "
            }
        }
        if len(v.SubWheres) == 0 {
            if v.Raw != "" {
                tempStr += v.Raw
                if len(v.RawBindings) > 0 {
                    *bindings = append(*bindings, v.RawBindings...)
                }
            } else {
                tempStr += v.Column + " " + v.Operator + " ?"
                *bindings = append(*bindings, v.Val)
            }
        } else {
            tempStr += "(" + m.generateWhereStr(v.SubWheres, bindings) + ")"
        }
        whereStr = append(whereStr, tempStr)
    }
    return strings.Join(whereStr, " ")
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) Where(column interface{}, vals ...interface{}) Query[T] {
    t := m.where(false, column, vals...)
    return t
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) OrWhere(column interface{}, vals ...interface{}) Query[T] {
    return m.where(true, column, vals...)
}

//where column(primary) = ? or column(primary) in (?)
func (m Query[T]) WherePrimary(val ...interface{}) Query[T] {
    if len(val) > 1 {
        return m.where(false, m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, val)
    } else if len(val) == 1 {
        if reflect.TypeOf(val[0]).Kind() == reflect.Slice {
            return m.where(false, m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, val[0])
        } else {
            return m.where(false, m.tables[0].tableStruct.Field(0).Addr().Interface(), val[0])
        }
    } else {
        return m
    }
}

func (m Query[T]) OrWherePrimary(val ...interface{}) Query[T] {
    if len(val) > 1 {
        return m.where(true, m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, val)
    } else if len(val) == 1 {
        if reflect.TypeOf(val[0]).Kind() == reflect.Slice {
            return m.where(true, m.tables[0].tableStruct.Field(0).Addr().Interface(), WhereIn, val[0])
        } else {
            return m.where(true, m.tables[0].tableStruct.Field(0).Addr().Interface(), val[0])
        }
    } else {
        return m
    }
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) WhereFunc(f func(Query[T]) Query[T]) Query[T] {
    return m.whereGroup(false, f)
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) OrWhereFunc(f func(Query[T]) Query[T]) Query[T] {
    return m.whereGroup(true, f)
}

func (m Query[T]) whereGroup(isOr bool, f func(Query[T]) Query[T]) Query[T] {
    temp, err := m.generateWhereGroup(f)
    m.setErr(err)
    if len(temp.SubWheres) > 0 {
        temp.IsOr = isOr
        m.wheres = append(m.wheres, temp)
    }
    return m
}

func (m Query[T]) generateWhereGroup(f func(Query[T]) Query[T]) (where, error) {
    start := len(m.wheres)
    nq := f(m)
    newWheres := nq.wheres[start:]

    if len(newWheres) > 0 {
        subwheres := make([]where, 0)
        m.wheres = m.wheres[:start]
        return where{SubWheres: append(subwheres, newWheres...)}, nq.result.Err
    }
    return where{}, nq.result.Err
}
