package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun_JSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"-seed", "42", "-format", "json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run error: %v (stderr: %s)", err, stderr.String())
	}
	// Decode into a generic map; verify StellarCount is at least 1.
	var form map[string]any
	if derr := json.Unmarshal(stdout.Bytes(), &form); derr != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", derr, stdout.String())
	}
	if v, ok := form["StellarCount"].(float64); !ok || v < 1 {
		t.Fatalf("StellarCount missing or invalid: %v", form["StellarCount"])
	}
}

func TestRun_Short(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"-seed", "42", "-format", "short"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run error: %v (stderr: %s)", err, stderr.String())
	}
	out := strings.TrimSpace(stdout.String())
	// Short profile starts with "<count>-" then class.
	if !strings.HasPrefix(out, "1-") && !strings.HasPrefix(out, "2-") &&
		!strings.HasPrefix(out, "3-") && !strings.HasPrefix(out, "4-") &&
		!strings.HasPrefix(out, "5-") {
		t.Fatalf("short profile prefix unexpected: %q", out)
	}
	// Single-star profiles have no colons; multi-star have at least one.
	// Either is valid; just check we got SOMETHING with hyphens.
	if !strings.Contains(out, "-") {
		t.Fatalf("short profile missing field separators: %q", out)
	}
}

func TestRun_UnknownFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"-seed", "42", "-format", "xml"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("error doesn't mention 'unknown format': %v", err)
	}
}

func TestRun_TimeBasedSeed(t *testing.T) {
	// seed=0 -> time-based. Just verify no error and the output contains JSON.
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-format", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if !strings.Contains(stdout.String(), "StellarCount") {
		t.Fatalf("expected JSON output for time-based seed")
	}
}

func TestRun_DefaultIsMarkdown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v (stderr=%q)", err, stderr.String())
	}
	out := stdout.String()
	trimmed := strings.TrimLeft(out, "\n")
	if !strings.HasPrefix(trimmed, "# Star System:") {
		preview := out
		if len(preview) > 80 {
			preview = preview[:80]
		}
		t.Errorf("expected output to start with '# Star System:'; got first 80 chars: %q", preview)
	}
	if !strings.Contains(out, "## IISS Class 0/I Survey") {
		t.Errorf("expected Class 0/I H2 in default output")
	}
}

func TestRun_FormatMarkdownExplicit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "markdown"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "## IISS Class 0/I Survey") {
		t.Errorf("expected Class 0/I H2 with -format markdown")
	}
}

func TestRun_FormatJSONStillWorks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Errorf("expected valid JSON; got error %v on output:\n%s", err, stdout.String())
	}
}

func TestRun_FormatShortStillWorks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-seed", "42", "-format", "short"}, &stdout, &stderr); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	out := strings.TrimSpace(stdout.String())
	if strings.Contains(out, "\n") {
		t.Errorf("expected single-line short output; got:\n%s", out)
	}
}
