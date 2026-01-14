package internal

import (
	"testing"
)

func TestFindProgramTitle(t *testing.T) {
	testXMLData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<radiko>
  <stations>
    <station id="TBS">
      <name>TBSラジオ</name>
      <progs>
        <prog ft="20240115180000" to="20240115210000" ftl="1800" tol="2100" dur="10800">
          <title>アフター６ジャンクション</title>
        </prog>
        <prog ft="20240116030000" to="20240116040000" ftl="0300" tol="0400" dur="3600">
          <title>火曜JUNK 爆笑問題カーボーイ</title>
        </prog>
      </progs>
    </station>
  </stations>
</radiko>`)

	tests := []struct {
		name            string
		programData     []byte
		targetTime      string
		targetDayOfWeek string
		wantTitle       string
		wantErr         bool
	}{
		{
			name:            "Success: Find program on Monday",
			programData:     testXMLData,
			targetTime:      "1800",
			targetDayOfWeek: "Mon",
			wantTitle:       "アフター６ジャンクション",
			wantErr:         false,
		},
		{
			name:            "Success: Find program on Tuesday morning (edge case)",
			programData:     testXMLData,
			targetTime:      "0300",
			targetDayOfWeek: "Tue",
			wantTitle:       "火曜JUNK 爆笑問題カーボーイ",
			wantErr:         false,
		},
		{
			name:            "Failure: Program not found on specified time",
			programData:     testXMLData,
			targetTime:      "1900", // No program starts exactly at 19:00
			targetDayOfWeek: "Mon",
			wantTitle:       "",
			wantErr:         true,
		},
		{
			name:            "Failure: Program not found on specified day",
			programData:     testXMLData,
			targetTime:      "1800",
			targetDayOfWeek: "Fri",
			wantTitle:       "",
			wantErr:         true,
		},
		{
			name:            "Failure: Malformed XML",
			programData:     []byte(`<radiko><station`),
			targetTime:      "1800",
			targetDayOfWeek: "Mon",
			wantTitle:       "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, err := FindProgramTitle(tt.programData, tt.targetTime, tt.targetDayOfWeek)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProgramTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTitle != tt.wantTitle {
				t.Errorf("FindProgramTitle() gotTitle = %v, want %v", gotTitle, tt.wantTitle)
			}
		})
	}
}
