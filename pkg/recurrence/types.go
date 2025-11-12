package recurrence

import "time"

// Frequency represents how often an event recurs
type Frequency string

const (
	FrequencyDaily   Frequency = "DAILY"
	FrequencyWeekly  Frequency = "WEEKLY"
	FrequencyMonthly Frequency = "MONTHLY"
	FrequencyYearly  Frequency = "YEARLY"
)

// Weekday represents days of the week
type Weekday string

const (
	Sunday    Weekday = "SU"
	Monday    Weekday = "MO"
	Tuesday   Weekday = "TU"
	Wednesday Weekday = "WE"
	Thursday  Weekday = "TH"
	Friday    Weekday = "FR"
	Saturday  Weekday = "SA"
)

// MonthlyType specifies how monthly recurrence works
type MonthlyType string

const (
	MonthlyByDay     MonthlyType = "day"      // e.g., "15th of each month"
	MonthlyByWeekday MonthlyType = "weekday"  // e.g., "2nd Tuesday of each month"
)

// RecurrenceInput represents user-friendly recurrence parameters
type RecurrenceInput struct {
	// Required
	Frequency Frequency `json:"frequency"`
	Interval  int       `json:"interval"` // Every N days/weeks/months/years (default: 1)

	// End conditions (one of these should be set)
	Until *time.Time `json:"until,omitempty"` // End date
	Count *int       `json:"count,omitempty"` // Number of occurrences

	// Weekly-specific
	ByWeekday []Weekday `json:"by_weekday,omitempty"` // Days of week (for weekly)

	// Monthly-specific
	MonthlyType   MonthlyType `json:"monthly_type,omitempty"`    // How to repeat monthly
	MonthDayNum   *int        `json:"month_day_num,omitempty"`   // Day of month (1-31) for MonthlyByDay
	MonthWeekNum  *int        `json:"month_week_num,omitempty"`  // Week number (1-5, -1 for last) for MonthlyByWeekday
	MonthWeekday  *Weekday    `json:"month_weekday,omitempty"`   // Day of week for MonthlyByWeekday
}

// EventInstance represents a single occurrence of a recurring event
type EventInstance struct {
	StartTime time.Time
	EndTime   time.Time
}

// GenerationOptions controls how many instances to generate
type GenerationOptions struct {
	StartDate    time.Time // Start generating from this date
	LookAhead    int       // Number of months to look ahead (default: 12)
	MaxInstances int       // Maximum number of instances to generate (default: 365)
}
