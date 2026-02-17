"""Test sandbox models."""
import pytest
from pkg.models.sandbox import (
    ExecutionStatus,
    ResourceLimits,
    SandboxRequest,
    SandboxResult,
)

def test_execution_status():
    """Test ExecutionStatus enum."""
    assert ExecutionStatus.PENDING.value == "pending"
    assert ExecutionStatus.SUCCESS.value == "success"

def test_resource_limits_defaults():
    """Test ResourceLimits defaults."""
    limits = ResourceLimits()
    assert limits.max_duration_ms == 30000
    assert limits.max_output_bytes == 1_000_000
    assert limits.allow_network is False
    assert limits.allow_filesystem is False

def test_resource_limits_custom():
    """Test custom ResourceLimits."""
    limits = ResourceLimits(
        max_duration_ms=5000,
        allow_network=True,
        allowed_paths=["/tmp"]
    )
    assert limits.max_duration_ms == 5000
    assert limits.allow_network is True
    assert limits.allowed_paths == ["/tmp"]

def test_sandbox_request():
    """Test SandboxRequest creation."""
    req = SandboxRequest(tool_name="test_tool")
    assert req.tool_name == "test_tool"
    assert req.request_id == ""
    assert req.agent_id == ""

def test_sandbox_result_defaults():
    """Test SandboxResult defaults."""
    result = SandboxResult(tool_name="test")
    assert result.status == ExecutionStatus.PENDING
    assert result.output is None
    assert result.error is None
    assert result.duration_ms == 0.0
