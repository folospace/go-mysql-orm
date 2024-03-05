package orm

type where struct {
    IsOr        bool
    Column      string
    Operator    string
    Val         any
    Raw         string
    RawBindings []any
    SubWheres   []where
}
