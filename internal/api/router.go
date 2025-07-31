package api

import (
	"github.com/gin-gonic/gin"

	"weex-watchdog/internal/api/handler"
	"weex-watchdog/internal/api/middleware"
)

// Router API路由器
type Router struct {
	traderHandler       *handler.TraderHandler
	orderHandler        *handler.OrderHandler
	notificationHandler *handler.NotificationHandler
}

// NewRouter 创建新的路由器
func NewRouter(
	traderHandler *handler.TraderHandler,
	orderHandler *handler.OrderHandler,
	notificationHandler *handler.NotificationHandler,
) *Router {
	return &Router{
		traderHandler:       traderHandler,
		orderHandler:        orderHandler,
		notificationHandler: notificationHandler,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// 中间件
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 交易员管理
		traders := v1.Group("/traders")
		{
			traders.GET("", r.traderHandler.GetTraders)
			traders.POST("", r.traderHandler.CreateTrader)
			traders.PUT("/:id", r.traderHandler.UpdateTrader)
			traders.DELETE("/:id", r.traderHandler.DeleteTrader)
			traders.POST("/:id/toggle", r.traderHandler.ToggleMonitor)
		}

		// 订单管理
		orders := v1.Group("/orders")
		{
			orders.GET("", r.orderHandler.GetOrderHistory)
			orders.GET("/active", r.orderHandler.GetActiveOrders)
			orders.GET("/statistics", r.orderHandler.GetStatistics)
		}

		// 通知管理
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", r.notificationHandler.GetNotificationLogs)
			notifications.POST("/test", r.notificationHandler.TestNotification)
		}
	}

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Weex Monitor API is running",
		})
	})

	// 静态文件服务
	engine.Static("/static", "./web/static")

	// 默认页面 - 重定向到静态文件
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/static/index.html")
	})
}
