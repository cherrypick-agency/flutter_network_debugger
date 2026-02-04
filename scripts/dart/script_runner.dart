import 'dart:io';
import 'dart:convert';

/// Dart Script Runner for Network Debugger
///
/// This is a JSON-RPC 2.0 server that executes Dart scripts via stdin/stdout.
/// It's used by the DartExecutor to run user-provided Dart code.
///
/// Protocol: JSON-RPC 2.0
/// Transport: stdin/stdout
///
/// Supported methods:
/// - execute: Runs user-provided Dart code with input
void main() async {
  // JSON-RPC server over stdin/stdout
  final stdinStream = stdin
      .transform(utf8.decoder)
      .transform(const LineSplitter())
      .map((line) {
        try {
          return json.decode(line) as Map<String, dynamic>;
        } catch (e) {
          stderr.writeln('Failed to parse JSON: $e');
          return null;
        }
      })
      .where((request) => request != null)
      .cast<Map<String, dynamic>>();

  await for (final request in stdinStream) {
    try {
      final method = request['method'] as String?;
      final params = request['params'] as Map<String, dynamic>?;
      final id = request['id'];

      if (method == 'ping') {
        final response = {
          'jsonrpc': '2.0',
          'result': {'pong': true},
          'id': id,
        };
        stdout.writeln(json.encode(response));
        continue;
      }

      if (method == 'execute' && params != null) {
        final code = params['code'] as String;
        final input = params['input'] as String;

        try {
          // Execute Dart code
          final result = await executeScript(code, input);

          // Send success response
          final response = {
            'jsonrpc': '2.0',
            'result': {'output': result, 'logs': <String>[]},
            'id': id,
          };

          stdout.writeln(json.encode(response));
        } catch (e, stack) {
          // Send error response
          final response = {
            'jsonrpc': '2.0',
            'error': {
              'code': -32000,
              'message': e.toString(),
              'data': stack.toString(),
            },
            'id': id,
          };

          stdout.writeln(json.encode(response));
        }
      } else {
        // Method not supported
        final response = {
          'jsonrpc': '2.0',
          'error': {'code': -32601, 'message': 'Method not found'},
          'id': id,
        };

        stdout.writeln(json.encode(response));
      }
    } catch (e) {
      stderr.writeln('Error processing request: $e');
    }
  }
}

/// Executes user-provided Dart code
///
/// For MVP: Simple subprocess execution for E2E testing
/// For production: Integrate dart_eval or use Isolate-based execution.
///
/// This implementation writes the code to a temp file and runs it as a subprocess.
/// The script receives input via stdin and should return a JSON-RPC response via stdout.
Future<String> executeScript(String code, String input) async {
  // Write code to temporary file
  final tempDir = await Directory.systemTemp.createTemp('dart_script_');
  final scriptFile = File('${tempDir.path}/script.dart');
  Process? child;
  var childExited = false;
  final stderrLines = <String>[];

  try {
    await scriptFile.writeAsString(code);

    // Run the script as a subprocess
    child = await Process.start('dart', [scriptFile.path]);

    // Important: always read stderr, otherwise the process may hang when the buffer fills up.
    // Store only the last N lines for diagnostics.
    child.stderr.transform(utf8.decoder).transform(const LineSplitter()).listen(
      (line) {
        stderrLines.add(line);
        if (stderrLines.length > 50) {
          stderrLines.removeAt(0);
        }
      },
    );

    // Parse the input (should be ScriptContext JSON)
    final inputData = json.decode(input) as Map<String, dynamic>;

    // Send JSON-RPC request to the script
    final rpcRequest = {
      'jsonrpc': '2.0',
      'method': 'execute',
      'params': inputData,
      'id': 1,
    };

    child.stdin.writeln(json.encode(rpcRequest));
    await child.stdin.close();

    // Read response with timeout
    final output = await readFirstLineLimited(
      child.stdout,
      maxBytes: 1024 * 1024,
    ).timeout(const Duration(seconds: 5));

    // Wait for process to complete
    await child.exitCode.timeout(const Duration(seconds: 5));
    childExited = true;

    // Parse the response and extract the result
    final rpcResponse = json.decode(output) as Map<String, dynamic>;

    if (rpcResponse['result'] != null) {
      // Return the modified request/response
      return json.encode(rpcResponse['result']);
    } else if (rpcResponse['error'] != null) {
      return json.encode({
        'error': rpcResponse['error']['message'] ?? 'Script execution error',
      });
    } else {
      return json.encode({'error': 'Invalid response from script'});
    }
  } catch (e, stack) {
    stderr.writeln('[script_runner] Error executing script: $e');
    stderr.writeln('[script_runner] Stack: $stack');
    if (stderrLines.isNotEmpty) {
      stderr.writeln(
        '[script_runner] Child stderr (last ${stderrLines.length} lines):',
      );
      for (final line in stderrLines) {
        stderr.writeln('[script_runner] > $line');
      }
    }
    return json.encode({'error': 'Script execution failed: $e'});
  } finally {
    if (child != null && !childExited) {
      try {
        // If the script hung/didn't respond — must kill the process, otherwise there will be "zombie" processes and resource leaks.
        child.kill();
        await child.exitCode.timeout(const Duration(seconds: 1));
      } catch (_) {}
    }

    // Cleanup temp directory
    try {
      await tempDir.delete(recursive: true);
    } catch (e) {
      stderr.writeln('[script_runner] Failed to cleanup temp dir: $e');
    }
  }
}

Future<String> readFirstLineLimited(
  Stream<List<int>> stream, {
  required int maxBytes,
}) async {
  final buffer = <int>[];
  var sawNewline = false;

  await for (final chunk in stream) {
    for (final b in chunk) {
      if (buffer.length >= maxBytes) {
        throw Exception('stdout exceeds ${maxBytes} bytes');
      }
      if (b == 10) {
        sawNewline = true;
        break;
      }
      buffer.add(b);
    }
    if (sawNewline) {
      break;
    }
  }

  return utf8.decode(buffer, allowMalformed: true);
}
