package logging

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
)

// Config captures logging configuration options.
type Config struct {
	Level    string
	Dir      string
	Filename string
}

// Logger provides access to slog logging APIs.
type Logger struct {
	logger *slog.Logger
	writer *RotatableFileWriter
}

// New creates a new Logger instance.
func New(cfg Config) (*Logger, error) {
	// 1. Create RotatableFileWriter
	writer, err := NewRotatableFileWriter(cfg.Dir, cfg.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	// 2. Determine Log Level
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 3. Create Handlers
	jsonHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: level,
	})

	textHandler := NewCustomTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	multiHandler := NewMultiHandler(jsonHandler, textHandler)

	logger := slog.New(multiHandler)

	return &Logger{
		logger: logger,
		writer: writer,
	}, nil
}

// Legacy returns the logger itself for backward compatibility.
// Deprecated: Use the Logger methods directly.
func (l *Logger) Legacy() *Logger {
	return l
}

// Slog returns the underlying slog.Logger.
func (l *Logger) Slog() *slog.Logger {
	return l.logger
}

// Close closes the underlying log writer.
func (l *Logger) Close() error {
	return l.writer.Close()
}

// Helper to handle legacy args (map[string]interface{})
func (l *Logger) handleArgs(args []any) []any {
	if len(args) > 0 {
		if m, ok := args[0].(map[string]interface{}); ok {
			var newArgs []any
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				newArgs = append(newArgs, k, m[k])
			}
			return newArgs
		}
	}
	return args
}

func containsFormatPlaceholders(s string) bool {
	return strings.Contains(s, "%")
}

func (l *Logger) Debug(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		l.logger.Debug(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Debug(msg, l.handleArgs(args)...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		l.logger.Info(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Info(msg, l.handleArgs(args)...)
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		l.logger.Warn(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Warn(msg, l.handleArgs(args)...)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		l.logger.Error(fmt.Sprintf(msg, args...))
	} else {
		l.logger.Error(msg, l.handleArgs(args)...)
	}
}

// Tag helpers

func formatLog(tag, message string) string {
	tag = strings.TrimSpace(tag)
	message = strings.TrimSpace(message)
	if tag == "" {
		return message
	}
	if strings.HasPrefix(message, "[") {
		return message
	}
	return fmt.Sprintf("[%s] %s", tag, message)
}

func (l *Logger) DebugTag(tag, msg string, args ...any) {
	l.Debug(formatLog(tag, msg), args...)
}

func (l *Logger) InfoTag(tag, msg string, args ...any) {
	l.Info(formatLog(tag, msg), args...)
}

func (l *Logger) WarnTag(tag, msg string, args ...any) {
	l.Warn(formatLog(tag, msg), args...)
}

func (l *Logger) ErrorTag(tag, msg string, args ...any) {
	l.Error(formatLog(tag, msg), args...)
}

// Module helpers

func (l *Logger) InfoASR(msg string, args ...any) {
	l.Info("[ASR] "+msg, args...)
}

func (l *Logger) InfoLLM(msg string, args ...any) {
	l.Info("[LLM] "+msg, args...)
}

func (l *Logger) InfoTTS(msg string, args ...any) {
	l.Info("[TTS] "+msg, args...)
}

func (l *Logger) InfoTiming(msg string, args ...any) {
	l.Info("[TIMING] "+msg, args...)
}
