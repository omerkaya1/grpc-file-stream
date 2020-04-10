package internal

import "log"

type (
	// FormatLogger .
	FormatLogger interface {
		Printf(format string, v ...interface{})
		Fatalf(format string, v ...interface{})
	}
	// SimpleLogger .
	SimpleLogger interface {
		Print(v ...interface{})
		Fatal(v ...interface{})
	}
	// MetaLogger .
	MetaLogger interface {
		FormatLogger
		SimpleLogger
	}
	// Log is a simple wrapper object for basic logging
	Log struct{}
)

// NewLog returns a new instance of Log
func NewLog(level int) *Log {
	return &Log{}
}

func (l *Log) Printf(format string, v ...interface{}) {
	log.Printf(format, v)
}

func (l *Log) Print(v ...interface{}) {
	log.Print(v)
}

func (l *Log) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}

func (l *Log) Fatal(v ...interface{}) {
	log.Fatal(v)
}
