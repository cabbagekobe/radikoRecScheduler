package internal

import (
	"context"
	"os/exec"
	"reflect"
	"testing"
)

func TestRecRadikoTSRecorder_Record_BuildsExpectedArgs(t *testing.T) {
	var gotName string
	var gotArgs []string

	rec := &RecRadikoTSRecorder{
		ScriptPath: "rec_radiko_ts.sh",
		execCommand: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			gotName = name
			gotArgs = args
			return exec.CommandContext(ctx, "true")
		},
	}

	const programURL = "https://radiko.jp/#!/ts/LFR/20260111010000"
	const outputPath = "output/20260111010000-LFR-Test.m4a"
	if err := rec.Record(context.Background(), programURL, outputPath); err != nil {
		t.Fatalf("Record returned unexpected error: %v", err)
	}

	if gotName != "rec_radiko_ts.sh" {
		t.Errorf("script name = %q, want %q", gotName, "rec_radiko_ts.sh")
	}
	wantArgs := []string{"-u", programURL, "-o", outputPath}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestRecRadikoTSRecorder_Record_PropagatesScriptFailure(t *testing.T) {
	rec := &RecRadikoTSRecorder{
		ScriptPath: "rec_radiko_ts.sh",
		execCommand: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "false")
		},
	}

	err := rec.Record(context.Background(), "https://radiko.jp/#!/ts/LFR/20260111010000", "/tmp/out.m4a")
	if err == nil {
		t.Fatal("expected error from failing script, got nil")
	}
}

func TestNewRecRadikoTSRecorder_UsesEnvOverride(t *testing.T) {
	t.Setenv(scriptPathEnvVar, "/opt/custom/rec_radiko_ts.sh")
	rec := NewRecRadikoTSRecorder()
	if rec.ScriptPath != "/opt/custom/rec_radiko_ts.sh" {
		t.Errorf("ScriptPath = %q, want %q", rec.ScriptPath, "/opt/custom/rec_radiko_ts.sh")
	}
}

func TestNewRecRadikoTSRecorder_DefaultsToScriptName(t *testing.T) {
	t.Setenv(scriptPathEnvVar, "")
	rec := NewRecRadikoTSRecorder()
	if rec.ScriptPath != "rec_radiko_ts.sh" {
		t.Errorf("ScriptPath = %q, want %q", rec.ScriptPath, "rec_radiko_ts.sh")
	}
}
