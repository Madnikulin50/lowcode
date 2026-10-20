package cmp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Диалоговый режим просмотра: специалист разговаривает с моделью о найденных расхождениях. Модель
// видит только факты сравнения (выбранное расхождение, соседние находки, текст листов вокруг него) и
// не оценивает проектные решения; ссылки на листы она ставит маркерами [[ПД:12]], а сервер оставляет
// только те листы, о которых ей рассказали.

// DialogFinding is a finding as the dialog shows it to the model.
type DialogFinding struct {
	ID          string
	Severity    string
	Kind        string
	Description string
	RefDoc      string
	TargetDoc   string
	RefDocName  string
	TgtDocName  string
	PageRef     int
	PageTarget  int
}

// DialogExcerpt is the text of a sheet around a finding.
type DialogExcerpt struct {
	Side  string
	Doc   string // document ID
	Name  string
	Page  int
	Lines []string
}

// DialogPage is a sheet of the document on screen, as the model sees it.
type DialogPage struct {
	Side  string // ref | target | "" (a document viewed on its own)
	Doc   string // document ID the viewer can jump to
	Name  string
	Title string // volume / heading, if known
	Page  int
	Count int // pages in the file
	Text  string
	Thin  bool // almost no text layer: a drawing or a scan
}

// DialogFacts is everything the model may rely on.
type DialogFacts struct {
	Pages     []DialogPage // the sheets on screen
	Skill     string       // method for reading drawings, added for drawing sheets
	ImageNote string       // what the attached picture shows / why there is none

	Title, SetsLabel, Similarity string
	RefName, TargetName          string
	Selected                     *DialogFinding
	Related                      []DialogFinding
	Excerpts                     []DialogExcerpt
	Remarks                      []string // short lines about expertise remarks that concern the comparison
}

// Jump is a sheet the user can be taken to.
type Jump struct {
	Side string `json:"side"` // ref | target
	Doc  string `json:"doc"`
	Page int    `json:"page"`
	Name string `json:"name"` // "ПД стр. 12"
}

func findingLine(f DialogFinding, ref, tgt string) string {
	var where []string
	if f.PageRef > 0 {
		where = append(where, fmt.Sprintf("%s стр. %d (%s)", ref, f.PageRef, f.RefDocName))
	}
	if f.PageTarget > 0 {
		where = append(where, fmt.Sprintf("%s стр. %d (%s)", tgt, f.PageTarget, f.TgtDocName))
	}
	return fmt.Sprintf("[%s, %s] %s — %s", SeverityLabel(f.Severity), KindLabel(f.Kind), strings.Join(where, " · "), f.Description)
}

// DialogSystem is the system prompt: the rules and the facts.
func DialogSystem(f DialogFacts) string {
	ref, tgt := firstNonEmpty(f.RefName, "ПД"), firstNonEmpty(f.TargetName, "РД")
	var b strings.Builder
	fmt.Fprintf(&b, `Ты помощник специалиста производственно-технического отдела. Помогаешь разобраться в расхождениях, найденных при сравнении документации (%s — эталон, %s — проверяемая).
Правила: опирайся ТОЛЬКО на факты ниже; не оценивай качество проектных решений и не предлагай улучшений; не выдумывай значений, номеров листов и документов; если по фактам ответить нельзя — так и скажи. Решение о замечании принимает специалист. Отвечай по-русски, кратко.
Чтобы направить специалиста к листу, пиши маркер [[%s:12]] или [[%s:7]] (имя стороны и номер страницы), а для документа без сторон — [[стр:12]] — только для листов, названных в фактах.
`, ref, tgt, ref, tgt)
	fmt.Fprintf(&b, "\nСравнение: %s\n", f.Title)
	if f.SetsLabel != "" {
		fmt.Fprintf(&b, "Версии комплектов: %s\n", f.SetsLabel)
	}
	if f.Similarity != "" {
		fmt.Fprintf(&b, "Схожесть: %s %%\n", f.Similarity)
	}
	for _, p := range f.Pages {
		side := ""
		switch p.Side {
		case SideRef:
			side = ref + ": "
		case SideTarget:
			side = tgt + ": "
		}
		fmt.Fprintf(&b, "\nОткрытый лист: %s%s, стр. %d из %d", side, p.Name, p.Page, p.Count)
		if p.Title != "" {
			fmt.Fprintf(&b, " (%s)", p.Title)
		}
		b.WriteString("\n")
		if p.Thin {
			b.WriteString("Текстового слоя почти нет (чертёж или скан): ниже только то, что удалось прочитать.\n")
		}
		fmt.Fprintf(&b, "Текст листа:\n%s\n", p.Text)
	}
	if f.ImageNote != "" {
		fmt.Fprintf(&b, "\n%s\n", f.ImageNote)
	}
	if f.Skill != "" {
		fmt.Fprintf(&b, "\nМетодика чтения чертежей (используй её, когда спрашивают о чертеже):\n%s\n", f.Skill)
	}
	if f.Selected != nil {
		fmt.Fprintf(&b, "\nВыбранное расхождение:\n%s\n", findingLine(*f.Selected, ref, tgt))
	}
	if len(f.Related) > 0 {
		b.WriteString("\nСоседние находки:\n")
		for _, r := range f.Related {
			b.WriteString("— " + findingLine(r, ref, tgt) + "\n")
		}
	}
	for _, e := range f.Excerpts {
		side := ref
		if e.Side == SideTarget {
			side = tgt
		}
		fmt.Fprintf(&b, "\nТекст листа: %s стр. %d (%s):\n%s\n", side, e.Page, e.Name, strings.Join(e.Lines, "\n"))
	}
	if len(f.Remarks) > 0 {
		b.WriteString("\nЗамечания экспертизы по этому сравнению:\n")
		for _, r := range f.Remarks {
			b.WriteString("— " + r + "\n")
		}
	}
	return b.String()
}

