package service

import (
	"weex-watchdog/internal/model"
	"weex-watchdog/internal/repository"
	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/notification"
)

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