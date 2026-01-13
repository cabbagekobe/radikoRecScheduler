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
		fmt.Fprintln(os.Stderr, "\nSchedule is loaded from schedule.json.")
		fmt.Fprintln(os.Stderr, "For detailed usage and configuration, refer to README.md.")
	}

	scheduleFilePath := flag.String("file", "schedule.json", "Path to the schedule JSON file.")
	flag.Parse()

	scheduleEntries, err := internal.LoadSchedule(*scheduleFilePath)
	if err != nil {
		log.Fatalf("Failed to load schedule: %v", err)
	}

	now := time.Now().In(internal.JST)
	for _, entry := range scheduleEntries {
		recentPastTime, err := internal.CalculateRecentPastRunTime(entry, now)
		if err != nil {
			log.Printf("Error calculating recent past run time for '%s': %v", entry.ProgramName, err)
			continue
		}

		if err := internal.ExecuteJob(entry, recentPastTime); err != nil {
			log.Printf("Error executing job for '%s': %v", entry.ProgramName, err)
		}
	}

	log.Println("All scheduled past broadcasts processed. Exiting.")
}


