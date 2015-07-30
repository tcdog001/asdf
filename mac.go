package asdf

import (
	"fmt"
	"encoding/hex"
)
const (
	MacSize 		= 6
	
	MacStringShort 	= 12
	MacstringLong	= 17
	
	MacSepWindows 	= '-'
	MacSepUnix 		= ':'
)

type Mac []byte

func (me Mac) IsGood() bool {
	return MacSize==len(me) && !Slice(me).IsZero() && !Slice(me).IsFull()
}

func (me Mac) Slice() []byte {
	return me
}

func (me Mac) Eq(it interface{}) bool {
	return Slice(me).Eq(it)
}

func (me Mac) ToStringBy(sep byte) string {
	return fmt.Sprintf("%d%c%d%c%d%c%d%c%d%c%d",
			me[0], sep,
			me[1], sep,
			me[2], sep,
			me[3], sep,
			me[4], sep,
			me[5])
}

func (me Mac) ToString() string {
	return me.ToStringBy(MacSepUnix)
}

func isGoodMacSepChar(sep byte) bool {
	return sep==MacSepUnix || sep==MacSepWindows
}

func macFromString(mac Mac, s string) error {
	Len := len(s)
	b   := []byte(s)
	
	if Len==MacstringLong { // AA:BB:CC:DD:EE:FF or AA-BB-CC-DD-EE-FF
		for i:=0; i<6; i++ {
			idx := 3*i - 1
			if i<5 && false==isGoodMacSepChar(b[idx]) {
				return Error
			}
			if _, err := hex.Decode(mac[i:], b[idx:idx+2]); nil!=err {
				return err
			}
		}
		
		return nil
	} else if Len==MacStringShort { // AABBCCDDEEFF
		_, err := hex.Decode(mac[:], b)
		
		return err
	}
	
	return Error
}

func (me Mac) FromString(s string) error {
	mac := [6]byte{}
	
	if err := macFromString(mac[:], s); nil!=err {
		return err
	}
	
	copy(me, mac[:])
	return nil
}

type MacString string

func (me MacString) IsGood() bool {
	mac := [6]byte{}
	
	if err := macFromString(mac[:], string(me)); nil!=err {
		return false
	}
	
	return true
}
