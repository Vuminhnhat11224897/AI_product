package weekmanager

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// WeekRange represents a week's date range
type WeekRange struct {
	WeekNumber int
	Label      string
	StartDate  time.Time
	EndDate    time.Time
}

// WeekManager handles automatic week calculation from database
type WeekManager struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewWeekManager(db *sql.DB, logger *logrus.Logger) *WeekManager {
	return &WeekManager{
		db:     db,
		logger: logger,
	}
}

// GetAvailableWeeks gets all distinct weeks from database data
func (wm *WeekManager) GetAvailableWeeks() ([]WeekRange, error) {
	query := `
		SELECT DISTINCT 
			DATE_TRUNC('week', created_at)::date as week_start
		FROM wallet_transactions
		WHERE created_at >= '2025-10-01'
		ORDER BY week_start ASC
	`

	rows, err := wm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query weeks: %w", err)
	}
	defer rows.Close()

	var weeks []WeekRange
	weekNum := 1

	for rows.Next() {
		var weekStart time.Time
		if err := rows.Scan(&weekStart); err != nil {
			return nil, fmt.Errorf("failed to scan week: %w", err)
		}

		// Calculate week end (7 days later)
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Format label
		label := fmt.Sprintf("Tuần %d - Tháng %02d/2025", weekNum, weekStart.Month())

		weeks = append(weeks, WeekRange{
			WeekNumber: weekNum,
			Label:      label,
			StartDate:  weekStart,
			EndDate:    weekEnd,
		})

		weekNum++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating weeks: %w", err)
	}

	wm.logger.Infof("📅 Found %d weeks in database", len(weeks))
	for _, w := range weeks {
		wm.logger.Infof("   %s: %s to %s", w.Label, w.StartDate.Format("2006-01-02"), w.EndDate.Format("2006-01-02"))
	}

	return weeks, nil
}

// GetWeekData returns data for specific week with historical context
func (wm *WeekManager) GetWeekData(currentWeek WeekRange, allWeeks []WeekRange) *WeekData {
	data := &WeekData{
		CurrentWeek: currentWeek,
	}

	// Find index of current week
	currentIdx := -1
	for i, w := range allWeeks {
		if w.WeekNumber == currentWeek.WeekNumber {
			currentIdx = i
			break
		}
	}

	// Get previous weeks if available
	if currentIdx > 0 {
		data.PreviousWeek = &allWeeks[currentIdx-1]
	}
	if currentIdx > 1 {
		data.TwoWeeksAgo = &allWeeks[currentIdx-2]
	}

	return data
}

// WeekData contains current week and historical weeks
type WeekData struct {
	CurrentWeek  WeekRange
	PreviousWeek *WeekRange
	TwoWeeksAgo  *WeekRange
}

// HasHistoricalData checks if there are previous weeks for comparison
func (wd *WeekData) HasHistoricalData() bool {
	return wd.PreviousWeek != nil
}

// HasTwoWeeksHistory checks if there are 2 weeks of history
func (wd *WeekData) HasTwoWeeksHistory() bool {
	return wd.PreviousWeek != nil && wd.TwoWeeksAgo != nil
}

// FormatDateRange formats date range for SQL queries
func (wr *WeekRange) FormatDateRange() (string, string) {
	return wr.StartDate.Format("2006-01-02"), wr.EndDate.Format("2006-01-02")
}
