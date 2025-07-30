package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"weex-watchdog/internal/model"
	"weex-watchdog/pkg/constant"
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
func (s *WecomNotificationClient) BuildNotificationMessage(order *model.OrderHistory, position string) string {
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
