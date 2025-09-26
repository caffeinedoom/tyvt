package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  Level
	logger *log.Logger
}

func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", format, args...)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", format, args...)
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log("WARN", format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", format, args...)
	}
}

func (l *Logger) log(level string, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s: %s", timestamp, level, message)
}