import argparse
from pathlib import Path

from detector.version_comparator import detect_drift
from fix_script.script_generator import generate_fix_script
from manifest.schema import load_canonical_policy, load_tool_versions
from reporting.console_report import format_drift_report


def main() -> None:
    parser = argparse.ArgumentParser(description="Detect local toolchain drift.")
    parser.add_argument(
        "--canonical",
        type=Path,
        default=Path("config/canonical-toolchain.yaml"),
    )
    parser.add_argument("--local", type=Path, required=True)
    args = parser.parse_args()

    canonical = load_canonical_policy(args.canonical)
    local = load_tool_versions(args.local)
    drifts = detect_drift(canonical, local)
    print(format_drift_report(drifts), end="")
    print(generate_fix_script(drifts), end="")


if __name__ == "__main__":
    main()
