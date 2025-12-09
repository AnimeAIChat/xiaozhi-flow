package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	domainimage "xiaozhi-server-go/internal/domain/image"
	domainauth "xiaozhi-server-go/internal/domain/auth"
	domainmcp "xiaozhi-server-go/internal/domain/mcp"
	domainllm "xiaozhi-server-go/internal/domain/llm"
	llminfra "xiaozhi-server-go/internal/domain/llm/infrastructure"
	llmrepo "xiaozhi-server-go/internal/domain/llm/repository"
	"xiaozhi-server-go/internal/plugin/builtin/openai"
	"xiaozhi-server-go/internal/plugin/capability"
	"xiaozhi-server-go/internal/plugin/providers/chatglm"
	"xiaozhi-server-go/internal/plugin/providers/coze"
	"xiaozhi-server-go/internal/plugin/providers/deepgram"
	"xiaozhi-server-go/internal/plugin/providers/doubao"
	"xiaozhi-server-go/internal/plugin/providers/edge"
	"xiaozhi-server-go/internal/plugin/providers/gosherpa"
	"xiaozhi-server-go/internal/plugin/providers/ollama"
	"xiaozhi-server-go/internal/plugin/providers/stepfun"
	authstore "xiaozhi-server-go/internal/domain/auth/store"
	configmanager "xiaozhi-server-go/internal/domain/config/manager"
	"xiaozhi-server-go/internal/domain/config/types"
	"xiaozhi-server-go/internal/domain/device/service"
	"xiaozhi-server-go/internal/domain/device/repository"
	"xiaozhi-server-go/internal/domain/eventbus"
	platformerrors "xiaozhi-server-go/internal/platform/errors"
	platformlogging "xiaozhi-server-go/internal/platform/logging"
	platformobservability "xiaozhi-server-go/internal/platform/observability"
	platformstorage "xiaozhi-server-go/internal/platform/storage"
	platformconfig "xiaozhi-server-go/internal/platform/config"
	httptransport "xiaozhi-server-go/internal/transport/http"
	httpvision "xiaozhi-server-go/internal/transport/http/vision"
	httpwebapi "xiaozhi-server-go/internal/transport/http/webapi"
	httpota "xiaozhi-server-go/internal/transport/http/ota"
	httpstartup "xiaozhi-server-go/internal/transport/http/startup"
	authv1 "xiaozhi-server-go/internal/transport/http/v1"
	systemv1 "xiaozhi-server-go/internal/transport/http/v1"
	devicev1 "xiaozhi-server-go/internal/transport/http/v1"
	"xiaozhi-server-go/internal/core/transport"
	"xiaozhi-server-go/internal/contracts/adapters"
	"xiaozhi-server-go/internal/contracts/config/integration"
	"xiaozhi-server-go/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
	"golang.org/x/sync/errgroup"

	// 注意：移除了对src/core的直接依赖，将通过适配器层来访问
	// 提供者注册将延迟到第二阶段进行
)

const scalarHTML = `<!DOCTYPE html>
<html lang="zh-CN">
	<head>
		<meta charset="utf-8" />
		<title>小智 API Reference</title>
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script
			id="api-reference"
			data-url="/openapi.json"
			data-layout="modern"
			src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"
		></script>
	</body>
</html>`

type stepFn func(context.Context, *appState) error

type initStep struct {
	ID        string
	Title     string
	DependsOn []string
	Kind      platformerrors.Kind
	Execute   stepFn
}

type appState struct {
	config                *platformconfig.Config
	configPath            string
	configRepo            types.Repository
	logProvider           *platformlogging.Logger
	logger                *utils.Logger
	slogger               *slog.Logger
	observabilityShutdown platformobservability.ShutdownFunc
	authManager           *domainauth.AuthManager
	domainMCPManager      *domainmcp.Manager   // New domain MCP manager
	bootstrapManager      *adapters.BootstrapManager // 新增：引导管理器
	componentContainer    *adapters.ComponentContainer // 新增：组件容器
	configIntegrator      *integration.ConfigIntegrator   // 新增：配置集成器
	llmManager            llmrepo.LLMRepository // 新增：LLM管理器
	llmService            domainllm.Service     // 新增：LLM服务
	registry              *capability.Registry  // 新增：插件注册表
}

