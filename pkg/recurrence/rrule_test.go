package recurrence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRRule_Daily(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	count := 10

	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	rruleStr, err := GenerateRRule(input, startDate)
	require.NoError(t, err)
	assert.Contains(t, rruleStr, "FREQ=DAILY")
	assert.Contains(t, rruleStr, "INTERVAL=1")
	assert.Contains(t, rruleStr, "COUNT=10")
}

func TestGenerateRRule_Weekly(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	until := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	input := RecurrenceInput{
		Frequency: FrequencyWeekly,
		Interval:  2, // Every 2 weeks
		Until:     &until,
		ByWeekday: []Weekday{Monday, Wednesday, Friday},
	}

	rruleStr, err := GenerateRRule(input, startDate)
	require.NoError(t, err)
	assert.Contains(t, rruleStr, "FREQ=WEEKLY")
	assert.Contains(t, rruleStr, "INTERVAL=2")
	assert.Contains(t, rruleStr, "BYDAY=MO,WE,FR")
}

func TestGenerateRRule_MonthlyByDay(t *testing.T) {
	startDate := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	count := 12

	dayNum := 15
	input := RecurrenceInput{
		Frequency:   FrequencyMonthly,
		Interval:    1,
		Count:       &count,
		MonthlyType: MonthlyByDay,
		MonthDayNum: &dayNum,
	}

	rruleStr, err := GenerateRRule(input, startDate)
	require.NoError(t, err)
	assert.Contains(t, rruleStr, "FREQ=MONTHLY")
	assert.Contains(t, rruleStr, "BYMONTHDAY=15")
}

func TestGenerateRRule_MonthlyByWeekday(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	count := 12

	weekNum := 2
	weekday := Tuesday
	input := RecurrenceInput{
		Frequency:    FrequencyMonthly,
		Interval:     1,
		Count:        &count,
		MonthlyType:  MonthlyByWeekday,
		MonthWeekNum: &weekNum,
		MonthWeekday: &weekday,
	}

	rruleStr, err := GenerateRRule(input, startDate)
	require.NoError(t, err)
	assert.Contains(t, rruleStr, "FREQ=MONTHLY")
	// Should contain +2TU (second Tuesday)
	assert.Contains(t, rruleStr, "TU")
}

func TestGenerateRRule_Yearly(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	count := 5

	input := RecurrenceInput{
		Frequency: FrequencyYearly,
		Interval:  1,
		Count:     &count,
	}

	rruleStr, err := GenerateRRule(input, startDate)
	require.NoError(t, err)
	assert.Contains(t, rruleStr, "FREQ=YEARLY")
	assert.Contains(t, rruleStr, "COUNT=5")
}

func TestValidateRecurrenceInput_Valid(t *testing.T) {
	count := 10
	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	err := ValidateRecurrenceInput(input)
	assert.NoError(t, err)
}

func TestValidateRecurrenceInput_NoEndCondition(t *testing.T) {
	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
	}

	err := ValidateRecurrenceInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must specify either")
}

func TestValidateRecurrenceInput_BothEndConditions(t *testing.T) {
	count := 10
	until := time.Now().Add(30 * 24 * time.Hour)
	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
		Until:     &until,
	}

	err := ValidateRecurrenceInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both")
}

func TestValidateRecurrenceInput_InvalidFrequency(t *testing.T) {
	count := 10
	input := RecurrenceInput{
		Frequency: Frequency("INVALID"),
		Interval:  1,
		Count:     &count,
	}

	err := ValidateRecurrenceInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid frequency")
}

func TestValidateRecurrenceInput_InvalidInterval(t *testing.T) {
	count := 10
	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  0,
		Count:     &count,
	}

	err := ValidateRecurrenceInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interval must be at least 1")
}

func TestValidateRecurrenceInput_MonthlyMissingParams(t *testing.T) {
	count := 10
	input := RecurrenceInput{
		Frequency:   FrequencyMonthly,
		Interval:    1,
		Count:       &count,
		MonthlyType: MonthlyByDay,
		// Missing MonthDayNum
	}

	err := ValidateRecurrenceInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "month_day_num required")
}

func TestGenerateInstances_Daily(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)
	count := 5

	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	rruleStr, err := GenerateRRule(input, startTime)
	require.NoError(t, err)

	options := GenerationOptions{
		StartDate:    startTime,
		LookAhead:    1,
		MaxInstances: 10,
	}

	instances, err := GenerateInstances(rruleStr, startTime, endTime, options)
	require.NoError(t, err)
	assert.Len(t, instances, 5)

	// Verify first instance
	assert.Equal(t, startTime, instances[0].StartTime)
	assert.Equal(t, endTime, instances[0].EndTime)

	// Verify second instance (next day)
	expectedStart := startTime.AddDate(0, 0, 1)
	expectedEnd := endTime.AddDate(0, 0, 1)
	assert.Equal(t, expectedStart, instances[1].StartTime)
	assert.Equal(t, expectedEnd, instances[1].EndTime)
}

func TestGenerateInstances_Weekly(t *testing.T) {
	// Start on Monday, Jan 6, 2025
	startTime := time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 1, 6, 11, 0, 0, 0, time.UTC)
	count := 8

	input := RecurrenceInput{
		Frequency: FrequencyWeekly,
		Interval:  1,
		Count:     &count,
		ByWeekday: []Weekday{Monday, Wednesday, Friday},
	}

	rruleStr, err := GenerateRRule(input, startTime)
	require.NoError(t, err)

	options := GenerationOptions{
		StartDate:    startTime,
		LookAhead:    2,
		MaxInstances: 10,
	}

	instances, err := GenerateInstances(rruleStr, startTime, endTime, options)
	require.NoError(t, err)

	// Should have 8 instances (count limit)
	assert.Len(t, instances, 8)

	// First should be Monday
	assert.Equal(t, time.Monday, instances[0].StartTime.Weekday())

	// Second should be Wednesday (2 days later)
	assert.Equal(t, time.Wednesday, instances[1].StartTime.Weekday())

	// Third should be Friday (4 days after start)
	assert.Equal(t, time.Friday, instances[2].StartTime.Weekday())
}

func TestPreviewInstances(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)
	count := 30

	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	instances, err := PreviewInstances(input, startTime, endTime, 10)
	require.NoError(t, err)

	// Should be limited to 10 (maxPreview)
	assert.Len(t, instances, 10)

	// Verify they're sequential days
	for i := 0; i < 10; i++ {
		expected := startTime.AddDate(0, 0, i)
		assert.Equal(t, expected.Day(), instances[i].StartTime.Day())
	}
}

func TestGetNextOccurrence(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	count := 10

	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	rruleStr, err := GenerateRRule(input, startTime)
	require.NoError(t, err)

	// Get next occurrence after Jan 3
	after := time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC)
	next, err := GetNextOccurrence(rruleStr, startTime, after)
	require.NoError(t, err)
	require.NotNil(t, next)

	// Should be Jan 4 at 10:00 AM
	assert.Equal(t, 2025, next.Year())
	assert.Equal(t, time.January, next.Month())
	assert.Equal(t, 4, next.Day())
	assert.Equal(t, 10, next.Hour())
}

func TestGetNextOccurrence_NoMore(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	count := 3

	input := RecurrenceInput{
		Frequency: FrequencyDaily,
		Interval:  1,
		Count:     &count,
	}

	rruleStr, err := GenerateRRule(input, startTime)
	require.NoError(t, err)

	// Try to get next occurrence after all instances are done
	after := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	next, err := GetNextOccurrence(rruleStr, startTime, after)
	require.NoError(t, err)
	assert.Nil(t, next) // No more occurrences
}
