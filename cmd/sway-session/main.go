package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/marang/sway-title-animator/internal/diagnostic"
)

const (
	exitSuccess     = 0
	exitUsage       = 2
	exitUnavailable = 3
)

type commandSpec struct {
	usage   string
	summary string
	minimum int
	maximum int
}

var commandSpecs = map[string]commandSpec{
	"register": {usage: "register [options]", summary: "Register a persistent work context", minimum: 0, maximum: -1},
	"restore":  {usage: "restore [context]", summary: "Restore active or selected contexts", minimum: 0, maximum: 1},
	"list":     {usage: "list", summary: "List registered contexts", minimum: 0, maximum: 0},
	"archive":  {usage: "archive <context>", summary: "Exclude a context from automatic restore", minimum: 1, maximum: 1},
	"activate": {usage: "activate <context>", summary: "Return an archived context to automatic restore", minimum: 1, maximum: 1},
	"purge":    {usage: "purge <context>", summary: "Permanently remove a context and its saved state", minimum: 1, maximum: 1},
}

var commandOrder = []string{"register", "restore", "list", "archive", "activate", "purge"}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout io.Writer, stderr io.Writer) int {
	arguments, structured, help := globalOptions(arguments)
	if len(arguments) == 0 {
		if help {
			writeUsage(stdout)
			return exitSuccess
		}
		writeUsage(stderr)
		writeDiagnostic(stderr, structured, "usage", "a command is required", "Run sway-session --help to see available commands.")
		return exitUsage
	}

	if arguments[0] == "help" {
		if len(arguments) == 1 {
			writeUsage(stdout)
			return exitSuccess
		}
		if len(arguments) == 2 {
			return writeCommandHelp(arguments[1], stdout, stderr, structured)
		}
		writeDiagnostic(stderr, structured, "usage", "help accepts at most one command", "Run sway-session --help to see available commands.")
		return exitUsage
	}

	name := arguments[0]
	spec, exists := commandSpecs[name]
	if !exists {
		writeDiagnostic(stderr, structured, "unknown_command", fmt.Sprintf("unknown command %q", name), "Run sway-session --help to see available commands.")
		return exitUsage
	}
	if help {
		writeCommandUsage(stdout, name, spec)
		return exitSuccess
	}

	count := len(arguments) - 1
	if count < spec.minimum || spec.maximum >= 0 && count > spec.maximum {
		writeDiagnostic(stderr, structured, "usage", fmt.Sprintf("invalid arguments for %q", name), "Usage: sway-session "+spec.usage)
		return exitUsage
	}

	writeDiagnostic(
		stderr,
		structured,
		"not_implemented",
		fmt.Sprintf("command %q is not available in the Phase 1 foundation", name),
		"LAB-80 Phase 4 adds context lifecycle and Herdr launch behavior.",
	)
	return exitUnavailable
}

func globalOptions(arguments []string) ([]string, bool, bool) {
	filtered := make([]string, 0, len(arguments))
	structured := false
	help := false
	for _, argument := range arguments {
		switch argument {
		case "--json":
			structured = true
		case "--help", "-h":
			help = true
		default:
			filtered = append(filtered, argument)
		}
	}
	return filtered, structured, help
}

func writeUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Usage: sway-session [--json] <command> [options]")
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintln(writer, "Persist explicitly registered Sway work contexts.")
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintln(writer, "Commands:")
	for _, name := range commandOrder {
		spec := commandSpecs[name]
		_, _ = fmt.Fprintf(writer, "  %-20s %s\n", spec.usage, spec.summary)
	}
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintln(writer, "Options:")
	_, _ = fmt.Fprintln(writer, "  --json               Emit machine-readable diagnostics")
	_, _ = fmt.Fprintln(writer, "  -h, --help           Show help")
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintln(writer, "Exit status:")
	_, _ = fmt.Fprintln(writer, "  0  Help or successful operation")
	_, _ = fmt.Fprintln(writer, "  2  Invalid command or arguments")
	_, _ = fmt.Fprintln(writer, "  3  Recognized command unavailable in this foundation build")
}

func writeCommandHelp(name string, stdout io.Writer, stderr io.Writer, structured bool) int {
	spec, exists := commandSpecs[name]
	if !exists {
		writeDiagnostic(stderr, structured, "unknown_command", fmt.Sprintf("unknown command %q", name), "Run sway-session --help to see available commands.")
		return exitUsage
	}
	writeCommandUsage(stdout, name, spec)
	return exitSuccess
}

func writeCommandUsage(writer io.Writer, name string, spec commandSpec) {
	_, _ = fmt.Fprintf(writer, "Usage: sway-session [--json] %s\n\n%s.\n", spec.usage, spec.summary)
	if slices.Contains([]string{"archive", "activate", "purge", "restore"}, name) {
		_, _ = fmt.Fprintln(writer, "A context is an unambiguous UUID or label.")
	}
}

func writeDiagnostic(writer io.Writer, structured bool, code string, message string, hint string) {
	_ = diagnostic.Write(writer, "sway-session", diagnostic.Diagnostic{
		Level:   diagnostic.LevelError,
		Code:    strings.TrimSpace(code),
		Message: message,
		Hint:    hint,
	}, structured)
}
