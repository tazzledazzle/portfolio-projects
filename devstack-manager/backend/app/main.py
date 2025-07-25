from fastapi import FastAPI, WebSocket, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import asyncio
from typing import List, Dict, Optional

app = FastAPI(title="DevStack Manager API", version="1.0.0")

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize services lazily
docker_service = None
vm_service = None
profile_manager = None
mock_service = None

def get_services():
    """Lazy initialization of services"""
    global docker_service, vm_service, profile_manager, mock_service
    
    if docker_service is None:
        try:
            from .services.docker_control import DockerService
            from .services.vm_control import VMService
            from .services.profile_manager import ProfileManager
            from .services.mock_service import MockService
            
            docker_service = DockerService()
            vm_service = VMService()
            profile_manager = ProfileManager()
            mock_service = MockService()
        except Exception as e:
            print(f"Error initializing services: {e}")
            # Create dummy services that return errors
            docker_service = DummyDockerService()
            vm_service = DummyVMService()
            profile_manager = DummyProfileManager()
            mock_service = DummyMockService()
    
    return docker_service, vm_service, profile_manager, mock_service

class DummyDockerService:
    def list_containers(self):
        return [{"error": "Docker service not available"}]
    
    def start_container(self, name):
        return {"status": "error", "message": "Docker service not available"}
    
    def stop_container(self, name):
        return {"status": "error", "message": "Docker service not available"}
    
    def get_container_status(self, name):
        return {"error": "Docker service not available"}
    
    def get_container_logs(self, name, lines=100):
        return ["Docker service not available"]

class DummyVMService:
    def list_vms(self):
        return []
    
    def start_vm(self, name):
        return {"status": "error", "message": "VM service not available"}
    
    def stop_vm(self, name):
        return {"status": "error", "message": "VM service not available"}
    
    def get_vm_status(self, name):
        return {"error": "VM service not available"}

class DummyProfileManager:
    def list_profiles(self):
        return []
    
    def get_profile(self, name):
        return None
    
    async def start_profile(self, name):
        return {"status": "error", "message": "Profile service not available"}
    
    async def stop_profile(self, name):
        return {"status": "error", "message": "Profile service not available"}

class DummyMockService:
    def list_services(self):
        return []
    
    async def start_service(self, name):
        return {"status": "error", "message": "Mock service not available"}
    
    async def stop_service(self, name):
        return {"status": "error", "message": "Mock service not available"}
    
    async def create_service(self, config):
        return {"status": "error", "message": "Mock service not available"}
    
    async def delete_service(self, name):
        return {"status": "error", "message": "Mock service not available"}

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "DevStack Manager API", "status": "running"}

@app.get("/api/health")
async def health_check():
    """Health check endpoint"""
    try:
        docker_service, vm_service, profile_manager, mock_service = get_services()
        
        # Test Docker connection
        docker_status = "ok"
        try:
            containers = docker_service.list_containers()
            if any("error" in str(container) for container in containers):
                docker_status = "unavailable"
        except:
            docker_status = "unavailable"
        
        return {
            "status": "ok", 
            "message": "DevStack Manager API is running",
            "services": {
                "docker": docker_status,
                "vm": "ok",
                "profiles": "ok",
                "mock": "ok"
            }
        }
    except Exception as e:
        return {
            "status": "ok", 
            "message": "DevStack Manager API is running",
            "services": {
                "docker": "unavailable",
                "vm": "unavailable", 
                "profiles": "unavailable",
                "mock": "unavailable"
            },
            "error": str(e)
        }

@app.get("/api/profiles")
async def get_profiles():
    """Get all available profiles"""
    try:
        _, _, profile_manager, _ = get_services()
        profiles = profile_manager.list_profiles()
        return {"profiles": profiles}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/profiles/{profile_name}")
async def get_profile(profile_name: str):
    """Get specific profile configuration"""
    try:
        _, _, profile_manager, _ = get_services()
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
        _, _, profile_manager, _ = get_services()
        result = await profile_manager.start_profile(profile_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/profiles/{profile_name}/stop")
async def stop_profile(profile_name: str):
    """Stop all services in a profile"""
    try:
        _, _, profile_manager, _ = get_services()
        result = await profile_manager.stop_profile(profile_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/services")
async def list_services():
    """List all services across all types"""
    try:
        docker_service, vm_service, _, mock_service = get_services()
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
        docker_service, vm_service, _, mock_service = get_services()
        
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
        docker_service, vm_service, _, mock_service = get_services()
        
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
        docker_service, _, _, _ = get_services()
        logs = docker_service.get_container_logs(service_name, lines)
        return {"logs": logs}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.websocket("/ws/logs/{service_name}")
async def websocket_logs(websocket: WebSocket, service_name: str):
    """WebSocket endpoint for streaming logs"""
    try:
        from .websocket.log_stream import stream_logs
        await stream_logs(websocket, service_name)
    except Exception as e:
        print(f"WebSocket error: {e}")
        try:
            await websocket.close()
        except:
            pass

@app.get("/api/services/{service_name}/status")
async def get_service_status(service_name: str):
    """Get detailed status of a service"""
    try:
        docker_service, _, _, _ = get_services()
        status = docker_service.get_container_status(service_name)
        return status
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/mock-services")
async def create_mock_service(config: dict):
    """Create a new mock service"""
    try:
        _, _, _, mock_service = get_services()
        result = await mock_service.create_service(config)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.delete("/api/mock-services/{service_name}")
async def delete_mock_service(service_name: str):
    """Delete a mock service"""
    try:
        _, _, _, mock_service = get_services()
        result = await mock_service.delete_service(service_name)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000, reload=True)