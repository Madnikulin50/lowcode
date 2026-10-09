package store

import (
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
	"github.com/madnikulin50/lowcode/server/system/types"
)

type (
	setting struct {
		cfg *EncoderConfig

		res *resource.Setting
		st  *types.SettingValue

		ux *userIndex
	}
)
