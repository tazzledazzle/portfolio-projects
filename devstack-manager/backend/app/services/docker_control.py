try:
    import docker
    DOCKER_AVAILABLE = True
except ImportError:
    DOCKER_AVAILABLE = False
    docker = None

import json
from typing import List, Dict, Optional
from datetime import datetime

class DockerService:
    def __init__(self):
        self.client = None
    
    def _get_docker_client(self):
        """Lazy initialization of Docker client"""
        if not DOCKER_AVAILABLE:
            return None
            
        if self.client is None:
            try:
                self.client = docker.from_env()
            except Exception as e:
                print(f"Failed to initialize Docker client: {e}")
                self.client = None
        return self.client
    
    def list_containers(self) -> List[Dict]:
        """List all containers with detailed information"""
        try:
            client = self._get_docker_client()
            if not client:
                return [{"error": "Docker client not available"}]
            
            containers = client.containers.list(all=True)
            result = []
            
            for container in containers:
                result.append({
                    "name": container.name,
                    "id": container.short_id,
                    "status": container.status,
                    "image": container.image.tags[0] if container.image.tags else "unknown",
                    "ports": self._format_ports(container.ports),
                    "created": container.attrs.get("Created", ""),
                    "labels": container.labels
                })
            
            return result
        except Exception as e:
            return [{"error": f"Failed to list containers: {str(e)}"}]
    
    def start_container(self, name: str) -> Dict:
        """Start a container"""
        try:
            client = self._get_docker_client()
            if not client:
                return {"status": "error", "message": "Docker client not available"}
            
            container = client.containers.get(name)
            container.start()
            return {
                "status": "started",
                "name": name,
                "message": f"Container {name} started successfully"
            }
        except docker.errors.NotFound:
            return {"status": "error", "message": f"Container {name} not found"}
        except Exception as e:
            return {"status": "error", "message": str(e)}
    
    def stop_container(self, name: str) -> Dict:
        """Stop a container"""
        try:
            client = self._get_docker_client()
            if not client:
                return {"status": "error", "message": "Docker client not available"}
            
            container = client.containers.get(name)
            container.stop()
            return {
                "status": "stopped",
                "name": name,
                "message": f"Container {name} stopped successfully"
            }
        except docker.errors.NotFound:
            return {"status": "error", "message": f"Container {name} not found"}
        except Exception as e:
            return {"status": "error", "message": str(e)}
    
    def restart_container(self, name: str) -> Dict:
        """Restart a container"""
        try:
            client = self._get_docker_client()
            if not client:
                return {"status": "error", "message": "Docker client not available"}
            
            container = client.containers.get(name)
            container.restart()
            return {
                "status": "restarted",
                "name": name,
                "message": f"Container {name} restarted successfully"
            }
        except docker.errors.NotFound:
            return {"status": "error", "message": f"Container {name} not found"}
        except Exception as e:
            return {"status": "error", "message": str(e)}
    
    def get_container_status(self, name: str) -> Dict:
        """Get detailed status of a container"""
        try:
            client = self._get_docker_client()
            if not client:
                return {"error": "Docker client not available"}
            
            container = client.containers.get(name)
            stats = container.stats(stream=False)
            
            return {
                "name": container.name,
                "status": container.status,
                "image": container.image.tags[0] if container.image.tags else "unknown",
                "ports": self._format_ports(container.ports),
                "created": container.attrs.get("Created", ""),
                "started": container.attrs.get("State", {}).get("StartedAt", ""),
                "cpu_usage": self._calculate_cpu_usage(stats),
                "memory_usage": self._calculate_memory_usage(stats),
                "network": self._get_network_stats(stats)
            }
        except docker.errors.NotFound:
            return {"error": f"Container {name} not found"}
        except Exception as e:
            return {"error": str(e)}
    
    def get_container_logs(self, name: str, lines: int = 100) -> List[str]:
        """Get recent logs from a container"""
        try:
            client = self._get_docker_client()
            if not client:
                return ["Docker client not available"]
            
            container = client.containers.get(name)
            logs = container.logs(tail=lines, timestamps=True).decode('utf-8')
            return logs.split('\n') if logs else []
        except docker.errors.NotFound:
            return [f"Container '{name}' not found"]
        except Exception as e:
            return [f"Error getting logs: {str(e)}"]
    
    def create_container_from_config(self, config: Dict) -> Dict:
        """Create a container from configuration"""
        try:
            client = self._get_docker_client()
            if not client:
                return {"status": "error", "message": "Docker client not available"}
            
            container = client.containers.run(
                image=config.get("image"),
                name=config.get("name"),
                ports=self._parse_ports(config.get("ports", [])),
                volumes=self._parse_volumes(config.get("volumes", [])),
                environment=config.get("environment", {}),
                detach=True,
                remove=False
            )
            
            return {
                "status": "created",
                "name": container.name,
                "id": container.short_id
            }
        except Exception as e:
            return {"status": "error", "message": str(e)}
    
    def _format_ports(self, ports: Dict) -> List[Dict]:
        """Format port mappings for display"""
        formatted = []
        for container_port, host_bindings in ports.items():
            if host_bindings:
                for binding in host_bindings:
                    formatted.append({
                        "container_port": container_port,
                        "host_port": binding.get("HostPort"),
                        "host_ip": binding.get("HostIp", "0.0.0.0")
                    })
        return formatted
    
    def _parse_ports(self, ports: List) -> Dict:
        """Parse port configuration"""
        port_dict = {}
        for port in ports:
            if isinstance(port, str) and ":" in port:
                host_port, container_port = port.split(":")
                port_dict[f"{container_port}/tcp"] = host_port
            elif isinstance(port, int):
                port_dict[f"{port}/tcp"] = port
        return port_dict
    
    def _parse_volumes(self, volumes: List[str]) -> Dict:
        """Parse volume configuration"""
        volume_dict = {}
        for volume in volumes:
            if ":" in volume:
                host_path, container_path = volume.split(":", 1)
                volume_dict[host_path] = {"bind": container_path, "mode": "rw"}
        return volume_dict
    
    def _calculate_cpu_usage(self, stats: Dict) -> float:
        """Calculate CPU usage percentage"""
        try:
            cpu_stats = stats.get("cpu_stats", {})
            precpu_stats = stats.get("precpu_stats", {})
            
            cpu_delta = cpu_stats.get("cpu_usage", {}).get("total_usage", 0) - \
                       precpu_stats.get("cpu_usage", {}).get("total_usage", 0)
            system_delta = cpu_stats.get("system_cpu_usage", 0) - \
                          precpu_stats.get("system_cpu_usage", 0)
            
            if system_delta > 0:
                cpu_percent = (cpu_delta / system_delta) * len(cpu_stats.get("cpu_usage", {}).get("percpu_usage", [])) * 100
                return round(cpu_percent, 2)
        except:
            pass
        return 0.0
    
    def _calculate_memory_usage(self, stats: Dict) -> Dict:
        """Calculate memory usage"""
        try:
            memory_stats = stats.get("memory_stats", {})
            usage = memory_stats.get("usage", 0)
            limit = memory_stats.get("limit", 0)
            
            return {
                "usage_bytes": usage,
                "limit_bytes": limit,
                "usage_mb": round(usage / (1024 * 1024), 2),
                "limit_mb": round(limit / (1024 * 1024), 2),
                "percentage": round((usage / limit) * 100, 2) if limit > 0 else 0
            }
        except:
            return {"usage_bytes": 0, "limit_bytes": 0, "usage_mb": 0, "limit_mb": 0, "percentage": 0}
    
    def _get_network_stats(self, stats: Dict) -> Dict:
        """Get network statistics"""
        try:
            networks = stats.get("networks", {})
            total_rx = sum(net.get("rx_bytes", 0) for net in networks.values())
            total_tx = sum(net.get("tx_bytes", 0) for net in networks.values())
            
            return {
                "rx_bytes": total_rx,
                "tx_bytes": total_tx,
                "rx_mb": round(total_rx / (1024 * 1024), 2),
                "tx_mb": round(total_tx / (1024 * 1024), 2)
            }
        except:
            return {"rx_bytes": 0, "tx_bytes": 0, "rx_mb": 0, "tx_mb": 0}