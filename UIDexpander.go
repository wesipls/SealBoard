package main

import (
	"fmt"
	"os"
	"strings"
)

// expandUIDVariable replaces ${UID} with the current user's UID
func expandUIDVariable(path string) string {
	if strings.Contains(path, "${UID}") {
		uid := os.Getuid()
		return strings.ReplaceAll(path, "${UID}", fmt.Sprintf("%d", uid))
	}
	return path
}

