import 'dart:io';
import 'package:flutter/material.dart';
import '../../services/prefs.dart';
import '../../features/startup/startup_dialog.dart';
import '../go_server/go_server_manager.dart';
import 'dart:async';

/// Bootstrap service для desktop приложений
/// Управляет запуском Go сервера и инициализацией приложения
class DesktopBootstrap {
  static final GoServerManager _serverManager = GoServerManager();

  /// Проверяет, является ли текущая платформа desktop
  static bool isDesktop() {
    return Platform.isMacOS || Platform.isWindows || Platform.isLinux;
  }

  /// Загружает сохраненную конфигурацию портов из preferences
  static Future<StartupConfig> _loadSavedConfig() async {
    try {
      final prefs = await PrefsService().load();
      // Пытаемся получить сохраненные порты из настроек
      // Если нет - возвращаем defaults
      final apiPort = int.tryParse(prefs['apiPort'] ?? '') ?? 9092;
      final proxyPort = int.tryParse(prefs['proxyPort'] ?? '') ?? 9091;

      return StartupConfig(apiPort: apiPort, forwardProxyPort: proxyPort);
    } catch (_) {
      return StartupConfig.getDefaults();
    }
  }

  /// Проверяет что сервер действительно работает
  static Future<bool> _verifyServerHealth(int apiPort) async {
    try {
      final client = HttpClient();
      client.connectionTimeout = const Duration(seconds: 3);

      final request = await client.getUrl(
        Uri.parse('http://localhost:$apiPort/_health'),
      );

      final response = await request.close().timeout(
        const Duration(seconds: 3),
      );

      await response.drain();
      client.close();

      return response.statusCode == 200;
    } catch (e) {
      return false;
    }
  }

  /// Сохраняет конфигурацию портов в preferences
  static Future<void> _saveConfig(StartupConfig config) async {
    try {
      final prefs = await PrefsService().load();
      prefs['apiPort'] = config.apiPort.toString();
      prefs['proxyPort'] = config.forwardProxyPort.toString();
      // Сохраняем обратно
      await PrefsService().save(
        baseUrl: 'http://localhost:${config.apiPort}',
        targetWs: prefs['targetWs'] ?? '',
        q: prefs['q'] ?? '',
        targetFilter: prefs['targetFilter'] ?? '',
        opcode: prefs['opcode'] ?? 'all',
        direction: prefs['direction'] ?? 'all',
        namespace: prefs['namespace'] ?? '',
        httpMethod: prefs['httpMethod'] ?? 'any',
        httpStatus: prefs['httpStatus'] ?? 'any',
        httpMime: prefs['httpMime'] ?? '',
        httpMinDurationMs: int.tryParse(prefs['httpMinDuration'] ?? '0') ?? 0,
        groupBy: prefs['groupBy'] ?? 'none',
        headerKey: prefs['headerKey'] ?? '',
        headerVal: prefs['headerVal'] ?? '',
        respDelayEnabled: (prefs['respDelayEnabled'] ?? 'false') == 'true',
        respDelayValue: prefs['respDelayValue'] ?? '',
      );
    } catch (_) {
      // Ignore save errors
    }
  }

