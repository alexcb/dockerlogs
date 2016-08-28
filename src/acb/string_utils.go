package dockerlogs

import "strings"

func PadLeft(s string, l int) string {
	needed := l - len(s)
	if needed > 0 {
		return strings.Repeat(" ", needed) + s
	}
	return s
}

func PadRight(s string, l int) string {
	needed := l - len(s)
	if needed > 0 {
		return s + strings.Repeat(" ", needed)
	}
	return s
}
