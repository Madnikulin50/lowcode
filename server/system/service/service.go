package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bep/godartsass/v2"

	automationService "github.com/madnikulin50/lowcode/server/automation/service"
	discoveryService "github.com/madnikulin50/lowcode/server/discovery/service"
	"github.com/madnikulin50/lowcode/server/pkg/actionlog"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/healthcheck"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/label"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/pkg/objstore"
	"github.com/madnikulin50/lowcode/server/pkg/objstore/minio"
	"github.com/madnikulin50/lowcode/server/pkg/objstore/plain"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/pkg/rbac"
	"github.com/madnikulin50/lowcode/server/pkg/valuestore"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/system/automation"
	"github.com/madnikulin50/lowcode/server/system/types"
	"go.uber.org/zap"
)

type (
	websocketSender interface {
		Send(kind string, payload interface{}, userIDs ...uint64) error
	}

	Config struct {
		ActionLog  options.ActionLogOpt
		Discovery  options.DiscoveryOpt
		Storage    options.ObjectStoreOpt
		DB         options.DBOpt
		Template   options.TemplateOpt
		Auth       options.AuthOpt
		RBAC       options.RbacOpt
		Limit      options.LimitOpt
		Attachment options.AttachmentOpt
		Webapps    options.WebappOpt
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

	// DefaultSettings controls system's settings
	DefaultSettings *settings

	DefaultStylesheet *stylesheet

	// DefaultAccessControl Access control checking
	DefaultAccessControl *accessControl

	DefaultAuthNotification AuthNotificationService

	// CurrentSettings represents current system settings
	CurrentSettings = &types.AppSettings{}

	DefaultActionlog actionlog.Recorder

	DefaultSink *sink

	DefaultAuth                *auth
	DefaultAuthClient          *authClient
	DefaultUser                *user
	DefaultCredentials         *credentials
	DefaultDalConnection       *dalConnection
	DefaultDalSensitivityLevel *dalSensitivityLevel
	DefaultDalSchemaAlteration *DalSchemaAlteration
	DefaultRole                *role
	DefaultUserGroup           *userGroup
	DefaultApplication         *application
	DefaultReminder            ReminderService
	DefaultNotification        NotificationService
	DefaultAttachment          AttachmentService
	DefaultRenderer            TemplateService
	DefaultResourceTranslation ResourceTranslationService
	DefaultQueue               *queue
	DefaultApigwRoute          *apigwRoute
	DefaultApigwFilter         *apigwFilter
	DefaultApigwProfiler       *apigwProfiler
	DefaultReport              *report
	DefaultDataPrivacy         *dataPrivacy
	DefaultSMTPChecker         *smtpConfigurationChecker
	DefaultExpression          *expression

	DefaultStatistics *statistics

	// wrapper around time.Now() that will aid service testing
	now = func() *time.Time {
		c := time.Now().Round(time.Second)
		return &c
	}

	// wrapper around nextID that will aid service testing
	nextID = func() uint64 {
		return id.Next()
	}
)

func Initialize(ctx context.Context, log *zap.Logger, s store.Storer, ws websocketSender, c Config) (err error) {
	var (
		hcd = healthcheck.Defaults()
	)

	// we're doing conversion to avoid having
	// store interface exposed or generated inside app package
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

	// Activity log for system resources
	{
		l := log
		if !c.Discovery.Debug {
			l = zap.NewNop()
		}

		DefaultResourceActivity := discoveryService.ResourceActivity(l, c.Discovery, DefaultStore, eventbus.Service())
		err = DefaultResourceActivity.InitResourceActivityLog(ctx, []string{
			// (types.User{}).RbacResource(), // @todo user?? suppose to be system:user
			"system:user",
		})
		if err != nil {
			return err
		}
	}

	sassTranspiler := dartSassTranspiler(log)

	DefaultAccessControl = AccessControl(s)
	CurrentSettings.Auth.Internal.Enabled = true
	CurrentSettings.Auth.Internal.Signup.Enabled = true
	CurrentSettings.AI.Enabled = true
	DefaultSettings = Settings(ctx, DefaultStore, DefaultLogger, DefaultAccessControl, DefaultActionlog, CurrentSettings, c.Webapps)
	wireChatAIConfig()
	DefaultStylesheet = Stylesheet(sassTranspiler, log)

	DefaultDalConnection = Connection(ctx, dal.Service(), c.DB)

	DefaultDalSensitivityLevel = SensitivityLevel(ctx, dal.Service())

	DefaultDalSchemaAlteration = NewDalSchemaAlteration(dal.Service())

	if DefaultObjectStore == nil {
		var (
			opt    = c.Storage
			bucket string
		)
		const svcPath = "system"
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

		hcd.Add(objstore.Healthcheck(DefaultObjectStore), "ObjectStore/System")

		if err != nil {
			return err
		}

	}

	DefaultRenderer = Renderer(c.Template)
	DefaultResourceTranslation = ResourceTranslation()
	DefaultAuthNotification = AuthNotification(CurrentSettings, DefaultRenderer, c.Auth)
	DefaultAuth = Auth(AuthOptions{LimitUsers: c.Limit.SystemUsers})
	DefaultAuthClient = AuthClient(DefaultStore, DefaultAccessControl, DefaultActionlog, eventbus.Service(), c.Auth)
	DefaultAttachment = Attachment(DefaultObjectStore, c.Attachment, DefaultLogger)
	DefaultUser = User(UserOptions{LimitUsers: c.Limit.SystemUsers})
	DefaultCredentials = Credentials()
	DefaultReport = Report(DefaultStore, DefaultAccessControl, DefaultActionlog, eventbus.Service())
	DefaultRole = Role(rbac.Global())
	DefaultUserGroup = UserGroup(rbac.Global())
	DefaultApplication = Application(DefaultStore, DefaultAccessControl, DefaultActionlog, eventbus.Service())
	DefaultReminder = Reminder(ctx, DefaultLogger.Named("reminder"), ws)
	DefaultNotification = Notification(ctx, DefaultLogger.Named("notification"), ws)
	DefaultSink = Sink()
	DefaultStatistics = Statistics()
	DefaultQueue = Queue()
	DefaultApigwRoute = Route()
	DefaultApigwProfiler = Profiler()
	DefaultApigwFilter = Filter()
	DefaultDataPrivacy = DataPrivacy(DefaultStore, DefaultAccessControl, DefaultActionlog, eventbus.Service())
	DefaultSMTPChecker = SmtpConfigurationChecker(CurrentSettings, DefaultRenderer, DefaultAccessControl, c.Auth)
	DefaultExpression = Expression()

	if err = initRoles(ctx, log.Named("rbac.roles"), c.RBAC, eventbus.Service(), rbac.Global()); err != nil {
		return err
	}

	// Register a user resolver with the RBAC service so that contextual role
	// expressions can access user properties (email, username, handle, name, labels).
	rbac.Global().SetUserResolver(func(userID uint64) map[string]interface{} {
		u, err := store.LookupUserByID(ctx, s, userID)
		if err != nil || u == nil {
			return nil
		}

		// Load user labels
		if err = label.Load(ctx, s, u); err != nil {
			log.Warn("failed to load user labels for RBAC scope", zap.Uint64("userID", userID), zap.Error(err))
		}

		// Convert labels to map[string]interface{} for expression evaluation.
		// Single-value labels are exposed as strings, multi-value as string slices.
		ll := make(map[string]interface{}, len(u.Labels))
		for k, v := range u.Labels {
			if len(v.Values) > 0 {
				ll[k] = v.Values
			} else {
				ll[k] = v.Val
			}
		}

		return map[string]interface{}{
			"email":    u.Email,
			"username": u.Username,
			"handle":   u.Handle,
			"name":     u.Name,
			"labels":   ll,
		}
	})

	automationService.DefaultUser = DefaultUser

	automationService.Registry().AddTypes(
		automation.User{},
		automation.Role{},
		automation.Template{},
		automation.RenderOptions{},
		automation.RenderedDocument{},
		automation.RbacResource{},
		automation.Reminder{},
	)

	automation.UsersHandler(
		automationService.Registry(),
		DefaultUser,
		DefaultRole,
	)

	automation.TemplatesHandler(
		automationService.Registry(),
		DefaultRenderer,
	)

	automation.RolesHandler(
		automationService.Registry(),
		DefaultRole,
		DefaultUser,
	)
	automation.RemindersHandler(
		automationService.Registry(),
		DefaultReminder,
	)

	automation.RbacHandler(
		automationService.Registry(),
		rbac.Global(),
		DefaultUser,
		DefaultRole,
	)

	// Register notification handler
	automation.NotificationHandler(
		automationService.Registry(),
		DefaultNotification,
		DefaultUser,
		log,
	)

	// ValuestoreHandler isn't (yet) a system thing but this initialization resides
	// here just so we can easily register it

	automation.ValuestoreHandler(
		automationService.Registry(),
		valuestore.Global(),
	)

	if c.ActionLog.WorkflowFunctionsEnabled {
		// register action-log functions & types only when enabled
		automation.ActionlogHandler(
			automationService.Registry(),
			DefaultActionlog,
		)

		automationService.Registry().AddTypes(
			automation.Action{},
		)
	}

	// Reload DAL sensitivity levels
	err = DefaultDalSensitivityLevel.ReloadSensitivityLevels(ctx, DefaultStore)
	if err != nil {
		return
	}

	// Reload DAL connections
	err = DefaultDalConnection.ReloadConnections(ctx)
	if err != nil {
		return
	}

	return
}

func Watchers(ctx context.Context) {
	DefaultReminder.Watch(ctx)
	return
}

func Activate(ctx context.Context) (err error) {
	// Run initial update of current settings
	err = DefaultSettings.UpdateCurrent(ctx)
	if err != nil {
		return
	}

	err = DefaultUserGroup.Activate(ctx)
	if err != nil {
		return
	}

	return
}

// isGeneric returns true if given error is generic
func isGeneric(err error) bool {
	g, ok := err.(interface{ IsGeneric() bool })
	return ok && g != nil && g.IsGeneric()
}

// unwrapGeneric unwraps error if error is generic (and wrapped)
func unwrapGeneric(err error) error {
	for {
		if isGeneric(err) {
			err = errors.Unwrap(err)
			continue
		}

		return err
	}
}

// Data is stale when new date does not match updatedAt or createdAt (before first update)
//
// @todo This is the same as in compose.service; do we want to make an util thing?
func isStale(new *time.Time, updatedAt *time.Time, createdAt time.Time) bool {
	if new == nil {
		// Change to true for stale-data-check
		return false
	}

	if updatedAt != nil {
		return !new.Equal(*updatedAt)
	}

	return new.Equal(createdAt)
}

func dartSassTranspiler(log *zap.Logger) *godartsass.Transpiler {
	transpiler, err := godartsass.Start(godartsass.Options{
		DartSassEmbeddedFilename: "sass",
	})

	if err != nil {
		log.Warn("dart sass is not installed in your system", zap.Error(err))
		return nil
	}

	return transpiler
}

func wireChatAIConfig() {
	chat.SetConfigProvider(func() chat.Config {
		s := CurrentSettings.AI
		cfg := chat.Config{
			Enabled:   s.Enabled,
			OllamaURL: s.OllamaURL,
			Roles: chat.RoleModels{
				ComposeChat:    s.Roles.ComposeChat,
				MCPAgent:       s.Roles.MCPAgent,
				AutomationChat: s.Roles.AutomationChat,
				RulesgoAI:      s.Roles.RulesgoAI,
			},
		}
		if len(s.Catalog) > 0 {
			cfg.Catalog = make([]chat.CatalogEntry, 0, len(s.Catalog))
			for _, e := range s.Catalog {
				cfg.Catalog = append(cfg.Catalog, chat.CatalogEntry{
					Name:    e.Name,
					Enabled: e.Enabled,
					Label:   e.Label,
					Note:    e.Note,
				})
			}
		}
		return cfg
	})

	aiagent.SetExtrasProvider(func() []aiagent.AgentSpec {
		return agentSpecsFromSettings(CurrentSettings.AI.Agents)
	})
	aiagent.SetConnectorsProvider(func() []aiagent.Connector {
		return toolkitConnectorsFromSettings(CurrentSettings.AI.Toolkits)
	})
	if DefaultSettings != nil {
		DefaultSettings.Register("ai.", func(ctx context.Context, current interface{}, _ types.SettingValueSet) {
			if cat := aiagent.DefaultCatalog(); cat != nil {
				cat.RefreshRemotesNow(ctx)
			}
			if reg := aiagent.DefaultRegistry(); reg != nil {
				reg.Reload(nil)
			}
		})
	}
}

func toolkitConnectorsFromSettings(entries []types.AIToolkitEntry) []aiagent.Connector {
	if len(entries) == 0 {
		return nil
	}
	out := make([]aiagent.Connector, 0, len(entries))
	for _, e := range entries {
		c := aiagent.Connector{
			Handle: strings.TrimSpace(e.Handle),
			URL:    strings.TrimSpace(e.URL),
			Token:  strings.TrimSpace(e.Token),
			Source: "settings",
		}
		if e.Enabled != nil {
			v := *e.Enabled
			c.Enabled = &v
		}
		if c.Handle == "" || aiagent.ReservedKitName(c.Handle) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func agentSpecsFromSettings(entries []types.AIAgentEntry) []aiagent.AgentSpec {
	if len(entries) == 0 {
		return nil
	}
	out := make([]aiagent.AgentSpec, 0, len(entries))
	for _, e := range entries {
		s := aiagent.AgentSpec{
			Handle:      strings.TrimSpace(e.Handle),
			Description: e.Description,
			Prompt:      e.Prompt,
			Model:       e.Model,
			Toolkits:    append([]string(nil), e.Toolkits...),
			MaxSteps:    e.MaxSteps,
			Confirm:     e.Confirm,
			Source:      "settings",
		}
		if e.Enabled != nil {
			v := *e.Enabled
			s.Enabled = &v
		}
		if s.Handle == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
