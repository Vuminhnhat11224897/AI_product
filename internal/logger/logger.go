package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var globalLogger *logrus.Logger

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level     string
	Output    string
	LogToFile bool
	LogDir    string
}

// InitLogger initializes the global logger
func InitLogger(cfg *LoggingConfig) error {
	globalLogger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", cfg.Level, err)
	}
	globalLogger.SetLevel(level)

	// Configure output
	if cfg.LogToFile && cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		timestamp := time.Now().Format("20060102_150405")
		logFile := filepath.Join(cfg.LogDir, fmt.Sprintf("pipeline_%s.log", timestamp))

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}

		// Log to both console and file
		globalLogger.SetOutput(io.MultiWriter(os.Stdout, file))
		globalLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     false,
		})
	} else {
		globalLogger.SetOutput(os.Stdout)
		globalLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if globalLogger == nil {
		// Create default logger if not initialized
		globalLogger = logrus.New()
		globalLogger.SetLevel(logrus.InfoLevel)
		globalLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}
	return globalLogger
}
