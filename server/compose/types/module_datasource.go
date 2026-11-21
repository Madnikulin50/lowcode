package types

import (
	"github.com/madnikulin50/lowcode/server/system/types"
)

type (
	ModuleConfigDataSource struct {
		Items  types.ReportDataSourceSet `json:"items,omitempty"`
		Enable bool                      `json:"enable,omitempty"`
	}
)
