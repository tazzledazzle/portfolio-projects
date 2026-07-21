# Threat / failure notes — ENG-N-012

- **T-3-11 split-brain / false sync claims:** replication is asynchronous;
  `lag_ms` and `pending` remain non-zero until regional copies are present. The
  demo verifies a secondary GET before reporting `replicated=true`.
- **T-3-13 region and digest traversal:** region names reject separators and
  `..`; reads require `sha256:<64 lowercase hex>` identifiers.
- Blob request bodies are capped at 100 MiB.
- This local model has no production conflict resolution. A real deployment
  needs authenticated writes, regional retry queues, checksums, and repair.
