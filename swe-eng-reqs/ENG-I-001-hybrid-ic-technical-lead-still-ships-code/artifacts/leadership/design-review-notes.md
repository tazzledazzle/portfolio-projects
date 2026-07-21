# Design Review Notes (template)

Lightweight checklist used when a hybrid IC/tech lead facilitates a platform design review. Original portfolio material — paraphrase of common industry practice catalogs only; no copyrighted book excerpts.

## Before the review

1. Problem statement in one paragraph (who hurts, how often, blast radius).
2. Non-goals listed explicitly.
3. At least two alternatives with trade-offs (cost, operability, migration).
4. Threat / failure notes for trust boundaries touched by the change.

## During the review

- Ask: what is the rollback / degrade path?
- Ask: which metrics prove the design worked two weeks after launch?
- Ask: who owns the on-call page when this breaks at 2am?
- Capture open questions with owners and dates — do not leave "TBD forever".

## After the review

- Publish the decision (ADR or short note) in the same repo as the code when practical.
- File follow-ups as tracked work, not chat folklore.
- If the lead is also the implementer, schedule a second reviewer for merge.
