package {{ .package }}

{{ template "gocode/header-gentext.tpl" }}

import (
	"context"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
{{- range .imports }}
    {{ . }}
{{- end }}
)


type (
	modelReplacer interface  {
		ReplaceModel(ctx context.Context, model *dal.Model) (err error)
	}
)

var (
	models []*dal.Model
)

func Models() dal.ModelSet {
	return models
}
