package orm

type where struct {
	IsOr        bool
	Column      string
	Operator    string
	Val         interface{}
	Raw         string
	RawBindings []interface{}
	SubWheres   []where
}
