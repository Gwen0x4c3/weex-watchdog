package repository

import (
	"time"
	"weex-watchdog/internal/model"

	"gorm.io/gorm"
)

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

func (r *orderRepository) GetOrderHistoryWithFilters(traderUserID string, filters map[string]interface{}, offset, limit int) ([]model.OrderHistory, int64, error) {
	var orders []model.OrderHistory
	var count int64

	query := r.db.Model(&model.OrderHistory{})
	
	// 基础筛选：交易员ID
	if traderUserID != "" {
		query = query.Where("trader_user_id = ?", traderUserID)
	}

	// 交易员名称模糊搜索
	if traderName, ok := filters["trader_name"].(string); ok && traderName != "" {
		query = query.Where("trader_name LIKE ?", "%"+traderName+"%")
	}

	// 币种模糊搜索
	if contractSymbol, ok := filters["contract_symbol"].(string); ok && contractSymbol != "" {
		query = query.Where("contract_symbol LIKE ?", "%"+contractSymbol+"%")
	}

	// 状态筛选
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 时间筛选
	if dateFilter, ok := filters["date_filter"].(string); ok && dateFilter != "" {
		now := time.Now()
		var startTime time.Time
		
		switch dateFilter {
		case "today":
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "7days":
			startTime = now.AddDate(0, 0, -7)
		case "10days":
			startTime = now.AddDate(0, 0, -10)
		case "30days":
			startTime = now.AddDate(0, 0, -30)
		}
		
		if !startTime.IsZero() {
			query = query.Where("first_seen_at >= ?", startTime)
		}
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

// DeleteByTraderID 删除指定交易员的所有订单
func (r *orderRepository) DeleteByTraderID(traderID uint) error {
	return r.db.Where("trader_user_id = ?", traderID).Delete(&model.OrderHistory{}).Error
}
