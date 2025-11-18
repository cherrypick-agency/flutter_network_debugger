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

      // Запускаем Go сервер и показываем отдельный диалог с логами старта
      final serverConfig = GoServerConfig(
        apiPort: config.apiPort,
        forwardProxyPort: config.forwardProxyPort,
      );

      if (!context.mounted) {
        return null;
      }

      // Диалог старта сервера с живыми логами и явной кнопкой Continue
      final startResult = await showDialog<bool>(
        context: context,
        barrierDismissible: false,
        builder: (ctx) => ServerStartupDialog(
          config: serverConfig,
          serverManager: _serverManager,
        ),
      );

      // null — пользователь отменил запуск полностью
      if (startResult == null) {
        return null;
      }

      // true — сервер запущен, продолжаем инициализацию
      if (startResult) {
        return config.apiPort;
      }

      // false — пользователь нажал Back в диалоге логов,
      // повторяем цикл и снова показываем диалог выбора портов
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

/// Диалог, который показывает логи старта Go сервера и его текущий статус.
/// Пользователь видит живой вывод stdout/stderr и вручную нажимает Continue,
/// когда сервер успешно запустился.
class ServerStartupDialog extends StatefulWidget {
  final GoServerConfig config;
  final GoServerManager serverManager;

  const ServerStartupDialog({
    required this.config,
    required this.serverManager,
    super.key,
  });

  @override
  State<ServerStartupDialog> createState() => _ServerStartupDialogState();
}

class _ServerStartupDialogState extends State<ServerStartupDialog> {
  late ServerStatus _status;
  final List<String> _logs = [];
  StreamSubscription<ServerStatus>? _statusSub;
  StreamSubscription<String>? _logSub;
  bool _startRequested = false;
  bool _startFinished = false;
  bool _startSuccess = false;

  @override
  void initState() {
    super.initState();
    // Стартовое состояние и последние логи (если есть)
    _status = widget.serverManager.status;
    _logs.addAll(widget.serverManager.recentLogs);

    _statusSub = widget.serverManager.statusStream.listen((s) {
      setState(() {
        _status = s;
      });
    });
    _logSub = widget.serverManager.logStream.listen((line) {
      setState(() {
        _logs.add(line);
        // Оставляем только последние ~200 строк, чтобы список не раздувался
        if (_logs.length > 200) {
          _logs.removeRange(0, _logs.length - 200);
        }
      });
    });

    _startServer();
  }

  Future<void> _startServer() async {
    if (_startRequested) {
      return;
    }
    _startRequested = true;

    final ok = await widget.serverManager.start(widget.config);
    if (!mounted) {
      return;
    }
    setState(() {
      _startFinished = true;
      _startSuccess = ok && widget.serverManager.status == ServerStatus.running;
    });
  }

  @override
  void dispose() {
    _statusSub?.cancel();
    _logSub?.cancel();
    super.dispose();
  }

  String _statusLabel() {
    switch (_status) {
      case ServerStatus.stopped:
        return 'Stopped';
      case ServerStatus.starting:
        return 'Starting...';
      case ServerStatus.running:
        return 'Running';
      case ServerStatus.stopping:
        return 'Stopping...';
      case ServerStatus.error:
        return 'Error';
    }
  }

