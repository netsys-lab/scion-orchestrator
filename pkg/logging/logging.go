package logging

import (
	"fmt"
	"os"
	"time"
)

const (
	LOG_LEVEL_DEBUG = 0
	LOG_LEVEL_INFO  = 1
	LOG_LEVEL_WARN  = 2
	LOG_LEVEL_ERROR = 3
	LOG_LEVEL_FATAL = 4
)

type Logger interface {
	Info(string, ...interface{})
	Error(string, ...interface{})
	Debug(string, ...interface{})
	Warn(string, ...interface{})
	Fatal(string, ...interface{})
}

type MainLogger struct {
	out   *os.File
	Level int
}

var Log *MainLogger

func SetLogOutput(logger *os.File) {
	Log.out = logger
}

func SetLogLevel(level int) {
	Log.Level = level
}

func init() {
	Log = &MainLogger{
		out:   os.Stdout,
		Level: LOG_LEVEL_INFO,
	}
}

func (l *MainLogger) Info(value string, args ...interface{}) {

	if l.Level > LOG_LEVEL_INFO {
		return
	}

	t := time.Now().Format(("2006-02-01 15:04:05"))
	l.out.WriteString(t + " INFO: " + value)
	for _, arg := range args {
		l.out.WriteString(fmt.Sprintf("%v", arg))
	}

	l.out.WriteString("\n")
}

func (l *MainLogger) Debug(value string, args ...interface{}) {

	if l.Level > LOG_LEVEL_DEBUG {
		return
	}

	t := time.Now().Format(("2006-02-01 15:04:05"))
	l.out.WriteString(t + " DEBUG: " + value)
	for _, arg := range args {
		l.out.WriteString(fmt.Sprintf("%v", arg))
	}

	l.out.WriteString("\n")
}

func (l *MainLogger) Error(value string, args ...interface{}) {

	if l.Level > LOG_LEVEL_ERROR {
		return
	}

	t := time.Now().Format(("2006-02-01 15:04:05"))
	l.out.WriteString(t + " ERROR: " + value)
	for _, arg := range args {
		l.out.WriteString(fmt.Sprintf("%v", arg))
	}

	l.out.WriteString("\n")
}

func (l *MainLogger) Warn(value string, args ...interface{}) {

	if l.Level > LOG_LEVEL_WARN {
		return
	}

	t := time.Now().Format(("2006-02-01 15:04:05"))
	l.out.WriteString(t + " WARN: " + value)
	for _, arg := range args {
		l.out.WriteString(fmt.Sprintf("%v", arg))
	}

	l.out.WriteString("\n")
}

func (l *MainLogger) Fatal(value string, args ...interface{}) {

	if l.Level > LOG_LEVEL_FATAL {
		return
	}

	t := time.Now().Format(("2006-02-01 15:04:05"))
	l.out.WriteString(t + " FATAL: " + value)
	for _, arg := range args {
		l.out.WriteString(fmt.Sprintf("%v", arg))
	}

	l.out.WriteString("\n")
	os.Exit(1)
}
