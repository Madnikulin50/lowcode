package request

import (
	"fmt"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

type (
	// Internal API interface
	ChatAsk struct {
		ChatID      uint64 `json:",string"`
		NamespaceID uint64 `json:",string"`
		PageID      uint64 `json:",string"`
		ModuleID    uint64 `json:",string"`
		RecordID    uint64 `json:",string"`
		Facts       []string
		Prompt      string
	}
)

// NewPageList request
func NewChatAsk() *ChatAsk {
	return &ChatAsk{}
}

// Auditable returns all auditable/loggable parameters
func (r ChatAsk) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"namespaceID": r.NamespaceID,
		"moduleID":    r.ModuleID,
		"pageID":      r.PageID,
		"recordID":    r.RecordID,
		"prompt":      r.Prompt,
	}
}

// Auditable returns all auditable/loggable parameters
func (r ChatAsk) GetNamespaceID() uint64 {
	return r.NamespaceID
}

// Auditable returns all auditable/loggable parameters
func (r ChatAsk) GetModuleID() uint64 {
	return r.ModuleID
}

// Auditable returns all auditable/loggable parameters
func (r ChatAsk) GetPrompt() string {
	return r.Prompt
}

// Fill processes request and fills internal variables
func (r *ChatAsk) Fill(req *http.Request) (err error) {

	{
		// GET params
		uTmp := req.URL.Query()
		data := map[string]string{}
		for k, v := range uTmp {
			data[k] = v[0]
		}

		if strings.HasPrefix(strings.ToLower(req.Header.Get("content-type")), "application/json") {
			var d map[string]interface{}
			err = json.NewDecoder(req.Body).Decode(&d)
			for k, v := range d {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
		if val, ok := data["moduleID"]; ok && len(val) > 0 {
			r.ModuleID, err = payload.ParseUint64(val), nil
			if err != nil {
				return err
			}
		}
		if val, ok := data["pageID"]; ok && len(val) > 0 {
			r.PageID, err = payload.ParseUint64(val), nil
			if err != nil {
				return err
			}
		}

		if val, ok := data["prompt"]; ok && len(val) > 0 {
			r.Prompt, err = val, nil
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
