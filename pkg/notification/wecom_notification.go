package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"weex-watchdog/internal/model"
)

// WecomNotificationClient HTTP通知服务实现
type WecomNotificationClient struct {
	wecomCID     string
	wecomAgentID string
	wecomSecret  string
	client       *http.Client
}

// NewWecomNotificationClient 创建HTTP通知服务
func NewWecomNotificationClient(config *Config) *WecomNotificationClient {
	return &WecomNotificationClient{
		wecomCID:     config.Wecom.CID,
		wecomAgentID: config.Wecom.AgentID,
		wecomSecret:  config.Wecom.Secret,
		client: &http.Client{
			Timeout: config.Wecom.Timeout,
		},
	}
}

// BuildNotificationMessage 构建通知消息
func (s *WecomNotificationClient) BuildNotificationMessage(orders []*model.OrderHistory, isOpen bool) string {
	if len(orders) == 0 {
		return ""
	}

	const maxLength = 40000
	var result string

	// 获取交易员ID（所有订单的交易员都相同）
	traderUserID := orders[0].TraderUserID

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
<div style="color: #333; font-weight: bold; margin-bottom: 8px;">交易员: %s</div>
`, titleColor, titleColor, getWecomActionIcon(isOpen), actionText, traderUserID)

	result = headerHtml
	if len(result) >= maxLength {
		return result[:maxLength]
	}

	// 按币种和方向分组
	groupedOrders := make(map[string][]*model.OrderHistory)
	for _, order := range orders {
		key := order.ContractSymbol + "_" + order.PositionSide
		groupedOrders[key] = append(groupedOrders[key], order)
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
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; font-size: 11px;">杠杆: %s</span>
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; font-size: 11px;">价格: %s</span>
<span style="background-color: #e9ecef; padding: 2px 6px; border-radius: 8px; font-size: 11px;">时间: %s</span>
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

// getWecomActionIcon 获取操作图标
func getWecomActionIcon(isOpen bool) string {
	if isOpen {
		return "🆕"
	}
	return "❌"
}

// sendMessage 发送消息
func (s *WecomNotificationClient) SendMessage(notification NotificationMessage) error {
	if s.wecomCID == "" || s.wecomAgentID == "" || s.wecomSecret == "" {
		return errors.New("wecom configuration is incomplete")
	}

	token, err := s.getToken()
	if err != nil || token == "" {
		return fmt.Errorf("failed to get WeCom token: %w", err)
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	data := map[string]interface{}{
		"touser":  "@all",
		"agentid": s.wecomAgentID,
		"msgtype": "text",
		"text": map[string]string{
			"content": notification.Message,
		},
		"duplicate_check_interval": 600,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal notification message: %w", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	fmt.Println("Sending notification to WeCom:", string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// 读取并打印响应内容
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
	} else {
		fmt.Println("WeCom response:", respBody.String())
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *WecomNotificationClient) getToken() (string, error) {
	if s.wecomCID == "" || s.wecomSecret == "" {
		return "", errors.New("wecom configuration is incomplete")
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", s.wecomCID, s.wecomSecret)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get WeCom token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get WeCom token, status code: %d", resp.StatusCode)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Token   string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode WeCom token response: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("WeCom API error: %s", result.ErrMsg)
	}

	return result.Token, nil
}
