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

type TimerCallback func(proxy ITimerProxy) (bool/* safe */, error)

type ITimerProxy interface {
	Get(tidx uint) ITimer
	Bind(tidx uint, timer ITimer)
	Name(tidx uint) string
}

type ITimer interface {
	IsPending() bool
	IsCycle() bool
	
	Left() uint
	Change(after uint/* ms */) error
	Remove() error
}

type IClock interface {
	Insert(
		proxy ITimerProxy,
		tidx uint, // timer index
		after uint, // ms
		cb TimerCallback, 
		cycle bool) (ITimer, error)
	Trigger(times uint) uint
	Ticks() uint64
	Unit() uint
	Debug(debug bool)
}

type tmTimer struct {
	cb 		TimerCallback
	
	tidx 	uint
	flag	uint
	slot 	uint	// ring slot
	expires	uint	// ticks
	create	uint64	// ticks
	
	ring 	*tmRing
	clock 	*tmClock
	node 	*list.Element
	proxy	ITimerProxy
}

func (me *tmTimer) isDebug() bool {
	return me.clock.debug
}

func (me *tmTimer) dump(action string) {
	if me.isDebug() {
		fmt.Printf("%s timer(%s) tidx(%d) ring(%d) slot(%d) expires(%d) create(%d)" + Crlf, 
			action,
			me.proxy.Name(me.tidx),
			me.tidx,
			me.ring.idx,
			me.slot,
			me.expires,
			me.create)
	}
}

func (me *tmTimer) IsPending() bool {
	return tm_PENDING==(tm_PENDING & me.flag)
}

func (me *tmTimer) IsCycle() bool {
	return tm_CYCLE==(tm_CYCLE & me.flag)
}

/*
func (me *tmTimer) undoCycle() {
	me.flag &= ^tm_CYCLE
}

func (me *tmTimer) cycle() {
	me.flag |= tm_CYCLE
}
*/

func (me *tmTimer) left() uint /* ticks */ {
	timeout := me.create + uint64(me.expires)
	
	ticks := me.clock.ticks
	if timeout > ticks {
		return uint(timeout - ticks)
	}
	
	return 0
}

func (me *tmTimer) Left() uint /* ms */ {
	return me.left() * me.clock.unit // ticks==>ms
}

func (me *tmTimer) Change(after uint/* ms */) error {
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

func (me *tmTimer) findRing() (uint, *tmRing) {
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

func (me *tmTimer) bind(tidx uint, proxy ITimerProxy) {
	me.proxy 	= proxy
	me.tidx		= tidx

	proxy.Bind(tidx, me)
}

func (me *tmTimer) unbind() {
	me.proxy.Bind(me.tidx, nil)
	
	me.proxy 	= nil
	me.tidx  	= 0
}

func (me *tmTimer) insert() {
	if nil==me.node {
		slot, r := me.findRing()
		
		me.node = r.list(slot).PushBack(me)
		me.slot = slot
		me.ring = r
		me.flag |= tm_PENDING
				
		me.dump("insert")
	}
}

func (me *tmTimer) remove() {
	if nil!=me.node {
		me.ring.list(me.slot).Remove(me.node)
		me.node	= nil
		me.ring	= nil
		me.flag &= ^tm_PENDING
	}
}

func (me *tmTimer) Remove() error {
	if nil==me {
		return ErrNilObj
	}
	
	if !me.IsPending() {
		return ErrNoPending
	}
	
	me.remove()
	me.unbind()
	
	return nil
}

type tmRing struct {
	hash [tm_SLOT]list.List
	
	current uint
	idx 	uint // ring index
	
	clock 	*tmClock
}

func (me *tmRing) init(idx uint, clock *tmClock) {
	me.idx 		= idx
	me.clock 	= clock
	me.current 	= 0
	
	for i:=uint(0); i<tm_SLOT; i++ {
		me.list(i).Init()
	}
}

func (me *tmRing) list(slot uint) *list.List {
	return &me.hash[slot]
}

func (me *tmRing) isDebug() bool {
	return me.clock.debug
}

func (me *tmRing) dumpList(slot uint, action string) {
	if Len := me.list(slot).Len(); Len > 0 && me.isDebug() {
		for e := me.list(slot).Front(); e != nil; e = e.Next() {
			if t, ok := e.Value.(*tmTimer); ok {
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

func (me *tmRing) __trigger() uint {
	count := uint(0)
	
	var next *list.Element
	for e := me.list(me.current).Front(); e != nil; e = next {
		next = e.Next()
		
		t, ok := e.Value.(*tmTimer)
		if !ok {
			continue
		}
		
		t.remove()
		
		if t.left() > 0 {
			t.insert()
			
			continue
		}
		
		if safe, err := t.cb(t.proxy); nil==err {
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

func (me *tmRing) trigger() uint {
	count := uint(0)	
	
	me.current++
	me.current &= tm_MASK
	
	count += me.__trigger()
	
	if idx := me.idx; 0==me.current && idx < tm_RINGMAX {
		count += me.clock.ringX(idx).__trigger()
	}
	
	return count
}

type tmClock struct {
	init 	bool
	ticks 	uint64
	
	ring 	[tm_RING]tmRing
	unit 	uint
	debug 	bool
}

func (me *tmClock) ringX(idx uint) *tmRing {
	return &me.ring[idx]
}

func (me *tmClock) Ticks() uint64 {
	return me.ticks
}

func (me *tmClock) Unit() uint {
	return me.unit
}

func (me *tmClock) Trigger(times uint) uint {
	count := uint(0)
	
	me.dump("dump")
	
	for i:=uint(0); i<times; i++ {
		me.ticks++
		
		count += me.ringX(0).trigger()
	}
	
	return count
}

func (me *tmClock) Insert(
		proxy ITimerProxy,
		tidx uint, // timer index
		after uint, // ms
		cb TimerCallback, 
		cycle bool) (ITimer, error) {
	if nil==me {
		return nil, ErrNilObj
	} else if nil==cb {
		return nil, ErrNilObj
	} else if nil==proxy {
		return nil, ErrNilObj
	} else if nil!=proxy.Get(tidx) {
		return nil, ErrExist
	}
	
	flag := uint(0)
	if cycle {
		flag = tm_CYCLE
	}
	
	timer := &tmTimer{
		cb:		cb,
		create:	me.ticks,
		expires:after/me.unit, // ms==>ticks
		flag:	flag,
		clock:	me,
	}
	
	timer.bind(tidx, proxy)
	timer.insert()
	
	return timer, nil
}

func (me *tmClock) Debug(debug bool) {
	me.debug = debug
}

func (me *tmClock) dump(action string) {
	if me.debug {
		fmt.Println("======================")
		for i:=uint(0); i<tm_RING; i++ {
			me.ringX(i).dump(action)
		}
		fmt.Println("======================")
	}
}

func TmClock(unit uint/* ms */) IClock {
	c := &tmClock{
		unit:unit,
	}
	
	for i:=uint(0); i<tm_RING; i++ {
		c.ringX(i).init(i, c)
	}
	
	fmt.Println("clock create")
	
	return c
}
