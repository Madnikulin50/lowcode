package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"go.uber.org/zap"
)

// Version history of pages, modules and charts.
//
// Every successful create/update saves a snapshot of the resource as a new
// numbered version (identical consecutive snapshots are not repeated).
// Restoring a version applies its snapshot through the regular update, so
// permissions, validation and events work as for any edit and the restore
// itself becomes the newest version.

const (
	// versions kept per resource; the oldest are dropped first
	resourceVersionsKept = 200
)

type (
	resourceVersion struct {
		store store.Storer
	}

	// pageSnapshot is what a page version holds: the page and its layouts
	// (the blocks of a page live in the layouts)
	pageSnapshot struct {
		Page    *types.Page         `json:"page"`
		Layouts []*types.PageLayout `json:"layouts,omitempty"`
	}

	ResourceVersionService interface {
		List(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64) (types.ResourceVersionSet, error)
		Get(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64, version int) (*types.ResourceVersion, error)
		Restore(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64, version int) (any, error)
	}
)

var _ ResourceVersionService = &resourceVersion{}

func ResourceVersion() *resourceVersion {
	return &resourceVersion{
		store: DefaultStore,
	}
}

// recordResourceVersion saves the current state of a resource as a new
// version. History is an aid: a failure here is logged and never fails the
// change that triggered it.
func recordResourceVersion(ctx context.Context, s store.Storer, resourceType string, namespaceID, resourceID uint64) {
	if s == nil {
		return
	}

	if err := saveResourceVersion(ctx, s, resourceType, namespaceID, resourceID, ""); err != nil && DefaultLogger != nil {
		DefaultLogger.Warn("could not save resource version",
			zap.String("type", resourceType), zap.Uint64("id", resourceID), zap.Error(err))
	}
}

