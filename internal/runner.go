package internal

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// ExecuteJob runs the recording command for a given schedule entry and time.
func ExecuteJob(entry ScheduleEntry, pastTime time.Time) error {
	log.Printf("Executing: %s (%s) for past broadcast at %s\n", entry.ProgramName, entry.StationID, pastTime.Format("2006-01-02 15:04:05"))
	
	cmdArgs := []string{
		"rec",
		"-id=" + entry.StationID,
		"-s=" + pastTime.Format("20060102150405"),
	}

	cmd := exec.Command(AppConfig.RadigoCommandPath, cmdArgs...)
	
	// Capture stdout and stderr to log, instead of directly to os.Stdout/Stderr
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing command '%s %v': %s, %w", AppConfig.RadigoCommandPath, cmdArgs, cmdOutput, err)
	}

	log.Printf("Command output for '%s %v':\n%s", AppConfig.RadigoCommandPath, cmdArgs, cmdOutput)
	
	return nil
}
