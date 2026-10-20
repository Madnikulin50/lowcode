package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// ИИ-разбор неоднозначного замечания экспертизы: модель получает замечание, ответ проектировщика и
// фрагменты связанных томов — и говорит, видно ли по ним, что замечание устранено. Только по
// предъявленным фрагментам; каждая цитата проверяется по тексту; рабочий статус замечания ИИ не
// меняет — результат лежит в отдельных полях и всегда «к проверке» специалистом.

// AI verdicts.
const (
	AIFixed    = "fixed"
	AIPartial  = "partial"
	AINotFixed = "not_fixed"
	AIUnclear  = "unclear"
)

// Passage is a fragment of a volume offered to the model.
type Passage struct {
	Ref  string `json:"ref"` // Ф1, Ф2…
	Doc  string `json:"doc"`
	Page int    `json:"page"`
	Text string `json:"text"`
}

// PassageDoc is a document whose pages are searched.
type PassageDoc struct {
	Name  string
	Pages []Page
}

// RetrievePassages picks the fragments of the documents that share the most (words, and more
// strongly codes and values) with the query: at most one per page, best first.
func RetrievePassages(query string, docs []PassageDoc, limit int) []Passage {
	qt := remarkTokens(query)
	wq := total(qt)
	if wq == 0 {
		return nil
	}
	type cand struct {
		p     Passage
		score float64
	}
	var all []cand
	for _, d := range docs {
		for _, pg := range d.Pages {
			if pg.Drawing || !pg.HasText {
				continue
			}
			var lines []string
			for _, l := range pg.Lines {
				if !l.Stamp && !skipLine(l.Key) {
					lines = append(lines, l.Text)
				}
			}
			best, bi, bj := 0.0, -1, 0
			const win = 4
			for i := 0; i < len(lines); i++ {
				j := i + win
				if j > len(lines) {
					j = len(lines)
				}
				wt := remarkTokens(strings.Join(lines[i:j], " "))
				s, common := shared(qt, wt)
				if len(common) < 2 || total(wt) == 0 {
					continue
				}
				if sc := s / math.Sqrt(wq*total(wt)); sc > best {
					best, bi, bj = sc, i, j
				}
			}
			if bi < 0 {
				continue
			}
			text := strings.Join(lines[bi:bj], "\n")
			all = append(all, cand{Passage{Doc: d.Name, Page: pg.Number, Text: truncate(text, 600)}, best})
		}
	}
	if len(all) == 0 {
		return nil
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].score > all[j].score })
	if limit <= 0 || limit > len(all) {
		limit = len(all)
	}
	out := make([]Passage, 0, limit)
	for i := 0; i < limit; i++ {
		p := all[i].p
		p.Ref = fmt.Sprintf("Ф%d", i+1)
		out = append(out, p)
	}
	return out
}

// RemarkAIInput is what the model is shown.
type RemarkAIInput struct {
	Number       string
	Text         string
	Basis        string
	Answer       string
	RejectReason string
	PrevAnswer   string
	Findings     []string // differences found by the comparison that match the remark
	Passages     []Passage
}

// AICite is a quotation the verdict rests on.
type AICite struct {
	Ref   string `json:"ref"`
	Doc   string `json:"doc"`
	Page  int    `json:"page"`
	Quote string `json:"quote"`
}

// AIVerdict is the model's reading of one remark.
type AIVerdict struct {
	Verdict     string   `json:"verdict"`
	Explanation string   `json:"explanation"`
	Cited       []AICite `json:"cited"`
	Dropped     int      `json:"-"` // quotations that are not in the fragments and were thrown away
}

const interpretSystem = `Ты помощник эксперта строительной экспертизы. Тебе дано замечание, ответ проектировщика и фрагменты текста связанных томов документации (каждый с кодом Ф1, Ф2…).
Определи, видно ли по фрагментам, что замечание устранено. Используй ТОЛЬКО эти фрагменты и приведённые расхождения: не оценивай качество проектных решений, не предлагай улучшений, не выдумывай значений и номеров листов.
Верни JSON: {"verdict": "fixed|partial|not_fixed|unclear", "explanation": "...", "cited": [{"ref": "Ф1", "quote": "точная цитата из фрагмента, до 200 символов"}]}.
fixed — по фрагментам видно, что устранено так, как сказано в ответе; partial — устранено частично; not_fixed — фрагменты показывают, что не устранено или противоречат ответу; unclear — по фрагментам судить нельзя.
explanation — не более 4 предложений по-русски. В cited — только дословные цитаты; если цитировать нечего, оставь пустой массив и поставь unclear.`

func interpretPrompt(in RemarkAIInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Замечание №%s:\n%s\n", in.Number, in.Text)
	if in.Basis != "" {
		fmt.Fprintf(&b, "Основание: %s\n", in.Basis)
	}
	if in.Answer != "" {
		fmt.Fprintf(&b, "\nОтвет проектировщика:\n%s\n", in.Answer)
	}
	if in.RejectReason != "" {
		fmt.Fprintf(&b, "Причина отклонения предыдущего ответа: %s\n", in.RejectReason)
	}
	if in.PrevAnswer != "" {
		fmt.Fprintf(&b, "Предыдущий ответ: %s\n", in.PrevAnswer)
	}
	if len(in.Findings) > 0 {
		b.WriteString("\nРасхождения, найденные сравнением документации:\n")
		for _, f := range in.Findings {
			b.WriteString("— " + f + "\n")
		}
	}
	b.WriteString("\nФрагменты томов:\n")
	for _, p := range in.Passages {
		fmt.Fprintf(&b, "[%s] %s, стр. %d:\n%s\n\n", p.Ref, p.Doc, p.Page, p.Text)
	}
	return b.String()
}

