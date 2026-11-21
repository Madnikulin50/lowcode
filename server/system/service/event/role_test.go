package event

import (
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/system/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleMatching(t *testing.T) {
	var (
		a   = assert.New(t)
		res = &roleBase{
			role: &types.Role{Handle: "admin"},
		}

		cRol = eventbus.MustMakeConstraint("role", "eq", "admin")
	)

	a.True(res.Match(cRol))
}
