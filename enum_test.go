package asdf

import (
	"errors"
	"fmt"
	"testing"
)

type aEnum int

func (me aEnum) Begin() int {
	return int(enumBegin)
}

func (me aEnum) End() int {
	return int(enumEnd)
}

func (me aEnum) Int() int {
	return int(me)
}

func (me aEnum) IsGood() bool {
	return IsGoodEnum(me) &&
		len(enumTestBind) == me.End() &&
		len(enumTestBind[me]) > 0
}

func (me aEnum) ToString() string {
	var b EnumBinding = enumTestBind[:]

	return b.EntryShow(me)
}

func (me *aEnum) Read(s string) error {
	v, ok := enumTestMap[s]
	if !ok {
		return errors.New("invalid aEnum string")
	}

	*me = v
	return nil
}

const (
	enumBegin aEnum = 0

	enum0 aEnum = 0
	enum1 aEnum = 1

	enumEnd aEnum = 2
)

var enumTestBind = [enumEnd]string{
	enum0: "Enum-0",
}

var enumTestMap = map[string]aEnum{}

func initEnumTest() {
	for i := enumBegin; i < enumEnd; i++ {
		if len(enumTestBind[i]) > 0 {
			enumTestMap[enumTestBind[i]] = i
		}
	}
}

func TestEnum(t *testing.T) {
	initEnumTest()

	fmt.Println("enum0 - 1 =", (enum0 - 1).ToString()) // empty
	fmt.Println("enum0 =", enum0.ToString())           // Good
	fmt.Println("enum1 =", enum1.ToString())           // empty
	fmt.Println("enum1 + 1 =", (enum1 + 1).ToString()) // empty

	var e aEnum

	e = -1
	(&e).Read("Enum-0")
	fmt.Println("e =", e.Int(), e.ToString())

	e = -1
	(&e).Read("Enum-1")
	fmt.Println("e =", e.Int(), e.ToString())
}
