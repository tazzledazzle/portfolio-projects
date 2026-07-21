SERVICE := eng-n-015
PORT := 18915
COMPOSE := docker compose

.PHONY: up down test demo logs build demo-local

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

demo-local: test
	@echo "Starting local demo (no Docker) on :$(PORT)..."
	@lsof -ti :$(PORT) | xargs kill -9 2>/dev/null || true
	@PORT=$(PORT) ADDR=127.0.0.1 python main.py >/tmp/$(SERVICE)-local.log 2>&1 & echo $$! > /tmp/$(SERVICE)-local.pid
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do curl -sf http://127.0.0.1:$(PORT)/healthz && break; sleep 0.3; done
	@curl -s http://127.0.0.1:$(PORT)/v1/demo | tee demo-output.json
	@echo ""
	@kill `cat /tmp/$(SERVICE)-local.pid` 2>/dev/null || true
	@lsof -ti :$(PORT) | xargs kill -9 2>/dev/null || true
	@rm -f /tmp/$(SERVICE)-local.pid
	@echo "Local demo complete → demo-output.json"
