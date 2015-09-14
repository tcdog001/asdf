package asdf

import (
	"container/list"
	"fmt"
)

const (
	tm_BIT 		uint = 8			// 8
	tm_MASK 	uint = 1<<tm_BIT - 1// 255
	tm_SLOT 	uint = tm_MASK + 1	// 256
	tm_RING 	uint = 4*8/tm_BIT	// 4
	tm_RINGMAX 	uint = tm_RING - 1	// 3

	tm_PENDING 	uint = 1
	tm_CYCLE 	uint = 2
)

type TimerCallback func(entry interface{}) (bool/* safe */, error)

type ITimer interface {
	GetTimer(tidx uint) *Timer
	Name(tidx uint) string
}

type Timer struct {
	cb 		TimerCallback
	
	tidx 	uint
	flag	uint
	slot 	uint	// ring slot
	expires	uint	// ticks
	create	uint64	// ticks
	
	ring 	*tmRing
	clock 	*Clock
	node 	*list.Element
	father 	interface{}
}

func (me *Timer) IsDebug() bool {
	return me.clock.IsDebug()
}

func (me *Timer) dump(action string) {
	if me.IsDebug() {
		if iTimer, ok := me.father.(ITimer); ok {
			fmt.Printf("%s timer(%s) tidx(%d) ring(%d) slot(%d) expires(%d) create(%d)" + Crlf, 
				action,
				iTimer.Name(me.tidx),
				me.tidx,
				me.ring.idx,
				me.slot,
				me.expires,
				me.create)
		}
	}
}

func (me *Timer) IsPending() bool {
	return tm_PENDING==(tm_PENDING & me.flag)
}

func (me *Timer) IsCycle() bool {
	return tm_CYCLE==(tm_CYCLE & me.flag)
}

func (me *Timer) UnCycle() {
	me.flag &= ^tm_CYCLE
}

func (me *Timer) Cycle() {
	me.flag |= tm_CYCLE
}

func (me *Timer) left() uint /* ticks */ {
	timeout := me.create + uint64(me.expires)
	
	ticks := me.clock.ticks
	if timeout > ticks {
		return uint(timeout - ticks)
	}
	
	return 0
}

func (me *Timer) Left() uint /* ms */ {
	return me.left() * me.clock.unit // ticks==>ms
}

func (me *Timer) Change(after uint/* ms */) error {
	if nil==me {
		return ErrNilObj
	}
	
	if me.IsPending() {
		left := me.left()
		after /= me.clock.unit // ms==>ticks
		
		me.remove()
		
		if after > left {
			me.expires += after - left
		} else {
			me.expires -= left - after
		}
		
		me.insert()
		
		return nil
	}
	
	return ErrNoExist
}

func (me *Timer) findRing() (uint, *tmRing) {
	left := me.left()
	
	offset := uint(0)
	idx := tm_RINGMAX
	
	for ;idx>0; idx-- {
		offset = left
		offset <<= tm_BIT*(tm_RINGMAX-idx)
		offset >>= tm_BIT*tm_RINGMAX
		
		if offset > 0 {
			break
		}
	}
	
	if 0==idx {
		offset = left
	}
	
	r := me.clock.ringX(idx)
	
	return (r.current + offset) & tm_MASK, r
}

func (me *Timer) insert() {
	if nil==me.node {
		slot, r := me.findRing()
		
		me.node = r.List(slot).PushBack(me)
		me.slot = slot
		me.ring = r
		me.flag |= tm_PENDING
		
		me.dump("insert")
	}
}

func (me *Timer) remove() {
	if nil!=me.node {
		me.ring.List(me.slot).Remove(me.node)
		me.node 	= nil
		me.ring 	= nil
		me.flag 	&= ^tm_PENDING
	}
}

func (me *Timer) Remove() error {
	if nil==me {
		return ErrNilObj
	}
	
	if !me.IsPending() {
		return ErrNoPending
	}
	
	me.remove()
	me.father 	= nil
	me.clock	= nil
	
	return nil
}

type tmRing struct {
	list [tm_SLOT]list.List
	
	current uint
	idx 	uint // ring index
	
	clock 	*Clock
}

func (me *tmRing) Init(idx uint, clock *Clock) {
	me.idx 		= idx
	me.clock 	= clock
	me.current 	= 0
	
	for i:=uint(0); i<tm_SLOT; i++ {
		me.List(i).Init()
	}
}

