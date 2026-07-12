# Synthetic Data Factory — Design

**Date:** 2026-07-12  
**Status:** Approved  
**Bead:** portfolio-projects-dmq  
**Author:** Terence Schumacher (via brainstorming)

## 1. Goal

Securely generate and exercise **synthetic marketplace traffic** against the local C2C stack (docker-compose or kind) so demos and light load tests are reproducible — without inventing real PII and without treating the LLM as the source of truth for payloads.

This is a **data factory**, not a chaos monkey. Fault injection is explicitly out of v1.

## 2. Decisions (locked)

| Decision | Choice |
|---|---|
| Primary goal | Data factory (demo + load), not chaos |
| Delivery shape | Hybrid: Cursor subagent **orchestrates**; in-repo harness **generates** |
| Security bar | Safe by construction (`synth-*` tags, wordlists); local kind/compose by convention |
| v1 surface | Full: listings → search → purchase (confirm/dispute) → chat |
| Approach | #3 Hybrid (Kotlin HTTP harness + thin chat driver + agent) |

## 3. Architecture

```
┌─────────────────────┐     invokes      ┌──────────────────────────┐
│ Cursor subagent     │ ───────────────► │ scripts/synth-run.sh     │
│ (orchestrator only) │                  │ (profile → harness args) │
└─────────────────────┘                  └────────────┬─────────────┘
                                                      │
                         ┌────────────────────────────┼────────────────┐
                         ▼                            ▼                ▼
              ┌──────────────────┐      ┌─────────────────┐   ┌────────────────┐
              │ :synth-harness   │      │ chat driver     │   │ assertions     │
              │ (Kotlin HTTP)    │      │ (WS helper)     │   │ (in harness)   │
              └────────┬─────────┘      └────────┬────────┘   └───────┬────────┘
                       │                         │                    │
                       ▼                         ▼                    ▼
              listings / search / payments   messaging WS      exit 0/1 + summary
```

- **Harness owns data** — seeded RNG, `synth-` prefixes, fixed wordlists; no LLM-invented PII.
- **Agent owns orchestration** — pick profile, run harness against localhost/kind NodePorts, report summary.
- **All writes via public HTTP/WS APIs** — no direct DB inserts (keeps Kafka → OpenSearch path honest).

## 4. Components

| Piece | Path | Responsibility |
|---|---|---|
| Cursor subagent | `.claude/agents/synth-data-factory.md` | Orchestrate only; never invent free-form user data |
| Runner | `scripts/synth-run.sh` | Env URLs, profile selection, invoke harness + chat, unify exit code |
| HTTP harness | `:synth-harness` Gradle module | Listings, search poll, orders confirm/dispute |
| Chat driver | `scripts/synth-chat.sh` (or WS client inside harness) | Two WS peers, N messages, delivery assert |
| Profiles | `synth/profiles/{demo,load-light}.json` | Counts, confirm/dispute mix, geo, seed |

### Synthetic identity rules

- User IDs: `synth-buyer-{n}`, `synth-seller-{n}`
- Listing titles: wordlist + index (no real names/phones/emails)
- Conversation IDs: `synth-buyer-a:synth-seller-b` (matches existing messaging convention)

### Profiles (v1)

| Profile | Listings | Orders | Chat | Intent |
|---|---|---|---|---|
| `demo` | 10 | 5 (mix confirm/dispute) | 1 pair, few messages | Fast demo / CI-ish |
| `load-light` | 100 | 20 | brief burst | Mild load without melting kind |

## 5. Assertions

Fail the run (non-zero exit) if:

1. Created listing does not appear in `GET /search` after retry/backoff.
2. Order does not reach expected terminal escrow status (RELEASED or REFUNDED per profile).
3. Chat peer does not observe at least one message within timeout.

Print a JSON summary on completion: `created`, `indexed`, `orders`, `chatOk`, `errors[]`.

## 6. Error handling

- Per-request HTTP failures: log status + body, count error; continue unless `--fail-fast`.
- Search lag: retry (~10× / 500ms) then fail that listing’s indexed assertion.
- Chat: hard timeout; no infinite wait.
- Exit `0` only if required assertions pass.

## 7. Testing strategy

- **Unit:** generators always emit `synth-` / wordlist-only strings.
- **Contract:** harness happy-path against mocked HTTP (or Ktor test doubles) for listing→order.
- **Live:** `demo` profile against kind/compose (extends `scripts/smoke-test.sh` purchase path).

## 8. Ops / SRE notes (v1 light)

- Tag synthetic actors so future LGTM metrics/logs can filter `synth-*`.
- Rollback = stop the harness; data is append-only mock rows (cluster recreate clears state).
- No prod gates, no chaos, no secret material in profiles.

## 9. Out of scope (v1)

- Chaos (pod kill, network partition, Kafka lag injection)
- Hard environment allowlists / kill switches
- Direct DB seeding
- Real PII or production targets
- Full k6/Gatling load suite

## 10. Success criteria

1. `./scripts/synth-run.sh demo` exits 0 on kind (listings + search + purchase + chat).
2. Subagent can run that path and return the summary without inventing PII.
3. `load-light` completes without crashing services.

## 11. Related

- Existing smoke: `scripts/smoke-test.sh` (purchase path only)
- Services: listings `:8081`, search `:8082`, messaging `:8083`, payments `:8084`
- Observability workstream: `.planning/quick/260711-x5m-*` (filter synthetic traffic later)

---

*Approved in brainstorming session 2026-07-12.*
