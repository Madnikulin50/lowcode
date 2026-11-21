package system

import (
	"context"
	discoveryType "github.com/madnikulin50/lowcode/server/discovery/types"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/system/service"
	"testing"
)

func (h helper) clearActivityLog() {
	h.noError(store.TruncateResourceActivitys(context.Background(), service.DefaultStore))
}

func (h helper) repoMakeActivityLog() *discoveryType.ResourceActivity {
	var res = &discoveryType.ResourceActivity{
		ID:             id.Next(),
		ResourceID:     id.Next(),
		ResourceType:   "compose:record",
		ResourceAction: "create",
	}

	h.a.NoError(store.CreateResourceActivity(context.Background(), service.DefaultStore, res))

	return res
}

func TestCreateActivityLog(t *testing.T) {
	h := newHelper(t)
	h.clearActionLog()

	h.repoMakeActivityLog()
}
