"""
Q12 — Word-wrap / Line Justification

Base: split a list of words into lines of exactly N characters, keeping whole
      words intact (greedy packing).
Ext 1: fully justify each line — distribute extra spaces evenly across gaps;
       for the last line, left-justify.
Ext 2: handle single-word lines (no internal gaps) by padding with trailing spaces.
"""

from __future__ import annotations


def wrap_words(words: list[str], width: int) -> list[str]:
    """
    Greedily pack words into lines of at most `width` characters.
    Words are separated by a single space; lines are NOT padded.

    Raises:
        ValueError: if any single word is longer than width.
    """
    raise NotImplementedError


def justify_line(words: list[str], width: int, *, is_last: bool = False) -> str:
    """
    Format a list of words into a string of exactly `width` characters.

    - Last line (or single word): left-justified with trailing spaces.
    - Other lines: spaces distributed as evenly as possible; extra spaces
      go to the leftmost gaps first.
    """
    raise NotImplementedError


def full_justify(words: list[str], width: int) -> list[str]:
    """
    End-to-end: pack words into lines then justify each line to exactly width.

    Equivalent to LeetCode #68 Text Justification.
    """
    raise NotImplementedError
