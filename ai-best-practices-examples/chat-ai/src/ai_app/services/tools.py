from collections.abc import Callable


ToolFn = Callable[[str], str]


class ToolService:
    def __init__(self) -> None:
        self._tools: dict[str, ToolFn] = {}

    def register(self, name: str, tool: ToolFn) -> None:
        self._tools[name] = tool

    def execute(self, name: str, query: str) -> str:
        if name not in self._tools:
            raise KeyError(f"Unknown tool: {name}")
        return self._tools[name](query)

    def has(self, name: str) -> bool:
        return name in self._tools
