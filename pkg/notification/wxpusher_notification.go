package notification

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"weex-watchdog/internal/model"

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
func (s *WxPusherNotificationClient) BuildNotificationMessage(orders []*model.OrderHistory, isOpen bool) string {
	if len(orders) == 0 {
		return ""
	}

	const maxLength = 40000
	var result string

	// 获取交易员
	traderName := orders[0].TraderName

	// 按币种和方向分组
	groupedOrders := make(map[string][]*model.OrderHistory)
	for _, order := range orders {
		key := order.ContractSymbol + "_" + order.PositionSide
		groupedOrders[key] = append(groupedOrders[key], order)
	}

	// 生成开单预览
	previewParts := make([]string, 0, len(groupedOrders))
	for _, groupOrders := range groupedOrders {
		if len(groupOrders) == 0 {
			continue
		}
		
		symbol := strings.Replace(groupOrders[0].ContractSymbol, "/USDT", "", -1)
		positionSide := groupOrders[0].PositionSide
		
		var directionText, directionColor string
		if positionSide == "LONG" {
			directionText = "多"
			directionColor = "#28a745" // 绿色
		} else {
			directionText = "空"
			directionColor = "#dc3545" // 红色
		}
		
		previewPart := fmt.Sprintf(`<span style="display: inline-block; margin: 2px 4px; padding: 2px 6px; background-color: %s; color: white; border-radius: 10px; font-size: 11px; font-weight: bold;">%s(%s)</span>`, 
			directionColor, symbol, directionText)
		previewParts = append(previewParts, previewPart)
	}
	
	previewHtml := ""
	if len(previewParts) > 0 {
		previewHtml = fmt.Sprintf(`<div style="margin-top: 6px; margin-bottom: 8px;">%s</div>`, 
			strings.Join(previewParts, ""))
	}

	// 设置统一的大标题
	var titleColor, actionText string
	if isOpen {
		titleColor = "#007bff"
		actionText = "新开仓"
	} else {
		titleColor = "#6c757d"
		actionText = "平仓"
	}

	headerHtml := fmt.Sprintf(`<div style="border: 2px solid %s; border-radius: 8px; padding: 8px; background-color: #f8f9fa; margin: 5px 0;">
<h3 style="color: %s; margin: 0 0 6px 0; padding: 0;">%s %s提醒</h3>
<div style="color: #333; font-weight: bold; margin-bottom: 4px;">交易员: %s</div>
%s`, titleColor, titleColor, getActionIcon(isOpen), actionText, traderName, previewHtml)

	result = headerHtml
	if len(result) >= maxLength {
		return result[:maxLength]
	}



	// 为每个分组生成消息体
	for _, groupOrders := range groupedOrders {
		if len(groupOrders) == 0 {
			continue
		}

		symbol := groupOrders[0].ContractSymbol
		positionSide := groupOrders[0].PositionSide

		// 确定方向和颜色
		var directionText, directionColor string
		if positionSide == "LONG" {
			directionText = "做多"
			directionColor = "#28a745" // 绿色
		} else {
			directionText = "做空"
			directionColor = "#dc3545" // 红色
		}

		// 构建这个组的消息体
		groupHtml := fmt.Sprintf(`<div style="border: 1px solid #dee2e6; border-radius: 6px; padding: 10px; margin: 6px 0; background-color: #ffffff;">
<div style="display: flex; align-items: center; margin-bottom: 8px;">
<span style="color: #f39c12; font-weight: bold; font-size: 14px; margin-right: 8px;">%s</span>
<span style="background-color: %s; color: white; padding: 2px 8px; border-radius: 12px; font-size: 12px; font-weight: bold;">%s</span>
<span style="margin-left: 8px; color: #666; font-size: 12px;">(%d单)</span>
</div>
`, symbol, directionColor, directionText, len(groupOrders))

		// 检查长度
		if len(result)+len(groupHtml) >= maxLength {
			break
		}
		result += groupHtml

		// 添加订单详情
		for _, order := range groupOrders {
			var timeStr string
			if isOpen {
				timeStr = order.FirstSeenAt.Format("15:04:05")
			} else if order.ClosedAt != nil {
				timeStr = order.ClosedAt.Format("15:04:05")
			} else {
				timeStr = "未知"
			}

			orderDetailHtml := fmt.Sprintf(`<div style="display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 4px;">
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; color: rgb(20,20,20); font-size: 11px;">杠杆: %s</span>
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; color: rgb(20,20,20); font-size: 11px;">价格: %s</span>
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; color: rgb(20,20,20); font-size: 11px;">开仓: %s</span>
</div>
`, order.OpenLeverage, order.OpenPrice, timeStr)

			// 检查长度
			if len(result)+len(orderDetailHtml) >= maxLength {
				break
			}
			result += orderDetailHtml
		}

		// 关闭这个组的div
		closingDiv := "</div>"
		if len(result)+len(closingDiv) >= maxLength {
			break
		}
		result += closingDiv
	}

	// 关闭整个消息的div
	closingDiv := "</div>"
	if len(result)+len(closingDiv) < maxLength {
		result += closingDiv
	}

	return result
}

// getActionIcon 获取操作图标
func getActionIcon(isOpen bool) string {
	if isOpen {
		return "🆕"
	}
	return "❌"
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
