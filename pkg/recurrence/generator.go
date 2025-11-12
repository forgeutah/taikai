package recurrence

import (
	"fmt"
	"time"
)

// GenerateInstances generates event instances from an RRULE string
func GenerateInstances(rruleStr string, startTime, endTime time.Time, options GenerationOptions) ([]EventInstance, error) {
	// Set defaults
	if options.LookAhead == 0 {
		options.LookAhead = 12 // 12 months ahead
	}
	if options.MaxInstances == 0 {
		options.MaxInstances = 365 // Max 365 instances
	}
	if options.StartDate.IsZero() {
		options.StartDate = time.Now()
	}

	// Parse RRULE
	rule, err := ParseRRule(rruleStr, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rrule: %w", err)
	}

	// Calculate look-ahead end date
	lookAheadEnd := options.StartDate.AddDate(0, options.LookAhead, 0)

	// Generate all occurrences up to look-ahead date
	occurrences := rule.Between(
		options.StartDate,
		lookAheadEnd,
		true, // inclusive
	)

	// Limit to max instances
	if len(occurrences) > options.MaxInstances {
		occurrences = occurrences[:options.MaxInstances]
	}

	// Calculate event duration
	duration := endTime.Sub(startTime)

	// Convert to EventInstance array
	instances := make([]EventInstance, len(occurrences))
	for i, occurrence := range occurrences {
		instances[i] = EventInstance{
			StartTime: occurrence,
			EndTime:   occurrence.Add(duration),
		}
	}

	return instances, nil
}

// PreviewInstances generates a preview of recurring event dates
// This is useful for the preview endpoint before creating the event
func PreviewInstances(input RecurrenceInput, startTime, endTime time.Time, maxPreview int) ([]EventInstance, error) {
	// Validate input
	if err := ValidateRecurrenceInput(input); err != nil {
		return nil, err
	}

	// Generate RRULE
	rruleStr, err := GenerateRRule(input, startTime)
	if err != nil {
		return nil, err
	}

	// Limit preview to reasonable number
	if maxPreview == 0 {
		maxPreview = 50 // Show first 50 occurrences in preview
	}

	// Generate instances
	options := GenerationOptions{
		StartDate:    startTime,
		LookAhead:    12, // Preview 12 months ahead
		MaxInstances: maxPreview,
	}

	instances, err := GenerateInstances(rruleStr, startTime, endTime, options)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// GetNextOccurrence returns the next occurrence of a recurring event after a given date
func GetNextOccurrence(rruleStr string, startTime time.Time, after time.Time) (*time.Time, error) {
	rule, err := ParseRRule(rruleStr, startTime)
	if err != nil {
		return nil, err
	}

	// Get next occurrence after the given date
	next := rule.After(after, false) // false = not inclusive

	if next.IsZero() {
		return nil, nil // No more occurrences
	}

	return &next, nil
}

// CountTotalOccurrences returns the total number of occurrences for a recurring event
func CountTotalOccurrences(rruleStr string, startTime time.Time) (int, error) {
	rule, err := ParseRRule(rruleStr, startTime)
	if err != nil {
		return 0, err
	}

	// Get all occurrences (with reasonable upper limit)
	maxDate := startTime.AddDate(10, 0, 0) // Look up to 10 years ahead
	occurrences := rule.Between(startTime, maxDate, true)

	return len(occurrences), nil
}
