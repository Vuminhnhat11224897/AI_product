package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Config holds all processor configuration
type Config struct {
	// OpenAI settings
	APIKey        string
	Model         string
	MaxTokens     int
	Temperature   float64
	Timeout       time.Duration
	SystemMessage string // System message for AI model

	// Batch settings
	BatchSize     int
	MaxConcurrent int

	// Rate limit settings
	RateLimitPerMin int

	// Retry settings
	MaxRetries         int
	InitialRetryDelay  time.Duration
	MaxRetryDelay      time.Duration
	ExponentialBackoff bool

	// Monitoring
	TrackTokenUsage bool
	TrackTiming     bool
	ShowProgress    bool
}

// AIProcessor handles AI model calls with production-grade features
type AIProcessor struct {
	config       Config
	logger       *logrus.Logger
	httpClient   *http.Client
	rateLimiter  *RateLimiter
	tokenTracker *TokenTracker
}

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	tokens     chan struct{}
	refillRate time.Duration
	mu         sync.Mutex
}

// OpenAIRequest represents the API request structure
type OpenAIRequest struct {
	Model               string         `json:"model"`
	Messages            []Message      `json:"messages"`
	ResponseFormat      ResponseFormat `json:"response_format,omitempty"`
	Temperature         float64        `json:"temperature,omitempty"`
	MaxCompletionTokens int            `json:"max_completion_tokens,omitempty"` // Updated for newer models
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat specifies JSON response format
type ResponseFormat struct {
	Type string `json:"type"`
}

// OpenAIResponse represents the API response structure
type OpenAIResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ProcessResult contains the result of processing a single item
type ProcessResult struct {
	Index      int
	Input      interface{}
	Output     string
	Success    bool
	Error      error
	Retries    int
	Duration   time.Duration
	TokenUsage Usage
}

// NewAIProcessor creates a new AI processor instance with all production features
func NewAIProcessor(config Config, logger *logrus.Logger) *AIProcessor {
	// Set defaults if not provided
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.InitialRetryDelay == 0 {
		config.InitialRetryDelay = 2 * time.Second
	}
	if config.MaxRetryDelay == 0 {
		config.MaxRetryDelay = 10 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 5
	}
	if config.RateLimitPerMin == 0 {
		config.RateLimitPerMin = 60
	}

	logger.WithFields(logrus.Fields{
		"model":            config.Model,
		"batch_size":       config.BatchSize,
		"max_concurrent":   config.MaxConcurrent,
		"rate_limit":       config.RateLimitPerMin,
		"max_retries":      config.MaxRetries,
		"timeout":          config.Timeout,
		"exponential_back": config.ExponentialBackoff,
	}).Info("✅ AI Processor initialized")

	return &AIProcessor{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter:  NewRateLimiter(config.RateLimitPerMin, logger),
		tokenTracker: NewTokenTracker(config.Model),
	}
}

// NewRateLimiter creates a new token bucket rate limiter
func NewRateLimiter(requestsPerMinute int, logger *logrus.Logger) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, requestsPerMinute),
		refillRate: time.Minute / time.Duration(requestsPerMinute),
	}

	// Fill initial tokens
	for i := 0; i < requestsPerMinute; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refilling goroutine
	go rl.refill(logger)

	logger.WithField("rate_limit", requestsPerMinute).Info("✅ Rate limiter initialized")
	return rl
}

// refill continuously adds tokens to the bucket
func (rl *RateLimiter) refill(logger *logrus.Logger) {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
			// Token added successfully
		default:
			// Channel is full, skip
		}
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// GetTokenTracker returns the token tracker for reporting
func (ap *AIProcessor) GetTokenTracker() *TokenTracker {
	return ap.tokenTracker
}

// PrintTokenReport logs the detailed token usage report
func (ap *AIProcessor) PrintTokenReport() {
	report := ap.tokenTracker.GetDetailedReport()
	ap.logger.Info("\n" + report)
}

// ProcessSingleWithWeek processes a single prompt and returns response with week tracking
func (ap *AIProcessor) ProcessSingleWithWeek(ctx context.Context, prompt, systemMessage, weekLabel string) (string, error) {
	// Wait for rate limit token
	ap.rateLimiter.Wait()

	startTime := time.Now()

	// Call OpenAI with retry
	var response string
	var usage Usage
	var err error

	for attempt := 0; attempt < ap.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ap.calculateRetryDelay(attempt)
			ap.logger.Warnf("Retry attempt %d/%d after %v", attempt, ap.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		// Create request with system message
		fullPrompt := prompt
		if systemMessage != "" {
			fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s", systemMessage, prompt)
		}

		response, usage, err = ap.callOpenAI(ctx, fullPrompt)
		if err == nil {
			// Record token usage
			ap.tokenTracker.RecordUsage(weekLabel, usage.PromptTokens, usage.CompletionTokens)
			break
		}

		ap.logger.Warnf("Attempt %d failed: %v", attempt+1, err)
	}

	duration := time.Since(startTime)

	if err != nil {
		ap.logger.Errorf("All %d attempts failed: %v", ap.config.MaxRetries, err)
		return "", fmt.Errorf("failed after %d attempts: %w", ap.config.MaxRetries, err)
	}

	if ap.config.TrackTiming {
		ap.logger.Infof("✅ Processed in %v", duration)
	}

	return response, nil
}

