"""Test FastAPI routes."""
import pytest
from fastapi.testclient import TestClient
from pkg.api.routes import router

client = TestClient(router)

def test_health():
    """Test health endpoint."""
    response = client.get("/v1/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["service"] == "agent-tool-sandbox"

def test_list_tools_empty():
    """Test listing tools when none registered."""
    response = client.get("/v1/tools")
    assert response.status_code == 200
    tools = response.json()["tools"]
    assert isinstance(tools, list)

def test_register_tool():
    """Test tool registration."""
    response = client.post(
        "/v1/tools/register",
        json={"name": "my_tool", "description": "test tool"}
    )
    assert response.status_code == 200
    data = response.json()
    assert data["registered"] is True
    assert data["tool_name"] == "my_tool"

def test_execute_unknown_tool():
    """Test executing unknown tool."""
    response = client.post(
        "/v1/execute",
        json={
            "tool_name": "unknown",
            "tool_input": {}
        }
    )
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "denied"

def test_get_history():
    """Test getting execution history."""
    response = client.get("/v1/history")
    assert response.status_code == 200
    data = response.json()
    assert "history" in data
    assert isinstance(data["history"], list)
