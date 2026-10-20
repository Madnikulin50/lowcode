package cmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func dialogFacts() DialogFacts {
	return DialogFacts{
		Title: "КЖ1", SetsLabel: "ПД v1 ↔ РД v2", Similarity: "79.2", RefName: "ПД", TargetName: "РД",
		Selected: &DialogFinding{ID: "f1", Severity: "high", Kind: KindParam, Description: "Класс бетона B25 → B30", RefDoc: "dp", TargetDoc: "dr", RefDocName: "pd.pdf", TgtDocName: "rd.pdf", PageRef: 3, PageTarget: 4},
		Related:  []DialogFinding{{ID: "f2", Severity: "low", Kind: KindText, Description: "Изменена формулировка", RefDoc: "dp", TargetDoc: "dr", RefDocName: "pd.pdf", TgtDocName: "rd.pdf", PageRef: 5, PageTarget: 6}},
		Excerpts: []DialogExcerpt{{Side: SideRef, Doc: "dp", Name: "pd.pdf", Page: 3, Lines: []string{"Фундаменты из бетона класса В25"}}},
		Remarks:  []string{"№7.1: принять класс бетона В30"},
	}
}

func TestDialogSystemHasRulesAndFacts(t *testing.T) {
	s := DialogSystem(dialogFacts())
	for _, want := range []string{"ТОЛЬКО на факты", "не оценивай качество проектных решений", "[[ПД:12]]", "Версии комплектов: ПД v1 ↔ РД v2",
		"Класс бетона B25 → B30", "Соседние находки", "Фундаменты из бетона класса В25", "№7.1"} {
		if !strings.Contains(s, want) {
			t.Fatalf("system prompt lacks %q:\n%s", want, s)
		}
	}
}

func TestExtractJumpsKeepsOnlySheetsFromTheFacts(t *testing.T) {
	text, jumps := ExtractJumps("Смотрите [[ПД:3]] и [[РД: 4]], а также [[РД:99]] и [[ПД:3]] ещё раз; [[ЧТО:1]].", dialogFacts())
	if len(jumps) != 2 || jumps[0].Side != SideRef || jumps[0].Doc != "dp" || jumps[0].Page != 3 || jumps[1].Side != SideTarget || jumps[1].Doc != "dr" || jumps[1].Page != 4 {
		t.Fatalf("jumps: %+v", jumps)
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "99") || !strings.Contains(text, "ПД стр. 3") || !strings.Contains(text, "РД стр. 4") {
		t.Fatalf("text: %q", text)
	}
	// in a versions comparison the sides are named До/После
	f := dialogFacts()
	f.RefName, f.TargetName = "До", "После"
	_, jumps = ExtractJumps("См. [[До:3]] и [[После:6]]", f)
	if len(jumps) != 2 || jumps[1].Page != 6 || jumps[0].Name != "До стр. 3" {
		t.Fatalf("%+v", jumps)
	}
}

func TestConverseSendsFactsHistoryAndFiltersAnswer(t *testing.T) {
	var got struct {
		Messages []Msg `json:"messages"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = json.NewDecoder(req.Body).Decode(&got)
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": "Класс бетона изменён, см. [[РД:4]]; ещё [[РД:50]]."}})
		_, _ = w.Write(out)
	}))
	defer srv.Close()
	s := NewSuggester(srv.URL, "m")
	var hist []Msg
	for i := 0; i < 14; i++ {
		hist = append(hist, Msg{Role: "user", Content: "вопрос"}, Msg{Role: "assistant", Content: "ответ"})
	}
	hist = append(hist, Msg{Role: "user", Content: "Объясни различие"})
	text, jumps, err := s.Converse(context.Background(), dialogFacts(), hist)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Messages) == 0 || got.Messages[0].Role != "system" || !strings.Contains(got.Messages[0].Content, "Класс бетона B25 → B30") {
		t.Fatalf("the first message must be the system prompt with the facts: %+v", got.Messages[0])
	}
	if last := got.Messages[len(got.Messages)-1]; last.Role != "user" || last.Content != "Объясни различие" {
		t.Fatalf("last: %+v", last)
	}
	if len(got.Messages) > maxDialogMessages+1 || got.Messages[1].Role != "user" {
		t.Fatalf("history must be cut to %d and start with a user message, got %d", maxDialogMessages, len(got.Messages))
	}
	if len(jumps) != 1 || jumps[0].Page != 4 || strings.Contains(text, "50") {
		t.Fatalf("text %q jumps %+v", text, jumps)
	}
	if _, _, err := s.Converse(context.Background(), dialogFacts(), []Msg{{Role: "assistant", Content: "x"}}); err == nil {
		t.Fatal("a conversation without a user message must fail")
	}
	if _, _, err := NewSuggester(srv.URL, "").Converse(context.Background(), dialogFacts(), hist); err == nil {
		t.Fatal("no model configured must fail")
	}
}