// ProcessSingle processes a single prompt and returns response (legacy, without week tracking)
func (ap *AIProcessor) ProcessSingle(ctx context.Context, prompt, systemMessage string) (string, error) {
	return ap.ProcessSingleWithWeek(ctx, prompt, systemMessage, "unknown")
}

// ProcessSingleDeprecated is the old implementation kept for compatibility
func (ap *AIProcessor) ProcessSingleDeprecated(ctx context.Context, prompt, systemMessage string) (string, error) {
	// Wait for rate limit token
	ap.rateLimiter.Wait()

	startTime := time.Now()

	// Call OpenAI with retry
	var response string
	var err error

	for attempt := 0; attempt < ap.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ap.calculateRetryDelay(attempt)
			ap.logger.Warnf("Retry attempt %d/%d after %v", attempt, ap.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		// Create request with system message
		fullPrompt := prompt
		if systemMessage != "" {
			fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s", systemMessage, prompt)
		}

		response, _, err = ap.callOpenAI(ctx, fullPrompt)
		if err == nil {
			break
		}

		ap.logger.Warnf("Attempt %d failed: %v", attempt+1, err)
	}

	duration := time.Since(startTime)

	if err != nil {
		ap.logger.Errorf("All %d attempts failed: %v", ap.config.MaxRetries, err)
		return "", fmt.Errorf("failed after %d attempts: %w", ap.config.MaxRetries, err)
	}

	if ap.config.TrackTiming {
		ap.logger.Infof("✅ Processed in %v", duration)
	}

	return response, nil
}

// ProcessBatch processes multiple items in batches with controlled concurrency and resilience
func (ap *AIProcessor) ProcessBatch(ctx context.Context, items []interface{}, promptTemplate func(interface{}) string) []ProcessResult {
	ap.logger.WithFields(logrus.Fields{
		"total_items":    len(items),
		"batch_size":     ap.config.BatchSize,
		"max_concurrent": ap.config.MaxConcurrent,
	}).Info("🚀 Starting batch processing")

	results := make([]ProcessResult, len(items))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, ap.config.MaxConcurrent)

	startTime := time.Now()
	processedCount := 0
	var progressMu sync.Mutex

	// Process in batches
	totalBatches := (len(items) + ap.config.BatchSize - 1) / ap.config.BatchSize
	for batchStart := 0; batchStart < len(items); batchStart += ap.config.BatchSize {
		batchEnd := batchStart + ap.config.BatchSize
		if batchEnd > len(items) {
			batchEnd = len(items)
		}

		batchNum := (batchStart / ap.config.BatchSize) + 1
		ap.logger.WithFields(logrus.Fields{
			"batch_num":   batchNum,
			"total":       totalBatches,
			"batch_start": batchStart,
			"batch_end":   batchEnd,
			"batch_items": batchEnd - batchStart,
		}).Info("📦 Processing batch")

		// Process items in current batch concurrently
		for i := batchStart; i < batchEnd; i++ {
			wg.Add(1)
			go func(index int, item interface{}) {
				defer wg.Done()

				// Acquire semaphore slot
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Check context cancellation
				if ctx.Err() != nil {
					results[index] = ProcessResult{
						Index:   index,
						Input:   item,
						Success: false,
						Error:   ctx.Err(),
					}
					return
				}

				// Process item with retry logic
				result := ap.processItemWithRetry(ctx, index, item, promptTemplate)
				results[index] = result

				// Update progress
				if ap.config.ShowProgress {
					progressMu.Lock()
					processedCount++
					progress := float64(processedCount) / float64(len(items)) * 100
					ap.logger.WithFields(logrus.Fields{
						"processed": processedCount,
						"total":     len(items),
						"progress":  fmt.Sprintf("%.1f%%", progress),
					}).Info("📊 Progress update")
					progressMu.Unlock()
				}

			}(i, items[i])
		}

		// Wait for current batch to complete before starting next batch
		wg.Wait()

		ap.logger.WithFields(logrus.Fields{
			"batch_num":       batchNum,
			"items_completed": batchEnd,
		}).Info("✅ Batch completed")
	}

	duration := time.Since(startTime)

	// Calculate summary statistics
	successful := 0
	failed := 0
	totalRetries := 0
	totalTokens := 0

	for _, result := range results {
		if result.Success {
			successful++
			totalTokens += result.TokenUsage.TotalTokens
		} else {
			failed++
		}
		totalRetries += result.Retries
	}

	ap.logger.Info("=" + strings.Repeat("=", 100))
	ap.logger.WithFields(logrus.Fields{
		"total_items":    len(items),
		"successful":     successful,
		"failed":         failed,
		"success_rate":   fmt.Sprintf("%.2f%%", float64(successful)/float64(len(items))*100),
		"total_retries":  totalRetries,
		"total_tokens":   totalTokens,
		"total_duration": duration,
		"avg_per_item":   duration / time.Duration(len(items)),
	}).Info("🎉 BATCH PROCESSING COMPLETED")
	ap.logger.Info("=" + strings.Repeat("=", 100))

	return results
}

