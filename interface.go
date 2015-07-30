package asdf

////////////////////////////////////////////////////////////////////////////////
// single interface
////////////////////////////////////////////////////////////////////////////////
type IBegin	interface {
	Begin() int
}

type IEnd interface {
	End() int
}

type IInt interface {
	Int() int
}

type IFloat interface {
	Float() float64
}

type ISlice interface {
	Slice() []byte
}

type IEq interface {
	Eq(interface{}) bool
}

type ILt interface {
	Lt(interface{}) bool
}

type IGt interface {
	Gt(interface{}) bool
}

type IFirst interface {
	First() interface{}
}

type ILast interface {
	Last() interface{}
}

type ITails interface {
	Tail() []interface{}
}

type IHead interface {
	Head() []interface{}
}

type IPrev interface {
	Prev() interface{}
}

type INext interface {
	Next() interface{}
}

type IToString interface {
	ToString() string
}

type IFromString interface {
	FromString(string) error
}

type IToBinary interface {
	ToBinary([]byte) error
}

type IFromBinary interface {
	FromBinary([]byte) error
}

type IGood interface {
	IsGood() bool
}

type IReverse interface {
	Reverse() []interface{}
}

type IRepeat interface {
	Repeat(int) []interface{}
}

////////////////////////////////////////////////////////////////////////////////
// combination interface
////////////////////////////////////////////////////////////////////////////////
type IBound	interface {
	IBegin
	IEnd
}

type INumber interface {
	IBound
	IInt
}

type IObjBinary interface {
	IToBinary
	IFromBinary
}

type ILogger interface {
	Emerg(format string, v ...interface{})
	Alert(format string, v ...interface{})
	Crit(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Notice(format string, v ...interface{})
	Info(format string, v ...interface{})
	Debug(format string, v ...interface{})
}