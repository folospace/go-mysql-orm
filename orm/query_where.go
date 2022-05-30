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
        return m.setErr(errors.New("where vals should be less than 2"))
    }

    if len(vals) == 0 {
        c, ok := m.isStringOrRaw(column)
        if ok == false {
            return m.setErr(errors.New("where only 1 param should be string as PrepareSql raw"))
        }
        m.wheres = append(m.wheres, where{Raw: c, IsOr: isOr})
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
                return m.setErr(errors.New("where middle param should be operator string"))
            }
            operator = operator2
            val = vals[1]
        } else {
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
            tempTable, ok := val.(*tempTable)
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
    return m.where(false, column, vals...)
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) OrWhere(column interface{}, vals ...interface{}) Query[T] {
    return m.where(true, column, vals...)
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) WhereGroup(f func(*Query[T])) Query[T] {
    return m.whereGroup(false, f)
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (m Query[T]) OrWhereGroup(f func(*Query[T])) Query[T] {
    return m.whereGroup(true, f)
}

func (m Query[T]) whereGroup(isOr bool, f func(*Query[T])) Query[T] {
    temp := m.generateWhereGroup(f)

    if len(temp.SubWheres) > 0 {
        temp.IsOr = isOr
        m.wheres = append(m.wheres, temp)
    }
    return m
}

func (m Query[T]) generateWhereGroup(f func(*Query[T])) where {
    start := len(m.wheres)
    f(&m)
    newWheres := m.wheres[start:]
    if len(newWheres) > 0 {
        subwheres := make([]where, 0)
        m.wheres = m.wheres[:start]
        return where{SubWheres: append(subwheres, newWheres...)}
    }
    return where{}
}
