package asdf

import (
	"fmt"
	"testing"
)

func TestMac(t *testing.T){
	bin := [6]byte{}
	mac := Mac(bin[:])
	macString := ""
	
	macString = "11-22-33-44-55-66"
	mac.FromString(macString)
	fmt.Println(macString, mac)
	
	macString = "11:22:33:44:55:66"
	mac.FromString(macString)
	fmt.Println(macString, mac)
	
	macString = "112233445566"
	mac.FromString(macString)
	fmt.Println(macString, mac)
	
	macString = "1122:3344:5566"
	mac.FromString(macString)
	fmt.Println(macString, mac)
}

