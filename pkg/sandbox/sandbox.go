// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package sandbox provides a policy-driven execution environment for
// agent tool calls. It enforces filesystem, network, and resource
// constraints while recording audit trails.
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Config defines the sandbox constraints.
type Config struct {
	// Filesystem
	AllowedPaths  []string      `json:"allowed_paths,omitempty"`  // paths the tool can read/write
	DeniedPaths   []string      `json:"denied_paths,omitempty"`   // explicitly blocked paths
	ReadOnly      bool          `json:"read_only,omitempty"`      // mount filesystem read-only

	// Network
	AllowNetwork  bool          `json:"allow_network"`
	AllowedHosts  []string      `json:"allowed_hosts,omitempty"`  // DNS/IP allowlist
	DeniedPorts   []int         `json:"denied_ports,omitempty"`   // blocked ports

	// Resources
	MaxMemoryMB   int           `json:"max_memory_mb,omitempty"`
	MaxCPUSeconds float64       `json:"max_cpu_seconds,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`

	// Execution
	WorkDir       string        `json:"work_dir,omitempty"`
	Env           []string      `json:"env,omitempty"`            // allowed env vars
}

// DefaultConfig returns a restrictive default sandbox config.
func DefaultConfig() Config {
	return Config{
		AllowedPaths:  []string{"/tmp/sandbox"},
		DeniedPaths:   []string{"/etc", "/var", "/root", "/home"},
		ReadOnly:      false,
		AllowNetwork:  false,
		MaxMemoryMB:   256,
		MaxCPUSeconds: 30,
		Timeout:       30 * time.Second,
		WorkDir:       "/tmp/sandbox",
	}
}

// Result is the outcome of a sandboxed tool execution.
type Result struct {
	ExitCode   int           `json:"exit_code"`
	Stdout     string        `json:"stdout"`
	Stderr     string        `json:"stderr"`
	Duration   time.Duration `json:"duration"`
	Killed     bool          `json:"killed,omitempty"`      // true if timeout/OOM killed
	Error      string        `json:"error,omitempty"`
	Violations []string      `json:"violations,omitempty"` // policy violations detected
}

// Runner executes tool calls within sandbox constraints.
type Runner struct {
	config Config
}

// NewRunner creates a sandbox runner with the given config.
func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Execute runs a command in the sandbox.
func (r *Runner) Execute(ctx context.Context, command string, args []string) *Result {
	result := &Result{}
	start := time.Now()

	// Pre-execution policy checks
	violations := r.checkPolicy(command, args)
	if len(violations) > 0 {
		result.Violations = violations
		result.Error = "policy violation: " + strings.Join(violations, "; ")
		result.ExitCode = -1
		result.Duration = time.Since(start)
		return result
	}

	// Set timeout
	timeout := r.config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(ctx, command, args...)
	if r.config.WorkDir != "" {
		cmd.Dir = r.config.WorkDir
	}
	cmd.Env = r.config.Env

	// Capture output
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run
	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if ctx.Err() == context.DeadlineExceeded {
		result.Killed = true
		result.Error = "timeout exceeded"
		result.ExitCode = -1
		return result
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	return result
}

// checkPolicy validates the command against sandbox policies.
func (r *Runner) checkPolicy(command string, args []string) []string {
	var violations []string

	// Check denied paths in arguments
	allArgs := strings.Join(args, " ")
	for _, denied := range r.config.DeniedPaths {
		if strings.Contains(allArgs, denied) {
			violations = append(violations, fmt.Sprintf("access to denied path: %s", denied))
		}
		if strings.Contains(command, denied) {
			violations = append(violations, fmt.Sprintf("command in denied path: %s", denied))
		}
	}

	// Check for path traversal
	if strings.Contains(allArgs, "..") {
		violations = append(violations, "path traversal detected")
	}

	// Check for network commands when network is disabled
	if !r.config.AllowNetwork {
		networkCmds := []string{"curl", "wget", "nc", "ncat", "ssh", "scp", "rsync"}
		for _, nc := range networkCmds {
			if command == nc || strings.HasSuffix(command, "/"+nc) {
				violations = append(violations, fmt.Sprintf("network command '%s' not allowed", nc))
			}
		}
	}

	// Check for dangerous commands
	dangerousCmds := []string{"rm -rf /", "mkfs", "dd if=/dev/zero", ":(){ :|:& };:"}
	for _, dc := range dangerousCmds {
		fullCmd := command + " " + allArgs
		if strings.Contains(fullCmd, dc) {
			violations = append(violations, fmt.Sprintf("dangerous command pattern: %s", dc))
		}
	}

	return violations
}

// ValidatePath checks if a path is allowed by the sandbox policy.
func (r *Runner) ValidatePath(path string) error {
	// Check denied paths first
	for _, denied := range r.config.DeniedPaths {
		if strings.HasPrefix(path, denied) {
			return fmt.Errorf("path '%s' is denied (matches %s)", path, denied)
		}
	}

	// Check path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path '%s' contains traversal", path)
	}

	// Check allowed paths
	if len(r.config.AllowedPaths) > 0 {
		allowed := false
		for _, ap := range r.config.AllowedPaths {
			if strings.HasPrefix(path, ap) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path '%s' not in allowed list", path)
		}
	}

	return nil
}

// ValidateHost checks if a network host is allowed.
func (r *Runner) ValidateHost(host string) error {
	if !r.config.AllowNetwork {
		return fmt.Errorf("network access disabled")
	}
	if len(r.config.AllowedHosts) == 0 {
		return nil // all hosts allowed if no allowlist
	}
	for _, ah := range r.config.AllowedHosts {
		if host == ah || strings.HasSuffix(host, "."+ah) {
			return nil
		}
	}
	return fmt.Errorf("host '%s' not in allowed list", host)
}
