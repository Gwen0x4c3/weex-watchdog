package notification

import (
	"errors"
	"fmt"
	"net/http"
	"weex-watchdog/internal/model"
	"weex-watchdog/pkg/constant"

	"github.com/wxpusher/wxpusher-sdk-go"
	wxpusher_model "github.com/wxpusher/wxpusher-sdk-go/model"
)

// WxPusherNotificationClient HTTP通知服务实现
type WxPusherNotificationClient struct {
	appToken string
	uid      string
	client   *http.Client
}

// NewWxPusherNotificationClient 创建HTTP通知服务
func NewWxPusherNotificationClient(config *Config) *WxPusherNotificationClient {
	return &WxPusherNotificationClient{
		appToken: config.Wxpusher.AppToken,
		uid:      config.Wxpusher.UID,
		client:   nil,
	}
}

// BuildNotificationMessage 构建通知消息
func (s *WxPusherNotificationClient) BuildNotificationMessage(order *model.OrderHistory, position string) string {
	// 确定方向和颜色
	var directionText, directionColor string
	if order.PositionSide == "LONG" {
		directionText = "做多"
		directionColor = "#28a745" // 绿色
	} else {
		directionText = "做空"
		directionColor = "#dc3545" // 红色
	}

	if position == constant.PositionLong {
		return fmt.Sprintf(`<div style="width: 100%%; border: 2px solid #e0e0e0; border-radius: 8px; padding: 12px; margin: 8px 0; background-color: #f9f9f9; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
<span style="color: #f39c12; font-weight: bold; font-size: 16px;">%s</span>
<span style="color: %s; font-weight: bold;">%s %sx</span>
</div>
<div style="color: #333; font-weight: bold; margin-bottom: 4px;">%s</div>
<div style="color: #333; margin-bottom: 4px;">开仓: %s</div>
<div style="color: #666; font-size: 12px;">%s</div>
</div>`,
			order.ContractSymbol,
			directionColor,
			directionText,
			order.OpenLeverage,
			order.TraderUserID,
			order.OpenPrice,
			order.FirstSeenAt.Format("2006-01-02 15:04:05"))
	} else {
		return fmt.Sprintf(`<div style="width: 100%%; border: 2px solid #e0e0e0; border-radius: 8px; padding: 12px; margin: 8px 0; background-color: #f9f9f9; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
<span style="color: #f39c12; font-weight: bold; font-size: 16px;">%s</span>
<span style="color: %s; font-weight: bold;">%s %sx</span>
</div>
<div style="color: #333; font-weight: bold; margin-bottom: 4px;">%s</div>
<div style="color: #333; margin-bottom: 4px;">开仓: %s</div>
<div style="color: #666; font-size: 12px;">平仓: %s</div>
</div>`,
			order.ContractSymbol,
			directionColor,
			directionText,
			order.OpenLeverage,
			order.TraderUserID,
			order.OpenPrice,
			order.ClosedAt.Format("2006-01-02 15:04:05"))
	}
}

// sendMessage 发送消息
func (s *WxPusherNotificationClient) SendMessage(notification NotificationMessage) error {
	if s.appToken == "" || s.uid == "" {
		return errors.New("wxpusher configuration is incomplete")
	}
	if notification.Message == "" {
		return errors.New("notification message cannot be empty")
	}
	msg := wxpusher_model.NewMessage(s.appToken).SetContent(notification.Message).AddUId(s.uid)
	msgArr, err := wxpusher.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	if len(msgArr) == 0 {
		return errors.New("no message sent, check your configuration")
	}
	return nil
}
