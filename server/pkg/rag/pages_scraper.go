package rag

import (
	"html"
	"regexp"
	"strings"

	composeType "github.com/madnikulin50/lowcode/server/compose/types"
)

var (
	htmlTagRe2    = regexp.MustCompile(`<[^>]*>`)
	htmlScriptRe2 = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	htmlStyleRe2  = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	multSpaceRe2  = regexp.MustCompile(`\n{3,}`)
	multWSRe2     = regexp.MustCompile(`[ \t]+`)
)

type ScrapedPage struct {
	PageID      uint64
	NamespaceID uint64
	Title       string
	Text        string
}

type ScrapedBlock struct {
	PageID      uint64
	NamespaceID uint64
	PageTitle   string
	BlockKind   string
	BlockTitle  string
	Text        string
}

func ScrapePage(p *composeType.Page) *ScrapedPage {
	sp := &ScrapedPage{
		PageID:      p.ID,
		NamespaceID: p.NamespaceID,
		Title:       p.Title,
	}
	var parts []string
	if p.Title != "" {
		parts = append(parts, p.Title)
	}
	if p.Description != "" {
		parts = append(parts, p.Description)
	}
	for _, block := range p.Blocks {
		text := extractBlockText(block)
		if text != "" {
			parts = append(parts, text)
		}
	}
	sp.Text = strings.Join(parts, "\n\n")
	return sp
}

// ScrapePageBlocks returns each block as its own ScrapedBlock so the RAG
// can create a separate chunk per block.
func ScrapePageBlocks(p *composeType.Page) []ScrapedBlock {
	var blocks []ScrapedBlock
	for _, block := range p.Blocks {
		text := extractBlockText(block)
		if text == "" {
			continue
		}
		blocks = append(blocks, ScrapedBlock{
			PageID:      p.ID,
			NamespaceID: p.NamespaceID,
			PageTitle:   p.Title,
			BlockKind:   block.Kind,
			BlockTitle:  block.Title,
			Text:        text,
		})
	}
	return blocks
}

func extractBlockText(block composeType.PageBlock) string {
	return ExtractBlockText(block)
}

func ExtractBlockText(block composeType.PageBlock) string {
	var parts []string
	if block.Title != "" {
		parts = append(parts, block.Title)
	}
	if block.Description != "" {
		parts = append(parts, block.Description)
	}
	switch block.Kind {
	case "Content":
		if body, ok := block.Options["body"]; ok {
			if s, ok := body.(string); ok {
				parts = append(parts, stripHTML(s))
			}
		}
	case "Metric":
		if metrics, ok := block.Options["metrics"]; ok {
			if arr, ok := metrics.([]interface{}); ok {
				for _, m := range arr {
					if mm, ok := m.(map[string]interface{}); ok {
						for _, f := range []string{"label", "prefix", "suffix"} {
							if v, ok := mm[f]; ok {
								if s, ok := v.(string); ok && s != "" {
									parts = append(parts, s)
								}
							}
						}
					}
				}
			}
		}
	case "Automation", "RecordList":
		for _, key := range []string{"buttons", "selectionButtons"} {
			if btns, ok := block.Options[key]; ok {
				if arr, ok := btns.([]interface{}); ok {
					for _, b := range arr {
						if bb, ok := b.(map[string]interface{}); ok {
							if v, ok := bb["label"]; ok {
								if s, ok := v.(string); ok && s != "" {
									parts = append(parts, s)
								}
							}
						}
					}
				}
			}
		}
	case "IFrame":
		if url, ok := block.Options["url"]; ok {
			if s, ok := url.(string); ok && s != "" {
				parts = append(parts, "Embedded: "+s)
			}
		}
	case "File":
		if label, ok := block.Options["label"]; ok {
			if s, ok := label.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
	}
	for _, v := range block.Options {
		if s, ok := v.(string); ok && s != "" && len(s) > 3 {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n")
}

func stripHTML(s string) string {
	s = htmlScriptRe2.ReplaceAllString(s, "")
	s = htmlStyleRe2.ReplaceAllString(s, "")
	s = htmlTagRe2.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = multSpaceRe2.ReplaceAllString(s, "\n\n")
	s = multWSRe2.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
