package wfexec

import (
	"github.com/madnikulin50/lowcode/server/pkg/cli"
	"github.com/madnikulin50/lowcode/server/pkg/id"
)

func init() {
	id.Init(cli.Context())
}
