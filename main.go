package main

import (
	"flag"
	"fmt" // Added
	"log"
	"os"   // Added
	"time"

	"radigoSchedule/internal" // Assuming radigoSchedule is the module name
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nSchedule is loaded from the path specified by -file flag, or from XDG config directory by default.")
		fmt.Fprintln(os.Stderr, "For detailed usage and configuration, refer to README.md.")
	}

	scheduleFilePath := flag.String("file", func() string {
		path, err := internal.GetScheduleConfigPath()
		if err != nil {
			log.Fatalf("Failed to get default schedule config path: %v", err)
		}
		return path
	}(), "Path to the schedule JSON file. Defaults to XDG config directory.")
	flag.Parse()

	scheduleEntries, err := internal.LoadSchedule(*scheduleFilePath)
	if err != nil {
		// If schedule.json does not exist in the XDG config path, try to load from the current directory for backward compatibility
		if os.IsNotExist(err) && *scheduleFilePath == func() string {
			path, _ := internal.GetScheduleConfigPath()
			return path
		}() {
			log.Printf("Schedule file not found at default XDG config path. Trying current directory for 'schedule.json'.")
			*scheduleFilePath = "schedule.json"
			scheduleEntries, err = internal.LoadSchedule(*scheduleFilePath)
			if err != nil {
				log.Fatalf("Failed to load schedule from XDG path and current directory: %v", err)
			}
		} else {
			log.Fatalf("Failed to load schedule: %v", err)
		}
	}

	now := time.Now().In(internal.JST)
	for _, entry := range scheduleEntries {
		recentPastTime, err := internal.CalculateRecentPastRunTime(entry, now)
		if err != nil {
			log.Printf("Error calculating recent past run time for '%s': %v", entry.ProgramName, err)
			continue
		}

		if err := internal.ExecuteJob(entry, recentPastTime, "output"); err != nil {
			log.Printf("Error executing job for '%s': %v", entry.ProgramName, err)
		}
	}

	log.Println("All scheduled past broadcasts processed. Exiting.")
}


