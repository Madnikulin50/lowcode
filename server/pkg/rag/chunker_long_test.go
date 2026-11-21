package rag

import (
	"strings"
	"testing"
)

func TestChunkTextNoOversizedChunks(t *testing.T) {
	text := strings.Repeat("слово ", 3000) + "\n\n" + strings.Repeat("x", 5000)
	for i, c := range ChunkText(text, 512, 64) {
		if n := len([]rune(c)); n > 512+64+1 {
			t.Fatalf("chunk %d has %d runes", i, n)
		}
	}
}
