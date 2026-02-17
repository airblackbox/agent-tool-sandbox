"""Limit enforcement logic."""
from __future__ import annotations
from pkg.models.sandbox import ResourceLimits, SandboxRequest

class LimitEnforcer:
    """Enforces resource limits across sandbox requests."""
    def __init__(self, global_limits: ResourceLimits | None = None) -> None:
        self.global_limits = global_limits or ResourceLimits()
        self._tool_limits: dict[str, ResourceLimits] = {}

    def set_tool_limits(self, tool_name: str, limits: ResourceLimits) -> None:
        """Set limits for a specific tool."""
        self._tool_limits[tool_name] = limits

    def get_effective_limits(self, request: SandboxRequest) -> ResourceLimits:
        """Get effective limits by merging request, tool, and global limits."""
        tool_limits = self._tool_limits.get(request.tool_name)
        if tool_limits:
            return ResourceLimits(
                max_duration_ms=min(
                    request.limits.max_duration_ms,
                    tool_limits.max_duration_ms,
                    self.global_limits.max_duration_ms
                ),
                max_output_bytes=min(
                    request.limits.max_output_bytes,
                    tool_limits.max_output_bytes,
                    self.global_limits.max_output_bytes
                ),
                max_memory_mb=min(
                    request.limits.max_memory_mb,
                    tool_limits.max_memory_mb,
                    self.global_limits.max_memory_mb
                ),
                allow_network=(
                    request.limits.allow_network and
                    tool_limits.allow_network and
                    self.global_limits.allow_network
                ),
                allow_filesystem=(
                    request.limits.allow_filesystem and
                    tool_limits.allow_filesystem and
                    self.global_limits.allow_filesystem
                ),
            )
        return request.limits

    def check_allowed(self, request: SandboxRequest) -> tuple[bool, str]:
        """Check if request is allowed under limits."""
        limits = self.get_effective_limits(request)
        if request.limits.max_duration_ms > self.global_limits.max_duration_ms:
            return False, (
                f"Duration {request.limits.max_duration_ms}ms exceeds "
                f"global limit {self.global_limits.max_duration_ms}ms"
            )
        return True, "ok"
