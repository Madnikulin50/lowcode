package decoder

import (
	"github.com/madnikulin50/lowcode/server/compose/types"
)

type (
	ComposeRecord struct {
		types.Record
	}
	ComposeRecordSet []*ComposeRecord

	ComposeRecordFilter struct {
		types.RecordFilter
	}
)
