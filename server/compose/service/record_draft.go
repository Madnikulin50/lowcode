package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/dalutils"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store"
)

// Record drafts kept on the server.
//
// A draft is an unsaved edit of a record (or of a record not created yet) that
// its author wants to come back to, from any browser. Drafts are private: every
// call works on the drafts of the signed-in user only, whatever the client says
// about the author. Saving one needs the permission the edit itself would need
// (update on an existing record, create on the module for a new one), so a
// draft cannot be used to keep values the user could not have entered.
//
// A draft is only data: applying it is the normal record save, which runs all
// the checks again.

const (
	// recordDraftsKept: drafts per user; the least recently updated go first
	recordDraftsKept = 200

	// recordDraftMaxPayload: the draft as JSON, in bytes
	recordDraftMaxPayload = 1 << 20

	recordDraftMaxChangeID = 64
)

type (
	recordDraft struct {
		store store.Storer
		ac    recordDraftAccessController
		dal   dalDater
	}

	recordDraftAccessController interface {
		CanUpdateRecord(context.Context, *types.Record) bool
		CanCreateRecordOnModule(context.Context, *types.Module) bool
	}

	RecordDraftService interface {
		List(ctx context.Context, f types.RecordDraftFilter) (types.RecordDraftSet, error)
		Save(ctx context.Context, d *types.RecordDraft) (*types.RecordDraft, error)
		Delete(ctx context.Context, changeID string) error
		Clear(ctx context.Context) error
	}
)

var _ RecordDraftService = &recordDraft{}

func RecordDraft() *recordDraft {
	return &recordDraft{store: DefaultStore, ac: DefaultAccessControl, dal: dal.Service()}
}

// List returns the drafts of the signed-in user, newest first. The filter may
// narrow them down to a namespace, module or record.
func (svc recordDraft) List(ctx context.Context, f types.RecordDraftFilter) (types.RecordDraftSet, error) {
	userID, err := draftUser(ctx)
	if err != nil {
		return nil, err
	}

	f.UserID = userID
	f.ChangeID = ""

	set, _, err := store.SearchComposeRecordDrafts(ctx, svc.store, f)
	if err != nil {
		return nil, err
	}

	sortDraftsNewestFirst(set)
	return set, nil
}

// Save creates the draft or replaces the user's draft with the same change ID
func (svc recordDraft) Save(ctx context.Context, d *types.RecordDraft) (*types.RecordDraft, error) {
	userID, err := draftUser(ctx)
	if err != nil {
		return nil, err
	}

	if err = validateRecordDraft(d); err != nil {
		return nil, err
	}

	_, m, err := loadModuleCombo(ctx, svc.store, d.NamespaceID, d.ModuleID)
	if err != nil {
		return nil, err
	}

	if d.RecordID > 0 {
		r, err := dalutils.ComposeRecordsFind(ctx, svc.dal, m, d.RecordID)
		if err != nil {
			return nil, err
		}
		if !svc.ac.CanUpdateRecord(ctx, r) {
			return nil, RecordErrNotAllowedToUpdate()
		}
	} else if !svc.ac.CanCreateRecordOnModule(ctx, m) {
		return nil, RecordErrNotAllowedToCreate()
	}

	return saveRecordDraft(ctx, svc.store, userID, d)
}

// Delete removes one draft of the user; a draft that is not there is not an error
func (svc recordDraft) Delete(ctx context.Context, changeID string) error {
	userID, err := draftUser(ctx)
	if err != nil {
		return err
	}

	set, _, err := store.SearchComposeRecordDrafts(ctx, svc.store, types.RecordDraftFilter{UserID: userID, ChangeID: changeID})
	if err != nil {
		return err
	}

	for _, d := range set {
		if err = store.DeleteComposeRecordDraftByID(ctx, svc.store, d.ID); err != nil {
			return err
		}
	}

	return nil
}

// Clear removes every draft of the user
func (svc recordDraft) Clear(ctx context.Context) error {
	userID, err := draftUser(ctx)
	if err != nil {
		return err
	}

	set, _, err := store.SearchComposeRecordDrafts(ctx, svc.store, types.RecordDraftFilter{UserID: userID})
	if err != nil {
		return err
	}

	for _, d := range set {
		if err = store.DeleteComposeRecordDraftByID(ctx, svc.store, d.ID); err != nil {
			return err
		}
	}

	return nil
}

func draftUser(ctx context.Context) (uint64, error) {
	userID := auth.GetIdentityFromContext(ctx).Identity()
	if userID == 0 {
		return 0, errors.Unauthorized("drafts need a signed-in user")
	}
	return userID, nil
}

func validateRecordDraft(d *types.RecordDraft) error {
	d.ChangeID = strings.TrimSpace(d.ChangeID)

	switch {
	case d.ChangeID == "" || len(d.ChangeID) > recordDraftMaxChangeID:
		return errors.InvalidData("the draft needs a change ID of up to %d characters", recordDraftMaxChangeID)
	case d.NamespaceID == 0 || d.ModuleID == 0:
		return errors.InvalidData("the draft needs a namespace and a module")
	case len(d.Payload) > recordDraftMaxPayload:
		return errors.InvalidData("the draft is larger than %d bytes", recordDraftMaxPayload)
	case !json.Valid([]byte(d.Payload)):
		return errors.InvalidData("the draft is not valid JSON")
	}

	return nil
}

// saveRecordDraft stores the draft for the user (permissions are checked by
// the caller): a draft with the same change ID is replaced, and the user's
// oldest drafts are dropped beyond recordDraftsKept.
func saveRecordDraft(ctx context.Context, s store.Storer, userID uint64, d *types.RecordDraft) (*types.RecordDraft, error) {
	mine, _, err := store.SearchComposeRecordDrafts(ctx, s, types.RecordDraftFilter{UserID: userID})
	if err != nil {
		return nil, err
	}

	t := now()
	out := &types.RecordDraft{
		UserID:      userID,
		ChangeID:    d.ChangeID,
		NamespaceID: d.NamespaceID,
		ModuleID:    d.ModuleID,
		RecordID:    d.RecordID,
		Payload:     d.Payload,
		UpdatedAt:   t,
	}

	var replaced *types.RecordDraft
	for _, e := range mine {
		if e.ChangeID == d.ChangeID {
			replaced = e
			break
		}
	}

	if replaced != nil {
		out.ID, out.CreatedAt = replaced.ID, replaced.CreatedAt
		if err = store.UpdateComposeRecordDraft(ctx, s, out); err != nil {
			return nil, err
		}
		return out, nil
	}

	out.ID, out.CreatedAt = nextID(), *t
	if err = store.CreateComposeRecordDraft(ctx, s, out); err != nil {
		return nil, err
	}

	if over := len(mine) + 1 - recordDraftsKept; over > 0 {
		sorted := append(types.RecordDraftSet(nil), mine...)
		sort.Slice(sorted, func(i, j int) bool { return draftTime(sorted[i]).Before(draftTime(sorted[j])) })
		for _, e := range sorted[:over] {
			if err = store.DeleteComposeRecordDraftByID(ctx, s, e.ID); err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

func draftTime(d *types.RecordDraft) time.Time {
	if d.UpdatedAt != nil {
		return *d.UpdatedAt
	}
	return d.CreatedAt
}

func sortDraftsNewestFirst(set types.RecordDraftSet) {
	sort.Slice(set, func(i, j int) bool { return draftTime(set[j]).Before(draftTime(set[i])) })
}
