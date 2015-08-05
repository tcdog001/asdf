package asdf

type Slice []byte

func (me Slice) IsValue(Value byte) bool {
	for i := 0; i < len(me); i++ {
		if Value != me[i] {
			return false
		}
	}

	return true
}

func (me Slice) IsZero() bool {
	return me.IsValue(0)
}

func (me Slice) IsFull() bool {
	return me.IsValue(255)
}

func (me Slice) Slice() []byte {
	return me
}

func (me Slice) Eq(it interface{}) bool {
	v, ok := it.(ISlice)
	if !ok {
		return false
	}
	b := v.Slice()

	if len(me) != len(b) {
		return false
	}

	for i := 0; i < len(me); i++ {
		if me[i] != b[i] {
			return false
		}
	}

	return true
}
