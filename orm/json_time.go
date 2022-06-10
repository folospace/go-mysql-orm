package orm

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type JsonTime struct {
	time.Time
}

// MarshalJSON on JsonTime format Time field with %Y-%m-%d %H:%M:%S
func (t JsonTime) MarshalJSON() ([]byte, error) {
	//formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	//return []byte(formatted), nil
	ts := t.UnixNano() / 1e6
	if ts < 0 {
		ts = 0
	}
	return []byte(strconv.FormatInt(ts, 10)), nil
}

// MarshalJSON on JsonTime format Time field with %Y-%m-%d %H:%M:%S
func (t *JsonTime) UnmarshalJSON(data []byte) error {
	//formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	//return []byte(formatted), nil
	val, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	if val == 0 {
		t.Time = time.Time{}
		return nil
	}
	t.Time = time.Unix(0, val*1e6)
	return nil
}

// Value insert timestamp into mysql need this function.
func (t JsonTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JsonTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok && t != nil {
		*t = JsonTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
