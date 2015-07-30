package asdf

import (

)

type IpAddress uint32

func (me IpAddress) IsGood() bool {
	return true
}

func (me IpAddress) Int() int {
	return int(me)
}

func (me IpAddress) Eq(it interface{}) bool {
	return true
}

func (me IpAddress) ToString() string {
	return ""
}

func (me IpAddress) FromString(s string) error {
	return nil
}