// Run 启动整个服务生命周期，负责加载配置、初始化依赖和优雅关停。
func Run(ctx context.Context) error {
	state := &appState{}

	steps := InitGraph()
	if err := executeInitSteps(ctx, steps, state); err != nil {
		return err
	}

	config := state.config
	logger := state.logger
	if config == nil || logger == nil {
		return platformerrors.Wrap(
			platformerrors.KindBootstrap,
			"bootstrap state validation",
			"config/logger not initialised",
			errors.New("config/logger not initialised"),
		)
	}

	authManager := state.authManager
	if authManager == nil {
		return platformerrors.Wrap(
			platformerrors.KindBootstrap,
			"bootstrap state validation",
			"auth manager not initialised",
			errors.New("auth manager not initialised"),
		)
	}

	domainMCPManager := state.domainMCPManager
	if domainMCPManager == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"bootstrap state validation",
			"domain MCP manager not initialised",
		)
	}

	logBootstrapGraph(steps, logger)

	if shutdown := state.observabilityShutdown; shutdown != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(shutdownCtx); err != nil {
				logger.WarnTag("引导", "可观测性未正常关闭: %v", err)
			}
		}()
	}

	defer func() {
		if authManager != nil {
			if closeErr := authManager.Close(); closeErr != nil {
				logger.ErrorTag("认证", "认证管理器未正常关闭: %v", closeErr)
			}
		}
	}()

	rootCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	signalCtx, stop := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	group, groupCtx := errgroup.WithContext(rootCtx)

	if err := startServices(state.config, logger, authManager, state.configRepo, state.domainMCPManager, state.componentContainer, state.registry, group, groupCtx); err != nil {
		cancel()
		return err
	}

	if err := waitForShutdown(signalCtx, cancel, logger, group); err != nil {
		return err
	}

	logger.InfoTag("引导", "服务已成功启动")
	logger.Close()
	return nil
}

func logBootstrapGraph(steps []initStep, logger *utils.Logger) {
	if logger == nil {
		return
	}
	logger.InfoTag("引导", "初始化依赖关系概览")

	// 阶段名称映射
	stepNames := map[string]string{
		"storage:init-config-store": "初始化配置存储",
		"storage:init-database":     "初始化数据库",
		"config:load-default":       "加载默认配置",
		"logging:init-provider":     "初始化日志提供者",
		"mcp:init-manager":          "初始化MCP管理器",
		"observability:setup-hooks": "设置可观测性钩子",
		"components:init-container": "初始化组件容器",
		"config:init-integrator":    "初始化配置集成器",
		"auth:init-manager":         "初始化认证管理器",
		"plugin:init-manager":       "初始化插件管理器",
	}

	for _, step := range steps {
		if name, ok := stepNames[step.ID]; ok {
			logger.InfoTag("引导", name)
		}
	}
	logger.InfoTag("引导", "启动服务")
}

func executeInitSteps(ctx context.Context, steps []initStep, state *appState) error {
	if state == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"execute init steps",
			"nil bootstrap state",
		)
	}

	completed := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		for _, dep := range step.DependsOn {
			if _, ok := completed[dep]; !ok {
				return platformerrors.New(
					platformerrors.KindBootstrap,
					step.ID,
					fmt.Sprintf("dependency %s not satisfied", dep),
				)
			}
		}
		if step.Execute == nil {
			return platformerrors.New(
				platformerrors.KindBootstrap,
				step.ID,
				"missing execute function",
			)
		}
		if err := step.Execute(ctx, state); err != nil {
			var typed *platformerrors.Error
			if errors.As(err, &typed) {
				return err
			}

			kind := step.Kind
			if kind == "" {
				kind = platformerrors.KindBootstrap
			}
			return platformerrors.Wrap(kind, step.ID, "bootstrap step failed", err)
		}
		completed[step.ID] = struct{}{}
	}
	return nil
}

