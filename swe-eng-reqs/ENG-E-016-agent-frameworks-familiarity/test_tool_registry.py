from tool_registry import Tool, ToolRegistry


def test_registry_register_and_call_readonly():
    registry = ToolRegistry()
    registry.register(
        Tool(
            name="inspect_pipeline",
            mutating=False,
            handler=lambda args: {"pipeline": args["pipeline"], "status": "failed"},
            input_schema={"pipeline": "string"},
        )
    )

    result = registry.call("inspect_pipeline", {"pipeline": "build-42"})

    assert result["isError"] is False
    assert result["content"]["status"] == "failed"


def test_registry_mutating_requires_approval():
    side_effects = []
    registry = ToolRegistry()
    registry.register(
        Tool(
            name="retry_pipeline",
            mutating=True,
            handler=lambda args: side_effects.append(args["pipeline"]),
            input_schema={"pipeline": "string"},
        )
    )

    result = registry.call("retry_pipeline", {"pipeline": "build-42"})

    assert result["isError"] is True
    assert result["reason"] == "approval_required"
    assert side_effects == []


def test_registry_unknown_tool_errors():
    result = ToolRegistry().call("missing", {})

    assert result["isError"] is True
    assert result["reason"] == "unknown_tool"
