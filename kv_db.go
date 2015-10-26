package asdf

import (
	"time"
//	"fmt"
)

type kvDbCmd uint

const (
	// get, NOT change it
	// input:key
	// output:entry+error
	kv_CMD_GET		kvDbCmd = 0
	
	// hold, need change it
	// input:key
	// output:entry+error
	kv_CMD_HOLD 	kvDbCmd = 1
	
	// input:key+entry
	// output:error
	kv_CMD_PUT 		kvDbCmd = 2
	
	// input:key+entry
	// output:error
	kv_CMD_CREATE 	kvDbCmd = 3
	
	// input:key
	// output:error
	kv_CMD_DELETE 	kvDbCmd = 4
	
	// input:key+entry
	// output:error
	kv_CMD_UPDATE 	kvDbCmd = 5
)

const (
	kv_TIMER_IDLE	= 0
	kv_TIMER_HOLD 	= 1
	
	kv_TIMER_END	= 2
)

type KvResponse struct {
	Key 	interface{}
	Entry	interface{}
	
	Error 	error
}

type kvRequest struct {
	Key 	interface{}
	Entry	interface{}
	
	cmd 	kvDbCmd
	ch 		chan KvResponse
}

type kvCache struct {
	Key 	interface{}
	Entry	interface{}
	
	ref 	uint
	timer 	[kv_TIMER_END]ITimer
	
	db 		*kvDB
}

// ITimerProxy
func (me *kvCache) TimerName(tidx uint) string {
	var name string
	
	switch tidx {
		case kv_TIMER_IDLE: name = "idler"
		case kv_TIMER_HOLD: name = "holder"
	}
	
	return name
}

// ITimerProxy
func (me *kvCache) GetTimer(tidx uint) ITimer {
	return me.timer[tidx]
}

// ITimerProxy
func (me *kvCache) SetTimer(tidx uint, timer ITimer) {
	me.timer[tidx] = timer
}

func (me *kvCache) holder() ITimer {
	return me.timer[kv_TIMER_HOLD]
}

func (me *kvCache) idler() ITimer {
	return me.timer[kv_TIMER_IDLE]
}

func (me *kvCache) del() {
	// when delete
	// remove idle timer
	// remove hold timer
	me.idler().Remove()
	me.holder().Remove()
	
	me.db.cache[me.Key] = nil
	me.db 		= nil
	me.Key 		= nil
	me.Entry 	= nil
}

type kvDB struct {
	ops 	KvDbOps
	
	ch 		chan kvRequest
	
	cache 	map[interface{}]*kvCache
	clock 	IClock
}

func (me *kvDB) handle (q *kvRequest) {
	p := KvResponse{}
	
	switch q.cmd {
		case kv_CMD_GET:		me.__get(q, &p)
		case kv_CMD_HOLD:		me.__hold(q, &p)
		case kv_CMD_PUT:		me.__put(q, &p)
		case kv_CMD_CREATE:	me.__create(q, &p)
		case kv_CMD_DELETE:	me.__delete(q, &p)
		case kv_CMD_UPDATE:	me.__update(q, &p)
		default: p.Error = ErrNoSupport
	}
	
	q.ch<-p
}

