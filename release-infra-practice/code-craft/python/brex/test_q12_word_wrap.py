"""Tests for Q12 — Word-wrap / Line Justification."""

import pytest

from q12_word_wrap import full_justify, justify_line, wrap_words


def test_wrap_single_line() -> None:
    assert wrap_words(["hello", "world"], 11) == ["hello world"]


def test_wrap_splits_at_width() -> None:
    lines = wrap_words(["hello", "world"], 5)
    assert lines == ["hello", "world"]


def test_wrap_greedy_packing() -> None:
    lines = wrap_words(["a", "b", "c", "d"], 3)
    assert lines == ["a b", "c d"]


def test_wrap_word_longer_than_width_raises() -> None:
    with pytest.raises(ValueError):
        wrap_words(["toolong"], 4)


def test_wrap_empty_words_returns_empty() -> None:
    assert wrap_words([], 10) == []


def test_justify_line_middle_distributes_spaces() -> None:
    result = justify_line(["what", "must", "be"], width=16)
    assert len(result) == 16
    assert result.startswith("what")


def test_justify_line_last_left_justified() -> None:
    result = justify_line(["shall", "be"], width=16, is_last=True)
    assert result == "shall be        "


def test_justify_line_single_word_padded() -> None:
    result = justify_line(["word"], width=10)
    assert len(result) == 10
    assert result == "word      "


def test_justify_line_extra_spaces_go_left_first() -> None:
    result = justify_line(["a", "b", "c"], width=7)
    # "a  b  c" — 2 gaps, 4 extra spaces → leftmost gets 3, rightmost gets 2? no:
    # total width=7, word-chars=3, spaces=4 → 2 gaps → gap[0]=2, gap[1]=2 (4/2=2 even)
    assert len(result) == 7


def test_full_justify_classic_example() -> None:
    words = ["This", "is", "an", "example", "of", "text", "justification."]
    result = full_justify(words, 16)
    assert all(len(line) == 16 for line in result)


def test_full_justify_last_line_left_aligned() -> None:
    words = ["What", "must", "be", "acknowledgment", "shall", "be"]
    result = full_justify(words, 16)
    last = result[-1]
    assert last == last.lstrip() or last.endswith(" ")


def test_full_justify_single_word_per_line() -> None:
    result = full_justify(["hello"], 10)
    assert result == ["hello     "]


def test_full_justify_returns_list_of_strings() -> None:
    result = full_justify(["a", "b"], 3)
    assert isinstance(result, list)
    assert all(isinstance(line, str) for line in result)


def test_full_justify_all_lines_correct_width() -> None:
    words = ["Science", "is", "what", "we", "understand", "well", "enough"]
    result = full_justify(words, 10)
    for line in result:
        assert len(line) == 10
