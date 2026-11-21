// Package idcheck checks primary as-built documentation (АОСР/АООК acts)
// against the ИД transfer registry (реестр передачи ИД), the КС-2 form and
// regulatory requirements. Parsing and checks are pure Go; OCR (tesseract)
// and structured extraction (Ollama vision model) sit behind small
// interfaces so the checks themselves stay deterministic and unit-testable.
package idcheck

import "time"

// RowKind classifies a registry row.
type RowKind string

const (
	RowSection  RowKind = "section"  // e.g. "2.14 (ИССО 2.2) Путепровод ..."
	RowPosition RowKind = "position" // ВРЦ position with unit/quantities
	RowAct      RowKind = "act"      // a document (АОСР/АООК) under a position
)

// RegistryRow is one meaningful row of the registry sheet. Quantity fields
// are nil when the cell is empty — "no volume" and "zero volume" differ.
type RegistryRow struct {
	Row        int // 1-based spreadsheet row
	Seq        string
	Kind       RowKind
	Position   string
	WorkName   string
	DocName    string
	DocNumber  string
	DocDate    time.Time
	DocDateRaw string
	Unit       string
	QtyVRC     *float64 // col 9
	QtyTotal   *float64 // col 10
	QtyKS2     *float64 // col 11
	QtyProject *float64 // col 12
	QtyFact    *float64 // col 13
}

// Registry is a parsed registry workbook.
type Registry struct {
	Title    string
	Period   string
	KS2Ref   string
	Customer string
	Rows     []RegistryRow
}

// KS2Row is one priced line of a КС-2 act.
type KS2Row struct {
	Row      int
	Seq      string
	Position string
	Name     string
	Unit     string
	Qty      float64
	Price    float64
	Amount   float64
}

// KS2 is a parsed КС-2 workbook.
type KS2 struct {
	Number     string
	Date       time.Time
	PeriodFrom time.Time
	PeriodTo   time.Time
	Rows       []KS2Row
}

// PageKind classifies a scanned page of an act PDF.
type PageKind string

const (
	PageAct         PageKind = "act"
	PageScheme      PageKind = "scheme"      // исполнительная схема
	PageCertificate PageKind = "certificate" // сертификат / паспорт / декларация
	PageProtocol    PageKind = "protocol"    // протокол испытаний
	PageRegister    PageKind = "register"    // реестр документов о качестве
	PageLetter      PageKind = "letter"
	PageOther       PageKind = "other"
)

// Page is one page of an act PDF after OCR/classification.
type Page struct {
	N    int      `json:"n"`
	Kind PageKind `json:"kind"`
	Text string   `json:"-"`
}

// Person is a representative listed in the act header.
type Person struct {
	Role      string `json:"role"`
	FIO       string `json:"fio"`
	Position  string `json:"position"`
	Org       string `json:"org"`
	OrderNo   string `json:"order_no"`
	OrderDate string `json:"order_date"`
	NRSID     string `json:"nrs_id"`
}

// Org is a legal entity listed in the act header.
type Org struct {
	Role    string `json:"role"`
	Name    string `json:"name"`
	OGRN    string `json:"ogrn"`
	INN     string `json:"inn"`
	Address string `json:"address"`
	SRO     string `json:"sro"`
	SROOGRN string `json:"sro_ogrn"`
	SROINN  string `json:"sro_inn"`
}

// DocRef is a document referenced by an act (п.3, п.4, приложения) or found
// as an attached scan.
type DocRef struct {
	Kind       string `json:"kind"` // схема / сертификат / паспорт / протокол / письмо / реестр / журнал / акт
	Name       string `json:"name"`
	Number     string `json:"number"`
	Date       string `json:"date"`
	Material   string `json:"material,omitempty"`
	ValidUntil string `json:"valid_until,omitempty"`
	Sheets     string `json:"sheets,omitempty"`
	Page       int    `json:"page,omitempty"`
}

// ActData is what extraction produced for one act. Dates are kept as the
// raw strings the document carries; checks parse them with ParseDate.
type ActData struct {
	Type          string   `json:"type"` // АОСР / АООК
	Number        string   `json:"number"`
	Date          string   `json:"date"`
	Object        string   `json:"object"`
	P1Work        string   `json:"p1_work"`
	P1VolProject  *float64 `json:"p1_vol_project"`
	P1VolFact     *float64 `json:"p1_vol_fact"`
	P1Unit        string   `json:"p1_unit"`
	P2RDCode      string   `json:"p2_rd_code"`
	P2Text        string   `json:"p2_text"`
	P3Materials   []DocRef `json:"p3_materials"`
	P4Docs        []DocRef `json:"p4_docs"`
	P5Start       string   `json:"p5_start"`
	P5End         string   `json:"p5_end"`
	P6Norms       []string `json:"p6_norms"`
	P6RDCode      string   `json:"p6_rd_code"`
	P7Next        string   `json:"p7_next"`
	Appendix      []DocRef `json:"appendix"`
	People        []Person `json:"people"`
	Orgs          []Org    `json:"orgs"`
	SchemeRDCode  string   `json:"scheme_rd_code"`
	SchemeVolProj *float64 `json:"scheme_vol_project"`
	SchemeVolFact *float64 `json:"scheme_vol_fact"`
	SchemePage    int      `json:"scheme_page"`
	Attached      []DocRef `json:"attached"` // documents actually scanned into the PDF
}

// ActFile is one act PDF as delivered (file name carries number/date too).
type ActFile struct {
	Name string
	Data []byte
}

// Severity of a finding.
type Severity string

const (
	High   Severity = "high"
	Medium Severity = "medium"
	Low    Severity = "low"
	Info   Severity = "info"
)

// Finding is one discrepancy. Check is the ТЗ check number (1–8).
type Finding struct {
	Check       int      `json:"check"`
	Severity    Severity `json:"severity"`
	Field       string   `json:"field"`
	Expected    string   `json:"expected"`
	Actual      string   `json:"actual"`
	Description string   `json:"description"`
	Page        int      `json:"page,omitempty"`
	Box         *Box     `json:"box,omitempty"` // where on Page (relative coords), if located
	Source      string   `json:"source"`        // det / ai / ext
	ActKey      string   `json:"act_key,omitempty"`
	RegistryRow int      `json:"registry_row,omitempty"`
	Position    string   `json:"position,omitempty"`
}

// CheckNames are the user-facing titles of the ТЗ checks.
var CheckNames = map[int]string{
	1: "Акт ↔ реестр (номер, дата, объёмы кол. 12/13)",
	2: "Приложения ↔ п.3/п.4 ↔ сканы, сроки годности",
	3: "Шифр РД: п.2 ↔ п.6 ↔ исп. схема",
	4: "Объём п.1 ↔ исполнительная схема",
	5: "Нормативные документы п.6",
	6: "Представители: приказы, НРС",
	7: "Реквизиты юрлиц ↔ ЕГРЮЛ",
	8: "Реестр ↔ КС-2 по позициям",
}

// PositionSummary is the result of check 8 for one ВРЦ position.
type PositionSummary struct {
	Position    string
	Name        string
	Unit        string
	RegistrySum float64  // Σ min(col12, col13) over act rows
	RegistryKS2 *float64 // col 11 on the position row
	KS2Qty      *float64
	Delta       float64
	Status      string // ok / mismatch / missing_ks2 / missing_registry
}
