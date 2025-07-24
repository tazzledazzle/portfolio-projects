### Why
Single entrypoint (`./tools/run --project <id>`) proves DX focus & makes demos trivial.

### What
- `tools/run` script
- Each project defines `run.yaml` (build/run commands, env vars, deps)
- README section “Run anything in 1 command”

### Checklist
- [ ] Every project runnable
- [ ] Helpful errors if config missing