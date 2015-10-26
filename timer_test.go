package asdf

import (
	. "strconv"
	"fmt"
	"time"
	"testing"
)

type Entry struct {
	timer ITimer
	
	number int
}

func (me *Entry) TimerName(tidx uint) string {
	return Itoa(me.number)
}

func (me *Entry) GetTimer(tidx uint) ITimer {
	return me.timer
}

func (me *Entry) SetTimer(tidx uint, timer ITimer) {
	me.timer = timer
}

func EntryCallback(proxy ITimerProxy) (bool, error) {
	e, ok := proxy.(*Entry)
	if !ok {
		return true, ErrBadType
	}
	
	fmt.Printf(" %d", e.number)
	
	return true, nil
}

const count = 1000
const ms = 1
const timeout = 300
const cpu = 1
const times = 100

var clock = [cpu]IClock{}
var entry = [cpu][count]Entry{}

func Init(idx int) {
	clock[idx] = TmClock(ms)
	
	for i:=0; i<count; i++ {
		entry[idx][i].number = i
		
		entry[idx][i].timer, _ =
			clock[idx].Insert(&entry[idx][i], 0, ms*uint(i)/timeout, EntryCallback, true)
	}
}

func run(idx int) {
	Init(idx)
	
	for i:=0;i<times;i++ {
		time.Sleep(ms*1000*1000)
		
		fmt.Printf("%.10d:", i+1)
		clock[idx].Trigger(1)
		fmt.Printf("\n")
	}
}

func TestTimer(t *testing.T) {
	for i:=cpu-1; i>0; i-- {
		go run(i)
	}
	
	run(0)
}
