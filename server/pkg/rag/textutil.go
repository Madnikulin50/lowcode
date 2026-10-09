package rag

import (
	"strings"
	"unicode/utf8"
)

// SanitizeExtractedText strips template delimiters so rulechain interpolation
// cannot be broken by CAD/Office payload, then collapses huge blank runs.
func SanitizeExtractedText(s string) string {
	s = strings.ReplaceAll(s, "{{", "{ {")
	s = strings.ReplaceAll(s, "}}", "} }")
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.TrimSpace(s)
	return s
}

// TruncateRunes cuts s to at most max runes, appending an ellipsis when clipped.
func TruncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	if max < 4 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

func looksBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	n := len(data)
	if n > 1024 {
		n = 1024
	}
	nul := 0
	nonPrint := 0
	for _, b := range data[:n] {
		if b == 0 {
			nul++
		}
		if b < 9 || (b > 13 && b < 32) {
			nonPrint++
		}
	}
	if nul > 4 {
		return true
	}
	return nonPrint*10 > n
}
