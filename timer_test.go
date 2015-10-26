package asdf

import (
	"testing"
	. "strconv"
	"fmt"
	"time"
)

type Entry struct {
	timer ITimer
	
	number int
}

func (me *Entry) Get(tidx uint) ITimer {
	return me.timer
}

func (me *Entry) Bind(tidx uint, timer ITimer) {
	me.timer = timer
}

func (me *Entry) Name(tidx uint) string {
	return Itoa(me.number)
}

func EntryCallback(proxy ITimerProxy) (bool, error) {
	e, ok := proxy.(*Entry)
	if !ok {
		return true, ErrBadType
	}
	
	fmt.Printf("Entry(%d) callback" + Crlf, e.number)
	
	return true, nil
}

const count = 100*1000
const ms = 1000
var clock = TmClock(ms)
var entry = [count]Entry{}

func testInit() {	
	for i:=0; i<count; i++ {
		entry[i].number = i
		
		entry[i].timer, _ =
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
