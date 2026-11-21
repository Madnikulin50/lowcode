package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// a column added to a table with rows reads back as the model's "{}" default
func TestPromptJSONColumns_ReadEmptyObjectAsEmpty(t *testing.T) {
	var (
		req = require.New(t)
		rq  PromptRequires
		rf  PromptFiles
		pc  PromptCases
	)
	req.NoError(rq.Scan([]byte("{}")))
	req.NoError(rf.Scan("{}"))
	req.NoError(pc.Scan([]byte(" {} ")))
	req.Empty(rq)
	req.Empty(rf)
	req.Empty(pc)

	req.NoError(rq.Scan([]byte(`["compose.records"]`)))
	req.Equal(PromptRequires{"compose.records"}, rq)
	req.NoError(rf.Scan([]byte(`[{"path":"a","content":"b"}]`)))
	req.Len(rf, 1)
	req.NoError(rq.Scan(nil))
}
