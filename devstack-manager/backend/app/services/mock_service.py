import asyncio
import json
import aiohttp
from aiohttp import web
from typing import Dict, List, Optional
import threading
import time

class MockService:
    def __init__(self):
        self.running_services = {}
        self.service_configs = {}
    
    def list_services(self) -> List[Dict]:
        """List all mock services"""
        services = []
        for name, info in self.running_services.items():
            services.append({
                "name": name,
                "status": "running" if info.get("running") else "stopped",
                "port": info.get("port"),
                "routes": list(self.service_configs.get(name, {}).get("routes", {}).keys())
            })
        return services
    
    async def start_service(self, service_name: str) -> Dict:
        """Start a mock service"""
        if service_name in self.running_services and self.running_services[service_name].get("running"):
            return {
                "status": "already_running",
                "message": f"Mock service {service_name} is already running"
            }
        
        config = self.service_configs.get(service_name)
        if not config:
            return {
                "status": "error",
                "message": f"No configuration found for service {service_name}"
            }
        
        try:
            port = config.get("port", 9000)
            app = self._create_mock_app(config)
            
            # Start the server in a separate thread
            server_thread = threading.Thread(
                target=self._run_server,
                args=(app, port, service_name),
                daemon=True
            )
            server_thread.start()
            
            # Give the server a moment to start
            await asyncio.sleep(0.5)
            
            self.running_services[service_name] = {
                "running": True,
                "port": port,
                "thread": server_thread
            }
            
            return {
                "status": "started",
                "name": service_name,
                "port": port,
                "message": f"Mock service {service_name} started on port {port}"
            }
        except Exception as e:
            return {
                "status": "error",
                "message": f"Failed to start mock service {service_name}: {str(e)}"
            }
    
    async def start_service_from_config(self, service_config: Dict) -> Dict:
        """Start a mock service from configuration"""
        service_name = service_config.get("name")
        mock_config = service_config.get("mock_config", {})
        
        # Extract port from service ports configuration
        ports = service_config.get("ports", [])
        port = 9000  # default
        if ports:
            port_str = ports[0]
            if isinstance(port_str, str) and ":" in port_str:
                port = int(port_str.split(":")[0])
            elif isinstance(port_str, int):
                port = port_str
        
        config = {
            "port": port,
            "routes": mock_config.get("routes", {})
        }
        
        self.service_configs[service_name] = config
        return await self.start_service(service_name)
    
    async def stop_service(self, service_name: str) -> Dict:
        """Stop a mock service"""
        if service_name not in self.running_services:
            return {
                "status": "not_found",
                "message": f"Mock service {service_name} not found"
            }
        
        service_info = self.running_services[service_name]
        service_info["running"] = False
        
        return {
            "status": "stopped",
            "name": service_name,
            "message": f"Mock service {service_name} stopped"
        }
    
    async def create_service(self, config: Dict) -> Dict:
        """Create a new mock service"""
        service_name = config.get("name")
        if not service_name:
            return {
                "status": "error",
                "message": "Service name is required"
            }
        
        self.service_configs[service_name] = config
        return {
            "status": "created",
            "name": service_name,
            "message": f"Mock service {service_name} configuration created"
        }
    
    async def delete_service(self, service_name: str) -> Dict:
        """Delete a mock service"""
        # Stop the service if running
        if service_name in self.running_services:
            await self.stop_service(service_name)
            del self.running_services[service_name]
        
        # Remove configuration
        if service_name in self.service_configs:
            del self.service_configs[service_name]
        
        return {
            "status": "deleted",
            "name": service_name,
            "message": f"Mock service {service_name} deleted"
        }
    
    def get_service_status(self, service_name: str) -> Dict:
        """Get status of a mock service"""
        if service_name not in self.running_services:
            return {
                "status": "not_found",
                "message": f"Mock service {service_name} not found"
            }
        
        service_info = self.running_services[service_name]
        return {
            "name": service_name,
            "status": "running" if service_info.get("running") else "stopped",
            "port": service_info.get("port"),
            "type": "mock"
        }
    
    def _create_mock_app(self, config: Dict) -> web.Application:
        """Create an aiohttp application for the mock service"""
        app = web.Application()
        routes = config.get("routes", {})
        
        for route_path, route_config in routes.items():
            method = route_config.get("method", "GET").upper()
            status = route_config.get("status", 200)
            headers = route_config.get("headers", {})
            body = route_config.get("body", {})
            
            # Create handler function
            async def handler(request, response_body=body, response_status=status, response_headers=headers):
                return web.json_response(
                    response_body,
                    status=response_status,
                    headers=response_headers
                )
            
            # Add route based on method
            if method == "GET":
                app.router.add_get(route_path, handler)
            elif method == "POST":
                app.router.add_post(route_path, handler)
            elif method == "PUT":
                app.router.add_put(route_path, handler)
            elif method == "DELETE":
                app.router.add_delete(route_path, handler)
            elif method == "PATCH":
                app.router.add_patch(route_path, handler)
        
        # Add a default health check endpoint
        async def health_check(request):
            return web.json_response({"status": "ok", "service": "mock"})
        
        app.router.add_get("/health", health_check)
        
        return app
    
    def _run_server(self, app: web.Application, port: int, service_name: str):
        """Run the aiohttp server"""
        try:
            web.run_app(app, host='0.0.0.0', port=port, print=None)
        except Exception as e:
            print(f"Error running mock service {service_name}: {e}")
            if service_name in self.running_services:
                self.running_services[service_name]["running"] = False