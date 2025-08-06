package config

import (
	"weex-watchdog/pkg/database"
	"weex-watchdog/pkg/logger"
	"weex-watchdog/pkg/notification"
)

// Config 应用配置
type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	Database     database.Config     `mapstructure:"database"`
	Log          logger.Config       `mapstructure:"log"`
	Weex         WeexConfig          `mapstructure:"weex"`
	Monitor      MonitorConfig       `mapstructure:"monitor"`
	Notification notification.Config `mapstructure:"notification"`
	Auth         AuthConfig          `mapstructure:"auth"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	AESKey   string `mapstructure:"aes_key"`
}

// WeexConfig Weex API配置
type WeexConfig struct {
	APIURL     string `mapstructure:"api_url"`
	Timeout    string `mapstructure:"timeout"`
	RetryTimes int    `mapstructure:"retry_times"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	DefaultInterval string `mapstructure:"default_interval"`
	MaxGoroutines   int    `mapstructure:"max_goroutines"`
}
