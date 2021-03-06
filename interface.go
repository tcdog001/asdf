package asdf

////////////////////////////////////////////////////////////////////////////////
// single interface
////////////////////////////////////////////////////////////////////////////////
type IBegin interface {
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

type IReverse interface {
	Reverse() []interface{}
}

type IRepeat interface {
	Repeat(int) []interface{}
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

////////////////////////////////////////////////////////////////////////////////
// combination interface
////////////////////////////////////////////////////////////////////////////////
type IBound interface {
	// [begin, end)
	IBegin
	IEnd
}

type INumber interface {
	IBound
	IInt
}

type IString interface {
	IToString
	IFromString
}

type IBinary interface {
	IToBinary
	IFromBinary
}

type ICompare interface {
	IEq
	IGt
}

type IList interface {
	IFirst
	ILast

	ITails
	IHead

	IReverse
}

type IListNode interface {
	IPrev
	INext
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
