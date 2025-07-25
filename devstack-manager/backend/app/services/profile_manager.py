import os
import yaml
import asyncio
from typing import List, Dict, Optional
from ..utils.yaml_loader import load_profiles
from .docker_control import DockerService
from .vm_control import VMService
from .mock_service import MockService

class ProfileManager:
    def __init__(self):
        self.docker_service = DockerService()
        self.vm_service = VMService()
        self.mock_service = MockService()
        self.config_dir = os.path.join(os.path.dirname(__file__), '../../configs')
    
    def list_profiles(self) -> List[Dict]:
        """List all available profiles"""
        profiles = []
        
        if not os.path.exists(self.config_dir):
            return profiles
        
        for filename in os.listdir(self.config_dir):
            if filename.endswith('.yaml') or filename.endswith('.yml'):
                try:
                    filepath = os.path.join(self.config_dir, filename)
                    with open(filepath, 'r') as f:
                        config = yaml.safe_load(f)
                        
                    profile_info = config.get('profile', {})
                    profile_info['filename'] = filename
                    profile_info['services_count'] = len(config.get('services', []))
                    profiles.append(profile_info)
                except Exception as e:
                    print(f"Error loading profile {filename}: {e}")
        
        return profiles
    
    def get_profile(self, profile_name: str) -> Optional[Dict]:
        """Get a specific profile configuration"""
        try:
            filepath = os.path.join(self.config_dir, f"{profile_name}.yaml")
            if not os.path.exists(filepath):
                # Try .yml extension
                filepath = os.path.join(self.config_dir, f"{profile_name}.yml")
                if not os.path.exists(filepath):
                    return None
            
            with open(filepath, 'r') as f:
                return yaml.safe_load(f)
        except Exception as e:
            print(f"Error loading profile {profile_name}: {e}")
            return None
    
    async def start_profile(self, profile_name: str) -> Dict:
        """Start all services in a profile"""
        profile = self.get_profile(profile_name)
        if not profile:
            return {"status": "error", "message": f"Profile {profile_name} not found"}
        
        results = []
        
        # Run pre-start hooks
        await self._run_hooks(profile.get('hooks', {}).get('pre_start', []))
        
        # Start services
        for service in profile.get('services', []):
            try:
                result = await self._start_service(service)
                results.append(result)
            except Exception as e:
                results.append({
                    "service": service.get('name'),
                    "status": "error",
                    "message": str(e)
                })
        
        # Run post-start hooks
        await self._run_hooks(profile.get('hooks', {}).get('post_start', []))
        
        return {
            "status": "completed",
            "profile": profile_name,
            "results": results
        }
    
    async def stop_profile(self, profile_name: str) -> Dict:
        """Stop all services in a profile"""
        profile = self.get_profile(profile_name)
        if not profile:
            return {"status": "error", "message": f"Profile {profile_name} not found"}
        
        results = []
        
        # Run pre-stop hooks
        await self._run_hooks(profile.get('hooks', {}).get('pre_stop', []))
        
        # Stop services in reverse order
        services = profile.get('services', [])
        for service in reversed(services):
            try:
                result = await self._stop_service(service)
                results.append(result)
            except Exception as e:
                results.append({
                    "service": service.get('name'),
                    "status": "error",
                    "message": str(e)
                })
        
        # Run post-stop hooks
        await self._run_hooks(profile.get('hooks', {}).get('post_stop', []))
        
        return {
            "status": "completed",
            "profile": profile_name,
            "results": results
        }
    
    async def _start_service(self, service: Dict) -> Dict:
        """Start a single service based on its type"""
        service_type = service.get('type')
        service_name = service.get('name')
        
        if service_type == 'docker':
            # Create container if it doesn't exist
            existing_containers = self.docker_service.list_containers()
            container_exists = any(c.get('name') == service_name for c in existing_containers)
            
            if not container_exists:
                create_result = self.docker_service.create_container_from_config(service)
                if create_result.get('status') == 'error':
                    return create_result
            
            return self.docker_service.start_container(service_name)
        
        elif service_type == 'vm':
            return self.vm_service.start_vm(service_name)
        
        elif service_type == 'mock':
            return await self.mock_service.start_service_from_config(service)
        
        else:
            return {
                "status": "error",
                "message": f"Unknown service type: {service_type}"
            }
    
    async def _stop_service(self, service: Dict) -> Dict:
        """Stop a single service based on its type"""
        service_type = service.get('type')
        service_name = service.get('name')
        
        if service_type == 'docker':
            return self.docker_service.stop_container(service_name)
        elif service_type == 'vm':
            return self.vm_service.stop_vm(service_name)
        elif service_type == 'mock':
            return await self.mock_service.stop_service(service_name)
        else:
            return {
                "status": "error",
                "message": f"Unknown service type: {service_type}"
            }
    
    async def _run_hooks(self, hooks: List[str]):
        """Run a list of hook commands"""
        for hook in hooks:
            try:
                process = await asyncio.create_subprocess_shell(
                    hook,
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                )
                stdout, stderr = await process.communicate()
                
                if process.returncode != 0:
                    print(f"Hook failed: {hook}")
                    print(f"Error: {stderr.decode()}")
                else:
                    print(f"Hook executed: {hook}")
                    if stdout:
                        print(f"Output: {stdout.decode()}")
            except Exception as e:
                print(f"Error running hook '{hook}': {e}")
    
    def get_profile_status(self, profile_name: str) -> Dict:
        """Get the status of all services in a profile"""
        profile = self.get_profile(profile_name)
        if not profile:
            return {"status": "error", "message": f"Profile {profile_name} not found"}
        
        services_status = []
        
        for service in profile.get('services', []):
            service_name = service.get('name')
            service_type = service.get('type')
            
            if service_type == 'docker':
                status = self.docker_service.get_container_status(service_name)
            elif service_type == 'vm':
                status = self.vm_service.get_vm_status(service_name)
            elif service_type == 'mock':
                status = self.mock_service.get_service_status(service_name)
            else:
                status = {"status": "unknown", "type": service_type}
            
            status['name'] = service_name
            status['type'] = service_type
            services_status.append(status)
        
        return {
            "profile": profile_name,
            "services": services_status
        }