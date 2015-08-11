package asdf

import (
	"container/list"
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
	GetTimer(Type uint) interface{}
}

type tmNode struct {
	cb 		TimerCallback
	
	flag	uint
	slot 	uint	// ring slot
	expires	uint
	create	uint64
	
	ring 	*tmRing
	node *list.Element
}

func (me *tmNode) IsCycle() bool {
	return tm_CYCLE==(tm_CYCLE & me.flag)
}

func (me *tmNode) IsPending() bool {
	return tm_PENDING==(tm_PENDING & me.flag)
}

func (me *tmNode) Clock() *Clock {
	return me.ring.clock
}

func (me *tmNode) Left() uint {
	timeout := me.create + uint64(me.expires)
	
	if timeout > me.Clock().ticks {
		return uint(timeout - me.Clock().ticks)
	}
	
	return 0
}

func (me *tmNode) FindRing() (uint, *tmRing) {
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

func (me *tmNode) Insert() error {
	slot, r := me.FindRing()
	
	me.node = r.List(slot).PushFront(me)
	me.slot = slot
	me.flag |= tm_PENDING
	
	if me.Clock().ticks > 0 {
		// test ???
	}
	
	return nil
}

func (me *tmNode) Remove() error {
	r := me.ring
	
	if me.Clock().ticks > 0 {
		// test ???
	}
	
	r.List(me.slot).Remove(me.node)
	me.node = nil
	me.flag &= ^tm_PENDING
	
	return nil
}

type tmRing struct {
	list [tm_SLOT]list.List
	
	current uint
	idx 	uint // ring index
	
	clock 	*Clock
}

func (me *tmRing) List(idx uint) *list.List {
	return &me.list[idx]
}

type Clock struct {
	init 	bool
	ticks 	uint64
	
	ring 	[tm_RING]tmRing
	unit 	uint
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

func (me *Clock) Trigger(times uint) error {
	return nil
}

func (me *Clock) Insert(
		entry interface{}, 
		after uint, 
		cb TimerCallback, 
		Cycle bool) (interface{}, error) {
	if _, ok := entry.(ITimer); !ok {
		return nil, ErrBadIntf
	}

	flag := uint(0)
	if Cycle {
		flag = tm_CYCLE
	}
	
	t := &tmNode{
		cb:cb,
		create:me.ticks,
		expires:after,
		flag:flag,
	}
	
	if err := t.Insert(); nil!=err {
		return nil, err
	}
	
	return t, nil
}

func (me *Clock) Remove(entry interface{}, Type uint) error {
	t, err := tmNodeGet(entry, Type)
	if nil!=err {
		return err
	}
	
	return t.Remove()
}

func tmNodeGet(entry interface{}, Type uint) (*tmNode, error) {
	var iEntry ITimer
	var ok bool
	var t *tmNode
	
	iEntry, ok = entry.(ITimer)
	if !ok {
		return nil, ErrBadIntf
	}
	
	t, ok = iEntry.GetTimer(Type).(*tmNode)
	if !ok {
		return nil, ErrBadType
	}
	
	return t, nil
}

func TmClock() *Clock {
	c := &Clock{}
	
	for i:=uint(0); i<tm_RING; i++ {
		r := c.ringX(i)
		
		r.idx = i
		r.clock = c
		r.current = 0
		
		for j:=uint(0); j<tm_SLOT; j++ {
			r.List(j).Init()
		}
	}
	
	return c
}
