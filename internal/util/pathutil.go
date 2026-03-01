package util

import (
	"strings"
)

// SplitAndClean splits a path into segments and discards empty ones
func SplitAndClean(path string) []string {
	segments := strings.Split(path, "/")
	var out []string
	for _, seg := range segments {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

