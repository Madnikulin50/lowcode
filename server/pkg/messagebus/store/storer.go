package store

import (
	"github.com/madnikulin50/lowcode/server/pkg/messagebus/types"
)

type (
	Storer interface {
		SetStore(types.QueueStorer)
		GetStore() types.QueueStorer
	}
)
