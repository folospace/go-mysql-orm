package orm

import (
    "database/sql/driver"
    "encoding/json"
    "unsafe"
)

//json ojbect/slice <=> go struct/slice <=> db json string
type JsonField[T any] struct {
    Data T
}

func NewJsonField[T any](data T) JsonField[T] {
    return JsonField[T]{Data: data}
}

func (t JsonField[T]) MarshalJSON() ([]byte, error) {
    return json.Marshal(t.Data)
}

func (t *JsonField[T]) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &t.Data)
}

func (t JsonField[T]) Value() (driver.Value, error) {
    data, err := json.Marshal(t.Data)
    return *(*string)(unsafe.Pointer(&data)), err
}

func (t *JsonField[T]) Scan(raw any) error {
    rawData, ok := raw.([]byte)
    if !ok {
        return nil
    }

    if len(rawData) == 0 {
        return nil
    }

    var i T
    err := json.Unmarshal(rawData, &i)
    if err != nil {
        return err
    }
    t.Data = i
    return nil
}
