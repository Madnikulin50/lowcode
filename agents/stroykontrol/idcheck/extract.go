package idcheck

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"regexp"
	"strings"
)

// cropBottomRight writes the bottom-right (w×h fraction) part of a page
// image next to it and returns its path.
func cropBottomRight(img string, wf, hf float64) (string, error) {
	out := strings.TrimSuffix(img, ".jpg") + "-stamp.jpg"
	if _, err := os.Stat(out); err == nil {
		return out, nil
	}
	f, err := os.Open(img)
	if err != nil {
		return "", err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	b := src.Bounds()
	r := image.Rect(b.Max.X-int(float64(b.Dx())*wf), b.Max.Y-int(float64(b.Dy())*hf), b.Max.X, b.Max.Y)
	sub, ok := src.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return "", fmt.Errorf("image type cannot be cropped")
	}
	w, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer w.Close()
	return out, jpeg.Encode(w, sub.SubImage(r), &jpeg.Options{Quality: 90})
}

func landscape(img string) bool {
	f, err := os.Open(img)
	if err != nil {
		return false
	}
	defer f.Close()
	c, _, err := image.DecodeConfig(f)
	return err == nil && c.Width > c.Height
}

// Extractor turns an act PDF into ActData: rasterize → (tesseract OCR) →
// classify pages → vision model on the act pages, the first execution
// scheme and up to MaxAttached attached quality documents.
type Extractor struct {
	CacheDir    string
	DPI         int
	OCR         Tesseract
	Vision      *Vision
	MaxAttached int // attached pages sent to the vision model per act
	// ClassifyPages asks the vision model for each page's kind when tesseract
	// is unavailable. Off by default: on CPU it costs ~45 s/page and layout
	// rules (act → landscape scheme → attachments) work as well in practice.
	ClassifyPages bool
}

// ExtractResult is ActData plus page classification.
type ExtractResult struct {
	Data     *ActData
	Pages    []Page
	Notes    []string
	Words    map[int][]Word // OCR words per page (1-based), for locating findings
	ActPages []int          // 1-based pages of the act itself
}

