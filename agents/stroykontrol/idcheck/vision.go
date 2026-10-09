package idcheck

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Vision is an Ollama chat client for a multimodal model (qwen3.5 by
// default). Replies are constrained with a JSON schema (`format`) and cached
// on disk keyed by model + prompt + images, so re-running a package never
// re-asks the (slow, CPU-only) model about the same page.
type Vision struct {
	URL      string // http://localhost:11434
	Model    string
	CacheDir string
	Timeout  time.Duration
	http     *http.Client
}

// NewVision returns a client; an empty model disables vision extraction.
func NewVision(url, model, cacheDir string, timeout time.Duration) *Vision {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	return &Vision{URL: strings.TrimRight(url, "/"), Model: model, CacheDir: cacheDir, Timeout: timeout,
		http: &http.Client{Timeout: timeout}}
}

// Enabled reports whether a model is configured.
func (v *Vision) Enabled() bool { return v != nil && v.Model != "" }

type ollamaMsg struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

// Ask sends prompt + images and decodes the JSON reply into out.
func (v *Vision) Ask(ctx context.Context, prompt string, images []string, schema map[string]any, out any) error {
	h := sha256.New()
	h.Write([]byte(v.Model + "\x00" + prompt))
	var imgs []string
	for _, p := range images {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h.Write(b)
		imgs = append(imgs, base64.StdEncoding.EncodeToString(b))
	}
	sb, _ := json.Marshal(schema)
	h.Write(sb)
	cache := filepath.Join(v.CacheDir, "vision", hex.EncodeToString(h.Sum(nil))[:32]+".json")
	if b, err := os.ReadFile(cache); err == nil {
		return json.Unmarshal(b, out)
	}

	body, _ := json.Marshal(map[string]any{
		"model":    v.Model,
		"stream":   false,
		"think":    false,
		"format":   schema,
		"options":  map[string]any{"temperature": 0, "num_ctx": 16384},
		"messages": []ollamaMsg{{Role: "user", Content: prompt, Images: imgs}},
	})
	ctx, cancel := context.WithTimeout(ctx, v.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.URL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := v.http.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var r struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return fmt.Errorf("ollama: decode: %w", err)
	}
	content := strings.TrimSpace(r.Message.Content)
	if i := strings.Index(content, "{"); i > 0 {
		content = content[i:]
	}
	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("model reply is not the expected JSON: %w (%s)", err, truncate(content, 200))
	}
	_ = os.MkdirAll(filepath.Dir(cache), 0o755)
	_ = os.WriteFile(cache, []byte(content), 0o644)
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// --- JSON schemas ----------------------------------------------------------

func str() map[string]any { return map[string]any{"type": "string"} }
func num() map[string]any { return map[string]any{"type": []string{"number", "null"}} }
func arr(item map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": item}
}
func obj(props map[string]any) map[string]any {
	req := make([]string, 0, len(props))
	for k := range props {
		req = append(req, k)
	}
	sort.Strings(req) // map order is random; the schema feeds the cache key
	return map[string]any{"type": "object", "properties": props, "required": req}
}

var docRefSchema = obj(map[string]any{
	"kind": str(), "name": str(), "number": str(), "date": str(), "sheets": str(),
})

var actSchema = obj(map[string]any{
	"type": str(), "number": str(), "date": str(), "object": str(),
	"p1_work": str(), "p1_vol_project": num(), "p1_vol_fact": num(), "p1_unit": str(),
	"p2_text": str(), "p2_rd_code": str(),
	"p3_materials": arr(docRefSchema), "p4_docs": arr(docRefSchema),
	"p5_start": str(), "p5_end": str(),
	"p6_norms": arr(str()), "p6_rd_code": str(), "p7_next": str(),
	"appendix": arr(docRefSchema),
	"people": arr(obj(map[string]any{
		"role": str(), "fio": str(), "position": str(), "org": str(),
		"order_no": str(), "order_date": str(), "nrs_id": str(),
	})),
	"orgs": arr(obj(map[string]any{
		"role": str(), "name": str(), "ogrn": str(), "inn": str(), "address": str(),
		"sro": str(), "sro_ogrn": str(), "sro_inn": str(),
	})),
})

const actPrompt = `Это страницы акта освидетельствования скрытых работ (АОСР) или ответственных конструкций (АООК) по форме Приказа Минстроя № 344/пр. Извлеки данные в JSON строго по схеме. Переписывай значения ТОЧНО как напечатано/вписано в документе, ничего не додумывай; если поля нет — пустая строка, для чисел null. Даты в формате ДД.ММ.ГГГГ.
- type: "АОСР" или "АООК"; number: номер акта после «№»; date: дата акта (справа от номера).
- p1_work: текст п.1 «К освидетельствованию предъявлены следующие работы»; p1_vol_project / p1_vol_fact: числа Vпр / Vф из п.1 (если указаны), p1_unit: единица (м3, т, шт).
- p2_text: текст п.2 целиком; p2_rd_code: шифр проектной/рабочей документации из п.2.
- p3_materials: материалы из п.3 (name, документ о качестве — number/date если есть).
- p4_docs: документы из п.4 (kind: схема/акт/протокол/письмо/журнал/прочее, name, number, date).
- p5_start, p5_end: даты начала и окончания работ (п.5).
- p6_norms: каждый нормативный документ и проектный документ из п.6 отдельной строкой; p6_rd_code: шифр проекта/РД из п.6.
- p7_next: п.7.
- appendix: каждый пункт раздела «Приложения» (name, number, date, sheets — число листов).
- people: все представители из шапки акта: role (как в заголовке, напр. «Представитель застройщика … по вопросам строительного контроля»), fio, position (должность), org, order_no (реквизиты приказа/распорядительного документа целиком), order_date, nrs_id (идентификационный номер в НРС, напр. С-77-246659).
- orgs: все юрлица из шапки: role (застройщик/лицо, осуществляющее строительство/проектировщик/исполнитель работ), name, ogrn, inn, address, sro (наименование СРО), sro_ogrn, sro_inn.`

var schemeSchema = obj(map[string]any{
	"title": str(), "rd_code": str(), "vol_project": num(), "vol_fact": num(), "unit": str(), "date": str(),
})

const schemePrompt = `Это правый нижний угол исполнительной схемы (исполнительного чертежа): основная надпись (штамп) и, возможно, таблица объёмов. Извлеки в JSON: title — наименование схемы; rd_code — шифр (обозначение) из основной надписи (штампа) в правом нижнем углу: это верхняя широкая графа штампа, обычно вида «<номер договора>/РД-<объект>-<раздел>-<марка>» (например 03102022/РД-ИССО2.2-ОП-КЖ4.1), точно как напечатано, вместе со списком листов; НЕ путай с короткими регистрационными/инвентарными номерами в рамках над штампом; vol_project и vol_fact — проектный и фактический объём работ, если указаны на схеме или в таблице объёмов (числа, иначе null); unit — единица; date — дата из штампа. Ничего не додумывай.`

var attachedSchema = obj(map[string]any{
	"docs": arr(obj(map[string]any{
		"kind": str(), "name": str(), "number": str(), "date": str(), "material": str(), "valid_until": str(),
	})),
})

const attachedPrompt = `Это страница(ы), приложенные к акту освидетельствования: сертификаты, паспорта качества, декларации, протоколы испытаний, реестры документов о качестве, письма. Для КАЖДОГО документа на странице (для реестра — каждой строки реестра) верни: kind (сертификат/паспорт/декларация/протокол/реестр/письмо/прочее), name, number, date (ДД.ММ.ГГГГ), material (материал/изделие), valid_until (срок действия или годности, ДД.ММ.ГГГГ, если указан). Переписывай точно, ничего не додумывай.`

var classifySchema = obj(map[string]any{"kind": map[string]any{"type": "string", "enum": []string{
	string(PageAct), string(PageScheme), string(PageCertificate), string(PageProtocol), string(PageRegister), string(PageLetter), string(PageOther),
}}})

const classifyPrompt = `Определи тип отсканированной страницы из комплекта исполнительной документации по её виду и заголовку:
- act — страница самого акта освидетельствования (АОСР/АООК): сплошной текст с пунктами 1–7, «Представитель…», подписи;
- scheme — исполнительная схема/чертёж: графика, размеры, отметки, рамка с основной надписью (штампом) внизу справа;
- register — таблица «Реестр документов…», «Реестр исполнительной документации», перечень сертификатов;
- certificate — сертификат качества, паспорт качества/изделия, документ о качестве, декларация о соответствии (логотип завода, таблица химсостава/свойств, печать ОТК);
- protocol — протокол испытаний, акт отбора проб, заключение лаборатории;
- letter — письмо на бланке организации (исх. №, адресат, подпись);
- other — что-то иное.
Ответь одним значением kind.`
