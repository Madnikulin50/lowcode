package provision

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/system/types"
	"go.uber.org/zap"
)

// SystemRoles creates system roles
func SystemRoles(ctx context.Context, log *zap.Logger, s store.Storer) (rr []*types.Role, err error) {
	rr = types.RoleSet{
		&types.Role{
			Name:   "Super administrator",
			Handle: "super-admin",
			Meta: &types.RoleMeta{
				Description: "Super admin is a 'bypass' role that auto-allows all operations to it's members",
				Context:     nil,
			},
		},

		&types.Role{
			Name:   "Authenticated",
			Handle: "authenticated",
			Meta: &types.RoleMeta{
				Description: "Authenticated role is auto-assigned to all authenticated sessions",
				Context:     nil,
			},
		},

		&types.Role{
			Name:   "Anonymous",
			Handle: "anonymous",
			Meta: &types.RoleMeta{
				Description: "Authenticated role is auto-assigned to all non-authenticated sessions",
				Context:     nil,
			}},
	}

	m, err := loadRoles(ctx, s)
	if err != nil {
		return
	}

	var toCreate, toUpdate types.RoleSet

	for i := range rr {
		r := rr[i]
		existing := m[r.Handle]
		if existing == nil {
			r.ID = id.Next()
			r.CreatedAt = *now()
			m[r.Handle] = r
			toCreate = append(toCreate, r)
			log.Info("creating system role", zap.String("handle", r.Handle), logger.Uint64("ID", r.ID))
			continue
		}

		existing.DeletedAt = nil
		existing.ArchivedAt = nil
		rr[i] = existing
		toUpdate = append(toUpdate, existing)
		log.Info("updating system role", zap.String("handle", r.Handle), logger.Uint64("ID", existing.ID))
	}

	// Create/Update instead of Upsert: Postgres INSERT ... ON CONFLICT (id)
	// requires a matching unique constraint, and inference can fail even when
	// roles_pkey exists (search_path / schema mismatch). Provision already
	// knows whether the row is new.
	if len(toCreate) > 0 {
		if err := store.CreateRole(ctx, s, toCreate...); err != nil {
			return nil, fmt.Errorf("failed to provision system roles: %w", err)
		}
	}
	if len(toUpdate) > 0 {
		if err := store.UpdateRole(ctx, s, toUpdate...); err != nil {
			return nil, fmt.Errorf("failed to provision system roles: %w", err)
		}
	}

	return
}

func loadRoles(ctx context.Context, s store.Roles) (m map[string]*types.Role, err error) {
	var (
		f = types.RoleFilter{
			Archived: filter.StateInclusive,
			Deleted:  filter.StateInclusive,
		}
	)

	m = make(map[string]*types.Role)

	if set, _, err := store.SearchRoles(ctx, s, f); err == nil {
		for _, r := range set {
			m[r.Handle] = r
		}
	}

	return
}
