package asdf

import (
	"errors"
)

var Error 			= errors.New(Empty)

var ErrNoSupport	= errors.New("no support")
var ErrNoFound 		= errors.New("no found")
var ErrNoMatch		= errors.New("no match")
