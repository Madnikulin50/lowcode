package idcheck

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

func f3(p *float64) string {
	if p == nil {
		return "—"
	}
	return trimFloat(*p)
}

func trimFloat(f float64) string {
	s := fmt.Sprintf("%.4f", f)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s
}

// MatchRegistryRow finds the registry act row for an act number + date.
// The registry lists a multi-position act once per position; the first
// row (in sheet order) is returned, use MatchRegistryRows for all of them.
func MatchRegistryRow(reg *Registry, num ActNumber, date time.Time) *RegistryRow {
	rows := MatchRegistryRows(reg, num, date)
	if len(rows) == 0 {
		return nil
	}
	return rows[0]
}

// MatchRegistryRows returns every registry act row that carries the act's
// number. Matching is on the canonical number key; when the date is known
// rows with a different date are still returned (check 1 reports the date
// mismatch) only if no row matches both.
func MatchRegistryRows(reg *Registry, num ActNumber, date time.Time) []*RegistryRow {
	key := num.Key()
	var exact, byKey []*RegistryRow
	for _, r := range reg.Acts() {
		rn := ParseActNumber(r.DocNumber, '/')
		if !sameActKey(rn, num, key) {
			continue
		}
		byKey = append(byKey, r)
		if date.IsZero() || r.DocDate.IsZero() || r.DocDate.Equal(date) {
			exact = append(exact, r)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	if len(byKey) > 0 {
		return byKey
	}
	// Fallback for a mistyped number in the registry (e.g. "3/2.14.1.6/…"
	// listed under position 2.14.2.6): same sequence number and date, listed
	// under one of the act's positions. Check 1 then reports the number
	// mismatch instead of "not found".
	if date.IsZero() {
		return nil
	}
	var loose []*RegistryRow
	for _, r := range reg.Acts() {
		rn := ParseActNumber(r.DocNumber, '/')
		if rn.Seq == num.Seq && r.DocDate.Equal(date) && num.HasPosition(r.Position) {
			loose = append(loose, r)
		}
	}
	return loose
}

// sameActKey compares numbers tolerating a registry cell truncated by the
// sheet ("…/ИССО17." vs "…/ИССО17.2/Б4.3-Б7.3/МТК"): seq and positions must
// match exactly, the tail must agree on the common prefix.
func sameActKey(a, b ActNumber, bKey string) bool {
	if a.Key() == bKey {
		return true
	}
	if a.Seq != b.Seq || strings.Join(a.Positions, ";") != strings.Join(b.Positions, ";") {
		return false
	}
	at, bt := strings.Join(a.Rest, "/"), strings.Join(b.Rest, "/")
	if at == "" || bt == "" {
		return true
	}
	return strings.HasPrefix(at, strings.TrimRight(bt, ".")) || strings.HasPrefix(bt, strings.TrimRight(at, "."))
}

// CheckRegistryConsistency runs the registry-only part of check 1: every
// act row's number must name the position it is listed under, and its date
// must parse.
func CheckRegistryConsistency(reg *Registry) []Finding {
	var out []Finding
	for _, r := range reg.Acts() {
		if r.DocNumber == "" {
			continue
		}
		n := ParseActNumber(r.DocNumber, '/')
		if len(n.Positions) > 0 && !n.HasPosition(r.Position) {
			out = append(out, Finding{
				Check: 1, Severity: Medium, Field: "doc_number", Source: "det",
				RegistryRow: r.Row, Position: r.Position,
				Expected:    "позиция " + r.Position,
				Actual:      r.DocNumber,
				Description: fmt.Sprintf("Реестр, строка %d: номер акта %q не соответствует позиции ВРЦ %s, под которой он указан", r.Row, r.DocNumber, r.Position),
			})
		}
		if r.DocDate.IsZero() {
			out = append(out, Finding{
				Check: 1, Severity: Low, Field: "doc_date", Source: "det",
				RegistryRow: r.Row, Position: r.Position, Actual: r.DocDateRaw,
				Description: fmt.Sprintf("Реестр, строка %d: дата акта не распознана (%q)", r.Row, r.DocDateRaw),
			})
		}
		if (r.QtyProject == nil) != (r.QtyFact == nil) {
			out = append(out, Finding{
				Check: 1, Severity: Low, Field: "qty", Source: "det",
				RegistryRow: r.Row, Position: r.Position,
				Description: fmt.Sprintf("Реестр, строка %d: заполнен только один из столбцов 12/13 (проект %s, факт %s)", r.Row, f3(r.QtyProject), f3(r.QtyFact)),
			})
		}
	}
	return out
}

// ReconcileKS2 is check 8: per position, Σ min(col 12, col 13) over the
// registry's act rows vs the КС-2 quantity (and vs the registry's own
// col 11 "КС-2" on the position row).
func ReconcileKS2(reg *Registry, ks *KS2) []PositionSummary {
	sums := map[string]float64{}
	has := map[string]bool{}
	for _, r := range reg.Acts() {
		if r.QtyProject == nil && r.QtyFact == nil {
			continue
		}
		v := math.Inf(1)
		if r.QtyProject != nil {
			v = *r.QtyProject
		}
		if r.QtyFact != nil && *r.QtyFact < v {
			v = *r.QtyFact
		}
		sums[r.Position] += v
		has[r.Position] = true
	}
	ksQty := map[string]*KS2Row{}
	for i := range ks.Rows {
		ksQty[ks.Rows[i].Position] = &ks.Rows[i]
	}
	positions := reg.Positions()
	keys := map[string]bool{}
	for p := range positions {
		keys[p] = true
	}
	for p := range ksQty {
		keys[p] = true
	}
	var order []string
	for p := range keys {
		order = append(order, p)
	}
	sort.Slice(order, func(i, j int) bool { return positionLess(order[i], order[j]) })

	var out []PositionSummary
	for _, p := range order {
		s := PositionSummary{Position: p, RegistrySum: sums[p]}
		if pr := positions[p]; pr != nil {
			s.Name, s.Unit, s.RegistryKS2 = pr.WorkName, pr.Unit, pr.QtyKS2
		}
		if k := ksQty[p]; k != nil {
			q := k.Qty
			s.KS2Qty = &q
			if s.Name == "" {
				s.Name, s.Unit = k.Name, k.Unit
			}
		}
		switch {
		case s.KS2Qty == nil && !has[p]:
			continue // position with nothing in this period
		case s.KS2Qty == nil:
			s.Status = "missing_ks2"
			s.Delta = s.RegistrySum
		case !has[p]:
			s.Status = "missing_registry"
			s.Delta = -*s.KS2Qty
		default:
			s.Delta = s.RegistrySum - *s.KS2Qty
			s.Status = "ok"
			if !NearlyEqual(s.RegistrySum, *s.KS2Qty) {
				s.Status = "mismatch"
			}
		}
		out = append(out, s)
	}
	return out
}

func positionLess(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		var ai, bi int
		fmt.Sscan(as[i], &ai)
		fmt.Sscan(bs[i], &bi)
		if ai != bi {
			return ai < bi
		}
	}
	return len(as) < len(bs)
}

// Check8Findings turns position summaries into findings.
func Check8Findings(sum []PositionSummary) []Finding {
	var out []Finding
	for _, s := range sum {
		switch s.Status {
		case "mismatch":
			out = append(out, Finding{
				Check: 8, Severity: High, Field: "qty", Source: "det", Position: s.Position,
				Expected: f3(s.KS2Qty) + " " + s.Unit, Actual: trimFloat(s.RegistrySum) + " " + s.Unit,
				Description: fmt.Sprintf("Позиция %s: Σ min(кол.12, кол.13) по реестру = %s, в КС-2 = %s (разница %s %s)",
					s.Position, trimFloat(s.RegistrySum), f3(s.KS2Qty), trimFloat(s.Delta), s.Unit),
			})
		case "missing_ks2":
			out = append(out, Finding{
				Check: 8, Severity: High, Field: "qty", Source: "det", Position: s.Position,
				Actual:      trimFloat(s.RegistrySum),
				Description: fmt.Sprintf("Позиция %s: объём по реестру %s %s, в КС-2 позиция отсутствует", s.Position, trimFloat(s.RegistrySum), s.Unit),
			})
		case "missing_registry":
			out = append(out, Finding{
				Check: 8, Severity: High, Field: "qty", Source: "det", Position: s.Position,
				Expected:    f3(s.KS2Qty),
				Description: fmt.Sprintf("Позиция %s: в КС-2 %s %s, в реестре нет актов с объёмом", s.Position, f3(s.KS2Qty), s.Unit),
			})
		}
		if s.RegistryKS2 != nil && s.KS2Qty != nil && !NearlyEqual(*s.RegistryKS2, *s.KS2Qty) {
			out = append(out, Finding{
				Check: 8, Severity: Medium, Field: "qty_ks2", Source: "det", Position: s.Position,
				Expected: f3(s.KS2Qty), Actual: f3(s.RegistryKS2),
				Description: fmt.Sprintf("Позиция %s: столбец 11 реестра «КС-2» = %s не совпадает с формой КС-2 = %s", s.Position, f3(s.RegistryKS2), f3(s.KS2Qty)),
			})
		}
	}
	return out
}

// ActInput bundles everything known about one act.
type ActInput struct {
	Key      string // stable id for findings (file name)
	FileNum  ActNumber
	FileDate time.Time
	Rows     []*RegistryRow
	Data     *ActData
	Norms    *NormCatalog
}

func (in *ActInput) actDate() time.Time {
	if in.Data != nil {
		if d, ok := ParseDate(in.Data.Date); ok {
			return d
		}
	}
	if !in.FileDate.IsZero() {
		return in.FileDate
	}
	if len(in.Rows) > 0 {
		return in.Rows[0].DocDate
	}
	return time.Time{}
}

func (in *ActInput) find(check int, sev Severity, field, expected, actual, desc string) Finding {
	f := Finding{Check: check, Severity: sev, Field: field, Expected: expected, Actual: actual, Description: desc, Source: "det", ActKey: in.Key}
	if len(in.Rows) > 0 {
		f.RegistryRow = in.Rows[0].Row
		f.Position = in.Rows[0].Position
	}
	return f
}

// CheckAct runs checks 1–6 for one act (7 and the online part of 6 need
// network access and run separately, see External).
func CheckAct(in *ActInput) []Finding {
	var out []Finding
	out = append(out, check1(in)...)
	if in.Data == nil {
		return out
	}
	out = append(out, check2(in)...)
	out = append(out, check3(in)...)
	out = append(out, check4(in)...)
	out = append(out, check5(in)...)
	out = append(out, check6(in)...)
	out = append(out, checkOrgRequisites(in)...)
	return out
}

func check1(in *ActInput) []Finding {
	var out []Finding
	if len(in.Rows) == 0 {
		return append(out, in.find(1, High, "registry", "строка реестра", in.FileNum.Raw,
			fmt.Sprintf("Акт %s от %s не найден в реестре", in.FileNum.Raw, FmtDate(in.FileDate))))
	}
	row := in.Rows[0]
	actDate := in.actDate()
	if !actDate.IsZero() && !row.DocDate.IsZero() && !actDate.Equal(row.DocDate) {
		out = append(out, in.find(1, High, "date", FmtDate(row.DocDate), FmtDate(actDate),
			fmt.Sprintf("Дата акта %s не совпадает с датой в реестре %s (строка %d)", FmtDate(actDate), FmtDate(row.DocDate), row.Row)))
	}
	if in.Data != nil && in.Data.Number != "" {
		dn := ParseActNumber(in.Data.Number, '/')
		rn := ParseActNumber(row.DocNumber, '/')
		if !sameActKey(dn, rn, rn.Key()) {
			// the file name agrees with the registry and the document differs
			// by a character or two: most likely a recognition slip
			if levenshtein(dn.Key(), rn.Key()) <= 2 && sameActKey(in.FileNum, rn, rn.Key()) {
				out = append(out, in.find(1, Low, "number", row.DocNumber, in.Data.Number,
					fmt.Sprintf("Номер акта в документе распознан как %q, в реестре %q — вероятна ошибка распознавания, сверьте со сканом", in.Data.Number, row.DocNumber)))
			} else {
				out = append(out, in.find(1, Medium, "number", row.DocNumber, in.Data.Number,
					fmt.Sprintf("Номер акта в документе %q не совпадает с номером в реестре %q", in.Data.Number, row.DocNumber)))
			}
		}
	}
	if in.Data == nil {
		return out
	}
	// volumes: rows that carry quantities
	var qtyRows []*RegistryRow
	for _, r := range in.Rows {
		if r.QtyProject != nil || r.QtyFact != nil {
			qtyRows = append(qtyRows, r)
		}
	}
	if len(qtyRows) == 0 {
		return out
	}
	proj, fact, src := in.Data.P1VolProject, in.Data.P1VolFact, "п.1 акта"
	if proj == nil && fact == nil {
		proj, fact, src = in.Data.SchemeVolProj, in.Data.SchemeVolFact, "исполнительной схеме"
	}
	if proj == nil && fact == nil {
		return append(out, in.find(1, Medium, "volume", "объём", "",
			"Объём работ не указан ни в п.1 акта, ни на исполнительной схеме — сверка с кол. 12/13 реестра невозможна"))
	}
	for _, r := range qtyRows {
		okP := proj == nil || r.QtyProject == nil || NearlyEqual(*proj, *r.QtyProject)
		okF := fact == nil || r.QtyFact == nil || NearlyEqual(*fact, *r.QtyFact)
		if okP && okF {
			return out
		}
	}
	r := qtyRows[0]
	return append(out, in.find(1, High, "volume",
		fmt.Sprintf("проект %s / факт %s", f3(r.QtyProject), f3(r.QtyFact)),
		fmt.Sprintf("проект %s / факт %s", f3(proj), f3(fact)),
		fmt.Sprintf("Объём по %s (Vпр=%s, Vф=%s) не совпадает с реестром, строка %d (кол.12=%s, кол.13=%s)",
			src, f3(proj), f3(fact), r.Row, f3(r.QtyProject), f3(r.QtyFact))))
}

// docMatch reports whether two references point at the same document.
func docMatch(a, b DocRef) bool {
	an, bn := compact(a.Number), compact(b.Number)
	if an != "" && bn != "" {
		return an == bn || strings.Contains(an, bn) || strings.Contains(bn, an)
	}
	ad, aok := ParseDate(a.Date)
	bd, bok := ParseDate(b.Date)
	if aok && bok && ad.Equal(bd) {
		return a.Kind == "" || b.Kind == "" || a.Kind == b.Kind
	}
	return false
}

func check2(in *ActInput) []Finding {
	var out []Finding
	d := in.Data
	actDate := in.actDate()
	// п.4 documents must be listed in the appendix
	for _, ref := range d.P4Docs {
		if ref.Number == "" && ref.Date == "" {
			continue
		}
		found := false
		for _, a := range d.Appendix {
			if docMatch(ref, a) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, in.find(2, Medium, "appendix", describeDoc(ref), "",
				fmt.Sprintf("Документ из п.4 «%s» отсутствует в перечне приложений", describeDoc(ref))))
		}
	}
	// appendix items must be physically attached (scanned) — only for kinds
	// that are expected inside the act file
	for _, a := range d.Appendix {
		if a.Number == "" && a.Date == "" {
			continue
		}
		found := false
		for _, s := range d.Attached {
			if docMatch(a, s) {
				found = true
				break
			}
		}
		if !found && len(d.Attached) > 0 {
			out = append(out, in.find(2, Medium, "attached", describeDoc(a), "",
				fmt.Sprintf("Приложение «%s» не найдено среди вложенных в акт сканов", describeDoc(a))))
		}
		if dt, ok := ParseDate(a.Date); ok && !actDate.IsZero() && dt.After(actDate) {
			out = append(out, in.find(2, High, "appendix_date", "не позднее "+FmtDate(actDate), FmtDate(dt),
				fmt.Sprintf("Приложение «%s» датировано позже акта (%s > %s)", describeDoc(a), FmtDate(dt), FmtDate(actDate))))
		}
	}
	// attached scans: dates, validity at the time of work
	workEnd, _ := ParseDate(d.P5End)
	workStart, _ := ParseDate(d.P5Start)
	if workEnd.IsZero() {
		workEnd = actDate
	}
	seenLate := map[string]bool{}
	for _, s := range d.Attached {
		// the same document appears in the materials register and on its own
		// page — report it once
		key := compact(s.Number) + "|" + s.Date
		if dt, ok := ParseDate(s.Date); ok && !actDate.IsZero() && dt.After(actDate) && !seenLate[key] {
			seenLate[key] = true
			out = append(out, Finding{Check: 2, Severity: High, Field: "attached_date", Source: "det", ActKey: in.Key, Page: s.Page,
				Expected: "не позднее " + FmtDate(actDate), Actual: FmtDate(dt),
				Description: fmt.Sprintf("Вложенный документ «%s» (стр. %d) датирован позже акта", describeDoc(s), s.Page)})
		}
		if vu, ok := ParseDate(s.ValidUntil); ok {
			ref := workEnd
			if !workStart.IsZero() && workStart.After(ref) {
				ref = workStart
			}
			if !ref.IsZero() && vu.Before(ref) {
				out = append(out, Finding{Check: 2, Severity: High, Field: "valid_until", Source: "det", ActKey: in.Key, Page: s.Page,
					Expected: "действует на " + FmtDate(ref), Actual: "до " + FmtDate(vu),
					Description: fmt.Sprintf("Срок действия/годности «%s» (%s) истёк до производства работ (%s)", describeDoc(s), FmtDate(vu), FmtDate(ref))})
			}
		}
	}
	if len(d.P3Materials) > 0 {
		hasQuality := false
		for _, s := range d.Attached {
			if s.Kind == "сертификат" || s.Kind == "паспорт" || s.Kind == "реестр" || s.Kind == "декларация" {
				hasQuality = true
			}
		}
		for _, a := range append(append([]DocRef{}, d.Appendix...), d.P4Docs...) {
			t := strings.ToLower(a.Kind + " " + a.Name)
			if strings.Contains(t, "реестр") || strings.Contains(t, "сертификат") || strings.Contains(t, "паспорт") || strings.Contains(t, "документ о качестве") {
				hasQuality = true
			}
		}
		if !hasQuality {
			out = append(out, in.find(2, Medium, "p3", "документы о качестве", "",
				"В п.3 указаны материалы, но ни реестр документов о качестве, ни сертификаты/паспорта не приложены"))
		}
	}
	return out
}

