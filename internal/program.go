package internal

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Radiko is the root element of the program guide XML.
type Radiko struct {
	XMLName  xml.Name `xml:"radiko"`
	Stations Stations `xml:"stations"`
}

// Stations contains a list of stations.
type Stations struct {
	XMLName xml.Name  `xml:"stations"`
	Station []Station `xml:"station"`
}

// Station contains the program guide for a single station.
type Station struct {
	XMLName xml.Name `xml:"station"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name"`
	Progs   Progs    `xml:"progs"`
}

// Progs contains a list of programs.
type Progs struct {
	XMLName xml.Name `xml:"progs"`
	Prog    []Prog   `xml:"prog"`
	Date    string   `xml:"date"`
}

// Prog represents a single program.
type Prog struct {
	XMLName  xml.Name `xml:"prog"`
	Ft       string   `xml:"ft,attr"`
	To       string   `xml:"to,attr"`
	Ftl      string   `xml:"ftl,attr"`
	Tol      string   `xml:"tol,attr"`
	Dur      string   `xml:"dur,attr"`
	Title    string   `xml:"title"`
	SubTitle string   `xml:"sub_title"`
	Pfm      string   `xml:"pfm"`
	Desc     string   `xml:"desc"`
	Info     string   `xml:"info"`
	URL      string   `xml:"url"`
}

// GetProgramGuide fetches the program guide for a given station.
func GetProgramGuide(stationID string) ([]byte, error) {
	url := fmt.Sprintf("http://radiko.jp/v3/program/station/weekly/%s.xml", stationID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get program guide: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get program guide: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read program guide: %w", err)
	}

	return body, nil
}

// FindProgramTitle finds a program title by start time and day of week from the program guide XML.
func FindProgramTitle(programData []byte, targetTime, targetDayOfWeek string) (string, error) {
	var radiko Radiko
	if err := xml.Unmarshal(programData, &radiko); err != nil {
		return "", fmt.Errorf("failed to unmarshal program guide: %w", err)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone: %w", err)
	}

	for _, station := range radiko.Stations.Station {
		for _, prog := range station.Progs.Prog {
			// "20060102150405" is the layout for "YYYYMMDDHHmmss"
			startTime, err := time.ParseInLocation("20060102150405", prog.Ft, jst)
			if err != nil {
				// skip if format is wrong
				continue
			}

			// Format the start time to "HHmm" and day of the week to "Mon"
			progStartTime := startTime.Format("1504")
			progDayOfWeek := startTime.Weekday().String()[:3]

			if progStartTime == targetTime && strings.EqualFold(progDayOfWeek, targetDayOfWeek) {
				return prog.Title, nil
			}
		}
	}

	return "", fmt.Errorf("program not found for time %s on %s", targetTime, targetDayOfWeek)
}
