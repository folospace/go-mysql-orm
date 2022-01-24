package orm

import (
    "database/sql/driver"
    "fmt"
    "strconv"
    "strings"
)

type JsonInt int64

func (i JsonInt) MarshalJSON() ([]byte, error) {
    return []byte("\"" + strconv.FormatInt(int64(i), 10) + "\""), nil
}

func (i *JsonInt) UnmarshalJSON(data []byte) error {
    val, err := strconv.ParseInt(strings.Trim(string(data), "\""), 10, 64)
    if err != nil {
        return err
    }
    *i = JsonInt(val)
    return nil
}

func (i JsonInt) Value() (driver.Value, error) {
    return int64(i), nil
}

func (i JsonInt) ToString() string {
    return strconv.FormatInt(int64(i), 10)
}

func (i *JsonInt) Scan(v interface{}) error {
    val, ok := v.(int64)
    if ok {
        *i = JsonInt(val)
        return nil
    } else {
        val, ok := v.([]uint8)
        if ok {
            v, _ := strconv.ParseInt(string(val), 10, 64)
            *i = JsonInt(v)
            return nil
        }
    }
    return fmt.Errorf("can not convert %v to json int", v)
}
