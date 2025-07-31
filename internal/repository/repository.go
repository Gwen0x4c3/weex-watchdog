package repository

import (
	"time"

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
	GetOrderHistoryWithFilters(traderUserID string, filters map[string]interface{}, offset, limit int) ([]model.OrderHistory, int64, error)
	GetStatistics(traderUserID string) (map[string]interface{}, error)
	DeleteByTraderID(traderID uint) error
}

// NotificationRepository 通知仓库接口
type NotificationRepository interface {
	Create(log *model.NotificationLog) error
	UpdateStatus(id uint, status model.NotificationStatus, errorMsg string) error
	UpdateStatusBatch(ids []uint, status model.NotificationStatus, errorMsg string) error
	GetLogs(traderUserID string, offset, limit int) ([]model.NotificationLog, int64, error)
}