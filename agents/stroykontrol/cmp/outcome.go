package cmp

import "fmt"

// Результат проверки устранения замечания в терминах ТЗ (п. 3.4.6): четыре значения.
// Система выводит его из двух независимых источников — проверки по документам (Verdict*) и
// разбора ИИ (AI*) — и никогда не меняет по нему рабочий статус замечания. Решение специалиста
// (подтвердить / изменить) хранится отдельно от вывода системы.
const (
	OutcomeFixed    = "fixed"     // Устранено
	OutcomePartial  = "partial"   // Частично устранено
	OutcomeNotFixed = "not_fixed" // Не устранено
	OutcomeCheck    = "check"     // Требует проверки
)

// OutcomeLabels are the ТЗ names of the outcomes.
var OutcomeLabels = map[string]string{
	OutcomeFixed:    "Устранено",
	OutcomePartial:  "Частично устранено",
	OutcomeNotFixed: "Не устранено",
	OutcomeCheck:    "Требует проверки",
}

// Decisions of the specialist on the system's outcome.
const (
	DecisionConfirmed = "confirmed" // specialist agrees with the system's outcome
	DecisionChanged   = "changed"   // specialist set another outcome
)

// docOutcome is what the check against the documents says; "" when it says nothing.
func docOutcome(verdict string) string {
	switch verdict {
	case VerdictResolved:
		return OutcomeFixed
	case VerdictStillDiffers:
		return OutcomeNotFixed
	}
	return "" // no_match, no_documents, not checked
}

// aiOutcome is what the AI reading says; "" when it says nothing.
func aiOutcome(ai string) string {
	switch ai {
	case AIFixed:
		return OutcomeFixed
	case AIPartial:
		return OutcomePartial
	case AINotFixed:
		return OutcomeNotFixed
	}
	return "" // unclear, not run
}

// CheckOutcome combines the check against the documents and the AI reading into one of the four
// outcomes of ТЗ 3.4.6. One source silent → the other decides; both speak → they must agree,
// otherwise (and when both are silent) the outcome is «Требует проверки».
func CheckOutcome(verdict, ai string) string {
	d, a := docOutcome(verdict), aiOutcome(ai)
	switch {
	case d == "" && a == "":
		return OutcomeCheck
	case d == "":
		return a
	case a == "":
		return d
	case d == a:
		return d
	}
	return OutcomeCheck
}

// ValidOutcome reports whether code is one of the four outcomes.
func ValidOutcome(code string) bool {
	_, ok := OutcomeLabels[code]
	return ok
}

// ResolveDecision validates the specialist's decision on a remark whose system outcome is
// system and returns its kind. A changed outcome needs a comment (the decision must be explainable);
// an empty final outcome clears the decision.
func ResolveDecision(system, final, comment string) (string, error) {
	if final == "" {
		return "", nil
	}
	if !ValidOutcome(final) {
		return "", fmt.Errorf("unknown outcome %q", final)
	}
	if final == system {
		return DecisionConfirmed, nil
	}
	if comment == "" {
		return "", fmt.Errorf("outcome comment is required when the specialist changes the system outcome")
	}
	return DecisionChanged, nil
}
