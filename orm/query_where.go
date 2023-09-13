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

func (q *Query[T]) where(isOr bool, column interface{}, vals ...interface{}) *Query[T] {
    if len(vals) > 2 {
        return q.setErr(errors.New("two many where-params"))
    }

    if len(vals) == 0 {
        c, ok := q.isStringOrRaw(column)
        if ok == false {
            return q.setErr(errors.New("where-param should be string while only 1 param exist"))
        }
        if c != "" {
            q.wheres = append(q.wheres, where{Raw: c, IsOr: isOr})
        } else {
            return q.setErr(errors.New("where-param should not be empty string"))
        }
    } else {
        c, err := q.parseColumn(column)
        if err != nil {
            return q.setErr(err)
        }
        operator := "="
        var val interface{}
        if len(vals) == 2 {
            operator2, ok := q.isStringOrRaw(vals[0])
            if ok == false {
                return q.setErr(errors.New("the second where-param should be operator as string"))
            }
            operator = operator2
            val = vals[1]
        } else {
            if vals[0] == nil {
                vals[0] = WhereIsNull
            }
            tempVal, ok := q.isOperator(vals[0])
            if ok {
                if tempVal != string(WhereIsNull) && tempVal != string(WhereIsNotNull) {
                    return q.setErr(errors.New("operator \"" + tempVal + "\" must have params"))
                }
                operator = ""
                val = Raw(tempVal)
            } else {
                val = vals[0]
            }
        }

        value, ok := q.isRaw(val)
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
                    rawColumn, err := q.parseColumn(val)
                    if err == nil {
                        raw = c + " " + operator + " " + rawColumn
                    } else {
                        return q.setErr(errors.New("Error where " + c + " " + operator + " ? val is invalid"))
                    }
                }
            }
        }
        q.wheres = append(q.wheres, where{Raw: raw, Column: c, Val: val, Operator: operator, IsOr: isOr, RawBindings: rawBindings})
    }
    return q
}

func (q *Query[T]) generateWhereStr(wheres []where, bindings *[]interface{}) string {
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
            tempStr += "(" + q.generateWhereStr(v.SubWheres, bindings) + ")"
        }
        whereStr = append(whereStr, tempStr)
    }
    return strings.Join(whereStr, " ")
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (q *Query[T]) Where(column interface{}, vals ...interface{}) *Query[T] {
    t := q.where(false, column, vals...)
    return t
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (q *Query[T]) OrWhere(column interface{}, vals ...interface{}) *Query[T] {
    return q.where(true, column, vals...)
}

//short for Where(primaryKey, vals...)
func (q *Query[T]) WherePrimary(operator interface{}, vals ...interface{}) *Query[T] {
    //operator as vals
    if len(vals) == 0 {
        vals = []interface{}{operator}
        reflectVar := reflect.ValueOf(operator)
        if reflectVar.Kind() == reflect.Slice {
            if reflectVar.Len() == 0 {
                return q
            }
            operator = WhereIn
        } else {
            operator = WhereEqual
        }
    }

    return q.where(false, q.tables[0].tableStruct.Field(0).Addr().Interface(), operator, vals[0])
}

//short for OrWhere(primaryKey, vals...)
func (q *Query[T]) OrWherePrimary(operator interface{}, vals ...interface{}) *Query[T] {
    //operator as vals
    if len(vals) == 0 {
        reflectVar := reflect.ValueOf(operator)
        if reflectVar.Kind() == reflect.Slice {
            if reflectVar.Len() == 0 {
                return q
            }
            operator = WhereIn
        } else {
            operator = WhereEqual
        }
    }

    return q.where(true, q.tables[0].tableStruct.Field(0).Addr().Interface(), operator, vals[0])
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (q *Query[T]) WhereFunc(f func(*Query[T]) *Query[T]) *Query[T] {
    return q.whereGroup(false, f)
}

//"id=1"
//&obj.id, 1
//&obj.id, "=", 1
func (q *Query[T]) OrWhereFunc(f func(*Query[T]) *Query[T]) *Query[T] {
    return q.whereGroup(true, f)
}

func (q *Query[T]) whereGroup(isOr bool, f func(*Query[T]) *Query[T]) *Query[T] {
    temp, err := q.generateWhereGroup(f)
    q.setErr(err)
    if len(temp.SubWheres) > 0 {
        temp.IsOr = isOr
        q.wheres = append(q.wheres, temp)
    }
    return q
}

func (q *Query[T]) generateWhereGroup(f func(*Query[T]) *Query[T]) (where, error) {
    start := len(q.wheres)
    nq := *q
    f(&nq)
    newWheres := nq.wheres[start:]

    if len(newWheres) > 0 {
        return where{SubWheres: append([]where{}, newWheres...)}, nq.result.Err
    }
    return where{}, nq.result.Err
}
