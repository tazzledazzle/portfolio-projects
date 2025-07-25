import asyncio
import docker
from typing import AsyncGenerator

client = docker.from_env()

async def stream_container_logs(container_name: str) -> AsyncGenerator[str, None]:
    """Stream logs from a Docker container"""
    try:
        container = client.containers.get(container_name)
        
        # Stream logs in real-time
        for line in container.logs(stream=True, follow=True, tail=50):
            yield line.decode('utf-8').strip()
            await asyncio.sleep(0.01)  # Small delay to prevent overwhelming
            
    except docker.errors.NotFound:
        yield f"Container '{container_name}' not found"
    except Exception as e:
        yield f"Error streaming logs: {str(e)}"

def get_container_logs(container_name: str, lines: int = 100) -> list:
    """Get recent logs from a container"""
    try:
        container = client.containers.get(container_name)
        logs = container.logs(tail=lines).decode('utf-8')
        return logs.split('\n')
    except docker.errors.NotFound:
        return [f"Container '{container_name}' not found"]
    except Exception as e:
        return [f"Error getting logs: {str(e)}"]