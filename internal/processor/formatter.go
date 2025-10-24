package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// TableFormatter formats processing results into a readable table
type TableFormatter struct {
	logger     *logrus.Logger
	tableWidth int
}

// ResultSummary contains aggregated statistics
type ResultSummary struct {
	TotalItems       int
	SuccessCount     int
	FailureCount     int
	TotalRetries     int
	TotalTokens      int
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	SuccessRate      float64
	AvgTokensPerItem int
}

// NewTableFormatter creates a new table formatter instance
func NewTableFormatter(logger *logrus.Logger, tableWidth int) *TableFormatter {
	if tableWidth == 0 {
		tableWidth = 150
	}
	return &TableFormatter{
		logger:     logger,
		tableWidth: tableWidth,
	}
}

// FormatResultsTable formats and displays the processing results as a table
func (tf *TableFormatter) FormatResultsTable(results []ProcessResult) {
	if len(results) == 0 {
		tf.logger.Warn("No results to format")
		return
	}

	// Calculate summary statistics
	summary := tf.calculateSummary(results)

	// Display header
	tf.printSeparator("=")
	tf.printCentered("AI PROCESSING RESULTS SUMMARY")
	tf.printSeparator("=")
	tf.logger.Info("")

	// Display summary statistics
	tf.displaySummaryStats(summary)

	// Display detailed results table
	tf.logger.Info("")
	tf.printSeparator("-")
	tf.displayDetailedResults(results)

	// Display footer
	tf.printSeparator("=")
}

// calculateSummary aggregates statistics from all results
func (tf *TableFormatter) calculateSummary(results []ProcessResult) ResultSummary {
	summary := ResultSummary{
		TotalItems: len(results),
	}

	for _, result := range results {
		if result.Success {
			summary.SuccessCount++
			summary.TotalTokens += result.TokenUsage.TotalTokens
		} else {
			summary.FailureCount++
		}
		summary.TotalRetries += result.Retries
		summary.TotalDuration += result.Duration
	}

	if summary.TotalItems > 0 {
		summary.AverageDuration = summary.TotalDuration / time.Duration(summary.TotalItems)
		summary.SuccessRate = float64(summary.SuccessCount) / float64(summary.TotalItems) * 100
	}

	if summary.SuccessCount > 0 {
		summary.AvgTokensPerItem = summary.TotalTokens / summary.SuccessCount
	}

	return summary
}

// displaySummaryStats displays the summary statistics
func (tf *TableFormatter) displaySummaryStats(summary ResultSummary) {
	// Processing statistics
	tf.logger.WithFields(logrus.Fields{
		"total_items":  summary.TotalItems,
		"successful":   summary.SuccessCount,
		"failed":       summary.FailureCount,
		"success_rate": fmt.Sprintf("%.2f%%", summary.SuccessRate),
	}).Info("📊 Processing Statistics")

	// Performance metrics
	tf.logger.WithFields(logrus.Fields{
		"total_duration":   summary.TotalDuration,
		"average_per_item": summary.AverageDuration,
		"total_retries":    summary.TotalRetries,
	}).Info("⚡ Performance Metrics")

	// Token usage
	if summary.TotalTokens > 0 {
		tf.logger.WithFields(logrus.Fields{
			"total_tokens":        summary.TotalTokens,
			"avg_tokens_per_item": summary.AvgTokensPerItem,
		}).Info("🎯 Token Usage")
	}
}

// displayDetailedResults displays a detailed table of individual results
func (tf *TableFormatter) displayDetailedResults(results []ProcessResult) {
	tf.printCentered("DETAILED RESULTS")
	tf.printSeparator("-")

	// Table header
	header := fmt.Sprintf("%-6s | %-10s | %-8s | %-10s | %-10s | %-30s",
		"Index", "Status", "Retries", "Duration", "Tokens", "Error")
	tf.logger.Info(header)
	tf.printSeparator("-")

	// Table rows
	for _, result := range results {
		status := "✅ SUCCESS"
		errorMsg := "-"
		tokens := fmt.Sprintf("%d", result.TokenUsage.TotalTokens)

		if !result.Success {
			status = "❌ FAILED"
			if result.Error != nil {
				errorMsg = result.Error.Error()
				if len(errorMsg) > 28 {
					errorMsg = errorMsg[:28] + "..."
				}
			}
			tokens = "-"
		}

		row := fmt.Sprintf("%-6d | %-10s | %-8d | %-10s | %-10s | %-30s",
			result.Index,
			status,
			result.Retries,
			result.Duration.Round(time.Millisecond),
			tokens,
			errorMsg,
		)
		tf.logger.Info(row)
	}
}

// printSeparator prints a separator line
func (tf *TableFormatter) printSeparator(char string) {
	tf.logger.Info(strings.Repeat(char, tf.tableWidth))
}

// printCentered prints centered text
func (tf *TableFormatter) printCentered(text string) {
	padding := (tf.tableWidth - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	centered := strings.Repeat(" ", padding) + text
	tf.logger.Info(centered)
}

// FormatFinalSummary formats the final summary box
func (tf *TableFormatter) FormatFinalSummary(total, success, failed int) {
	successRate := float64(success) / float64(total) * 100

	tf.logger.Info("")
	tf.logger.Info("╔" + strings.Repeat("═", 78) + "╗")
	tf.logger.Info(fmt.Sprintf("║%s║", tf.centerText("FINAL PROCESSING SUMMARY", 78)))
	tf.logger.Info("╠" + strings.Repeat("═", 78) + "╣")
	tf.logger.Info(fmt.Sprintf("║  Total: %-3d  |  Success: %-3d  |  Failed: %-3d  |  Success Rate: %6.2f%%  ║",
		total, success, failed, successRate))
	tf.logger.Info("╚" + strings.Repeat("═", 78) + "╝")
	tf.logger.Info("")
}

// centerText centers text within a given width
func (tf *TableFormatter) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}
