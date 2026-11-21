package dalutils

import (
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/secrets"
)

// Secret-поля модулей: значения шифруются перед передачей в dal
// (recToGetters → encryptRecordSecrets) и расшифровываются после чтения
// (WalkIterator/ComposeRecordsFind → decryptRecordSecrets). Сервисный слой
// и автоматизация работают с plaintext в памяти; REST маскирует
// (server/compose/rest). Маска secrets.Mask на входе обновления означает
// «оставить старое значение» — её обрабатывает сервис записей до dalutils.

// secretFieldSet — имена Secret-полей модуля.
func secretFieldSet(mod *types.Module) map[string]bool {
	if mod == nil {
		return nil
	}
	out := map[string]bool{}
	for _, f := range mod.Fields {
		if f != nil && f.Kind == "Secret" {
			out[f.Name] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// encryptRecordSecrets шифрует значения Secret-полей записи in-place.
// Маска- sentinel пропускается (сервис уже заменил её старым значением; если
// попала сюда — не шифруем очевидный не-секрет).
func encryptRecordSecrets(mod *types.Module, r *types.Record) error {
	sec := secretFieldSet(mod)
	if len(sec) == 0 || r == nil {
		return nil
	}
	nsID := r.NamespaceID
	if nsID == 0 {
		nsID = mod.NamespaceID
	}
	for i := range r.Values {
		v := r.Values[i]
		if !sec[v.Name] {
			continue
		}
		if v.Value == "" || v.Value == secrets.Mask || secrets.IsEncrypted(v.Value) {
			continue
		}
		enc, err := secrets.Encrypt(nsID, v.Value)
		if err != nil {
			return fmt.Errorf("record %d field %q: %w", r.ID, v.Name, err)
		}
		v.Value = enc
	}
	return nil
}

// decryptRecordSecrets расшифровывает Secret-поля записи in-place. Значение,
// которое не расшифровалось ни одним ключом, оставляется как есть — клиент
// увидит его через маску/как есть, но целостность записи не роняем.
func decryptRecordSecrets(mod *types.Module, r *types.Record) {
	sec := secretFieldSet(mod)
	if len(sec) == 0 || r == nil {
		return
	}
	nsID := r.NamespaceID
	if nsID == 0 {
		nsID = mod.NamespaceID
	}
	for i := range r.Values {
		v := r.Values[i]
		if !sec[v.Name] || !secrets.IsEncrypted(v.Value) {
			continue
		}
		if dec, err := secrets.Decrypt(nsID, v.Value); err == nil {
			v.Value = dec
		}
	}
}
