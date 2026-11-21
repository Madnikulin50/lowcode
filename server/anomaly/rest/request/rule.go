package request

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from anomaly/rest.yaml

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// dummy vars to prevent
// unused imports complain
var (
	_ = chi.URLParam
	_ = multipart.ErrMessageTooLarge
	_ = payload.ParseUint64s
	_ = strings.ToLower
	_ = io.EOF
	_ = fmt.Errorf
	_ = json.NewEncoder
)

type (
	// Internal API interface
	RuleSearch struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// ModuleID GET parameter
		//
		// Filter by module
		ModuleID uint64 `json:",string"`

		// Query GET parameter
		//
		// Filter by field name
		Query string

		// Enabled GET parameter
		//
		// Filter by enabled state
		Enabled bool

		// Limit GET parameter
		//
		// Limit
		Limit uint

		// PageCursor GET parameter
		//
		// Page cursor
		PageCursor string

		// Sort GET parameter
		//
		// Sort items
		Sort string
	}

	RuleCreate struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// ModuleID POST parameter
		//
		// Module ID
		ModuleID uint64 `json:",string"`

		// Field POST parameter
		//
		// Module field to watch
		Field string

		// Detector POST parameter
		//
		// Detector to use (e.g. zscore)
		Detector string

		// Threshold POST parameter
		//
		// Detector threshold
		Threshold float64

		// Enabled POST parameter
		//
		// Whether the rule is active
		Enabled bool

		// Params POST parameter
		//
		// Detector-specific parameters as a JSON object (e.g. {"min":0,"max":100} for the range detector)
		Params string
	}

	RuleUpdate struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// RuleID PATH parameter
		//
		// Rule ID
		RuleID uint64 `json:",string"`

		// Detector POST parameter
		//
		// Detector to use (e.g. zscore)
		Detector string

		// Threshold POST parameter
		//
		// Detector threshold
		Threshold float64

		// Enabled POST parameter
		//
		// Whether the rule is active
		Enabled bool

		// Params POST parameter
		//
		// Detector-specific parameters as a JSON object (e.g. {"min":0,"max":100} for the range detector)
		Params string
	}

	RuleDelete struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// RuleID PATH parameter
		//
		// Rule ID
		RuleID uint64 `json:",string"`
	}
)