// processItemWithRetry processes a single item with retry logic and exponential backoff
func (ap *AIProcessor) processItemWithRetry(ctx context.Context, index int, item interface{}, promptTemplate func(interface{}) string) ProcessResult {
	startTime := time.Now()
	var lastError error
	retryCount := 0

	for attempt := 0; attempt <= ap.config.MaxRetries; attempt++ {
		// Check context before attempting
		if ctx.Err() != nil {
			return ProcessResult{
				Index:    index,
				Input:    item,
				Success:  false,
				Error:    ctx.Err(),
				Retries:  retryCount,
				Duration: time.Since(startTime),
			}
		}

		// Wait for rate limiter
		ap.rateLimiter.Wait()

		// Generate prompt
		prompt := promptTemplate(item)
		if prompt == "" {
			return ProcessResult{
				Index:    index,
				Input:    item,
				Success:  false,
				Error:    fmt.Errorf("empty prompt generated"),
				Retries:  0,
				Duration: time.Since(startTime),
			}
		}

		// Call OpenAI API
		output, usage, err := ap.callOpenAI(ctx, prompt)
		if err == nil {
			// Success
			duration := time.Since(startTime)
			ap.logger.WithFields(logrus.Fields{
				"index":    index,
				"retries":  retryCount,
				"duration": duration,
				"tokens":   usage.TotalTokens,
			}).Info("✅ Item processed successfully")

			return ProcessResult{
				Index:      index,
				Input:      item,
				Output:     output,
				Success:    true,
				Retries:    retryCount,
				Duration:   duration,
				TokenUsage: usage,
			}
		}

		// Handle error
		lastError = err
		retryCount++

		if attempt < ap.config.MaxRetries {
			// Calculate retry delay
			delay := ap.calculateRetryDelay(attempt)

			ap.logger.WithFields(logrus.Fields{
				"index":        index,
				"attempt":      attempt + 1,
				"max_attempts": ap.config.MaxRetries + 1,
				"error":        err.Error(),
				"retry_in":     delay,
			}).Warn("⚠️ Request failed, retrying...")

			// Wait before retry
			select {
			case <-time.After(delay):
				// Continue to retry
			case <-ctx.Done():
				return ProcessResult{
					Index:    index,
					Input:    item,
					Success:  false,
					Error:    ctx.Err(),
					Retries:  retryCount,
					Duration: time.Since(startTime),
				}
			}
		}
	}

	// All retries exhausted
	duration := time.Since(startTime)
	ap.logger.WithFields(logrus.Fields{
		"index":    index,
		"retries":  retryCount,
		"duration": duration,
		"error":    lastError.Error(),
	}).Error("❌ Item processing failed after all retries")

	return ProcessResult{
		Index:    index,
		Input:    item,
		Success:  false,
		Error:    lastError,
		Retries:  retryCount,
		Duration: duration,
	}
}

// calculateRetryDelay calculates the delay before next retry
func (ap *AIProcessor) calculateRetryDelay(attempt int) time.Duration {
	if !ap.config.ExponentialBackoff {
		return ap.config.InitialRetryDelay
	}

	// Exponential backoff: delay = initialDelay * 2^attempt
	delay := ap.config.InitialRetryDelay * time.Duration(1<<uint(attempt))
	if delay > ap.config.MaxRetryDelay {
		delay = ap.config.MaxRetryDelay
	}
	return delay
}

// callOpenAI makes a call to the OpenAI API
func (ap *AIProcessor) callOpenAI(ctx context.Context, prompt string) (string, Usage, error) {
	// Use configured system message or default
	systemMsg := ap.config.SystemMessage
	if systemMsg == "" {
		systemMsg = "Bạn là chuyên gia phân tích dữ liệu dành cho ứng dụng giáo dục tài chính trẻ em. Trả về CHÍNH XÁC định dạng JSON được yêu cầu, không thêm markdown hay text khác."
	}

	// Prepare request
	reqBody := OpenAIRequest{
		Model: ap.config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemMsg,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat:      ResponseFormat{Type: "json_object"},
		Temperature:         ap.config.Temperature,
		MaxCompletionTokens: ap.config.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ap.config.APIKey)

	// Execute request
	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var apiResp OpenAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", Usage{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if apiResp.Error != nil {
		return "", Usage{}, fmt.Errorf("API error: %s (%s)", apiResp.Error.Message, apiResp.Error.Type)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", Usage{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Extract content
	if len(apiResp.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("no choices in response")
	}

	content := apiResp.Choices[0].Message.Content
	usage := apiResp.Usage

	return content, usage, nil
}
