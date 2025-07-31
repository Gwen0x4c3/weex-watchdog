package service

import (
	"fmt"
	"strconv"
	"weex-watchdog/internal/model"
	"weex-watchdog/internal/repository"
	"weex-watchdog/pkg/logger"
)

// TraderService 交易员服务
type TraderService struct {
	traderRepo     repository.TraderRepository
	orderService   *OrderService
	logger         *logger.Logger
	monitorService *MonitorService // 添加对监控服务的引用
}

// NewTraderService 创建交易员服务
func NewTraderService(traderRepo repository.TraderRepository, orderService *OrderService, logger *logger.Logger) *TraderService {
	return &TraderService{
		traderRepo:   traderRepo,
		orderService: orderService,
		logger:       logger,
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
	// 删除交易员
	if err := s.traderRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete trader: %w", err)
	}
	// 删除交易员的订单
	if err := s.orderService.DeleteByTraderId(id); err != nil {
		return fmt.Errorf("failed to delete trader's orders: %w", err)
	}
	// 清理监控缓存
	s.monitorService.ClearTraderCache(strconv.FormatUint(uint64(id), 10))
	return nil
}

// ToggleTraderMonitor 启用/禁用交易员监控
func (s *TraderService) ToggleTraderMonitor(id uint, isActive bool) error {
	return s.traderRepo.ToggleActive(id, isActive)
}

// GetActiveTraders 获取活跃交易员
func (s *TraderService) GetActiveTraders() ([]model.TraderMonitor, error) {
	return s.traderRepo.GetActiveTraders()
}
