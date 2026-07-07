import argparse

from changelog_generator.diff_engine import diff_api
from changelog_generator.transforms.python_rewriter import generate_python_migration


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="changelog-generator")
    parser.add_argument("--old", required=True)
    parser.add_argument("--new", required=True)
    return parser


def main() -> None:
    args = build_parser().parse_args()
    delta = diff_api(args.old, args.new)
    print("changes", delta)
    print("python_migration", generate_python_migration(delta))


if __name__ == "__main__":
    main()
