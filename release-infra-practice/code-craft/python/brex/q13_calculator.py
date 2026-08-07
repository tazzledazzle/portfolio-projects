"""
Q13 — Calculator

Base: evaluate a string expression containing +, -, *, / with integers and
      correct operator precedence (no parentheses).
Ext 1: support parentheses for grouping.
Ext 2: handle unary negation (e.g. "-3 + 5").

All arithmetic uses integer division (truncation toward zero) for /.
"""

from __future__ import annotations


def calculate(expression: str) -> int:
    """
    Evaluate an arithmetic expression string.

    Supports: integers, +, -, *, /, (, ), unary minus, whitespace.
    Division truncates toward zero.

    Raises:
        ValueError: if the expression is malformed.
        ZeroDivisionError: if division by zero occurs.
    """
    raise NotImplementedError
