# Scale ADR: {{DECISION_TITLE}}

- Status and date: {{STATUS_AND_DATE}}
- Scope and owners: {{SCOPE_AND_OWNERS}}

## Scale axis

- Current load with units and window: {{CURRENT_LOAD}}
- Forecast load and horizon: {{FORECAST_LOAD}}
- Tenant count: {{TENANT_COUNT}}
- Owning and consuming team count: {{TEAM_COUNT}}
- Assumption sources: {{ASSUMPTION_SOURCES}}

## Capacity math

- Capacity per unit or worker: {{UNIT_CAPACITY}}
- Required capacity: {{REQUIRED_CAPACITY}}
- Reserved headroom: {{HEADROOM}}
- First limiting resource: {{LIMITING_RESOURCE}}

## What breaks at 10×

- Throughput or storage: {{TEN_X_SYSTEM_FAILURE}}
- Failure blast radius: {{TEN_X_FAILURE_DOMAIN}}
- Team coordination or ownership: {{TEN_X_TEAM_FAILURE}}

## Isolation and tenancy

- Proposed boundary: {{FAILURE_DOMAIN_OR_TENANCY_BOUNDARY}}
- Noisy-neighbor control: {{CONTROL}}
- Ownership model: {{OWNERSHIP_MODEL}}

## Scale-driven decision

- Architecture change: {{DECISION}}
- Scale threshold that requires it: {{THRESHOLD}}
- Downside accepted: {{ACCEPTED_DOWNSIDE}}

## Validation and revisit

- Capacity or isolation test: {{TEST}}
- Observed result: {{OBSERVED_RESULT}}
- Recalculation trigger: {{RECALCULATION_TRIGGER}}
- Next review: {{REVIEW_DATE}}

---
*Truthful evidence template: replace placeholders with measured scale data. Use the general trade-off ADR when scale is not the first-class decision axis.*
