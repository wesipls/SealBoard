package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleAPIError logs, sets status code, and writes JSON error response
func HandleAPIError(w http.ResponseWriter, status int, label, errmsg string) {
	logMsg := errmsg
	if label != "" {
		logMsg = fmt.Sprintf("[%s] %s", label, errmsg)
	}
	LogError(logMsg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errArr := []map[string]interface{}{{
		"host":   label,
		"status": "error",
		"error":  errmsg,
	}}
	arrBytes, _ := json.Marshal(errArr)
	w.Write(arrBytes)
}

// APIError is a standard error object for API responses
func APIErrorArray(label, errmsg string) []byte {
	errArr := []map[string]interface{}{{
		"host":   label,
		"status": "error",
		"error":  errmsg,
	}}
	arrBytes, _ := json.Marshal(errArr)
	return arrBytes
}

// APIErrorObj returns a single error object as JSON
func APIErrorObj(label, errmsg string) []byte {
	errObj := map[string]interface{}{
		"host":   label,
		"status": "error",
		"error":  errmsg,
	}
	b, _ := json.Marshal(errObj)
	return b
}

// FormatErrorMsg returns formatted error string for API use
func FormatErrorMsg(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
