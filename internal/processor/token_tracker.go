package processor

import (
	"fmt"
	"sync"
	"time"
)

// TokenUsage tracks token usage and costs
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	EstimatedCost    float64
	Timestamp        time.Time
}

// TokenTracker tracks all token usage across pipeline runs
type TokenTracker struct {
	mu          sync.RWMutex
	usageByWeek map[string][]TokenUsage
	totalUsage  TokenUsage
	model       string

	// GPT-4o pricing (as of 2024)
	// Input: $2.50 per 1M tokens
	// Output: $10.00 per 1M tokens
	inputPricePer1M  float64
	outputPricePer1M float64
}

// NewTokenTracker creates a new token tracker
func NewTokenTracker(model string) *TokenTracker {
	// Set pricing based on model
	inputPrice, outputPrice := getPricing(model)

	return &TokenTracker{
		usageByWeek:      make(map[string][]TokenUsage),
		model:            model,
		inputPricePer1M:  inputPrice,
		outputPricePer1M: outputPrice,
	}
}

// getPricing returns pricing for different models
func getPricing(model string) (input, output float64) {
	switch model {
	case "gpt-4o", "gpt-4o-2024-08-06":
		return 2.50, 10.00 // $2.50 input, $10.00 output per 1M tokens
	case "gpt-4o-mini":
		return 0.15, 0.60 // $0.15 input, $0.60 output per 1M tokens
	case "gpt-4-turbo", "gpt-4-turbo-2024-04-09":
		return 10.00, 30.00 // $10.00 input, $30.00 output per 1M tokens
	case "gpt-3.5-turbo":
		return 0.50, 1.50 // $0.50 input, $1.50 output per 1M tokens
	default:
		// Default to GPT-4o pricing
		return 2.50, 10.00
	}
}

// RecordUsage records token usage for a request
func (tt *TokenTracker) RecordUsage(weekLabel string, promptTokens, completionTokens int) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	totalTokens := promptTokens + completionTokens

	// Calculate cost
	inputCost := float64(promptTokens) * tt.inputPricePer1M / 1_000_000
	outputCost := float64(completionTokens) * tt.outputPricePer1M / 1_000_000
	totalCost := inputCost + outputCost

	usage := TokenUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		EstimatedCost:    totalCost,
		Timestamp:        time.Now(),
	}

	// Add to week-specific tracking
	tt.usageByWeek[weekLabel] = append(tt.usageByWeek[weekLabel], usage)

	// Update total
	tt.totalUsage.PromptTokens += promptTokens
	tt.totalUsage.CompletionTokens += completionTokens
	tt.totalUsage.TotalTokens += totalTokens
	tt.totalUsage.EstimatedCost += totalCost
}

// GetWeekSummary returns summary for a specific week
func (tt *TokenTracker) GetWeekSummary(weekLabel string) TokenUsage {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	weekUsages := tt.usageByWeek[weekLabel]
	if len(weekUsages) == 0 {
		return TokenUsage{}
	}

	var summary TokenUsage
	for _, usage := range weekUsages {
		summary.PromptTokens += usage.PromptTokens
		summary.CompletionTokens += usage.CompletionTokens
		summary.TotalTokens += usage.TotalTokens
		summary.EstimatedCost += usage.EstimatedCost
	}

	return summary
}

// GetTotalSummary returns total summary across all weeks
func (tt *TokenTracker) GetTotalSummary() TokenUsage {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	return tt.totalUsage
}

// GetDetailedReport returns detailed report string
func (tt *TokenTracker) GetDetailedReport() string {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	report := fmt.Sprintf("📊 TOKEN USAGE & COST REPORT\n")
	report += fmt.Sprintf("=" + repeatString("=", 80) + "\n")
	report += fmt.Sprintf("Model: %s\n", tt.model)
	report += fmt.Sprintf("Pricing: Input $%.2f/1M tokens | Output $%.2f/1M tokens\n\n",
		tt.inputPricePer1M, tt.outputPricePer1M)

	// Per-week breakdown
	report += fmt.Sprintf("📅 Per-Week Breakdown:\n")
	for weekLabel, usages := range tt.usageByWeek {
		if len(usages) == 0 {
			continue
		}

		var weekTotal TokenUsage
		for _, usage := range usages {
			weekTotal.PromptTokens += usage.PromptTokens
			weekTotal.CompletionTokens += usage.CompletionTokens
			weekTotal.TotalTokens += usage.TotalTokens
			weekTotal.EstimatedCost += usage.EstimatedCost
		}

		report += fmt.Sprintf("\n   %s (%d requests):\n", weekLabel, len(usages))
		report += fmt.Sprintf("      Input:  %7d tokens ($%.4f)\n",
			weekTotal.PromptTokens,
			float64(weekTotal.PromptTokens)*tt.inputPricePer1M/1_000_000)
		report += fmt.Sprintf("      Output: %7d tokens ($%.4f)\n",
			weekTotal.CompletionTokens,
			float64(weekTotal.CompletionTokens)*tt.outputPricePer1M/1_000_000)
		report += fmt.Sprintf("      Total:  %7d tokens ($%.4f)\n",
			weekTotal.TotalTokens, weekTotal.EstimatedCost)
	}

	// Total summary
	report += fmt.Sprintf("\n" + repeatString("=", 80) + "\n")
	report += fmt.Sprintf("💰 TOTAL SUMMARY:\n")
	report += fmt.Sprintf("   Input tokens:      %10d ($%.4f)\n",
		tt.totalUsage.PromptTokens,
		float64(tt.totalUsage.PromptTokens)*tt.inputPricePer1M/1_000_000)
	report += fmt.Sprintf("   Output tokens:     %10d ($%.4f)\n",
		tt.totalUsage.CompletionTokens,
		float64(tt.totalUsage.CompletionTokens)*tt.outputPricePer1M/1_000_000)
	report += fmt.Sprintf("   Total tokens:      %10d\n", tt.totalUsage.TotalTokens)
	report += fmt.Sprintf("   Estimated cost:    $%.4f USD\n", tt.totalUsage.EstimatedCost)
	report += fmt.Sprintf("=" + repeatString("=", 80) + "\n")

	return report
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
