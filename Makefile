SHELL := /bin/bash
ADDR ?= :9091

build:
	cd cmd/wsproxy && go build -race -o ../../bin/wsproxy

run: build
	ADDR=$(ADDR) ./bin/wsproxy

# Development auto-reload using Air (https://github.com/cosmtrek/air)
# Install: go install github.com/cosmtrek/air@latest
dev:
	@if ! command -v air >/dev/null 2>&1; then echo "air not found. Install with: go install github.com/air-verse/air@latest"; exit 1; fi
	ADDR=$(ADDR) air -c .air.toml

tidy:
	go mod tidy

test:
	go test ./...

itest:
	go test -v ./internal/integration

e2e-echo:
	# Requires wscat: npm i -g wscat
	wscat -c "ws://localhost:8080/wsproxy?target=wss://echo.websocket.events"

docker-build:
	docker build -t go-proxy -f deploy/Dockerfile .

docker-up:
	cd deploy && docker-compose up --build

frontend-dev-web:
	cd frontend && flutter run -d chrome

frontend-dev-macos:
	cd frontend && flutter run -d macos

frontend-dev-windows:
	cd frontend && flutter run -d windows


