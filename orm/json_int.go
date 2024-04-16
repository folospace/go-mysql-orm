package orm

import (
    "database/sql/driver"
    "fmt"
    "strconv"
    "strings"
)

//json int or long string int <=> go int64
type JsonInt int64

func (t JsonInt) MarshalJSON() ([]byte, error) {
    return []byte("\"" + strconv.FormatInt(int64(t), 10) + "\""), nil
}

func (t *JsonInt) UnmarshalJSON(data []byte) error {
    val, err := strconv.ParseInt(strings.Trim(string(data), "\""), 10, 64)
    if err != nil {
        return err
    }
    *t = JsonInt(val)
    return nil
}

func (t JsonInt) Value() (driver.Value, error) {
    return int64(t), nil
}

func (t JsonInt) ToString() string {
    return strconv.FormatInt(int64(t), 10)
}

func (t *JsonInt) Scan(v any) error {
    val, ok := v.(int64)
    if ok {
        *t = JsonInt(val)
        return nil
    } else {
        val, ok := v.([]uint8)
        if ok {
            v, _ := strconv.ParseInt(string(val), 10, 64)
            *t = JsonInt(v)
            return nil
        }
    }
    return fmt.Errorf("can not convert %v to json int", v)
}
