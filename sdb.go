package asdf

import (
	"time"
//	"fmt"
)

const (
	SDB_ACT_BEGIN 	= 0
	
	// input:key
	// output:entry+error
	SDB_ACT_GETONLY	= 0
	
	// input:key
	// output:entry+error
	SDB_ACT_GET 	= 1
	
	// input:key+entry
	// output:error
	SDB_ACT_PUT 	= 2
	
	// input:key+entry
	// output:error
	SDB_ACT_CREATE 	= 3
	
	// input:key
	// output:error
	SDB_ACT_DELETE 	= 4
	
	// input:key+entry
	// output:error
	SDB_ACT_UPDATE 	= 5
	
	SDB_ACT_END 	= 6
)

const (
	SDB_TIMER_BEGIN	= 0
	
	SDB_TIMER_IDLE 	= 0
	SDB_TIMER_HOLD 	= 1
	
	SDB_TIMER_END 	= 2
)

type SdbOps struct {
	Max 	uint
	Idle 	uint /* ms */
	Hold 	uint /* ms */
	Unit 	uint /* ms */
	
	Create func (entry interface{}) error
	Delete func (Key interface{}) error
	Update func (entry interface{}) error
}

type SdbPire struct {
	Key 	interface{}
	Entry	interface{}
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

func (me *sdb) Name(tidx uint) string {
	var name string
	
	switch tidx {
		case SDB_TIMER_IDLE: name = "idler"
		case SDB_TIMER_HOLD: name = "holder"
	}
	
	return name
}

func (me *sdb) GetTimer(tidx uint) *Timer {
	return &me.timer[tidx]
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
	
	Log.Info("sdb init")
}

func (me *SDB) handle (q *SdbRequest) {
	p := SdbResponse{}
	
	switch q.Act {
		case SDB_ACT_GETONLY:	me.getOnly(q, &p)
		case SDB_ACT_GET:		me.get(q, &p)
		case SDB_ACT_PUT:		me.put(q, &p)
		case SDB_ACT_CREATE:	me.create(q, &p)
		case SDB_ACT_DELETE:	me.delete(q, &p)
		case SDB_ACT_UPDATE:	me.update(q, &p)
	}
	
	q.Chan<-p
}

func (me *SDB) getEx (q *SdbRequest, p *SdbResponse, only bool) {
	sdb, ok := me.db[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if only && sdb.ref > 0 {
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

func (me *SDB) getOnly (q *SdbRequest, p *SdbResponse) {
	me.getEx(q, p, true)
}

func (me *SDB) get (q *SdbRequest, p *SdbResponse) {
	me.getEx(q, p, false)
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
		// just log
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
	} else if sdb.ref > 0 {
		p.Error = ErrHolding
	} else {
		sdb.delete()		
		
		p.Error = me.Delete(q.Key)
	}
}

func (me *SDB) update (q *SdbRequest, p *SdbResponse) {
	if sdb, ok := me.db[q.Key]; !ok {
		p.Error = ErrNoExist
	} else if sdb.ref > 0 {
		p.Error = ErrHolding
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
	
	Log.Info("sdb %v idle timeout", entry)
	
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
	
	Log.Info("sdb %v hold timeout", entry)
	
	return false, nil
}

// go it
func SdbRun(sdb *SDB) {
	Log.Info("sdb run")
	
	timeout := time.After(time.Duration(sdb.Unit) * time.Millisecond)
	
	for {
		select {
			case q := <- sdb.ch:
				sdb.handle(&q)
			case <- timeout:
				sdb.clock.Trigger(1)
		}
	}
}
