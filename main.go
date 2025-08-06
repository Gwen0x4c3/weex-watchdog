package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"weex-watchdog/internal/api"
	"weex-watchdog/internal/api/handler"
	"weex-watchdog/internal/config"
	"weex-watchdog/internal/repository"
	"weex-watchdog/internal/service"
	"weex-watchdog/pkg/cache"
	"weex-watchdog/pkg/database"
	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/notification"
	"weex-watchdog/pkg/weex"
)


func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	appLogger := logger.NewLogger(&config.Log)
	appLogger.Info("Starting Weex Monitor application")

	// 初始化数据库
	db, err := database.InitDB(&config.Database)
	if err != nil {
		appLogger.Error("Failed to initialize database:", err)
		os.Exit(1)
	}
	appLogger.Info("Database initialized successfully")

	// 加载合约映射
	appLogger.Info("Loading contract mappings...")
	contractMapper := weex.GetContractMapper()
	if err := contractMapper.LoadContractMapping(); err != nil {
		appLogger.Error("Failed to load contract mappings (will continue with empty mappings):", err)
	} else {
		mappingCount := contractMapper.GetMappingCount()
		appLogger.Info("Contract mappings loaded successfully, total contracts:", mappingCount)
	}

	// 初始化仓库
	traderRepo := repository.NewTraderRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// 初始化缓存
	memoryCache := cache.NewMemoryCache()

	// 初始化通知服务
	notificationClient, err := notification.CreateClient(&config.Notification)
	if err != nil {
		appLogger.Error("Failed to create notification client:", err)
		os.Exit(1)
	}

	// 初始化业务服务
	orderService := service.NewOrderService(orderRepo, appLogger)
	traderService := service.NewTraderService(traderRepo, orderService, appLogger)
	traderAnalysisService := service.NewTraderAnalysisService(memoryCache, appLogger)  // 添加交易员分析服务
	notificationService := service.NewNotificationService(notificationRepo, notificationClient, appLogger)
	monitorService := service.NewMonitorService(
		traderRepo,
		orderRepo,
		notificationRepo,
		notificationClient,
		appLogger,
		config.Weex.APIURL,
	)
	traderService.SetMonitorService(monitorService)

	// 初始化处理器
	traderHandler := handler.NewTraderHandler(traderService, appLogger)
	orderHandler := handler.NewOrderHandler(orderService, appLogger)
	notificationHandler := handler.NewNotificationHandler(notificationService, appLogger)
	analysisHandler := handler.NewTraderAnalysisHandler(traderAnalysisService, appLogger)  // 添加分析处理器
	authHandler := handler.NewAuthHandler(config.Auth.Username, config.Auth.Password, []byte(config.Auth.AESKey), appLogger)

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)

	// 创建Gin引擎
	engine := gin.New()

	// 设置路由
	router := api.NewRouter(traderHandler, orderHandler, notificationHandler, analysisHandler, authHandler, []byte(config.Auth.AESKey), config.Auth.Username, config.Auth.Password)
	router.SetupRoutes(engine)

	// 启动监控服务
	go monitorService.StartMonitoring()

	// 启动HTTP服务器
	port := config.Server.Port
	if port == "" {
		port = "8080"
	}

	appLogger.Info("Server starting on port:", port)
	if err := engine.Run(":" + port); err != nil {
		appLogger.Error("Failed to start server:", err)
		os.Exit(1)
	}
}

// loadConfig 加载配置文件
func loadConfig(configFile string) (*config.Config, error) {
	viper.SetConfigFile(configFile)

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parse_time", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "both")
	viper.SetDefault("monitor.default_interval", "30s")
	viper.SetDefault("monitor.max_goroutines", 100)
	viper.SetDefault("notification.timeout", "10s")

	// 环境变量映射
	viper.SetEnvPrefix("WEEX")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
