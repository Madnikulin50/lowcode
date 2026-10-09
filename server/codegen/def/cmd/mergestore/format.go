package main

import "go/format"

// formatGo runs the merged source through the same formatter
// codegen/tool's writeFormattedGo applies, so splicing artifacts (an extra
// blank line, minor indentation drift) come out canonical either way.
func formatGo(src string) ([]byte, error) {
	return format.Source([]byte(src))
}
