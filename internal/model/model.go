package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// TraderMonitor 监控交易员配置
type TraderMonitor struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	TraderUserID    string    `json:"trader_user_id" gorm:"type:varchar(50);not null;uniqueIndex"`
	TraderName      string    `json:"trader_name" gorm:"type:varchar(100)"`
	IsActive        bool      `json:"is_active" gorm:"default:true;index"`
	MonitorInterval int       `json:"monitor_interval" gorm:"default:30"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (TraderMonitor) TableName() string {
	return "trader_monitors"
}

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusActive OrderStatus = "ACTIVE"
	OrderStatusClosed OrderStatus = "CLOSED"
)

// JSON 自定义JSON类型
type JSON map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// OrderHistory 订单历史记录
type OrderHistory struct {
	ID             uint        `json:"id" gorm:"primaryKey"`
	TraderUserID   string      `json:"trader_user_id" gorm:"type:varchar(50);not null;index"`
	OrderID        string      `json:"order_id" gorm:"type:varchar(50);not null"`
	OrderData      JSON        `json:"order_data" gorm:"type:json"`
	ContractSymbol string      `json:"contract_symbol" gorm:"type:varchar(50);not null"`
	Status         OrderStatus `json:"status" gorm:"type:enum('ACTIVE','CLOSED');default:'ACTIVE';index"`
	PositionSide   string      `json:"position_side" gorm:"type:varchar(10)"`
	OpenSize       string      `json:"open_size" gorm:"type:decimal(20,8)"`
	OpenPrice      string      `json:"open_price" gorm:"type:decimal(20,8)"`
	OpenLeverage   string      `json:"open_leverage" gorm:"type:varchar(10)"`
	FirstSeenAt    time.Time   `json:"first_seen_at" gorm:"index"`
	LastSeenAt     time.Time   `json:"last_seen_at"`
	ClosedAt       *time.Time  `json:"closed_at"`
}

// TableName 指定表名
func (OrderHistory) TableName() string {
	return "order_history"
}

// NotificationType 通知类型枚举
type NotificationType string

const (
	NotificationTypeNewOrder    NotificationType = "NEW_ORDER"
	NotificationTypeOrderClosed NotificationType = "ORDER_CLOSED"
)

// NotificationStatus 通知状态枚举
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSuccess NotificationStatus = "SUCCESS"
	NotificationStatusFailed  NotificationStatus = "FAILED"
)

// NotificationLog 通知记录
type NotificationLog struct {
	ID               uint               `json:"id" gorm:"primaryKey"`
	TraderUserID     string             `json:"trader_user_id" gorm:"type:varchar(50);not null;index"`
	OrderID          string             `json:"order_id" gorm:"type:varchar(50)"`
	NotificationType NotificationType   `json:"notification_type" gorm:"type:enum('NEW_ORDER','ORDER_CLOSED');not null;index"`
	Message          string             `json:"message" gorm:"type:text"`
	Status           NotificationStatus `json:"status" gorm:"type:enum('PENDING','SUCCESS','FAILED');default:'PENDING';index"`
	SentAt           time.Time          `json:"sent_at"`
	ErrorMsg         string             `json:"error_msg" gorm:"type:text"`
}

// TableName 指定表名
func (NotificationLog) TableName() string {
	return "notification_logs"
}
