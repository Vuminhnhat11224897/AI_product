package silver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"ai-production-pipeline/internal/weekmanager"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// SilverLayer handles enhanced transformation with historical comparison
type SilverLayer struct {
	db     *sql.DB
	logger *logrus.Logger
}

// EnhancedKidData represents complete kid analysis with historical context
type EnhancedKidData struct {
	ProfileID   string `json:"profile_id"`
	Nickname    string `json:"nickname"`
	Age         int    `json:"age"`
	DateOfBirth string `json:"date_of_birth"`

	// Multi-week data
	CurrentWeek  WeekMetrics  `json:"current_week"`
	PreviousWeek *WeekMetrics `json:"previous_week,omitempty"`
	TwoWeeksAgo  *WeekMetrics `json:"two_weeks_ago,omitempty"`

	// Analysis (only if historical data available)
	Trends     *TrendData      `json:"trends,omitempty"`
	Statistics *StatisticsData `json:"statistics,omitempty"`

	// Scores
	ActivityScore    float64 `json:"activity_score"`
	ConsistencyScore float64 `json:"consistency_score,omitempty"`
	ImprovementRate  float64 `json:"improvement_rate,omitempty"`
}

// WeekMetrics represents data for one week
type WeekMetrics struct {
	WeekLabel string `json:"week_label"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	// Wallet balances
	JoyWallet      float64 `json:"joy_wallet"`
	SpendingWallet float64 `json:"spending_wallet"`
	CharityWallet  float64 `json:"charity_wallet"`
	StudyWallet    float64 `json:"study_wallet"`
	TotalBalance   float64 `json:"total_balance"`

	// Transaction summary
	MoneyReceived      float64 `json:"money_received"`
	MoneyReceivedCount int     `json:"money_received_count"`
	TotalSpent         float64 `json:"total_spent"`
	JoySpent           float64 `json:"joy_spent"`
	SpendingSpent      float64 `json:"spending_spent"`
	CharitySpent       float64 `json:"charity_spent"`
	StudySpent         float64 `json:"study_spent"`
	SpentCount         int     `json:"spent_count"`

	// Mission data
	MissionsTotal     int     `json:"missions_total"`
	MissionsCompleted int     `json:"missions_completed"`
	MissionsPending   int     `json:"missions_pending"`
	CompletionRate    float64 `json:"completion_rate"`

	// Activity
	TransactionCount   int     `json:"transaction_count"`
	AvgTransactionSize float64 `json:"avg_transaction_size"`
	ActiveDays         int     `json:"active_days"`
}

// TrendData represents trends across weeks
type TrendData struct {
	BalanceTrend         string  `json:"balance_trend"` // increasing, decreasing, stable
	BalanceChangePercent float64 `json:"balance_change_percent"`

	SpendingTrend         string  `json:"spending_trend"`
	SpendingChangePercent float64 `json:"spending_change_percent"`

	MissionCompletionTrend string  `json:"mission_completion_trend"`
	CompletionRateChange   float64 `json:"completion_rate_change"`

	ActivityTrend  string `json:"activity_trend"`
	ActivityChange int    `json:"activity_change"`

	ConsistencyLevel string `json:"consistency_level"` // high, medium, low
}

// StatisticsData represents calculated statistics
type StatisticsData struct {
	// Spending patterns (current week)
	JoySpendingRatio float64 `json:"joy_spending_ratio"`
	SavingsRatio     float64 `json:"savings_ratio"` // (spending_wallet + study_wallet) / total
	CharityRatio     float64 `json:"charity_ratio"`
	StudyRatio       float64 `json:"study_ratio"`

	// Averages (across all available weeks)
	AvgWeeklyIncome      float64 `json:"avg_weekly_income"`
	AvgWeeklySpending    float64 `json:"avg_weekly_spending"`
	AvgMissionCompletion float64 `json:"avg_mission_completion"`

	// Growth rates
	IncomeGrowthRate  float64 `json:"income_growth_rate"` // % change
	SavingsGrowthRate float64 `json:"savings_growth_rate"`

	// Behavioral patterns
	SpendingConsistency float64 `json:"spending_consistency"` // 0-1
	SavingsBehavior     string  `json:"savings_behavior"`     // aggressive, moderate, minimal
	CharityInvolvement  string  `json:"charity_involvement"`  // high, medium, low
}

// EnhancedOutput represents the final JSON output
type EnhancedOutput struct {
	GeneratedAt string            `json:"generated_at"`
	Week        string            `json:"week"`
	TotalKids   int               `json:"total_kids"`
	Kids        []EnhancedKidData `json:"kids"`
}

func NewSilverLayer(db *sql.DB, logger *logrus.Logger) *SilverLayer {
	return &SilverLayer{
		db:     db,
		logger: logger,
	}
}

// Transform performs enhanced transformation for a specific week
func (s *SilverLayer) Transform(weekData *weekmanager.WeekData, outputPath string) error {
	s.logger.Info("=" + repeatString("=", 80))
	s.logger.Infof("🔄 Silver Layer V3: Processing %s", weekData.CurrentWeek.Label)
	s.logger.Info("=" + repeatString("=", 80))

	if weekData.HasHistoricalData() {
		s.logger.Infof("📊 Historical data available: %d previous weeks",
			func() int {
				if weekData.HasTwoWeeksHistory() {
					return 2
				} else {
					return 1
				}
			}())
	} else {
		s.logger.Warn("⚠️  First week - no historical comparison available")
	}

	// Get ALL kid profiles (not filtered by activity)
	profiles, err := s.getAllKidProfiles()
	if err != nil {
		return fmt.Errorf("failed to get kid profiles: %w", err)
	}

	s.logger.Infof("👥 Processing %d kids (including inactive)", len(profiles))

	// Analyze each kid
	var kidsData []EnhancedKidData
	activeCount := 0
	inactiveCount := 0

	for _, profile := range profiles {
		s.logger.Infof("   Analyzing: %s (ID: %s)", profile.Nickname, profile.ProfileID)

		kidData, err := s.analyzeKidEnhanced(profile, weekData)
		if err != nil {
			s.logger.Errorf("   ❌ Error analyzing %s: %v", profile.Nickname, err)
			continue
		}

		// Include ALL kids regardless of activity
		kidsData = append(kidsData, *kidData)

		if kidData.CurrentWeek.TransactionCount > 0 || kidData.CurrentWeek.MissionsCompleted > 0 {
			activeCount++
			s.logger.Infof("   ✅ Active: Activity Score %.2f, Trends: %v",
				kidData.ActivityScore, kidData.Trends != nil)
		} else {
			inactiveCount++
			s.logger.Infof("   ⚪ Inactive: No activity this week (Trends: %v)",
				kidData.Trends != nil)
		}
	}

	s.logger.Infof("📊 Summary: %d active, %d inactive, %d total",
		activeCount, inactiveCount, len(kidsData)) // Create output
	output := EnhancedOutput{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Week:        weekData.CurrentWeek.Label,
		TotalKids:   len(kidsData),
		Kids:        kidsData,
	}

	// Save to JSON
	if err := s.saveJSON(output, outputPath); err != nil {
		return fmt.Errorf("failed to save JSON: %w", err)
	}

	s.logger.Infof("✅ Silver Layer V3 Complete: %s", outputPath)
	return nil
}

// analyzeKidEnhanced performs complete analysis with historical comparison
func (s *SilverLayer) analyzeKidEnhanced(profile KidProfile, weekData *weekmanager.WeekData) (*EnhancedKidData, error) {
	data := &EnhancedKidData{
		ProfileID:   profile.ProfileID,
		Nickname:    profile.Nickname,
		Age:         profile.Age,
		DateOfBirth: profile.DateOfBirth,
	}

	// Get current week metrics
	currentMetrics, err := s.getWeekMetrics(profile.ProfileID, &weekData.CurrentWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get current week metrics: %w", err)
	}
	data.CurrentWeek = *currentMetrics

	// Get historical metrics if available
	if weekData.HasHistoricalData() {
		prevMetrics, err := s.getWeekMetrics(profile.ProfileID, weekData.PreviousWeek)
		if err == nil {
			data.PreviousWeek = prevMetrics
		}

		if weekData.HasTwoWeeksHistory() {
			twoWeeksMetrics, err := s.getWeekMetrics(profile.ProfileID, weekData.TwoWeeksAgo)
			if err == nil {
				data.TwoWeeksAgo = twoWeeksMetrics
			}
		}
	}

	// Calculate activity score
	data.ActivityScore = s.calculateActivityScore(currentMetrics)

	// Calculate trends and statistics if historical data available
	if data.PreviousWeek != nil {
		s.logger.Debugf("      📈 Calculating trends for %s (has previous week)", profile.Nickname)
		data.Trends = s.calculateTrends(data)
		data.Statistics = s.calculateStatistics(data)
		data.ConsistencyScore = s.calculateConsistencyScore(data)
		data.ImprovementRate = s.calculateImprovementRate(data)
		s.logger.Debugf("      ✅ Trends calculated: Balance=%s, Spending=%s",
			data.Trends.BalanceTrend, data.Trends.SpendingTrend)
	} else {
		s.logger.Debugf("      ⏭️  No previous week data for %s - skipping trends", profile.Nickname)
	}

	return data, nil
}

// getWeekMetrics gets all metrics for a kid in a specific week
func (s *SilverLayer) getWeekMetrics(profileID string, week *weekmanager.WeekRange) (*WeekMetrics, error) {
	startDate, endDate := week.FormatDateRange()

	metrics := &WeekMetrics{
		WeekLabel: week.Label,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Get wallet balances (current state, not time-ranged)
	walletQuery := `
		SELECT slug, balance
		FROM wallets
		WHERE profile_id = $1::uuid
	`
	rows, err := s.db.Query(walletQuery, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalBalance := 0.0
	for rows.Next() {
		var walletType string
		var balance float64
		if err := rows.Scan(&walletType, &balance); err != nil {
			return nil, err
		}

		totalBalance += balance
		switch walletType {
		case "joy":
			metrics.JoyWallet = balance
		case "spending":
			metrics.SpendingWallet = balance
		case "charity":
			metrics.CharityWallet = balance
		case "study":
			metrics.StudyWallet = balance
		}
	}
	metrics.TotalBalance = totalBalance

	// Get transaction data for this week
	txQuery := `
		SELECT 
			w.slug,
			wt.type,
			SUM(wt.amount) as total,
			COUNT(*) as count
		FROM wallet_transactions wt
		JOIN wallets w ON wt.wallet_id = w.id
		WHERE wt.profile_id = $1::uuid
		  AND wt.created_at >= $2::date
		  AND wt.created_at < $3::date
		GROUP BY w.slug, wt.type
	`
	txRows, err := s.db.Query(txQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer txRows.Close()

	for txRows.Next() {
		var walletType, txType string
		var amount float64
		var count int
		if err := txRows.Scan(&walletType, &txType, &amount, &count); err != nil {
			return nil, err
		}

		if txType == "deposit" {
			metrics.MoneyReceived += amount
			metrics.MoneyReceivedCount += count
		} else if txType == "withdraw" {
			metrics.TotalSpent += amount
			metrics.SpentCount += count

			switch walletType {
			case "joy":
				metrics.JoySpent += amount
			case "spending":
				metrics.SpendingSpent += amount
			case "charity":
				metrics.CharitySpent += amount
			case "study":
				metrics.StudySpent += amount
			}
		}
	}

	metrics.TransactionCount = metrics.MoneyReceivedCount + metrics.SpentCount
	if metrics.TransactionCount > 0 {
		metrics.AvgTransactionSize = (metrics.MoneyReceived + metrics.TotalSpent) / float64(metrics.TransactionCount)
	}

	// Get mission data
	missionQuery := `
		SELECT 
			COALESCE(COUNT(*), 0) as total,
			COALESCE(SUM(CASE WHEN status = 'complete' THEN 1 ELSE 0 END), 0) as completed
		FROM missions
		WHERE profile_id = $1::uuid
		  AND created_at >= $2::date
		  AND created_at < $3::date
	`
	var completed sql.NullInt64
	err = s.db.QueryRow(missionQuery, profileID, startDate, endDate).Scan(
		&metrics.MissionsTotal,
		&completed,
	)
	if completed.Valid {
		metrics.MissionsCompleted = int(completed.Int64)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	metrics.MissionsPending = metrics.MissionsTotal - metrics.MissionsCompleted
	if metrics.MissionsTotal > 0 {
		metrics.CompletionRate = float64(metrics.MissionsCompleted) / float64(metrics.MissionsTotal) * 100
	}

	// Get active days
	activeDaysQuery := `
		SELECT COUNT(DISTINCT DATE(created_at))
		FROM wallet_transactions
		WHERE profile_id = $1::uuid
		  AND created_at >= $2::date
		  AND created_at < $3::date
	`
	if err := s.db.QueryRow(activeDaysQuery, profileID, startDate, endDate).Scan(&metrics.ActiveDays); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	return metrics, nil
}

// calculateTrends calculates trends by comparing weeks
func (s *SilverLayer) calculateTrends(data *EnhancedKidData) *TrendData {
	trends := &TrendData{}

	current := &data.CurrentWeek
	previous := data.PreviousWeek

	if previous == nil {
		return trends
	}

	// Balance trend
	if previous.TotalBalance > 0 {
		balanceChange := current.TotalBalance - previous.TotalBalance
		trends.BalanceChangePercent = (balanceChange / previous.TotalBalance) * 100

		if trends.BalanceChangePercent > 10 {
			trends.BalanceTrend = "strongly_increasing"
		} else if trends.BalanceChangePercent > 0 {
			trends.BalanceTrend = "increasing"
		} else if trends.BalanceChangePercent < -10 {
			trends.BalanceTrend = "strongly_decreasing"
		} else if trends.BalanceChangePercent < 0 {
			trends.BalanceTrend = "decreasing"
		} else {
			trends.BalanceTrend = "stable"
		}
	}

	// Spending trend
	if previous.TotalSpent > 0 {
		spendingChange := current.TotalSpent - previous.TotalSpent
		trends.SpendingChangePercent = (spendingChange / previous.TotalSpent) * 100

		if math.Abs(trends.SpendingChangePercent) < 10 {
			trends.SpendingTrend = "stable"
		} else if trends.SpendingChangePercent > 0 {
			trends.SpendingTrend = "increasing"
		} else {
			trends.SpendingTrend = "decreasing"
		}
	}

	// Mission completion trend
	completionChange := current.CompletionRate - previous.CompletionRate
	trends.CompletionRateChange = completionChange

	if math.Abs(completionChange) < 5 {
		trends.MissionCompletionTrend = "stable"
	} else if completionChange > 0 {
		trends.MissionCompletionTrend = "improving"
	} else {
		trends.MissionCompletionTrend = "declining"
	}

	// Activity trend
	activityChange := current.TransactionCount - previous.TransactionCount
	trends.ActivityChange = activityChange

	if activityChange > 2 {
		trends.ActivityTrend = "increasing"
	} else if activityChange < -2 {
		trends.ActivityTrend = "decreasing"
	} else {
		trends.ActivityTrend = "stable"
	}

	// Consistency level (using coefficient of variation)
	weeks := []float64{current.TotalSpent}
	if previous != nil {
		weeks = append(weeks, previous.TotalSpent)
	}
	if data.TwoWeeksAgo != nil {
		weeks = append(weeks, data.TwoWeeksAgo.TotalSpent)
	}

	if len(weeks) >= 2 {
		stdDev := calculateStdDev(weeks)
		mean := calculateMean(weeks)
		if mean > 0 {
			cv := stdDev / mean
			if cv < 0.2 {
				trends.ConsistencyLevel = "high"
			} else if cv < 0.5 {
				trends.ConsistencyLevel = "medium"
			} else {
				trends.ConsistencyLevel = "low"
			}
		}
	}

	return trends
}

// calculateStatistics calculates aggregate statistics
func (s *SilverLayer) calculateStatistics(data *EnhancedKidData) *StatisticsData {
	stats := &StatisticsData{}
	current := &data.CurrentWeek

	// Spending ratios (current week)
	if current.TotalSpent > 0 {
		stats.JoySpendingRatio = current.JoySpent / current.TotalSpent
		stats.CharityRatio = current.CharitySpent / current.TotalSpent
		stats.StudyRatio = current.StudySpent / current.TotalSpent
	}

	// Savings ratio (savings wallets / total balance)
	if current.TotalBalance > 0 {
		savingsAmount := current.SpendingWallet + current.StudyWallet
		stats.SavingsRatio = savingsAmount / current.TotalBalance
	}

	// Averages across all available weeks
	incomes := []float64{current.MoneyReceived}
	spendings := []float64{current.TotalSpent}
	completions := []float64{current.CompletionRate}

	if data.PreviousWeek != nil {
		incomes = append(incomes, data.PreviousWeek.MoneyReceived)
		spendings = append(spendings, data.PreviousWeek.TotalSpent)
		completions = append(completions, data.PreviousWeek.CompletionRate)
	}
	if data.TwoWeeksAgo != nil {
		incomes = append(incomes, data.TwoWeeksAgo.MoneyReceived)
		spendings = append(spendings, data.TwoWeeksAgo.TotalSpent)
		completions = append(completions, data.TwoWeeksAgo.CompletionRate)
	}

	stats.AvgWeeklyIncome = calculateMean(incomes)
	stats.AvgWeeklySpending = calculateMean(spendings)
	stats.AvgMissionCompletion = calculateMean(completions)

	// Growth rates (if at least 2 weeks)
	if len(incomes) >= 2 {
		oldestIncome := incomes[len(incomes)-1]
		if oldestIncome > 0 {
			stats.IncomeGrowthRate = ((current.MoneyReceived - oldestIncome) / oldestIncome) * 100
		}

		savingsCurrent := current.SpendingWallet + current.StudyWallet
		if data.TwoWeeksAgo != nil {
			savingsOldest := data.TwoWeeksAgo.SpendingWallet + data.TwoWeeksAgo.StudyWallet
			if savingsOldest > 0 {
				stats.SavingsGrowthRate = ((savingsCurrent - savingsOldest) / savingsOldest) * 100
			}
		} else if data.PreviousWeek != nil {
			savingsOldest := data.PreviousWeek.SpendingWallet + data.PreviousWeek.StudyWallet
			if savingsOldest > 0 {
				stats.SavingsGrowthRate = ((savingsCurrent - savingsOldest) / savingsOldest) * 100
			}
		}
	}

	// Spending consistency
	if len(spendings) >= 2 {
		stdDev := calculateStdDev(spendings)
		mean := calculateMean(spendings)
		if mean > 0 {
			cv := stdDev / mean
			stats.SpendingConsistency = 1.0 - math.Min(cv, 1.0) // 0-1 scale, higher is more consistent
		}
	}

	// Savings behavior
	if stats.SavingsRatio >= 0.5 {
		stats.SavingsBehavior = "aggressive"
	} else if stats.SavingsRatio >= 0.3 {
		stats.SavingsBehavior = "moderate"
	} else {
		stats.SavingsBehavior = "minimal"
	}

	// Charity involvement
	if stats.CharityRatio >= 0.15 {
		stats.CharityInvolvement = "high"
	} else if stats.CharityRatio >= 0.05 {
		stats.CharityInvolvement = "medium"
	} else {
		stats.CharityInvolvement = "low"
	}

	return stats
}

// calculateActivityScore calculates activity score for a week
func (s *SilverLayer) calculateActivityScore(metrics *WeekMetrics) float64 {
	score := 0.0

	// Transaction activity (max 40 points)
	score += math.Min(float64(metrics.TransactionCount)*4, 40)

	// Mission completion (max 30 points)
	score += (metrics.CompletionRate / 100) * 30

	// Active days (max 20 points)
	score += math.Min(float64(metrics.ActiveDays)*2.86, 20) // 7 days = 20 points

	// Balance management (max 10 points)
	if metrics.TotalBalance > 0 {
		score += 10
	}

	return math.Min(score, 100)
}

// calculateConsistencyScore calculates consistency across weeks
func (s *SilverLayer) calculateConsistencyScore(data *EnhancedKidData) float64 {
	values := []float64{}

	if data.TwoWeeksAgo != nil {
		values = append(values, data.TwoWeeksAgo.TotalSpent)
	}
	if data.PreviousWeek != nil {
		values = append(values, data.PreviousWeek.TotalSpent)
	}
	values = append(values, data.CurrentWeek.TotalSpent)

	if len(values) < 2 {
		return 0
	}

	stdDev := calculateStdDev(values)
	mean := calculateMean(values)

	if mean == 0 {
		return 0
	}

	cv := stdDev / mean
	// Convert to 0-1 scale (lower CV = higher consistency)
	return math.Max(0, 1.0-cv)
}

// calculateImprovementRate calculates overall improvement rate
func (s *SilverLayer) calculateImprovementRate(data *EnhancedKidData) float64 {
	if data.Trends == nil || data.Statistics == nil {
		return 0
	}

	improvements := 0.0
	count := 0.0

	// Balance improvement
	if data.Trends.BalanceChangePercent > 0 {
		improvements += math.Min(data.Trends.BalanceChangePercent/100, 1.0)
	}
	count++

	// Mission completion improvement
	if data.Trends.CompletionRateChange > 0 {
		improvements += math.Min(data.Trends.CompletionRateChange/100, 1.0)
	}
	count++

	// Savings growth
	if data.Statistics.SavingsGrowthRate > 0 {
		improvements += math.Min(data.Statistics.SavingsGrowthRate/100, 1.0)
	}
	count++

	if count == 0 {
		return 0
	}

	return (improvements / count) * 100
}

// Helper: getKidProfiles gets all kid profiles
// getAllKidProfiles returns ALL kids in the system (used for comprehensive weekly analysis)
func (s *SilverLayer) getAllKidProfiles() ([]KidProfile, error) {
	query := `
		SELECT 
			id::text,
			COALESCE(full_name, 'Unknown'),
			COALESCE(full_name, 'Kid'),
			COALESCE(EXTRACT(YEAR FROM AGE(CURRENT_DATE, date_of_birth)), 0)::int,
			COALESCE(date_of_birth::text, '')
		FROM profiles
		WHERE profile_type = 'kid'
		ORDER BY created_at
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []KidProfile
	for rows.Next() {
		var p KidProfile
		if err := rows.Scan(&p.ProfileID, &p.FullName, &p.Nickname, &p.Age, &p.DateOfBirth); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}

// getActiveKidProfiles returns kids who had transactions or missions in the given week
// NOTE: Currently not used - kept for potential future filtering needs
func (s *SilverLayer) getActiveKidProfiles(week *weekmanager.WeekRange) ([]KidProfile, error) {
	startDate, endDate := week.FormatDateRange()

	query := `
		SELECT DISTINCT
			p.id::text,
			COALESCE(p.full_name, 'Unknown'),
			COALESCE(p.full_name, 'Kid'),
			COALESCE(EXTRACT(YEAR FROM AGE(CURRENT_DATE, p.date_of_birth)), 0)::int,
			COALESCE(p.date_of_birth::text, ''),
			p.created_at
		FROM profiles p
		WHERE p.profile_type = 'kid'
		AND (
			-- Has transactions in this week
			EXISTS (
				SELECT 1 FROM wallet_transactions wt
				WHERE wt.profile_id = p.id
				AND wt.created_at >= $1::timestamp
				AND wt.created_at < $2::timestamp
			)
			OR
			-- Has missions in this week
			EXISTS (
				SELECT 1 FROM missions m
				WHERE m.profile_id = p.id
				AND m.created_at >= $1::timestamp
				AND m.created_at < $2::timestamp
			)
		)
		ORDER BY p.created_at
	`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []KidProfile
	for rows.Next() {
		var p KidProfile
		var createdAt interface{} // Ignore this field, only used for ORDER BY
		if err := rows.Scan(&p.ProfileID, &p.FullName, &p.Nickname, &p.Age, &p.DateOfBirth, &createdAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}

// saveJSON saves data to JSON file
func (s *SilverLayer) saveJSON(data interface{}, filepath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Helper functions
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateMean(values)
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	return math.Sqrt(variance / float64(len(values)))
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
