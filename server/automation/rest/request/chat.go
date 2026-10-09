package request

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

type (
	ChatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	ChatAsk struct {
		ChatID     string        `json:",string"`
		WorkflowID uint64        `json:",string"`
		TriggerID  uint64        `json:",string"`
		SessionID  uint64        `json:",string"`
		Facts      []string      `json:"facts"`
		Prompt     string        `json:"prompt"`
		Messages   []ChatMessage `json:"messages"`
	}
)

func NewChatAsk() *ChatAsk {
	return &ChatAsk{}
}

func (r *ChatAsk) Fill(req *http.Request) (err error) {
	defer req.Body.Close()

	r.WorkflowID = payload.ParseUint64(chi.URLParam(req, "workflowID"))

	if err = json.NewDecoder(req.Body).Decode(r); err != nil {
		return err
	}

	r.Prompt = strings.TrimSpace(r.Prompt)
	return
}

func (r *ChatAsk) Auditable() map[string]interface{} {
	return map[string]interface{}{
		"prompt": r.Prompt,
	}
}
