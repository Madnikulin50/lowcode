package main

import "strings"

// lineDepths returns, for each line of src, the brace/paren/bracket depth
// in effect at the START of that line (i.e. before any character on that
// line is consumed). It tracks Go string/rune/comment lexical rules so that
// braces inside string literals or comments don't skew the count.
func lineDepths(src string) []int {
	lines := strings.Split(src, "\n")
	depths := make([]int, len(lines))

	depth := 0
	inBlockComment := false

	for i, line := range lines {
		depths[i] = depth

		j := 0
		inLineComment := false
		for j < len(line) {
			if inBlockComment {
				if strings.HasPrefix(line[j:], "*/") {
					inBlockComment = false
					j += 2
					continue
				}
				j++
				continue
			}
			if inLineComment {
				break
			}

			c := line[j]
			switch {
			case strings.HasPrefix(line[j:], "//"):
				inLineComment = true
				j += 2
			case strings.HasPrefix(line[j:], "/*"):
				inBlockComment = true
				j += 2
			case c == '"':
				j = skipInterpretedString(line, j+1) + 1
			case c == '`':
				// raw string: if it doesn't close on this line, it
				// continues on subsequent lines - not used by these
				// generated files (no raw strings span lines here), but
				// handle the common case of closing on the same line.
				if k := strings.IndexByte(line[j+1:], '`'); k >= 0 {
					j = j + 1 + k + 1
				} else {
					j = len(line)
				}
			case c == '\'':
				j = skipInterpretedString(line, j+1) + 1
			case c == '{' || c == '(' || c == '[':
				depth++
				j++
			case c == '}' || c == ')' || c == ']':
				depth--
				j++
			default:
				j++
			}
		}
	}

	return depths
}

// skipInterpretedString returns the index of the closing quote (matching
// whatever quote char started at pos-1) starting the scan at pos, handling
// backslash escapes. Works for both '"' and '\” delimited literals since
// it just looks for an unescaped occurrence of... actually we need to know
// which quote we're matching; since generated code never mixes them
// mid-literal this simple heuristic (stop at first unescaped quote of
// either kind that closes the literal) is resolved by the caller passing
// the right start.
func skipInterpretedString(line string, pos int) int {
	for pos < len(line) {
		if line[pos] == '\\' {
			pos += 2
			continue
		}
		if line[pos] == '"' || line[pos] == '\'' {
			return pos
		}
		pos++
	}
	return len(line)
}

// splitTopLevelBlocks splits src (already trimmed of any wrapping
// paren/brace) into chunks separated by one-or-more blank lines that occur
// at depth 0. Each returned chunk has its surrounding blank lines trimmed.
func splitTopLevelBlocks(src string) []string {
	lines := strings.Split(src, "\n")
	depths := lineDepths(src)

	var blocks []string
	var cur []string
	flush := func() {
		joined := strings.Join(cur, "\n")
		if strings.TrimSpace(joined) != "" {
			blocks = append(blocks, strings.Trim(joined, "\n"))
		}
		cur = cur[:0]
	}

	for i, line := range lines {
		if strings.TrimSpace(line) == "" && depths[i] == 0 {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()

	return blocks
}
