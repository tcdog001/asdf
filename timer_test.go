package asdf

import (
	"testing"
	. "strconv"
	"fmt"
	"time"
)

type Entry struct {
	Timer
	
	number int
}

func (me *Entry) GetTimer(tidx uint) *Timer {
	return &me.Timer
}

func (me *Entry) Name(tidx uint) string {
	return Itoa(me.number)
}

func EntryCallback(entry interface{}) (bool, error) {
	e, ok := entry.(*Entry)
	if !ok {
		return true, ErrBadType
	}
	
	fmt.Printf("Entry(%d) callback" + Crlf, e.number)
	
	return true, nil
}

const count = 100*1000
const ms = 1000
var debug = false
var clock = TmClock(ms, &debug)
var entry = [count]Entry{}

func testInit() {	
	for i:=0; i<count; i++ {
		entry[i].number = i
		
		clock.Insert(&entry[i], 0, ms*uint(i), EntryCallback, true)
	}
}

func TestTimer(t *testing.T) {
	testInit()
	
	for i:=0;i<count;i++ {
		time.Sleep(ms*1000*1000)
		
		clock.Trigger(1)
		
		fmt.Println("Trigger", i+1)
	}
}
