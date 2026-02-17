"""Sandbox request/result models."""
from __future__ import annotations
from enum import Enum
from typing import Any
from pydantic import BaseModel, Field

class ExecutionStatus(str, Enum):
    """Execution status enumeration."""
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    TIMEOUT = "timeout"
    DENIED = "denied"

class ResourceLimits(BaseModel):
    """Resource limits configuration."""
    max_duration_ms: int = 30000
    max_output_bytes: int = 1_000_000
    max_memory_mb: int = 512
    allow_network: bool = False
    allow_filesystem: bool = False
    allowed_paths: list[str] = Field(default_factory=list)

class SandboxRequest(BaseModel):
    """Sandbox execution request."""
    request_id: str = ""
    agent_id: str = ""
    tool_name: str
    tool_input: dict[str, Any] = Field(default_factory=dict)
    limits: ResourceLimits = Field(default_factory=ResourceLimits)

class SandboxResult(BaseModel):
    """Sandbox execution result."""
    request_id: str = ""
    status: ExecutionStatus = ExecutionStatus.PENDING
    tool_name: str = ""
    output: Any = None
    error: str | None = None
    duration_ms: float = 0.0
    output_bytes: int = 0
    metadata: dict[str, Any] = Field(default_factory=dict)
