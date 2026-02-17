# Agent Tool Sandbox

An isolated execution environment for AI agent tool calls with resource limits, timeouts, output capture, and rollback support.

## Features

- Sandboxed tool execution with async/sync support
- Resource limits: duration, output size, memory
- Network and filesystem access control
- Execution history and metadata tracking
- RESTful API with FastAPI
- CLI for tool management

## Quick Start

```bash
pip install -e .
python -m app.server
```

API runs on `http://localhost:8500/v1`

## API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/v1/health` | Health check |
| POST | `/v1/execute` | Execute sandboxed tool |
| GET | `/v1/history` | Get execution history |
| GET | `/v1/tools` | List registered tools |
| POST | `/v1/tools/register` | Register new tool |

## Configuration

Resource limits in `ResourceLimits` model:
- `max_duration_ms`: Default 30000ms
- `max_output_bytes`: Default 1MB
- `max_memory_mb`: Default 512MB
- `allow_network`: Default false
- `allow_filesystem`: Default false

## Testing

```bash
pytest tests/ -v
```

## License

MIT
