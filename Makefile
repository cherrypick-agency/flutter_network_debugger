SHELL := /bin/bash
ADDR ?= :9091

# Build metadata
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X network-debugger/internal/infrastructure/observability.Version=$(VERSION) -X network-debugger/internal/infrastructure/observability.Commit=$(COMMIT) -X network-debugger/internal/infrastructure/observability.Date=$(DATE)

# Where to publish packaged artifacts by default
PUBLISH_DIR ?= web

build:
	cd cmd/network-debugger && go build -race -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger

run: build
	ADDR=$(ADDR) ./bin/network-debugger

# Development auto-reload using Air (https://github.com/cosmtrek/air)
# Install: go install github.com/cosmtrek/air@latest
dev:
	@if ! command -v air >/dev/null 2>&1; then echo "air not found. Install with: go install github.com/air-verse/air@latest"; exit 1; fi
	ADDR=$(ADDR) DEV_MODE=1 air -c .air.toml

tidy:
	go mod tidy

test:
	go test ./...

itest:
	go test -v ./internal/integration

e2e-echo:
	# Requires wscat: npm i -g wscat
	wscat -c "ws://localhost:8080/network-debugger?_target=wss://echo.websocket.events"

docker-build:
	docker build -t network-debugger -f deploy/Dockerfile .

docker-up:
	cd deploy && docker-compose up --build

frontend-dev-web:
	cd frontend && flutter run -d chrome

frontend-build-web:
	cd frontend && flutter build web --release
	rm -rf cmd/network-debugger-web/_web
	mkdir -p cmd/network-debugger-web/_web
	cp -R frontend/build/web/* cmd/network-debugger-web/_web/

build-app:
	$(MAKE) frontend-build-web
	cd cmd/network-debugger-web && go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web

run-app: build-app
	ADDR=$(ADDR) ./bin/network-debugger-web

win-app:
	$(MAKE) frontend-build-web
	cd cmd/network-debugger-web && GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_amd64.exe
	cd cmd/network-debugger-web && GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_386.exe
	cd cmd/network-debugger-web && GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_arm64.exe

# Cross-platform builds for network-debugger
.PHONY: build-cross network-debugger-darwin network-debugger-linux network-debugger-windows
build-cross: network-debugger-darwin network-debugger-linux network-debugger-windows

network-debugger-darwin:
	@mkdir -p bin
	cd cmd/network-debugger && \
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_darwin_amd64 && \
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_darwin_arm64

network-debugger-linux:
	@mkdir -p bin
	cd cmd/network-debugger && \
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_linux_amd64 && \
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_linux_arm64

network-debugger-windows:
	@mkdir -p bin
	cd cmd/network-debugger && \
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_windows_amd64.exe && \
	GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_windows_386.exe && \
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger_windows_arm64.exe

# Cross-platform builds for wsapp (includes embedding built web)
.PHONY: build-app-cross network-debugger-web-darwin network-debugger-web-linux network-debugger-web-windows
build-app-cross: frontend-build-web network-debugger-web-darwin network-debugger-web-linux network-debugger-web-windows

network-debugger-web-darwin:
	@mkdir -p bin
	cd cmd/network-debugger-web && \
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_darwin_amd64 && \
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_darwin_arm64

network-debugger-web-linux:
	@mkdir -p bin
	cd cmd/network-debugger-web && \
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_linux_amd64 && \
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_linux_arm64

network-debugger-web-windows:
	@mkdir -p bin
	cd cmd/network-debugger-web && \
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_amd64.exe && \
	GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_386.exe && \
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_arm64.exe

# Packaging
.PHONY: package
package: build-cross build-app-cross
	@mkdir -p $(PUBLISH_DIR)/downloads
	cd bin && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger_windows_amd64.zip network-debugger_windows_amd64.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger_windows_386.zip network-debugger_windows_386.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger_windows_arm64.zip network-debugger_windows_arm64.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger-web_windows_amd64.zip network-debugger-web_windows_amd64.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger-web_windows_386.zip network-debugger-web_windows_386.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger-web_windows_arm64.zip network-debugger-web_windows_arm64.exe || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger_darwin_amd64.tar.gz network-debugger_darwin_amd64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger_darwin_arm64.tar.gz network-debugger_darwin_arm64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger-web_darwin_amd64.tar.gz network-debugger-web_darwin_amd64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger-web_darwin_arm64.tar.gz network-debugger-web_darwin_arm64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger_linux_amd64.tar.gz network-debugger_linux_amd64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger_linux_arm64.tar.gz network-debugger_linux_arm64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger-web_linux_amd64.tar.gz network-debugger-web_linux_amd64 || true && \
	tar -C . -czf ../$(PUBLISH_DIR)/downloads/network-debugger-web_linux_arm64.tar.gz network-debugger-web_linux_arm64 || true

.PHONY: package-docs
package-docs:
	$(MAKE) package PUBLISH_DIR=docs

frontend-dev-macos:
	cd frontend && flutter run -d macos

frontend-dev-windows:
	cd frontend && flutter run -d windows

