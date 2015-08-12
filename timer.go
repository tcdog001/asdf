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

type TimerCallback func(entry interface{}) (bool, error)

type ITimer interface {
	GetTimer(entry interface{}, tag uint) *Timer
	Name() string
}

type Timer struct {
	cb 		TimerCallback
	
	tag 	uint
	flag	uint
	slot 	uint	// ring slot
	expires	uint
	create	uint64
	
	ring 	*tmRing
	node 	*list.Element
	father 	interface{}
}

func (me *Timer) dump() {
	if *me.ring.clock.debug {
		iTimer, ok := me.father.(ITimer)
		if !ok {
			return
		}
		
		fmt.Printf("timer(%s) tag(%d) ring(%d) slot(%d) expires(%d) create(%ld)" + Crlf, 
			iTimer.Name(),
			me.tag,
			me.ring.idx,
			me.slot,
			me.expires,
			me.create)
	}
}

func (me *Timer) IsCycle() bool {
	return tm_CYCLE==(tm_CYCLE & me.flag)
}

func (me *Timer) IsPending() bool {
	return tm_PENDING==(tm_PENDING & me.flag)
}

func (me *Timer) Clock() *Clock {
	return me.ring.clock
}

func (me *Timer) Left() uint {
	timeout := me.create + uint64(me.expires)
	
	if timeout > me.Clock().ticks {
		return uint(timeout - me.Clock().ticks)
	}
	
	return 0
}

func (me *Timer) UnCycle() {
	me.flag &= ^tm_CYCLE
}

func (me *Timer) Cycle() {
	me.flag |= tm_CYCLE
}

func (me *Timer) ChangeExpires(expires uint/* ticks */) error {
	if nil==me {
		return ErrNilObj
	}
	
	if me.IsCycle() {
		me.expires = expires
		
		return nil
	}
	
	return ErrNoSupport
}

func (me *Timer) Change(after uint/* ticks */) error {
	if nil==me {
		return ErrNilObj
	}
	
	if me.IsPending() {
		left := me.Left()
		
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
	left := me.Left()
	
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
	
	r := me.Clock().ringX(idx)
	
	return (r.current + offset) & tm_MASK, r
}

func (me *Timer) insert() {
	slot, r := me.findRing()
	
	me.node = r.List(slot).PushFront(me)
	me.slot = slot
	me.ring = r
	me.flag |= tm_PENDING
}

func (me *Timer) remove() {
	r := me.ring
	
	r.List(me.slot).Remove(me.node)
	me.node = nil
	me.ring = nil
	me.flag &= ^tm_PENDING
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

func (me *tmRing) dumpList(slot uint) {
	if Len := me.List(slot).Len(); Len > 0 && *me.clock.debug {
		for e := me.List(me.current).Front(); e != nil; e = e.Next() {
			if t, ok := e.Value.(*Timer); ok {
				t.dump()
			}
		}
	}
}

func (me *tmRing) dump() {
	fmt.Printf("ring(%d) current(%d)" + Crlf, 
		me.idx,
		me.current)
	
	for i:=uint(0); i<tm_SLOT; i++ {
		me.dumpList(i)
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
		
		if t.Left() > 0 {
			t.insert()
			
			continue
		}
		
		if safe, err := t.cb(t.father); nil==err {
			count++
		} else if !safe {
			continue
		}
		
		if t.IsCycle() {
			t.create = t.Clock().ticks
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

func (me *Clock) SetUnit(unit uint/* ms */) {
	me.unit = unit
}

func (me *Clock) GetUnit() uint {
	return me.unit
}

func (me *Clock) Trigger(times uint) uint {
	count := uint(0)
	
	for i:=uint(0); i<times; i++ {
		me.ticks++
		
		count += me.ring0().Trigger()
	}
	
	return count
}

func (me *Clock) Insert(
		entry interface{},
		tag uint,
		after uint, 
		cb TimerCallback, 
		Cycle bool) (*Timer, error) {
	if nil==entry {
		return nil, ErrNilObj
	}
	
	if nil==cb {
		return nil, ErrNilObj
	}
	
	var t *Timer
	if iTimer, ok := entry.(ITimer); !ok {
		return nil, ErrBadIntf
	} else if t = iTimer.GetTimer(entry, tag); nil==t {
		return nil, ErrBadIntf
	}
	
	flag := uint(0)
	if Cycle {
		flag = tm_CYCLE
	}
	
	t.cb 		= cb
	t.tag 		= tag
	t.create	= me.ticks
	t.expires	= after
	t.flag 		= flag
	t.father	= entry
	
	t.insert()
	
	return t, nil
}

func (me *Clock) Remove(entry interface{}) error {
	t, ok := entry.(*Timer)
	if !ok {
		return ErrBadType
	}
	
	t.remove()
	t.father = nil
	
	return nil
}

func (me *Clock) dump() {
	if *me.debug {
		fmt.Println("======================")
		for i:=uint(0); i<tm_RING; i++ {
			me.ringX(i).dump()
		}
		
		fmt.Println("======================")
	}
}

func TmClock(debug *bool) *Clock {
	c := &Clock{
		debug:debug,
	}
	
	for i:=uint(0); i<tm_RING; i++ {
		c.ringX(i).Init(i, c)
	}
	
	return c
}
