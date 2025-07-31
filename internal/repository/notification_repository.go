package repository

import (
	"weex-watchdog/internal/model"

	"gorm.io/gorm"
)

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

func (r *notificationRepository) UpdateStatusBatch(ids []uint, status model.NotificationStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	return r.db.Model(&model.NotificationLog{}).Where("id IN ?", ids).Updates(updates).Error
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
