package recurrence

import (
	"fmt"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// GenerateRRule converts user-friendly input to RRULE string (RFC 5545)
func GenerateRRule(input RecurrenceInput, startDate time.Time) (string, error) {
	if input.Interval < 1 {
		input.Interval = 1
	}

	// Map our frequency to rrule frequency
	var freq rrule.Frequency
	switch input.Frequency {
	case FrequencyDaily:
		freq = rrule.DAILY
	case FrequencyWeekly:
		freq = rrule.WEEKLY
	case FrequencyMonthly:
		freq = rrule.MONTHLY
	case FrequencyYearly:
		freq = rrule.YEARLY
	default:
		return "", fmt.Errorf("invalid frequency: %s", input.Frequency)
	}

	// Build ROption
	rOption := &rrule.ROption{
		Freq:     freq,
		Interval: input.Interval,
		Dtstart:  startDate,
	}

	// Set end condition
	if input.Until != nil {
		rOption.Until = *input.Until
	} else if input.Count != nil {
		rOption.Count = *input.Count
	}

	// Handle weekly recurrence with specific days
	if input.Frequency == FrequencyWeekly && len(input.ByWeekday) > 0 {
		weekdays := make([]rrule.Weekday, len(input.ByWeekday))
		for i, wd := range input.ByWeekday {
			weekdays[i] = mapWeekday(wd)
		}
		rOption.Byweekday = weekdays
	}

	// Handle monthly recurrence
	if input.Frequency == FrequencyMonthly {
		if input.MonthlyType == MonthlyByDay && input.MonthDayNum != nil {
			// e.g., "15th of each month"
			rOption.Bymonthday = []int{*input.MonthDayNum}
		} else if input.MonthlyType == MonthlyByWeekday && input.MonthWeekNum != nil && input.MonthWeekday != nil {
			// e.g., "2nd Tuesday of each month"
			wd := mapWeekday(*input.MonthWeekday)
			// In rrule-go, we use Byweekday with offset
			// +1TU = first Tuesday, +2TU = second Tuesday, -1TU = last Tuesday
			rOption.Byweekday = []rrule.Weekday{wd.Nth(*input.MonthWeekNum)}
		}
	}

	// Generate RRULE
	rule, err := rrule.NewRRule(*rOption)
	if err != nil {
		return "", fmt.Errorf("failed to create rrule: %w", err)
	}

	return rule.OrigOptions.RRuleString(), nil
}

// ParseRRule parses an RRULE string and validates it
func ParseRRule(rruleStr string, startDate time.Time) (*rrule.RRule, error) {
	// Prepend DTSTART if not present
	if !strings.Contains(rruleStr, "DTSTART") {
		rruleStr = fmt.Sprintf("DTSTART:%s\n%s", startDate.Format("20060102T150405Z"), rruleStr)
	}

	rule, err := rrule.StrToRRule(rruleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rrule: %w", err)
	}

	return rule, nil
}

// ValidateRecurrenceInput validates user input
func ValidateRecurrenceInput(input RecurrenceInput) error {
	if input.Interval < 1 {
		return fmt.Errorf("interval must be at least 1")
	}

	// Must have at least one end condition
	if input.Until == nil && input.Count == nil {
		return fmt.Errorf("must specify either 'until' date or 'count'")
	}

	// Can't have both end conditions
	if input.Until != nil && input.Count != nil {
		return fmt.Errorf("cannot specify both 'until' and 'count'")
	}

	// Validate frequency
	validFreqs := []Frequency{FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly}
	valid := false
	for _, f := range validFreqs {
		if input.Frequency == f {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid frequency: %s", input.Frequency)
	}

	// Validate weekly parameters
	if input.Frequency == FrequencyWeekly && len(input.ByWeekday) > 0 {
		for _, wd := range input.ByWeekday {
			if !isValidWeekday(wd) {
				return fmt.Errorf("invalid weekday: %s", wd)
			}
		}
	}

	// Validate monthly parameters
	if input.Frequency == FrequencyMonthly {
		if input.MonthlyType == MonthlyByDay {
			if input.MonthDayNum == nil {
				return fmt.Errorf("month_day_num required for monthly by day")
			}
			if *input.MonthDayNum < 1 || *input.MonthDayNum > 31 {
				return fmt.Errorf("month_day_num must be between 1 and 31")
			}
		} else if input.MonthlyType == MonthlyByWeekday {
			if input.MonthWeekNum == nil || input.MonthWeekday == nil {
				return fmt.Errorf("month_week_num and month_weekday required for monthly by weekday")
			}
			if (*input.MonthWeekNum < 1 || *input.MonthWeekNum > 5) && *input.MonthWeekNum != -1 {
				return fmt.Errorf("month_week_num must be 1-5 or -1 (for last)")
			}
			if !isValidWeekday(*input.MonthWeekday) {
				return fmt.Errorf("invalid month_weekday: %s", *input.MonthWeekday)
			}
		}
	}

	return nil
}

// Helper functions

func mapWeekday(wd Weekday) rrule.Weekday {
	switch wd {
	case Sunday:
		return rrule.SU
	case Monday:
		return rrule.MO
	case Tuesday:
		return rrule.TU
	case Wednesday:
		return rrule.WE
	case Thursday:
		return rrule.TH
	case Friday:
		return rrule.FR
	case Saturday:
		return rrule.SA
	default:
		return rrule.MO // Default
	}
}

func isValidWeekday(wd Weekday) bool {
	validWeekdays := []Weekday{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
	for _, vwd := range validWeekdays {
		if wd == vwd {
			return true
		}
	}
	return false
}
