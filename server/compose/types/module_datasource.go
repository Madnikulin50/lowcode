package types

import (
	"github.com/cortezaproject/corteza/server/system/types"
)

type (
	ModuleConfigDataSource struct {
		Items  types.ReportDataSourceSet `json:"items,omitempty"`
		Enable bool                      `json:"enable,omitempty"`
	}
)
