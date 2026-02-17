"""FastAPI server entry point."""
import uvicorn
from pkg.api.routes import router

def main():
    """Run the server."""
    uvicorn.run(
        router,
        host="0.0.0.0",
        port=8500,
        log_level="info"
    )

if __name__ == "__main__":
    main()
