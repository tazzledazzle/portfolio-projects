### Why
One-command onboarding: VS Code devcontainer / `make bootstrap`.

### What
- `.devcontainer/devcontainer.json`
- `Makefile` or `justfile` with bootstrap/lint/test/docs/run targets
- `.pre-commit-config.yaml` for ruff/black/ktlint/markdownlint

### Checklist
- [ ] Fresh clone → “Open in container” works
- [ ] `make bootstrap` sets hooks & deps