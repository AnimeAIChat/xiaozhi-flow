package utils

import (
"xiaozhi-server-go/internal/platform/logging"
)

// Logger alias to platform/logging.Logger
type Logger = logging.Logger

// LogCfg configuration for legacy logger
type LogCfg struct {
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogDir   string `yaml:"log_dir" json:"log_dir"`
	LogFile  string `yaml:"log_file" json:"log_file"`
}

var DefaultLogger *Logger

// NewLogger creates a new logger using the platform logging implementation
func NewLogger(cfg *LogCfg) (*Logger, error) {
return logging.New(logging.Config{
Level:    cfg.LogLevel,
Dir:      cfg.LogDir,
Filename: cfg.LogFile,
})
}