func (me *kvDB) __get (q *kvRequest, p *KvResponse) {
	e, ok := me.cache[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else {
		p.Entry = e.Entry
		
		// when get
		// update idle timer
		e.idler().Change(me.ops.Idle)
	}
}

func (me *kvDB) __hold (q *kvRequest, p *KvResponse) {
	e, ok := me.cache[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if e.ref > 0 { // is holding
		p.Error = ErrHolding
	} else { // not holding
		e.ref = 1
		p.Entry = e.Entry
		
		// when hold
		// insert hold timer
		// update idle timer
		e.timer[kv_TIMER_HOLD], _ =
			me.clock.Insert(e, kv_TIMER_HOLD, me.ops.Hold, holdTimeout, true)
		e.idler().Change(me.ops.Idle)
	}
}

func (me *kvDB) __put (q *kvRequest, p *KvResponse) {
	e, ok := me.cache[q.Key]
	if !ok {
		p.Error = ErrNoExist
	} else if e.ref > 0 {
		e.ref = 0
		
		// when put
		// remove hold timer
		// change idle timer
		e.holder().Remove()
		e.idler().Change(me.ops.Idle)
	} else {
		// just log
	}
}

func (me *kvDB) __create (q *kvRequest, p *KvResponse) {
	if _, ok := me.cache[q.Key]; ok {
		p.Error = ErrExist
	} else if p.Error = me.ops.Create(q.Entry); nil==p.Error {
		e := &kvCache{
			Key:	q.Key,
			Entry:	q.Entry,
			db:		me,
		}
		
		// when create
		// insert idle timer
		e.timer[kv_TIMER_IDLE], _ =
			me.clock.Insert(e, kv_TIMER_IDLE, me.ops.Idle, idleTimeout, true)
		me.cache[q.Key] = e
	}
}

func (me *kvDB) __delete (q *kvRequest, p *KvResponse) {
	if e, ok := me.cache[q.Key]; !ok {
		p.Error = ErrNoExist
	} else if e.ref > 0 {
		p.Error = ErrHolding
	} else {
		e.del()		
		
		p.Error = me.ops.Delete(q.Key)
	}
}

func (me *kvDB) __update (q *kvRequest, p *KvResponse) {
	if e, ok := me.cache[q.Key]; !ok {
		p.Error = ErrNoExist
	} else if e.ref > 0 {
		p.Error = ErrHolding
	} else {
		// when update
		// change idle timer
		e.idler().Change(me.ops.Idle)
		
		p.Error = me.ops.Update(q.Entry)
	}
}

func idleTimeout(proxy ITimerProxy) (bool, error) {
	e, ok := proxy.(*kvCache)
	if !ok {
		return true, ErrBadType
	}
	
	Log.Info("kvCache %v idle timeout", e)
	
	e.del()
	
	return false, nil
}

func holdTimeout(proxy ITimerProxy) (bool, error) {
	e, ok := proxy.(*kvCache)
	if !ok {
		return true, ErrBadType
	}
	
	e.holder().Remove()
	e.ref = 0
	
	Log.Info("kvCache %v hold timeout", e)
	
	return false, nil
}

type IKvDB interface {
	Get (
		ch chan KvResponse, 
		key interface{}) (interface{}, error)
	
	Hold (
		ch chan KvResponse, 
		key interface{}) (interface{}, error)
	
	Put (
		ch chan KvResponse, 
		key interface{}, 
		entry interface{}) error
	
	Create (
		ch chan KvResponse, 
		key interface{}, 
		entry interface{}) error
	
	Delete (
		ch chan KvResponse, 
		key interface{}) error
	
	Update (
		ch chan KvResponse, 
		key interface{}, 
		entry interface{}) error
	
	// go it
	Run()
}

func (me *kvDB) request (act kvDbCmd,
							ch chan KvResponse,
							key interface{}, 
							entry interface{}) {
	me.ch <- kvRequest{
		Key: key,
		Entry: entry,
		
		ch: ch,
		cmd: act,
	}
}

func (me *kvDB) Get (ch chan KvResponse, 
						key interface{}) (interface{}, error) {
	me.request(kv_CMD_GET, ch, key, nil)
	
	p := <- ch

	return p.Entry, p.Error
}

func (me *kvDB) Hold (ch chan KvResponse, 
						key interface{}) (interface{}, error) {
	me.request(kv_CMD_HOLD, ch, key, nil)
	
	p := <- ch
	
	return p.Entry, p.Error
}

func (me *kvDB) Put (ch chan KvResponse, 
						key interface{}, 
						entry interface{}) error {
	me.request(kv_CMD_PUT, ch, key, entry)
	
	p := <- ch

	return p.Error
}

func (me *kvDB) Create (	ch chan KvResponse, 
							key interface{}, 
							entry interface{}) error {
	me.request(kv_CMD_CREATE, ch, key, entry)
	
	p := <- ch

	return p.Error
}

func (me *kvDB) Delete (	ch chan KvResponse, 
							key interface{}) error {
	me.request(kv_CMD_DELETE, ch, key, nil)
	
	p := <- ch

	return p.Error
}

func (me *kvDB) Update (	ch chan KvResponse, 
							key interface{}, 
							entry interface{}) error {
	me.request(kv_CMD_UPDATE, ch, key, entry)
	
	p := <- ch

	return p.Error
}

func (me *kvDB) Run() {
	Log.Info("kvDB run")
	
	timeout := time.After(time.Duration(me.ops.Unit) * time.Millisecond)
	
	for {
		select {
			case q := <- me.ch:
				me.handle(&q)
			case <- timeout:
				me.clock.Trigger(1)
		}
	}
}

type KvDbOps struct {
	Max 	uint
	Cap 	uint /* request channel cap */
	Idle 	uint /* ms */
	Hold 	uint /* ms */
	Unit 	uint /* ms */
	
	Create func (entry interface{}) error
	Delete func (key interface{}) error
	Update func (entry interface{}) error
}

const (
	kv_deft_MAX 	= 1000*1000
	kv_min_MAX 		= 10*1000
	kv_max_MAX 		= 10*1000*1000
	
	kv_deft_CAP 	= 1000
	kv_min_CAP 		= 10
	kv_max_CAP 		= 10*1000
	
	kv_deft_IDLE 	= 300*1000
	kv_min_IDLE 	= 3*1000
	kv_max_IDLE		= 3600*1000
	
	kv_deft_HOLD 	= 30*1000
	kv_min_HOLD 	= 3*1000
	kv_max_HOLD 	= 300*1000
	
	kv_deft_UNIT 	= 1*1000
	kv_min_UNIT 	= 10
	kv_max_UNIT 	= 10*1000
)

func DbCache (ops KvDbOps) IKvDB {
	if ops.Max < kv_min_MAX || ops.Max > kv_max_MAX {
		ops.Max = kv_deft_MAX
	}
	
	if ops.Cap < kv_min_CAP || ops.Cap > kv_max_CAP {
		ops.Cap = kv_deft_CAP
	}
	
	if ops.Idle < kv_min_IDLE || ops.Idle > kv_max_IDLE {
		ops.Idle = kv_deft_IDLE
	}
	
	if ops.Hold < kv_min_HOLD || ops.Hold > kv_max_HOLD {
		ops.Hold = kv_deft_HOLD
	}
	
	if ops.Unit < kv_min_UNIT || ops.Unit > kv_max_UNIT {
		ops.Unit = kv_deft_UNIT
	}
	
	db := &kvDB{
		ops: ops,
	}

	db.ch 		= make(chan kvRequest, ops.Cap)
	db.cache 	= make(map[interface{}]*kvCache, ops.Max)
	db.clock 	= TmClock(ops.Unit)
	
	Log.Info("kvDB init")
	
	return db
}

