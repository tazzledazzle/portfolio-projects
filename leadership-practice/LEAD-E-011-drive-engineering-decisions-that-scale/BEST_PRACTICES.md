# LEAD-E-011 — Drive engineering decisions that scale

## Best practices

- Define current and forecast load with units, time windows, headroom, and the source of each assumption.
- Identify what fails first at 10× across throughput, state, coordination, and operational ownership.
- Select failure domains and tenancy boundaries from blast-radius and noisy-neighbor evidence.
- Model team growth and ownership handoffs alongside machine capacity.
- Tie the architecture change to a threshold and verify it with observed capacity or isolation evidence.
- Keep the scale axis explicit so the proof does not collapse into a generic alternatives ADR.

## Four habits

1. Include capacity math and named assumptions in every scale-sensitive decision.
2. Ask what breaks at 10× and record the first limiting resource or coordination boundary.
3. Review failure-domain and tenant-isolation evidence before approving shared state.
4. Recalculate the scale model when load, tenant count, or owning-team count crosses a threshold.

## Sources

- *The Staff Engineer’s Path* (Tanya Reilly) — O’Reilly — catalog themes on technical decisions that hold across teams and organizational scope: https://www.oreilly.com/library/view/the-staff-engineers/9781098118723/
- *Software Architecture: The Hard Parts* (Ford, Richards, et al.) — O’Reilly — catalog themes on distributed architecture, coupling, and trade-off analysis: https://www.oreilly.com/library/view/software-architecture-the/9781492086888/
- *Architecture Modernization* (Nick Tune & Perrin) — Manning — catalog themes on modernization and socio-technical boundaries: https://www.manning.com/books/architecture-modernization

---
*This kit paraphrases catalog-level ideas only; it contains no copyrighted book text. Interview stories and artifacts are truthful templates to complete with real evidence.*