func (me *tmRing) List(slot uint) *list.List {
	return &me.list[slot]
}

func (me *tmRing) IsDebug() bool {
	return me.clock.IsDebug()
}

func (me *tmRing) dumpList(slot uint, action string) {
	if Len := me.List(slot).Len(); Len > 0 && me.IsDebug() {
		for e := me.List(slot).Front(); e != nil; e = e.Next() {
			if t, ok := e.Value.(*Timer); ok {
				t.dump(action)
			}
		}
	}
}

func (me *tmRing) dump(action string) {
	fmt.Printf("%s ring(%d) current(%d)" + Crlf,
		action,
		me.idx,
		me.current)
	
	for i:=uint(0); i<tm_SLOT; i++ {
		me.dumpList(i, action)
	}
}

func (me *tmRing) trigger() uint {
	count := uint(0)
	
	var next *list.Element
	for e := me.List(me.current).Front(); e != nil; e = next {
		next = e.Next()
		
		t, ok := e.Value.(*Timer)
		if !ok {
			continue
		}
		
		t.remove()
		
		if t.left() > 0 {
			t.insert()
			
			continue
		}
		
		if safe, err := t.cb(t.father); nil==err {
			count++
		} else if !safe {
			continue
		}
		
		if t.IsCycle() {
			t.create = t.clock.ticks
			t.insert()
		}
	}
	
	return count
}

func (me *tmRing) Trigger() uint {
	count := uint(0)	
	
	me.current++
	me.current &= tm_MASK
	
	count += me.trigger()
	
	if idx := me.idx; 0==me.current && idx < tm_RINGMAX {
		r := me.clock.ringX(idx)
		
		count += r.trigger()
	}
	
	return count
}

type Clock struct {
	init 	bool
	ticks 	uint64
	
	ring 	[tm_RING]tmRing
	unit 	uint
	Type 	uint
	debug 	*bool
}

func (me *Clock) ringX(idx uint) *tmRing {
	return &me.ring[idx]
}

func (me *Clock) ring0() *tmRing {
	return me.ringX(0)
}

func (me *Clock) ringMax() *tmRing {
	return me.ringX(tm_RINGMAX)
}

func (me *Clock) Ticks() uint64 {
	return me.ticks
}

func (me *Clock) GetUnit() uint {
	return me.unit
}

func (me *Clock) Trigger(times uint) uint {
	count := uint(0)
	
	me.dump("dump")
	
	for i:=uint(0); i<times; i++ {
		me.ticks++
		
		count += me.ring0().Trigger()
	}
	
	return count
}

func getTimer(entry interface{}, tidx uint) *Timer {
	var t *Timer
	
	if iTimer, ok := entry.(ITimer); !ok {
		return nil
	} else if t = iTimer.GetTimer(tidx); nil==t {
		return nil
	}
	
	return t
}

func (me *Clock) Insert(
		entry interface{},
		tidx uint, // timer index
		after uint, // ms
		cb TimerCallback, 
		Cycle bool) (*Timer, error) {
	if nil==me {
		return nil, ErrNilObj
	}
	
	if nil==entry {
		return nil, ErrNilObj
	}
	
	if nil==cb {
		return nil, ErrNilObj
	}
	
	t := getTimer(entry, tidx)
	if nil==t {
		return nil, ErrBadIntf
	}
	
	if t.IsPending() {
		return nil, ErrPending
	}
	
	flag := uint(0)
	if Cycle {
		flag = tm_CYCLE
	}
	
	t.cb 		= cb
	t.tidx 		= tidx
	t.create	= me.ticks
	t.expires	= after/me.unit // ms==>ticks
	t.flag 		= flag
	t.clock 	= me
	t.father	= entry
	
	t.insert()
	
	return t, nil
}

func (me *Clock) IsDebug() bool {
	return *me.debug
}

func (me *Clock) dump(action string) {
	if me.IsDebug() {
		fmt.Println("======================")
		for i:=uint(0); i<tm_RING; i++ {
			me.ringX(i).dump(action)
		}
		
		fmt.Println("======================")
	}
}

func TmClock(unit uint/* ms */, debug *bool) *Clock {
	c := &Clock{
		debug:debug,
		unit:unit,
	}
	
	for i:=uint(0); i<tm_RING; i++ {
		c.ringX(i).Init(i, c)
	}
	
	fmt.Println("clock create")
	return c
}
