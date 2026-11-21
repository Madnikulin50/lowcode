package service

import (
	"context"
	"time"

	composeService "github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/store"
	"go.uber.org/zap"
)

var (
	// DefaultStore mirrors federation/service.DefaultStore: kept here (not
	// referenced from the base store package, which has no singleton of its
	// own) so REST handlers and the scanner share one instance without a
	// dedicated business-logic layer for what is, so far, plain CRUD over
	// config rows gated by the referenced Compose module's own permissions.
	DefaultStore store.Storer

	DefaultLogger *zap.Logger

	DefaultScanner *scanner

	// ScanInterval is how often the scanner walks watched modules; not yet
	// exposed as a user-facing option (see options.AnomalyOpt in a later
	// pass if this needs to be admin-configurable).
	ScanInterval = 5 * time.Minute
)

func Initialize(_ context.Context, log *zap.Logger, s store.Storer) error {
	DefaultStore = s
	DefaultLogger = log.Named("anomaly")
	DefaultScanner = Scanner(DefaultLogger, DefaultStore, composeService.DefaultRecord)

	return nil
}

// Watch starts the periodic anomaly scan. Called once at boot, alongside the
// other service Watch()ers (see server/app/boot_levels.go).
func Watch(ctx context.Context) {
	DefaultScanner.Watch(ctx, ScanInterval)
}
