package rag

import (
	"strings"
	"unicode"
)

func harvestDocument(data []byte, kind string, partial bool) (*ParsedDocument, error) {
	ascii := harvestASCII(data)
	u16 := harvestUTF16LE(data)
	text := ascii
	if len(u16) > len(ascii) {
		text = u16
	}
	text = strings.TrimSpace(text)
	if kind == "" {
		kind = "harvest"
	}
	return &ParsedDocument{
		Text:    text,
		Title:   extractTitle(text),
		Kind:    kind,
		Partial: partial || text == "",
	}, nil
}

func harvestASCII(data []byte) string {
	var b strings.Builder
	var run []byte
	flush := func() {
		if len(run) >= 6 && mostlyLetters(run) {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.Write(run)
		}
		run = run[:0]
	}
	for _, c := range data {
		if c >= 32 && c < 127 {
			run = append(run, c)
			continue
		}
		flush()
	}
	flush()
	return b.String()
}

func harvestUTF16LE(data []byte) string {
	var b strings.Builder
	var run []byte
	flush := func() {
		if len(run) >= 6 && mostlyLetters(run) {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.Write(run)
		}
		run = run[:0]
	}
	for i := 0; i+1 < len(data); i += 2 {
		lo, hi := data[i], data[i+1]
		if hi == 0 && lo >= 32 && lo < 127 {
			run = append(run, lo)
			continue
		}
		flush()
		if hi != 0 {
			i--
		}
	}
	flush()
	return b.String()
}

func mostlyLetters(run []byte) bool {
	letters := 0
	for _, c := range run {
		if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) {
			letters++
		}
	}
	return letters*2 >= len(run)
}
