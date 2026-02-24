# REST API Test Demo

**Problem:** Show a minimal but production-style REST API with a clear contract, automated tests, and runnable deployment.

**Solution:** A small FastAPI service with OpenAPI/Swagger, unit and integration tests, Docker, and CI that runs tests and builds the image.

**What this demonstrates:** Shipping a well-tested, documented service and articulating QA practices (test cases, automation)—aligned with RESTful web services, Dockerized apps, and test plans.

---

## Contents

- **API:** FastAPI app with a few resources (e.g. items CRUD-style), OpenAPI at `/docs` and `/openapi.json`.
- **Tests:** Pytest unit and integration tests; see [Test approach](docs/test-plan.md).
- **Docker:** Dockerfile and docker-compose for local run.
- **CI:** GitHub Actions: run tests, optionally build image.

---

## How to run

**Local (no Docker):**

```bash
python -m venv .venv
source .venv/bin/activate   # or .venv\Scripts\activate on Windows
pip install -r requirements.txt
uvicorn app.main:app --reload
```

Open http://localhost:8000/docs for Swagger UI.

**Docker:**

```bash
docker compose up --build
```

API at http://localhost:8000; docs at http://localhost:8000/docs.

---

## How to run tests

```bash
pip install -r requirements.txt
pytest -v
# With coverage:
pytest -v --cov=app --cov-report=term-missing
```

---

## License

MIT.
