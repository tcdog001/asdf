package asdf

import (
	"time"
//	"fmt"
)

type DbCacheAct uint

const (
	// get, NOT change it
	// input:key
	// output:entry+error
	DB_CACHE_GET	DbCacheAct = 0
	
	// hold, need change it
	// input:key
	// output:entry+error
	DB_CACHE_HOLD 		DbCacheAct = 1
	
	// input:key+entry
	// output:error
	DB_CACHE_PUT 		DbCacheAct = 2
	
	// input:key+entry
	// output:error
	DB_CACHE_CREATE 	DbCacheAct = 3
	
	// input:key
	// output:error
	DB_CACHE_DELETE 	DbCacheAct = 4
	
	// input:key+entry
	// output:error
	DB_CACHE_UPDATE 	DbCacheAct = 5
)

const (
	DB_CACHE_TIMER_BEGIN 	= 0
	
	DB_CACHE_TIMER_IDLE		= 0
	DB_CACHE_TIMER_HOLD 	= 1
	
	DB_CACHE_TIMER_END 		= 2
)

type DbCacheOps struct {
	Max 	uint
	Idle 	uint /* ms */
	Hold 	uint /* ms */
	Unit 	uint /* ms */
	
	Create func (entry interface{}) error
	Delete func (Key interface{}) error
	Update func (entry interface{}) error
}

type DbCachePire struct {
	Key 	interface{}
	Entry	interface{}
}

type DbCacheResponse struct {
	DbCachePire
	
	Error 	error
}

type DbCacheRequest struct {
	DbCachePire
	
	Act 	DbCacheAct
	Chan 	chan DbCacheResponse
}

type dbcache struct {
	DbCachePire
	
	ref 	uint
	timer 	[DB_CACHE_TIMER_END]Timer
	
	SDB 	*DbCache
}

func (me *dbcache) Name(tidx uint) string {
	var name string
	
	switch tidx {
		case DB_CACHE_TIMER_IDLE: name = "idler"
		case DB_CACHE_TIMER_HOLD: name = "holder"
	}
	
	return name
}

func (me *dbcache) GetTimer(tidx uint) *Timer {
	return &me.timer[tidx]
}

func (me *dbcache) holder() *Timer {
	return &me.timer[DB_CACHE_TIMER_HOLD]
}

func (me *dbcache) idler() *Timer {
	return &me.timer[DB_CACHE_TIMER_IDLE]
}

func (me *dbcache) delete() {
	// when delete
	// remove idle timer
	// remove hold timer
	me.idler().Remove()
	me.holder().Remove()
	
	me.SDB.cache[me.Key] = nil
	me.SDB 		= nil
	me.Key 		= nil
	me.Entry 	= nil
}

type DbCache struct {
	DbCacheOps
	
	Ch 		chan DbCacheRequest
	Debug 	bool
	
	cache 	map[interface{}]*dbcache
	clock 	*Clock
}

func (me *DbCache) handle (q *DbCacheRequest) {
	p := DbCacheResponse{}
	
	switch q.Act {
		case DB_CACHE_GET:		me.get(q, &p)
		case DB_CACHE_HOLD:		me.hold(q, &p)
		case DB_CACHE_PUT:		me.put(q, &p)
		case DB_CACHE_CREATE:	me.create(q, &p)
		case DB_CACHE_DELETE:	me.delete(q, &p)
		case DB_CACHE_UPDATE:	me.update(q, &p)
		default: p.Error = ErrNoSupport
	}
	
	q.Chan<-p
}

func (me *DbCache) get (q *DbCacheRequest, p *DbCacheResponse) {
	sdb, ok := me.cache[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else {
		p.Entry = sdb.Entry
		
		// when get
		// update idle timer
		sdb.idler().Change(me.Idle)
	}
}

func (me *DbCache) hold (q *DbCacheRequest, p *DbCacheResponse) {
	sdb, ok := me.cache[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if sdb.ref > 0 { // is holding
		p.Error = ErrHolding
	} else { // not holding
		sdb.ref = 1
		p.Entry = sdb.Entry
		
		// when hold
		// insert hold timer
		// update idle timer
		me.clock.Insert(sdb, DB_CACHE_TIMER_HOLD, me.Hold, holdTimeout, true)
		sdb.idler().Change(me.Idle)
	}
}

func (me *DbCache) put (q *DbCacheRequest, p *DbCacheResponse) {
	sdb, ok := me.cache[q.Key]
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

func (me *DbCache) create (q *DbCacheRequest, p *DbCacheResponse) {
	if _, ok := me.cache[q.Key]; ok {
		p.Error = ErrExist
	} else if p.Error = me.Create(q.Entry); nil==p.Error {
		sdb := &dbcache{
			DbCachePire:DbCachePire{
				Key:q.Key,
				Entry:q.Entry,
			},
			SDB:me,
		}
		
		// when create
		// insert idle timer
		me.clock.Insert(sdb, DB_CACHE_TIMER_IDLE, me.Idle, idleTimeout, true)
		me.cache[q.Key] = sdb
	}
}

func (me *DbCache) delete (q *DbCacheRequest, p *DbCacheResponse) {
	if sdb, ok := me.cache[q.Key]; !ok {
		p.Error = ErrNoExist
	} else if sdb.ref > 0 {
		p.Error = ErrHolding
	} else {
		sdb.delete()		
		
		p.Error = me.Delete(q.Key)
	}
}

func (me *DbCache) update (q *DbCacheRequest, p *DbCacheResponse) {
	if sdb, ok := me.cache[q.Key]; !ok {
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
	sdb, ok := entry.(*dbcache)
	if !ok {
		return true, ErrBadType
	}
	
	Log.Info("sdb %v idle timeout", entry)
	
	sdb.delete()
	
	return false, nil
}

func holdTimeout(entry interface{}) (bool, error) {
	sdb, ok := entry.(*dbcache)
	if !ok {
		return true, ErrBadType
	}
	
	sdb.holder().Remove()
	sdb.ref = 0
	
	Log.Info("sdb %v hold timeout", entry)
	
	return false, nil
}

func (me *DbCache) init () {
	me.Ch = make(chan DbCacheRequest, me.Max/100)
	me.cache = make(map[interface{}]*dbcache, me.Max)
	
	me.clock = TmClock(me.Unit, &me.Debug)
	
	Log.Info("sdb init")
}

// go it
func SdbRun(sdb *DbCache) {
	sdb.init()
	
	Log.Info("sdb run")
	
	timeout := time.After(time.Duration(sdb.Unit) * time.Millisecond)
	
	for {
		select {
			case q := <- sdb.Ch:
				sdb.handle(&q)
			case <- timeout:
				sdb.clock.Trigger(1)
		}
	}
}
