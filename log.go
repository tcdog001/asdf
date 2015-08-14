package asdf

import (
	"fmt"
)

type logDefault struct{}

var defaultLog logDefault

func (me *logDefault) Emerg(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Alert(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Crit(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Error(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Warning(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Notice(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Info(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

func (me *logDefault) Debug(format string, v ...interface{}) {
	fmt.Printf(format+Crlf, v...)
}

var Log ILogger = &defaultLog

func SetLogger(r ILogger) {
	Log = r
}