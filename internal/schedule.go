package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

// ScheduleEntry corresponds to an entry in the schedule.json file.
type ScheduleEntry struct {
	ProgramName string `json:"program_name"`
	DayOfWeek   string `json:"day_of_week"`
	StartTime   string `json:"start_time"`
	StationID   string `json:"station_id"`
}

// LoadSchedule reads and parses the schedule file from the given path.
func LoadSchedule(filePath string) ([]ScheduleEntry, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading schedule file '%s': %w", filePath, err)
	}

	var scheduleEntries []ScheduleEntry
	if err := json.Unmarshal(file, &scheduleEntries); err != nil {
		return nil, fmt.Errorf("error parsing JSON from '%s': %w", filePath, err)
	}

	return scheduleEntries, nil
}
