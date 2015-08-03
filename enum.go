package asdf

type IEnum interface {
	INumber
	IGood
	IToString
	// todo: IFromString
}

func IsGoodEnum(idx interface{}) bool {
	n, ok := idx.(INumber)
	if !ok {
		return false
	}
	v := n.Int()
	
	return v >= n.Begin() && v < n.End()
}

type EnumBinding []string

// todo: reutrn string and error
func (me EnumBinding) EntryShow(idx interface{}) string {
	if nil==me {
		return Empty
	}
	
	e, ok := idx.(IEnum);
	if !ok {
		return Empty
	}
	
	if !e.IsGood() {
		return Empty
	}
	
	return me[e.Int()]
}