func describeDoc(d DocRef) string {
	parts := []string{}
	if d.Name != "" {
		parts = append(parts, d.Name)
	} else if d.Kind != "" {
		parts = append(parts, d.Kind)
	}
	if d.Number != "" {
		parts = append(parts, "№"+d.Number)
	}
	if d.Date != "" {
		parts = append(parts, "от "+d.Date)
	}
	return strings.Join(parts, " ")
}

func hasLetters(s string) bool {
	for _, r := range s {
		if (r >= 'А' && r <= 'я') || (r >= 'A' && r <= 'z') {
			return true
		}
	}
	return false
}

// rdEqual tolerates a sheet list or extra suffix on one side and stray
// dots ("КЖ.4.1" vs "КЖ4.1").
func rdEqual(a, b string) bool {
	a, b = NormalizeRDCode(a), NormalizeRDCode(b)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	if a == "" || b == "" {
		return true
	}
	return a == b || strings.Contains(a, b) || strings.Contains(b, a)
}

// primaryRDCode returns the first code-looking token of a free-text "шифр"
// field ("03102022/РД-ИССО2.2-ОП-КЖ4.1 (лист №1,5,6). ИССО2.2" → the code).
func primaryRDCode(s string) string {
	if c := FindRDCodes(rdSheetsRe.ReplaceAllString(s, " ")); len(c) > 0 {
		return c[0]
	}
	return ""
}

