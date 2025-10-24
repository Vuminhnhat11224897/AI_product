package gold

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-production-pipeline/internal/config"
	"ai-production-pipeline/internal/processor"

	"github.com/sirupsen/logrus"
)

// GoldLayer handles AI inference with enhanced prompts
type GoldLayer struct {
	config         *config.Config
	logger         *logrus.Logger
	aiProcessor    *processor.AIProcessor
	promptTemplate string // Cached prompt template from file
	systemMessage  string // Cached system message from file
}

// GetAIProcessor returns the AI processor for external access (e.g., token reporting)
func (gl *GoldLayer) GetAIProcessor() *processor.AIProcessor {
	return gl.aiProcessor
}

// KidDataV2 represents enriched kid data for AI prompt
type KidDataV2 struct {
	Nickname           string  `json:"nickname"`
	Age                int     `json:"age"`
	JoyWallet          float64 `json:"joy_wallet"`
	SpendingWallet     float64 `json:"spending_wallet"`
	CharityWallet      float64 `json:"charity_wallet"`
	StudyWallet        float64 `json:"study_wallet"`
	MoneyReceived      float64 `json:"money_received"`
	MoneyReceivedCount int     `json:"money_received_count"`
	JoySpent           float64 `json:"joy_spent"`
	SpendingSpent      float64 `json:"spending_spent"`
	CharitySpent       float64 `json:"charity_spent"`
	StudySpent         float64 `json:"study_spent"`
	MissionsCompleted  int     `json:"missions_completed"`
	MissionsTotal      int     `json:"missions_total"`
	ActivityScore      float64 `json:"activity_score"`
}

// AIReport represents the structured Vietnamese AI report for a kid
type AIReport struct {
	ChildName           string               `json:"child_name"`
	Week                string               `json:"week"`
	FinancialTendencies []FinancialTendency  `json:"financial_tendencies"`
	PerformanceSections []PerformanceSection `json:"performance_sections"`
	NextWeekGoals       []string             `json:"next_week_goals"`
	ParentSuggestions   []string             `json:"parent_suggestions"`
	GeneratedAt         string               `json:"generated_at"`
}

// FinancialTendency represents a financial behavior tendency
type FinancialTendency struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// PerformanceSection represents a performance evaluation section
type PerformanceSection struct {
	Title   string `json:"title"`
	Level   string `json:"level"`
	Score   int    `json:"score"`
	Summary string `json:"summary"`
}

func NewGoldLayer(cfg *config.Config, logger *logrus.Logger) (*GoldLayer, error) {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Load prompt template from file
	promptTemplate, err := loadPromptTemplate(cfg.Prompts.TemplateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}
	logger.WithField("template_file", cfg.Prompts.TemplateFile).Info("✅ Loaded prompt template")

	// Load system message from file
	systemMessage, err := loadSystemMessage(cfg.Prompts.SystemMessageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load system message: %w", err)
	}
	logger.WithField("system_message_file", cfg.Prompts.SystemMessageFile).Info("✅ Loaded system message")

	// Configure AI Processor
	aiConfig := processor.Config{
		APIKey:             apiKey,
		Model:              cfg.OpenAI.Model, // Use model from config
		MaxTokens:          cfg.OpenAI.MaxTokens,
		Temperature:        cfg.OpenAI.Temperature,
		MaxRetries:         cfg.Retry.MaxAttempts,
		InitialRetryDelay:  time.Duration(cfg.Retry.InitialDelaySeconds) * time.Second,
		MaxRetryDelay:      time.Duration(cfg.Retry.MaxDelaySeconds) * time.Second,
		ExponentialBackoff: cfg.Retry.ExponentialBackoff,
		Timeout:            time.Duration(cfg.OpenAI.TimeoutSeconds) * time.Second,
		BatchSize:          cfg.Batch.Size,
		MaxConcurrent:      cfg.Batch.MaxConcurrent,
		RateLimitPerMin:    cfg.RateLimit.RequestsPerMinute,
		TrackTokenUsage:    cfg.Monitoring.TrackTokenUsage,
		TrackTiming:        cfg.Monitoring.TrackTiming,
		ShowProgress:       cfg.Monitoring.ShowProgress,
		SystemMessage:      systemMessage, // Pass loaded system message
	}

	aiProcessor := processor.NewAIProcessor(aiConfig, logger)

	logger.Info("✅ Gold Layer V2 initialized successfully")
	logger.WithFields(logrus.Fields{
		"model":          aiConfig.Model,
		"batch_size":     aiConfig.BatchSize,
		"max_concurrent": aiConfig.MaxConcurrent,
		"rate_limit":     aiConfig.RateLimitPerMin,
		"max_retries":    aiConfig.MaxRetries,
	}).Info("AI Processor V2 Configuration")

	return &GoldLayer{
		config:         cfg,
		logger:         logger,
		aiProcessor:    aiProcessor,
		promptTemplate: promptTemplate,
		systemMessage:  systemMessage,
	}, nil
}

