package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"weex-watchdog/internal/model"
	"weex-watchdog/internal/repository"
	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/notification"
)

// TraderService 交易员服务
type TraderService struct {
	traderRepo repository.TraderRepository
	logger     *logger.Logger
}

// NewTraderService 创建交易员服务
func NewTraderService(traderRepo repository.TraderRepository, logger *logger.Logger) *TraderService {
	return &TraderService{
		traderRepo: traderRepo,
		logger:     logger,
	}
}

// CreateTrader 创建交易员监控
func (s *TraderService) CreateTrader(trader *model.TraderMonitor) error {
	// 检查是否已存在
	existing, err := s.traderRepo.GetByTraderUserID(trader.TraderUserID)
	if err == nil && existing != nil {
		return fmt.Errorf("trader %s already exists", trader.TraderUserID)
	}

	return s.traderRepo.Create(trader)
}

// GetTraders 获取交易员列表
func (s *TraderService) GetTraders(page, pageSize int) ([]model.TraderMonitor, int64, error) {
	offset := (page - 1) * pageSize
	return s.traderRepo.GetAll(offset, pageSize)
}

// GetTraderByID 根据ID获取交易员
func (s *TraderService) GetTraderByID(id uint) (*model.TraderMonitor, error) {
	return s.traderRepo.GetByID(id)
}

// UpdateTrader 更新交易员信息
func (s *TraderService) UpdateTrader(trader *model.TraderMonitor) error {
	return s.traderRepo.Update(trader)
}

// DeleteTrader 删除交易员
func (s *TraderService) DeleteTrader(id uint) error {
	return s.traderRepo.Delete(id)
}

// ToggleTraderMonitor 启用/禁用交易员监控
func (s *TraderService) ToggleTraderMonitor(id uint, isActive bool) error {
	return s.traderRepo.ToggleActive(id, isActive)
}

// GetActiveTraders 获取活跃交易员
func (s *TraderService) GetActiveTraders() ([]model.TraderMonitor, error) {
	return s.traderRepo.GetActiveTraders()
}

// MonitorService 监控服务
type MonitorService struct {
	traderRepo          repository.TraderRepository
	orderRepo           repository.OrderRepository
	notificationRepo    repository.NotificationRepository
	notificationService notification.Service
	httpClient          *http.Client
	logger              *logger.Logger
	apiURL              string
}

// NewMonitorService 创建监控服务
func NewMonitorService(
	traderRepo repository.TraderRepository,
	orderRepo repository.OrderRepository,
	notificationRepo repository.NotificationRepository,
	notificationService notification.Service,
	logger *logger.Logger,
	apiURL string,
) *MonitorService {
	return &MonitorService{
		traderRepo:          traderRepo,
		orderRepo:           orderRepo,
		notificationRepo:    notificationRepo,
		notificationService: notificationService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
		apiURL: apiURL,
	}
}

// WeexOrder Weex API返回的订单结构
type WeexOrder struct {
	OrderID      string `json:"orderId"`
	PositionSide string `json:"positionSide"`
	Size         string `json:"size"`
	Price        string `json:"price"`
	Leverage     string `json:"leverage"`
	// 其他字段...
}

// WeexResponse Weex API响应结构
type WeexResponse struct {
	Success bool        `json:"success"`
	Data    []WeexOrder `json:"data"`
	Message string      `json:"message"`
}

// StartMonitoring 启动监控
func (s *MonitorService) StartMonitoring() {
	s.logger.Logger.Info("Starting monitoring service")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.monitorAllTraders()
		}
	}
}

// monitorAllTraders 监控所有交易员
func (s *MonitorService) monitorAllTraders() {
	traders, err := s.traderRepo.GetActiveTraders()
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to get active traders")
		return
	}

	s.logger.WithField("count", len(traders)).Info("Monitoring traders")

	for _, trader := range traders {
		go s.monitorSingleTrader(trader)
	}
}

// monitorSingleTrader 监控单个交易员
func (s *MonitorService) monitorSingleTrader(trader model.TraderMonitor) {
	s.logger.WithField("trader_id", trader.TraderUserID).Debug("Monitoring trader")

	// 获取当前订单
	orders, err := s.fetchTraderOrders(trader.TraderUserID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"trader_id": trader.TraderUserID,
			"error":     err,
		}).Error("Failed to fetch trader orders")
		return
	}

	// 检测新订单
	s.detectNewOrders(trader.TraderUserID, orders)

	// 检测平仓订单
	s.detectClosedOrders(trader.TraderUserID, orders)
}

// fetchTraderOrders 获取交易员订单
func (s *MonitorService) fetchTraderOrders(traderUserID string) ([]WeexOrder, error) {
	url := fmt.Sprintf("%s?userId=%s", s.apiURL, traderUserID)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var response WeexResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	return response.Data, nil
}

// detectNewOrders 检测新订单
func (s *MonitorService) detectNewOrders(traderUserID string, currentOrders []WeexOrder) {
	for _, order := range currentOrders {
		// 检查订单是否已存在
		existing, err := s.orderRepo.GetByTraderAndOrderID(traderUserID, order.OrderID)
		if err != nil || existing == nil {
			// 新订单
			orderHistory := &model.OrderHistory{
				TraderUserID: traderUserID,
				OrderID:      order.OrderID,
				OrderData:    s.convertToJSON(order),
				Status:       model.OrderStatusActive,
				PositionSide: order.PositionSide,
				OpenSize:     order.Size,
				OpenPrice:    order.Price,
				OpenLeverage: order.Leverage,
				FirstSeenAt:  time.Now(),
				LastSeenAt:   time.Now(),
			}

			if err := s.orderRepo.Create(orderHistory); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"trader_id": traderUserID,
					"order_id":  order.OrderID,
					"error":     err,
				}).Error("Failed to save new order")
				continue
			}

			// 发送通知
			s.sendNewOrderNotification(orderHistory)

			s.logger.WithFields(map[string]interface{}{
				"trader_id": traderUserID,
				"order_id":  order.OrderID,
			}).Info("New order detected")
		}
	}
}

