// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package sandbox

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AllowNetwork {
		t.Fatal("default should deny network")
	}
	if cfg.MaxMemoryMB != 256 {
		t.Fatalf("expected 256MB, got %d", cfg.MaxMemoryMB)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected 30s timeout, got %v", cfg.Timeout)
	}
}

func TestDeniedPathViolation(t *testing.T) {
	cfg := DefaultConfig()
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "cat", []string{"/etc/passwd"})
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for /etc path")
	}
	if result.ExitCode != -1 {
		t.Fatalf("expected exit code -1, got %d", result.ExitCode)
	}
}

func TestPathTraversalViolation(t *testing.T) {
	cfg := DefaultConfig()
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "cat", []string{"../../etc/passwd"})
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for path traversal")
	}
}

func TestNetworkCommandBlocked(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowNetwork = false
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "curl", []string{"https://example.com"})
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for curl when network disabled")
	}
}

func TestValidCommand(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowedPaths = []string{"/"}
	cfg.DeniedPaths = nil
	cfg.WorkDir = ""
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "echo", []string{"hello"})
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d (error: %s)", result.ExitCode, result.Error)
	}
	if result.Stdout != "hello\n" {
		t.Fatalf("expected 'hello\\n', got '%s'", result.Stdout)
	}
}

func TestTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowedPaths = []string{"/"}
	cfg.DeniedPaths = nil
	cfg.Timeout = 100 * time.Millisecond
	cfg.WorkDir = ""
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "sleep", []string{"5"})
	if !result.Killed {
		t.Fatal("expected timeout kill")
	}
}

func TestValidatePath(t *testing.T) {
	cfg := DefaultConfig()
	runner := NewRunner(cfg)

	// Denied path
	if err := runner.ValidatePath("/etc/passwd"); err == nil {
		t.Fatal("expected error for /etc")
	}

	// Path traversal
	if err := runner.ValidatePath("/tmp/sandbox/../etc"); err == nil {
		t.Fatal("expected error for traversal")
	}

	// Allowed path
	if err := runner.ValidatePath("/tmp/sandbox/file.txt"); err != nil {
		t.Fatalf("expected no error for allowed path: %v", err)
	}

	// Not in allowed list
	if err := runner.ValidatePath("/opt/data/file.txt"); err == nil {
		t.Fatal("expected error for non-allowed path")
	}
}

func TestValidateHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowNetwork = true
	cfg.AllowedHosts = []string{"api.openai.com", "api.anthropic.com"}
	runner := NewRunner(cfg)

	if err := runner.ValidateHost("api.openai.com"); err != nil {
		t.Fatalf("expected allowed: %v", err)
	}
	if err := runner.ValidateHost("evil.com"); err == nil {
		t.Fatal("expected denied for evil.com")
	}

	// Network disabled
	cfg2 := DefaultConfig()
	runner2 := NewRunner(cfg2)
	if err := runner2.ValidateHost("anything"); err == nil {
		t.Fatal("expected denied when network disabled")
	}
}

func TestDangerousCommand(t *testing.T) {
	cfg := DefaultConfig()
	runner := NewRunner(cfg)
	result := runner.Execute(context.Background(), "rm", []string{"-rf", "/"})
	if len(result.Violations) == 0 {
		t.Fatal("expected violation for rm -rf /")
	}
}
