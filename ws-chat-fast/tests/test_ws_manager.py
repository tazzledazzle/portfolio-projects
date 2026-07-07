from ws_chat.ws_manager import ConnectionManager


def test_connection_manager_starts_empty() -> None:
    manager = ConnectionManager()
    assert manager.active_connections == []


def test_disconnect_removes_connection() -> None:
    manager = ConnectionManager()
    websocket = object()
    manager.active_connections.append(websocket)
    manager.disconnect(websocket)
    assert websocket not in manager.active_connections
