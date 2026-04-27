package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ExecuteJob resolves the program metadata for a schedule entry and delegates
// the actual recording to the supplied Recorder.
func ExecuteJob(rec Recorder, entry ScheduleEntry, pastTime time.Time, outputDir string) error {
	log.Printf("INFO: Starting recording for: %s (%s) for past broadcast at %s", entry.ProgramName, entry.StationID, pastTime.Format("2006-01-02 15:04:05"))

	programName := resolveProgramName(entry)

	outputFileName := fmt.Sprintf("%s-%s-%s.m4a", pastTime.Format("20060102150405"), entry.StationID, programName)
	outputFilePath := filepath.Join(outputDir, outputFileName)

	if _, err := os.Stat(outputFilePath); err == nil {
		log.Printf("INFO: File already exists, skipping: %s", outputFilePath)
		return nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	programURL := fmt.Sprintf("https://radiko.jp/#!/ts/%s/%s", entry.StationID, pastTime.Format("20060102150405"))
	log.Printf("INFO: Recording from %s to %s", programURL, outputFilePath)

	if err := rec.Record(context.Background(), programURL, outputFilePath); err != nil {
		return fmt.Errorf("failed to record %s: %w", entry.ProgramName, err)
	}
	log.Printf("INFO: Successfully recorded and saved to: %s", outputFilePath)
	return nil
}

// getProgramGuide is the program-guide fetcher used by ExecuteJob. It is a
// package-level variable so tests can stub it without touching the network.
var getProgramGuide = GetProgramGuide

// resolveProgramName returns the program title from the radiko program guide,
// falling back to the schedule entry's program name on any error.
func resolveProgramName(entry ScheduleEntry) string {
	programData, err := getProgramGuide(entry.StationID)
	if err != nil {
		log.Printf("WARNING: Failed to get program guide for station %s, falling back to schedule.json: %v", entry.StationID, err)
		return entry.ProgramName
	}
	dayOfWeek, err := toEnglishDayOfWeek(entry.DayOfWeek)
	if err != nil {
		log.Printf("WARNING: %v, falling back to schedule.json", err)
		return entry.ProgramName
	}
	name, err := FindProgramTitle(programData, entry.StartTime, dayOfWeek)
	if err != nil {
		log.Printf("WARNING: Failed to find program name for %s at %s on %s, falling back to schedule.json: %v", entry.StationID, entry.StartTime, entry.DayOfWeek, err)
		return entry.ProgramName
	}
	log.Printf("INFO: Successfully found program name: %s", name)
	return name
}

func toEnglishDayOfWeek(dayOfWeek string) (string, error) {
	switch dayOfWeek {
	case "日":
		return "Sun", nil
	case "月":
		return "Mon", nil
	case "火":
		return "Tue", nil
	case "水":
		return "Wed", nil
	case "木":
		return "Thu", nil
	case "金":
		return "Fri", nil
	case "土":
		return "Sat", nil
	default:
		return "", fmt.Errorf("invalid day of week: %s", dayOfWeek)
	}
}
