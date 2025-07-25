from fastapi import FastAPI, WebSocket, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import asyncio
from typing import List, Dict, Optional

from .services.docker_control import DockerService
from .services.vm_control import VMService
from .services.profile_manager import ProfileManager
from .services.mock_service import MockService
from .websocket.log_stream import stream_logs
from .utils.yaml_loader import load_profiles

app = FastAPI(title="DevStack Manager API", version="1.0.0")

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize services
docker_service = DockerService()
vm_service = VMService()
profile_manager = ProfileManager()
mock_service = MockService()

@app.get("/api/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "ok", "message": "DevStack Manager API is running"}

@app.get("/api/profiles")
async def get_profiles():
    """Get all available profiles"""
    try:
        profiles = profile_manager.list_profiles()
        return {"profiles": profiles}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/profiles/{profile_name}")
async def get_profile(profile_name: str):
    """Get specific profile configuration"""
    try:
        profile = profile_manager.get_profile(profile_name)
        if not profile:
            raise HTTPException(status_code=404, detail="Profile not found")
        return profile
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/profiles/{profile_name}/start")
async def start_profile(profile_name: str):
    """Start all services in a profile"""
    try:
        result = await profile_manager.start_profile(profile_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/profiles/{profile_name}/stop")
async def stop_profile(profile_name: str):
    """Stop all services in a profile"""
    try:
        result = await profile_manager.stop_profile(profile_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/services")
async def list_services():
    """List all services across all types"""
    try:
        docker_containers = docker_service.list_containers()
        vms = vm_service.list_vms()
        mock_services = mock_service.list_services()
        
        return {
            "docker": docker_containers,
            "vms": vms,
            "mocks": mock_services
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/services/{service_name}/start")
async def start_service(service_name: str, service_type: str = "docker"):
    """Start a specific service"""
    try:
        if service_type == "docker":
            result = docker_service.start_container(service_name)
        elif service_type == "vm":
            result = vm_service.start_vm(service_name)
        elif service_type == "mock":
            result = await mock_service.start_service(service_name)
        else:
            raise HTTPException(status_code=400, detail="Invalid service type")
        
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/services/{service_name}/stop")
async def stop_service(service_name: str, service_type: str = "docker"):
    """Stop a specific service"""
    try:
        if service_type == "docker":
            result = docker_service.stop_container(service_name)
        elif service_type == "vm":
            result = vm_service.stop_vm(service_name)
        elif service_type == "mock":
            result = await mock_service.stop_service(service_name)
        else:
            raise HTTPException(status_code=400, detail="Invalid service type")
        
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/services/{service_name}/logs")
async def get_service_logs(service_name: str, lines: int = 100):
    """Get recent logs from a service"""
    try:
        logs = docker_service.get_container_logs(service_name, lines)
        return {"logs": logs}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.websocket("/ws/logs/{service_name}")
async def websocket_logs(websocket: WebSocket, service_name: str):
    """WebSocket endpoint for streaming logs"""
    await stream_logs(websocket, service_name)

@app.get("/api/services/{service_name}/status")
async def get_service_status(service_name: str):
    """Get detailed status of a service"""
    try:
        status = docker_service.get_container_status(service_name)
        return status
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/mock-services")
async def create_mock_service(config: dict):
    """Create a new mock service"""
    try:
        result = await mock_service.create_service(config)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.delete("/api/mock-services/{service_name}")
async def delete_mock_service(service_name: str):
    """Delete a mock service"""
    try:
        result = await mock_service.delete_service(service_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000, reload=True)