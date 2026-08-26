package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/compose/automation"
	"github.com/madnikulin50/lowcode/server/compose/types"
	discoveryService "github.com/madnikulin50/lowcode/server/discovery/service"
	"github.com/madnikulin50/lowcode/server/pkg/actionlog"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/corredor"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/healthcheck"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/locale"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/pkg/objstore"
	"github.com/madnikulin50/lowcode/server/pkg/objstore/minio"
	"github.com/madnikulin50/lowcode/server/pkg/objstore/plain"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/store"
	systemService "github.com/madnikulin50/lowcode/server/system/service"
	systemTypes "github.com/madnikulin50/lowcode/server/system/types"
	"go.uber.org/zap"

	"github.com/madnikulin50/lowcode/server/pkg/rag"
)

type (
	userFinder interface {
		FindByID(context.Context, uint64) (*systemTypes.User, error)
	}

	schemaAltManager interface {
		ModelAlterations(context.Context, *dal.Model) (out []*dal.Alteration, err error)
		SetAlterations(ctx context.Context, s store.Storer, m *dal.Model, stale []*dal.Alteration, set ...*dal.Alteration) (err error)
	}

	Config struct {
		ActionLog        options.ActionLogOpt
		Discovery        options.DiscoveryOpt
		Storage          options.ObjectStoreOpt
		Limit            options.LimitOpt
		ImageSearch      options.ImageSearchOpt
		Cache            options.CacheOpt
		UserFinder       userFinder
		SchemaAltManager schemaAltManager
	}

	eventDispatcher interface {
		WaitFor(ctx context.Context, ev eventbus.Event) (err error)
		Dispatch(ctx context.Context, ev eventbus.Event)
	}
)

var (
	DefaultObjectStore objstore.Store

	// DefaultStore is an interface to storage backend(s)
	// ng (next-gen) is a temporary prefix
	// so that we can differentiate between it and the file-only store
	DefaultStore store.Storer

	DefaultLogger *zap.Logger

	DefaultActionlog actionlog.Recorder

	// DefaultAccessControl Access control checking
	DefaultAccessControl *accessControl

	DefaultNamespace           NamespaceService
	DefaultImportSession       ImportSessionService
	DefaultRecord              *record
	DefaultModule              ModuleService
	DefaultChart               *chart
	DefaultPage                *page
	DefaultChat                *chatService
	DefaultPageLayout          *pageLayout
	DefaultAttachment          AttachmentService
	DefaultNotification        *notification
	DefaultResourceTranslation ResourceTranslationsManagerService
	DefaultDataPrivacy         DataPrivacyService
	DefaultImageSearch         ImageSearchService
	DefaultETL                 ETLService
	DefaultRAG                 *RAGService
	DefaultPagesRAG            *PagesRAGService

	// wrapper around time.Now() that will aid service testing
	now = func() *time.Time {
		c := time.Now().Round(time.Second)
		return &c
	}

	// wrapper around time.Now() that will aid service testing
	nowUTC = func() *time.Time {
		c := time.Now().Round(time.Second).UTC()
		return &c
	}

	// wrapper around nextID that will aid service testing
	nextID = func() uint64 {
		return id.Next()
	}
)

