package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weex-watchdog/internal/model"
	"weex-watchdog/internal/service"
	"weex-watchdog/pkg/logger"
)

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// TraderHandler 交易员处理器
type TraderHandler struct {
	traderService *service.TraderService
	logger        *logger.Logger
}

// NewTraderHandler 创建交易员处理器
func NewTraderHandler(traderService *service.TraderService, logger *logger.Logger) *TraderHandler {
	return &TraderHandler{
		traderService: traderService,
		logger:        logger,
	}
}

// CreateTraderRequest 创建交易员请求
type CreateTraderRequest struct {
	TraderUserID    string `json:"trader_user_id" binding:"required"`
	TraderName      string `json:"trader_name"`
	MonitorInterval int    `json:"monitor_interval"`
}

// CreateTrader 创建交易员监控
func (h *TraderHandler) CreateTrader(c *gin.Context) {
	var req CreateTraderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	trader := &model.TraderMonitor{
		TraderUserID:    req.TraderUserID,
		TraderName:      req.TraderName,
		MonitorInterval: req.MonitorInterval,
		IsActive:        true,
	}

	if trader.MonitorInterval <= 0 {
		trader.MonitorInterval = 30
	}

	if err := h.traderService.CreateTrader(trader); err != nil {
		h.logger.WithField("error", err).Error("Failed to create trader")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to create trader: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Trader created successfully",
		Data:    trader,
	})
}

// GetTraders 获取交易员列表
func (h *TraderHandler) GetTraders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}

	traders, total, err := h.traderService.GetTraders(page, size)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get traders")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to get traders: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginationResponse{
		Success: true,
		Message: "Traders retrieved successfully",
		Data:    traders,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// UpdateTrader 更新交易员
func (h *TraderHandler) UpdateTrader(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid trader ID",
		})
		return
	}

	trader, err := h.traderService.GetTraderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "Trader not found",
		})
		return
	}

	var req CreateTraderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	trader.TraderName = req.TraderName
	if req.MonitorInterval > 0 {
		trader.MonitorInterval = req.MonitorInterval
	}

	if err := h.traderService.UpdateTrader(trader); err != nil {
		h.logger.WithField("error", err).Error("Failed to update trader")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to update trader: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Trader updated successfully",
		Data:    trader,
	})
}

// DeleteTrader 删除交易员
func (h *TraderHandler) DeleteTrader(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid trader ID",
		})
		return
	}

	if err := h.traderService.DeleteTrader(uint(id)); err != nil {
		h.logger.WithField("error", err).Error("Failed to delete trader")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to delete trader: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Trader deleted successfully",
	})
}

// ToggleMonitor 切换监控状态
func (h *TraderHandler) ToggleMonitor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid trader ID",
		})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.traderService.ToggleTraderMonitor(uint(id), req.IsActive); err != nil {
		h.logger.WithField("error", err).Error("Failed to toggle trader monitor")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to toggle trader monitor: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Trader monitor status updated successfully",
	})
}

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService *service.OrderService
	logger       *logger.Logger
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(orderService *service.OrderService, logger *logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

// GetOrderHistory 获取订单历史
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	traderUserID := c.Query("trader_user_id")
	traderName := c.Query("trader_name")
	contractSymbol := c.Query("contract_symbol")
	status := c.Query("status")
	dateFilter := c.Query("date_filter") // today, 7days, 10days, 30days
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	// 构建筛选参数
	filters := map[string]interface{}{
		"trader_name":     traderName,
		"contract_symbol": contractSymbol,
		"status":          status,
		"date_filter":     dateFilter,
	}

	orders, total, err := h.orderService.GetOrderHistoryWithFilters(traderUserID, filters, page, size)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get order history")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to get order history: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginationResponse{
		Success: true,
		Message: "Order history retrieved successfully",
		Data:    orders,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// GetActiveOrders 获取指定交易员的活跃订单
func (h *OrderHandler) GetActiveOrders(c *gin.Context) {
	traderUserID := c.Query("trader_user_id")
	if traderUserID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "trader_user_id is required",
		})
		return
	}

	orders, err := h.orderService.GetActiveOrdersByTrader(traderUserID)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get active orders")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to get active orders: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Active orders retrieved successfully",
		Data:    orders,
	})
}

// GetStatistics 获取统计数据
func (h *OrderHandler) GetStatistics(c *gin.Context) {
	traderUserID := c.Query("trader_user_id")

	stats, err := h.orderService.GetStatistics(traderUserID)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get statistics")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to get statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Statistics retrieved successfully",
		Data:    stats,
	})
}

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService *service.NotificationService
	logger              *logger.Logger
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notificationService *service.NotificationService, logger *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// GetNotificationLogs 获取通知记录
func (h *NotificationHandler) GetNotificationLogs(c *gin.Context) {
	traderUserID := c.Query("trader_user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	logs, total, err := h.notificationService.GetNotificationLogs(traderUserID, page, size)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get notification logs")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to get notification logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginationResponse{
		Success: true,
		Message: "Notification logs retrieved successfully",
		Data:    logs,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// TestNotification 测试通知
func (h *NotificationHandler) TestNotification(c *gin.Context) {
	// 获取message
	message := c.GetString("message")
	if message == "" {
		message = "This is a test notification"
	}
	if err := h.notificationService.TestNotification(message); err != nil {
		h.logger.WithField("error", err).Error("Failed to send test notification")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to send test notification: " + err.Error(),
		})
		return
	}
	
	// 这里可以实现测试通知功能
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Test notification sent successfully",
	})
}
