"""Small MCP-inspired tools/list and tools/call surface with mutation policy."""

from __future__ import annotations

from policy_gateway import PolicyGateway, call_with_policy


class MutatingMCPServer:
    def __init__(self, gateway: PolicyGateway | None = None) -> None:
        self.gateway = gateway or PolicyGateway()
        self.restarted_pipelines: list[str] = []
        self.gateway.register(
            "get_pipeline",
            mutating=False,
            description="Read fixture pipeline status",
            input_schema={
                "type": "object",
                "required": ["pipeline"],
                "properties": {"pipeline": {"type": "string"}},
            },
            handler=self._get_pipeline,
        )
        self.gateway.register(
            "restart_pipeline",
            mutating=True,
            description="Restart a fixture pipeline",
            input_schema={
                "type": "object",
                "required": ["pipeline"],
                "properties": {"pipeline": {"type": "string"}},
            },
            handler=self._restart_pipeline,
        )

    def _get_pipeline(self, args: dict) -> dict:
        return {"pipeline": args.get("pipeline", "build-42"), "status": "failed"}

    def _restart_pipeline(self, args: dict) -> dict:
        pipeline = args.get("pipeline", "build-42")
        self.restarted_pipelines.append(pipeline)
        return {"restarted": pipeline}

    def list_tools(self) -> list[dict]:
        return self.gateway.list_tools()

    def call_tool(
        self, name: str, args: dict, approval_token: str | None = None
    ) -> dict:
        return call_with_policy(self.gateway, name, args, approval_token)


def handle_mcp(message: dict, server: MutatingMCPServer) -> dict:
    request_id = message.get("id")
    method = message.get("method")
    if method == "tools/list":
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {"tools": server.list_tools()},
        }
    if method == "tools/call":
        params = message.get("params") or {}
        name = params.get("name")
        args = params.get("arguments") or {}
        if not isinstance(name, str) or not isinstance(args, dict):
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "error": {"code": -32602, "message": "invalid params"},
            }
        try:
            result = server.call_tool(name, args, params.get("approval_token"))
        except KeyError as exc:
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "error": {"code": -32602, "message": str(exc)},
            }
        return {"jsonrpc": "2.0", "id": request_id, "result": result}
    return {
        "jsonrpc": "2.0",
        "id": request_id,
        "error": {"code": -32601, "message": "method not found"},
    }
