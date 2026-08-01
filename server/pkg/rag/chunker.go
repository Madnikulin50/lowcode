package rag

import (
	"strings"
	"unicode"
)

const (
	defaultChunkSize    = 512
	defaultChunkOverlap = 64
)

func ChunkText(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}
	if overlap < 0 {
		overlap = defaultChunkOverlap
	}

	paragraphs := splitParagraphs(text)
	var chunks []string
	var current strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if current.Len()+len(para)+1 > chunkSize && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()

			// Add overlap from the previous chunk
			if overlap > 0 && len(chunks) > 0 {
				prev := chunks[len(chunks)-1]
				words := strings.Fields(prev)
				if len(words) > overlap/5 {
					// Take last ~overlap characters worth of words
					overlapText := ""
					for i := len(words) - 1; i >= 0; i-- {
						if len(overlapText)+len(words[i])+1 > overlap {
							break
						}
						if overlapText != "" {
							overlapText = words[i] + " " + overlapText
						} else {
							overlapText = words[i]
						}
					}
					current.WriteString(overlapText)
					current.WriteString(" ")
				}
			}
		}

		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(para)
	}

	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}

	if len(chunks) == 0 && text != "" {
		// If no paragraphs found, split by size
		runes := []rune(text)
		for i := 0; i < len(runes); i += chunkSize - overlap {
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			chunk := string(runes[i:end])
			chunks = append(chunks, strings.TrimSpace(chunk))
		}
	}

	return chunks
}

func splitParagraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var paragraphs []string
	current := strings.Builder{}

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current.Len() > 0 {
				paragraphs = append(paragraphs, strings.TrimSpace(current.String()))
				current.Reset()
			}
			continue
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(trimmed)
	}
	if current.Len() > 0 {
		paragraphs = append(paragraphs, strings.TrimSpace(current.String()))
	}

	// Also split long paragraphs by sentences
	var result []string
	for _, p := range paragraphs {
		if len([]rune(p)) <= defaultChunkSize*2 {
			result = append(result, p)
			continue
		}
		sentences := splitSentences(p)
		result = append(result, sentences...)
	}
	return result
}

func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			// Check if it's a real sentence end (followed by space or capital letter)
			currentStr := current.String()
			if len(currentStr) > 10 && !isAbbreviation(currentStr) {
				sentences = append(sentences, strings.TrimSpace(currentStr))
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}
	return sentences
}

func isAbbreviation(s string) bool {
	lastWord := s
	if idx := strings.LastIndex(s, " "); idx >= 0 {
		lastWord = s[idx+1:]
	}
	lastWord = strings.TrimRight(lastWord, ".!?")
	abbrevs := []string{"Mr", "Mrs", "Ms", "Dr", "Prof", "Sr", "Jr", "etc", "e.g", "i.e", "vs", "St"}
	for _, ab := range abbrevs {
		if lastWord == ab {
			return true
		}
	}
	// Check single capital letter (like "A." or "B.")
	if len(lastWord) == 1 && unicode.IsUpper([]rune(lastWord)[0]) {
		return true
	}
	// Check numbered lists like "1." or "1.1"
	if len(lastWord) > 0 {
		allDigits := true
		for _, r := range lastWord {
			if !unicode.IsDigit(r) && r != '.' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return true
		}
	}
	return false
}
