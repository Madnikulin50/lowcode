package envoy

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/madnikulin50/lowcode/server/pkg/envoyx"
)

type (
	CsvEncoder struct{}
)

func (e CsvEncoder) Encode(ctx context.Context, p envoyx.EncodeParams, rt string, nodes envoyx.NodeSet, tt envoyx.Traverser) (err error) {
	w, err := e.getWriter(p)
	if err != nil {
		return
	}

	cw := csv.NewWriter(w)

	switch rt {
	case ComposeRecordDatasourceAuxType:
		_, err = e.encodeRecordDatasources(ctx, cw, p, nodes, tt)
		if err != nil {
			return
		}
	}

	cw.Flush()
	return
}

func (e CsvEncoder) encodeRecordDatasources(ctx context.Context, writer *csv.Writer, p envoyx.EncodeParams, nodes envoyx.NodeSet, tt envoyx.Traverser) (out any, err error) {
	for _, n := range nodes {
		_, err = e.encodeRecordDatasource(ctx, writer, p, n, tt)
		if err != nil {
			return
		}
	}

	return
}

func (e CsvEncoder) encodeRecordDatasource(ctx context.Context, writer *csv.Writer, p envoyx.EncodeParams, node *envoyx.Node, tt envoyx.Traverser) (_ any, err error) {
	err = streamRecordTable(ctx, p, node, writer.Write, writer.Write)
	return
}

func (e CsvEncoder) getWriter(p envoyx.EncodeParams) (out io.Writer, err error) {
	return getEncoderWriter(p, "csv")
}