var jumpRe = regexp.MustCompile(`\[\[\s*([^\]:：]{1,10}?)\s*[:：]\s*(\d{1,4})\s*\]\]`)

// allowedJumps are the sheets the facts name.
func (f DialogFacts) allowedJumps() map[string]Jump {
	ref, tgt := firstNonEmpty(f.RefName, "ПД"), firstNonEmpty(f.TargetName, "РД")
	out := map[string]Jump{}
	add := func(side, doc string, page int) {
		if doc == "" || page <= 0 {
			return
		}
		name := ref
		if side == SideTarget {
			name = tgt
		}
		out[side+":"+strconv.Itoa(page)] = Jump{Side: side, Doc: doc, Page: page, Name: fmt.Sprintf("%s стр. %d", name, page)}
	}
	fs := append([]DialogFinding{}, f.Related...)
	if f.Selected != nil {
		fs = append(fs, *f.Selected)
	}
	for _, x := range fs {
		add(SideRef, x.RefDoc, x.PageRef)
		add(SideTarget, x.TargetDoc, x.PageTarget)
	}
	for _, e := range f.Excerpts {
		add(e.Side, e.Doc, e.Page)
	}
	for _, p := range f.Pages {
		if p.Side == SideRef || p.Side == SideTarget {
			add(p.Side, p.Doc, p.Page)
		}
		if p.Side == "" { // a document on its own: any page of it, marked [[стр:N]]
			for n := 1; n <= p.Count; n++ {
				out["doc:"+strconv.Itoa(n)] = Jump{Side: "doc", Doc: p.Doc, Page: n, Name: fmt.Sprintf("стр. %d", n)}
			}
		}
	}
	return out
}

// ExtractJumps turns the markers of an answer into text and the list of sheets that really exist in
// the facts; a marker for any other sheet is dropped.
func ExtractJumps(answer string, f DialogFacts) (string, []Jump) {
	ref, tgt := firstNonEmpty(f.RefName, "ПД"), firstNonEmpty(f.TargetName, "РД")
	allowed := f.allowedJumps()
	var jumps []Jump
	seen := map[string]bool{}
	clean := jumpRe.ReplaceAllStringFunc(answer, func(m string) string {
		sm := jumpRe.FindStringSubmatch(m)
		name := strings.ToLower(strings.TrimSpace(sm[1]))
		side := ""
		switch name {
		case strings.ToLower(ref):
			side = SideRef
		case strings.ToLower(tgt):
			side = SideTarget
		case "стр", "стр.", "лист":
			side = "doc"
		}
		page, _ := strconv.Atoi(sm[2])
		j, ok := allowed[side+":"+strconv.Itoa(page)]
		if side == "" || !ok {
			return "" // a sheet nobody told the model about
		}
		if !seen[side+":"+strconv.Itoa(page)] {
			seen[side+":"+strconv.Itoa(page)] = true
			jumps = append(jumps, j)
		}
		return j.Name
	})
	return strings.TrimSpace(strings.Join(strings.Fields(clean), " ")), jumps
}

// maxDialogMessages bounds the history sent to the model.
const maxDialogMessages = 10

// Converse answers the user's last message in the context of the facts.
func (s *Suggester) Converse(ctx context.Context, f DialogFacts, msgs []Msg) (string, []Jump, error) {
	if !s.Enabled() {
		return "", nil, fmt.Errorf("модель не настроена: задайте REMARK_MODEL или --remark-model")
	}
	var clean []Msg
	for _, m := range msgs {
		if (m.Role == "user" || m.Role == "assistant") && strings.TrimSpace(m.Content) != "" {
			clean = append(clean, Msg{Role: m.Role, Content: m.Content, Images: m.Images})
		}
	}
	if len(clean) == 0 || clean[len(clean)-1].Role != "user" {
		return "", nil, fmt.Errorf("нужно сообщение пользователя")
	}
	if len(clean) > maxDialogMessages {
		clean = clean[len(clean)-maxDialogMessages:]
		for len(clean) > 0 && clean[0].Role != "user" {
			clean = clean[1:]
		}
	}
	raw, err := s.chatMessages(ctx, DialogSystem(f), clean, false, 0.3)
	if err != nil {
		return "", nil, err
	}
	text, jumps := ExtractJumps(raw, f)
	return text, jumps, nil
}

// LoadSkillText reads the body of a skill (SKILL.md without its header) from the agent's skills
// directory (AI_SKILLS_DIR, or ./skills next to the agent) and cuts it to max runes at a paragraph.
// "" if the skill is not there: the assistant then works without the method.
func LoadSkillText(name string, max int, dirs ...string) string {
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, name, "SKILL.md"))
		if err != nil {
			continue
		}
		text := string(b)
		if strings.HasPrefix(text, "---") {
			if i := strings.Index(text[3:], "\n---"); i >= 0 {
				text = text[3+i+4:]
			}
		}
		text = strings.TrimSpace(text)
		r := []rune(text)
		if len(r) <= max {
			return text
		}
		cut := string(r[:max])
		if i := strings.LastIndex(cut, "\n\n"); i > max/2 {
			cut = cut[:i]
		}
		return strings.TrimSpace(cut) + "\n…"
	}
	return ""
}
