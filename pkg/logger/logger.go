package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger 日志器
type Logger struct {
	*logrus.Logger
}

// Config 日志配置
type Config struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// NewLogger 创建新的日志器
func NewLogger(config *Config) *Logger {
	log := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	switch config.Output {
	case "file":
		file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	case "both":
		_, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(os.Stdout) // 同时输出到控制台和文件需要额外处理
		}
	default:
		log.SetOutput(os.Stdout)
	}

	return &Logger{log}
}

// WithFields 添加字段
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithField 添加单个字段
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// Info 记录信息级别日志
func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Error 记录错误级别日志
func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}
