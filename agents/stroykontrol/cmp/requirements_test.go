package cmp

import (
	"strings"
	"testing"
)

func texts(rs []Requirement) []string {
	var out []string
	for _, r := range rs {
		out = append(out, r.Text)
	}
	return out
}

func TestSplitRequirements(t *testing.T) {
	cases := []struct {
		name, in string
		want     []string
	}{
		{"inline numbering", "Нарушены требования. 1) Представить обоснование глубины заложения фундамента. 2) Уточнить класс бетона на листе 5; 3) Привести в соответствие марки.",
			[]string{"Представить обоснование глубины заложения фундамента.", "Уточнить класс бетона на листе 5;", "Привести в соответствие марки."}},
		{"lines with dashes", "Необходимо:\n– уточнить толщину стен по осям А–В;\n– указать марку арматуры в спецификации.",
			[]string{"уточнить толщину стен по осям А–В;", "указать марку арматуры в спецификации."}},
		{"numbered lines", "1. Представить расчёт плиты.\n2. Указать нагрузки на фундамент.", []string{"Представить расчёт плиты.", "Указать нагрузки на фундамент."}},
		{"letters", "а) уточнить класс бетона; б) указать морозостойкость", []string{"уточнить класс бетона;", "указать морозостойкость"}},
		{"sentences with a modal word", "Толщина стен не совпадает с ПД. Необходимо уточнить толщину стен. Следует привести в соответствие марку бетона.",
			[]string{"Необходимо уточнить толщину стен.", "Следует привести в соответствие марку бетона."}},
		{"one requirement is the whole text", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). Обосновать.", []string{"Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). Обосновать."}},
		{"a modal word in one sentence only", "По п. 4.2 СП 63 условие не выполнено. Необходимо представить расчёт.", []string{"По п. 4.2 СП 63 условие не выполнено. Необходимо представить расчёт."}},
		{"decimals and abbreviations are not markers", "Глубина 2.5 м, см. стр. 5. Необходимо уточнить отметку 2.1 на листе 3. Следует указать п. 4.2.", []string{"Необходимо уточнить отметку 2.1 на листе 3.", "Следует указать п. 4.2."}},
		{"a sentence with a modal preamble is a requirement", "Следует исправить нумерацию. 1) уточнить лист; 2) указать том.", []string{"Следует исправить нумерацию.", "уточнить лист;", "указать том."}},
		{"empty", "  ", nil},
	}
	for _, c := range cases {
		got := texts(SplitRequirements(c.in))
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
	rs := SplitRequirements("1) уточнить толщину стен; 2) указать марку; 3) уточнить толщину стен;")
	if len(rs) != 2 || rs[0].Idx != 1 || rs[1].Idx != 2 {
		t.Errorf("a repeated requirement is one, the numbers run on: %+v", rs)
	}
}

func TestAggregateOutcome(t *testing.T) {
	F, P, N, C := OutcomeFixed, OutcomePartial, OutcomeNotFixed, OutcomeCheck
	for _, c := range []struct {
		in   []string
		want string
	}{
		{nil, C}, {[]string{F}, F}, {[]string{F, F}, F}, {[]string{N, N}, N}, {[]string{C, C}, C},
		{[]string{F, N}, P}, {[]string{F, P}, P}, {[]string{P, N}, P}, {[]string{F, C}, C}, {[]string{P, C}, C},
		{[]string{N, C}, N}, {[]string{F, N, C}, P},
	} {
		if got := AggregateOutcome(c.in); got != c.want {
			t.Errorf("AggregateOutcome(%v) = %s, want %s", c.in, got, c.want)
		}
	}
	items := []Requirement{{Outcome: N}, {Outcome: N, OutcomeFinal: F}}
	if AggregateRequirements(items) != P {
		t.Error("the specialist's decision on a requirement counts instead of the system's")
	}
	if c := RequirementCounts(items); c[N] != 1 || c[F] != 1 {
		t.Errorf("counts: %v", c)
	}
}

func TestEnsureRequirementsKeepsTheListWhileTheTextIsTheSame(t *testing.T) {
	text := "1) уточнить толщину стен; 2) указать марку бетона"
	set, changed := EnsureRequirements(RequirementSet{}, text)
	if !changed || len(set.Items) != 2 || set.For != TextHash(text) {
		t.Fatalf("a new remark gets its list: %+v", set)
	}
	set.Items[0].Verdict = VerdictResolved
	if again, changed := EnsureRequirements(set, text+"  "); changed || again.Items[0].Verdict != VerdictResolved {
		t.Errorf("the same text (spaces aside) keeps the list and its results: %+v", again)
	}
	// the text changed, the list is the program's: split again
	if again, changed := EnsureRequirements(set, "1) уточнить толщину стен; 2) указать марку; 3) привести узел"); !changed || len(again.Items) != 3 {
		t.Errorf("a new text, a new list: %+v", again)
	}
	// the list was edited by hand: it stays, and says that the text moved on
	set.Manual = true
	again, changed := EnsureRequirements(set, "совсем другой текст замечания, который нужно уточнить")
	if !changed || !again.Stale || len(again.Items) != 2 {
		t.Errorf("a hand-made list outlives a change of the text: %+v", again)
	}
	if again2, changed := EnsureRequirements(again, "совсем другой текст замечания, который нужно уточнить"); changed || !again2.Stale {
		t.Error("already stale: nothing more to write")
	}
}

func TestApplyRequirementEdit(t *testing.T) {
	text := "1) уточнить толщину стен; 2) указать марку бетона"
	stored, _ := EnsureRequirements(RequirementSet{}, text)
	stored.Items[0].Verdict, stored.Items[0].Outcome = VerdictResolved, OutcomeFixed
	stored.Items[1].Outcome = OutcomeNotFixed
	in := []RequirementInput{
		{Text: "уточнить толщину стен;"},                                      // unchanged: keeps the result
		{Text: "указать марку бетона и арматуры", OutcomeFinal: OutcomeFixed}, // edited: no result yet, a decision needs a comment
		{Text: "привести узел примыкания"},
	}
	if _, err := ApplyRequirementEdit(stored, text, in, "ПТО", "2026-10-22T10:00:00Z"); err == nil {
		t.Fatal("a decision with another result than the system's needs a comment")
	}
	in[1].OutcomeComment = "проверено по листу 7"
	set, err := ApplyRequirementEdit(stored, text, in, "ПТО", "2026-10-22T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if !set.Manual || len(set.Items) != 3 || set.Items[0].Outcome != OutcomeFixed || set.Items[1].Outcome != "" || set.Items[1].OutcomeFinal != OutcomeFixed || set.Items[1].OutcomeBy != "ПТО" || !set.Items[2].Manual {
		t.Errorf("%+v", set)
	}
	if _, err := ApplyRequirementEdit(stored, text, nil, "", ""); err == nil {
		t.Error("an empty list is refused")
	}
	if _, err := ApplyRequirementEdit(stored, text, []RequirementInput{{Text: " "}}, "", ""); err == nil {
		t.Error("an empty requirement is refused")
	}
	// the set round-trips through its JSON
	back, err := ParseRequirementSet(set.JSON())
	if err != nil || len(back.Items) != 3 || back.Items[2].Idx != 3 || !back.Manual {
		t.Errorf("json: %v %+v", err, back)
	}
	if s, err := ParseRequirementSet(""); err != nil || len(s.Items) != 0 {
		t.Error("no value, no list")
	}
}
