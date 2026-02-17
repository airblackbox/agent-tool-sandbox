"""CLI commands for sandbox."""
import click
from rich.console import Console
from rich.table import Table
import httpx
import json

console = Console()
BASE_URL = "http://localhost:8500/v1"

@click.group()
def cli():
    """Agent Tool Sandbox CLI."""
    pass

@cli.command()
def health():
    """Check sandbox health."""
    try:
        with httpx.Client() as client:
            resp = client.get(f"{BASE_URL}/health")
            resp.raise_for_status()
            data = resp.json()
            console.print("[green]✓[/green] Sandbox is healthy")
            console.print(f"  Tools registered: {data['tools_registered']}")
    except Exception as e:
        console.print(f"[red]✗[/red] Health check failed: {e}")

@cli.command()
def list_tools():
    """List registered tools."""
    try:
        with httpx.Client() as client:
            resp = client.get(f"{BASE_URL}/tools")
            resp.raise_for_status()
            tools = resp.json()["tools"]
            if tools:
                table = Table(title="Registered Tools")
                table.add_column("Tool Name")
                for tool in tools:
                    table.add_row(tool)
                console.print(table)
            else:
                console.print("No tools registered")
    except Exception as e:
        console.print(f"[red]Error:[/red] {e}")

@cli.command()
@click.option("--name", required=True, help="Tool name")
@click.option("--description", default="", help="Tool description")
def register(name: str, description: str):
    """Register a new tool."""
    try:
        with httpx.Client() as client:
            resp = client.post(
                f"{BASE_URL}/tools/register",
                json={"name": name, "description": description}
            )
            resp.raise_for_status()
            data = resp.json()
            console.print(
                f"[green]✓[/green] Registered tool '{name}' "
                f"({data['total_tools']} total)"
            )
    except Exception as e:
        console.print(f"[red]Error:[/red] {e}")

if __name__ == "__main__":
    cli()
