// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command toolbox runs tool calls through the policy-driven sandbox.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/nostalgicskinco/agent-tool-sandbox/pkg/audit"
	"github.com/nostalgicskinco/agent-tool-sandbox/pkg/policy"
	"github.com/nostalgicskinco/agent-tool-sandbox/pkg/sandbox"
)

func main() {
	profileFile := flag.String("profiles", "", "Path to sandbox profiles JSON")
	toolName := flag.String("tool", "", "Tool name (for profile matching)")
	format := flag.String("format", "text", "Output format: text or json")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: toolbox [-profiles <file>] [-tool <name>] <command> [args...]\n")
		os.Exit(1)
	}

	command := args[0]
	cmdArgs := args[1:]

	// Load profile
	var cfg sandbox.Config
	profileName := "default"
	if *profileFile != "" {
		ps, err := policy.LoadProfileSet(*profileFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profiles: %v\n", err)
			os.Exit(1)
		}
		name := *toolName
		if name == "" {
			name = command
		}
		p := ps.FindProfile(name)
		if p != nil {
			cfg = p.Config.ToSandboxConfig()
			profileName = p.Name
		} else {
			cfg = sandbox.DefaultConfig()
		}
	} else {
		cfg = sandbox.DefaultConfig()
	}

	// Setup audit
	auditor := audit.NewLogger(os.Stderr)
	auditor.LogStart(*toolName, command, profileName)

	// Execute
	runner := sandbox.NewRunner(cfg)
	result := runner.Execute(context.Background(), command, cmdArgs)

	auditor.LogResult(*toolName, command, profileName, result)

	switch *format {
	case "json":
		json.NewEncoder(os.Stdout).Encode(result)
	default:
		if result.Error != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		}
		if len(result.Violations) > 0 {
			fmt.Fprintf(os.Stderr, "Policy violations:\n")
			for _, v := range result.Violations {
				fmt.Fprintf(os.Stderr, "  - %s\n", v)
			}
		}
		if result.Stdout != "" {
			fmt.Print(result.Stdout)
		}
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
}