  Color _statusColor(ColorScheme scheme) {
    switch (_status) {
      case ServerStatus.running:
        return scheme.primary;
      case ServerStatus.error:
        return scheme.error;
      case ServerStatus.starting:
      case ServerStatus.stopping:
        return scheme.secondary;
      case ServerStatus.stopped:
        return scheme.outline;
    }
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final canContinue = _status == ServerStatus.running && _startSuccess;
    final hasError =
        _status == ServerStatus.error || (_startFinished && !_startSuccess);

    return AlertDialog(
      backgroundColor: Theme.of(
        context,
      ).dialogTheme.backgroundColor?.withValues(alpha: 0.97),
      title: Row(
        children: [
          Icon(
            hasError
                ? Icons.error
                : (canContinue ? Icons.check_circle : Icons.hourglass_empty),
            color: hasError
                ? scheme.error
                : (canContinue ? scheme.primary : scheme.secondary),
          ),
          const SizedBox(width: 12),
          const Expanded(child: Text('Starting Go server')),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: _statusColor(scheme).withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              _statusLabel(),
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: _statusColor(scheme),
              ),
            ),
          ),
        ],
      ),
      content: SizedBox(
        width: 520,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'API: ${widget.config.apiPort}, Proxy: ${widget.config.forwardProxyPort}',
              style: const TextStyle(fontSize: 13),
            ),
            const SizedBox(height: 8),
            if (hasError && widget.serverManager.lastError != null) ...[
              Text(
                'Last error:',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: scheme.error,
                ),
              ),
              const SizedBox(height: 4),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: scheme.error.withValues(alpha: 0.06),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(
                    color: scheme.error.withValues(alpha: 0.4),
                  ),
                ),
                child: SelectableText(
                  widget.serverManager.lastError!,
                  style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
                ),
              ),
              const SizedBox(height: 8),
            ],
            Text(
              'Startup logs:',
              style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 4),
            Container(
              constraints: const BoxConstraints(maxHeight: 260, minHeight: 140),
              width: double.infinity,
              decoration: BoxDecoration(
                color: scheme.surfaceVariant.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(4),
                border: Border.all(
                  color: scheme.outline.withValues(alpha: 0.6),
                ),
              ),
              child: Scrollbar(
                thumbVisibility: true,
                child: ListView.builder(
                  padding: const EdgeInsets.all(8),
                  itemCount: _logs.isEmpty ? 1 : _logs.length,
                  itemBuilder: (context, index) {
                    const baseStyle = TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 12,
                    );
                    if (_logs.isEmpty) {
                      return Text(
                        'No output yet...',
                        style: baseStyle.copyWith(
                          color: scheme.outline.withValues(alpha: 0.9),
                        ),
                      );
                    }
                    final line = _logs[index];
                    final lower = line.toLowerCase();

                    TextStyle style;

                    // Простая подсветка: ошибки, warn и ключевые фразы старта
                    if (lower.contains('error') ||
                        lower.contains('[stderr]') ||
                        lower.contains('"level":"error"')) {
                      style = baseStyle.copyWith(
                        color: scheme.error,
                        fontWeight: FontWeight.w600,
                      );
                    } else if (lower.contains('warn') ||
                        lower.contains('"level":"warn"')) {
                      style = baseStyle.copyWith(color: scheme.tertiary);
                    } else if (lower.contains('starting') ||
                        lower.contains('listening') ||
                        lower.contains('ready') ||
                        lower.contains('"addr"')) {
                      style = baseStyle.copyWith(color: scheme.primary);
                    } else {
                      style = baseStyle.copyWith(
                        color: scheme.onSurface.withValues(alpha: 0.9),
                      );
                    }

                    return Text(line, style: style);
                  },
                ),
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () async {
            // При отмене пробуем остановить сервер, если он успел стартовать
            if (_status == ServerStatus.starting ||
                _status == ServerStatus.running) {
              await widget.serverManager.stop();
            }
            if (!mounted) {
              return;
            }
            Navigator.of(context).pop(null);
          },
          child: const Text('Cancel'),
        ),
        if (hasError)
          ElevatedButton.icon(
            onPressed: () async {
              // На Back тоже стараемся не оставлять висящий процесс
              if (_status == ServerStatus.starting ||
                  _status == ServerStatus.running) {
                await widget.serverManager.stop();
              }
              if (!mounted) {
                return;
              }
              Navigator.of(context).pop(false);
            },
            icon: const Icon(Icons.arrow_back),
            label: const Text('Back'),
          ),
        if (canContinue)
          ElevatedButton.icon(
            onPressed: () {
              Navigator.of(context).pop(true);
            },
            icon: const Icon(Icons.check),
            label: const Text('Continue'),
          ),
      ],
    );
  }
}
