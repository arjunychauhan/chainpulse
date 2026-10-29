package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(stdout.String(), "Commands:") {
		t.Errorf("expected help output to contain Commands:, got %q", stdout.String())
	}

	if stderr.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunHelpAliases(t *testing.T) {
	for _, command := range []string{"--help", "-h"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := run([]string{command}, &stdout, &stderr)

			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}

			if !strings.Contains(stdout.String(), "ChainPulse - EVM RPC monitoring CLI") {
				t.Errorf("expected CLI help text, got %q", stdout.String())
			}
		})
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(stdout.String(), "ChainPulse dev") {
		t.Errorf("expected version output to contain ChainPulse dev, got %q", stdout.String())
	}
}

func TestRunMissingCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage instructions, got %q", stdout.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"unknown"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(stderr.String(), "Unknown command: unknown") {
		t.Errorf("expected unknown command error on stderr, got %q", stderr.String())
	}

	if !strings.Contains(stdout.String(), "Commands:") {
		t.Errorf("expected help text on stdout, got %q", stdout.String())
	}
}

func TestCheckHelp(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			code := checkWithRPC([]string{arg}, nil)

			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
		})
	}
}

func TestCheckUnknownFlag(t *testing.T) {
	code := checkWithRPC([]string{"--bad-flag"}, nil)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}
