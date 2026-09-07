package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"entgo.io/ent/dialect"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/encryption"
	"github.com/school-api/school-api-v1/internal/handler"
	"github.com/school-api/school-api-v1/internal/relay"
	"github.com/school-api/school-api-v1/internal/repository"
	"github.com/school-api/school-api-v1/internal/server"
	"github.com/school-api/school-api-v1/internal/server/middleware"
	"github.com/school-api/school-api-v1/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := os.Getenv("APP_CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize structured logger.
	logLevel := parseLogLevel(cfg.Server.LogLevel)
	var logHandler slog.Handler
	if cfg.Server.Env == "development" || cfg.Server.Env == "test" {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("config loaded",
		"port", cfg.Server.Port,
		"env", cfg.Server.Env,
		"log_level", cfg.Server.LogLevel,
	)

	// Set Gin mode.
	if cfg.Server.Env == "development" || cfg.Server.Env == "test" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to PostgreSQL with timeout.
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	dbPool, err := pgxpool.New(dbCtx, cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer dbPool.Close()
	// pgxpool.New is lazy: ping once at startup so a bad DATABASE_URL fails
	// fast instead of surfacing on the first request.
	if err := dbPool.Ping(dbCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	// Connect to Redis with timeout.
	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisCancel()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}

	// Open Ent client on top of the same database. entDB is the underlying
	// *sql.DB, handed to repositories that need raw SQL (statistics reports).
	entClient, entDB, err := repository.OpenEntClientWithDB(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("open ent client: %w", err)
	}
	defer entClient.Close()

	// Auth wiring.
	adminRepo := repository.NewEntAdminRepository(entClient)
	tokenExpiry := time.Duration(cfg.JWT.ExpireHours) * time.Hour
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, tokenExpiry)
	authService := service.NewAuthService(adminRepo, tokenManager, auth.BcryptCostForEnv(cfg.Server.Env), tokenExpiry)
	authHandler := handler.NewAuthHandler(authService)

	// The encryption key is needed by both account storage and user token
	// ciphertext, so it is parsed here before either is wired.
	encryptKey, err := encryption.ParseKey(cfg.Encryption.Key)
	if err != nil {
		return fmt.Errorf("parse encryption key: %w", err)
	}

	// API key wiring.
	apiKeyLookup := repository.NewEntAPIKeyLookup(entClient)
	apiKeyValidator := auth.NewAPIKeyValidator(apiKeyLookup)
	apiKeyHandler := handler.NewAPIKeyHandler()

	// User management wiring.
	userRepo := repository.NewEntUserRepository(entClient)
	userService := service.NewUserService(userRepo, auth.BcryptCostForEnv(cfg.Server.Env))
	userHandler := handler.NewUserHandler(userService)

	// API token management wiring. The settings service feeds the default
	// quota for new tokens, so it is constructed first.
	settingRepo := repository.NewEntSettingRepository(entClient)
	settingsService := service.NewSettingsService(settingRepo)
	settingsHandler := handler.NewSettingsHandler(settingsService)

	userTokenRepo := repository.NewEntUserTokenRepository(entClient)
	userTokenService := service.NewUserTokenService(userTokenRepo, userRepo, encryptKey, settingsService)
	userTokenHandler := handler.NewUserTokenHandler(userTokenService)

	// User portal account wiring (login / change-password / token reveal).
	userAuthService := service.NewUserAuthService(userRepo, userTokenRepo, tokenManager, auth.BcryptCostForEnv(cfg.Server.Env), encryptKey, int64(tokenExpiry.Seconds()))
	userAuthHandler := handler.NewUserAuthHandler(userAuthService)

	// Department management wiring.
	departmentRepo := repository.NewEntDepartmentRepository(entClient)
	departmentService := service.NewDepartmentService(departmentRepo)
	departmentHandler := handler.NewDepartmentHandler(departmentService)

	// Platform / Account / AI Model / Group-Platform wiring.
	platformRepo := repository.NewEntPlatformRepository(entClient)
	platformService := service.NewPlatformService(platformRepo)
	platformHandler := handler.NewPlatformHandler(platformService)

	accountRepo := repository.NewEntAccountRepository(entClient)
	accountService := service.NewAccountService(accountRepo, platformRepo, encryptKey)
	accountHandler := handler.NewAccountHandler(accountService)

	aiModelRepo := repository.NewEntAIModelRepository(entClient)
	aiModelService := service.NewAIModelService(aiModelRepo)
	aiModelHandler := handler.NewAIModelHandler(aiModelService)

	groupRepo := repository.NewEntGroupRepository(entClient)
	groupService := service.NewGroupService(groupRepo)
	groupHandler := handler.NewGroupHandler(groupService)
	groupPlatformRepo := repository.NewEntGroupPlatformRepository(entClient)
	groupPlatformService := service.NewGroupPlatformService(groupPlatformRepo, groupRepo, platformRepo)
	groupPlatformHandler := handler.NewGroupPlatformHandler(groupPlatformService)

	// Billing wiring (used by relay and balance record services).
	billingRepo := repository.NewEntBillingRepository(entClient)
	billingService := service.NewBillingService(billingRepo, service.BillingConfig{DefaultMaxTokens: cfg.Billing.DefaultMaxTokens})

	// Quota request wiring.
	quotaRequestRepo := repository.NewEntQuotaRequestRepository(entClient)
	quotaRequestService := service.NewQuotaRequestService(quotaRequestRepo, userTokenRepo, settingsService)
	quotaRequestHandler := handler.NewQuotaRequestHandler(quotaRequestService)

	// Batch quota wiring (class-level / multi-user quota assignment).
	quotaBatchRepo := repository.NewEntQuotaBatchRepository(entClient)
	quotaBatchService := service.NewQuotaBatchService(quotaBatchRepo, settingsService)
	quotaBatchHandler := handler.NewQuotaBatchHandler(quotaBatchService)

	// Recharge order wiring.
	rechargeOrderRepo := repository.NewEntRechargeOrderRepository(entClient)
	rechargeOrderService := service.NewRechargeOrderService(rechargeOrderRepo, settingsService)
	rechargeOrderHandler := handler.NewRechargeOrderHandler(rechargeOrderService)

	// Balance and quota record wiring.
	balanceRecordRepo := repository.NewEntBalanceRecordRepository(entClient)
	balanceRecordService := service.NewBalanceRecordService(balanceRecordRepo, billingRepo)
	balanceRecordHandler := handler.NewBalanceRecordHandler(balanceRecordService)

	quotaRecordRepo := repository.NewEntQuotaRecordRepository(entClient)
	quotaRecordService := service.NewQuotaRecordService(quotaRecordRepo)
	quotaRecordHandler := handler.NewQuotaRecordHandler(quotaRecordService)

	// Phase 6: operational support wiring.
	notificationRepo := repository.NewEntNotificationRepository(entClient)
	notificationService := service.NewNotificationService(notificationRepo, userRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	alertRuleRepo := repository.NewEntAlertRuleRepository(entClient)
	alertRecordRepo := repository.NewEntAlertRecordRepository(entClient)
	alertService := service.NewAlertService(alertRuleRepo, alertRecordRepo, entClient)
	alertHandler := handler.NewAlertHandler(alertService)

	statsRepo := repository.NewEntStatsRepositoryWithDB(entClient, entDB, dialect.Postgres)
	statsService := service.NewStatsService(statsRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	// XLSX export wiring (usage stats, finance summary, records and orders).
	exportService := service.NewExportService(statsService, balanceRecordRepo, rechargeOrderRepo)
	exportHandler := handler.NewExportHandler(exportService)

	auditLogRepo := repository.NewEntAuditLogRepository(entClient)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService)

	reconciliationRepo := repository.NewEntReconciliationRepository(entClient)
	reconciliationService := service.NewReconciliationService(reconciliationRepo)
	reconciliationHandler := handler.NewReconciliationHandler(reconciliationService)

	// Refund request flow (user-initiated refunds, admin approval).
	refundRequestRepo := repository.NewEntRefundRequestRepository(entClient)
	refundRequestService := service.NewRefundRequestService(refundRequestRepo, reconciliationRepo)
	refundRequestHandler := handler.NewRefundRequestHandler(refundRequestService)

	// Usage guide wiring.
	guideSectionRepo := repository.NewEntGuideSectionRepository(entClient)
	guideSectionService := service.NewGuideSectionService(guideSectionRepo)
	guideSectionHandler := handler.NewGuideSectionHandler(guideSectionService)

	// AI relay wiring.
	callLogRepo := repository.NewEntCallLogRepository(entClient)
	forwarder := relay.NewForwarder(cfg.Relay.Timeout)
	relayService := service.NewRelayService(
		aiModelRepo,
		groupPlatformRepo,
		platformRepo,
		accountRepo,
		callLogRepo,
		accountService,
		billingService,
		forwarder,
		encryptKey,
		cfg.Relay.MaxBodyBytes,
		cfg.Relay.MaxRetries,
	)
	relayHandler := handler.NewRelayHandler(relayService)

	// Call-log retention sweep (requirement #7: keep 30 days, prune daily).
	logRetention := service.NewLogRetentionService(callLogRepo, cfg.Retention)
	maintenanceHandler := handler.NewMaintenanceHandler(logRetention)

	// Health checkers.
	postgresHealth := repository.NewPostgresHealthChecker(dbPool)
	redisHealth := repository.NewRedisHealthChecker(redisClient)
	healthService := service.NewHealthService(map[string]domain.HealthChecker{
		"postgres": postgresHealth,
		"redis":    redisHealth,
	})
	healthHandler := handler.NewHealthHandler(healthService)

	engine := gin.New()
	// In production only trust RFC1918 proxy hops (typical for LB / ingress
	// setups) so ClientIP cannot be spoofed via X-Forwarded-For from the open
	// internet. In dev/test trust nothing and use the direct peer address.
	if cfg.Server.Env == "production" {
		if err := engine.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}); err != nil {
			return fmt.Errorf("set trusted proxies: %w", err)
		}
	} else if err := engine.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("set trusted proxies: %w", err)
	}
	// Recovery must be the outermost middleware so panics in RequestID,
	// RequestLogger or any handler are captured instead of crashing the process.
	engine.Use(
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.RequestLogger(logger),
	)

	engine.GET("/health", healthHandler.Health)

	// Public logins, rate-limited per IP.
	engine.POST("/api/v1/admin/login", middleware.LoginRateLimit(redisClient), authHandler.Login)
	engine.POST("/api/v1/user/login", middleware.LoginRateLimit(redisClient), userAuthHandler.Login)
	// Protected admin routes.
	adminGroup := engine.Group("/api/v1/admin", middleware.JWTAuth(tokenManager), middleware.AuditMiddleware(auditLogService))
	adminGroup.GET("/me", authHandler.Me)
	adminGroup.POST("/users", userHandler.Create)
	adminGroup.GET("/users", userHandler.List)
	adminGroup.GET("/users/:id", userHandler.Get)
	adminGroup.PUT("/users/:id", userHandler.Update)
	adminGroup.PATCH("/users/:id/status", userHandler.UpdateStatus)
	// Class-roster bulk import (static segment registers alongside /users/:id).
	adminGroup.POST("/users/bulk-import", userHandler.BulkImport)
	adminGroup.POST("/users/bulk-import-file", userHandler.BulkImportFile)

	adminGroup.POST("/users/:id/tokens", userTokenHandler.Create)
	adminGroup.GET("/users/:id/tokens", userTokenHandler.List)
	adminGroup.DELETE("/users/:id/tokens/:token_id", userTokenHandler.Delete)
	adminGroup.PATCH("/users/:id/tokens/:token_id/status", userTokenHandler.UpdateStatus)

	adminGroup.POST("/departments", departmentHandler.Create)
	adminGroup.GET("/departments", departmentHandler.List)
	adminGroup.GET("/departments/:id", departmentHandler.Get)
	adminGroup.PUT("/departments/:id", departmentHandler.Update)
	adminGroup.DELETE("/departments/:id", departmentHandler.Delete)

	// Platform management.
	adminGroup.POST("/platforms", platformHandler.Create)
	adminGroup.GET("/platforms", platformHandler.List)
	adminGroup.GET("/platforms/:id", platformHandler.Get)
	adminGroup.PUT("/platforms/:id", platformHandler.Update)
	adminGroup.DELETE("/platforms/:id", platformHandler.Delete)

	// Account management.
	adminGroup.POST("/accounts", accountHandler.Create)
	adminGroup.GET("/accounts", accountHandler.List)
	adminGroup.GET("/accounts/:id", accountHandler.Get)
	adminGroup.PUT("/accounts/:id", accountHandler.Update)
	adminGroup.DELETE("/accounts/:id", accountHandler.Delete)

	// AI Model management.
	adminGroup.POST("/models", aiModelHandler.Create)
	adminGroup.GET("/models", aiModelHandler.List)
	adminGroup.GET("/models/:id", aiModelHandler.Get)
	adminGroup.PUT("/models/:id", aiModelHandler.Update)
	adminGroup.DELETE("/models/:id", aiModelHandler.Delete)

	// Group-Platform bindings.
	adminGroup.GET("/groups", groupHandler.List)
	adminGroup.POST("/groups/:id/platforms", groupPlatformHandler.Bind)
	adminGroup.GET("/groups/:id/platforms", groupPlatformHandler.List)
	adminGroup.DELETE("/groups/:id/platforms/:platform_id", groupPlatformHandler.Unbind)

	// Quota request management.
	adminGroup.GET("/quota-requests", quotaRequestHandler.ListAdmin)
	adminGroup.PATCH("/quota-requests/:id/approve", quotaRequestHandler.Approve)
	adminGroup.PATCH("/quota-requests/:id/reject", quotaRequestHandler.Reject)
	adminGroup.POST("/quota-batch", quotaBatchHandler.Apply)

	// Recharge order management.
	adminGroup.GET("/recharge-orders", rechargeOrderHandler.ListAdmin)

	// Platform settings.
	adminGroup.GET("/settings/feature-mode", settingsHandler.GetFeatureMode)
	adminGroup.PUT("/settings/feature-mode", settingsHandler.PutFeatureMode)
	adminGroup.GET("/settings/default-quota", settingsHandler.GetDefaultQuota)
	adminGroup.PUT("/settings/default-quota", settingsHandler.PutDefaultQuota)
	adminGroup.GET("/settings/quota-range", settingsHandler.GetQuotaRange)
	adminGroup.PUT("/settings/quota-range", settingsHandler.PutQuotaRange)
	adminGroup.GET("/settings/recharge", settingsHandler.GetRechargeLimits)
	adminGroup.PUT("/settings/recharge", settingsHandler.PutRechargeLimits)

	// Usage guide management.
	adminGroup.POST("/guide-sections", guideSectionHandler.Create)
	adminGroup.GET("/guide-sections", guideSectionHandler.ListAdmin)
	adminGroup.GET("/guide-sections/:id", guideSectionHandler.Get)
	adminGroup.PUT("/guide-sections/:id", guideSectionHandler.Update)
	adminGroup.DELETE("/guide-sections/:id", guideSectionHandler.Delete)

	// Balance management.
	adminGroup.POST("/users/:id/balance-adjust", balanceRecordHandler.AdminAdjust)
	adminGroup.GET("/balance-records", balanceRecordHandler.ListAdmin)
	adminGroup.GET("/quota-records", quotaRecordHandler.ListAdmin)

	// Notifications and announcements.
	adminGroup.POST("/notifications", notificationHandler.Create)
	adminGroup.GET("/notifications", notificationHandler.ListAdmin)

	// Alert rules and alert records.
	adminGroup.POST("/alert-rules", alertHandler.CreateRule)
	adminGroup.GET("/alert-rules", alertHandler.ListRules)
	adminGroup.GET("/alert-rules/:id", alertHandler.GetRule)
	adminGroup.PUT("/alert-rules/:id", alertHandler.UpdateRule)
	adminGroup.DELETE("/alert-rules/:id", alertHandler.DeleteRule)
	adminGroup.POST("/alerts/evaluate", alertHandler.Evaluate)
	adminGroup.GET("/alerts", alertHandler.ListRecords)
	adminGroup.PATCH("/alerts/:id/resolve", alertHandler.ResolveRecord)

	// Statistics.
	adminGroup.GET("/dashboard/stats", statsHandler.DashboardStats)
	adminGroup.GET("/stats/usage", statsHandler.AdminUsage)
	adminGroup.GET("/stats/rankings", statsHandler.Rankings)
	adminGroup.GET("/stats/trend", statsHandler.AdminTrend)
	adminGroup.GET("/stats/model-dist", statsHandler.ModelDist)
	adminGroup.GET("/stats/dept-dist", statsHandler.DeptDist)
	adminGroup.GET("/finance/summary", statsHandler.FinanceSummary)

	// XLSX exports (admin downloads and the user-facing balance export).
	adminGroup.GET("/export/usage", exportHandler.AdminUsageExport)
	adminGroup.GET("/export/summary", exportHandler.AdminSummaryExport)
	adminGroup.GET("/export/balance-records", exportHandler.AdminBalanceRecordsExport)
	adminGroup.GET("/export/recharge-orders", exportHandler.AdminRechargeOrdersExport)

	// Audit logs.
	adminGroup.GET("/audit-logs", auditLogHandler.ListAdmin)

	// Maintenance: manually trigger the call-log retention sweep.
	adminGroup.POST("/maintenance/call-logs/sweep", maintenanceHandler.SweepCallLogs)

	// Reconciliation and refunds.
	adminGroup.GET("/recharge-orders/:id/reconciliation", reconciliationHandler.GetReconciliation)
	adminGroup.POST("/recharge-orders/:id/refunds", reconciliationHandler.Refund)

	// Refund request approval (user-initiated flow).
	adminGroup.GET("/refund-requests", refundRequestHandler.ListAdmin)
	adminGroup.POST("/refund-requests/:id/approve", refundRequestHandler.Approve)
	adminGroup.POST("/refund-requests/:id/reject", refundRequestHandler.Reject)

	// API-key protected relay routes (OpenAI-compatible endpoints).
	relayGroup := engine.Group("/v1", middleware.OpenAIAPIKeyAuth(apiKeyValidator))
	relayGroup.GET("/models", relayHandler.ListModels)
	relayGroup.POST("/chat/completions", relayHandler.ChatCompletions)
	relayGroup.POST("/messages", relayHandler.Messages)

	// API-key protected internal routes.
	apiGroup := engine.Group("/api/v1", middleware.APIKeyAuth(apiKeyValidator))
	apiGroup.GET("/me", apiKeyHandler.Me)
	apiGroup.POST("/quota-requests", quotaRequestHandler.Create)
	apiGroup.GET("/quota-requests", quotaRequestHandler.ListMy)
	apiGroup.POST("/recharge/orders", rechargeOrderHandler.Create)
	apiGroup.GET("/recharge/config", settingsHandler.GetPublicRechargeConfig)
	apiGroup.GET("/recharge/orders", rechargeOrderHandler.List)
	apiGroup.GET("/recharge/orders/:id", rechargeOrderHandler.Get)
	apiGroup.POST("/recharge/orders/:id/mock-callback", rechargeOrderHandler.MockCallback)
	apiGroup.GET("/balance-records", balanceRecordHandler.ListMy)
	apiGroup.GET("/quota-records", quotaRecordHandler.ListMy)

	// User-initiated refund requests.
	apiGroup.POST("/refund-requests", refundRequestHandler.Create)
	apiGroup.GET("/refund-requests", refundRequestHandler.ListMy)
	apiGroup.DELETE("/refund-requests/:id", refundRequestHandler.Cancel)

	// User-facing notifications and stats.
	apiGroup.GET("/notifications", notificationHandler.ListMy)
	apiGroup.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	apiGroup.POST("/notifications/read-all", notificationHandler.ReadAll)
	apiGroup.PATCH("/notifications/:id/read", notificationHandler.MarkRead)
	apiGroup.GET("/stats/usage", statsHandler.UserStats)
	apiGroup.GET("/stats/trend", statsHandler.UserTrend)
	apiGroup.GET("/stats/model-stats", statsHandler.UserModelStats)
	apiGroup.GET("/guide/sections", guideSectionHandler.ListUser)

	// User-facing balance export (XLSX download).
	apiGroup.GET("/export/balance-records", exportHandler.UserBalanceRecordsExport)

	// User portal account routes (user JWT only; admin and API-key tokens are
	// rejected by the audience check).
	userGroup := engine.Group("/api/v1/user", middleware.UserJWTAuth(tokenManager))
	userGroup.POST("/change-password", userAuthHandler.ChangePassword)
	userGroup.POST("/token/reveal", userAuthHandler.RevealToken)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := server.New(addr, engine, logger)

	// Start the periodic call-log sweep in the background. The goroutine
	// stops on its own when the server is shut down and its context cancelled.
	retentionCtx, retentionCancel := context.WithCancel(context.Background())
	defer retentionCancel()
	go logRetention.Start(retentionCtx, logger)

	if err := srv.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	slog.Info("server started", "addr", addr)

	if err := srv.WaitForShutdown(30 * time.Second); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