func check3(in *ActInput) []Finding {
	var out []Finding
	d := in.Data
	p2 := primaryRDCode(d.P2RDCode)
	if p2 == "" {
		p2 = primaryRDCode(d.P2Text)
	}
	if p2 == "" {
		return append(out, in.find(3, Medium, "p2_rd_code", "шифр РД", "", "В п.2 акта не указан шифр рабочей документации"))
	}
	// п.6 lists ППР, РД and norms together; the RD code must be among them
	var p6 []string
	for _, src := range append([]string{d.P6RDCode}, d.P6Norms...) {
		if isNormCode(src) {
			continue
		}
		p6 = append(p6, FindRDCodes(rdSheetsRe.ReplaceAllString(src, " "))...)
	}
	switch {
	case len(p6) == 0:
		out = append(out, in.find(3, Medium, "p6_rd_code", p2, "", "В п.6 акта не указан шифр проектной/рабочей документации"))
	default:
		match := false
		for _, c := range p6 {
			match = match || rdEqual(p2, c)
		}
		if !match {
			out = append(out, in.find(3, High, "p6_rd_code", p2, strings.Join(p6, "; "),
				fmt.Sprintf("Шифр РД в п.2 (%s) не найден среди документов п.6 (%s)", p2, strings.Join(p6, "; "))))
		}
	}
	scheme := primaryRDCode(d.SchemeRDCode)
	if scheme == "" {
		scheme = d.SchemeRDCode
	}
	if scheme == "" {
		out = append(out, in.find(3, Low, "scheme_rd_code", p2, "", "Шифр в основной надписи исполнительной схемы не распознан"))
	} else if !rdEqual(p2, scheme) {
		sev := High
		desc := fmt.Sprintf("Шифр РД в п.2 (%s) не совпадает с шифром в основной надписи исполнительной схемы (%s)", p2, d.SchemeRDCode)
		if !hasLetters(scheme) {
			// "2.2-0824-178": a registration number near the stamp, not a code
			sev = Low
			desc = fmt.Sprintf("С исполнительной схемы распознан номер %q, не похожий на шифр РД (в п.2 — %s) — сверьте основную надпись со сканом", d.SchemeRDCode, p2)
		}
		f := in.find(3, sev, "scheme_rd_code", p2, d.SchemeRDCode, desc)
		f.Page = d.SchemePage
		out = append(out, f)
	}
	return out
}

