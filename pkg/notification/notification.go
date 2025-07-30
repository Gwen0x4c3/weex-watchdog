package notification

import (
	"errors"
	"time"

	"weex-watchdog/internal/model"
)

// Client 通知服务接口
type Client interface {
	BuildNotificationMessage(orders []*model.OrderHistory, isOpen bool) string
	SendMessage(notification NotificationMessage) error
}

// Config 通知配置
type Config struct {
	Supplier string `mapstructure:"supplier"`
	// 企业微信配置
	Wecom struct {
		CID     string        `mapstructure:"cid"`
		AgentID string        `mapstructure:"agent_id"`
		Secret  string        `mapstructure:"secret"`
		Timeout time.Duration `mapstructure:"timeout"`
	} `mapstructure:"wecom"`
	// WxPusher 配置
	Wxpusher struct {
		AppToken string `mapstructure:"app_token"`
		UID      string `mapstructure:"uid"`
	} `mapstructure:"wxpusher"`
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func CreateClient(config *Config) (Client, error) {
	switch config.Supplier {
	case "wecom":
		return NewWecomNotificationClient(config), nil
	case "wxpusher":
		return NewWxPusherNotificationClient(config), nil
	default:
		return nil, errors.New("unsupported notification supplier: " + config.Supplier)
	}
}