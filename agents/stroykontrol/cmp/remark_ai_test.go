package cmp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func aiPages() []PassageDoc {
	d := textDoc("d", SideTarget,
		[]string{"Общие данные", "Фундаменты выполнить из бетона класса В30, W6, F150", "Глубина заложения фундамента 2,4 м"},
		[]string{"Кабельные линии прокладывать в траншее", "Глубина траншеи 0,7 м"},
	)
	return []PassageDoc{{Name: "rd.pdf", Pages: d.Pages}}
}

func TestRetrievePassagesFindsTheRelevantPage(t *testing.T) {
	ps := RetrievePassages("Глубина заложения фундамента принять 2,4 м, класс бетона В30", aiPages(), 3)
	if len(ps) == 0 || ps[0].Page != 1 || ps[0].Ref != "Ф1" || !strings.Contains(ps[0].Text, "2,4") {
		t.Fatalf("%+v", ps)
	}
	if RetrievePassages("совсем другое про вентиляцию", aiPages(), 3) != nil {
		t.Fatal("nothing shared → no passages")
	}
}

func TestParseAIVerdictDropsInventedQuotes(t *testing.T) {
	ps := []Passage{{Ref: "Ф1", Doc: "rd.pdf", Page: 1, Text: "Глубина заложения фундамента 2,4 м\nКласс бетона В30"}}
	good := `{"verdict":"fixed","explanation":"Глубина приведена.","cited":[{"ref":"Ф1","quote":"Глубина заложения фундамента 2,4 м"},{"ref":"Ф1","quote":"Совсем выдуманная цитата из ниоткуда"},{"ref":"Ф9","quote":"Класс бетона В30"}]}`
	v, err := parseAIVerdict("Вот ответ: "+good, ps)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdict != AIFixed || len(v.Cited) != 1 || v.Cited[0].Page != 1 || v.Dropped != 2 {
		t.Fatalf("%+v", v)
	}
	if !strings.Contains(v.NoteText(), "отброшено цитат") {
		t.Fatalf("the note must say that quotes were thrown away: %s", v.NoteText())
	}
	// a firm verdict with no checkable quotation is not accepted
	v, _ = parseAIVerdict(`{"verdict":"not_fixed","explanation":"Не устранено.","cited":[{"ref":"Ф1","quote":"нет такого текста совсем"}]}`, ps)
	if v.Verdict != AIUnclear || !strings.Contains(v.Explanation, "не привела проверяемых цитат") {
		t.Fatalf("%+v", v)
	}
	v, _ = parseAIVerdict(`{"verdict":"maybe","explanation":"?","cited":[]}`, ps)
	if v.Verdict != AIUnclear {
		t.Fatalf("an unknown verdict is unclear: %+v", v)
	}
	if _, err := parseAIVerdict("не json", ps); err == nil {
		t.Fatal("garbage must fail")
	}
}

// fakeOllama answers with a verdict that quotes the first fragment it was shown.
func fakeOllama(t *testing.T, verdict string, seen *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Format   string                           `json:"format"`
			Messages []struct{ Role, Content string } `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		user := body.Messages[1].Content
		*seen = user
		if body.Format != "json" {
			http.Error(w, "json mode expected", 400)
			return
		}
		i := strings.Index(user, "[Ф1]")
		quote := ""
		if i >= 0 {
			rest := user[i:]
			lines := strings.Split(rest, "\n")
			if len(lines) > 1 {
				quote = lines[1]
			}
		}
		ans, _ := json.Marshal(map[string]any{"verdict": verdict, "explanation": "Фрагмент соответствует ответу.", "cited": []map[string]string{{"ref": "Ф1", "quote": quote}}})
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": string(ans)}})
		_, _ = w.Write(out)
	}))
}

func TestAnalyzeRemarkAIWritesOnlyTheAIFields(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	doc := f.put(ModDocuments, idcheck.V("file", "pack-1/rd-kzh.pdf", "side", SideTarget, "volume_code", "2024-15-КЖ1"))
	rem := f.put(ModExpertise, idcheck.V("object", obj, "number", "7.1", "text", "Глубина заложения фундамента принять 2,6 м, класс бетона В30",
		"answer", "Глубина заложения уточнена", "status", RemarkAnswerRev, "linked_documents", doc))
	var seen string
	srv := fakeOllama(t, AIFixed, &seen)
	defer srv.Close()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), LLM: NewSuggester(srv.URL, "m1")}

	v, err := r.AnalyzeRemarkAI(context.Background(), "ns", rem)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdict != AIFixed || len(v.Cited) != 1 {
		t.Fatalf("%+v", v)
	}
	if !strings.Contains(seen, "Ответ проектировщика") || !strings.Contains(seen, "Глубина заложения") {
		t.Fatalf("the model must be shown the remark, the answer and fragments:\n%s", seen)
	}
	got, _ := f.GetRecord("ns", ModExpertise, rem)
	if got.Get("ai_verdict") != AIFixed || got.Get("ai_note") == "" || got.Get("ai_at") == "" {
		t.Fatalf("ai fields: %+v", got.Values)
	}
	if got.Get("status") != RemarkAnswerRev {
		t.Fatalf("the working status must not be changed by the model, got %q", got.Get("status"))
	}
}

func TestAnalyzeRemarkAIFailsClearly(t *testing.T) {
	f := newFake()
	rem := f.put(ModExpertise, idcheck.V("number", "1", "text", "т"))
	if _, err := (&Runner{Backend: f}).AnalyzeRemarkAI(context.Background(), "ns", rem); err == nil || !strings.Contains(err.Error(), "модель не настроена") {
		t.Fatalf("%v", err)
	}
	r := &Runner{Backend: f, LLM: NewSuggester("http://127.0.0.1:1", "m")}
	if _, err := r.AnalyzeRemarkAI(context.Background(), "ns", rem); err == nil || !strings.Contains(err.Error(), "не привязано ни одного") {
		t.Fatalf("%v", err)
	}
}
