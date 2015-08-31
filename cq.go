package asdf

type Cq struct {
	reader 	uint
	writer 	uint
	size 	uint
	array	[]interface{}
}

func (me *Cq) align(idx uint) uint {
	return Align(idx, me.size)
}

func (me *Cq) Count() uint {
	return me.align(me.size + me.writer + me.reader)
}

func (me *Cq) isWriteable(idx uint) bool {
	idx = me.align(idx)
	
	if me.reader == me.writer {
		return true
	} else if me.reader < me.writer {
		return idx >= me.writer || idx < me.reader
	} else { // reader > writer
		return idx >= me.writer && idx < me.reader
	}
}

func (me *Cq) isReadable(idx uint) bool {
	return false==me.isWriteable(idx)
}

func (me *Cq) get(idx uint) interface{} {
	return me.array[me.align(idx)]
}

func (me *Cq) set(idx uint, entry interface{}) {
	me.array[me.align(idx)] = entry
}

func (me *Cq) Read() (interface{}, error) {
	if !me.isReadable(me.reader) {
		return nil, ErrEmpty
	}
	
	entry := me.get(me.reader)
	me.reader = me.align(me.reader + 1)
	
	return entry, nil
}

func (me *Cq) Write(entry interface{}) error {
	if !me.isWriteable(me.writer) {
		return ErrFull
	}
	
	me.set(me.writer, entry)
	me.writer = me.align(me.writer + 1)
	
	return nil
}

func (me *Cq) Init(size uint) {
	me.reader 	= 0
	me.writer 	= 0
	me.size 	= size
	me.array 	= make([]interface{}, size)
}

func CqNew(size uint) *Cq {
	cq := &Cq{}
	
	cq.Init(size)
	
	return cq
}