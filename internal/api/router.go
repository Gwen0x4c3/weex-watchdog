package api

import (
	"weex-watchdog/internal/api/handler"
	"weex-watchdog/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// Router API路由器
type Router struct {
	traderHandler       *handler.TraderHandler
	orderHandler        *handler.OrderHandler
	notificationHandler *handler.NotificationHandler
	analysisHandler     *handler.TraderAnalysisHandler
	authHandler         *handler.AuthHandler
	aesKey              []byte
	username 			string
	password 			string
}

// NewRouter 创建新的路由器
func NewRouter(
	traderHandler *handler.TraderHandler,
	orderHandler *handler.OrderHandler,
	notificationHandler *handler.NotificationHandler,
	analysisHandler *handler.TraderAnalysisHandler,
	authHandler *handler.AuthHandler,
	aesKey []byte,
	username string,
	password string,
) *Router {
	return &Router{
		traderHandler:       traderHandler,
		orderHandler:        orderHandler,
		notificationHandler: notificationHandler,
		analysisHandler:     analysisHandler,
		authHandler:         authHandler,
		aesKey:              aesKey,
		username:          	 username,
		password: 			 password,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// 全局中间件
	engine.Use(middleware.CORSMiddleware())
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// API v1
	apiV1 := engine.Group("/api/v1")

	// 登录接口，不受保护
	apiV1.POST("/login", r.authHandler.Login)

	// 受保护的API组
	protected := apiV1.Group("/")
	protected.Use(middleware.AuthMiddleware(r.aesKey, r.username, r.password))
	{
		// 交易员管理
		traders := protected.Group("/traders")
		{
			traders.GET("", r.traderHandler.GetTraders)
			traders.POST("", r.traderHandler.CreateTrader)
			traders.PUT("/:id", r.traderHandler.UpdateTrader)
			traders.DELETE("/:id", r.traderHandler.DeleteTrader)
			traders.POST("/:id/toggle", r.traderHandler.ToggleMonitor)
			traders.GET("/:id/analysis", r.analysisHandler.AnalyzeTrader)
		}

		// 订单管理
		orders := protected.Group("/orders")
		{
			orders.GET("", r.orderHandler.GetOrderHistory)
			orders.GET("/active", r.orderHandler.GetActiveOrders)
			orders.GET("/statistics", r.orderHandler.GetStatistics)
		}

		// 通知管理
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", r.notificationHandler.GetNotificationLogs)
			notifications.POST("/test", r.notificationHandler.TestNotification)
		}
	}

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 静态文件服务
	engine.Static("/static", "./web/static")

	// 根路径重定向到登录页
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/static/login.html")
	})
}
