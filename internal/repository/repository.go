package repository

import (
	"time"

	"gorm.io/gorm"

	"weex-watchdog/internal/model"
)

// TraderRepository 交易员仓库接口
type TraderRepository interface {
	Create(trader *model.TraderMonitor) error
	GetByID(id uint) (*model.TraderMonitor, error)
	GetByTraderUserID(traderUserID string) (*model.TraderMonitor, error)
	GetActiveTraders() ([]model.TraderMonitor, error)
	GetAll(offset, limit int) ([]model.TraderMonitor, int64, error)
	Update(trader *model.TraderMonitor) error
	Delete(id uint) error
	ToggleActive(id uint, isActive bool) error
}

// OrderRepository 订单仓库接口
type OrderRepository interface {
	Create(order *model.OrderHistory) error
	GetByTraderAndOrderID(traderUserID, orderID string) (*model.OrderHistory, error)
	GetActiveOrdersByTrader(traderUserID string) ([]model.OrderHistory, error)
	UpdateOrderStatus(id uint, status model.OrderStatus, closedAt *time.Time) error
	GetOrderHistory(traderUserID string, offset, limit int) ([]model.OrderHistory, int64, error)
	GetStatistics(traderUserID string) (map[string]interface{}, error)
}

// NotificationRepository 通知仓库接口
type NotificationRepository interface {
	Create(log *model.NotificationLog) error
	UpdateStatus(id uint, status model.NotificationStatus, errorMsg string) error
	GetLogs(traderUserID string, offset, limit int) ([]model.NotificationLog, int64, error)
}

// traderRepository 交易员仓库实现
type traderRepository struct {
	db *gorm.DB
}

// NewTraderRepository 创建交易员仓库
func NewTraderRepository(db *gorm.DB) TraderRepository {
	return &traderRepository{db: db}
}

func (r *traderRepository) Create(trader *model.TraderMonitor) error {
	return r.db.Create(trader).Error
}

func (r *traderRepository) GetByID(id uint) (*model.TraderMonitor, error) {
	var trader model.TraderMonitor
	err := r.db.First(&trader, id).Error
	if err != nil {
		return nil, err
	}
	return &trader, nil
}

func (r *traderRepository) GetByTraderUserID(traderUserID string) (*model.TraderMonitor, error) {
	var trader model.TraderMonitor
	err := r.db.Where("trader_user_id = ?", traderUserID).First(&trader).Error
	if err != nil {
		return nil, err
	}
	return &trader, nil
}

func (r *traderRepository) GetActiveTraders() ([]model.TraderMonitor, error) {
	var traders []model.TraderMonitor
	err := r.db.Where("is_active = ?", true).Find(&traders).Error
	return traders, err
}

func (r *traderRepository) GetAll(offset, limit int) ([]model.TraderMonitor, int64, error) {
	var traders []model.TraderMonitor
	var count int64

	err := r.db.Model(&model.TraderMonitor{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&traders).Error
	return traders, count, err
}

func (r *traderRepository) Update(trader *model.TraderMonitor) error {
	return r.db.Save(trader).Error
}

func (r *traderRepository) Delete(id uint) error {
	return r.db.Delete(&model.TraderMonitor{}, id).Error
}

func (r *traderRepository) ToggleActive(id uint, isActive bool) error {
	return r.db.Model(&model.TraderMonitor{}).Where("id = ?", id).Update("is_active", isActive).Error
}

// orderRepository 订单仓库实现
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.OrderHistory) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByTraderAndOrderID(traderUserID, orderID string) (*model.OrderHistory, error) {
	var order model.OrderHistory
	err := r.db.Where("trader_user_id = ? AND order_id = ?", traderUserID, orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetActiveOrdersByTrader(traderUserID string) ([]model.OrderHistory, error) {
	var orders []model.OrderHistory
	err := r.db.Where("trader_user_id = ? AND status = ?", traderUserID, model.OrderStatusActive).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateOrderStatus(id uint, status model.OrderStatus, closedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if closedAt != nil {
		updates["closed_at"] = closedAt
	}
	return r.db.Model(&model.OrderHistory{}).Where("id = ?", id).Updates(updates).Error
}

func (r *orderRepository) GetOrderHistory(traderUserID string, offset, limit int) ([]model.OrderHistory, int64, error) {
	var orders []model.OrderHistory
	var count int64

	query := r.db.Model(&model.OrderHistory{})
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("first_seen_at DESC").Offset(offset).Limit(limit).Find(&orders).Error
	return orders, count, err
}

func (r *orderRepository) GetStatistics(traderUserID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总订单数
	var totalOrders int64
	query := r.db.Model(&model.OrderHistory{})
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}
	query.Count(&totalOrders)
	stats["total_orders"] = totalOrders

	// 活跃订单数
	var activeOrders int64
	query = r.db.Model(&model.OrderHistory{}).Where("status = ?", model.OrderStatusActive)
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}
	query.Count(&activeOrders)
	stats["active_orders"] = activeOrders

	// 已平仓订单数
	var closedOrders int64
	query = r.db.Model(&model.OrderHistory{}).Where("status = ?", model.OrderStatusClosed)
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}
	query.Count(&closedOrders)
	stats["closed_orders"] = closedOrders

	return stats, nil
}

// notificationRepository 通知仓库实现
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓库
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}

func (r *notificationRepository) UpdateStatus(id uint, status model.NotificationStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	return r.db.Model(&model.NotificationLog{}).Where("id = ?", id).Updates(updates).Error
}

func (r *notificationRepository) GetLogs(traderUserID string, offset, limit int) ([]model.NotificationLog, int64, error) {
	var logs []model.NotificationLog
	var count int64

	query := r.db.Model(&model.NotificationLog{})
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("sent_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, count, err
}
