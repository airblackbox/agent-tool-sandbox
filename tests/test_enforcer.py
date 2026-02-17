"""Test LimitEnforcer."""
import pytest
from pkg.limits.enforcer import LimitEnforcer
from pkg.models.sandbox import ResourceLimits, SandboxRequest

def test_enforcer_defaults():
    """Test enforcer with defaults."""
    enforcer = LimitEnforcer()
    assert enforcer.global_limits.max_duration_ms == 30000

def test_set_tool_limits():
    """Test setting tool-specific limits."""
    enforcer = LimitEnforcer()
    limits = ResourceLimits(max_duration_ms=5000)
    enforcer.set_tool_limits("test_tool", limits)
    assert "test_tool" in enforcer._tool_limits

def test_effective_limits_no_tool_limits():
    """Test effective limits without tool limits."""
    enforcer = LimitEnforcer()
    req = SandboxRequest(tool_name="unknown")
    effective = enforcer.get_effective_limits(req)
    assert effective == req.limits

def test_effective_limits_with_tool_limits():
    """Test effective limits merging."""
    global_limits = ResourceLimits(max_duration_ms=10000)
    enforcer = LimitEnforcer(global_limits=global_limits)
    tool_limits = ResourceLimits(max_duration_ms=5000)
    enforcer.set_tool_limits("test", tool_limits)
    req = SandboxRequest(tool_name="test")
    effective = enforcer.get_effective_limits(req)
    assert effective.max_duration_ms == 5000

def test_check_allowed_valid():
    """Test checking allowed request."""
    enforcer = LimitEnforcer()
    req = SandboxRequest(tool_name="test")
    allowed, msg = enforcer.check_allowed(req)
    assert allowed is True
    assert msg == "ok"

def test_check_allowed_duration_exceeded():
    """Test request exceeding global duration."""
    global_limits = ResourceLimits(max_duration_ms=5000)
    enforcer = LimitEnforcer(global_limits=global_limits)
    req = SandboxRequest(
        tool_name="test",
        limits=ResourceLimits(max_duration_ms=10000)
    )
    allowed, msg = enforcer.check_allowed(req)
    assert allowed is False
    assert "exceeds global limit" in msg
