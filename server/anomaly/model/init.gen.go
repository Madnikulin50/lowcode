package model

// This file is auto-generated version 2.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from <no value>
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
