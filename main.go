package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ai-production-pipeline/internal/config"
	"ai-production-pipeline/internal/gold"
	"ai-production-pipeline/internal/processor"
	"ai-production-pipeline/internal/silver"
	"ai-production-pipeline/internal/weekmanager"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Run the application
	if err := runAutomatedPipeline(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}

func runAutomatedPipeline(ctx context.Context) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger
	logger := setupLogger(cfg)
	logger.Info("=" + repeatString("=", 100))
	logger.Info("🚀 AUTOMATED AI PRODUCTION PIPELINE - MULTI-WEEK ANALYSIS")
	logger.Info("=" + repeatString("=", 100))

	// Get OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Connect to database
	db, err := connectDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize Week Manager
	weekMgr := weekmanager.NewWeekManager(db, logger)

	// Get all available weeks from database
	logger.Info("📅 Detecting available weeks from database...")
	weeks, err := weekMgr.GetAvailableWeeks()
	if err != nil {
		return fmt.Errorf("failed to get available weeks: %w", err)
	}

	if len(weeks) == 0 {
		return fmt.Errorf("no data found in database")
	}

	logger.Infof("✅ Found %d weeks of data", len(weeks))

	// Check if we should only process the last week (for testing)
	testMode := os.Getenv("TEST_LAST_WEEK_ONLY")
	if testMode == "true" || testMode == "1" {
		logger.Warn("⚠️  TEST MODE: Processing ONLY the last week")
		lastWeek := weeks[len(weeks)-1]
		weeks = []weekmanager.WeekRange{lastWeek}
	}

	// Initialize Silver Layer
	silverLayer := silver.NewSilverLayer(db, logger)

	// Initialize Gold Layer (for AI reports)
	goldLayer, err := gold.NewGoldLayer(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Gold layer: %w", err)
	}

	// Process each week
	for i, week := range weeks {
		weekNum := i + 1
		logger.Info("")
		logger.Info("=" + repeatString("=", 100))
		logger.Infof("📊 PROCESSING WEEK %d/%d: %s", weekNum, len(weeks), week.Label)
		logger.Info("=" + repeatString("=", 100))

		// Get week data with historical context
		weekData := weekMgr.GetWeekData(week, weeks)

		// Display context info
		if weekData.HasHistoricalData() {
			logger.Infof("📈 Historical data available:")
			if weekData.PreviousWeek != nil {
				logger.Infof("   - Previous week: %s", weekData.PreviousWeek.Label)
			}
			if weekData.TwoWeeksAgo != nil {
				logger.Infof("   - Two weeks ago: %s", weekData.TwoWeeksAgo.Label)
			}
		} else {
			logger.Warn("⚠️  First week - no historical comparison")
		}

		// Run Silver Layer V3: Enhanced transformation with trends
		logger.Info("")
		logger.Info("📂 Running Silver Layer V3: Enhanced Transformation")
		silverOutputPath := filepath.Join(cfg.Data.OutputDir, fmt.Sprintf("kids_analysis_week_%d.json", weekNum))
		if err := silverLayer.Transform(weekData, silverOutputPath); err != nil {
			return fmt.Errorf("silver layer failed for week %d: %w", weekNum, err)
		}

		// Run Gold Layer V2: AI Report Generation
		logger.Info("")
		logger.Info("📂 Running Gold Layer V2: AI Report Generation")

		// Generate reports for this week
		reportOutputPath := filepath.Join(cfg.Data.OutputDir, fmt.Sprintf("kids_reports_week_%d.json", weekNum))
		successCount, err := goldLayer.GenerateReportsFromFile(ctx, silverOutputPath, reportOutputPath, week.Label)
		if err != nil {
			logger.Errorf("❌ Gold layer failed for week %d: %v", weekNum, err)
			// Continue to next week instead of failing completely
			continue
		}

		logger.Infof("✅ Week %d completed: %d reports generated", weekNum, successCount)
		logger.Infof("   📄 Silver output: %s", silverOutputPath)
		logger.Infof("   📄 Gold output: %s", reportOutputPath)
	}

	// Final summary
	logger.Info("")
	logger.Info("=" + repeatString("=", 100))
	logger.Info("🎉 AUTOMATED PIPELINE COMPLETED SUCCESSFULLY")
	logger.Infof("📊 Processed %d weeks", len(weeks))
	logger.Info("=" + repeatString("=", 100))

	// Print token usage and cost report
	logger.Info("")
	goldLayer.GetAIProcessor().PrintTokenReport()

	return nil
}

// connectDatabase establishes database connection
func connectDatabase(cfg *config.Config) (*sql.DB, error) {
	connStr := cfg.Database.ConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetimeMin) * time.Minute)

	return db, nil
}

// createAIProcessor creates configured AI processor
func createAIProcessor(cfg *config.Config, apiKey string, logger *logrus.Logger) *processor.AIProcessor {
	processorConfig := processor.Config{
		APIKey:             apiKey,
		Model:              cfg.OpenAI.Model,
		MaxTokens:          cfg.OpenAI.MaxTokens,
		Temperature:        cfg.OpenAI.Temperature,
		Timeout:            time.Duration(cfg.OpenAI.TimeoutSeconds) * time.Second,
		BatchSize:          cfg.Batch.Size,
		MaxConcurrent:      cfg.Batch.MaxConcurrent,
		RateLimitPerMin:    cfg.RateLimit.RequestsPerMinute,
		MaxRetries:         cfg.Retry.MaxAttempts,
		InitialRetryDelay:  time.Duration(cfg.Retry.InitialDelaySeconds) * time.Second,
		MaxRetryDelay:      time.Duration(cfg.Retry.MaxDelaySeconds) * time.Second,
		ExponentialBackoff: cfg.Retry.ExponentialBackoff,
		TrackTokenUsage:    cfg.Monitoring.TrackTokenUsage,
		TrackTiming:        cfg.Monitoring.TrackTiming,
		ShowProgress:       cfg.Monitoring.ShowProgress,
	}

	return processor.NewAIProcessor(processorConfig, logger)
}

// setupLogger configures and returns a logger instance
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set output format
	if cfg.Logging.Output == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Setup file logging if enabled
	if cfg.Logging.LogToFile {
		if err := os.MkdirAll(cfg.Logging.LogDir, 0755); err != nil {
			logger.Warnf("Failed to create log directory: %v", err)
		} else {
			logFile := filepath.Join(cfg.Logging.LogDir, fmt.Sprintf("pipeline_%s.log", time.Now().Format("20060102_150405")))
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.Warnf("Failed to open log file: %v", err)
			} else {
				logger.SetOutput(file)
				logger.Infof("Logging to file: %s", logFile)
			}
		}
	}

	return logger
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