func InitGraph() []initStep {
	return []initStep{
		{
			ID:      "storage:init-config-store",
			Title:   "Initialise configuration store",
			Kind:    platformerrors.KindStorage,
			Execute: initStorageStep,
		},
		{
			ID:      "storage:init-database",
			Title:   "Initialise database",
			Kind:    platformerrors.KindStorage,
			Execute: initDatabaseStep,
		},
		{
			ID:        "config:load-default",
			Title:     "Load configuration from database",
			DependsOn: []string{"storage:init-config-store", "storage:init-database"},
			Kind:      platformerrors.KindConfig,
			Execute:   loadDefaultConfigStep,
		},
		{
			ID:        "llm:init-manager",
			Title:     "Initialise LLM manager",
			DependsOn: []string{"config:load-default"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initLLMManagerStep,
		},
		{
			ID:        "logging:init-provider",
			Title:     "Initialise logging provider",
			DependsOn: []string{"config:load-default"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initLoggingStep,
		},
		{
			ID:        "mcp:init-manager",
			Title:     "Initialise MCP manager",
			DependsOn: []string{"logging:init-provider"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initMCPManagerStep,
		},
		{
			ID:        "observability:setup-hooks",
			Title:     "Setup observability hooks",
			DependsOn: []string{"logging:init-provider"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   setupObservabilityStep,
		},
		{
			ID:        "components:init-container",
			Title:     "Initialise component container",
			DependsOn: []string{"logging:init-provider"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initComponentsStep,
		},
		{
			ID:        "config:init-integrator",
			Title:     "Initialise config integrator",
			DependsOn: []string{"logging:init-provider", "components:init-container"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initConfigIntegratorStep,
		},
		{
			ID:        "auth:init-manager",
			Title:     "Initialise auth manager",
			DependsOn: []string{"observability:setup-hooks", "storage:init-database", "components:init-container"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initAuthStep,
		},
	}
}

func initLLMManagerStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"llm:init-manager",
			"config not loaded",
		)
	}

	// Initialize Capability Registry
	registry := capability.NewRegistry()
	
	// Register Builtin Plugins
	// In the future, this could be dynamic or loaded from external plugins
	openaiProvider := openai.NewProvider()
	registry.Register("openai", openaiProvider)

	// Register Vendor Plugins
	registry.Register("chatglm", chatglm.NewProvider())
	registry.Register("coze", coze.NewProvider())
	registry.Register("deepgram", deepgram.NewProvider())
	registry.Register("doubao", doubao.NewProvider())
	registry.Register("edge", edge.NewProvider())
	registry.Register("gosherpa", gosherpa.NewProvider())
	registry.Register("ollama", ollama.NewProvider())
	registry.Register("stepfun", stepfun.NewProvider())

	state.registry = registry

	manager, err := llminfra.NewLLMManager(state.config, registry)
	if err != nil {
		return platformerrors.Wrap(platformerrors.KindBootstrap, "llm:init-manager", "failed to create LLM manager", err)
	}

	state.llmManager = manager
	state.llmService = domainllm.NewService(manager)
	
	if state.logger != nil {
		state.logger.InfoTag("引导", "LLM管理器初始化完成 (Plugin System Enabled)")
	}

	return nil
}

func initStorageStep(_ context.Context, _ *appState) error {
	if err := platformstorage.InitConfigStore(); err != nil {
		return platformerrors.Wrap(platformerrors.KindStorage, "storage:init-config-store", "failed to initialize config store", err)
	}
	return nil
}

func initDatabaseStep(_ context.Context, _ *appState) error {
	// 尝试从 db.json 读取数据库配置
	configManager := platformstorage.NewDatabaseConfigManager()

	if configManager.Exists() {
		// 如果 db.json 存在，使用配置文件初始化数据库
		dbConfig, err := configManager.LoadConfig()
		if err != nil {
			return platformerrors.Wrap(platformerrors.KindConfig, "storage:init-database-config", "failed to load database config", err)
		}

		// 如果配置文件标记为已初始化，只连接数据库而不重新初始化
		if dbConfig.Initialized {
			if err := platformstorage.ConnectDatabaseWithConfig(dbConfig.Database); err != nil {
				// 如果连接失败，可能数据库文件被删除，需要重新初始化
				return platformerrors.Wrap(platformerrors.KindStorage, "storage:init-database", "database marked as initialized but connection failed, may need reinitialization", err)
			}
		} else {
			// 如果配置文件标记为未初始化，进行完整初始化
			if err := platformstorage.InitDatabaseWithConfig(dbConfig.Database); err != nil {
				return platformerrors.Wrap(platformerrors.KindStorage, "storage:init-database", "failed to initialize database with config", err)
			}
		}
	} else {
		// 如果 db.json 不存在，使用原有的初始化逻辑
		if err := platformstorage.InitDatabase(); err != nil {
			return platformerrors.Wrap(platformerrors.KindStorage, "storage:init-database", "failed to initialize database", err)
		}
	}

	return nil
}

func loadDefaultConfigStep(_ context.Context, state *appState) error {
	// Create database-backed config repository
	configRepo := configmanager.NewDatabaseRepository(platformstorage.GetDB())
	state.configRepo = configRepo

	// Load configuration from database, fallback to defaults if not found
	config, err := configRepo.LoadConfig()
	if err != nil {
		return platformerrors.Wrap(platformerrors.KindConfig, "config:load-default", "failed to load config from database", err)
	}

	state.config = config
	state.configPath = "database:config"
	return nil
}

func initLoggingStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"logging:init-provider",
			"config not loaded",
		)
	}

	logProvider, err := platformlogging.New(platformlogging.Config{
		Level:    state.config.Log.Level,
		Dir:      state.config.Log.Dir,
		Filename: state.config.Log.File,
	})
	if err != nil {
		return platformerrors.Wrap(platformerrors.KindBootstrap, "logging:init-provider", "failed to initialize logging provider", err)
	}

	state.logProvider = logProvider
	state.logger = logProvider.Legacy()
	state.slogger = logProvider.Slog()
	utils.DefaultLogger = state.logger

	if state.logger != nil {
		state.logger.InfoTag(
			"引导",
			"日志模块就绪 [%s] %s",
			state.config.Log.Level,
			state.configPath,
		)
	}

	// 设置事件处理器
	eventbus.SetupEventHandlers()

	return nil
}

func setupObservabilityStep(ctx context.Context, state *appState) error {
	if state == nil || state.logger == nil || state.config == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"observability:setup-hooks",
			"config/logger not initialised",
		)
	}

	slogger := state.slogger
	if slogger == nil && state.logger != nil {
		slogger = state.logger.Slog()
	}

	cfg := platformobservability.Config{
		Enabled: strings.EqualFold(state.config.Log.Level, "debug"),
	}

	shutdown, err := platformobservability.Setup(ctx, cfg, slogger)
	if err != nil {
		return platformerrors.Wrap(platformerrors.KindBootstrap, "observability:setup-hooks", "failed to setup observability hooks", err)
	}
	state.observabilityShutdown = shutdown

	return nil
}

func initComponentsStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil || state.logger == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"components:init-container",
			"missing config/logger",
		)
	}

	// 创建引导管理器
	state.bootstrapManager = adapters.NewBootstrapManager(state.config, state.logger)

	// 初始化组件容器
	container, err := state.bootstrapManager.InitializeComponents()
	if err != nil {
		return platformerrors.Wrap(
			platformerrors.KindBootstrap,
			"components:init-container",
			"failed to initialize component container",
			err,
		)
	}

	state.componentContainer = container

	if state.logger != nil {
		state.logger.InfoTag("引导", "组件容器初始化完成")
	}

	return nil
}

func initConfigIntegratorStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil || state.logger == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"config:init-integrator",
			"missing config/logger",
		)
	}

	// 创建配置集成器
	configIntegrator, err := integration.NewConfigIntegrator(state.config, state.logger)
	if err != nil {
		return platformerrors.Wrap(
			platformerrors.KindBootstrap,
			"config:init-integrator",
			"failed to create config integrator",
			err,
		)
	}

	// 初始化配置集成器
	if err := configIntegrator.Initialize(context.Background()); err != nil {
		return platformerrors.Wrap(
			platformerrors.KindBootstrap,
			"config:init-integrator",
			"failed to initialize config integrator",
			err,
		)
	}

	state.configIntegrator = configIntegrator

	if state.logger != nil {
		state.logger.InfoTag("引导", "配置集成器初始化完成")
	}

	return nil
}

func initAuthStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil || state.logger == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"auth:init-manager",
			"missing config/logger",
		)
	}

	authManager, err := initAuthManager(state.config, state.logger)
	if err != nil {
		return err
	}
	state.authManager = authManager
	return nil
}

func initMCPManagerStep(_ context.Context, state *appState) error {
	if state == nil || state.config == nil || state.logger == nil {
		return platformerrors.New(
			platformerrors.KindBootstrap,
			"mcp:init-manager",
			"missing config/logger",
		)
	}

	// Initialize global MCP tools using the new unified manager
	globalMCPManager := domainmcp.GetGlobalMCPManager()
	if err := globalMCPManager.Initialize(state.config, state.logger); err != nil {
		state.logger.WarnTag("引导", "全局MCP管理器初始化失败: %v", err)
		// Continue anyway, local tools will still work
	}

	// Create domain manager directly from config (will automatically use global clients)
	domainManager, err := domainmcp.NewFromConfig(state.config, state.logger)
	if err != nil {
		return platformerrors.Wrap(platformerrors.KindBootstrap, "mcp:init-manager", "failed to create domain MCP manager", err)
	}

	state.domainMCPManager = domainManager
	state.logger.InfoTag("引导", "MCP管理器初始化完成（使用统一全局管理器）")

	return nil
}



func initAuthManager(config *platformconfig.Config, logger *utils.Logger) (*domainauth.AuthManager, error) {
	storeType := strings.ToLower(strings.TrimSpace(config.Server.Auth.Store.Type))
	storeCfg := authstore.Config{
		Driver: storeType,
		TTL:    config.Server.Auth.Store.Expiry,
	}

	if storeCfg.Driver == "" || storeCfg.Driver == "database" || storeCfg.Driver == "sqlite" {
		storeCfg.Driver = authstore.DriverSQLite
	}

	cleanupInterval := config.Server.Auth.Store.Cleanup
	if cleanupInterval <= 0 {
		cleanupInterval = 10 * time.Minute // default cleanup interval
	}

	switch storeCfg.Driver {
	case authstore.DriverMemory:
		if config.Server.Auth.Store.Memory.Cleanup > 0 {
			cleanupInterval = config.Server.Auth.Store.Memory.Cleanup
		}
		storeCfg.Memory = &authstore.MemoryConfig{
			GCInterval: cleanupInterval,
		}
	case authstore.DriverSQLite:
		storeCfg.SQLite = &authstore.SQLiteConfig{
			DSN: config.Server.Auth.Store.SQLite.DSN,
		}
	case authstore.DriverRedis:
		storeCfg.Redis = &authstore.RedisConfig{
			Addr:     config.Server.Auth.Store.Redis.Addr,
			Username: config.Server.Auth.Store.Redis.Username,
			Password: config.Server.Auth.Store.Redis.Password,
			DB:       config.Server.Auth.Store.Redis.DB,
			Prefix:   config.Server.Auth.Store.Redis.Prefix,
		}
		if storeCfg.Redis.Addr == "" {
			return nil, platformerrors.Wrap(
				platformerrors.KindBootstrap,
				"auth:init-manager",
				"redis store addr is required",
				errors.New("redis store addr is required"),
			)
		}
	default:
		logger.WarnTag("认证", "不支持的存储类型 %s，已自动回退至内存模式", storeType)
		storeCfg.Driver = authstore.DriverMemory
		storeCfg.Memory = &authstore.MemoryConfig{GCInterval: cleanupInterval}
	}

	storeDeps := authstore.Dependencies{
		SQLiteDB: platformstorage.GetDB(),
	}
	authStore, err := authstore.New(storeCfg, storeDeps)
	if err != nil {
		return nil, platformerrors.Wrap(platformerrors.KindBootstrap, "auth:init-manager", "failed to create auth store", err)
	}

	crypto := domainauth.NewMemoryCryptoManager(logger, storeCfg.TTL)
	opts := domainauth.Options{
		Store:           authStore,
		Logger:          logger,
		Crypto:          crypto,
		SessionTTL:      storeCfg.TTL,
		CleanupInterval: cleanupInterval,
	}

	authManager, err := domainauth.NewManager(opts)
	if err != nil {
		return nil, platformerrors.Wrap(platformerrors.KindBootstrap, "auth:init-manager", "failed to create auth manager", err)
	}

	return authManager, nil
}

