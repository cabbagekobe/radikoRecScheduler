package internal

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Checker verifies that rec_radiko_ts.sh and its runtime tool dependencies
// are reachable from the current environment.
type Checker struct {
	LookPath   func(name string) (string, error)
	ScriptPath string
}

// NewChecker constructs a Checker using exec.LookPath and the configured
// rec_radiko_ts.sh path (RADIKO_REC_TS_SCRIPT or the bare script name).
func NewChecker() *Checker {
	return &Checker{
		LookPath:   exec.LookPath,
		ScriptPath: resolveScriptPath(),
	}
}

type dependency struct {
	Label  string
	Lookup string
	Hint   string
}

// Check writes a human-readable installation report to out and returns a
// non-nil error when any required dependency is missing.
func (c *Checker) Check(out io.Writer) error {
	deps := []dependency{
		{
			Label:  "rec_radiko_ts.sh",
			Lookup: c.ScriptPath,
			Hint:   "git clone https://github.com/uru2/rec_radiko_ts and put rec_radiko_ts.sh on PATH or set RADIKO_REC_TS_SCRIPT to its absolute path",
		},
		{
			Label:  "ffmpeg",
			Lookup: "ffmpeg",
			Hint:   "macOS: brew install ffmpeg / Debian/Ubuntu: apt install ffmpeg",
		},
		{
			Label:  "curl",
			Lookup: "curl",
			Hint:   "macOS: brew install curl / Debian/Ubuntu: apt install curl",
		},
		{
			Label:  "xmllint",
			Lookup: "xmllint",
			Hint:   "macOS: brew install libxml2 / Debian/Ubuntu: apt install libxml2-utils",
		},
	}

	fmt.Fprintln(out, "Checking dependencies for radikoRecScheduler:")
	fmt.Fprintln(out)

	var missing []string
	for _, d := range deps {
		path, err := c.LookPath(d.Lookup)
		if err != nil {
			fmt.Fprintf(out, "  [MISSING] %-18s install hint: %s\n", d.Label, d.Hint)
			missing = append(missing, d.Label)
			continue
		}
		fmt.Fprintf(out, "  [OK]      %-18s -> %s\n", d.Label, path)
	}

	fmt.Fprintln(out)
	if len(missing) > 0 {
		fmt.Fprintf(out, "Result: %d missing (%s).\n", len(missing), strings.Join(missing, ", "))
		return errors.New("missing dependencies")
	}
	fmt.Fprintln(out, "Result: All dependencies OK.")
	return nil
}
