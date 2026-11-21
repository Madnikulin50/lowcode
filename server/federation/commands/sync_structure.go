package commands

import (
	"context"

	cs "github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/federation/service"
	ss "github.com/madnikulin50/lowcode/server/system/service"
	"github.com/spf13/cobra"
)

func commandSyncStructure(ctx context.Context) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		syncService := service.NewSync(
			&service.Syncer{},
			&service.Mapper{},
			service.DefaultSharedModule,
			cs.DefaultRecord,
			ss.DefaultUser,
			ss.DefaultRole)

		syncStructure := service.WorkerStructure(syncService, service.DefaultLogger)
		syncStructure.Watch(ctx, service.DefaultOptions.StructureMonitorInterval, service.DefaultOptions.StructurePageSize)
	}
}
