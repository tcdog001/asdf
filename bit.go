package asdf


func SetFlag(x, bit uint32) uint32 {
	return x | bit
}

func ClrFlag(x, bit uint32) uint32 {
	return x & ^bit
}

func HasFlag(x, flag uint32) bool {
	return flag==(x & flag) 
}

func SetBit(x, bit uint32) uint32 {
	return SetFlag(x, 1<<bit)
}

func ClrBit(x, bit uint32) uint32 {
	return ClrFlag(x, 1<<bit)
}

func HasBit(x, bit uint32) bool {
	return HasFlag(x, 1<<bit)
}

type BitMap []uint32

const BitMapSlot = 32

func (me BitMap) isGoodIdx(idx uint32) bool {
	return int(idx)<len(me)
}

func (me BitMap) SetBit(bit uint32) {
	idx := bit/BitMapSlot
	
	if me.isGoodIdx(idx) {
		SetBit(me[idx], bit % BitMapSlot)
	}
}

func (me BitMap) ClrBit(bit uint32) {
	idx := bit/BitMapSlot
	
	if me.isGoodIdx(idx) {
		ClrBit(me[idx], bit % BitMapSlot)
	}
}

func (me BitMap) HasBit(bit uint32) bool {
	idx := bit/BitMapSlot
	
	if !me.isGoodIdx(idx) {
		return false
	}
	
	return HasBit(me[idx], bit % BitMapSlot)
}