// Extract runs the pipeline for one act. progress (optional) receives
// short human-readable steps.
func (e *Extractor) Extract(ctx context.Context, file ActFile, progress func(string)) (*ExtractResult, error) {
	say := func(f string, a ...any) {
		if progress != nil {
			progress(fmt.Sprintf(f, a...))
		}
	}
	dpi := e.DPI
	if dpi <= 0 {
		dpi = 150
	}
	imgs, err := PageImages(ctx, e.CacheDir, file.Data, dpi)
	if err != nil {
		return nil, err
	}
	res := &ExtractResult{}
	pages := make([]Page, len(imgs))
	for i, img := range imgs {
		pages[i] = Page{N: i + 1, Kind: PageOther}
		if e.OCR.Available() {
			say("OCR стр. %d/%d", i+1, len(imgs))
			txt, err := e.OCR.Text(ctx, img)
			if err != nil {
				res.Notes = append(res.Notes, err.Error())
			}
			pages[i].Text = txt
			pages[i].Kind = ClassifyText(txt)
			if ws, err := e.OCR.Words(ctx, img); err == nil {
				if res.Words == nil {
					res.Words = map[int][]Word{}
				}
				res.Words[i+1] = ws
			}
		}
	}
	if !e.OCR.Available() && e.Vision.Enabled() && e.ClassifyPages {
		// no OCR: classify with the vision model on low-res images (short
		// replies, so this is the cheap part of the pipeline)
		small, err := PageImages(ctx, e.CacheDir, file.Data, 75)
		if err == nil {
			for i := range pages {
				if i < len(small) {
					say("классификация стр. %d/%d", i+1, len(pages))
					var c struct {
						Kind string `json:"kind"`
					}
					if err := e.Vision.Ask(ctx, classifyPrompt, []string{small[i]}, classifySchema, &c); err == nil {
						pages[i].Kind = PageKind(c.Kind)
					} else {
						res.Notes = append(res.Notes, fmt.Sprintf("стр. %d: %v", i+1, err))
					}
				}
			}
		}
	}
	// The act itself always starts on page 1; continuation pages follow
	// until the first page of another kind.
	pages[0].Kind = PageAct
	actPages := []int{0}
	for i := 1; i < len(pages) && i < 4 && pages[i].Kind == PageAct; i++ {
		actPages = append(actPages, i)
	}
	if len(actPages) == 1 && len(pages) > 1 && pages[1].Kind == PageOther {
		pages[1].Kind = PageAct
		actPages = append(actPages, 1)
	}
	// No scheme recognized: the first landscape page after the act is the
	// execution scheme in practice (drawings are A3/A4 landscape).
	hasScheme := false
	for _, p := range pages {
		if p.Kind == PageScheme {
			hasScheme = true
		}
	}
	if !hasScheme {
		for i := actPages[len(actPages)-1] + 1; i < len(pages); i++ {
			if landscape(imgs[i]) {
				pages[i].Kind = PageScheme
				break
			}
		}
	}
	res.Pages = pages
	for _, i := range actPages {
		res.ActPages = append(res.ActPages, i+1)
	}

	data := &ActData{}
	if !e.Vision.Enabled() {
		res.Data = data
		res.Notes = append(res.Notes, "модель распознавания не настроена — проверяются только реестр и имя файла")
		return res, nil
	}

	say("распознавание акта (стр. %s)", pageList(actPages))
	var actImgs []string
	hint := ""
	for _, i := range actPages {
		actImgs = append(actImgs, imgs[i])
		if pages[i].Text != "" {
			hint += fmt.Sprintf("\n--- OCR стр. %d ---\n%s", i+1, truncate(pages[i].Text, 6000))
		}
	}
	prompt := actPrompt
	if hint != "" {
		prompt += "\n\nТекст, распознанный OCR (может содержать ошибки, сверяй с изображением):" + hint
	}
	if err := e.Vision.Ask(ctx, prompt, actImgs, actSchema, data); err != nil {
		return nil, fmt.Errorf("распознавание акта: %w", err)
	}
	postProcess(data)
	var actText strings.Builder
	for _, i := range actPages {
		actText.WriteString(pages[i].Text)
		actText.WriteString("\n")
	}
	fillFromOCR(data, actText.String())

	for i, p := range pages {
		if p.Kind != PageScheme {
			continue
		}
		say("исполнительная схема (стр. %d)", i+1)
		var s struct {
			Title      string   `json:"title"`
			RDCode     string   `json:"rd_code"`
			VolProject *float64 `json:"vol_project"`
			VolFact    *float64 `json:"vol_fact"`
		}
		// only the title block corner: the stamp and the volume table sit in
		// the bottom-right, and a full A3 landscape scan takes the CPU model
		// longer than the timeout
		img := imgs[i]
		if c, err := cropBottomRight(imgs[i], 0.55, 0.5); err == nil {
			img = c
		}
		if err := e.Vision.Ask(ctx, schemePrompt, []string{img}, schemeSchema, &s); err != nil {
			res.Notes = append(res.Notes, fmt.Sprintf("схема стр. %d: %v", i+1, err))
			continue
		}
		data.SchemeRDCode, data.SchemeVolProj, data.SchemeVolFact, data.SchemePage = s.RDCode, s.VolProject, s.VolFact, i+1
		data.Attached = append(data.Attached, DocRef{Kind: "схема", Name: s.Title, Page: i + 1})
		break
	}

	sent := 0
	for i, p := range pages {
		// everything bound into the act file after the act itself is an
		// attachment, whatever the classifier called it
		if p.Kind == PageAct || p.Kind == PageScheme {
			continue
		}
		if e.MaxAttached >= 0 && sent >= e.MaxAttached {
			res.Notes = append(res.Notes, fmt.Sprintf("стр. %d (%s) не распознавалась: лимит %d вложенных страниц на акт", i+1, p.Kind, e.MaxAttached))
			continue
		}
		sent++
		say("вложение стр. %d (%s)", i+1, p.Kind)
		var a struct {
			Docs []DocRef `json:"docs"`
		}
		if err := e.Vision.Ask(ctx, attachedPrompt, []string{imgs[i]}, attachedSchema, &a); err != nil {
			res.Notes = append(res.Notes, fmt.Sprintf("стр. %d: %v", i+1, err))
			continue
		}
		for _, d := range a.Docs {
			d.Page = i + 1
			if d.Kind == "" {
				d.Kind = kindName(p.Kind)
			}
			data.Attached = append(data.Attached, d)
		}
	}
	res.Data = data
	log.Printf("idcheck: %s: %d pages, act pages %v, %d attached docs", file.Name, len(pages), actPages, len(data.Attached))
	return res, nil
}

