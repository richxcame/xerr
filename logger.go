package xerr

import "log"

type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type defaultLogger struct{}

func (d *defaultLogger) Error(args ...interface{}) {
	log.Println(args...)
}

func (d *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

var logger Logger = &defaultLogger{}

func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
}
