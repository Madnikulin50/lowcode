package cmp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Remark is a remark card (карточка замечания) drafted from discrepancies.
type Remark struct {
	ID             string   `json:"id"`
	DiscrepancyIDs []string `json:"discrepancyIDs"`
	DocumentRef    string   `json:"documentRef"` // which documents/sheets
	Sheet          string   `json:"sheet"`
	WhatDiverges   string   `json:"whatDiverges"`
	MustAlign      string   `json:"mustAlign"`
	DraftText      string   `json:"draftText"` // LLM draft, only on request
	FinalText      string   `json:"finalText"` // what the specialist approved
	Status         string   `json:"status"`    // draft | issued | closed
}

// Fact is one discrepancy as the card builder needs it.
type Fact struct {
	Description string
	DocRef      string // display name of the ПД file
	DocTarget   string
	PageRef     string
	PageTarget  string
	Kind        string
	Evidence    Evidence
}

const alignTemplate = "Привести рабочую документацию в соответствие с проектной документацией (эталон — стадия П) либо представить обоснование изменения."

// BuildRemark fills a card from discrepancies deterministically. The
// specialist edits the text; nothing here involves an LLM.
func BuildRemark(facts []Fact) Remark {
	var docs, sheets, what []string
	seenDoc, seenSheet := map[string]bool{}, map[string]bool{}
	for _, f := range facts {
		if f.DocRef != "" && !seenDoc["ПД: "+f.DocRef] {
			seenDoc["ПД: "+f.DocRef] = true
			docs = append(docs, "ПД: "+f.DocRef)
		}
		if f.DocTarget != "" && !seenDoc["РД: "+f.DocTarget] {
			seenDoc["РД: "+f.DocTarget] = true
			docs = append(docs, "РД: "+f.DocTarget)
		}
		var parts []string
		if f.PageRef != "" {
			parts = append(parts, "ПД стр. "+f.PageRef)
		}
		if f.PageTarget != "" {
			parts = append(parts, "РД стр. "+f.PageTarget)
		}
		if s := strings.Join(parts, ", "); s != "" && !seenSheet[s] {
			seenSheet[s] = true
			sheets = append(sheets, s)
		}
		what = append(what, "— "+strings.TrimPrefix(strings.TrimPrefix(f.Description, "[детерм.] "), "[ИИ] "))
	}
	return Remark{
		DocumentRef:  strings.Join(docs, "; "),
		Sheet:        strings.Join(sheets, "; "),
		WhatDiverges: strings.Join(what, "\n"),
		MustAlign:    alignTemplate,
		Status:       "draft",
	}
}

// Suggester asks a local Ollama model to phrase a remark. It is only called
// on an explicit user action; the answer is a draft.
type Suggester struct {
	URL   string
	Model string
	http  *http.Client
}

func NewSuggester(url, model string) *Suggester {
	return &Suggester{URL: strings.TrimRight(url, "/"), Model: model, http: &http.Client{Timeout: 10 * time.Minute}}
}

// Enabled reports whether a model is configured.
func (s *Suggester) Enabled() bool { return s != nil && s.Model != "" }

const suggestSystem = `Ты помощник специалиста производственно-технического отдела. Сформулируй текст замечания по результатам сравнения проектной (ПД, эталон) и рабочей (РД) документации.
Правила: используй ТОЛЬКО факты из данных; не оценивай качество проектных решений и не предлагай улучшений; не выдумывай номера листов и значения; пиши по-русски, официально и кратко (до 6 предложений); укажи, в каком документе и на каком листе расхождение, в чём оно состоит и что требуется (привести в соответствие или пояснить).`

// Suggest returns a draft wording for the card.
func (s *Suggester) Suggest(ctx context.Context, r Remark) (string, error) {
	if !s.Enabled() {
		return "", fmt.Errorf("модель для формулировок не настроена")
	}
	user := fmt.Sprintf("Документы: %s\nЛисты: %s\nРасхождения:\n%s\nТребуется: %s", r.DocumentRef, r.Sheet, r.WhatDiverges, r.MustAlign)
	body, _ := json.Marshal(map[string]any{
		"model":   s.Model,
		"stream":  false,
		"think":   false,
		"options": map[string]any{"temperature": 0.2, "num_ctx": 8192},
		"messages": []map[string]string{
			{"role": "system", "content": suggestSystem},
			{"role": "user", "content": user},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	text := strings.TrimSpace(out.Message.Content)
	if text == "" {
		return "", fmt.Errorf("модель вернула пустой ответ")
	}
	return text, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