// NewRuleSearch request
func NewRuleSearch() *RuleSearch {
	return &RuleSearch{}
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"moduleID":    r.ModuleID,
		"query":       r.Query,
		"enabled":     r.Enabled,
		"limit":       r.Limit,
		"pageCursor":  r.PageCursor,
		"sort":        r.Sort,
	}
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetModuleID() uint64 {
	return r.ModuleID
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetQuery() string {
	return r.Query
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetEnabled() bool {
	return r.Enabled
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetLimit() uint {
	return r.Limit
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetPageCursor() string {
	return r.PageCursor
}

// Auditable returns all auditable/loggable parameters
func (r RuleSearch) GetSort() string {
	return r.Sort
}

// Fill processes request and fills internal variables
func (r *RuleSearch) Fill(req *http.Request) (err error) {

	{
		// GET params
		tmp := req.URL.Query()

		if val, ok := tmp["moduleID"]; ok && len(val) > 0 {
			r.ModuleID, err = payload.ParseUint64(val[0]), nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["query"]; ok && len(val) > 0 {
			r.Query, err = val[0], nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["enabled"]; ok && len(val) > 0 {
			r.Enabled, err = payload.ParseBool(val[0]), nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["limit"]; ok && len(val) > 0 {
			r.Limit, err = payload.ParseUint(val[0]), nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["pageCursor"]; ok && len(val) > 0 {
			r.PageCursor, err = val[0], nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["sort"]; ok && len(val) > 0 {
			r.Sort, err = val[0], nil
			if err != nil {
				return err
			}
		}
	}

	{
		var val string
		// path params

		val = chi.URLParam(req, "namespaceID")
		r.NamespaceID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

	}

	return err
}

// NewRuleCreate request
func NewRuleCreate() *RuleCreate {
	return &RuleCreate{}
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"moduleID":    r.ModuleID,
		"field":       r.Field,
		"detector":    r.Detector,
		"threshold":   r.Threshold,
		"enabled":     r.Enabled,
		"params":      r.Params,
	}
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetModuleID() uint64 {
	return r.ModuleID
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetField() string {
	return r.Field
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetDetector() string {
	return r.Detector
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetThreshold() float64 {
	return r.Threshold
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetEnabled() bool {
	return r.Enabled
}

// Auditable returns all auditable/loggable parameters
func (r RuleCreate) GetParams() string {
	return r.Params
}

// Fill processes request and fills internal variables
func (r *RuleCreate) Fill(req *http.Request) (err error) {

	if strings.HasPrefix(strings.ToLower(req.Header.Get("content-type")), "application/json") {
		err = json.NewDecoder(req.Body).Decode(r)

		switch {
		case err == io.EOF:
			err = nil
		case err != nil:
			return fmt.Errorf("error parsing http request body: %w", err)
		}
	}

	{
		// Caching 32MB to memory, the rest to disk
		if err = req.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
			return err
		} else if err == nil {
			// Multipart params

			if val, ok := req.MultipartForm.Value["moduleID"]; ok && len(val) > 0 {
				r.ModuleID, err = payload.ParseUint64(val[0]), nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["field"]; ok && len(val) > 0 {
				r.Field, err = val[0], nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["detector"]; ok && len(val) > 0 {
				r.Detector, err = val[0], nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["threshold"]; ok && len(val) > 0 {
				r.Threshold, err = payload.ParseFloat64(val[0]), nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["enabled"]; ok && len(val) > 0 {
				r.Enabled, err = payload.ParseBool(val[0]), nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["params"]; ok && len(val) > 0 {
				r.Params, err = val[0], nil
				if err != nil {
					return err
				}
			}
		}
	}

	{
		if err = req.ParseForm(); err != nil {
			return err
		}

		// POST params

		if val, ok := req.Form["moduleID"]; ok && len(val) > 0 {
			r.ModuleID, err = payload.ParseUint64(val[0]), nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["field"]; ok && len(val) > 0 {
			r.Field, err = val[0], nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["detector"]; ok && len(val) > 0 {
			r.Detector, err = val[0], nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["threshold"]; ok && len(val) > 0 {
			r.Threshold, err = payload.ParseFloat64(val[0]), nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["enabled"]; ok && len(val) > 0 {
			r.Enabled, err = payload.ParseBool(val[0]), nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["params"]; ok && len(val) > 0 {
			r.Params, err = val[0], nil
			if err != nil {
				return err
			}
		}
	}

	{
		var val string
		// path params

		val = chi.URLParam(req, "namespaceID")
		r.NamespaceID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

	}

	return err
}

// NewRuleUpdate request
func NewRuleUpdate() *RuleUpdate {
	return &RuleUpdate{}
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"ruleID":      r.RuleID,
		"detector":    r.Detector,
		"threshold":   r.Threshold,
		"enabled":     r.Enabled,
		"params":      r.Params,
	}
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetRuleID() uint64 {
	return r.RuleID
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetDetector() string {
	return r.Detector
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetThreshold() float64 {
	return r.Threshold
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetEnabled() bool {
	return r.Enabled
}

// Auditable returns all auditable/loggable parameters
func (r RuleUpdate) GetParams() string {
	return r.Params
}

// Fill processes request and fills internal variables
func (r *RuleUpdate) Fill(req *http.Request) (err error) {

	if strings.HasPrefix(strings.ToLower(req.Header.Get("content-type")), "application/json") {
		err = json.NewDecoder(req.Body).Decode(r)

		switch {
		case err == io.EOF:
			err = nil
		case err != nil:
			return fmt.Errorf("error parsing http request body: %w", err)
		}
	}

	{
		// Caching 32MB to memory, the rest to disk
		if err = req.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
			return err
		} else if err == nil {
			// Multipart params

			if val, ok := req.MultipartForm.Value["detector"]; ok && len(val) > 0 {
				r.Detector, err = val[0], nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["threshold"]; ok && len(val) > 0 {
				r.Threshold, err = payload.ParseFloat64(val[0]), nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["enabled"]; ok && len(val) > 0 {
				r.Enabled, err = payload.ParseBool(val[0]), nil
				if err != nil {
					return err
				}
			}

			if val, ok := req.MultipartForm.Value["params"]; ok && len(val) > 0 {
				r.Params, err = val[0], nil
				if err != nil {
					return err
				}
			}
		}
	}

	{
		if err = req.ParseForm(); err != nil {
			return err
		}

		// POST params

		if val, ok := req.Form["detector"]; ok && len(val) > 0 {
			r.Detector, err = val[0], nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["threshold"]; ok && len(val) > 0 {
			r.Threshold, err = payload.ParseFloat64(val[0]), nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["enabled"]; ok && len(val) > 0 {
			r.Enabled, err = payload.ParseBool(val[0]), nil
			if err != nil {
				return err
			}
		}

		if val, ok := req.Form["params"]; ok && len(val) > 0 {
			r.Params, err = val[0], nil
			if err != nil {
				return err
			}
		}
	}

	{
		var val string
		// path params

		val = chi.URLParam(req, "namespaceID")
		r.NamespaceID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

		val = chi.URLParam(req, "ruleID")
		r.RuleID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

	}

	return err
}

// NewRuleDelete request
func NewRuleDelete() *RuleDelete {
	return &RuleDelete{}
}

// Auditable returns all auditable/loggable parameters
func (r RuleDelete) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"ruleID":      r.RuleID,
	}
}

// Auditable returns all auditable/loggable parameters
func (r RuleDelete) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r RuleDelete) GetRuleID() uint64 {
	return r.RuleID
}

// Fill processes request and fills internal variables
func (r *RuleDelete) Fill(req *http.Request) (err error) {

	{
		var val string
		// path params

		val = chi.URLParam(req, "namespaceID")
		r.NamespaceID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

		val = chi.URLParam(req, "ruleID")
		r.RuleID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

	}

	return err
}
