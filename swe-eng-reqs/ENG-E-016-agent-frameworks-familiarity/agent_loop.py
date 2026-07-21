"""Deterministic Plan-Execute loop for agent-framework pedagogy."""

from __future__ import annotations

from copy import deepcopy

from tool_registry import ToolRegistry

MAX_AGENT_STEPS = 5


class FixturePlanner:
    """Emit a deterministic plan rather than calling a live model."""

    def __init__(self, steps: list[dict] | None = None):
        self._steps = steps or [
            {
                "tool": "inspect_pipeline",
                "args": {"pipeline": "build-42"},
                "rationale": "inspect the failing pipeline fixture",
            }
        ]

    def plan(self, goal: str) -> list[dict]:
        if not isinstance(goal, str) or not goal.strip():
            raise ValueError("goal must be a non-empty string")
        return deepcopy(self._steps)


def run_agent(
    goal: str,
    registry: ToolRegistry,
    planner: FixturePlanner,
    *,
    max_steps: int = MAX_AGENT_STEPS,
) -> dict:
    if isinstance(max_steps, bool) or not isinstance(max_steps, int) or max_steps < 1:
        raise ValueError("max_steps must be a positive integer")
    bounded_max = min(max_steps, MAX_AGENT_STEPS)
    planned = planner.plan(goal)
    trace = []
    for index, step in enumerate(planned[:bounded_max], start=1):
        if not isinstance(step, dict):
            observation = {"isError": True, "reason": "invalid_plan_step"}
            tool_name = ""
        else:
            tool_name = step.get("tool", "")
            observation = registry.call(tool_name, step.get("args", {}))
        trace.append(
            {
                "index": index,
                "phase": "act_observe",
                "tool": tool_name,
                "observation": observation,
            }
        )
    deterministic_eval_pass = bool(trace) and all(
        not entry["observation"]["isError"] for entry in trace
    )
    return {
        "goal": goal,
        "plan_steps": len(trace),
        "total_planned": len(planned),
        "tools_invoked": len(trace),
        "trace": trace,
        "deterministic_eval_pass": deterministic_eval_pass,
        "max_steps": bounded_max,
    }
