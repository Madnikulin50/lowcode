package agent

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ResolveSecret(handle string) string {
	handle = strings.TrimSpace(handle)
	if handle == "" {
		return ""
	}
	keys := []string{
		"BACKUP_SECRET_" + handle,
		"BACKUP_SECRET_" + strings.ToUpper(handle),
		"BACKUP_SECRET_" + strings.ReplaceAll(strings.ToUpper(handle), "-", "_"),
	}
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on", "да":
		return true
	default:
		return false
	}
}

func ParseUint(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func ParseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func CleanRecordID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return ""
	}
	if strings.Contains(s, "{{") || strings.Contains(s, "${") {
		return ""
	}
	return s
}

func (r JobRequest) ResolvedSourceID() string {
	return firstNonEmpty(CleanRecordID(r.SourceID), CleanRecordID(r.Source))
}

func (r JobRequest) ResolvedPolicyID() string {
	return CleanRecordID(r.PolicyID)
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
