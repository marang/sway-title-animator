package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHelpListsStableCommandContract(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, &stdout, &stderr)

	if exitCode != exitSuccess {
		t.Fatalf("expected success, got %d", exitCode)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
	for _, expected := range []string{"register [options]", "restore [context]", "list", "archive <context>", "activate <context>", "purge <context>", "--json", "3  Recognized command unavailable"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("help does not contain %q:\n%s", expected, stdout.String())
		}
	}
}

func TestCommandHelpWorksAfterCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"archive", "--help"}, &stdout, &stderr)

	if exitCode != exitSuccess || stderr.Len() != 0 {
		t.Fatalf("expected help success, got code=%d stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage: sway-session [--json] archive <context>") {
		t.Fatalf("unexpected command help: %q", stdout.String())
	}
}

func TestInvalidArityHasActionableTextDiagnostic(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"purge"}, &stdout, &stderr)

	if exitCode != exitUsage {
		t.Fatalf("expected usage exit code, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	for _, expected := range []string{"sway-session: error:", "invalid arguments", "Usage: sway-session purge <context>"} {
		if !strings.Contains(stderr.String(), expected) {
			t.Fatalf("diagnostic does not contain %q: %s", expected, stderr.String())
		}
	}
}

func TestJSONDiagnosticIsStableAndContainsNoPresentationText(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"list", "--json"}, &stdout, &stderr)

	if exitCode != exitUnavailable {
		t.Fatalf("expected unavailable exit code, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	var output struct {
		Diagnostics []struct {
			Level   string `json:"level"`
			Code    string `json:"code"`
			Message string `json:"message"`
			Hint    string `json:"hint"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &output); err != nil {
		t.Fatalf("decode JSON diagnostic: %v\n%s", err, stderr.String())
	}
	if len(output.Diagnostics) != 1 || output.Diagnostics[0].Level != "error" || output.Diagnostics[0].Code != "not_implemented" {
		t.Fatalf("unexpected diagnostic: %+v", output.Diagnostics)
	}
	if strings.Contains(stderr.String(), "sway-session: error") {
		t.Fatalf("JSON output contains human presentation prefix: %s", stderr.String())
	}
}

func TestUnknownCommandUsesUsageExitCode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"launch"}, &stdout, &stderr)

	if exitCode != exitUsage || !strings.Contains(stderr.String(), "unknown command \"launch\"") {
		t.Fatalf("unexpected result code=%d stderr=%q", exitCode, stderr.String())
	}
}
