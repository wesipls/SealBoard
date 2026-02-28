package util

import (
	"log"
)

// LogError logs an error with a standard prefix
func LogError(format string, a ...interface{}) {
	log.Printf("[ERROR] "+format, a...)
}

// LogWarning logs a warning with a standard prefix
func LogWarning(format string, a ...interface{}) {
	log.Printf("[WARN] "+format, a...)
}

// LogInfo logs an info message with a standard prefix
func LogInfo(format string, a ...interface{}) {
	log.Printf("[INFO] "+format, a...)
}