func parseDurationOrWarn(logger *utils.Logger, value string, field string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.WarnTag("配置", "无法解析 %s，原始值 %s（%v）", field, value, err)
		return 0
	}
	if duration <= 0 {
		logger.WarnTag("配置", "%s 必须为正数，当前值为 %s", field, value)
		return 0
	}
	return duration
}

func startTransportServer(
	config *platformconfig.Config,
	logger *utils.Logger,
	authManager *domainauth.AuthManager,
	domainMCPManager *domainmcp.Manager,
	componentContainer *adapters.ComponentContainer,
	deviceRepo repository.DeviceRepository,
	registry *capability.Registry,
	g *errgroup.Group,
	groupCtx context.Context,
) (adapters.TransportManager, error) {
	// 使用适配器来避免循环依赖
	if componentContainer == nil {
		return nil, fmt.Errorf("component container is required")
	}

	// 获取旧版适配器
	legacyAdapter := componentContainer.GetLegacyAdapter()

	// 创建传输适配器
	transportAdapter := adapters.NewTransportAdapter(config, logger, legacyAdapter, deviceRepo, registry)

	// 创建真正的传输管理器
	transportManager := transport.NewTransportManager(config, logger)

	// 注册WebSocket传输
	if wsTransport := transportAdapter.GetWebSocketTransport(); wsTransport != nil {
		transportManager.RegisterTransport("websocket", wsTransport)
	}

	// 启动传输服务器
	if err := transportAdapter.StartTransportServer(groupCtx, authManager, domainMCPManager); err != nil {
		return nil, platformerrors.Wrap(
			platformerrors.KindTransport,
			"transport:start-server",
			"failed to start transport server",
			err,
		)
	}

	// 启动传输管理器
	g.Go(func() error {
		go func() {
			<-groupCtx.Done()
			logger.InfoTag("传输", "收到关闭信号，正在关闭传输服务器")
			if err := transportAdapter.StopTransportServer(); err != nil {
				logger.ErrorTag("传输", "关闭传输服务器失败: %v", err)
			} else {
				logger.InfoTag("传输", "传输服务器已优雅关闭")
			}
		}()

		// 启动传输管理器
		if err := transportManager.Start(groupCtx); err != nil {
			if groupCtx.Err() != nil {
				return nil
			}
			logger.ErrorTag("传输", "传输管理器运行失败: %v", err)
			return err
		}
		return nil
	})

	logger.InfoTag("传输", "传输服务已成功启动（适配器模式）")
	return transportManager, nil
}