// detectClosedOrders 检测平仓订单
func (s *MonitorService) detectClosedOrders(traderUserID string, currentOrders []WeexOrder) {
	// 获取数据库中的活跃订单
	activeOrders, err := s.orderRepo.GetActiveOrdersByTrader(traderUserID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"trader_id": traderUserID,
			"error":     err,
		}).Error("Failed to get active orders")
		return
	}

	// 创建当前订单ID映射
	currentOrderIDs := make(map[string]bool)
	for _, order := range currentOrders {
		currentOrderIDs[order.OrderID] = true
	}

	// 检查哪些活跃订单在当前订单中不存在（已平仓）
	for _, activeOrder := range activeOrders {
		if !currentOrderIDs[activeOrder.OrderID] {
			// 订单已平仓
			now := time.Now()
			if err := s.orderRepo.UpdateOrderStatus(activeOrder.ID, model.OrderStatusClosed, &now); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"trader_id": traderUserID,
					"order_id":  activeOrder.OrderID,
					"error":     err,
				}).Error("Failed to update order status")
				continue
			}

			// 更新本地对象用于发送通知
			activeOrder.Status = model.OrderStatusClosed
			activeOrder.ClosedAt = &now

			// 发送通知
			s.sendCloseOrderNotification(&activeOrder)

			s.logger.WithFields(map[string]interface{}{
				"trader_id": traderUserID,
				"order_id":  activeOrder.OrderID,
			}).Info("Order closed detected")
		}
	}
}

// sendNewOrderNotification 发送新订单通知
func (s *MonitorService) sendNewOrderNotification(order *model.OrderHistory) {
	// 记录通知日志
	notificationLog := &model.NotificationLog{
		TraderUserID:     order.TraderUserID,
		OrderID:          order.OrderID,
		NotificationType: model.NotificationTypeNewOrder,
		Status:           model.NotificationStatusPending,
		SentAt:           time.Now(),
	}

	if err := s.notificationRepo.Create(notificationLog); err != nil {
		s.logger.WithField("error", err).Error("Failed to create notification log")
		return
	}

	// 发送通知
	if err := s.notificationService.SendNewOrderNotification(order); err != nil {
		s.notificationRepo.UpdateStatus(notificationLog.ID, model.NotificationStatusFailed, err.Error())
		s.logger.WithField("error", err).Error("Failed to send new order notification")
	} else {
		s.notificationRepo.UpdateStatus(notificationLog.ID, model.NotificationStatusSuccess, "")
	}
}

// sendCloseOrderNotification 发送平仓通知
func (s *MonitorService) sendCloseOrderNotification(order *model.OrderHistory) {
	// 记录通知日志
	notificationLog := &model.NotificationLog{
		TraderUserID:     order.TraderUserID,
		OrderID:          order.OrderID,
		NotificationType: model.NotificationTypeOrderClosed,
		Status:           model.NotificationStatusPending,
		SentAt:           time.Now(),
	}

	if err := s.notificationRepo.Create(notificationLog); err != nil {
		s.logger.WithField("error", err).Error("Failed to create notification log")
		return
	}

	// 发送通知
	if err := s.notificationService.SendCloseOrderNotification(order); err != nil {
		s.notificationRepo.UpdateStatus(notificationLog.ID, model.NotificationStatusFailed, err.Error())
		s.logger.WithField("error", err).Error("Failed to send close order notification")
	} else {
		s.notificationRepo.UpdateStatus(notificationLog.ID, model.NotificationStatusSuccess, "")
	}
}

// convertToJSON 转换为JSON
func (s *MonitorService) convertToJSON(order WeexOrder) model.JSON {
	data := make(map[string]interface{})
	orderBytes, _ := json.Marshal(order)
	json.Unmarshal(orderBytes, &data)
	return data
}

// OrderService 订单服务
type OrderService struct {
	orderRepo repository.OrderRepository
	logger    *logger.Logger
}

// NewOrderService 创建订单服务
func NewOrderService(orderRepo repository.OrderRepository, logger *logger.Logger) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

// GetOrderHistory 获取订单历史
func (s *OrderService) GetOrderHistory(traderUserID string, page, pageSize int) ([]model.OrderHistory, int64, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.GetOrderHistory(traderUserID, offset, pageSize)
}

// GetStatistics 获取统计数据
func (s *OrderService) GetStatistics(traderUserID string) (map[string]interface{}, error) {
	return s.orderRepo.GetStatistics(traderUserID)
}

// NotificationService 通知服务
type NotificationService struct {
	notificationRepo repository.NotificationRepository
	logger           *logger.Logger
}

// NewNotificationService 创建通知服务
func NewNotificationService(notificationRepo repository.NotificationRepository, logger *logger.Logger) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

// GetNotificationLogs 获取通知记录
func (s *NotificationService) GetNotificationLogs(traderUserID string, page, pageSize int) ([]model.NotificationLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.notificationRepo.GetLogs(traderUserID, offset, pageSize)
}
