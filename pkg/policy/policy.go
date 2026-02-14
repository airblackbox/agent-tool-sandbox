// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package policy defines sandbox profiles — named configurations
// that map tool names to sandbox constraint sets.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nostalgicskinco/agent-tool-sandbox/pkg/sandbox"
)

// Profile is a named sandbox configuration for a specific tool or category.
type Profile struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	ToolPatterns []string       `json:"tool_patterns"` // tool names this profile applies to
	Config       ProfileConfig  `json:"config"`
}

// ProfileConfig is a JSON-serializable version of sandbox.Config.
type ProfileConfig struct {
	AllowedPaths  []string `json:"allowed_paths,omitempty"`
	DeniedPaths   []string `json:"denied_paths,omitempty"`
	ReadOnly      bool     `json:"read_only,omitempty"`
	AllowNetwork  bool     `json:"allow_network"`
	AllowedHosts  []string `json:"allowed_hosts,omitempty"`
	MaxMemoryMB   int      `json:"max_memory_mb,omitempty"`
	MaxCPUSeconds float64  `json:"max_cpu_seconds,omitempty"`
	TimeoutSec    int      `json:"timeout_seconds,omitempty"`
}

// ToSandboxConfig converts a ProfileConfig to a sandbox.Config.
func (pc ProfileConfig) ToSandboxConfig() sandbox.Config {
	cfg := sandbox.DefaultConfig()
	if len(pc.AllowedPaths) > 0 {
		cfg.AllowedPaths = pc.AllowedPaths
	}
	if len(pc.DeniedPaths) > 0 {
		cfg.DeniedPaths = pc.DeniedPaths
	}
	cfg.ReadOnly = pc.ReadOnly
	cfg.AllowNetwork = pc.AllowNetwork
	if len(pc.AllowedHosts) > 0 {
		cfg.AllowedHosts = pc.AllowedHosts
	}
	if pc.MaxMemoryMB > 0 {
		cfg.MaxMemoryMB = pc.MaxMemoryMB
	}
	if pc.MaxCPUSeconds > 0 {
		cfg.MaxCPUSeconds = pc.MaxCPUSeconds
	}
	if pc.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(pc.TimeoutSec) * time.Second
	}
	return cfg
}

// ProfileSet is a collection of sandbox profiles.
type ProfileSet struct {
	Profiles []Profile `json:"profiles"`
	Default  string    `json:"default,omitempty"` // default profile name
}

// LoadProfileSet loads profiles from a JSON file.
func LoadProfileSet(path string) (*ProfileSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profiles: %w", err)
	}
	return ParseProfileSet(data)
}

// ParseProfileSet parses profiles from JSON.
func ParseProfileSet(data []byte) (*ProfileSet, error) {
	var ps ProfileSet
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}
	return &ps, nil
}

// FindProfile finds the best matching profile for a tool name.
// Exact matches take priority over wildcards.
func (ps *ProfileSet) FindProfile(toolName string) *Profile {
	// First pass: exact match (no wildcards)
	for i, p := range ps.Profiles {
		for _, pattern := range p.ToolPatterns {
			if pattern == toolName {
				return &ps.Profiles[i]
			}
		}
	}
	// Second pass: wildcard match
	for i, p := range ps.Profiles {
		for _, pattern := range p.ToolPatterns {
			if pattern == "*" {
				return &ps.Profiles[i]
			}
		}
	}
	// Fall back to default profile
	if ps.Default != "" {
		for i, p := range ps.Profiles {
			if p.Name == ps.Default {
				return &ps.Profiles[i]
			}
		}
	}
	return nil
}
