import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:logging/logging.dart';
import 'go_server_path.dart';

enum ServerStatus { stopped, starting, running, stopping, error }

class GoServerConfig {
  final int apiPort;
  final int forwardProxyPort;
  final String? dataDir;

  GoServerConfig({
    required this.apiPort,
    required this.forwardProxyPort,
    this.dataDir,
  });

  List<String> toArgs() {
    return [
      '--api-port=$apiPort',
      '--proxy-port=$forwardProxyPort',
      if (dataDir != null) '--data-dir=$dataDir',
    ];
  }
}

class GoServerManager {
  final _log = Logger('GoServerManager');
  Process? _process;
  ServerStatus _status = ServerStatus.stopped;
  final _statusController = StreamController<ServerStatus>.broadcast();
  final _logController = StreamController<String>.broadcast();
  String? _lastError;
  final List<String> _recentLogs = [];

  ServerStatus get status => _status;
  Stream<ServerStatus> get statusStream => _statusController.stream;
  Stream<String> get logStream => _logController.stream;
  String? get lastError => _lastError;
  List<String> get recentLogs => List.unmodifiable(_recentLogs);

  /// Starts Go server with specified configuration
  Future<bool> start(GoServerConfig config) async {
    if (_status == ServerStatus.running || _status == ServerStatus.starting) {
      _log.warning('Server is already running or starting');
      return false;
    }

    try {
      _updateStatus(ServerStatus.starting);
      _lastError = null;
      _recentLogs.clear();

      final serverPath = await getGoServerPath();
      if (serverPath == null) {
        _lastError = 'Go server binary not found in application bundle';
        throw Exception(_lastError);
      }

      _log.info('Starting Go server at: $serverPath');
      _log.info(
        'Config: API port=${config.apiPort}, Proxy port=${config.forwardProxyPort}',
      );

      // For desktop app, disable auto-opening browser
      // in bundled Go server via NO_BROWSER environment variable.
      final environment = <String, String>{
        ...Platform.environment,
        'NO_BROWSER': '1',
      };

      _process = await Process.start(
        serverPath,
        config.toArgs(),
        environment: environment,
        mode: ProcessStartMode.normal,
      );

      // Listen to stdout for logs and readiness detection
      _process!.stdout
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen(
            (line) {
              _logController.add('[STDOUT] $line');
              _log.info('Server: $line');
              _recentLogs.add(line);
              if (_recentLogs.length > 50) _recentLogs.removeAt(0);

              // Check server readiness by logs
              if (line.contains('started') ||
                  line.contains('listening') ||
                  line.contains('ready')) {
                if (_status == ServerStatus.starting) {
                  _updateStatus(ServerStatus.running);
                }
              }
            },
            onError: (error) {
              _log.severe('stdout error: $error');
              _lastError = 'STDOUT error: $error';
            },
          );

      // Listen to stderr for errors
      _process!.stderr
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen(
            (line) {
              _logController.add('[STDERR] $line');
              _log.warning('Server error: $line');
              _recentLogs.add('[ERROR] $line');
              if (_recentLogs.length > 50) _recentLogs.removeAt(0);

              // Save last error
              if (line.toLowerCase().contains('error') ||
                  line.toLowerCase().contains('failed') ||
                  line.toLowerCase().contains('address already in use') ||
                  line.toLowerCase().contains('permission denied')) {
                _lastError = line;
              }
            },
            onError: (error) {
              _log.severe('stderr error: $error');
              _lastError = 'STDERR error: $error';
            },
          );

      // Monitor process completion
      _process!.exitCode.then((exitCode) {
        _log.info('Server exited with code: $exitCode');
        // Don't change status if graceful shutdown is in progress
        if (_status != ServerStatus.stopping) {
          if (exitCode != 0) {
            _updateStatus(ServerStatus.error);
          } else {
            _updateStatus(ServerStatus.stopped);
          }
        }
        _process = null;
      });

      // Check health endpoint with retry logic (exponential backoff)
      // Attempts: 500ms, 1s, 2s
      bool isHealthy = false;
      for (final delayMs in [500, 1000, 2000]) {
        await Future.delayed(Duration(milliseconds: delayMs));

        isHealthy = await _checkHealth(config.apiPort);
        if (isHealthy) {
          _updateStatus(ServerStatus.running);
          _log.info('Server is healthy and ready');
          return true;
        }

        _log.fine('Health check attempt failed, retrying...');
      }

      // If health check failed but process is alive, still consider it running
      if (_process != null) {
        _updateStatus(ServerStatus.running);
        _log.warning('Server started but health check failed after retries');
        return true;
      } else {
        _lastError ??= 'Server process died unexpectedly';
        _updateStatus(ServerStatus.error);
        return false;
      }
    } catch (e, stack) {
      _log.severe('Failed to start server', e, stack);
      _lastError ??= e.toString();
      _updateStatus(ServerStatus.error);
      return false;
    }
  }

  /// Stops Go server
  Future<void> stop() async {
    if (_status == ServerStatus.stopped || _status == ServerStatus.stopping) {
      return;
    }

    if (_process == null) {
      _updateStatus(ServerStatus.stopped);
      return;
    }

    try {
      _updateStatus(ServerStatus.stopping);
      _log.info('Stopping Go server...');

      // Graceful shutdown via SIGTERM
      _process!.kill(ProcessSignal.sigterm);

      // Wait for process completion (max 5 sec)
      final exitCode = await _process!.exitCode.timeout(
        const Duration(seconds: 5),
        onTimeout: () {
          _log.warning('Server did not stop gracefully, forcing kill');
          _process!.kill(ProcessSignal.sigkill);
          return -1;
        },
      );

      _log.info('Server stopped with exit code: $exitCode');
      _process = null;
      _updateStatus(ServerStatus.stopped);
    } catch (e, stack) {
      _log.severe('Error stopping server', e, stack);
      _process = null;
      _updateStatus(ServerStatus.error);
    }
  }

  /// Restarts server with new configuration
  Future<bool> restart(GoServerConfig config) async {
    await stop();
    await Future.delayed(const Duration(milliseconds: 500));
    return await start(config);
  }

  /// Checks server health endpoint
  Future<bool> _checkHealth(int apiPort) async {
    try {
      final client = HttpClient();
      final request = await client.getUrl(
        Uri.parse('http://localhost:$apiPort/_health'),
      );
      final response = await request.close().timeout(
        const Duration(seconds: 2),
      );
      await response.drain();
      client.close();
      return response.statusCode == 200;
    } catch (e) {
      _log.fine('Health check failed: $e');
      return false;
    }
  }

  void _updateStatus(ServerStatus newStatus) {
    if (_status != newStatus) {
      _status = newStatus;
      _statusController.add(_status);
    }
  }

  /// Resource cleanup
  Future<void> dispose() async {
    try {
      await stop();
    } finally {
      await _statusController.close();
      await _logController.close();
    }
  }
}
