SERVICE := eng-e-018
PORT := 18818
COMPOSE := docker compose

.PHONY: up down test demo logs build

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

test:
	python -m pytest -q

demo: up
	@echo "Waiting for health..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  curl -sf http://127.0.0.1:$$(docker compose port $(SERVICE) 8080 | cut -d: -f2)/healthz && break; \
	  sleep 2; \
	done
	@PORT=$$(docker compose port $(SERVICE) 8080 | cut -d: -f2); \
	  curl -s http://127.0.0.1:$$PORT/v1/demo | tee demo-output.json
	@echo ""
	@echo "Demo complete → demo-output.json"

logs:
	$(COMPOSE) logs -f

build:
	python main.py

.PHONY: demo-local
demo-local: test
	@echo "Starting local demo (no Docker) on :$(PORT)..."
	@set -e; \
	  lsof -ti :$(PORT) | xargs kill -9 2>/dev/null || true; \
	  PORT=$(PORT) ADDR=127.0.0.1 python main.py >/tmp/$(SERVICE)-local.log 2>&1 & pid=$$!; \
	  echo $$pid > /tmp/$(SERVICE)-local.pid; \
	  trap 'kill $$pid 2>/dev/null || true; lsof -ti :$(PORT) | xargs kill -9 2>/dev/null || true; rm -f /tmp/$(SERVICE)-local.pid' EXIT; \
	  ready=0; \
	  for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
	    if curl -sf http://127.0.0.1:$(PORT)/healthz >/dev/null; then ready=1; break; fi; \
	    sleep 0.3; \
	  done; \
	  test $$ready -eq 1; \
	  curl -sf http://127.0.0.1:$(PORT)/v1/demo | tee demo-output.json; \
	  echo ""; \
	  echo "Local demo complete → demo-output.json"
