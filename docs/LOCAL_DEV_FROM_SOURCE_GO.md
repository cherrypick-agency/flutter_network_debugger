# Local development from source (Go + Web UI)

## Dev workflow (recommended for UI development)

Run Go backend and Flutter Web UI separately. This gives you fast hot-reload on
frontend without rebuilding and re-embedding `_web` on every change.

Terminal A (backend + proxy):

```bash
# API/UI backend on :9092, forward proxy on :9091
NO_BROWSER=1 DEV_MODE=1 go run ./cmd/network-debugger
```

Terminal B (frontend with hot-reload):

```bash
cd frontend
flutter pub get
flutter run -d chrome
# or: flutter run -d web-server --web-port 8080
```

Notes:
- Backend API base URL is `http://localhost:9092`.
- Forward proxy is `http://localhost:9091`.
- If you previously opened UI from `network-debugger-web`, your browser may have
  cached an old `base_url`. In that case clear site data for localhost.

## Build and run Go binaries

```bash
# server/desktop binary
go build -o ./network-debugger ./cmd/network-debugger
./network-debugger

# web version that opens browser automatically
go build -o ./network-debugger-web ./cmd/network-debugger-web
./network-debugger-web
```

## Important: embedded Web UI artifacts

`cmd/network-debugger-web` serves the Web UI from embedded artifacts in
`cmd/network-debugger-web/_web`. If you change the frontend (or you see an
outdated UI / "No connection to backend"), rebuild the frontend and copy the
build output before running `go run ./cmd/network-debugger-web`:

```bash
cd frontend
flutter pub get
flutter build web --release --web-renderer canvaskit
rm -rf ../cmd/network-debugger-web/_web
mkdir -p ../cmd/network-debugger-web/_web
cp -R build/web/* ../cmd/network-debugger-web/_web/
cd ..

go run ./cmd/network-debugger-web
```

