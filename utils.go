package logger

import "strings"

// Cut the front of a path, and add "..." if it was cut.
func CutFrontPath(s string, length int) string {
	return CutStart(s, length, "/", true)
}

// Cut the string if it is longer than the specified length, and add "..." if it was cut.
func CutStart(s string, length int, delim string, prefixIfCut bool) string {
	if len(s) > length {
		var cut = len(s) - length
		s = s[cut:]
		var parts = strings.Split(s, delim)
		if len(parts) > 1 {
			if prefixIfCut {
				return "..." + delim + strings.Join(parts[1:], delim)
			}
			return "..." + strings.Join(parts[1:], delim)
		}
		return "..." + s
	}
	return s
}