  /// Показывает startup dialog и запускает сервер
  /// Возвращает порт API для использования в setupDI
  static Future<int?> bootstrap(BuildContext context) async {
    if (!isDesktop()) {
      // Для web возвращаем default port
      return 9092;
    }

    // Цикл для возможности повторной попытки
    // ignore: use_build_context_synchronously
    while (true) {
      // Загружаем сохраненную конфигурацию
      final savedConfig = await _loadSavedConfig();

      // Показываем startup dialog
      final config = await showStartupDialog(
        context,
        initialConfig: savedConfig,
      );

      if (config == null) {
        // Пользователь отменил запуск
        return null;
      }

      // Сохраняем конфигурацию для следующего раза
      await _saveConfig(config);

      // Если сервер уже запущен, проверяем что он действительно работает
      if (config.serverAlreadyRunning) {
        // Показываем индикатор проверки
        if (context.mounted) {
          showDialog(
            context: context,
            barrierDismissible: false,
            builder: (ctx) => const Center(
              child: Card(
                child: Padding(
                  padding: EdgeInsets.all(24.0),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      CircularProgressIndicator(),
                      SizedBox(height: 16),
                      Text('Verifying server connection...'),
                    ],
                  ),
                ),
              ),
            ),
          );
        }

        // Проверяем здоровье сервера
        final isHealthy = await _verifyServerHealth(config.apiPort);

        // Закрываем индикатор
        if (context.mounted) {
          Navigator.of(context, rootNavigator: true).pop();
        }

        if (!isHealthy) {
          // Сервер не отвечает - показываем ошибку
          if (context.mounted) {
            final retry = await showDialog<bool>(
              context: context,
              barrierDismissible: false,
              builder: (ctx) => AlertDialog(
                title: const Row(
                  children: [
                    Icon(Icons.warning, color: Colors.orange),
                    SizedBox(width: 12),
                    Text('Server Not Responding'),
                  ],
                ),
                content: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Cannot connect to server on the specified port.',
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 16),
                    Text('API Port: ${config.apiPort}'),
                    const SizedBox(height: 8),
                    const Text(
                      'Make sure the Go server is running with:',
                      style: TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.grey.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: Colors.grey.withValues(alpha: 0.3),
                        ),
                      ),
                      child: SelectableText(
                        'go run cmd/network-debugger/main.go --api-port=${config.apiPort} --proxy-port=${config.forwardProxyPort}',
                        style: const TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 11,
                        ),
                      ),
                    ),
                  ],
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.of(ctx).pop(false),
                    child: const Text('Cancel'),
                  ),
                  ElevatedButton.icon(
                    onPressed: () => Navigator.of(ctx).pop(true),
                    icon: const Icon(Icons.arrow_back),
                    label: const Text('Back'),
                  ),
                ],
              ),
            );

            if (retry != true) {
              return null;
            }
            // Продолжаем цикл - покажем startup dialog снова
            continue;
          }
        }

        // Сервер работает - возвращаем порт
        return config.apiPort;
      }

      // Показываем индикатор загрузки
      if (context.mounted) {
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (ctx) => const Center(
            child: Card(
              child: Padding(
                padding: EdgeInsets.all(24.0),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    CircularProgressIndicator(),
                    SizedBox(height: 16),
                    Text('Starting server...'),
                  ],
                ),
              ),
            ),
          ),
        );
      }

      // Запускаем Go сервер
      final serverConfig = GoServerConfig(
        apiPort: config.apiPort,
        forwardProxyPort: config.forwardProxyPort,
      );

      final started = await _serverManager.start(serverConfig);

      // Закрываем индикатор загрузки
      if (context.mounted) {
        Navigator.of(context, rootNavigator: true).pop();
      }

      if (!started) {
        // Получаем детали ошибки
        final errorMessage = _serverManager.lastError ?? 'Unknown error';
        final recentLogs = _serverManager.recentLogs;

        // Показываем ошибку
        if (context.mounted) {
          final retry = await showDialog<bool>(
            context: context,
            barrierDismissible: false,
            builder: (ctx) => AlertDialog(
              title: const Row(
                children: [
                  Icon(Icons.error, color: Colors.red),
                  SizedBox(width: 12),
                  Text('Server Start Failed'),
                ],
              ),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Failed to start the Go server.',
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 16),
                    const Text('Error:'),
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.red.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: Colors.red.withValues(alpha: 0.3),
                        ),
                      ),
                      child: SelectableText(
                        errorMessage,
                        style: const TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 12,
                        ),
                      ),
                    ),
                    if (recentLogs.isNotEmpty) ...[
                      const SizedBox(height: 16),
                      const Text('Recent logs:'),
                      const SizedBox(height: 8),
                      Container(
                        constraints: const BoxConstraints(maxHeight: 200),
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.grey.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: Colors.grey.withValues(alpha: 0.3),
                          ),
                        ),
                        child: SingleChildScrollView(
                          child: SelectableText(
                            recentLogs.join('\n'),
                            style: const TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 11,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(false),
                  child: const Text('Cancel'),
                ),
                ElevatedButton.icon(
                  onPressed: () => Navigator.of(ctx).pop(true),
                  icon: const Icon(Icons.arrow_back),
                  label: const Text('Back'),
                ),
              ],
            ),
          );

          if (retry != true) {
            return null;
          }
          // Иначе продолжаем цикл и показываем startup dialog снова
        }
      } else {
        // Сервер успешно запущен
        return config.apiPort;
      }
    }
  }

  /// Останавливает сервер при выходе из приложения
  static Future<void> shutdown() async {
    await _serverManager.stop();
  }

  /// Получить статус сервера
  static ServerStatus get serverStatus => _serverManager.status;

  /// Получить поток статуса сервера
  static Stream<ServerStatus> get serverStatusStream =>
      _serverManager.statusStream;

  /// Получить поток логов сервера
  static Stream<String> get serverLogStream => _serverManager.logStream;
}
