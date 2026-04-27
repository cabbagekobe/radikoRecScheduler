package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockRecorder records the arguments it was called with. By default it
// returns nil; tests that need other behaviour set RecordFn explicitly.
type MockRecorder struct {
	RecordFn  func(ctx context.Context, programURL, outputPath string) error
	GotURL    string
	GotOutput string
	CallCount int
}

func (m *MockRecorder) Record(ctx context.Context, programURL, outputPath string) error {
	m.CallCount++
	m.GotURL = programURL
	m.GotOutput = outputPath
	if m.RecordFn != nil {
		return m.RecordFn(ctx, programURL, outputPath)
	}
	return nil
}

// stubGetProgramGuide replaces the package-level getProgramGuide for the
// duration of the test, restoring it on cleanup.
func stubGetProgramGuide(t *testing.T, response []byte, err error) {
	t.Helper()
	orig := getProgramGuide
	getProgramGuide = func(stationID string) ([]byte, error) {
		return response, err
	}
	t.Cleanup(func() { getProgramGuide = orig })
}

func TestExecuteJob_SuccessBuildsURLAndOutputPath(t *testing.T) {
	stubGetProgramGuide(t, nil, errors.New("network disabled in test"))
	tempOutputDir := t.TempDir()

	pastTime := time.Date(2026, time.January, 12, 1, 0, 0, 0, JST) // Mon
	entry := ScheduleEntry{
		ProgramName: "Test Program",
		DayOfWeek:   "月",
		StartTime:   "010000",
		StationID:   "ST1",
	}

	rec := &MockRecorder{}
	if err := ExecuteJob(rec, entry, pastTime, tempOutputDir); err != nil {
		t.Fatalf("ExecuteJob returned unexpected error: %v", err)
	}

	wantURL := "https://radiko.jp/#!/ts/ST1/20260112010000"
	if rec.GotURL != wantURL {
		t.Errorf("programURL = %q, want %q", rec.GotURL, wantURL)
	}

	// With the program guide stubbed to error, the fallback is entry.ProgramName.
	wantOutput := filepath.Join(tempOutputDir, "20260112010000-ST1-Test Program.m4a")
	if rec.GotOutput != wantOutput {
		t.Errorf("outputPath = %q, want %q", rec.GotOutput, wantOutput)
	}
}

func TestExecuteJob_UsesProgramTitleFromGuideWhenAvailable(t *testing.T) {
	guideXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<radiko>
  <stations>
    <station id="ST1">
      <name>Station 1</name>
      <progs>
        <date>20260112</date>
        <prog ft="20260112010000" to="20260112030000" ftl="0100" tol="0300" dur="7200">
          <title>Looked Up Title</title>
        </prog>
      </progs>
    </station>
  </stations>
</radiko>`)
	stubGetProgramGuide(t, guideXML, nil)
	tempOutputDir := t.TempDir()

	pastTime := time.Date(2026, time.January, 12, 1, 0, 0, 0, JST) // Mon
	entry := ScheduleEntry{
		ProgramName: "Fallback Name",
		DayOfWeek:   "月",
		StartTime:   "010000",
		StationID:   "ST1",
	}

	rec := &MockRecorder{}
	if err := ExecuteJob(rec, entry, pastTime, tempOutputDir); err != nil {
		t.Fatalf("ExecuteJob returned unexpected error: %v", err)
	}

	wantOutput := filepath.Join(tempOutputDir, "20260112010000-ST1-Looked Up Title.m4a")
	if rec.GotOutput != wantOutput {
		t.Errorf("outputPath = %q, want %q", rec.GotOutput, wantOutput)
	}
}

func TestExecuteJob_SkipsWhenOutputFileExists(t *testing.T) {
	stubGetProgramGuide(t, nil, errors.New("network disabled in test"))
	tempOutputDir := t.TempDir()

	pastTime := time.Date(2026, time.January, 12, 1, 0, 0, 0, JST)
	entry := ScheduleEntry{
		ProgramName: "Existing Program",
		DayOfWeek:   "月",
		StartTime:   "010000",
		StationID:   "ST1",
	}

	fallbackName := fmt.Sprintf("%s-%s-%s.m4a", pastTime.Format("20060102150405"), entry.StationID, entry.ProgramName)
	if err := os.WriteFile(filepath.Join(tempOutputDir, fallbackName), []byte("EXISTING"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	rec := &MockRecorder{
		RecordFn: func(ctx context.Context, programURL, outputPath string) error {
			t.Fatalf("Record should not be called when output already exists; got url=%q out=%q", programURL, outputPath)
			return nil
		},
	}
	if err := ExecuteJob(rec, entry, pastTime, tempOutputDir); err != nil {
		t.Fatalf("ExecuteJob returned unexpected error: %v", err)
	}
	if rec.CallCount != 0 {
		t.Errorf("expected Record not to be called, but it was called %d times", rec.CallCount)
	}
}

func TestExecuteJob_PropagatesRecorderError(t *testing.T) {
	stubGetProgramGuide(t, nil, errors.New("network disabled in test"))
	tempOutputDir := t.TempDir()

	pastTime := time.Date(2026, time.January, 12, 1, 0, 0, 0, JST)
	entry := ScheduleEntry{
		ProgramName: "Failing Program",
		DayOfWeek:   "月",
		StartTime:   "010000",
		StationID:   "ST1",
	}

	rec := &MockRecorder{
		RecordFn: func(ctx context.Context, programURL, outputPath string) error {
			return fmt.Errorf("rec_radiko_ts.sh exited 1")
		},
	}
	err := ExecuteJob(rec, entry, pastTime, tempOutputDir)
	if err == nil {
		t.Fatal("expected error from failing recorder, got nil")
	}
	if !strings.Contains(err.Error(), "rec_radiko_ts.sh exited 1") {
		t.Errorf("error %q does not contain expected substring", err.Error())
	}
}

func TestExecuteJob_CreatesMissingOutputDir(t *testing.T) {
	stubGetProgramGuide(t, nil, errors.New("network disabled in test"))
	parent := t.TempDir()
	nestedDir := filepath.Join(parent, "does", "not", "exist", "yet")

	pastTime := time.Date(2026, time.January, 12, 1, 0, 0, 0, JST)
	entry := ScheduleEntry{
		ProgramName: "Nested Program",
		DayOfWeek:   "月",
		StartTime:   "010000",
		StationID:   "ST1",
	}

	rec := &MockRecorder{}
	if err := ExecuteJob(rec, entry, pastTime, nestedDir); err != nil {
		t.Fatalf("ExecuteJob returned unexpected error: %v", err)
	}
	if _, err := os.Stat(nestedDir); err != nil {
		t.Errorf("expected nested output dir to exist: %v", err)
	}
}