func check4(in *ActInput) []Finding {
	d := in.Data
	if d.SchemeVolProj == nil && d.SchemeVolFact == nil {
		if d.P1VolProject != nil || d.P1VolFact != nil {
			f := in.find(4, Low, "scheme_volume", "объём на схеме", "", "Объём на исполнительной схеме не распознан — сверка с п.1 не выполнена")
			f.Page = d.SchemePage
			return []Finding{f}
		}
		return nil
	}
	var out []Finding
	cmp := func(field, label string, act, scheme *float64) {
		if act == nil || scheme == nil || NearlyEqual(*act, *scheme) {
			return
		}
		f := in.find(4, High, field, f3(scheme), f3(act), fmt.Sprintf("%s объём в п.1 акта (%s) не совпадает с исполнительной схемой (%s)", label, f3(act), f3(scheme)))
		f.Page = d.SchemePage
		out = append(out, f)
	}
	cmp("vol_project", "Проектный", d.P1VolProject, d.SchemeVolProj)
	cmp("vol_fact", "Фактический", d.P1VolFact, d.SchemeVolFact)
	return out
}

func check5(in *ActInput) []Finding {
	var out []Finding
	d := in.Data
	if in.Norms == nil {
		return nil
	}
	start, _ := ParseDate(d.P5Start)
	end, _ := ParseDate(d.P5End)
	if start.IsZero() {
		start = in.actDate()
	}
	if end.IsZero() {
		end = start
	}
	cited := ExtractNormCodes(strings.Join(d.P6Norms, "\n"))
	if len(cited) == 0 {
		return append(out, in.find(5, High, "p6", "нормативные документы", "", "В п.6 акта не указаны нормативные документы"))
	}
	for _, code := range cited {
		n := in.Norms.Lookup(code)
		if n == nil {
			out = append(out, in.find(5, Low, "p6", "", code, fmt.Sprintf("%s отсутствует в справочнике нормативных документов — актуальность не проверена", code)))
			continue
		}
		for _, dt := range []time.Time{start, end} {
			if dt.IsZero() {
				continue
			}
			if !n.ValidFrom.IsZero() && dt.Before(n.ValidFrom) {
				out = append(out, in.find(5, High, "p6", "действует с "+FmtDate(n.ValidFrom), code,
					fmt.Sprintf("%s на дату работ %s ещё не действовал (введён %s)", code, FmtDate(dt), FmtDate(n.ValidFrom))))
				break
			}
			if !n.ValidTo.IsZero() && !dt.Before(n.ValidTo) {
				repl := ""
				if n.ReplacedBy != "" {
					repl = ", заменён " + n.ReplacedBy
				}
				out = append(out, in.find(5, High, "p6", n.ReplacedBy, code,
					fmt.Sprintf("%s на дату работ %s утратил силу (с %s%s)", code, FmtDate(dt), FmtDate(n.ValidTo), repl)))
				break
			}
		}
	}
	work := strings.ToLower(d.P1Work + " " + d.P2Text)
	for _, req := range in.Norms.Required(work) {
		ok := false
		for _, c := range cited {
			if strings.HasPrefix(normKey(c), normKey(req.Prefix)) {
				ok = true
				break
			}
		}
		if !ok {
			out = append(out, in.find(5, Medium, "p6_sufficiency", req.Prefix, "",
				fmt.Sprintf("Для работ «%s» в п.6 не указан %s (%s)", req.Tag, req.Prefix, req.Title)))
		}
	}
	text := strings.ToLower(strings.Join(d.P6Norms, " "))
	if d.P6RDCode == "" && !strings.Contains(text, "рд") && !strings.Contains(text, "рабоч") && !strings.Contains(text, "проект") {
		out = append(out, in.find(5, Medium, "p6_project", "проектная/рабочая документация", "", "В п.6 не указана проектная (рабочая) документация"))
	}
	return out
}

