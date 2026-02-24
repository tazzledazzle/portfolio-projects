# Test Plan / Test Approach

## Scope

- **Unit tests:** Handlers and service logic in isolation (mocked dependencies).
- **Integration tests:** API endpoints against the running app (TestClient); no external DB required for the minimal demo (in-memory or mocked).

## Coverage goals

- All API routes have at least one happy-path and one error-path test.
- Critical business logic (e.g. validation, status transitions) is unit-tested.
- No secrets or real external services in tests.

## Running tests

```bash
pytest -v
pytest -v --cov=app --cov-report=term-missing
```

## CI

GitHub Actions runs `pytest` on push/PR. See `.github/workflows/build.yml`.
