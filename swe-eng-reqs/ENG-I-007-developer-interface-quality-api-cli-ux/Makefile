SERVICE := eng-i-007
COMPOSE := docker compose

.PHONY: up down test demo logs build

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

test:
	go test -race ./...

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
	go run .

.PHONY: demo-local
demo-local: test
	@echo "Starting local demo (no Docker) on :18607..."
	@go build -o /tmp/$(SERVICE)-local.bin .
	@lsof -ti :18607 | xargs kill -9 2>/dev/null || true
	@ADDR=:18607 /tmp/$(SERVICE)-local.bin >/tmp/$(SERVICE)-local.log 2>&1 & echo $$! > /tmp/$(SERVICE)-local.pid
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do curl -sf http://127.0.0.1:18607/healthz && break; sleep 0.3; done
	@curl -s http://127.0.0.1:18607/v1/demo | tee demo-output.json
	@echo ""
	@kill `cat /tmp/$(SERVICE)-local.pid` 2>/dev/null || true
	@lsof -ti :18607 | xargs kill -9 2>/dev/null || true
	@rm -f /tmp/$(SERVICE)-local.pid /tmp/$(SERVICE)-local.bin
	@echo "Local demo complete → demo-output.json"