// Initialize compose-only services
func Initialize(ctx context.Context, log *zap.Logger, s store.Storer, c Config) (err error) {
	var (
		hcd = healthcheck.Defaults()
	)

	DefaultStore = s

	DefaultLogger = log.Named("service")

	{
		tee := zap.NewNop()
		policy := actionlog.MakeProductionPolicy()

		if !c.ActionLog.Enabled {
			policy = actionlog.MakeDisabledPolicy()
		} else if c.ActionLog.Debug {
			policy = actionlog.MakeDebugPolicy()
			tee = logger.MakeDebugLogger()
		}

		DefaultActionlog = actionlog.NewService(DefaultStore, log, tee, policy)
	}

	// Activity log for compose resources
	{
		l := log
		if !c.Discovery.Debug {
			l = zap.NewNop()
		}

		DefaultResourceActivity := discoveryService.ResourceActivity(l, c.Discovery, DefaultStore, eventbus.Service())
		err = DefaultResourceActivity.InitResourceActivityLog(ctx, []string{
			(types.Namespace{}).LabelResourceKind(),
			(types.Module{}).LabelResourceKind(),
			"compose:record",
		})
		if err != nil {
			return err
		}
	}

	DefaultAccessControl = AccessControl(s)
	DefaultResourceTranslation = ResourceTranslationsManager(locale.Global())

	if DefaultObjectStore == nil {
		var (
			opt    = c.Storage
			bucket string
		)
		const svcPath = "compose"
		if opt.MinioEndpoint != "" {
			bucket = minio.GetBucket(opt.MinioBucket, svcPath)

			DefaultObjectStore, err = minio.New(bucket, opt.MinioPathPrefix, svcPath, minio.Options{
				Endpoint:        opt.MinioEndpoint,
				Secure:          opt.MinioSecure,
				Strict:          opt.MinioStrict,
				AccessKeyID:     opt.MinioAccessKey,
				SecretAccessKey: opt.MinioSecretKey,

				ServerSideEncryptKey: []byte(opt.MinioSSECKey),
			})

			log.Info("initializing minio",
				zap.String("bucket", bucket),
				zap.String("endpoint", opt.MinioEndpoint),
				zap.Error(err))
		} else {
			path := opt.Path + "/" + svcPath
			DefaultObjectStore, err = plain.New(path)
			log.Info("initializing store",
				zap.String("path", path),
				zap.Error(err))

		}

		hcd.Add(objstore.Healthcheck(DefaultObjectStore), "ObjectStore/Compose")

		if err != nil {
			return err
		}
	}

	DefaultNamespace = Namespace()
	DefaultModule = Module(c.SchemaAltManager)

	DefaultImportSession = ImportSession()
	DefaultRecord = Record(RecordOptions{LimitRecords: c.Limit.RecordCountPerModule, ReportCacheTTL: c.Cache.ReportTTL})
	DefaultPage = Page()
	DefaultPageLayout = PageLayout()
	DefaultChart = Chart()
	DefaultChat = Chat()
	DefaultNotification = Notification(c.UserFinder)
	DefaultAttachment = Attachment(DefaultObjectStore, dal.Service())
	DefaultDataPrivacy = DataPrivacy()
	DefaultImageSearch = ImageSearch(c.ImageSearch.Enabled)
	DefaultETL = ETL()

	ragStore, err := rag.NewStore("data/rag.db")
	if err != nil {
		panic(fmt.Errorf("rag store: %w", err))
	}
	ollamaHost := chat.EffectiveOllamaURL()
	DefaultRAG = NewRAGService(ragStore, rag.NewEmbedder(ollamaHost, "nomic-embed-text"))

	DefaultPagesRAG = NewPagesRAGService(ragStore, rag.NewEmbedder(ollamaHost, "nomic-embed-text"), log, DefaultPage, DefaultNamespace, DefaultRecord, DefaultResourceTranslation)
	DefaultPagesRAG.StartDailyCrawl(ctx)

	RegisterIteratorProviders()

	automationService.Registry().AddTypes(
		automation.ComposeNamespace{},
		automation.ComposeModule{},
		automation.ComposeRecord{},
		automation.ComposeRecordValues{},
		automation.Attachment{},
	)

	automation.RecordsHandler(
		automationService.Registry(),
		DefaultNamespace,
		DefaultModule,
		DefaultRecord,
	)

	automationService.Registry().AddFunctions(LoopIncidentApply())

	automation.ModulesHandler(
		automationService.Registry(),
		DefaultNamespace,
		DefaultModule,
	)

	automation.NamespacesHandler(
		automationService.Registry(),
		DefaultNamespace,
	)

	automation.AttachmentHandler(
		automationService.Registry(),
		DefaultAttachment,
	)

	automation.NotificationHandler(
		automationService.Registry(),
		systemService.DefaultNotification,
		systemService.DefaultUser,
		DefaultNamespace,
		DefaultModule,
		log,
	)

	return nil
}

func Activate(ctx context.Context) (err error) {
	err = DefaultModule.ReloadDALModels(ctx)
	if err != nil {
		return err
	}

	return
}

func Watchers(ctx context.Context) {
	//
}

func RegisterIteratorProviders() {
	// Register resource finders on iterator
	corredor.Service().RegisterIteratorProvider(
		"compose:record",
		func(ctx context.Context, f map[string]string, h eventbus.HandlerFn, action string) (err error) {
			rf := types.RecordFilter{Query: f["query"]}
			rf.Sort.Set(f["sort"])

			var limit uint
			if lString, has := f["limit"]; has {
				if limit64, err := strconv.ParseUint(lString, 10, 32); err != nil {
					return fmt.Errorf("cannot parse iterator limit param: %w", err)
				} else {
					// We specify the bit size during parsing, so this is fine
					limit = uint(limit64)
				}
			}

			page := f["page"]
			rf.Paging, err = filter.NewPaging(uint(limit), page)
			if err != nil {
				return err
			}

			if nsLookup, has := f["namespace"]; !has {
				return fmt.Errorf("namespace for record iteration filter not defined")
			} else if ns, err := DefaultNamespace.FindByAny(ctx, nsLookup); err != nil {
				return err
			} else {
				rf.NamespaceID = ns.ID
			}

			if mLookup, has := f["module"]; !has {
				return fmt.Errorf("module for record iteration filter not defined")
			} else if m, err := DefaultModule.FindByAny(ctx, rf.NamespaceID, mLookup); err != nil {
				return err
			} else {
				rf.ModuleID = m.ID
			}

			return DefaultRecord.Iterator(ctx, rf, h, action)
		},
	)
}

// sameInstant compares timestamps at second precision in UTC.
// JS Date only has milliseconds and Date.toISOString() is always UTC, while
// Postgres timestamptz and Go time.Time may carry microseconds and a location.
func sameInstant(a, b time.Time) bool {
	return a.UTC().Truncate(time.Second).Equal(b.UTC().Truncate(time.Second))
}

// Data is stale when new date does not match updatedAt or createdAt (before first update)
func isStale(new *time.Time, updatedAt *time.Time, createdAt time.Time) bool {
	if new == nil {
		// Change to true for stale-data-check
		return false
	}

	if updatedAt != nil {
		return !sameInstant(*new, *updatedAt)
	}

	return !sameInstant(*new, createdAt)
}

// trim1st removes 1st param and returns only error
func trim1st(_ interface{}, err error) error {
	return err
}
