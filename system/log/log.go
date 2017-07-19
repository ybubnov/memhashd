package log

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

const (
	// LevelEmerg is an emergency log level.
	LevelEmerg = iota
	// LevelAlert is an alert log level.
	LevelAlert
	// LevelCrit is a critical log level.
	LevelCrit
	// LevelErr is an error logs level.
	LevelErr
	// LevelWarn is a warning log level.
	LevelWarn
	// LevelNotice is a notice log level.
	LevelNotice
	// LevelInfo is an info log level.
	LevelInfo
	// LevelDebug is a debug log level.
	LevelDebug
)

var (
	logger   Logger
	loggerMu sync.RWMutex
)

func init() {
	logger = StdoutLogger()
}

// Logger describes the types used for logging.
type Logger interface {
	Log(level int, text string) error
}

// LoggerFunc is a function adapter for Logger interface.
type LoggerFunc func(int, string) error

// Log implements Logger interface.
func (fn LoggerFunc) Log(level int, text string) error {
	return fn(level, text)
}

// StdoutLogger creates a logger with standard output log destination.
func StdoutLogger() Logger {
	return LoggerFunc(func(level int, text string) error {
		_, err := fmt.Fprint(os.Stdout, logPrefix(), text)
		return err
	})
}

// DebugLogf logs a message at level Debug on the standard logger.
func DebugLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	DebugLog(event, text)
}

// DebugLog logs a message at level Debug on the standard logger.
func DebugLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelDebug, text)
}

// InfoLogf logs a message at level Info on the standard logger.
func InfoLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	InfoLog(event, text)
}

// InfoLog logs a message at level Info on the standard logger.
func InfoLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelInfo, text)
}

// ErrorLogf logs a message at level Error on the standard logger.
func ErrorLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	ErrorLog(event, text)
}

// ErrorLog logs a message at level Error on the standard logger.
func ErrorLog(event string, args ...interface{}) {
	text := fmt.Sprintln(event, fmt.Sprint(args...))
	logger.Log(LevelErr, text)
}

// FatalLogf logs a message at level Fatal on the standard logger.
func FatalLogf(event, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	FatalLog(event, text)
}

// FatalLog logs a message at level Fatal on the standard logger.
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
