package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

type ParsedDocument struct {
	Text  string
	Title string
}

func ParseDocument(data []byte, filename, mimetype string) (*ParsedDocument, error) {
	ext := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(ext, ".txt") || mimetype == "text/plain":
		return parseText(data)
	case strings.HasSuffix(ext, ".html") || strings.HasSuffix(ext, ".htm") || mimetype == "text/html":
		return parseHTML(data)
	case strings.HasSuffix(ext, ".docx") || mimetype == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return parseDocx(data)
	case strings.HasSuffix(ext, ".pdf") || mimetype == "application/pdf":
		return parsePDF(data)
	default:
		// Try as plain text
		return parseText(data)
	}
}

func parseText(data []byte) (*ParsedDocument, error) {
	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	title := extractTitle(text)
	return &ParsedDocument{Text: text, Title: title}, nil
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)
var htmlScriptRe = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
var htmlStyleRe = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
var htmlTitleRe = regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
var multSpaceRe = regexp.MustCompile(`\n{3,}`)

func parseHTML(data []byte) (*ParsedDocument, error) {
	text := string(data)
	title := ""
	if m := htmlTitleRe.FindStringSubmatch(text); len(m) > 1 {
		title = strings.TrimSpace(m[1])
	}
	text = htmlScriptRe.ReplaceAllString(text, "")
	text = htmlStyleRe.ReplaceAllString(text, "")
	text = htmlTagRe.ReplaceAllString(text, " ")
	text = decodeHTMLEntities(text)
	text = multSpaceRe.ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)
	return &ParsedDocument{Text: text, Title: title}, nil
}

func decodeHTMLEntities(s string) string {
	repl := strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&apos;", "'",
		"&#39;", "'",
		"&mdash;", "—",
		"&ndash;", "–",
		"&hellip;", "…",
	)
	return repl.Replace(s)
}

type docxBody struct {
	Paragraphs []docxParagraph `xml:"p"`
}

type docxParagraph struct {
	Runs []docxRun `xml:"r"`
}

type docxRun struct {
	Text string `xml:"t"`
}

func parseDocx(data []byte) (*ParsedDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("parse docx: %w", err)
	}
	var docXML []byte
	for _, f := range reader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("parse docx: open document.xml: %w", err)
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("parse docx: read document.xml: %w", err)
			}
			break
		}
	}
	if docXML == nil {
		return nil, fmt.Errorf("parse docx: word/document.xml not found")
	}
	var body docxBody
	if err := xml.Unmarshal(docXML, &body); err != nil {
		// Try to extract text directly from XML
		text := extractDocxTextSimple(docXML)
		return &ParsedDocument{Text: text}, nil
	}
	var paragraphs []string
	for _, p := range body.Paragraphs {
		var line strings.Builder
		for _, r := range p.Runs {
			line.WriteString(r.Text)
		}
		if s := strings.TrimSpace(line.String()); s != "" {
			paragraphs = append(paragraphs, s)
		}
	}
	text := strings.Join(paragraphs, "\n")
	title := extractTitle(text)
	return &ParsedDocument{Text: text, Title: title}, nil
}

var docxTextRe = regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)

func extractDocxTextSimple(xmlData []byte) string {
	var parts []string
	matches := docxTextRe.FindAllSubmatch(xmlData, -1)
	for _, m := range matches {
		if len(m) > 1 {
			parts = append(parts, string(m[1]))
		}
	}
	return strings.Join(parts, "")
}

func parsePDF(data []byte) (*ParsedDocument, error) {
	reader := bytes.NewReader(data)
	size := reader.Size()
	pdfReader, err := pdf.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("parse pdf: %w", err)
	}
	var text strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content := page.Content()
		for _, t := range content.Text {
			text.WriteString(t.S)
			text.WriteString(" ")
		}
		text.WriteString("\n")
	}
	result := strings.TrimSpace(text.String())
	title := extractTitle(result)
	return &ParsedDocument{Text: result, Title: title}, nil
}

var titleRe = regexp.MustCompile(`(?m)^(.{1,100})$`)

func extractTitle(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 5 && len(line) <= 100 {
			return line
		}
	}
	return ""
}
