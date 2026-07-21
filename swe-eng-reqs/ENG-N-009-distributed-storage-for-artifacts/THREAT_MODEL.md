# Threat / failure notes — ENG-N-009

- **T-3-10 false durability claims:** PUT checks healthy-node quorum before
  acknowledging a write. Durability reports replicas and healthy nodes from
  current node state instead of a configured target.
- **T-3-13 path/digest traversal:** GET and durability routes accept only
  `sha256:<64 lowercase hex>` digests, rejecting malformed and traversal input.
- Object request bodies are capped at 100 MiB with `http.MaxBytesReader`.
- This is a local in-process storage model. A production deployment requires
  authenticated APIs, tenant quotas, independent failure domains, and repair.
