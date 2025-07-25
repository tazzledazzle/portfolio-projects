from fastapi import WebSocket, WebSocketDisconnect
import asyncio
import docker
import json
from typing import AsyncGenerator

class LogStreamer:
    def __init__(self):
        self.client = docker.from_env()
        self.active_connections = {}
    
    async def stream_logs(self, websocket: WebSocket, container_name: str):
        """WebSocket endpoint for streaming container logs"""
        await websocket.accept()
        
        # Store the connection
        if container_name not in self.active_connections:
            self.active_connections[container_name] = []
        self.active_connections[container_name].append(websocket)
        
        try:
            # Send initial recent logs
            recent_logs = self._get_recent_logs(container_name, 50)
            for log_line in recent_logs:
                await websocket.send_text(json.dumps({
                    "type": "log",
                    "container": container_name,
                    "message": log_line,
                    "timestamp": None
                }))
            
            # Stream live logs
            async for log_line in self._stream_container_logs(container_name):
                await websocket.send_text(json.dumps({
                    "type": "log",
                    "container": container_name,
                    "message": log_line,
                    "timestamp": None
                }))
                
        except WebSocketDisconnect:
            print(f"WebSocket disconnected for container: {container_name}")
        except Exception as e:
            error_msg = json.dumps({
                "type": "error",
                "container": container_name,
                "message": f"Error streaming logs: {str(e)}"
            })
            try:
                await websocket.send_text(error_msg)
            except:
                pass
        finally:
            # Remove the connection
            if container_name in self.active_connections:
                try:
                    self.active_connections[container_name].remove(websocket)
                    if not self.active_connections[container_name]:
                        del self.active_connections[container_name]
                except ValueError:
                    pass
    
    async def _stream_container_logs(self, container_name: str) -> AsyncGenerator[str, None]:
        """Stream logs from a Docker container"""
        try:
            container = self.client.containers.get(container_name)
            
            # Stream logs in real-time
            for line in container.logs(stream=True, follow=True, tail=0):
                yield line.decode('utf-8').strip()
                await asyncio.sleep(0.01)  # Small delay to prevent overwhelming
                
        except docker.errors.NotFound:
            yield f"Container '{container_name}' not found"
        except Exception as e:
            yield f"Error streaming logs: {str(e)}"
    
    def _get_recent_logs(self, container_name: str, lines: int = 100) -> list:
        """Get recent logs from a container"""
        try:
            container = self.client.containers.get(container_name)
            logs = container.logs(tail=lines, timestamps=True).decode('utf-8')
            return logs.split('\n') if logs else []
        except docker.errors.NotFound:
            return [f"Container '{container_name}' not found"]
        except Exception as e:
            return [f"Error getting logs: {str(e)}"]
    
    async def broadcast_to_container_watchers(self, container_name: str, message: str):
        """Broadcast a message to all WebSocket connections watching a container"""
        if container_name in self.active_connections:
            disconnected = []
            for websocket in self.active_connections[container_name]:
                try:
                    await websocket.send_text(json.dumps({
                        "type": "broadcast",
                        "container": container_name,
                        "message": message
                    }))
                except:
                    disconnected.append(websocket)
            
            # Remove disconnected websockets
            for ws in disconnected:
                try:
                    self.active_connections[container_name].remove(ws)
                except ValueError:
                    pass

# Global instance
log_streamer = LogStreamer()

async def stream_logs(websocket: WebSocket, container_name: str):
    """WebSocket endpoint for streaming container logs"""
    await log_streamer.stream_logs(websocket, container_name)