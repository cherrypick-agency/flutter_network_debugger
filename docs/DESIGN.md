# Design Doc (MVP)

Goals: Local WebSocket/Socket.IO proxy for debugging with minimal latency and clean API.

Flow: client <-> wsproxy <-> upstream (ws/wss). Proxy mirrors frames both ways, logs metadata in in-memory ring buffer, exposes REST for sessions/frames/events and WS monitor for live updates.

Entities: Session, Frame, Event. Usecase layer defines repositories and services. Adapters implement storage (memory) and decoders (Socket.IO v4, partial v3). Infrastructure exposes HTTP API, metrics, health.

Constraints: local dev only (no auth/TLS), redact sensitive logs, target filter optional, >=10k msg/min, <5ms average overhead.

Risks/Workarounds:
- Socket.IO decoding differences v4 vs v3: start with '42' event packets and best-effort parse. Add engine.io support later.
- Memory pressure: ring buffer with caps and TTL eviction.
- Backpressure: buffered monitor channel and drop on slow consumers.


