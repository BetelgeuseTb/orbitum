package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelError = "ERROR"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Issuer    string `json:"issuer,omitempty"`
}

var logLevel = LevelInfo

func SetLevel(level string) {
	logLevel = strings.ToUpper(level)
}

func shouldLog(level string) bool {
	switch logLevel {
	case LevelDebug:
		return true
	case LevelInfo:
		return level != LevelDebug
	case LevelError:
		return level == LevelError
	default:
		return false
	}
}

func LogDebug(message, issuer string) {
	if !shouldLog(LevelDebug) {
		return
	}
	logJSON(LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     LevelDebug,
		Message:   message,
		Issuer:    issuer,
	})
}

func LogInfo(message, issuer string) {
	if !shouldLog(LevelInfo) {
		return
	}
	logJSON(LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     LevelInfo,
		Message:   message,
		Issuer:    issuer,
	})
}

func LogError(message, issuer string) {
	if !shouldLog(LevelError) {
		return
	}
	logJSON(LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     LevelError,
		Message:   message,
		Issuer:    issuer,
	})
}

func logJSON(entry LogEntry) {
	jsonData, _ := json.Marshal(entry)
	fmt.Println(string(jsonData))
}
