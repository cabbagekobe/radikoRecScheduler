package internal

import (
	"errors" // Added import
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSchedule_ValidFile(t *testing.T) {
	// Create a temporary valid schedule file
	content := `[
		{
			"program_name": "Test Program 1",
			"day_of_week": "月",
			"start_time": "100000",
			"station_id": "ST1"
		},
		{
			"program_name": "Test Program 2",
			"day_of_week": "火",
			"start_time": "110000",
			"station_id": "ST2"
		}
	]`
	tmpfile, err := os.CreateTemp("", "schedule-valid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// Load the schedule
	entries, err := LoadSchedule(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadSchedule failed: %v", err)
	}

	// Define expected entries
	expected := []ScheduleEntry{
		{
			ProgramName: "Test Program 1",
			DayOfWeek:   "月",
			StartTime:   "100000",
			StationID:   "ST1",
		},
		{
			ProgramName: "Test Program 2",
			DayOfWeek:   "火",
			StartTime:   "110000",
			StationID:   "ST2",
		},
	}

	// Compare actual and expected
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("LoadSchedule returned %+v, want %+v", entries, expected)
	}
}

func TestLoadSchedule_NonExistentFile(t *testing.T) {
	// Try to load a non-existent file
	nonExistentFile := filepath.Join(os.TempDir(), "non-existent-schedule.json")
	entries, err := LoadSchedule(nonExistentFile)

	if entries != nil {
		t.Errorf("LoadSchedule returned entries for non-existent file: %+v", entries)
	}
	if err == nil {
		t.Error("LoadSchedule did not return an error for non-existent file")
	}
	if !errors.Is(err, os.ErrNotExist) { // Corrected
		t.Errorf("LoadSchedule returned wrong error type for non-existent file: %v", err)
	}
}

func TestLoadSchedule_InvalidJson(t *testing.T) {
	// Create a temporary invalid JSON file
	content := `[
		{
			"program_name": "Test Program 1",
			"day_of_week": "月",
			"start_time": "100000",
			"station_id": "ST1"
		}
		// Missing comma and invalid structure
		{
			"program_name": "Test Program 2",
			"day_of_week": "火",
			"start_time": "110000",
			"station_id": "ST2"
		}
	`
	tmpfile, err := os.CreateTemp("", "schedule-invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// Load the schedule
	entries, err := LoadSchedule(tmpfile.Name())

	if entries != nil {
		t.Errorf("LoadSchedule returned entries for invalid JSON: %+v", entries)
	}
	if err == nil {
		t.Error("LoadSchedule did not return an error for invalid JSON")
	}
	// Check if the error indicates a JSON unmarshaling problem
	expectedErrSubstring := "error parsing JSON"
	if err != nil && !reflect.DeepEqual(err.Error()[:len(expectedErrSubstring)], expectedErrSubstring) {
		t.Errorf("LoadSchedule returned wrong error for invalid JSON: %v", err)
	}
}