// parseAIVerdict reads the model's answer and checks it against the fragments it was given.
func parseAIVerdict(raw string, passages []Passage) (*AIVerdict, error) {
	raw = strings.TrimSpace(raw)
	if i, j := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var v AIVerdict
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, fmt.Errorf("ответ модели не разобран: %w", err)
	}
	switch v.Verdict {
	case AIFixed, AIPartial, AINotFixed, AIUnclear:
	default:
		v.Verdict = AIUnclear
	}
	byRef := map[string]Passage{}
	for _, p := range passages {
		byRef[p.Ref] = p
	}
	var kept []AICite
	for _, c := range v.Cited {
		p, ok := byRef[strings.TrimSpace(c.Ref)]
		q := NormKey(c.Quote)
		if !ok || len([]rune(q)) < 8 || !strings.Contains(NormKey(strings.ReplaceAll(p.Text, "\n", " ")), q) {
			v.Dropped++ // a quotation that is not in the fragment is not evidence
			continue
		}
		kept = append(kept, AICite{Ref: p.Ref, Doc: p.Doc, Page: p.Page, Quote: strings.TrimSpace(c.Quote)})
	}
	v.Cited = kept
	v.Explanation = strings.TrimSpace(v.Explanation)
	if (v.Verdict == AIFixed || v.Verdict == AIPartial || v.Verdict == AINotFixed) && len(v.Cited) == 0 {
		v.Explanation = "Модель не привела проверяемых цитат — вывод не принят. " + v.Explanation
		v.Verdict = AIUnclear
	}
	return &v, nil
}

// Interpret asks the model whether the fragments show the remark fixed.
func (s *Suggester) Interpret(ctx context.Context, in RemarkAIInput) (*AIVerdict, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("модель не настроена: задайте REMARK_MODEL или --remark-model")
	}
	if len(in.Passages) == 0 {
		return nil, fmt.Errorf("нет фрагментов томов, по которым можно судить о замечании")
	}
	raw, err := s.chat(ctx, interpretSystem, interpretPrompt(in), true, 0.1)
	if err != nil {
		return nil, err
	}
	return parseAIVerdict(raw, in.Passages)
}

// NoteText is the text kept in the remark: the explanation and the quotations.
func (v *AIVerdict) NoteText() string {
	var b strings.Builder
	b.WriteString(v.Explanation)
	for _, c := range v.Cited {
		fmt.Fprintf(&b, "\n— %s, стр. %d: «%s»", c.Doc, c.Page, c.Quote)
	}
	if v.Dropped > 0 {
		fmt.Fprintf(&b, "\n(отброшено цитат, которых нет в тексте: %d)", v.Dropped)
	}
	return strings.TrimSpace(b.String())
}

// AnalyzeRemarkAI reads the volumes the remark points at, lets the model judge and writes the result
// into ai_verdict / ai_note / ai_at of the remark. The working status of the remark is not touched.
func (r *Runner) AnalyzeRemarkAI(ctx context.Context, ns, remarkID string) (*AIVerdict, error) {
	if r.LLM == nil || !r.LLM.Enabled() {
		return nil, fmt.Errorf("модель не настроена: задайте REMARK_MODEL или --remark-model")
	}
	b := r.Backend
	rem, err := b.GetRecord(ns, ModExpertise, remarkID)
	if err != nil {
		return nil, err
	}
	var docs []PassageDoc
	for _, id := range rem.Values["linked_documents"] {
		if id == "" {
			continue
		}
		d, err := b.GetRecord(ns, ModDocuments, id)
		if err != nil || d.Get("file") == "" {
			continue
		}
		data, name, err := b.Download(ns, d.Get("file"))
		if err != nil || !IsPDF(data) {
			continue
		}
		pages, _, _, err := ExtractPagesOCR(ctx, id, data, r.CacheDir, r.OCR)
		if err != nil {
			continue
		}
		label := name
		if c := d.Get("volume_code"); c != "" {
			label += " [" + c + "]"
		}
		docs = append(docs, PassageDoc{Name: label, Pages: pages})
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("к замечанию не привязано ни одного загруженного тома — сначала «Проверить по документам»")
	}
	in := RemarkAIInput{
		Number: rem.Get("number"), Text: rem.Get("text"), Basis: rem.Get("basis"), Answer: rem.Get("answer"),
		RejectReason: rem.Get("reject_reason"), PrevAnswer: rem.Get("prev_answer"),
	}
	for _, fid := range rem.Values["matched_discrepancies"] {
		if fid == "" {
			continue
		}
		if f, err := b.GetRecord(ns, ModDiscrepancies, fid); err == nil {
			in.Findings = append(in.Findings, cleanDescription(f.Get("description")))
		}
	}
	in.Passages = RetrievePassages(in.Text+" "+in.Answer+" "+rem.Get("section"), docs, 6)
	v, err := r.LLM.Interpret(ctx, in)
	if err != nil {
		return nil, err
	}
	upd := idcheck.V("ai_verdict", v.Verdict, "ai_note", v.NoteText(), "ai_at", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	if err := b.Update(ns, ModExpertise, remarkID, upd); err != nil {
		return nil, err
	}
	return v, nil
}

// AIBatch runs AnalyzeRemarkAI for several remarks one after another, reporting progress.
func (r *Runner) AIBatch(ctx context.Context, ns, key string, ids []string) (done, failed int, firstErr error) {
	for i, id := range ids {
		r.setStatus(key, strconv.Itoa(i+1)+" из "+strconv.Itoa(len(ids)))
		if _, err := r.AnalyzeRemarkAI(ctx, ns, id); err != nil {
			failed++
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		done++
	}
	r.setStatus(key, "")
	return
}