// GenerateReports generates AI reports using enhanced prompts
func (gl *GoldLayer) GenerateReports(ctx context.Context) (int, int, error) {
	gl.logger.Info("==============================================================================================================")
	gl.logger.Info("GOLD LAYER V2: AI REPORT GENERATION WITH ENHANCED PROMPTS")
	gl.logger.Info("==============================================================================================================")
	startTime := time.Now()

	// Read Silver layer output
	inputPath := filepath.Join("data", "kids_analysis.json")
	gl.logger.Infof("📖 Reading Silver layer output from: %s", inputPath)

	kidsData, err := gl.readSilverData(inputPath)
	if err != nil {
		return 0, 0, err
	}

	gl.logger.Infof("✅ Total kids found in Silver layer: %d", len(kidsData))

	// Convert kids data to interface slice for processing
	items := make([]interface{}, len(kidsData))
	for i, kid := range kidsData {
		items[i] = kid
	}

	// Define prompt template function
	promptTemplate := func(item interface{}) string {
		kid, ok := item.(KidDataV2)
		if !ok {
			return ""
		}
		return gl.createEnhancedPromptForKid(kid)
	}

	// Process all kids with batching and controlled concurrency
	gl.logger.Info("🚀 Starting AI batch processing...")
	results := gl.aiProcessor.ProcessBatch(ctx, items, promptTemplate)

	// Parse successful results into reports
	reports := []AIReport{}
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			var report AIReport
			if err := json.Unmarshal([]byte(result.Output), &report); err != nil {
				gl.logger.WithFields(logrus.Fields{
					"index": result.Index,
					"error": err,
				}).Error("Failed to parse AI report")
				continue
			}

			// Add metadata
			report.GeneratedAt = time.Now().Format(time.RFC3339)

			reports = append(reports, report)
		}
	}

	// Save reports to file
	if err := gl.saveReports(reports); err != nil {
		return successCount, len(kidsData), err
	}

	// Final summary
	duration := time.Since(startTime)
	gl.logger.Info("==============================================================================================================")
	gl.logger.WithFields(logrus.Fields{
		"total_kids":        len(kidsData),
		"reports_generated": len(reports),
		"success_rate":      fmt.Sprintf("%.2f%%", float64(len(reports))/float64(len(kidsData))*100),
		"total_duration":    duration,
		"avg_per_kid":       duration / time.Duration(len(kidsData)),
	}).Info("🎉 GOLD LAYER V2 PROCESSING COMPLETED")
	gl.logger.Info("==============================================================================================================")

	return successCount, len(kidsData), nil
}

