package agent

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MatchCron reports whether t matches a 5-field cron (min hour dom mon dow)
// or a shorthand (@hourly, @daily, @weekly, @monthly). Supports *, N, */N, A-B, lists.
func MatchCron(expr string, t time.Time) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}
	switch strings.ToLower(expr) {
	case "@hourly":
		expr = "0 * * * *"
	case "@daily", "@midnight":
		expr = "0 0 * * *"
	case "@weekly":
		expr = "0 0 * * 0"
	case "@monthly":
		expr = "0 0 1 * *"
	}
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}
	t = t.In(time.Local)
	return fieldMatch(parts[0], t.Minute(), 0, 59) &&
		fieldMatch(parts[1], t.Hour(), 0, 23) &&
		fieldMatch(parts[2], t.Day(), 1, 31) &&
		fieldMatch(parts[3], int(t.Month()), 1, 12) &&
		fieldMatch(parts[4], int(t.Weekday()), 0, 6)
}

func fieldMatch(field string, value, min, max int) bool {
	for _, tok := range strings.Split(field, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if tok == "*" {
			return true
		}
		if strings.HasPrefix(tok, "*/") {
			step, err := strconv.Atoi(tok[2:])
			if err != nil || step <= 0 {
				continue
			}
			if (value-min)%step == 0 {
				return true
			}
			continue
		}
		if a, b, ok := strings.Cut(tok, "-"); ok {
			lo, err1 := strconv.Atoi(a)
			hi, err2 := strconv.Atoi(b)
			if err1 == nil && err2 == nil && value >= lo && value <= hi {
				return true
			}
			continue
		}
		n, err := strconv.Atoi(tok)
		if err == nil && n == value {
			return true
		}
	}
	return false
}

func CronValid(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("empty cron")
	}
	switch strings.ToLower(expr) {
	case "@hourly", "@daily", "@midnight", "@weekly", "@monthly":
		return nil
	}
	if len(strings.Fields(expr)) != 5 {
		return fmt.Errorf("cron must have 5 fields or a shorthand")
	}
	return nil
}
