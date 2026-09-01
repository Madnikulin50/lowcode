package rulesgo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// AttachmentBytes is one compose File-field attachment for document.extract.
type AttachmentBytes struct {
	ID   uint64
	Name string
	MIME string
	Data []byte
}

// AttachmentLoader loads a record attachment by ID.
type AttachmentLoader interface {
	Load(ctx context.Context, namespaceID, attachmentID uint64) (*AttachmentBytes, error)
}

// AttachmentIDsFromValue accepts a File field (id, newline/comma list, or array).
func AttachmentIDsFromValue(v interface{}) []uint64 {
	var ids []uint64
	switch t := v.(type) {
	case nil:
		return nil
	case uint64:
		if t > 0 {
			ids = append(ids, t)
		}
	case int:
		if t > 0 {
			ids = append(ids, uint64(t))
		}
	case int64:
		if t > 0 {
			ids = append(ids, uint64(t))
		}
	case float64:
		if t > 0 {
			ids = append(ids, uint64(t))
		}
	case string:
		ids = append(ids, parseIDList(t)...)
	case []string:
		for _, s := range t {
			ids = append(ids, parseIDList(s)...)
		}
	case []interface{}:
		for _, item := range t {
			ids = append(ids, AttachmentIDsFromValue(item)...)
		}
	default:
		ids = append(ids, parseIDList(fmt.Sprintf("%v", t))...)
	}
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func parseIDList(s string) []uint64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return nil
	}
	repl := strings.NewReplacer(",", " ", ";", " ", "\n", " ", "\r", " ", "\t", " ")
	var ids []uint64
	for _, p := range strings.Fields(repl.Replace(s)) {
		n, err := strconv.ParseUint(p, 10, 64)
		if err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	return ids
}

type extractStub struct{}

func (extractStub) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	out := map[string]interface{}{
		"extract_ok":     false,
		"extract_status": "failed",
		"extract_error":  "document.extract not configured",
	}
	if ec != nil {
		for k, v := range out {
			ec.Set(k, v)
		}
	}
	return out, nil
}
