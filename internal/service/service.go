package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"weex-watchdog/internal/model"
	"weex-watchdog/internal/repository"
	"weex-watchdog/pkg/constant"
	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/notification"
	"weex-watchdog/pkg/weex"
)

// TraderService 交易员服务
type TraderService struct {
	traderRepo     repository.TraderRepository
	logger         *logger.Logger
	monitorService *MonitorService // 添加对监控服务的引用
}

// NewTraderService 创建交易员服务
func NewTraderService(traderRepo repository.TraderRepository, logger *logger.Logger) *TraderService {
	return &TraderService{
		traderRepo: traderRepo,
		logger:     logger,
	}
}

// SetMonitorService 设置监控服务引用（避免循环依赖）
func (s *TraderService) SetMonitorService(monitorService *MonitorService) {
	s.monitorService = monitorService
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
	err := s.traderRepo.Update(trader)
	if err != nil {
		return err
	}

	// 清理监控缓存，以便新的监控间隔立即生效
	if s.monitorService != nil {
		s.monitorService.ClearTraderCache(trader.TraderUserID)
	}

	return nil
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
	notificationService notification.Client
	httpClient          *http.Client
	logger              *logger.Logger
	apiURL              string
	traderLastCheck     map[string]time.Time // 记录每个交易员最后检查时间
	mu                  sync.RWMutex         // 保护 traderLastCheck 的并发访问
}

// NewMonitorService 创建监控服务
func NewMonitorService(
	traderRepo repository.TraderRepository,
	orderRepo repository.OrderRepository,
	notificationRepo repository.NotificationRepository,
	notificationService notification.Client,
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
		logger:          logger,
		apiURL:          apiURL,
		traderLastCheck: make(map[string]time.Time),
	}
}

// StartMonitoring 启动监控
func (s *MonitorService) StartMonitoring() {
	s.logger.Logger.Info("Starting monitoring service")

	ticker := time.NewTicker(1 * time.Second) // 改为1秒间隔
	defer ticker.Stop()

	for range ticker.C {
		s.monitorAllTraders()
	}
}

// monitorAllTraders 监控所有交易员
func (s *MonitorService) monitorAllTraders() {
	traders, err := s.traderRepo.GetActiveTraders()
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to get active traders")
		return
	}

	now := time.Now()

	for _, trader := range traders {
		// 检查是否需要监控这个交易员
		if s.shouldMonitorTrader(trader, now) {
			go s.monitorSingleTrader(trader, now)
		}
	}
}

// shouldMonitorTrader 判断是否应该监控某个交易员
func (s *MonitorService) shouldMonitorTrader(trader model.TraderMonitor, now time.Time) bool {
	s.mu.RLock()
	lastCheck, exists := s.traderLastCheck[trader.TraderUserID]
	s.mu.RUnlock()

	// 如果从未检查过，或者距离上次检查已经超过了设定的间隔
	if !exists || now.Sub(lastCheck) >= time.Duration(trader.MonitorInterval)*time.Second {
		// 更新最后检查时间
		s.mu.Lock()
		s.traderLastCheck[trader.TraderUserID] = now
		s.mu.Unlock()
		return true
	}

	return false
}

// monitorSingleTrader 监控单个交易员
func (s *MonitorService) monitorSingleTrader(trader model.TraderMonitor, _ time.Time) {
	s.logger.WithFields(map[string]interface{}{
		"trader_id": trader.TraderUserID,
		"interval":  trader.MonitorInterval,
	}).Info("Monitoring trader")

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
func (s *MonitorService) fetchTraderOrders(traderUserID string) ([]weex.OpenOrder, error) {
	// 将 traderUserID 转换为 uint
	traderID, err := strconv.ParseUint(traderUserID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid trader user ID: %w", err)
	}

	// 使用新的 GetOpenOrderList 函数
	orders, err := weex.GetOpenOrderList(uint(traderID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	return orders, nil
}

// detectNewOrders 检测新订单
func (s *MonitorService) detectNewOrders(traderUserID string, currentOrders []weex.OpenOrder) {
	newOrders := make([]*model.OrderHistory, 0)
	for _, order := range currentOrders {
		// 检查订单是否已存在
		existing, err := s.orderRepo.GetByTraderAndOrderID(traderUserID, order.OpenOrderID)
		if err != nil || existing == nil {
			// 新订单
			// 解析创建时间字符串为时间戳
			createdTime, err := strconv.ParseInt(order.CreatedTime, 10, 64)
			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"trader_id":    traderUserID,
					"order_id":     order.OpenOrderID,
					"created_time": order.CreatedTime,
					"error":        err,
				}).Error("Failed to parse created time")
				createdTime = time.Now().UnixMilli() // 使用当前时间作为fallback
			}

			// 获取交易对名称
			contractMapper := weex.GetContractMapper()
			symbolName := contractMapper.GetSymbolName(order.ContractID)

			orderHistory := &model.OrderHistory{
				TraderUserID:   traderUserID,
				OrderID:        order.OpenOrderID,
				OrderData:      s.convertToJSON(order),
				ContractSymbol: symbolName,
				Status:         model.OrderStatusActive,
				PositionSide:   order.PositionSide,
				OpenSize:       order.OpenSize,
				OpenPrice:      order.AverageOpenPrice,
				OpenLeverage:   order.OpenLeverage + "x",
				FirstSeenAt:    time.UnixMilli(createdTime),
				LastSeenAt:     time.Now(),
			}

			if err := s.orderRepo.Create(orderHistory); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"trader_id": traderUserID,
					"order_id":  order.OpenOrderID,
					"error":     err,
				}).Error("Failed to save new order")
				continue
			}

			// 添加到新订单列表
			newOrders = append(newOrders, orderHistory)

			s.logger.WithFields(map[string]interface{}{
				"trader_id": traderUserID,
				"order_id":  order.OpenOrderID,
			}).Info("New order detected")
		}
	}
	// 统一发送开仓通知
	s.sendNewOrderNotification(newOrders)
}

