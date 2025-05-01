# tools/gradle_to_bazel/tests/test_regex.py
import pytest
from migrate import DEPENDENCY_REGEX

@pytest.mark.parametrize("line, expected", [
    ("    implementation(\"org.example:lib:1.2.3\")",  "org.example:lib:1.2.3"),
    ("implementation(\"com.example:artifact:0.0.1\")", "com.example:artifact:0.0.1"),
    ("  implementation(\"my-lib:1.0\")",             "my-lib:1.0"),
    ("implementation('org.test:lib:2.0.0')",         "org.test:lib:2.0.0"),
    ("// implementation(\"org.ignore:lib:1.0.0\")",  None),
    ("implementation   (\"org.space:lib:3.4.5\")",   "org.space:lib:3.4.5"),
])
def test_regex_matches_and_groups(line, expected):
    m = DEPENDENCY_REGEX.match(line)
    if expected is not None:
        assert m, f"Expected a match for: {line!r}"
        assert m.group(1) == expected
    else:
        assert m is None, f"Did not expect a match for: {line!r}"
