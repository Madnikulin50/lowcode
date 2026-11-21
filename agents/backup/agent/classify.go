package agent

import (
	"path/filepath"
	"strings"
	"time"
)

// Критичность находки аудита. Совпадает с опциями Select поля criticality
// модуля audit_findings.
const (
	CritLow      = "low"
	CritMedium   = "medium"
	CritHigh     = "high"
	CritCritical = "critical"
)

// fileAssessment — оценка критичности одного файла по правилам
// (имя/каталог/расширение/возраст), без чтения содержимого.
type fileAssessment struct {
	Class       string // DATA_CLASS: public/internal/confidential/personal_special
	Criticality string
	Reasons     []string
}

var critRank = map[string]int{
	CritLow: 1, CritMedium: 2, CritHigh: 3, CritCritical: 4,
}

var rankCrit = map[int]string{
	1: CritLow, 2: CritMedium, 3: CritHigh, 4: CritCritical,
}

// personalMarkers — маркеры ПДн в имени файла/каталога.
var personalMarkers = []string{
	"персональн", "медицин", "снилс", "паспорт", "инн", "здоровь", "диагноз",
	"карта пациент", "полис", "трудовой", "зарплат", "personal", "medical", "patient",
}

// personalDirMarkers — маркеры каталогов, где ПДн ожидаемы даже без
// маркеров в имени файла.
var personalDirMarkers = []string{
	"кадры", "hr", "персонал", "бухгалт", "зарплат", "1с", "1c",
	"мед", "клиенты", "clients", "договоры",
}

// financeMarkers — финансовые/юридические документы.
var financeMarkers = []string{
	"договор", "счет", "счёт", "акт", "накладная", "счет-фактура", "платеж",
	"платёж", "invoice", "contract", "спецификац", "закуп",
}

// keyExts — ключевой материал и секреты.
var keyExts = map[string]bool{
	"pem": true, "key": true, "p12": true, "pfx": true, "ppk": true,
	"kdbx": true, "gpg": true, "asc": true, "env": true, "keystore": true,
	"jks": true, "p7b": true, "der": true, "crt": true, "cer": true,
}

// dbExts — базы и дампы: потеря почти всегда критична для бизнеса.
var dbExts = map[string]bool{
	"1cd": true, "bak": true, "sql": true, "dump": true, "pgdump": true,
	"mdb": true, "dbf": true, "mdf": true, "ldf": true, "sqlite": true,
	"db": true, "ib": true, "fdb": true, "myd": true, "frm": true,
}

// officeExts — документы, где живут договоры/отчёты/регистры.
var officeExts = map[string]bool{
	"doc": true, "docx": true, "xls": true, "xlsx": true, "pdf": true,
	"odt": true, "ods": true, "ppt": true, "pptx": true, "rtf": true,
	"csv": true, "tsv": true, "epub": true,
}

// archiveExts — архивы: неизвестно что внутри, не ниже среднего.
var archiveExts = map[string]bool{
	"zip": true, "7z": true, "rar": true, "tar": true, "gz": true,
	"tgz": true, "bz2": true, "xz": true, "iso": true,
}

// codeExts — код и конфиги без явных секретов.
var codeExts = map[string]bool{
	"go": true, "js": true, "ts": true, "py": true, "java": true, "php": true,
	"rb": true, "c": true, "cc": true, "cpp": true, "h": true, "hpp": true,
	"sh": true, "bat": true, "ps1": true, "conf": true, "cfg": true,
	"ini": true, "yml": true, "yaml": true, "json": true, "xml": true,
	"html": true, "htm": true, "css": true, "sql": true, "ipynb": true,
}

const (
	// активный файл изменялся недавно — критичность выше.
	activeWindow = 30 * 24 * time.Hour
	// давно не трогали — потеря переносится легче.
	staleWindow = 365 * 24 * time.Hour
)

// assessFile оценивает файл по имени, каталогам, расширению и возрасту.
// Контентное сканирование (scanFilePII) при необходимости эскалирует
// результат до personal_special/critical.
func assessFile(rel string, size int64, mod time.Time, now time.Time) fileAssessment {
	lower := strings.ToLower(rel)
	base := filepath.Base(lower)
	ext := strings.TrimPrefix(filepath.Ext(base), ".")
	score := 1
	class := "internal"
	var reasons []string

	raise := func(s int, r string) {
		if s > score {
			score = s
		}
		reasons = append(reasons, r)
	}
	setClass := func(c string) {
		if class == "personal_special" {
			return
		}
		class = c
	}

	for _, m := range personalMarkers {
		if strings.Contains(base, m) {
			score = critRank[CritCritical]
			setClass("personal_special")
			reasons = append(reasons, "имя:"+m)
			break
		}
	}
	for _, seg := range strings.Split(filepath.ToSlash(lower), "/") {
		for _, m := range personalDirMarkers {
			if seg == m || strings.HasPrefix(seg, m+" ") || strings.HasSuffix(seg, " "+m) || strings.HasPrefix(seg, m+"-") || strings.HasPrefix(seg, m+"_") {
				if score < critRank[CritHigh] {
					score = critRank[CritHigh]
				}
				setClass("personal_special")
				reasons = append(reasons, "каталог:"+seg)
			}
		}
	}
	switch {
	case keyExts[ext]:
		score = critRank[CritCritical]
		setClass("confidential")
		reasons = append(reasons, "ключи/сертификаты")
	case dbExts[ext]:
		raise(critRank[CritHigh], "база/дамп")
		setClass("confidential")
	case officeExts[ext]:
		raise(critRank[CritMedium], "документ")
		setClass("confidential")
	case archiveExts[ext]:
		raise(critRank[CritMedium], "архив")
	case codeExts[ext]:
		// код сам по себе не критичен
	case ext == "":
		// расширения нет — неизвестность, чуть выше низкой
		raise(2, "без расширения")
	}
	for _, m := range financeMarkers {
		if strings.Contains(base, m) {
			raise(critRank[CritMedium], "фин:"+m)
			setClass("confidential")
			break
		}
	}
	// Возраст: активные данные болезненнее терять, залежавшееся — легче.
	if age := now.Sub(mod); age >= 0 {
		switch {
		case age <= activeWindow:
			if score < critRank[CritCritical] {
				score++
			}
			reasons = append(reasons, "активный")
		case age >= staleWindow:
			if score > 1 {
				score--
			}
			reasons = append(reasons, "давно не изменялся")
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "нет явных признаков")
	}
	return fileAssessment{Class: class, Criticality: rankCrit[score], Reasons: reasons}
}

// escalatePII поднимает оценку до критичной по результатам контент-скана.
func (fa *fileAssessment) escalatePII(hits []string) {
	fa.Class = "personal_special"
	fa.Criticality = CritCritical
	fa.Reasons = append(fa.Reasons, "содержимое:"+strings.Join(hits, "+"))
}