func startHTTPServer(
	config *platformconfig.Config,
	logger *utils.Logger,
	configRepo types.Repository,
	transportManager adapters.TransportManager,
	deviceRepo repository.DeviceRepository,
	g *errgroup.Group,
	groupCtx context.Context,
) (*http.Server, error) {
	// 首先初始化webapi服务以获取认证中间件
	webapiService, err := httpwebapi.NewService(config, logger)
	if err != nil {
		logger.ErrorTag("WebAPI", "WebAPI 服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "webapi:new-service", "failed to create webapi service", err)
	}

	// 构建HTTP路由器，传入认证中间件
	httpRouter, err := httptransport.Build(httptransport.Options{
		Config:         config,
		Logger:         logger,
		AuthMiddleware: webapiService.AuthMiddleware(),
	})
	if err != nil {
		return nil, err
	}
	router := httpRouter.Engine
	apiGroup := httpRouter.API

	// 初始化设备服务
	db := platformstorage.GetDB()
	verificationRepo := platformstorage.NewVerificationCodeRepository(db)

	deviceService := service.NewDeviceService(
		deviceRepo,
		verificationRepo,
		config.Server.Device.RequireActivationCode,
		int(config.Server.Device.DefaultAdminUserID),
	)

	// 初始化图像处理管道
	imagePipeline, err := domainimage.NewPipeline(domainimage.Options{
		Security: &platformconfig.SecurityConfig{
			MaxFileSize:       5 * 1024 * 1024, // 5MB
			MaxPixels:         16777216,        // 16M pixels
			MaxWidth:          4096,
			MaxHeight:         4096,
			AllowedFormats:    []string{"jpeg", "jpg", "png", "webp", "gif"},
			EnableDeepScan:    true,
			ValidationTimeout: "10s",
		},
		Logger: logger,
	})
	if err != nil {
		return nil, platformerrors.Wrap(platformerrors.KindBootstrap, "http:init-image-pipeline", "failed to create image pipeline", err)
	}

	// 初始化新的HTTP服务
	visionService, err := httpvision.NewService(config, logger, imagePipeline)
	if err != nil {
		logger.ErrorTag("视觉", "Vision 服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindVision, "vision:new-service", "failed to create vision service", err)
	}

	otaService, err := httpota.NewService(config.Web.Websocket, config, deviceService, logger)
	if err != nil {
		logger.ErrorTag("OTA", "OTA 服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "ota:new-service", "failed to create ota service", err)
	}

	startupService, err := httpstartup.NewService(config, logger)
	if err != nil {
		logger.ErrorTag("Startup", "启动流程服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "startup:new-service", "failed to create startup service", err)
	}

	// 初始化V1认证服务
	authServiceV1, err := authv1.NewAuthServiceV1(config, logger)
	if err != nil {
		logger.ErrorTag("API", "V1认证服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "auth-v1:new-service", "failed to create auth v1 service", err)
	}

	// 初始化V1系统服务
	systemServiceV1, err := systemv1.NewSystemServiceV1(config, logger)
	if err != nil {
		logger.ErrorTag("API", "V1系统服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "system-v1:new-service", "failed to create system v1 service", err)
	}

	// 初始化V1设备服务
	deviceServiceV1, err := devicev1.NewDeviceServiceV1(config, logger, transportManager)
	if err != nil {
		logger.ErrorTag("API", "V1设备服务初始化失败: %v", err)
		return nil, platformerrors.Wrap(platformerrors.KindTransport, "device-v1:new-service", "failed to create device v1 service", err)
	}

	// 注册服务路由
	visionService.Register(groupCtx, apiGroup)
	webapiService.Register(groupCtx, apiGroup)
	otaService.Register(groupCtx, apiGroup)
	startupService.Register(groupCtx, apiGroup)

	// 注册V1 API服务路由
	authServiceV1.Register(httpRouter.V1)        // 公开的认证接口

	// 如果有认证中间件，注册需要认证的接口到V1Secure
	if httpRouter.V1Secure != nil {
		authServiceV1.RegisterSecure(httpRouter.V1Secure) // 需要认证的接口
		systemServiceV1.Register(httpRouter.V1Secure)     // 系统管理需要认证
		deviceServiceV1.Register(httpRouter.V1Secure)     // 设备管理需要认证
	} else {
		// 没有认证中间件时，注册到普通V1路由
		systemServiceV1.Register(httpRouter.V1)
		deviceServiceV1.Register(httpRouter.V1)
	}

	// Note: System config service removed as we no longer use database-backed configuration

	// 自动分配可用端口
	port, err := utils.GetAvailablePort(config.Web.Port)
	if err != nil {
		logger.ErrorTag("HTTP", "自动分配端口失败: %v", err)
		port = config.Web.Port // 如果自动分配失败，使用配置的端口
	} else {
		logger.InfoTag("HTTP", "自动分配端口: %d", port)
	}

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	router.GET("/openapi.json", func(c *gin.Context) {
		doc, err := swag.ReadDoc()
		if err != nil {
			logger.ErrorTag("HTTP", "生成 OpenAPI 文档失败: %v", err)
			c.JSON(http.StatusInternalServerError, httptransport.APIResponse{
				Success: false,
				Data:    gin.H{"error": err.Error()},
				Message: "failed to generate openapi spec",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(doc))
	})

	router.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarHTML))
	})

	g.Go(func() error {
		logger.InfoTag("HTTP", "Gin 服务已启动，访问地址 http://localhost:%d", port)
		logger.InfoTag("HTTP", "OTA 服务入口: http://localhost:%d/api/ota/", port)
		logger.InfoTag("HTTP", "在线文档入口: http://localhost:%d/docs", port)

		go func() {
			<-groupCtx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.ErrorTag("HTTP", "HTTP 服务关闭失败: %v", err)
			} else {
				logger.InfoTag("HTTP", "HTTP 服务已优雅关闭")
			}
		}()

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorTag("HTTP", "HTTP 服务启动失败: %v", err)
			return err
		}
		return nil
	})

	return httpServer, nil
}

func waitForShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	logger *utils.Logger,
	g *errgroup.Group,
) error {
	<-ctx.Done()
	logger.InfoTag("引导", "收到系统信号 %v，正在进行资源清理", context.Cause(ctx))

	cancel()

	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.ErrorTag("引导", "服务关闭过程中出现错误: %v", err)
			return err
		}
		logger.InfoTag("引导", "所有服务已成功关闭")
	case <-time.After(15 * time.Second):
		timeoutErr := errors.New("服务关闭超时")
		logger.ErrorTag("引导", "服务关闭超时，已强制退出")
		return timeoutErr
	}
	return nil
}

