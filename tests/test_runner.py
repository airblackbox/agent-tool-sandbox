"""Test SandboxRunner."""
import asyncio
import pytest
from pkg.executor.runner import SandboxRunner
from pkg.models.sandbox import ExecutionStatus, SandboxRequest, ResourceLimits

@pytest.mark.asyncio
async def test_execute_unknown_tool():
    """Test executing unknown tool."""
    runner = SandboxRunner()
    req = SandboxRequest(tool_name="unknown")
    result = await runner.execute(req)
    assert result.status == ExecutionStatus.DENIED
    assert "Unknown tool" in result.error

@pytest.mark.asyncio
async def test_execute_sync_tool():
    """Test executing sync tool."""
    runner = SandboxRunner()
    runner.register_tool("add", lambda a, b: a + b)
    req = SandboxRequest(tool_name="add", tool_input={"a": 5, "b": 3})
    result = await runner.execute(req)
    assert result.status == ExecutionStatus.SUCCESS
    assert result.output == 8

@pytest.mark.asyncio
async def test_execute_async_tool():
    """Test executing async tool."""
    runner = SandboxRunner()
    async def async_tool(x):
        await asyncio.sleep(0.01)
        return x * 2
    runner.register_tool("double", async_tool)
    req = SandboxRequest(tool_name="double", tool_input={"x": 5})
    result = await runner.execute(req)
    assert result.status == ExecutionStatus.SUCCESS
    assert result.output == 10

@pytest.mark.asyncio
async def test_execute_timeout():
    """Test timeout handling."""
    runner = SandboxRunner()
    async def slow_tool():
        await asyncio.sleep(10)
        return "done"
    runner.register_tool("slow", slow_tool)
    limits = ResourceLimits(max_duration_ms=100)
    req = SandboxRequest(tool_name="slow", limits=limits)
    result = await runner.execute(req)
    assert result.status == ExecutionStatus.TIMEOUT

@pytest.mark.asyncio
async def test_execute_exception():
    """Test exception handling."""
    runner = SandboxRunner()
    def bad_tool():
        raise ValueError("test error")
    runner.register_tool("bad", bad_tool)
    req = SandboxRequest(tool_name="bad")
    result = await runner.execute(req)
    assert result.status == ExecutionStatus.FAILED
    assert "test error" in result.error

@pytest.mark.asyncio
async def test_history_tracking():
    """Test execution history."""
    runner = SandboxRunner()
    runner.register_tool("test", lambda: "ok")
    for _ in range(5):
        req = SandboxRequest(tool_name="test")
        await runner.execute(req)
    history = runner.get_history()
    assert len(history) == 5
    assert all(r.status == ExecutionStatus.SUCCESS for r in history)
