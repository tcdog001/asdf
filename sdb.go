package asdf

import (
	"time"
//	"fmt"
)

const (
	SDB_ACT_BEGIN 	= 0
	
	// input:key
	// output:entry+error
	SDB_ACT_GET 	= 0
	
	// input:key+entry
	// output:error
	SDB_ACT_PUT 	= 1
	
	// input:key+entry
	// output:error
	SDB_ACT_CREATE 	= 2
	
	// input:key
	// output:error
	SDB_ACT_DELETE 	= 3
	
	// input:key+entry
	// output:error
	SDB_ACT_UPDATE 	= 4
	
	SDB_ACT_END 	= 5
)

const (
	SDB_TIMER_BEGIN	= 0
	
	SDB_TIMER_IDLE 	= 0
	SDB_TIMER_HOLD 	= 1
	
	SDB_TIMER_END 	= 2
)

type SdbPire struct {
	Key 	interface{}
	Entry	interface{}
}

type SdbOps struct {
	Max 	uint
	Idle 	uint /* ms */
	Hold 	uint /* ms */
	Unit 	uint /* ms */
	
	Create func (entry interface{}) error
	Delete func (Key interface{}) error
	Update func (entry interface{}) error
}

type SdbResponse struct {
	SdbPire
	
	Error 	error
}

type SdbRequest struct {
	SdbPire
	
	Act 	int
	Chan 	chan SdbResponse
}

type sdb struct {
	SdbPire
	
	ref 	uint
	timer 	[SDB_TIMER_END]Timer
	
	SDB 	*SDB
}

func (me *sdb) Name(tag uint) string {
	var name string
	
	switch tag {
		case SDB_TIMER_IDLE: name = "idler"
		case SDB_TIMER_HOLD: name = "holder"
	}
	
	return name
}

func (me *sdb) GetTimer(tag uint) *Timer {
	return &me.timer[tag]
}

func (me *sdb) holder() *Timer {
	return &me.timer[SDB_TIMER_HOLD]
}

func (me *sdb) idler() *Timer {
	return &me.timer[SDB_TIMER_IDLE]
}

func (me *sdb) delete() {
	me.SDB.db[me.Key] = nil
	me.SDB 		= nil
	me.Key 		= nil
	me.Entry 	= nil

	// when delete
	// remove idle timer
	// remove hold timer
	me.idler().Remove()
	me.holder().Remove()
}

type SDB struct {
	SdbOps
	
	debug 	bool
	db 		map[interface{}]*sdb
	ch 		chan SdbRequest
	clock 	*Clock
}

func (me *SDB) Init (ops SdbOps) {
	me.SdbOps = ops
	me.ch = make(chan SdbRequest, me.Max/100)
	me.db = make(map[interface{}]*sdb, me.Max)
	
	me.clock = TmClock(me.Unit, &me.debug)
}

func (me *SDB) handle (q *SdbRequest) {
	p := SdbResponse{}
	
	switch q.Act {
		case SDB_ACT_GET:	me.get(q, &p)
		case SDB_ACT_PUT:	me.put(q, &p)
		case SDB_ACT_CREATE:	me.create(q, &p)
		case SDB_ACT_DELETE:	me.delete(q, &p)
		case SDB_ACT_UPDATE:	me.update(q, &p)
	}
	
	q.Chan<-p
}

func (me *SDB) get (q *SdbRequest, p *SdbResponse) {
	sdb, ok := me.db[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if sdb.ref > 0 {
		p.Error = ErrHolding
	} else {
		sdb.ref = 1
		p.Entry = sdb.Entry

		// when get
		// insert hold timer
		// change idle timer
		me.clock.Insert(sdb, SDB_TIMER_HOLD, me.Hold, holdTimeout, true)
		sdb.idler().Change(me.Idle)
	}
}

func (me *SDB) put (q *SdbRequest, p *SdbResponse) {
	sdb, ok := me.db[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if sdb.ref > 0 {
		sdb.ref = 0
		
		// when put
		// remove hold timer
		// change idle timer
		sdb.holder().Remove()
		sdb.idler().Change(me.Idle)
	} else {
		// do nothing
	}
}

func (me *SDB) create (q *SdbRequest, p *SdbResponse) {
	if _, ok := me.db[q.Key]; ok {
		p.Error = ErrExist
	} else if p.Error = me.Create(q.Entry); nil==p.Error {
		sdb := &sdb{
			SdbPire:SdbPire{
				Key:q.Key,
				Entry:q.Entry,
			},
			SDB:me,
		}
		
		// when create
		// insert idle timer
		me.clock.Insert(sdb, SDB_TIMER_IDLE, me.Idle, idleTimeout, true)
		me.db[q.Key] = sdb
	}
}

func (me *SDB) delete (q *SdbRequest, p *SdbResponse) {
	if sdb, ok := me.db[q.Key]; !ok {
		p.Error = ErrNoExist
	} else {
		sdb.delete()		
		
		p.Error = me.Delete(q.Key)
	}
}

func (me *SDB) update (q *SdbRequest, p *SdbResponse) {
	if sdb, ok := me.db[q.Key]; !ok {
		p.Error = ErrNoExist
	} else {
		// when update
		// change idle timer
		sdb.idler().Change(me.Idle)
		
		p.Error = me.Update(q.Entry)
	}
}

func idleTimeout(entry interface{}) (bool, error) {
	sdb, ok := entry.(*sdb)
	if !ok {
		return true, ErrBadType
	}
	
	sdb.delete()
	
	return false, nil
}

func holdTimeout(entry interface{}) (bool, error) {
	sdb, ok := entry.(*sdb)
	if !ok {
		return true, ErrBadType
	}
	
	sdb.holder().Remove()
	sdb.ref = 0
	
	return false, nil
}

// go it
func SdbRun(sdb *SDB) {
	for {
		select {
			case q := <- sdb.ch:
				sdb.handle(&q)
			case <- time.After(time.Duration(sdb.Unit)*1000*1000):
				sdb.clock.Trigger(1)
		}
	}
}
