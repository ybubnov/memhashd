package log

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

const (
	LevelEmerg = iota
	LevelAlert
	LevelCrit
	LevelErr
	LevelWarn
	LevelNotice
	LevelInfo
	LevelDebug
)

var (
	logger   Logger
	loggerMu sync.RWMutex
)

func init() {
	logger = StdoutLogger()
}

type Logger interface {
	Log(level int, text string) error
}

type LoggerFunc func(int, string) error

func (fn LoggerFunc) Log(level int, text string) error {
	return fn(level, text)
}

func StdoutLogger() Logger {
	return LoggerFunc(func(level int, text string) error {
		_, err := fmt.Fprint(os.Stdout, logPrefix(), text)
		return err
	})
}

func DebugLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	DebugLog(event, text)
}

func DebugLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelDebug, text)
}

func InfoLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	InfoLog(event, text)
}

func InfoLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelInfo, text)
}

func ErrorLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	ErrorLog(event, text)
}

func ErrorLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelErr, text)
}

func FatalLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	FatalLog(event, text)
}

func FatalLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelEmerg, text)
	panic(text)
}

var (
	hostname, _ = os.Hostname()
	pid         = os.Getpid()
	proc        = path.Base(os.Args[0])
)

func logPrefix() string {
	now := time.Now().Format(time.StampMicro)
	return fmt.Sprintf("%s %s %s[%d]: ", now, hostname, proc, pid)
}
