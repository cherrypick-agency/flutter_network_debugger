SHELL := /bin/bash
API_PORT ?= 9092
PROXY_PORT ?= 9091

# Build metadata
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X network-debugger/internal/infrastructure/observability.Version=$(VERSION) -X network-debugger/internal/infrastructure/observability.Commit=$(COMMIT) -X network-debugger/internal/infrastructure/observability.Date=$(DATE)

# Release helper metadata (used by release-github target)
TAG              := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
FRONTEND_VERSION := $(shell echo $(TAG) | sed -E 's/^v//')
CURRENT_BUILD    := $(shell grep '^version:' frontend/pubspec.yaml | sed -E 's/.*\+([0-9]+).*/\1/')
BUILD_NUM        := $(shell echo $$(($(CURRENT_BUILD) + 1)))

# Where to publish packaged artifacts by default
PUBLISH_DIR ?= web

build:
	cd cmd/network-debugger && go build -race -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger

run: build
	API_PORT=$(API_PORT) FORWARD_PROXY_DEFAULT_PORT=$(PROXY_PORT) ./bin/network-debugger

# Development auto-reload using Air (https://github.com/cosmtrek/air)
# Install: go install github.com/cosmtrek/air@latest
dev:
	@if ! command -v air >/dev/null 2>&1; then echo "air not found. Install with: go install github.com/air-verse/air@latest"; exit 1; fi
	API_PORT=$(API_PORT) FORWARD_PROXY_DEFAULT_PORT=$(PROXY_PORT) DEV_MODE=1 air -c .air.toml

tidy:
	go mod tidy

# Code formatting
.PHONY: fmt fmt-go fmt-dart install-hooks
fmt: fmt-go fmt-dart

fmt-go:
	@echo "Formatting Go files..."
	@gofmt -w .
	@go fmt ./...
	@echo "✓ Go files formatted"

fmt-dart:
	@echo "Formatting Dart files..."
	@if command -v dart >/dev/null 2>&1; then \
		find . -name "*.dart" -not -path "*/.*" -not -path "*/build/*" | xargs dart format; \
		echo "✓ Dart files formatted"; \
	else \
		echo "⚠ dart not found, skipping Dart formatting"; \
	fi

install-hooks:
	@./scripts/install-git-hooks.sh

test:
	go test ./...

coverage:
	@bash scripts/coverage.sh

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
	cd frontend && flutter build web --release --no-tree-shake-icons --no-wasm-dry-run
	rm -rf cmd/network-debugger-web/_web
	mkdir -p cmd/network-debugger-web/_web
	cp -R frontend/build/web/* cmd/network-debugger-web/_web/

build-app:
	$(MAKE) frontend-build-web
	cd cmd/network-debugger-web && go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web

run-app: build-app
	API_PORT=$(API_PORT) FORWARD_PROXY_DEFAULT_PORT=$(PROXY_PORT) ./bin/network-debugger-web

win-app:
	$(MAKE) frontend-build-web
	cd cmd/network-debugger-web && GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_amd64.exe
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
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../../bin/network-debugger-web_windows_arm64.exe

# Packaging
.PHONY: package
package: build-cross build-app-cross
	@mkdir -p $(PUBLISH_DIR)/downloads
	cd bin && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger_windows_amd64.zip network-debugger_windows_amd64.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger_windows_arm64.zip network-debugger_windows_arm64.exe || true && \
	zip -q ../$(PUBLISH_DIR)/downloads/network-debugger-web_windows_amd64.zip network-debugger-web_windows_amd64.exe || true && \
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

.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=0.1.7"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "✗ VERSION='$(VERSION)' не похож на семвер X.Y.Z"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "✗ Working tree не пустой. Сначала закоммить изменения или сделай stash."; \
		exit 1; \
	fi
	@echo "→ Обновляем версию Flutter frontend..."
	@echo "   frontend/pubspec.yaml: version: $(VERSION)+$(BUILD_NUM)"
	@sed -i.bak -E 's/^version: .*/version: $(VERSION)+$(BUILD_NUM)/' frontend/pubspec.yaml && rm -f frontend/pubspec.yaml.bak
	@git add frontend/pubspec.yaml
	@echo "→ Создаём коммит и тег v$(VERSION)..."
	@git commit -m "chore(release): v$(VERSION)"
	@git tag v$(VERSION)
	@echo "→ Пушим в origin..."
	@git push origin HEAD
	@git push origin v$(VERSION)
	@echo ""
	@echo "✓ Release v$(VERSION) подготовлен и запушен."
	@echo "GitHub Actions уже соберут и опубликуют артефакты по тегу."

frontend-dev-macos:
	cd frontend && flutter run -d macos

frontend-dev-windows:
	cd frontend && flutter run -d windows

frontend-dev-linux:
	cd frontend && flutter run -d linux

# Desktop application packaging
.PHONY: desktop-macos desktop-windows desktop-linux desktop-all

desktop-macos:
	@echo "Building macOS desktop app..."
	@chmod +x scripts/package-macos.sh
	VERSION=$(VERSION) ./scripts/package-macos.sh
	@echo "✓ macOS DMG created in dist/"

desktop-windows:
	@echo "Building Windows desktop app..."
	powershell -ExecutionPolicy Bypass -File scripts/package-windows.ps1 -Version "$(VERSION)"
	@echo "✓ Windows ZIP created in dist/"

desktop-linux:
	@echo "Building Linux desktop app..."
	@chmod +x scripts/package-linux.sh
	VERSION=$(VERSION) ARCH=amd64 ./scripts/package-linux.sh
	@echo "✓ Linux packages created in dist/"

desktop-all:
	@echo "Building desktop apps for all platforms..."
	@echo "Note: This requires running on each platform separately or using CI/CD"
	@echo "Run 'make desktop-macos' on macOS"
	@echo "Run 'make desktop-windows' on Windows"
	@echo "Run 'make desktop-linux' on Linux"
	@echo "Or push a version tag to trigger GitHub Actions workflow"

# Desktop development helpers
.PHONY: desktop-clean desktop-flutter-clean

desktop-clean:
	@echo "Cleaning desktop build artifacts..."
	rm -rf dist/
	rm -rf frontend/build/macos/
	rm -rf frontend/build/windows/
	rm -rf frontend/build/linux/
	@echo "✓ Desktop artifacts cleaned"

desktop-flutter-clean:
	@echo "Running Flutter clean..."
	cd frontend && flutter clean
	cd frontend && flutter pub get
	@echo "✓ Flutter cleaned and dependencies updated"

