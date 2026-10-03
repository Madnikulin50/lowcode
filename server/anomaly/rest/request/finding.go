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
	FindingSearch struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// ModuleID GET parameter
		//
		// Filter by module
		ModuleID uint64 `json:",string"`

		// RecordID GET parameter
		//
		// Filter by record
		RecordID uint64 `json:",string"`

		// Status GET parameter
		//
		// Filter by status
		Status string

		// Severity GET parameter
		//
		// Filter by severity
		Severity string

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

	FindingUpdateStatus struct {
		// NamespaceID PATH parameter
		//
		// Namespace ID
		NamespaceID uint64 `json:",string"`

		// FindingID PATH parameter
		//
		// Finding ID
		FindingID uint64 `json:",string"`

		// Status POST parameter
		//
		// New status (acknowledged, resolved, false_positive)
		Status string
	}
)

// NewFindingSearch request
func NewFindingSearch() *FindingSearch {
	return &FindingSearch{}
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"moduleID":    r.ModuleID,
		"recordID":    r.RecordID,
		"status":      r.Status,
		"severity":    r.Severity,
		"limit":       r.Limit,
		"pageCursor":  r.PageCursor,
		"sort":        r.Sort,
	}
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetModuleID() uint64 {
	return r.ModuleID
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetRecordID() uint64 {
	return r.RecordID
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetStatus() string {
	return r.Status
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetSeverity() string {
	return r.Severity
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetLimit() uint {
	return r.Limit
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetPageCursor() string {
	return r.PageCursor
}

// Auditable returns all auditable/loggable parameters
func (r FindingSearch) GetSort() string {
	return r.Sort
}

// Fill processes request and fills internal variables
func (r *FindingSearch) Fill(req *http.Request) (err error) {

	{
		// GET params
		tmp := req.URL.Query()

		if val, ok := tmp["moduleID"]; ok && len(val) > 0 {
			r.ModuleID, err = payload.ParseUint64(val[0]), nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["recordID"]; ok && len(val) > 0 {
			r.RecordID, err = payload.ParseUint64(val[0]), nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["status"]; ok && len(val) > 0 {
			r.Status, err = val[0], nil
			if err != nil {
				return err
			}
		}
		if val, ok := tmp["severity"]; ok && len(val) > 0 {
			r.Severity, err = val[0], nil
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

// NewFindingUpdateStatus request
func NewFindingUpdateStatus() *FindingUpdateStatus {
	return &FindingUpdateStatus{}
}

// Auditable returns all auditable/loggable parameters
func (r FindingUpdateStatus) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"findingID":   r.FindingID,
		"status":      r.Status,
	}
}

// Auditable returns all auditable/loggable parameters
func (r FindingUpdateStatus) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r FindingUpdateStatus) GetFindingID() uint64 {
	return r.FindingID
}

// Auditable returns all auditable/loggable parameters
func (r FindingUpdateStatus) GetStatus() string {
	return r.Status
}

// Fill processes request and fills internal variables
func (r *FindingUpdateStatus) Fill(req *http.Request) (err error) {

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

			if val, ok := req.MultipartForm.Value["status"]; ok && len(val) > 0 {
				r.Status, err = val[0], nil
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

		if val, ok := req.Form["status"]; ok && len(val) > 0 {
			r.Status, err = val[0], nil
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

		val = chi.URLParam(req, "findingID")
		r.FindingID, err = payload.ParseUint64(val), nil
		if err != nil {
			return err
		}

	}

	return err
}
