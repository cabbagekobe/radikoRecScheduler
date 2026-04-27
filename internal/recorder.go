package internal

import (
	"context"
	"os"
	"os/exec"
)

const (
	scriptPathEnvVar  = "RADIKO_REC_TS_SCRIPT"
	defaultScriptName = "rec_radiko_ts.sh"
)

// resolveScriptPath returns the configured rec_radiko_ts.sh path, honoring
// the RADIKO_REC_TS_SCRIPT environment variable when set.
func resolveScriptPath() string {
	if p := os.Getenv(scriptPathEnvVar); p != "" {
		return p
	}
	return defaultScriptName
}

// Recorder records a radiko timefree program identified by programURL into outputPath.
type Recorder interface {
	Record(ctx context.Context, programURL, outputPath string) error
}

// RecRadikoTSRecorder records via the rec_radiko_ts.sh shell script.
type RecRadikoTSRecorder struct {
	ScriptPath  string
	execCommand func(ctx context.Context, name string, args ...string) *exec.Cmd
}

// NewRecRadikoTSRecorder returns a Recorder backed by rec_radiko_ts.sh.
// The script path can be overridden via the RADIKO_REC_TS_SCRIPT environment variable.
func NewRecRadikoTSRecorder() *RecRadikoTSRecorder {
	return &RecRadikoTSRecorder{
		ScriptPath:  resolveScriptPath(),
		execCommand: exec.CommandContext,
	}
}

func (r *RecRadikoTSRecorder) Record(ctx context.Context, programURL, outputPath string) error {
	cmd := r.execCommand(ctx, r.ScriptPath, "-u", programURL, "-o", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
