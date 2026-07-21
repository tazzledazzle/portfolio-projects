from agent_loop import FixturePlanner, run_agent
from tool_registry import Tool, ToolRegistry


def registry_with_inspect_tool():
    registry = ToolRegistry()
    registry.register(
        Tool(
            name="inspect_pipeline",
            mutating=False,
            handler=lambda args: {"pipeline": args["pipeline"], "status": "failed"},
            input_schema={"pipeline": "string"},
        )
    )
    return registry


def test_agent_loop_plan_execute():
    result = run_agent(
        "explain pipeline build-42",
        registry_with_inspect_tool(),
        FixturePlanner(),
    )

    assert result["plan_steps"] >= 1
    assert result["tools_invoked"] >= 1
    assert result["deterministic_eval_pass"] is True
    assert result["max_steps"] <= 5
    assert result["trace"][0]["phase"] == "act_observe"


def test_agent_loop_respects_max_steps():
    steps = [
        {
            "tool": "inspect_pipeline",
            "args": {"pipeline": f"build-{index}"},
            "rationale": "inspect fixture",
        }
        for index in range(7)
    ]

    result = run_agent(
        "inspect all fixtures",
        registry_with_inspect_tool(),
        FixturePlanner(steps),
        max_steps=3,
    )

    assert result["plan_steps"] == 3
    assert result["tools_invoked"] == 3
    assert len(result["trace"]) == 3
    assert result["max_steps"] == 3
