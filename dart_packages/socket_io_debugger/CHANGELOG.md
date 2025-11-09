# Changelog

## 1.0.0
- **BREAKING CHANGE**: Migrated to `socket_io_client` ^3.0.0
- **BREAKING CHANGE**: Now supports Socket.IO server v4.7+ (Engine.IO v4)
- **IMPORTANT**: For Socket.IO server v2.*/v3.*/v4.6 compatibility, use `socket_io_debugger: ^0.1.0` instead
- Updated: Minimum Dart SDK remains `>=3.0.0 <4.0.0`
- Compatible: socket_io_client v3.x API (no code changes required in user code)
- Docs: Added version compatibility information to README

## 0.1.2
- Docs: Complete README rewrite with comprehensive documentation
- Docs: Added "Part of network_debugger ecosystem" reference
- Docs: Added Features, Installation, and "Starting the Proxy" sections
- Docs: Added Quick start with full code example
- Docs: Added Advanced options (reverse/forward proxy, Android emulator)
- Docs: Added Configuration examples (--dart-define, ENV)
- Docs: Added Environment variables reference
- Docs: Added cross-references to network_debugger package

## 0.1.1
- Docs: small fixes in examples and configuration description.

## 0.1.0
- Initial release.
- One-call proxy attachment for Socket.IO client (reverse/forward).
- Env/--dart-define configuration, forward proxy overrides via HttpClient.


