### Code Quality & Analysis (5)

**27. Kotlin custom detekt rules library** — A library of custom `detekt` rules specific to team conventions: enforcing coroutine scope naming, flagging unchecked casts to platform types, detecting missing `@Transactional` annotations on service methods.

**28. License compliance scanner** — Scans all transitive dependencies (via Gradle dependency tree + SPDX database) and fails CI when a new dependency introduces an incompatible license (e.g. GPL in a proprietary codebase), with a SBOM output.

**29. Dead code surface reporter** — Combines call graph analysis (via Kotlin Symbol Processing) and git blame recency to identify code that is both unreachable and untouched for >90 days, generating a prioritized cleanup backlog.

**30. API breaking-change detector** — Compares OpenAPI specs between the current branch and `main`, classifies each diff as breaking/non-breaking per the OpenAPI compatibility rules, and blocks merge on unintentional breaking changes.

**31. Security hotspot annotator** — Integrates Semgrep SAST output into GitHub PR review comments, annotating specific lines with severity, CWE reference, and a plain-English explanation of the risk — linked to the team's remediation playbook.

## Project Scaffolds

- [kotlin-custom-detekt-rules-library](./kotlin-custom-detekt-rules-library)
- [license-compliance-scanner](./license-compliance-scanner)
- [dead-code-surface-reporter](./dead-code-surface-reporter)
- [api-breaking-change-detector](./api-breaking-change-detector)
- [security-hotspot-annotator](./security-hotspot-annotator)