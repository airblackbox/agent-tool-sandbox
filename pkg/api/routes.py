"""FastAPI routes for sandbox."""
from __future__ import annotations
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from pkg.models.sandbox import SandboxRequest, SandboxResult
from pkg.executor.runner import SandboxRunner
from pkg.limits.enforcer import LimitEnforcer

router = FastAPI(title="Agent Tool Sandbox")
runner = SandboxRunner()
enforcer = LimitEnforcer()

class ToolInput(BaseModel):
    """Tool registration input."""
    name: str
    description: str = ""

@router.get("/v1/health")
async def health():
    """Health check endpoint."""
    return {
        "status": "ok",
        "service": "agent-tool-sandbox",
        "tools_registered": len(runner._tools)
    }

@router.post("/v1/execute")
async def execute(request: SandboxRequest) -> SandboxResult:
    """Execute a tool in sandbox."""
    allowed, msg = enforcer.check_allowed(request)
    if not allowed:
        raise HTTPException(status_code=400, detail=msg)
    result = await runner.execute(request)
    return result

@router.get("/v1/history")
async def get_history(limit: int = 100):
    """Get execution history."""
    return {"history": runner.get_history(limit)}

@router.post("/v1/tools/register")
async def register_tool(tool: ToolInput):
    """Register a tool (in-memory)."""
    def default_impl(**kwargs):
        return {"echo": kwargs}
    runner.register_tool(tool.name, default_impl)
    return {
        "registered": True,
        "tool_name": tool.name,
        "total_tools": len(runner._tools)
    }

@router.get("/v1/tools")
async def list_tools():
    """List registered tools."""
    return {"tools": list(runner._tools.keys())}
