package service

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/revisions"
	"github.com/madnikulin50/lowcode/server/pkg/secrets"
)

type (
	// wrapper around revision.service
	//
	// handles all module-to-model conversion
	recordRevisions struct {
		r interface {
			Search(ctx context.Context, mf dal.ModelRef, f filter.Filter) (_ dal.Iterator, err error)
			Create(ctx context.Context, mf dal.ModelRef, revision *revisions.Revision) error
		}
	}
)

// constructs DAL model-ref from module
//
// Since we're using revisions model resource type needs to be altered.
func (svc *recordRevisions) modelRef(mod *types.Module) (mf dal.ModelRef) {
	if mod == nil {
		return
	}
	mf = mod.ModelRef()

	// removing resource ID from model filter
	// this will force DAL service to use FindModelByResourceIdent
	// to find model with same Resource value and RevisionResourceType
	// for ResourceType
	mf.ResourceID = 0
	mf.ResourceType = revisions.RevisionResourceType
	return
}

func (svc *recordRevisions) search(ctx context.Context, rec *types.Record, sorting filter.Sorting) (_ dal.Iterator, err error) {
	var (
		revModRef = svc.modelRef(rec.GetModule())
		sortExpr  = sorting.OrderBy()
	)

	if len(sortExpr) == 0 {
		sortExpr = filter.SortExprSet{
			&filter.SortExpr{Column: "revision", Descending: true},
		}
	}

	revFilter := filter.Generic(
		filter.WithConstraint("rel_resource", rec.ID),
		filter.WithOrderBy(sortExpr),
	)

	return svc.r.Search(ctx, revModRef, revFilter)
}

func (svc *recordRevisions) created(ctx context.Context, new *types.Record) (err error) {
	var (
		skipList  = svc.skippedField(new.GetModule())
		invokerID = auth.GetIdentityFromContext(ctx).Identity()
		rev       *revisions.Revision
	)

	rev = revisions.Make(revisions.Created, new.Revision, new.ID, invokerID)
	// Snapshot: к этому моменту значения записи уже зашифрованы (dalutils) —
	// клонируем и расшифровываем для честного диффа, в ревизию пишем
	// шифротекст.
	cloned := cloneRecordWithDecryptedSecrets(new)
	err = rev.CollectChanges(cloned, nil, skipList...)
	if err != nil {
		return
	}
	if err = encryptRevisionSecretChanges(new.GetModule(), rev); err != nil {
		return
	}

	return svc.r.Create(ctx, svc.modelRef(new.GetModule()), rev)
}

func (svc *recordRevisions) updated(ctx context.Context, upd, old *types.Record) (err error) {
	var (
		skipList  = svc.skippedField(old.GetModule())
		invokerID = auth.GetIdentityFromContext(ctx).Identity()
		rev       *revisions.Revision
	)

	rev = revisions.Make(revisions.Updated, upd.Revision, upd.ID, invokerID)
	cloned := cloneRecordWithDecryptedSecrets(upd)
	err = rev.CollectChanges(cloned, old, skipList...)
	if err != nil {
		return
	}
	if err = encryptRevisionSecretChanges(upd.GetModule(), rev); err != nil {
		return
	}

	return svc.r.Create(ctx, svc.modelRef(upd.GetModule()), rev)
}

func (svc *recordRevisions) softDeleted(ctx context.Context, del *types.Record) (err error) {
	var (
		invokerID = auth.GetIdentityFromContext(ctx).Identity()
		rev       *revisions.Revision
	)

	rev = revisions.Make(revisions.SoftDeleted, del.Revision, del.ID, invokerID)

	return svc.r.Create(ctx, svc.modelRef(del.GetModule()), rev)
}

func (svc *recordRevisions) undeleted(ctx context.Context, undel *types.Record) (err error) {
	var (
		invokerID = auth.GetIdentityFromContext(ctx).Identity()
		rev       *revisions.Revision
	)

	rev = revisions.Make(revisions.Undeleted, undel.Revision, undel.ID, invokerID)

	return svc.r.Create(ctx, svc.modelRef(undel.GetModule()), rev)
}

func (svc *recordRevisions) skippedField(mod *types.Module) []string {
	list := []string{
		"ID",
		"namespaceID",
		"moduleID",
		"revision",
		"meta",
		"createdBy",
		"createdAt",
		"updatedBy",
		"updatedAt",
		"deletedBy",
		"deletedAt",
	}

	for _, f := range mod.Fields {
		if f.Config.RecordRevisions.Skip {
			list = append(list, f.Name)
		}
	}

	return list
}

func (svc *recordRevisions) hardDeleted(ctx context.Context, del *types.Record) (err error) {
	var (
		invokerID = auth.GetIdentityFromContext(ctx).Identity()
		rev       *revisions.Revision
	)

	rev = revisions.Make(revisions.HardDeleted, del.Revision, del.ID, invokerID)
	return svc.r.Create(ctx, svc.modelRef(del.GetModule()), rev)
}

// load returns all revisions of the record
func (svc *recordRevisions) load(ctx context.Context, rec *types.Record) (out []*revisions.Revision, err error) {
	iter, err := svc.search(ctx, rec, filter.Sorting{})
	if err != nil {
		return
	}
	defer iter.Close()

	for iter.Next(ctx) {
		rev := &revisions.Revision{}
		if err = iter.Scan(rev); err != nil {
			return
		}
		out = append(out, rev)
	}

	return out, iter.Err()
}

// --- Secret-поля в revision-снапшотах ---------------------------------------

// cloneRecordWithDecryptedSecrets клонирует запись и расшифровывает в клоне
// Secret-поля: к моменту логирования ревизии значения уже зашифрованы
// (dalutils шифрует in-place перед записью), а для честного диффа нужен
// plaintext. Исходная запись не трогается.
func cloneRecordWithDecryptedSecrets(r *types.Record) *types.Record {
	if r == nil {
		return nil
	}
	mod := r.GetModule()
	out := *r
	out.Values = make(types.RecordValueSet, len(r.Values))
	copy(out.Values, r.Values)
	sec := map[string]bool{}
	if mod != nil {
		for _, f := range mod.Fields {
			if f.Kind == "Secret" {
				sec[f.Name] = true
			}
		}
	}
	if len(sec) == 0 {
		return &out
	}
	for i, v := range out.Values {
		if !sec[v.Name] || !secrets.IsEncrypted(v.Value) {
			continue
		}
		if dec, err := secrets.Decrypt(out.NamespaceID, v.Value); err == nil {
			out.Values[i].Value = dec
		}
	}
	return &out
}

// encryptRevisionSecretChanges шифрует значения Secret-полей в снапшоте
// ревизии, чтобы история изменений не хранила секреты открыто.
func encryptRevisionSecretChanges(mod *types.Module, rev *revisions.Revision) error {
	if mod == nil || rev == nil {
		return nil
	}
	sec := map[string]bool{}
	for _, f := range mod.Fields {
		if f.Kind == "Secret" {
			sec[f.Name] = true
		}
	}
	if len(sec) == 0 {
		return nil
	}
	for _, ch := range rev.Changes {
		if ch == nil || !sec[ch.Key] {
			continue
		}
		for _, side := range []*[]any{&ch.Old, &ch.New} {
			for i, v := range *side {
				s, ok := v.(string)
				if !ok || s == "" || s == secrets.Mask || secrets.IsEncrypted(s) {
					continue
				}
				enc, err := secrets.Encrypt(mod.NamespaceID, s)
				if err != nil {
					return err
				}
				(*side)[i] = enc
			}
		}
	}
	return nil
}