func kindName(k PageKind) string {
	switch k {
	case PageCertificate:
		return "сертификат"
	case PageProtocol:
		return "протокол"
	case PageRegister:
		return "реестр"
	case PageLetter:
		return "письмо"
	case PageScheme:
		return "схема"
	}
	return "прочее"
}

func pageList(idx []int) string {
	var s []string
	for _, i := range idx {
		s = append(s, fmt.Sprint(i+1))
	}
	return strings.Join(s, ",")
}

var (
	// volume or mass: "Vпр=26,3 м3 / Vф=26,3 м3", "Мрд=3993,4 кг./ Мф=3993,4 кг."
	vProjRe = regexp.MustCompile(`(?i)(?:V|М|M)\s*(?:пр|рд|п)\.?\s*=\s*([0-9]+(?:[.,][0-9]+)?)`)
	vFactRe = regexp.MustCompile(`(?i)(?:V|М|M)\s*ф(?:акт)?\.?\s*=\s*([0-9]+(?:[.,][0-9]+)?)`)
	unitRe  = regexp.MustCompile(`(?i)=\s*[0-9.,]+\s*(м3|м³|м2|м²|кг|тн|т|шт|п\.?м|м)(?:[^а-яa-z0-9]|$)`)
)

// postProcess fills gaps the model left using deterministic parsing of the
// text it did return (п.1 volumes are almost always "Vпр=26,3 м3 / Vф=26,3 м3").
func postProcess(d *ActData) {
	if d.P1VolProject == nil {
		if m := vProjRe.FindStringSubmatch(d.P1Work); m != nil {
			d.P1VolProject = numPtr(m[1])
		}
	}
	if d.P1VolFact == nil {
		if m := vFactRe.FindStringSubmatch(d.P1Work); m != nil {
			d.P1VolFact = numPtr(m[1])
		}
	}
	if d.P1Unit == "" {
		if m := unitRe.FindStringSubmatch(d.P1Work); m != nil {
			d.P1Unit = strings.ReplaceAll(m[1], "³", "3")
		}
	}
	if d.P2RDCode == "" {
		if c := FindRDCodes(d.P2Text); len(c) > 0 {
			d.P2RDCode = c[0]
		}
	}
	for i := range d.People {
		p := &d.People[i]
		if p.NRSID == "" {
			// the model sometimes files the NRS number under the order
			if n := NormalizeNRS(p.OrderNo); n != "" {
				p.NRSID = n
			}
		}
		if p.OrderDate == "" {
			if m := ruleDateRe.FindStringSubmatch(p.OrderNo); m != nil {
				p.OrderDate = m[1]
			}
		}
	}
}

// the order number is often garbled ("№1" → "Л") and separated from
// "приказ" by two lines of small-print captions; the date survives
var orderNearRe = regexp.MustCompile(`(?is)приказ.{0,500}?(?:№\s*)?(\S{1,12})\s+от\s+([0-9]{1,2}\.[0-9]{1,2}\.[0-9]{2,4})`)

// fillFromOCR fills a representative's order the model missed using the
// OCR text right after their surname: the act prints "Булатов В.Н.,
// идентификационный № С-77-246659, приказ" and "№1 от 10.01.2024г." on the
// next line, which the model tends to drop.
func fillFromOCR(d *ActData, text string) {
	if text == "" {
		return
	}
	for i := range d.People {
		p := &d.People[i]
		if _, ok := ParseDate(p.OrderDate); ok {
			continue
		}
		surname := strings.Fields(strings.TrimSpace(p.FIO))
		if len(surname) == 0 || len([]rune(surname[0])) < 3 {
			continue
		}
		at := strings.Index(text, surname[0])
		if at < 0 {
			continue
		}
		window := text[at:min(len(text), at+400)]
		if m := orderNearRe.FindStringSubmatch(window); m != nil {
			p.OrderNo = "приказ №" + strings.TrimPrefix(m[1], "№") + " от " + m[2] + " (по тексту скана)"
			p.OrderDate = m[2]
		}
	}
}
