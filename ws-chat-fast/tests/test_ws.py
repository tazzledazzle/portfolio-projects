import pytest
from websockets import connect

@pytest.mark.asyncio
async def test_echo_message(live_server):
    uri = f"ws://{live_server.host}:{live_server.port}/ws"
    async with connect(uri) as websocket:
        message = "Hello, WebSocket!"
        await websocket.send(message)
        response = await websocket.recv()
        assert response == message