### Why
README should self-sell with an up-to-date projects table.

### What
- `scripts/gen_readme_table.py` to generate table from `portfolio.yaml`
- Inject between `<!-- PROJECTS_TABLE_START -->` / `<!-- PROJECTS_TABLE_END -->`
- CI step fails if README drift

### Checklist
- [ ] Table shows Name/Problem/Stack/Highlights/Status/Links
- [ ] CI step guards drift