// detectClosedOrders 检测平仓订单
func (s *MonitorService) detectClosedOrders(traderUserID string, currentOrders []weex.OpenOrder) {
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
		currentOrderIDs[order.OpenOrderID] = true
	}

	closedOrders := make([]*model.OrderHistory, 0)
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

			closedOrders = append(closedOrders, &activeOrder)

			s.logger.WithFields(map[string]interface{}{
				"trader_id": traderUserID,
				"order_id":  activeOrder.OrderID,
			}).Info("Order closed detected")
		}
	}

	// 发送平仓通知
	s.sendCloseOrderNotification(closedOrders)

}

// sendNewOrderNotification 发送新订单通知
func (s *MonitorService) sendNewOrderNotification(newOrders []*model.OrderHistory) {
	if len(newOrders) == 0 {
		return
	}
	message := ""
	notificationIds := make([]uint, 0)
	for _, order := range newOrders {
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

		notificationIds = append(notificationIds, notificationLog.ID)
		if message != "" {
			message += "\n\n"
		} else {
			message = `<div style="border: 2px solid #007bff; border-radius: 8px; padding: 8px; background-color: #f8f9fa; margin: 5px 0;">
<h3 style="color: #007bff; margin: 0 0 6px 0;padding:0">🆕 新开仓提醒</h3><br/>`
		}
		message += s.notificationService.BuildNotificationMessage(order, constant.PositionLong)
	}
	// 发送通知
	notificationMsg := &notification.NotificationMessage{
		Type: string(model.NotificationTypeNewOrder),
		Message: message,
	}
	
	if err := s.notificationService.SendMessage(*notificationMsg); err != nil {
		s.logger.WithField("error", err).Error("Failed to send new order notification")
		s.notificationRepo.UpdateStatusBatch(notificationIds, model.NotificationStatusFailed, err.Error())
	} else {
		s.notificationRepo.UpdateStatusBatch(notificationIds, model.NotificationStatusSuccess, "")
	}
}

// sendCloseOrderNotification 发送平仓通知
func (s *MonitorService) sendCloseOrderNotification(closedOrders []*model.OrderHistory) {
	if len(closedOrders) == 0 {
		return
	}
	message := ""
	notificationIds := make([]uint, 0)
	
	for _, order := range closedOrders {
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
			continue
		}

		notificationIds = append(notificationIds, notificationLog.ID)
		if message != "" {
			message += "\n\n"
		} else {
			message = `<div style="border: 2px solid #6c757d; border-radius: 8px; padding: 8px; background-color: #f8f9fa; margin: 5px 0;">
<h3 style="color: #6c757d; margin: 0 0 6px 0;">❌ 平仓提醒</h3><br/>`
		}
		message += s.notificationService.BuildNotificationMessage(order, constant.PositionLong)
	}
	
	// 发送通知
	notificationMsg := &notification.NotificationMessage{
		Type:    string(model.NotificationTypeOrderClosed),
		Message: message,
	}
	
	if err := s.notificationService.SendMessage(*notificationMsg); err != nil {
		s.notificationRepo.UpdateStatusBatch(notificationIds, model.NotificationStatusFailed, err.Error())
		s.logger.WithField("error", err).Error("Failed to send close order notification")
	} else {
		s.notificationRepo.UpdateStatusBatch(notificationIds, model.NotificationStatusSuccess, "")
	}
}

// convertToJSON 转换为JSON
func (s *MonitorService) convertToJSON(order weex.OpenOrder) model.JSON {
	data := make(map[string]interface{})
	orderBytes, _ := json.Marshal(order)
	json.Unmarshal(orderBytes, &data)
	return data
}

// ClearTraderCache 清理交易员的监控缓存（当交易员信息更新时调用）
func (s *MonitorService) ClearTraderCache(traderUserID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.traderLastCheck, traderUserID)
	s.logger.WithField("trader_id", traderUserID).Debug("Cleared trader monitoring cache")
}

// RefreshAllTraderCache 刷新所有交易员的监控缓存
func (s *MonitorService) RefreshAllTraderCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traderLastCheck = make(map[string]time.Time)
	s.logger.Info("Refreshed all trader monitoring cache")
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
	notificationClient notification.Client
	logger           *logger.Logger
}

// NewNotificationService 创建通知服务
func NewNotificationService(notificationRepo repository.NotificationRepository, notificationClient notification.Client, logger *logger.Logger) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		notificationClient: notificationClient,
		logger:           logger,
	}
}

// GetNotificationLogs 获取通知记录
func (s *NotificationService) GetNotificationLogs(traderUserID string, page, pageSize int) ([]model.NotificationLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.notificationRepo.GetLogs(traderUserID, offset, pageSize)
}

// TestNotification 发送测试消息
func (s *NotificationService) TestNotification(message string) error {
	notificationMsg := &notification.NotificationMessage{
		Type: string(model.NotificationTypeNewOrder),
		Message: message,
	}
	if err := s.notificationClient.SendMessage(*notificationMsg); err != nil {
		s.logger.WithField("error", err).Error("Failed to send test notification")
		return err
	}
	return nil
}