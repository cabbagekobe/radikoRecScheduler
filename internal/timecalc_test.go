package internal

import (
	"testing"
	"time"
)

func TestCalculateRecentPastRunTime(t *testing.T) {
	tests := []struct {
		name        string
		entry       ScheduleEntry
		now         time.Time
		expected    time.Time
		expectError bool
	}{
		{
			name: "Target day today, time in past", // Today is Tuesday 10:00, Target Tuesday 03:00
			entry: ScheduleEntry{
				DayOfWeek: "火",
				StartTime: "030000",
			},
			now:      time.Date(2026, time.January, 13, 10, 0, 0, 0, JST), // Tuesday
			expected: time.Date(2026, time.January, 13, 3, 0, 0, 0, JST),
		},
		{
			name: "Target day today, time in future", // Today is Tuesday 10:00, Target Tuesday 15:00
			entry: ScheduleEntry{
				DayOfWeek: "火",
				StartTime: "150000",
			},
			now:      time.Date(2026, time.January, 13, 10, 0, 0, 0, JST), // Tuesday
			expected: time.Date(2026, time.January, 6, 15, 0, 0, 0, JST), // Last Tuesday
		},
		{
			name: "Target day before today (this week)", // Today is Tuesday, Target Monday
			entry: ScheduleEntry{
				DayOfWeek: "月",
				StartTime: "100000",
			},
			now:      time.Date(2026, time.January, 13, 10, 0, 0, 0, JST), // Tuesday
			expected: time.Date(2026, time.January, 12, 10, 0, 0, 0, JST), // This Monday
		},
		{
			name: "Target day after today (last week)", // Today is Tuesday, Target Wednesday
			entry: ScheduleEntry{
				DayOfWeek: "水",
				StartTime: "100000",
			},
			now:      time.Date(2026, time.January, 13, 10, 0, 0, 0, JST), // Tuesday
			expected: time.Date(2026, time.January, 7, 10, 0, 0, 0, JST), // Last Wednesday
		},
		{
			name: "Invalid DayOfWeek",
			entry: ScheduleEntry{
				DayOfWeek: "Invalid",
				StartTime: "100000",
			},
			now:         time.Date(2026, time.January, 13, 10, 0, 0, 0, JST),
			expectError: true,
		},
		{
			name: "Invalid StartTime format",
			entry: ScheduleEntry{
				DayOfWeek: "月",
				StartTime: "100", // Invalid format
			},
			now:         time.Date(2026, time.January, 13, 10, 0, 0, 0, JST),
			expectError: true,
		},
		{
			name: "Now is exactly target time", // Today is Tuesday 10:00, Target Tuesday 10:00
			entry: ScheduleEntry{
				DayOfWeek: "火",
				StartTime: "100000",
			},
			now:      time.Date(2026, time.January, 13, 10, 0, 0, 0, JST), // Tuesday
			expected: time.Date(2026, time.January, 13, 10, 0, 0, 0, JST),
		},
		{
			name: "Across Sunday to Monday, now Monday early morning, target Sunday evening",
			entry: ScheduleEntry{
				DayOfWeek: "日",
				StartTime: "200000", // 8 PM Sunday
			},
			now:      time.Date(2026, time.January, 12, 1, 0, 0, 0, JST), // Monday 1 AM
			expected: time.Date(2026, time.January, 11, 20, 0, 0, 0, JST), // Previous Sunday 8 PM
		},
		{
			name: "Across Sunday to Monday, now Sunday evening, target Monday early morning",
			entry: ScheduleEntry{
				DayOfWeek: "月",
				StartTime: "010000", // 1 AM Monday
			},
			now:      time.Date(2026, time.January, 11, 23, 0, 0, 0, JST), // Sunday 11 PM
			expected: time.Date(2026, time.January, 5, 1, 0, 0, 0, JST), // Previous Monday 1 AM
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateRecentPastRunTime(tt.entry, tt.now)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error for %s, but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for %s, but got: %v", tt.name, err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("for %s, expected %s, but got %s", tt.name, tt.expected.Format(time.RFC3339), result.Format(time.RFC3339))
				}
			}
		})
	}
}
