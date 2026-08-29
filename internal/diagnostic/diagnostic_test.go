package diagnostic

import (
	"bytes"
	"strings"
	"testing"
)

func TestTextDiagnosticEscapesTerminalControls(t *testing.T) {
	var output bytes.Buffer
	item := Diagnostic{
		Level:   LevelError,
		Message: "unknown \x1b[31mcommand",
		Hint:    "try\nagain",
	}

	if err := Write(&output, "sway-session", item, false); err != nil {
		t.Fatalf("write diagnostic: %v", err)
	}
	got := output.String()
	if strings.ContainsRune(got, '\x1b') {
		t.Fatalf("diagnostic contains a terminal escape: %q", got)
	}
	if !strings.Contains(got, `unknown \x1b[31mcommand`) || !strings.Contains(got, `try\nagain`) {
		t.Fatalf("diagnostic does not visibly escape controls: %q", got)
	}
	if strings.Count(got, "\n") != 2 {
		t.Fatalf("control characters injected extra lines: %q", got)
	}
}
