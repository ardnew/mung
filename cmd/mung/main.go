package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/ardnew/mung/cmd/mung/run"
)

//go:embed VERSION
var version string

// Version returns the semantic version of the mung command-line tool.
func Version() string { return strings.TrimSpace(version) }

// exit centralizes process termination for easy testing/wrapping.
func exit(code run.ExitCode) {
	if code.Int() != 0 {
		// Print only the error message. Newline for cleanliness.
		if msg := code.Error(); strings.TrimSpace(msg) != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	}
	os.Exit(code.Int())
}

func main() {
	out, code := run.Main(Version())
	if code.Int() == 0 && strings.TrimSpace(out) != "" {
		fmt.Fprint(os.Stdout, out)
	}
	exit(code)
}
