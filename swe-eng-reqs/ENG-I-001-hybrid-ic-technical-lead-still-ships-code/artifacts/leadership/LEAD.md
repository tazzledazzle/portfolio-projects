# Tech Lead Habits (portfolio notes)

Original catalog of practices for a hybrid IC / technical lead — not a soft-skill mentoring kit and not extracted book text.

## Still ships code

- Own at least one production-facing vertical slice each quarter (design → merge → operate).
- Prefer reviewing by pairing on the hard path over drive-by LGTM comments.
- Keep a personal merge cadence visible so the team sees IC work is not optional theater.

## Decision hygiene

- Write the decision, options considered, and what would change your mind — short ADRs when scope warrants.
- Separate "must ship" from "nice polish" before the design review ends.
- Escalate early when platform constraints (multi-tenant, SLOs, blast radius) collide with feature pressure.

## Coaching without a people-manager kit

- Give concrete feedback tied to artifacts (PRs, runbooks, incident notes), not personality labels.
- Rotate on-call ownership and design-review facilitation so expertise spreads.
- Protect deep-work blocks for ICs; cancel status theater that does not unblock delivery.