func saveResourceVersion(ctx context.Context, s store.Storer, resourceType string, namespaceID, resourceID uint64, note string) (err error) {
	var snapshot []byte

	if snapshot, err = takeSnapshot(ctx, s, resourceType, namespaceID, resourceID); err != nil {
		return
	}

	existing, _, err := store.SearchComposeResourceVersions(ctx, s, types.ResourceVersionFilter{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		return
	}

	latest := 0
	var latestSnapshot string
	for _, v := range existing {
		if v.Version > latest {
			latest, latestSnapshot = v.Version, v.Snapshot
		}
	}

	if latest > 0 && latestSnapshot == string(snapshot) {
		return nil
	}

	err = store.CreateComposeResourceVersion(ctx, s, &types.ResourceVersion{
		ID:           id.Next(),
		NamespaceID:  namespaceID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Version:      latest + 1,
		Snapshot:     string(snapshot),
		Note:         note,
		CreatedAt:    *nowUTC(),
		CreatedBy:    auth.GetIdentityFromContext(ctx).Identity(),
	})
	if err != nil {
		return
	}

	// drop the oldest versions beyond the cap
	if over := len(existing) + 1 - resourceVersionsKept; over > 0 {
		old := append(types.ResourceVersionSet(nil), existing...)
		sortVersionsAsc(old)
		for _, v := range old[:over] {
			if err = store.DeleteComposeResourceVersionByID(ctx, s, v.ID); err != nil {
				return
			}
		}
	}

	return nil
}

func sortVersionsAsc(set types.ResourceVersionSet) {
	for i := 1; i < len(set); i++ {
		for j := i; j > 0 && set[j-1].Version > set[j].Version; j-- {
			set[j-1], set[j] = set[j], set[j-1]
		}
	}
}

func takeSnapshot(ctx context.Context, s store.Storer, resourceType string, namespaceID, resourceID uint64) ([]byte, error) {
	switch resourceType {
	case types.ResourceVersionPage:
		pg, err := loadPage(ctx, s, namespaceID, resourceID)
		if err != nil {
			return nil, err
		}

		layouts, _, err := store.SearchComposePageLayouts(ctx, s, types.PageLayoutFilter{NamespaceID: namespaceID, PageID: resourceID})
		if err != nil {
			return nil, err
		}

		return json.Marshal(pageSnapshot{Page: pg, Layouts: layouts})

	case types.ResourceVersionModule:
		m, err := loadModule(ctx, s, namespaceID, resourceID)
		if err != nil {
			return nil, err
		}
		return json.Marshal(m)

	case types.ResourceVersionChart:
		c, err := loadChart(ctx, s, namespaceID, resourceID)
		if err != nil {
			return nil, err
		}
		return json.Marshal(c)
	}

	return nil, errors.InvalidData("unknown resource type %q", resourceType)
}

// authorize checks that the caller may see (write=false) or change the resource
func (svc resourceVersion) authorize(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64, write bool) error {
	switch resourceType {
	case types.ResourceVersionPage:
		pg, err := loadPage(ctx, svc.store, namespaceID, resourceID)
		if err != nil {
			return err
		}
		if write && !DefaultAccessControl.CanUpdatePage(ctx, pg) {
			return PageErrNotAllowedToUpdate()
		}
		if !write && !DefaultAccessControl.CanReadPage(ctx, pg) {
			return PageErrNotAllowedToRead()
		}

	case types.ResourceVersionModule:
		m, err := loadModule(ctx, svc.store, namespaceID, resourceID)
		if err != nil {
			return err
		}
		if write && !DefaultAccessControl.CanUpdateModule(ctx, m) {
			return ModuleErrNotAllowedToUpdate()
		}
		if !write && !DefaultAccessControl.CanReadModule(ctx, m) {
			return ModuleErrNotAllowedToRead()
		}

	case types.ResourceVersionChart:
		c, err := loadChart(ctx, svc.store, namespaceID, resourceID)
		if err != nil {
			return err
		}
		if write && !DefaultAccessControl.CanUpdateChart(ctx, c) {
			return ChartErrNotAllowedToUpdate()
		}
		if !write && !DefaultAccessControl.CanReadChart(ctx, c) {
			return ChartErrNotAllowedToRead()
		}

	default:
		return errors.InvalidData("unknown resource type %q", resourceType)
	}

	return nil
}

// List returns the versions of a resource, newest first, without snapshots
func (svc resourceVersion) List(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64) (types.ResourceVersionSet, error) {
	if err := svc.authorize(ctx, namespaceID, resourceType, resourceID, false); err != nil {
		return nil, err
	}

	set, _, err := store.SearchComposeResourceVersions(ctx, svc.store, types.ResourceVersionFilter{
		NamespaceID:  namespaceID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Sorting:      filter.Sorting{Sort: filter.SortExprSet{{Column: "version", Descending: true}}},
	})
	if err != nil {
		return nil, err
	}

	for _, v := range set {
		v.Snapshot = ""
	}

	return set, nil
}

// Get returns one version with its snapshot
func (svc resourceVersion) Get(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64, version int) (*types.ResourceVersion, error) {
	if err := svc.authorize(ctx, namespaceID, resourceType, resourceID, false); err != nil {
		return nil, err
	}

	set, _, err := store.SearchComposeResourceVersions(ctx, svc.store, types.ResourceVersionFilter{
		NamespaceID:  namespaceID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range set {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, errors.NotFound("version %d of the %s was not found", version, resourceType)
}

// Restore applies the snapshot of a version to the resource and returns the
// updated resource
func (svc resourceVersion) Restore(ctx context.Context, namespaceID uint64, resourceType string, resourceID uint64, version int) (any, error) {
	if err := svc.authorize(ctx, namespaceID, resourceType, resourceID, true); err != nil {
		return nil, err
	}

	v, err := svc.Get(ctx, namespaceID, resourceType, resourceID, version)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case types.ResourceVersionPage:
		var snap pageSnapshot
		if err = json.Unmarshal([]byte(v.Snapshot), &snap); err != nil || snap.Page == nil {
			return nil, fmt.Errorf("version %d has a damaged snapshot: %v", version, err)
		}

		snap.Page.ID, snap.Page.NamespaceID, snap.Page.UpdatedAt = resourceID, namespaceID, nil
		res, err := DefaultPage.Update(ctx, snap.Page)
		if err != nil {
			return nil, err
		}

		// layouts that still exist go back to their saved state; layouts
		// created after the version are left alone
		for _, l := range snap.Layouts {
			l.NamespaceID, l.PageID, l.UpdatedAt = namespaceID, resourceID, nil
			if _, err = store.LookupComposePageLayoutByID(ctx, svc.store, l.ID); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return nil, err
			}
			if _, err = DefaultPageLayout.Update(ctx, l); err != nil {
				return nil, err
			}
		}

		recordResourceVersion(ctx, svc.store, resourceType, namespaceID, resourceID)
		return res, nil

	case types.ResourceVersionModule:
		m := &types.Module{}
		if err = json.Unmarshal([]byte(v.Snapshot), m); err != nil {
			return nil, fmt.Errorf("version %d has a damaged snapshot: %v", version, err)
		}

		m.ID, m.NamespaceID, m.UpdatedAt = resourceID, namespaceID, nil
		return DefaultModule.Update(ctx, m)

	case types.ResourceVersionChart:
		c := &types.Chart{}
		if err = json.Unmarshal([]byte(v.Snapshot), c); err != nil {
			return nil, fmt.Errorf("version %d has a damaged snapshot: %v", version, err)
		}

		c.ID, c.NamespaceID, c.UpdatedAt = resourceID, namespaceID, nil
		return DefaultChart.Update(ctx, c)
	}

	return nil, errors.InvalidData("unknown resource type %q", resourceType)
}
