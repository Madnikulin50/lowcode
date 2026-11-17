package envoy

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
	"github.com/madnikulin50/lowcode/server/pkg/envoyx/datasource"
	"github.com/spf13/cast"
)

// streamRecordTable walks the record datasource and hands tabular data
// (header once, then every row) to the given callbacks.
//
// It is shared by all tabular encoders (csv, xlsx, pdf) so they produce
// identical columns and multi-value formatting.
func streamRecordTable(
	ctx context.Context,
	p envoyx.EncodeParams,
	node *envoyx.Node,
	writeHeader func(header []string) error,
	writeRow func(row []string) error,
) (err error) {
	rds := node.Datasource.(*RecordDatasource)
	resolved := map[string]string{}

	header := make([]string, 0, 4)

	hasID := false
	labels := map[string]string{}
	for _, m := range p.FieldMapping {
		header = append(header, m.Field)

		if m.Column != "" && m.Column != m.Field {
			labels[m.Field] = m.Column
		}

		hasID = hasID || strings.ToLower(m.Field) == "id"
	}

	if !hasID {
		header = append([]string{"ID"}, header...)
	}

	mvDelimiter := ";"
	wrapBrackets := false

	switch cast.ToString(p.Params["multiValueDelimiter"]) {
	case ",":
		mvDelimiter = ","
	case ";":
		mvDelimiter = ";"
	case "|":
		mvDelimiter = "|"
	case "[,]":
		mvDelimiter = ","
		wrapBrackets = true
	case "[;]":
		mvDelimiter = ";"
		wrapBrackets = true
	case "[|]":
		mvDelimiter = "|"
		wrapBrackets = true
	}

	getValue := func(cache datasource.RawRecord, h string) string {
		v := cache[h]
		if len(v.Values) == 0 {
			return ""
		}

		// Encode as mf if the OG field is multi value OR we're on the resolved record column
		src, ok := resolved[h]
		if !rds.multivalues[h] && !(ok && rds.multivalues[src]) {
			return v.Values[0]
		}

		auxv := make([]string, 0, len(v.Values))
		for _, vv := range v.Values {
			if strings.Contains(vv, mvDelimiter) {
				auxv = append(auxv, fmt.Sprintf("\"%s\"", vv))
			} else {
				auxv = append(auxv, vv)
			}
		}

		out := strings.Join(auxv, mvDelimiter)
		if wrapBrackets {
			out = fmt.Sprintf("[%s]", out)
		}

		return out
	}

	var (
		more     bool
		hWritten bool
	)
	for {
		cache := make(datasource.RawRecord)
		_, more, err = rds.Next(ctx, cache)
		if err != nil || !more {
			return
		}

		if len(header) == 0 {
			for k := range cache {
				header = append(header, k)
			}
		}

		if !hWritten {
			hWritten = true

			// Splice in resolved ref values
			for i, h := range header {
				if _, ok := cache[fmt.Sprintf("%s value", h)]; ok {
					header = append(header, "")
					copy(header[i+1:], header[i:])
					header[i] = fmt.Sprintf("%s value", h)

					resolved[header[i]] = h
				}
			}

			display := make([]string, len(header))
			for i, h := range header {
				display[i] = labels[h]
				if display[i] != "" {
					continue
				}

				// Resolved reference column: "<field> value" -> "<label> value"
				if src, ok := resolved[h]; ok && labels[src] != "" {
					display[i] = labels[src] + strings.TrimPrefix(h, src)
				} else {
					display[i] = h
				}
			}

			if err = writeHeader(display); err != nil {
				return
			}
		}

		row := make([]string, 0, len(header))
		for _, h := range header {
			row = append(row, getValue(cache, h))
		}

		if err = writeRow(row); err != nil {
			return
		}
	}
}

func getEncoderWriter(p envoyx.EncodeParams, name string) (out io.Writer, err error) {
	aux, ok := p.Params[paramsKeyWriter]
	if ok {
		out, ok = aux.(io.Writer)
		if ok {
			return
		}
	}

	err = fmt.Errorf("%s encoder expects a writer conforming to io.Writer interface", name)
	return
}