var (
	ruleDateRe = regexp.MustCompile(`(?i)(?:от|от\s*:)\s*([0-9]{1,2}\.[0-9]{1,2}\.[0-9]{2,4})`)
	nrsRe      = regexp.MustCompile(`(?i)[СC]\s*-\s*\d{2}\s*-\s*\d{3,7}|П[ИI]\s*-\s*\d{2}\s*-\s*\d{3,7}`)
)

// NormalizeNRS reduces "№ С-77-246659" to "С-77-246659".
func NormalizeNRS(s string) string {
	m := nrsRe.FindString(s)
	if m == "" {
		return ""
	}
	return compact(m)
}

// stroyControlRole reports whether a representative role requires an NRS
// identification number (строительный контроль застройщика / лица,
// осуществляющего строительство — ГрК РФ ст. 55.5-1).
func stroyControlRole(role string) bool {
	r := strings.ToLower(role)
	return strings.Contains(r, "строительного контроля") || strings.Contains(r, "строительному контролю") ||
		strings.Contains(r, "строительный контроль") || strings.Contains(r, "застройщ") || strings.Contains(r, "технического заказчика")
}

func check6(in *ActInput) []Finding {
	var out []Finding
	d := in.Data
	actDate := in.actDate()
	start, _ := ParseDate(d.P5Start)
	for _, p := range d.People {
		who := strings.TrimSpace(p.FIO)
		if who == "" {
			who = p.Role
		}
		od, ok := ParseDate(p.OrderDate)
		if !ok && p.OrderNo != "" {
			if m := ruleDateRe.FindStringSubmatch(p.OrderNo); m != nil {
				od, ok = ParseDate(m[1])
			}
		}
		switch {
		case p.OrderNo == "" && p.OrderDate == "":
			out = append(out, in.find(6, Medium, "order", "реквизиты распорядительного документа", "", fmt.Sprintf("%s: не указан приказ/распорядительный документ", who)))
		case !ok:
			out = append(out, in.find(6, Low, "order_date", "дата приказа", p.OrderNo, fmt.Sprintf("%s: дата приказа не распознана (%s)", who, p.OrderNo)))
		case !actDate.IsZero() && od.After(actDate):
			desc := fmt.Sprintf("%s: приказ %s от %s выпущен позже даты акта %s («приказ из будущего»)", who, p.OrderNo, FmtDate(od), FmtDate(actDate))
			if od.After(time.Now()) {
				desc += " — дата позже сегодняшней, вероятна опечатка в акте или ошибка распознавания, сверьте со сканом"
			}
			out = append(out, in.find(6, High, "order_date", "не позднее "+FmtDate(actDate), FmtDate(od), desc))
		case !start.IsZero() && od.After(start):
			out = append(out, in.find(6, Medium, "order_date", "не позднее начала работ "+FmtDate(start), FmtDate(od),
				fmt.Sprintf("%s: приказ от %s выпущен после начала работ %s", who, FmtDate(od), FmtDate(start))))
		}
		if stroyControlRole(p.Role) && NormalizeNRS(p.NRSID) == "" {
			out = append(out, in.find(6, Medium, "nrs_id", "идентификационный номер НРС", p.NRSID,
				fmt.Sprintf("%s (%s): не указан идентификационный номер в НРС", who, p.Role)))
		}
	}
	return out
}

