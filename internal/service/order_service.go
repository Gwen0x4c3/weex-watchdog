package service

import (
	"weex-watchdog/internal/model"
	"weex-watchdog/internal/repository"
	"weex-watchdog/pkg/logger"
)

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

// GetActiveOrdersByTrader 获取指定交易员的当前活跃订单
func (s *OrderService) GetActiveOrdersByTrader(traderUserID string) ([]model.OrderHistory, error) {
	return s.orderRepo.GetActiveOrdersByTrader(traderUserID)
}

// DeleteByTraderId 删除指定交易员的所有订单
func (s *OrderService) DeleteByTraderId(traderID uint) error {
	// 删除指定交易员的所有订单
	return s.orderRepo.DeleteByTraderID(traderID)
}
