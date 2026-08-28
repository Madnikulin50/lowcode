package sdk

import (
	"encoding/json"
	"strings"
)

// StartRequest is the POST /api/jobs body. Canonical envelope fields are
// lifted out; everything else (cidr, sourceID, …) lands in Params.
type StartRequest struct {
	Operation   string
	JobID       string
	RecordID    string
	NamespaceID FlexID
	Token       string
	CallbackURL string
	Params      map[string]any
}

func DecodeStartRequest(raw []byte) (StartRequest, error) {
	var req StartRequest
	if len(strings.TrimSpace(string(raw))) == 0 {
		req.Params = map[string]any{}
		return req, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return req, err
	}
	req.Params = map[string]any{}
	for k, v := range m {
		switch k {
		case "operation":
			req.Operation = stringify(v)
		case "jobID", "jobId":
			req.JobID = stringify(v)
		case "recordID", "createdRecordID", "scanRecordID", "jobRecordID":
			if req.RecordID == "" {
				req.RecordID = stringify(v)
			}
		case "namespaceID":
			req.NamespaceID = FlexID(ParseID(v))
		case "token":
			req.Token = stringify(v)
		case "callbackUrl", "callbackURL":
			req.CallbackURL = stringify(v)
		default:
			req.Params[k] = v
		}
	}
	if req.RecordID == "" {
		req.RecordID = req.JobID
	}
	return req, nil
}

func (r StartRequest) Param(key string) string {
	if r.Params == nil {
		return ""
	}
	return stringify(r.Params[key])
}

func (r StartRequest) ParamID(key string) uint64 {
	if r.Params == nil {
		return 0
	}
	return ParseID(r.Params[key])
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return strings.TrimSpace(strings.Trim(stringifyFallback(t), `"`))
		}
		s := strings.TrimSpace(string(b))
		if len(s) >= 2 && s[0] == '"' {
			var out string
			if json.Unmarshal(b, &out) == nil {
				return strings.TrimSpace(out)
			}
		}
		return strings.TrimSpace(s)
	}
}

func stringifyFallback(v any) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(jsonNumber(v), "\n", ""), " ", ""))
}

func jsonNumber(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
