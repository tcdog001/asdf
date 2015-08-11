package asdf

import (
	"container/list"
)

const (
	TM_BIT 		= 8				// 8
	TM_MASK 	= 1<<TM_BIT - 1	// 255
	TM_SLOT 	= TM_MASK + 1	// 256
	TM_RING 	= 4*8/TM_BIT	// 4
	TM_RINGMAX 	= TM_RING - 1	// 3
	
	TM_PENDING 	uint8 = 1
	TM_CYCLE 	uint8 = 2
)

type TimerCallback func() (bool, error)

type Timer struct {
	cb 		TimerCallback
	flag	uint8
	idx 	uint8
	slot 	uint16
	expires	uint32
	create	uint64
	
	node *list.Element
}

func (me *Timer) Idx() uint32 {
	return uint32(me.idx)
}

func (me *Timer) Slot() uint32 {
	return uint32(me.slot)
}

func (me *Timer) Flag() uint32 {
	return uint32(me.flag)
}

func (me *Timer) IsCycle() bool {
	return TM_CYCLE==(TM_CYCLE & me.flag)
}

func (me *Timer) IsPending() bool {
	return TM_PENDING==(TM_PENDING & me.flag)
}

func (me *Timer) left() uint32 {
	timeout := me.create + uint64(me.expires)
	
	if timeout > clock.ticks {
		return uint32(timeout - clock.ticks)
	}
	
	return 0
}

func (me *Timer) findRing() (uint32 /* slot */, *tmRing) {
	left := me.left()
	
	offset := uint32(0)
	idx := uint32(TM_RINGMAX)
	
	for ;idx>0; idx-- {
		offset = left
		offset <<= TM_BIT*(TM_RINGMAX-idx)
		offset >>= TM_BIT*TM_RINGMAX
		
		if offset > 0 {
			break
		}
	}
	
	if 0==idx {
		offset = left
	}
	
	r := ringX(idx)

	return (r.current + offset) & TM_MASK, r
}

func (me *Timer) Insert() {
	slot, r := me.findRing()
	
	me.node = r.Slot(slot).PushFront(me)
	me.idx 	= uint8(r.idx)
	me.slot = uint16(slot)
	me.flag |= TM_PENDING
	
	r.count++
	clock.count++
	clock.inserted++
	
	if clock.ticks > 0 {
		// test ???
	}
}

func (me *Timer) Remove() {
	r := ringX(me.Idx())
	
	if clock.ticks > 0 {
		// test ???
	}
	
	r.Slot(me.Slot()).Remove(me.node)
	me.node = nil
	me.flag &= ^TM_PENDING
	
	r.count--
	clock.count--
	clock.removed++
}

type tmRing struct {
	slot [TM_SLOT]list.List
	
	current uint32
	count 	uint32
	idx 	uint32
}

func (me *tmRing) Slot(idx uint32) *list.List {
	return &me.slot[idx]
}

type tmClock struct {
	init 	bool
	ticks 	uint64
	
	triggered_safe		uint64
	triggered_unsafe	uint64
	triggered_error		uint64
	triggered_ok		uint64
	inserted			uint64
	removed				uint64
	
	ring 	[TM_RING]tmRing
	count 	uint32
	unit 	uint32
}

var clock = &tmClock{}

func ring0() *tmRing {
	return &clock.ring[0]
}

func ringX(idx uint32) *tmRing {
	return &clock.ring[idx]
}

func ringMax() *tmRing {
	return &clock.ring[TM_RINGMAX]
}

func TmInit() {

}

func TmFini() {

}