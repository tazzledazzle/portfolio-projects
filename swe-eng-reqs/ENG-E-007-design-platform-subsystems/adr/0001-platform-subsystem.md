# ADR 0001: Platform Subsystem Boundary for Catalog Control Plane

**Status:** Accepted  
**ID:** `0001-platform-subsystem`  
**Date:** 2026-07-18

## Context

We need a thin control-plane skeleton that records architectural choices for a
platform subsystem (project/pipeline/environment catalog surface) without
collapsing into a mega-service or soft-skill mentoring kit. The design must
stay runnable as a local demo and stay honest about what it owns versus peer
requirement folders (production metrics, HPA packaging, leadership kits).

## Alternatives

### Alternative A — Monolithic platform service

Ship one process that owns catalog CRUD, OpenAPI craft, SLO metrics, HPA
packaging, and leadership artifacts together.

- **Pros:** Single deployable; fewer ports in local demos.
- **Cons:** Collapses atomic requirement IDs; weakens Boundary Matrix proofs;
  harder to evolve ownership independently.

### Alternative B — ADR-backed thin skeleton (chosen)

Keep a small Go HTTP service that loads Architecture Decision Records from
`adr/`, exposes list + skeleton endpoints, and references the accepted ADR ID
from a minimal runtime skeleton. Defer deep metrics, HPA, and mentoring kits
to their owning IDs.

- **Pros:** Preserves one-folder-per-ID; documents ≥2 alternatives with trade-
  offs; demo proves `decision_recorded` without overbuilding.
- **Cons:** More folders to operate; readers must follow ADR links to understand
  the full platform picture.

### Alternative C — Docs-only design without runnable skeleton

Publish ADRs in markdown only; skip a live service.

- **Pros:** Fastest to write.
- **Cons:** Fails the portfolio gate for live `/v1/demo` proofs and runnable
  slices (`make test` / `make demo-local`).

## Decision

Adopt **Alternative B**: ADR-backed thin skeleton. Record decisions in
`adr/0001-platform-subsystem.md` and expose them via `DesignStore` /
`GET /v1/adrs` and `GET /v1/skeleton`.

## Consequences

- Positive: Clear ownership boundary; interviewable design trail with trade-offs.
- Negative: Skeleton is intentionally thin — production metrics and HPA remain
  out of scope for ENG-E-007.
- Follow-ups: Peer folders own OpenAPI (E-021), production service craft (E-008),
  and leadership kits (Phase 7 / I-001 as applicable).
