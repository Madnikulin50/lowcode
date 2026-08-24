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
	ChatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	ChatFile struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// Internal API interface
	ChatAsk struct {
		ChatID      string        `json:",string"`
		NamespaceID uint64        `json:",string"`
		PageID      uint64        `json:",string"`
		ModuleID    uint64        `json:",string"`
		RecordID    uint64        `json:",string"`
		Facts       []string      `json:"facts"`
		Files       []ChatFile    `json:"files"`
		Prompt      string        `json:"prompt"`
		Messages    []ChatMessage `json:"messages"`
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

			// extract messages array before flattening to string map
			if raw, ok := d["messages"]; ok {
				if arr, ok2 := raw.([]interface{}); ok2 {
					for _, item := range arr {
						if m, ok3 := item.(map[string]interface{}); ok3 {
							role, _ := m["role"].(string)
							content, _ := m["content"].(string)
							r.Messages = append(r.Messages, ChatMessage{Role: role, Content: content})
						}
					}
				}
			}

			// extract files
			if raw, ok := d["files"]; ok {
				if arr, ok2 := raw.([]interface{}); ok2 {
					for _, item := range arr {
						if f, ok3 := item.(map[string]interface{}); ok3 {
							name, _ := f["name"].(string)
							content, _ := f["content"].(string)
							r.Files = append(r.Files, ChatFile{Name: name, Content: content})
						}
					}
				}
			}

			for k, v := range d {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
		if val, ok := data["moduleID"]; ok && len(val) > 0 {
			r.ModuleID = payload.ParseUint64(val)
		}

		if val, ok := data["chatID"]; ok && len(val) > 0 {
			r.ChatID = val
		}

		if val, ok := data["moduleID"]; ok && len(val) > 0 {
			r.ModuleID = payload.ParseUint64(val)
		}
		if val, ok := data["pageID"]; ok && len(val) > 0 {
			r.PageID = payload.ParseUint64(val)
		}

		if val, ok := data["prompt"]; ok && len(val) > 0 {
			r.Prompt = val
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
