"""Run every base problem + all 20 kata variations in one shot."""

import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).parent
FILES = [
    HERE / "02_base_solutions.py",
    HERE / "variations" / "kata_a_dedup.py",
    HERE / "variations" / "kata_b_parse_sort.py",
    HERE / "variations" / "kata_c_scheduling.py",
]

if __name__ == "__main__":
    failures = []
    for f in FILES:
        print(f"\n=== {f.relative_to(HERE)} ===")
        result = subprocess.run([sys.executable, str(f)], capture_output=True, text=True)
        print(result.stdout, end="")
        if result.returncode != 0:
            print(result.stderr)
            failures.append(f)
    print("\n" + "=" * 40)
    if failures:
        print(f"FAILED: {[str(f) for f in failures]}")
        sys.exit(1)
    print("ALL FILES PASSED (base problems + 20/20 kata variations)")
