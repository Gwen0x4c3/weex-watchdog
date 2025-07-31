package repository

import (
	"weex-watchdog/internal/model"

	"gorm.io/gorm"
)

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
