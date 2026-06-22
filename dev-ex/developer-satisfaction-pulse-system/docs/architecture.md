# Developer Satisfaction Pulse System Architecture

- Survey responses are ingested through API endpoints.
- Raw responses are persisted to Postgres for historical analysis.
- A scoring service computes rolling NPS and friction indicators per team.
- Dashboard adapters expose trend-ready aggregates to BI or internal portals.
