#!/usr/bin/env python3
"""
Startup script for DevStack Manager backend
"""
import os
import sys
import uvicorn

def main():
    """Main entry point"""
    print("Starting DevStack Manager Backend...")
    
    # Check if we're in development mode
    is_dev = os.getenv("ENVIRONMENT", "development") == "development"
    
    # Configure uvicorn
    config = {
        "app": "app.main:app",
        "host": "0.0.0.0",
        "port": 8000,
        "reload": is_dev,
        "log_level": "info"
    }
    
    try:
        uvicorn.run(**config)
    except Exception as e:
        print(f"Failed to start server: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()