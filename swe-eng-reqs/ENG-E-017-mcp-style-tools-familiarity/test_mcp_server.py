from audit import AuditLog, redact
from mcp_server import MCPServer, handle_mcp


def make_server() -> MCPServer:
    server = MCPServer(audit=AuditLog())
    server.register_read_tool(
        "list_pipelines",
        "List recent delivery pipelines",
        lambda arguments: {"pipelines": ["build-main", "release-api"]},
    )
    server.register_read_tool(
        "get_test_flakes",
        "Read recent test flake counts",
        lambda arguments: {"suite": arguments.get("suite", "all"), "flakes": 2},
    )
    return server


def test_tools_list_returns_readonly_tools():
    response = handle_mcp(
        {"jsonrpc": "2.0", "id": 1, "method": "tools/list"},
        make_server(),
    )

    tools = response["result"]["tools"]
    assert len(tools) >= 2
    assert all(tool["mutating"] is False for tool in tools)
    assert all(tool["read_only"] is True for tool in tools)


def test_tools_call_readonly_success():
    server = make_server()
    response = handle_mcp(
        {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "get_test_flakes",
                "arguments": {"suite": "unit"},
            },
        },
        server,
        token="Bearer demo-reader:tools:read",
    )

    assert response["result"]["isError"] is False
    assert response["result"]["content"]["suite"] == "unit"
    assert server.audit.entries[-1]["decision"] == "allow"


def test_tools_call_deny_without_token():
    server = make_server()
    response = handle_mcp(
        {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {"name": "list_pipelines", "arguments": {}},
        },
        server,
    )

    assert response["result"]["isError"] is True
    assert response["result"]["reason"] == "unauthorized"
    assert server.audit.entries[-1]["decision"] == "deny"


def test_tools_call_rejects_mutating_registration():
    server = make_server()

    try:
        server.register_tool(
            "promote_release",
            "Mutate a release",
            lambda arguments: {"promoted": True},
            mutating=True,
        )
    except ValueError as error:
        assert "read-only" in str(error)
    else:
        raise AssertionError("mutating tools must not be registered")


def test_audit_redacts_secrets():
    server = make_server()
    handle_mcp(
        {
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {
                "name": "get_test_flakes",
                "arguments": {
                    "api_key": "sk-secret",
                    "nested": {"password": "hunter2", "safe": "visible"},
                },
            },
        },
        server,
        token="Bearer demo-reader:tools:read",
    )

    entry = server.audit.entries[-1]
    assert entry["arguments"]["api_key"] == "[REDACTED]"
    assert entry["arguments"]["nested"]["password"] == "[REDACTED]"
    assert entry["arguments"]["nested"]["safe"] == "visible"
    assert redact({"token": "secret"}) == {"token": "[REDACTED]"}


def test_unknown_method_errors():
    response = handle_mcp(
        {"jsonrpc": "2.0", "id": 5, "method": "resources/list"},
        make_server(),
    )

    assert response["error"]["code"] == -32601
