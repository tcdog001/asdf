package asdf

import (
	"errors"
)

var (
	Error 		= errors.New(Empty)

	ErrEmpty    = errors.New("empty")
	ErrFull     = errors.New("full")
	ErrExist 	= errors.New("exist")
	ErrHolding	= errors.New("holding")
	ErrPending	= errors.New("pending")

	ErrNoPending= errors.New("no pending")
	ErrNoSupport= errors.New("no support")
	ErrNoExist 	= errors.New("no exist")
	ErrNoFound 	= errors.New("no found")
	ErrNoMatch 	= errors.New("no match")
	ErrNoSpace 	= errors.New("no space")
	ErrNoPermit	= errors.New("no permit")

	ErrBadObj 	= errors.New("bad obj")
	ErrNilObj 	= errors.New("nil obj")
	ErrBadIdx 	= errors.New("bad idx")
	ErrBadIntf	= errors.New("bad interface")
	ErrBadType	= errors.New("bad type")
	ErrNilBuffer= errors.New("nil buffer")
	ErrNilIntf  = errors.New("nil interface")

	ErrTooMore 			= errors.New("too more")
	ErrTooShortBuffer 	= errors.New("too short buffer")
	ErrBadPktLen		= errors.New("invalid packet length")
	ErrPktLenNoMatchBufferLen = errors.New("packet length not match buffer length")
	ErrBadPktDir		= errors.New("bad packet dir")
)
