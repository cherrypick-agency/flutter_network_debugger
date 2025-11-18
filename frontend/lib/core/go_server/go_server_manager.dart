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

  /// Запускает Go сервер с указанной конфигурацией
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

      // Для десктопного приложения отключаем авто‑открытие браузера
      // у bundled Go-сервера через переменную окружения NO_BROWSER.
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

      // Слушаем stdout для логов и определения готовности
      _process!.stdout
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen(
            (line) {
              _logController.add('[STDOUT] $line');
              _log.info('Server: $line');
              _recentLogs.add(line);
              if (_recentLogs.length > 50) _recentLogs.removeAt(0);

              // Проверяем готовность сервера по логам
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

      // Слушаем stderr для ошибок
      _process!.stderr
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen(
            (line) {
              _logController.add('[STDERR] $line');
              _log.warning('Server error: $line');
              _recentLogs.add('[ERROR] $line');
              if (_recentLogs.length > 50) _recentLogs.removeAt(0);

              // Сохраняем последнюю ошибку
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

      // Мониторим завершение процесса
      _process!.exitCode.then((exitCode) {
        _log.info('Server exited with code: $exitCode');
        // Не меняем статус если идет graceful shutdown
        if (_status != ServerStatus.stopping) {
          if (exitCode != 0) {
            _updateStatus(ServerStatus.error);
          } else {
            _updateStatus(ServerStatus.stopped);
          }
        }
        _process = null;
      });

      // Проверяем health endpoint с retry logic (exponential backoff)
      // Попытки: 500ms, 1s, 2s
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

      // Если health check не прошел, но процесс жив, все равно считаем running
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

  /// Останавливает Go сервер
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

      // Graceful shutdown через SIGTERM
      _process!.kill(ProcessSignal.sigterm);

      // Ждем завершения процесса (макс 5 сек)
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

  /// Перезапускает сервер с новой конфигурацией
  Future<bool> restart(GoServerConfig config) async {
    await stop();
    await Future.delayed(const Duration(milliseconds: 500));
    return await start(config);
  }

  /// Проверяет health endpoint сервера
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

  /// Очистка ресурсов
  Future<void> dispose() async {
    try {
      await stop();
    } finally {
      await _statusController.close();
      await _logController.close();
    }
  }
}
