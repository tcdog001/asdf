package asdf

import (
	"errors"
)

var Error = errors.New(Empty)

var ErrNoSupport = errors.New("no support")
var ErrNoFound = errors.New("no found")
var ErrNoMatch = errors.New("no match")
var ErrNoSpace = errors.New("no space")

var ErrBadObj = errors.New("bad obj")
var ErrNilObj = errors.New("nil obj")
var ErrBadIdx = errors.New("bad idx")
var ErrBadIntf= errors.New("bad interface")
var ErrBadType= errors.New("bad type")
var ErrNilBuffer = errors.New("nil buffer")

var ErrTooShortBuffer = errors.New("too short buffer")
var ErrBadPktLen = errors.New("invalid packet length")
var ErrPktLenNoMatchBufferLen = errors.New("packet length not match buffer length")
var ErrBadPktDir = errors.New("bad packet dir")