// readSilverData reads and parses the Silver layer output
func (gl *GoldLayer) readSilverData(inputPath string) ([]KidDataV2, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", inputPath, err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract kids array
	kidsArray, ok := rawData["kids"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid kids data format")
	}

	// Convert to KidDataV2 structs
	kids := make([]KidDataV2, 0, len(kidsArray))
	for _, kidInterface := range kidsArray {
		kidMap, ok := kidInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse profile
		profileMap, _ := kidMap["profile"].(map[string]interface{})

		// Parse wallet balances
		joyWallet := 0.0
		spendingWallet := 0.0
		charityWallet := 0.0
		studyWallet := 0.0

		if wallets, ok := kidMap["wallets"].([]interface{}); ok {
			for _, w := range wallets {
				wm, _ := w.(map[string]interface{})
				wtype, _ := wm["wallet_type"].(string)
				balance := getFloat64(wm, "balance")

				switch wtype {
				case "joy":
					joyWallet = balance
				case "spending":
					spendingWallet = balance
				case "charity":
					charityWallet = balance
				case "study":
					studyWallet = balance
				}
			}
		}

		// Parse wallet transactions
		joySpent := 0.0
		spendingSpent := 0.0
		charitySpent := 0.0
		studySpent := 0.0

		if wtrans, ok := kidMap["wallet_transactions"].([]interface{}); ok {
			for _, wt := range wtrans {
				wtm, _ := wt.(map[string]interface{})
				wtype, _ := wtm["wallet_type"].(string)
				spent := getFloat64(wtm, "total_spent")

				switch wtype {
				case "joy":
					joySpent = spent
				case "spending":
					spendingSpent = spent
				case "charity":
					charitySpent = spent
				case "study":
					studySpent = spent
				}
			}
		}

		// Parse missions
		missionsMap, _ := kidMap["missions"].(map[string]interface{})

		kid := KidDataV2{
			Nickname:           getString(profileMap, "nickname"),
			Age:                int(getFloat64(profileMap, "age")),
			JoyWallet:          joyWallet,
			SpendingWallet:     spendingWallet,
			CharityWallet:      charityWallet,
			StudyWallet:        studyWallet,
			MoneyReceived:      getFloat64(kidMap, "money_received"),
			MoneyReceivedCount: int(getFloat64(kidMap, "money_received_count")),
			JoySpent:           joySpent,
			SpendingSpent:      spendingSpent,
			CharitySpent:       charitySpent,
			StudySpent:         studySpent,
			MissionsCompleted:  int(getFloat64(missionsMap, "completed_missions")),
			MissionsTotal:      int(getFloat64(missionsMap, "total_missions")),
			ActivityScore:      getFloat64(kidMap, "activity_score"),
		}

		kids = append(kids, kid)
	}

	return kids, nil
}

// createEnhancedPromptForKid creates detailed Vietnamese prompt for financial education app
func (gl *GoldLayer) createEnhancedPromptForKid(kid KidDataV2) string {
	// Convert kid data to JSON for prompt
	kidJSON, _ := json.MarshalIndent(kid, "", "  ")

	// Replace placeholders in template
	prompt := gl.promptTemplate
	prompt = strings.ReplaceAll(prompt, "{{KIDS_DATA}}", string(kidJSON))
	prompt = strings.ReplaceAll(prompt, "{{CHILD_NAME}}", kid.Nickname)
	prompt = strings.ReplaceAll(prompt, "{{WEEK}}", gl.config.Prompts.Week)

	return prompt
}

// loadPromptTemplate loads prompt template from file
func loadPromptTemplate(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template file: %w", err)
	}
	return string(data), nil
}

// loadSystemMessage loads system message from file
func loadSystemMessage(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read system message file: %w", err)
	}
	return string(data), nil
}

