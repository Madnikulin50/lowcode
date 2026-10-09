package riskengine

import "strings"

// suggestRule is one keyword -> draft-factor mapping used by SuggestFactors.
type suggestRule struct {
	keywords []string
	factor   RiskFactorDef
}

// suggestRules is a small, hand-curated library covering the domains the
// architecture note itself uses as examples (retail ops, vendor, project).
// It is intentionally not exhaustive — SuggestFactors always returns a
// non-empty draft (falling back to a generic likelihood/impact pair), and
// every returned factor is unsaved (ID/NamespaceID left zero) for a human to
// edit or discard in the Risk Model Editor before publishing.
var suggestRules = []suggestRule{
	{[]string{"усадк", "shrink"}, RiskFactorDef{Handle: "shrinkPct", Label: "Усадка", Scale: ScaleNumeric, Max: 10, Source: SourceField}},
	{[]string{"инцидент"}, RiskFactorDef{Handle: "incidents90d", Label: "Инциденты за 90 дней", Scale: ScaleNumeric, Max: 20, Source: SourceField}},
	{[]string{"аудит"}, RiskFactorDef{Handle: "daysSinceAudit", Label: "Давность аудита", Scale: ScaleStates, Source: SourceField, States: []RiskFactorState{
		{Code: "recent", Label: "Недавно", Severity: SeverityLow, Order: 0},
		{Code: "aging", Label: "Давно", Severity: SeverityMedium, Order: 1},
		{Code: "overdue", Label: "Просрочено", Severity: SeverityHigh, Order: 2},
	}}},
	{[]string{"текучест", "персонал", "кадр"}, RiskFactorDef{Handle: "staffTurnover", Label: "Текучесть персонала", Scale: ScaleStates, Source: SourceField, States: []RiskFactorState{
		{Code: "stable", Label: "Стабильно", Severity: SeverityLow, Order: 0},
		{Code: "elevated", Label: "Повышенная", Severity: SeverityMedium, Order: 1},
		{Code: "exodus", Label: "Массовый отток", Severity: SeverityCritical, Order: 2},
	}}},
	{[]string{"выручк", "доход"}, RiskFactorDef{Handle: "revenueImpact", Label: "Влияние на выручку", Scale: ScaleNumeric, Max: 5, Source: SourceField}},
	{[]string{"поставщик", "вендор", "vendor", "supplier"}, RiskFactorDef{Handle: "onTimeDeliveryPct", Label: "Доля поставок в срок", Scale: ScaleNumeric, Max: 100, Invert: true, Source: SourceField}},
	{[]string{"контракт", "договор"}, RiskFactorDef{Handle: "contractExpiryDays", Label: "Дней до окончания договора", Scale: ScaleNumeric, Max: 365, Invert: true, Source: SourceField}},
	{[]string{"проект", "срок", "дедлайн"}, RiskFactorDef{Handle: "scheduleSlipDays", Label: "Отставание от графика (дни)", Scale: ScaleNumeric, Max: 90, Source: SourceField}},
	{[]string{"бюджет"}, RiskFactorDef{Handle: "budgetOverrunPct", Label: "Перерасход бюджета, %", Scale: ScaleNumeric, Max: 50, Source: SourceField}},
	{[]string{"безопасност", "инцидент ИБ", "утечк", "кибер"}, RiskFactorDef{Handle: "securityIncidents90d", Label: "Инциденты ИБ за 90 дней", Scale: ScaleNumeric, Max: 10, Source: SourceField}},
	{[]string{"контрол"}, RiskFactorDef{Handle: "controlEffectiveness", Label: "Эффективность контролей", Scale: ScaleNumeric, Max: 1, Invert: true, Source: SourceField}},
}

// SuggestFactors returns a draft factor list for a free-text domain
// description, for a human to edit in the Risk Model Editor before saving —
// it never writes to the library itself. See this file's package doc
// comment (narrate.go) for the deterministic-v1 rationale and the upgrade
// path to a live risk-analyst agent call.
func SuggestFactors(description string) []RiskFactorDef {
	text := strings.ToLower(description)
	seen := map[string]bool{}
	var out []RiskFactorDef
	for _, rule := range suggestRules {
		for _, kw := range rule.keywords {
			if strings.Contains(text, kw) {
				if !seen[rule.factor.Handle] {
					seen[rule.factor.Handle] = true
					out = append(out, rule.factor)
				}
				break
			}
		}
	}
	if len(out) == 0 {
		// Generic fallback so the endpoint always returns something to
		// start from, matching a "likelihood x impact" matrix strategy.
		out = []RiskFactorDef{
			{Handle: "likelihood", Label: "Вероятность", Scale: ScaleNumeric, Max: 5, Source: SourceManual},
			{Handle: "impact", Label: "Влияние", Scale: ScaleNumeric, Max: 5, Source: SourceManual},
		}
	}
	return out
}
