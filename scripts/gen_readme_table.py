#!/usr/bin/env python3
import yaml
import sys
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
manifest = yaml.safe_load((ROOT / "portfolio.yaml").read_text())

header = "| Name | Problem | Stack | Highlights | Status | Link |\n|---|---|---|---|---|---|\n"
rows = []
for p in manifest["projects"]:
    rows.append(
        f"| {p['name']} | {p['problem']} | {', '.join(p['stack'])} "
        f"| {', '.join(p['highlights'])} | {p['status']} | [{p['path']}](./{p['path']}) |"
    )

table = header + "\n".join(rows) + "\n"

readme = (ROOT / "README.md").read_text().splitlines()
out, in_block = [], False
for line in readme:
    if line.strip() == "<!-- PROJECTS_TABLE_START -->":
        in_block = True
        out.append(line)
        out.append(table)
        continue
    if line.strip() == "<!-- PROJECTS_TABLE_END -->":
        in_block = False
        out.append(line)
        continue
    if not in_block:
        out.append(line)

new_text = "\n".join(out) + "\n"
if "--check" in sys.argv:
    if new_text != (ROOT / "README.md").read_text():
        print("README.md out of date. Run scripts/gen_readme_table.py", file=sys.stderr)
        sys.exit(1)
else:
    (ROOT / "README.md").write_text(new_text)