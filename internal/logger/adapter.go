package logger

import (
	"xiaozhi-server-go/internal/platform/logging"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

// HCLogAdapter 是将 logging.Logger 适配到 hclog.Logger 接口的适配器
type HCLogAdapter struct {
	logger *logging.Logger
	name   string
}

// NewHCLogAdapter 创建一个新的 HCLog 适配器
func NewHCLogAdapter(logger *logging.Logger) *HCLogAdapter {
	return &HCLogAdapter{
		logger: logger,
	}
}

// Log 实现 hclog.Logger 接口
func (h *HCLogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		h.logger.DebugTag(h.name, msg, args...)
	case hclog.Debug:
		h.logger.DebugTag(h.name, msg, args...)
	case hclog.Info:
		h.logger.InfoTag(h.name, msg, args...)
	case hclog.Warn:
		h.logger.WarnTag(h.name, msg, args...)
	case hclog.Error:
		h.logger.ErrorTag(h.name, msg, args...)
	default:
		h.logger.InfoTag(h.name, msg, args...)
	}
}

// Trace 实现 hclog.Logger 接口
func (h *HCLogAdapter) Trace(msg string, args ...interface{}) {
	h.logger.DebugTag(h.name, msg, args...)
}

// Debug 实现 hclog.Logger 接口
func (h *HCLogAdapter) Debug(msg string, args ...interface{}) {
	h.logger.DebugTag(h.name, msg, args...)
}

// Info 实现 hclog.Logger 接口
func (h *HCLogAdapter) Info(msg string, args ...interface{}) {
	h.logger.InfoTag(h.name, msg, args...)
}

// Warn 实现 hclog.Logger 接口
func (h *HCLogAdapter) Warn(msg string, args ...interface{}) {
	h.logger.WarnTag(h.name, msg, args...)
}

// Error 实现 hclog.Logger 接口
func (h *HCLogAdapter) Error(msg string, args ...interface{}) {
	h.logger.ErrorTag(h.name, msg, args...)
}

// IsTrace 实现 hclog.Logger 接口
func (h *HCLogAdapter) IsTrace() bool {
	return false // 暂不支持Trace级别
}

// IsDebug 实现 hclog.Logger 接口
func (h *HCLogAdapter) IsDebug() bool {
	return true // 默认支持Debug级别
}

// IsInfo 实现 hclog.Logger 接口
func (h *HCLogAdapter) IsInfo() bool {
	return true // 默认支持Info级别
}

// IsWarn 实现 hclog.Logger 接口
func (h *HCLogAdapter) IsWarn() bool {
	return true // 默认支持Warn级别
}

// IsError 实现 hclog.Logger 接口
func (h *HCLogAdapter) IsError() bool {
	return true // 默认支持Error级别
}

// ImpliedArgs 实现 hclog.Logger 接口
func (h *HCLogAdapter) ImpliedArgs() []interface{} {
	return nil
}

// With 实现 hclog.Logger 接口
func (h *HCLogAdapter) With(args ...interface{}) hclog.Logger {
	// 创建新的适配器实例，复制现有配置
	newAdapter := &HCLogAdapter{
		logger: h.logger,
		name:   h.name,
	}
	// 注意：args 参数在当前的实现中被忽略，
	// 因为 logging.Logger 不支持动态字段
	return newAdapter
}

// Name 实现 hclog.Logger 接口
func (h *HCLogAdapter) Name() string {
	return h.name
}

// Named 实现 hclog.Logger 接口
func (h *HCLogAdapter) Named(name string) hclog.Logger {
	// 创建带有新名称的适配器
	newName := name
	if h.name != "" {
		newName = h.name + "." + name
	}

	return &HCLogAdapter{
		logger: h.logger,
		name:   newName,
	}
}

// ResetNamed 实现 hclog.Logger 接口
func (h *HCLogAdapter) ResetNamed(name string) hclog.Logger {
	return &HCLogAdapter{
		logger: h.logger,
		name:   name,
	}
}

// SetLevel 实现 hclog.Logger 接口
func (h *HCLogAdapter) SetLevel(level hclog.Level) {
	// logging.Logger 的级别通过配置管理，这里不做处理
}

// GetLevel 实现 hclog.Logger 接口
func (h *HCLogAdapter) GetLevel() hclog.Level {
	// 根据当前配置的日志级别返回对应的 hclog.Level
	// 这里返回一个默认值
	return hclog.Info
}

// StandardLogger 实现 hclog.Logger 接口
func (h *HCLogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	// 返回一个标准的 log.Logger，输出到标准错误
	return log.New(os.Stderr, "", 0)
}

// StandardWriter 实现 hclog.Logger 接口
func (h *HCLogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}


