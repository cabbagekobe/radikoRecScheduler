package internal

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestChecker_AllPresent(t *testing.T) {
	var buf bytes.Buffer
	c := &Checker{
		LookPath: func(name string) (string, error) {
			return "/usr/local/bin/" + name, nil
		},
		ScriptPath: "rec_radiko_ts.sh",
	}
	if err := c.Check(&buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "All dependencies OK") {
		t.Errorf("expected OK message, got:\n%s", out)
	}
	if strings.Contains(out, "[MISSING]") {
		t.Errorf("did not expect MISSING marker, got:\n%s", out)
	}
}

func TestChecker_SomeMissing(t *testing.T) {
	var buf bytes.Buffer
	c := &Checker{
		LookPath: func(name string) (string, error) {
			if name == "xmllint" {
				return "", errors.New("not found")
			}
			return "/usr/local/bin/" + name, nil
		},
		ScriptPath: "rec_radiko_ts.sh",
	}
	err := c.Check(&buf)
	if err == nil {
		t.Fatal("expected error for missing xmllint, got nil")
	}
	out := buf.String()
	if !strings.Contains(out, "[MISSING]") {
		t.Errorf("expected MISSING marker, got:\n%s", out)
	}
	if !strings.Contains(out, "xmllint") {
		t.Errorf("expected xmllint in output, got:\n%s", out)
	}
}

func TestChecker_RespectsScriptPath(t *testing.T) {
	var buf bytes.Buffer
	var lookedUp []string
	c := &Checker{
		LookPath: func(name string) (string, error) {
			lookedUp = append(lookedUp, name)
			return "/usr/local/bin/" + name, nil
		},
		ScriptPath: "/opt/custom/rec_radiko_ts.sh",
	}
	if err := c.Check(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lookedUp) == 0 || lookedUp[0] != "/opt/custom/rec_radiko_ts.sh" {
		t.Errorf("expected first LookPath call with custom script path, got %v", lookedUp)
	}
}

func TestNewChecker_UsesEnvOverride(t *testing.T) {
	t.Setenv(scriptPathEnvVar, "/opt/custom/rec_radiko_ts.sh")
	c := NewChecker()
	if c.ScriptPath != "/opt/custom/rec_radiko_ts.sh" {
		t.Errorf("ScriptPath = %q, want %q", c.ScriptPath, "/opt/custom/rec_radiko_ts.sh")
	}
}

func TestNewChecker_DefaultsToScriptName(t *testing.T) {
	t.Setenv(scriptPathEnvVar, "")
	c := NewChecker()
	if c.ScriptPath != "rec_radiko_ts.sh" {
		t.Errorf("ScriptPath = %q, want %q", c.ScriptPath, "rec_radiko_ts.sh")
	}
}
