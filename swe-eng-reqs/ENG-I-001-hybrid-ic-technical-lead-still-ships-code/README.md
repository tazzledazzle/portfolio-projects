# ENG-I-001: Hybrid IC + technical lead (still ships code)

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
designs/builds platforms + leadership duties in same role

## Rationale
Owns **shipped service code + leadership artifacts in the same folder**. Not Phase 7 soft-skill mentoring kits; not ADR-only without code (see ENG-E-007 for ADR emphasis).

## Acceptance demo
Live `/v1/demo` on port **18601** proves `code_shipped`, `leadership_artifacts`, and `hybrid_ic`. Leadership notes under `artifacts/leadership/` are original paraphrases — no copyrighted book text.

## Run

```bash
make test
make demo-local
make demo   # optional Docker Compose
make down
```

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET /v1/leadership` — lists leadership artifact paths + hybrid status
- `GET|POST /v1/demo` — live hybrid IC proof
- `GET /metrics`
