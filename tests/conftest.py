"""Pytest configuration and fixtures."""
import pytest
from pkg.executor.runner import SandboxRunner
from pkg.limits.enforcer import LimitEnforcer
from pkg.models.sandbox import ResourceLimits

@pytest.fixture
def runner():
    """Create a test runner."""
    return SandboxRunner()

@pytest.fixture
def enforcer():
    """Create a test enforcer."""
    limits = ResourceLimits(max_duration_ms=5000)
    return LimitEnforcer(global_limits=limits)

@pytest.fixture
def sample_tool():
    """Provide a sample tool function."""
    def echo_tool(message: str = ""):
        return {"echo": message}
    return echo_tool
