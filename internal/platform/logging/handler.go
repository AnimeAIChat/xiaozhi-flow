package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

// CustomTextHandler 自定义文本处理器，支持彩色输出和格式化
type CustomTextHandler struct {
	writer io.Writer
	level  slog.Level
	mu     sync.Mutex
}

var (
	colorReset  = "\x1b[0m"
	colorTime   = "\x1b[90m" // 时间：灰色
	colorDebug  = "\x1b[36m" // DEBUG：青色
	colorInfo   = "\x1b[32m" // INFO：绿色
	colorWarn   = "\x1b[33m" // WARN：黄色
	colorError  = "\x1b[31m" // ERROR：红色
	colorASR    = "\x1b[35m" // ASR：品红
	colorLLM    = "\x1b[34m" // LLM：蓝色
	colorTTS    = "\x1b[95m" // TTS：亮品红
	colorTiming = "\x1b[92m" // Timing：亮绿色
)

func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}
	return &CustomTextHandler{
		writer: w,
		level:  level,
	}
}

func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 获取时间戳
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")

	// 获取日志级别中文描述
	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = "调试"
	case slog.LevelInfo:
		levelStr = "信息"
	case slog.LevelWarn:
		levelStr = "警告"
	case slog.LevelError:
		levelStr = "错误"
	default:
		levelStr = "信息"
	}

	// 应用颜色
	var levelColor string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = colorDebug
	case slog.LevelInfo:
		levelColor = colorInfo
	case slog.LevelWarn:
		levelColor = colorWarn
	case slog.LevelError:
		levelColor = colorError
	default:
		levelColor = colorReset
	}

	// 检查是否是特殊阶段日志或模块日志
	var moduleColor string
	var isModuleLog bool
	msg := r.Message

	// 检测各种模块标签
	if strings.HasPrefix(msg, "[引导]") {
		moduleColor = "\x1b[96m" // 引导：亮青色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[传输]") {
		moduleColor = "\x1b[94m" // 传输：亮蓝色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[HTTP]") {
		moduleColor = "\x1b[95m" // HTTP：亮品红
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[WebSocket]") {
		moduleColor = "\x1b[92m" // WebSocket：亮绿色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[ASR]") {
		moduleColor = colorASR
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[LLM]") {
		moduleColor = colorLLM
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[TTS]") {
		moduleColor = colorTTS
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[TIMING]") {
		moduleColor = colorTiming
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[MCP]") {
		moduleColor = "\x1b[36m" // MCP：青蓝色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[认证]") {
		moduleColor = "\x1b[91m" // 认证：亮红色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[视觉]") {
		moduleColor = "\x1b[95m" // 视觉：亮品红
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[OTA]") {
		moduleColor = "\x1b[97m" // OTA：亮白色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[WebAPI]") {
		moduleColor = "\x1b[96m" // WebAPI：亮青色
		isModuleLog = true
	} else if strings.HasPrefix(msg, "[OBSERVABILITY]") {
		moduleColor = "\x1b[90m" // 可观测性：灰色
		isModuleLog = true
	}

	// 构建输出
	var output string
	if isModuleLog {
		// 模块日志格式: [时间] [模块] 消息
		output = fmt.Sprintf("%s[%s]%s %s%s%s",
			colorTime, timeStr, colorReset,
			moduleColor, msg, colorReset)
	} else {
		// 普通日志格式: [时间] [级别] 消息
		output = fmt.Sprintf("%s[%s]%s %s[%s]%s %s",
			colorTime, timeStr, colorReset,
			levelColor, levelStr, colorReset,
			msg)
	}

	// 添加属性（如果有）
	if r.NumAttrs() > 0 {
		output += " {"
		r.Attrs(func(a slog.Attr) bool {
			output += fmt.Sprintf(" %s=%v", a.Key, a.Value)
			return true
		})
		output += " }"
	}
	output += "\n"

	_, err := h.writer.Write([]byte(output))
	return err
}

func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // 简化实现
}

func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	return h // 简化实现
}

// MultiHandler 分发日志到多个 Handler
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []string
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multiple handler errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