// GenerateReportsFromFile reads Silver V3 output and generates AI reports
func (gl *GoldLayer) GenerateReportsFromFile(ctx context.Context, silverOutputPath, reportOutputPath, weekLabel string) (int, error) {
	gl.logger.Infof("📖 Loading Silver V3 data from: %s", silverOutputPath)

	// Read Silver V3 JSON output
	data, err := os.ReadFile(silverOutputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read silver output: %w", err)
	}

	var silverData map[string]interface{}
	if err := json.Unmarshal(data, &silverData); err != nil {
		return 0, fmt.Errorf("failed to parse silver output: %w", err)
	}

	kids, ok := silverData["kids"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid silver output format: missing 'kids' array")
	}

	gl.logger.Infof("✅ Loaded %d kids from Silver V3", len(kids))

	// Generate reports for each kid
	var reports []AIReport
	successCount := 0

	for i, kidData := range kids {
		kidMap, ok := kidData.(map[string]interface{})
		if !ok {
			gl.logger.Warnf("Skipping invalid kid data at index %d", i)
			continue
		}

		nickname := getString(kidMap, "nickname")
		gl.logger.Infof("   Processing: %s (%d/%d)", nickname, i+1, len(kids))

		// Convert to KidDataV2 format for existing prompt system
		kid := gl.convertEnhancedToV2(kidMap, weekLabel)

		// Generate AI report with week label for token tracking
		report, err := gl.generateReportForKid(ctx, kid, weekLabel)
		if err != nil {
			gl.logger.Errorf("   ❌ Failed to generate report for %s: %v", nickname, err)
			continue
		}

		reports = append(reports, *report)
		successCount++
		gl.logger.Infof("   ✅ Completed: %s", nickname)
	}

	// Save reports to specified output path
	if err := gl.saveReportsToPath(reports, reportOutputPath, weekLabel); err != nil {
		return successCount, fmt.Errorf("failed to save reports: %w", err)
	}

	gl.logger.Infof("✅ Generated %d/%d reports successfully", successCount, len(kids))
	return successCount, nil
}

// convertEnhancedToV2 converts Silver V3 enhanced data to V2 format
func (gl *GoldLayer) convertEnhancedToV2(kidMap map[string]interface{}, weekLabel string) KidDataV2 {
	// Get current week data
	currentWeek, _ := kidMap["current_week"].(map[string]interface{})

	return KidDataV2{
		Nickname:           getString(kidMap, "nickname"),
		Age:                int(getFloat64(kidMap, "age")),
		JoyWallet:          getFloat64(currentWeek, "joy_wallet"),
		SpendingWallet:     getFloat64(currentWeek, "spending_wallet"),
		CharityWallet:      getFloat64(currentWeek, "charity_wallet"),
		StudyWallet:        getFloat64(currentWeek, "study_wallet"),
		MoneyReceived:      getFloat64(currentWeek, "money_received"),
		MoneyReceivedCount: int(getFloat64(currentWeek, "money_received_count")),
		JoySpent:           getFloat64(currentWeek, "joy_spent"),
		SpendingSpent:      getFloat64(currentWeek, "spending_spent"),
		CharitySpent:       getFloat64(currentWeek, "charity_spent"),
		StudySpent:         getFloat64(currentWeek, "study_spent"),
		MissionsCompleted:  int(getFloat64(currentWeek, "missions_completed")),
		MissionsTotal:      int(getFloat64(currentWeek, "missions_total")),
		ActivityScore:      getFloat64(kidMap, "activity_score"),
	}
}

// generateReportForKid generates report for a single kid
func (gl *GoldLayer) generateReportForKid(ctx context.Context, kid KidDataV2, weekLabel string) (*AIReport, error) {
	// Create prompt
	prompt := gl.createEnhancedPromptForKid(kid)

	// Call AI with week tracking
	response, err := gl.aiProcessor.ProcessSingleWithWeek(ctx, prompt, gl.systemMessage, weekLabel)
	if err != nil {
		return nil, err
	}

	// Parse response
	var report AIReport
	if err := json.Unmarshal([]byte(response), &report); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	report.GeneratedAt = time.Now().Format(time.RFC3339)
	return &report, nil
}

// saveReportsToPath saves reports to a specific file path
func (gl *GoldLayer) saveReportsToPath(reports []AIReport, outputPath, weekLabel string) error {
	output := map[string]interface{}{
		"generated_at":  time.Now().Format(time.RFC3339),
		"week":          weekLabel,
		"total_reports": len(reports),
		"reports":       reports,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reports: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	gl.logger.Infof("✅ Reports saved to: %s", outputPath)
	return nil
}

// saveReports saves the generated reports to a JSON file
func (gl *GoldLayer) saveReports(reports []AIReport) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("kids_report_%s.json", timestamp)
	outputPath := filepath.Join("data", filename)

	output := map[string]interface{}{
		"generated_at":  time.Now().Format(time.RFC3339),
		"total_reports": len(reports),
		"reports":       reports,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reports: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	gl.logger.WithField("output_file", outputPath).Info("✅ Reports saved successfully")
	return nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}
