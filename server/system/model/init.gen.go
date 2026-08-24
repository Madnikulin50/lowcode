package model

// Formerly generated from CUE; now maintained by hand.
//

import (
	"context"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
)

type (
	modelReplacer interface {
		ReplaceModel(ctx context.Context, model *dal.Model) (err error)
	}
)

var (
	models []*dal.Model
)

func Models() dal.ModelSet {
	return models
}
