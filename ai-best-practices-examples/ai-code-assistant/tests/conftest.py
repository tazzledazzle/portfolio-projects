from pathlib import Path

import pytest


def pytest_collection_modifyitems(config: pytest.Config, items: list[pytest.Item]) -> None:
    del config
    for item in items:
        path = Path(str(item.fspath))
        path_parts = set(path.parts)

        if "e2e" in path_parts:
            item.add_marker(pytest.mark.e2e)
            continue
        if "integration" in path_parts:
            item.add_marker(pytest.mark.integration)
            continue

        # Default everything else to unit so legacy top-level tests stay selectable.
        item.add_marker(pytest.mark.unit)
