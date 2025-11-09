// Simple Dart script for E2E testing - logs requests and returns them unchanged
import 'dart:convert';
import 'dart:io';

void main() {
  stderr.writeln('[Dart E2E] Starting simple logger script');

  stdin.transform(utf8.decoder).transform(const LineSplitter()).listen((line) {
    try {
      final rpcRequest = jsonDecode(line);
      final params = rpcRequest['params'];

      if (params['request'] != null) {
        final req = params['request'];

        // Log for debugging
        stderr.writeln(
          '[Dart E2E] Processing request: ${req['method']} ${req['url']}',
        );

        // Add header to mark that Dart script processed it
        req['headers'] ??= {};
        req['headers']['X-Dart-Processed'] = ['true'];
        req['headers']['X-Dart-Test'] = ['E2E'];

        // Return ScriptResult format expected by ScriptService
        final rpcResponse = {
          'jsonrpc': '2.0',
          'id': rpcRequest['id'],
          'result': {'modified': true, 'modifiedRequest': req},
        };

        stdout.writeln(jsonEncode(rpcResponse));
      } else if (params['response'] != null) {
        // Handle response modification (if needed)
        final resp = params['response'];

        stderr.writeln('[Dart E2E] Processing response: ${resp['status']}');

        // Return ScriptResult format for response
        final rpcResponse = {
          'jsonrpc': '2.0',
          'id': rpcRequest['id'],
          'result': {'modified': false, 'modifiedResponse': resp},
        };

        stdout.writeln(jsonEncode(rpcResponse));
      }
    } catch (e, stack) {
      stderr.writeln('[Dart E2E] Error: $e');
      stderr.writeln('[Dart E2E] Stack: $stack');
    }
  });
}
