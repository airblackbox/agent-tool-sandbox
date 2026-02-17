"""Sandbox executor runner."""
from __future__ import annotations
import asyncio
import time
import uuid
from typing import Any, Callable
from pkg.models.sandbox import ExecutionStatus, ResourceLimits, SandboxRequest, SandboxResult

class SandboxRunner:
    """Sandboxed tool execution runner."""
    def __init__(self) -> None:
        self._tools: dict[str, Callable] = {}
        self._history: list[SandboxResult] = []

    def register_tool(self, name: str, func: Callable) -> None:
        """Register a tool for execution."""
        self._tools[name] = func

    async def execute(self, request: SandboxRequest) -> SandboxResult:
        """Execute a tool in sandbox with limits."""
        if not request.request_id:
            request.request_id = f"sbx-{uuid.uuid4().hex[:12]}"
        result = SandboxResult(
            request_id=request.request_id,
            tool_name=request.tool_name,
            status=ExecutionStatus.RUNNING
        )
        func = self._tools.get(request.tool_name)
        if not func:
            result.status = ExecutionStatus.DENIED
            result.error = f"Unknown tool: {request.tool_name}"
            self._history.append(result)
            return result
        start = time.monotonic()
        try:
            output = await asyncio.wait_for(
                self._run_tool(func, request.tool_input),
                timeout=request.limits.max_duration_ms / 1000.0
            )
            result.output = output
            result.output_bytes = len(str(output).encode())
            if result.output_bytes > request.limits.max_output_bytes:
                result.status = ExecutionStatus.FAILED
                result.error = (
                    f"Output exceeds limit: "
                    f"{result.output_bytes} > {request.limits.max_output_bytes}"
                )
            else:
                result.status = ExecutionStatus.SUCCESS
        except asyncio.TimeoutError:
            result.status = ExecutionStatus.TIMEOUT
            result.error = f"Timeout after {request.limits.max_duration_ms}ms"
        except Exception as e:
            result.status = ExecutionStatus.FAILED
            result.error = str(e)
        result.duration_ms = (time.monotonic() - start) * 1000
        self._history.append(result)
        return result

    async def _run_tool(self, func: Callable, inputs: dict) -> Any:
        """Run tool, handling both sync and async functions."""
        if asyncio.iscoroutinefunction(func):
            return await func(**inputs)
        return func(**inputs)

    def get_history(self, limit: int = 100) -> list[SandboxResult]:
        """Get execution history."""
        return self._history[-limit:]
