package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"weex-watchdog/internal/model"
)

// Service 通知服务接口
type Service interface {
	SendNewOrderNotification(order *model.OrderHistory) error
	SendCloseOrderNotification(order *model.OrderHistory) error
}

// Config 通知配置
type Config struct {
	WebhookURL string        `mapstructure:"webhook_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

// HTTPNotificationService HTTP通知服务实现
type HTTPNotificationService struct {
	webhookURL string
	client     *http.Client
}

// NewHTTPNotificationService 创建HTTP通知服务
func NewHTTPNotificationService(config *Config) *HTTPNotificationService {
	return &HTTPNotificationService{
		webhookURL: config.WebhookURL,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SendNewOrderNotification 发送新订单通知
func (s *HTTPNotificationService) SendNewOrderNotification(order *model.OrderHistory) error {
	message := fmt.Sprintf(`🆕 新开仓提醒
交易员：%s
订单ID：%s
方向：%s
杠杆：%sx
数量：%s
开仓价：%s
时间：%s`,
		order.TraderUserID,
		order.OrderID,
		order.PositionSide,
		order.OpenLeverage,
		order.OpenSize,
		order.OpenPrice,
		order.FirstSeenAt.Format("2006-01-02 15:04:05"))

	notification := NotificationMessage{
		Type:    "NEW_ORDER",
		Message: message,
		Data:    order,
	}

	return s.sendMessage(notification)
}

// SendCloseOrderNotification 发送平仓通知
func (s *HTTPNotificationService) SendCloseOrderNotification(order *model.OrderHistory) error {
	message := fmt.Sprintf(`❌ 平仓提醒
交易员：%s
订单ID：%s
方向：%s
杠杆：%sx
数量：%s
开仓价：%s
平仓时间：%s`,
		order.TraderUserID,
		order.OrderID,
		order.PositionSide,
		order.OpenLeverage,
		order.OpenSize,
		order.OpenPrice,
		order.ClosedAt.Format("2006-01-02 15:04:05"))

	notification := NotificationMessage{
		Type:    "ORDER_CLOSED",
		Message: message,
		Data:    order,
	}

	return s.sendMessage(notification)
}

// sendMessage 发送消息
func (s *HTTPNotificationService) sendMessage(notification NotificationMessage) error {
	if s.webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification failed with status: %d", resp.StatusCode)
	}

	return nil
}
