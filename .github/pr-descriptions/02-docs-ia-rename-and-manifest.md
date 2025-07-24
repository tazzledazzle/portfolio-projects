### Why
Design docs are under `src/main/resources/` and hard to discover. Manifest-driven docs prevent rot.

### What
- Move all design docs → `docs/design-docs/<slug>/design.md`
- Add `docs/templates/design_doc_template.md`
- Introduce `portfolio.yaml` with project metadata
- Update README links

### Checklist
- [ ] All 10 docs moved & follow template
- [ ] `portfolio.yaml` complete
- [ ] CI green