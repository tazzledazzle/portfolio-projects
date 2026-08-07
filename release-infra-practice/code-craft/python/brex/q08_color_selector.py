"""
Q8 — Color Multi-Select (backend state management)

UI round is React; this module models the pure state-machine logic so it can
be tested independently of a browser.
Base: maintain a set of selected colors from a fixed palette.
Ext 1: filter the available palette by a search string.
Ext 2: retrieve detail metadata for each selected color.
"""

from __future__ import annotations

from dataclasses import dataclass


PALETTE: list[str] = ["red", "blue", "green", "yellow", "purple", "orange", "pink"]

COLOR_DETAILS: dict[str, dict[str, str]] = {
    "red": {"hex": "#FF0000", "family": "warm"},
    "blue": {"hex": "#0000FF", "family": "cool"},
    "green": {"hex": "#00FF00", "family": "cool"},
    "yellow": {"hex": "#FFFF00", "family": "warm"},
    "purple": {"hex": "#800080", "family": "cool"},
    "orange": {"hex": "#FFA500", "family": "warm"},
    "pink": {"hex": "#FFC0CB", "family": "warm"},
}


@dataclass
class ColorSelector:
    _selected: set[str] = None  # type: ignore[assignment]

    def __post_init__(self) -> None:
        self._selected = set()

    def select(self, color: str) -> None:
        """Add color to selection. Raise ValueError if not in PALETTE."""
        raise NotImplementedError

    def deselect(self, color: str) -> None:
        """Remove color from selection. No-op if not selected."""
        raise NotImplementedError

    def toggle(self, color: str) -> None:
        """Select if not selected, deselect if already selected."""
        raise NotImplementedError

    @property
    def selected(self) -> list[str]:
        """Return sorted list of currently selected colors."""
        raise NotImplementedError

    def filter_palette(self, query: str) -> list[str]:
        """Return palette colors whose name contains query (case-insensitive)."""
        raise NotImplementedError

    def selected_details(self) -> list[dict[str, str]]:
        """Return COLOR_DETAILS entries for each selected color (sorted by name)."""
        raise NotImplementedError
