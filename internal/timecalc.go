package internal

import (
	"fmt"
	"time"
)

var JST *time.Location

func init() {
	var err error
	JST, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// This is a fatal error for the application, so panic.
		panic(fmt.Sprintf("Error loading location 'Asia/Tokyo': %v", err))
	}
}

// DayOfWeekMap maps Japanese day of the week to time.Weekday
var DayOfWeekMap = map[string]time.Weekday{
	"日": time.Sunday,
	"月": time.Monday,
	"火": time.Tuesday,
	"水": time.Wednesday,
	"木": time.Thursday,
	"金": time.Friday,
	"土": time.Saturday,
}

// CalculateRecentPastRunTime calculates the most recent past run time for a schedule entry.
func CalculateRecentPastRunTime(entry ScheduleEntry, now time.Time) (time.Time, error) {
	targetWeekday, ok := DayOfWeekMap[entry.DayOfWeek]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid day of week: %s", entry.DayOfWeek)
	}

	startTime, err := time.ParseInLocation("150405", entry.StartTime, JST)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid start time format '%s': %w", entry.StartTime, err)
	}

	// Calculate the difference in days from today to the target weekday
	daysOffset := int(targetWeekday) - int(now.Weekday())
	
	// Create a candidate time for this week at the target start time.
	candidate := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, JST)
	candidate = candidate.AddDate(0, 0, daysOffset)

	// Now check if this candidate is in the past or future relative to 'now'.
	if candidate.After(now) {
		// If the candidate is in the future, then the most recent past occurrence must be last week.
		return candidate.AddDate(0, 0, -7), nil
	} else {
		// If the candidate is in the past or exactly 'now', then this is the most recent past.
		return candidate, nil
	}
}
