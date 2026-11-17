package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/secrets"
)

// Secret-поля модулей в REST-представлении записей: по умолчанию значения
// заменяются маской; полное значение отдаётся только когда клиент явно
// запросил ?exposeSecrets=true И имеет право обновления записи (редактор
// формы). Сервисный слой и автоматизация работают с plaintext в памяти —
// маскирование — только здесь, на границе HTTP.

type exposeSecretsCtxKey struct{}

func withExposeSecrets(ctx context.Context) context.Context {
	return context.WithValue(ctx, exposeSecretsCtxKey{}, true)
}

func exposeSecretsRequested(ctx context.Context) bool {
	v, _ := ctx.Value(exposeSecretsCtxKey{}).(bool)
	return v
}

// maskRecordSecretValues закрывает маской все Secret-поля записи.
func maskRecordSecretValues(m *types.Module, r *types.Record) {
	if m == nil || r == nil {
		return
	}
	for _, f := range m.Fields {
		if f == nil || f.Kind != "Secret" {
			continue
		}
		for _, v := range r.Values {
			if v.Name == f.Name && v.Value != "" && v.Value != secrets.Mask {
				v.Value = secrets.Mask
			}
		}
	}
}

// exposeSecretsAllowed — true, если клиент запросил expose и имеет право
// обновления записи (секрет показываем редакторам, не зрителям).
func (ctrl Record) exposeSecretsAllowed(ctx context.Context, r *types.Record) bool {
	if !exposeSecretsRequested(ctx) || r == nil {
		return false
	}
	return ctrl.ac.CanUpdateRecord(ctx, r)
}