// checkOrgRequisites is the offline part of check 7: OGRN/INN checksums.
func checkOrgRequisites(in *ActInput) []Finding {
	var out []Finding
	for _, o := range in.Data.Orgs {
		name := o.Name
		check := func(field, label, raw string, lengths []int, valid func(string) bool) {
			if raw == "" {
				return
			}
			v := FirstDigitRun(raw)
			okLen := false
			for _, l := range lengths {
				okLen = okLen || len(v) == l
			}
			glued := len(v) > lengths[len(lengths)-1]+2 // e.g. ИНН run into the postcode without a separator
			switch {
			case !okLen && glued:
				out = append(out, in.find(7, Low, field, label, raw, fmt.Sprintf("%s: %s распознан неуверенно (%q) — сверьте со сканом", name, label, raw)))
			case !okLen:
				out = append(out, in.find(7, High, field, fmt.Sprintf("%s из %v цифр", label, lengths), v,
					fmt.Sprintf("%s: %s %s содержит %d цифр (должно быть %v) — ошибка в реквизитах акта (или распознавания, сверьте со сканом)", name, label, v, len(v), lengths)))
			case !valid(v):
				out = append(out, in.find(7, High, field, "корректный "+label, v, fmt.Sprintf("%s: %s %s не проходит проверку контрольного числа", name, label, v)))
			}
		}
		check("ogrn", "ОГРН", o.OGRN, []int{13, 15}, ValidOGRN)
		check("inn", "ИНН", o.INN, []int{10, 12}, ValidINN)
		if o.OGRN == "" && o.INN == "" && name != "" {
			out = append(out, in.find(7, Medium, "ogrn", "ОГРН/ИНН", "", fmt.Sprintf("%s: не указаны ОГРН и ИНН", name)))
		}
	}
	return out
}