func startServices(
	config *platformconfig.Config,
	logger *utils.Logger,
	authManager *domainauth.AuthManager,
	configRepo types.Repository,
	domainMCPManager *domainmcp.Manager,
	componentContainer *adapters.ComponentContainer,
	registry *capability.Registry,
	g *errgroup.Group,
	groupCtx context.Context,
) error {
	// 初始化设备仓库
	db := platformstorage.GetDB()
	deviceRepo := platformstorage.NewDeviceRepository(db)

	transportManager, err := startTransportServer(config, logger, authManager, domainMCPManager, componentContainer, deviceRepo, registry, g, groupCtx)
	if err != nil {
		return fmt.Errorf("启动 Transport 服务失败: %w", err)
	}

	if _, err := startHTTPServer(config, logger, configRepo, transportManager, deviceRepo, g, groupCtx); err != nil {
		return fmt.Errorf("启动 Http 服务失败: %w", err)
	}

	return nil
}

// loadConfigAndLogger 加载配置和日志记录器（用于测试）
func loadConfigAndLogger() (*platformconfig.Config, *utils.Logger, error) {
	state := &appState{}

	// 执行必要的初始化步骤
	steps := []initStep{
		{
			ID:      "storage:init-config-store",
			Title:   "Initialise configuration store",
			Kind:    platformerrors.KindStorage,
			Execute: initStorageStep,
		},
		{
			ID:      "storage:init-database",
			Title:   "Initialise database",
			Kind:    platformerrors.KindStorage,
			Execute: initDatabaseStep,
		},
		{
			ID:        "config:load-default",
			Title:     "Load configuration from database",
			DependsOn: []string{"storage:init-config-store", "storage:init-database"},
			Kind:      platformerrors.KindConfig,
			Execute:   loadDefaultConfigStep,
		},
		{
			ID:        "logging:init-provider",
			Title:     "Initialise logging provider",
			DependsOn: []string{"config:load-default"},
			Kind:      platformerrors.KindBootstrap,
			Execute:   initLoggingStep,
		},
	}

	if err := executeInitSteps(context.Background(), steps, state); err != nil {
		return nil, nil, err
	}

	return state.config, state.logger, nil
}

// setupStaticFiles 设置静态文件服务
func setupStaticFiles(router *gin.Engine, config *platformconfig.Config) {
	// 静态文件服务已经在 router.go 中处理
	// 这里只需要设置额外的路由处理（如果需要）

	// 为SPA路由设置fallback - 必须最后设置
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是API请求，返回404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, httptransport.APIResponse{
				Success: false,
				Data:    gin.H{},
				Message: "API Not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		// 对于所有其他非静态资源路径，返回index.html（SPA支持）
		if !strings.HasPrefix(path, "/static/") &&
		   !strings.HasPrefix(path, "/assets/") &&
		   path != "/favicon.ico" {
			// 优先返回 dist 目录的 index.html
			if _, err := os.Stat("./web/dist/index.html"); err == nil {
				c.File("./web/dist/index.html")
			} else {
				// fallback 到 web 目录
				c.File("./web/index.html")
			}
		}
	})
}
