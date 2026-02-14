# agent-tool-sandbox

**Policy-driven agent tool sandbox** — run untrusted tool calls safely with filesystem/network isolation, resource limits, and OTel-compatible audit trails.

Named sandbox profiles map tools to constraint sets: filesystem allowlists, network policies, memory/CPU limits, and timeout enforcement. Every execution decision is logged as structured JSON.

> Part of the **GenAI Infrastructure Standard** — a composable suite of open-source tools for enterprise-grade GenAI observability, security, and governance.
>
> | Layer | Component | Repo |
> |-------|-----------|------|
> | Privacy | Prompt Vault Processor | [prompt-vault-processor](https://github.com/nostalgicskinco/prompt-vault-processor) |
> | Normalization | Semantic Normalizer | [genai-semantic-normalizer](https://github.com/nostalgicskinco/genai-semantic-normalizer) |
> | Metrics | Cost & SLO Pack | [genai-cost-slo](https://github.com/nostalgicskinco/genai-cost-slo) |
> | Replay | Agent VCR | [agent-vcr](https://github.com/nostalgicskinco/agent-vcr) |
> | Testing | Regression Harness | [trace-regression-harness](https://github.com/nostalgicskinco/trace-regression-harness) |
> | Security | MCP Scanner | [mcp-security-scanner](https://github.com/nostalgicskinco/mcp-security-scanner) |
> | Gateway | MCP Policy Gateway | [mcp-policy-gateway](https://github.com/nostalgicskinco/mcp-policy-gateway) |
> | Inventory | Runtime AIBOM Emitter | [runtime-aibom-emitter](https://github.com/nostalgicskinco/runtime-aibom-emitter) |
> | Policy | AIBOM Policy Engine | [aibom-policy-engine](https://github.com/nostalgicskinco/aibom-policy-engine) |
> | **Sandbox** | **Agent Tool Sandbox** | **this repo** |

## Quick Start

```bash
go build -o toolbox ./cmd/toolbox

# Run with default restrictive sandbox
./toolbox echo "hello from sandbox"

# Run with profiles
./toolbox -profiles sandbox-profiles.json -tool web_search curl https://api.google.com
```

## Sandbox Constraints

| Constraint | Description |
|-----------|-------------|
| Filesystem allowlist | Only specified paths accessible |
| Filesystem denylist | Critical paths always blocked |
| Network toggle | Enable/disable all network access |
| Host allowlist | Restrict outbound connections by DNS |
| Memory limit | Max RSS in MB |
| CPU limit | Max CPU seconds |
| Timeout | Hard kill after duration |
| Path traversal | Automatic `..` detection |
| Dangerous commands | Block `rm -rf /`, `mkfs`, fork bombs |

## License

AGPL-3.0-or-later — see [LICENSE](LICENSE). Commercial licenses available.