// ValidOGRN checks the control digit of a 13-digit ОГРН / 15-digit ОГРНИП.
func ValidOGRN(s string) bool {
	s = digits(s)
	switch len(s) {
	case 13:
		var n uint64
		fmt.Sscan(s[:12], &n)
		return int(n%11%10) == int(s[12]-'0')
	case 15:
		var n uint64
		fmt.Sscan(s[:14], &n)
		return int(n%13%10) == int(s[14]-'0')
	}
	return false
}

// ValidINN checks the control digits of a 10/12-digit ИНН.
func ValidINN(s string) bool {
	s = digits(s)
	ctl := func(w []int) int {
		sum := 0
		for i, k := range w {
			sum += k * int(s[i]-'0')
		}
		return sum % 11 % 10
	}
	switch len(s) {
	case 10:
		return ctl([]int{2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(s[9]-'0')
	case 12:
		return ctl([]int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(s[10]-'0') &&
			ctl([]int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(s[11]-'0')
	}
	return false
}

// FirstDigitRun returns the first run of digits ("77313966110, 119034" →
// "77313966110"): OCR often glues the following postcode onto ИНН/ОГРН.
func FirstDigitRun(s string) string {
	start := -1
	for i, r := range s {
		if r >= '0' && r <= '9' {
			if start < 0 {
				start = i
			}
		} else if start >= 0 {
			return s[start:i]
		}
	}
	if start >= 0 {
		return s[start:]
	}
	return ""
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			c := 1
			if ar[i-1] == br[j-1] {
				c = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+c)
		}
		prev = cur
	}
	return prev[len(br)]
}
