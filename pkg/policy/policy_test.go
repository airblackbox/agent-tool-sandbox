// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"testing"
)

const sampleProfiles = `{
	"default": "restrictive",
	"profiles": [
		{
			"name": "restrictive",
			"description": "Default restrictive sandbox",
			"tool_patterns": ["*"],
			"config": {
				"allowed_paths": ["/tmp/sandbox"],
				"denied_paths": ["/etc", "/var"],
				"allow_network": false,
				"max_memory_mb": 128,
				"timeout_seconds": 10
			}
		},
		{
			"name": "web-tools",
			"description": "Tools that need network access",
			"tool_patterns": ["web_search", "url_fetch"],
			"config": {
				"allowed_paths": ["/tmp/sandbox"],
				"allow_network": true,
				"allowed_hosts": ["api.google.com", "api.bing.com"],
				"max_memory_mb": 256,
				"timeout_seconds": 30
			}
		},
		{
			"name": "file-tools",
			"description": "File operation tools",
			"tool_patterns": ["read_file", "write_file", "list_files"],
			"config": {
				"allowed_paths": ["/tmp/sandbox", "/home/agent/data"],
				"denied_paths": ["/etc", "/var", "/root"],
				"allow_network": false,
				"timeout_seconds": 15
			}
		}
	]
}`

func TestParseProfileSet(t *testing.T) {
	ps, err := ParseProfileSet([]byte(sampleProfiles))
	if err != nil {
		t.Fatalf("ParseProfileSet: %v", err)
	}
	if len(ps.Profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(ps.Profiles))
	}
	if ps.Default != "restrictive" {
		t.Fatalf("expected default 'restrictive', got '%s'", ps.Default)
	}
}

func TestFindProfileExactMatch(t *testing.T) {
	ps, _ := ParseProfileSet([]byte(sampleProfiles))

	p := ps.FindProfile("web_search")
	if p == nil {
		t.Fatal("expected to find profile for web_search")
	}
	if p.Name != "web-tools" {
		t.Fatalf("expected web-tools profile, got %s", p.Name)
	}
}

func TestFindProfileDefault(t *testing.T) {
	ps, _ := ParseProfileSet([]byte(sampleProfiles))

	p := ps.FindProfile("unknown_tool")
	if p == nil {
		t.Fatal("expected default profile for unknown tool")
	}
	if p.Name != "restrictive" {
		t.Fatalf("expected restrictive default, got %s", p.Name)
	}
}

func TestToSandboxConfig(t *testing.T) {
	ps, _ := ParseProfileSet([]byte(sampleProfiles))
	p := ps.FindProfile("web_search")
	cfg := p.Config.ToSandboxConfig()

	if !cfg.AllowNetwork {
		t.Fatal("expected network allowed for web-tools")
	}
	if cfg.MaxMemoryMB != 256 {
		t.Fatalf("expected 256MB, got %d", cfg.MaxMemoryMB)
	}
	if len(cfg.AllowedHosts) != 2 {
		t.Fatalf("expected 2 allowed hosts, got %d", len(cfg.AllowedHosts))
	}
}

func TestFileToolsProfile(t *testing.T) {
	ps, _ := ParseProfileSet([]byte(sampleProfiles))
	p := ps.FindProfile("read_file")
	if p == nil || p.Name != "file-tools" {
		t.Fatal("expected file-tools profile for read_file")
	}
	cfg := p.Config.ToSandboxConfig()
	if cfg.AllowNetwork {
		t.Fatal("file-tools should not allow network")
	}
	if len(cfg.AllowedPaths) != 2 {
		t.Fatalf("expected 2 allowed paths, got %d", len(cfg.AllowedPaths))
	}
}
