import 'dart:io';
import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../services/prefs.dart';
import '../../features/startup/startup_dialog.dart';
import '../go_server/go_server_manager.dart';

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

/// Виджет для отображения строки лога с подсветкой ключевых слов
class _LogLine extends StatelessWidget {
  final String line;
  final ColorScheme scheme;

  const _LogLine({required this.line, required this.scheme});

  @override
  Widget build(BuildContext context) {
    const baseStyle = TextStyle(fontFamily: 'monospace', fontSize: 12);
    final lower = line.toLowerCase();

    // [stderr] - серый
    final isStderr = lower.contains('[stderr]');
    // Успешные сообщения - синий
    final isSuccess =
        lower.contains('started') ||
        lower.contains('listening') ||
        lower.contains('ready') ||
        lower.contains('running') ||
        lower.contains('"addr"');
    // Предупреждения - оранжевый
    final isWarn =
        lower.contains('"level":"warn"') ||
        lower.contains('level=warn') ||
        lower.contains('warning');

    // Базовый цвет строки
    Color lineColor;
    if (isStderr) {
      lineColor = scheme.outline;
    } else if (isSuccess) {
      lineColor = scheme.primary;
    } else if (isWarn) {
      lineColor = scheme.tertiary;
    } else {
      lineColor = scheme.onSurface.withValues(alpha: 0.9);
    }

    // Паттерны для подсветки жирным красным
    final errorPatterns = RegExp(
      r'\b(error|fail|failed|cannot|unable|panic|fatal)\b',
      caseSensitive: false,
    );

    if (!errorPatterns.hasMatch(line)) {
      return Text(line, style: baseStyle.copyWith(color: lineColor));
    }

    // Разбиваем строку на части и подсвечиваем ключевые слова
    final spans = <TextSpan>[];
    int lastEnd = 0;

    for (final match in errorPatterns.allMatches(line)) {
      if (match.start > lastEnd) {
        spans.add(
          TextSpan(
            text: line.substring(lastEnd, match.start),
            style: baseStyle.copyWith(color: lineColor),
          ),
        );
      }
      spans.add(
        TextSpan(
          text: match.group(0),
          style: baseStyle.copyWith(
            color: scheme.error,
            fontWeight: FontWeight.bold,
          ),
        ),
      );
      lastEnd = match.end;
    }

    if (lastEnd < line.length) {
      spans.add(
        TextSpan(
          text: line.substring(lastEnd),
          style: baseStyle.copyWith(color: lineColor),
        ),
      );
    }

    return Text.rich(TextSpan(children: spans));
  }
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
  final ScrollController _logsScrollController = ScrollController();
  bool _autoScrollEnabled = true;

  @override
  void initState() {
    super.initState();
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
        if (_logs.length > 200) {
          _logs.removeRange(0, _logs.length - 200);
        }
      });
      _scrollToBottomIfEnabled();
    });

    _startServer();
  }

  void _scrollToBottomIfEnabled() {
    if (!_autoScrollEnabled || !_logsScrollController.hasClients) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_logsScrollController.hasClients) {
        _logsScrollController.animateTo(
          _logsScrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 150),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _copyAllLogs() {
    final text = _logs.join('\n');
    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Logs copied to clipboard'),
        duration: Duration(seconds: 2),
      ),
    );
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
    _logsScrollController.dispose();
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
            Row(
              children: [
                const Text(
                  'Startup logs:',
                  style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600),
                ),
                const Spacer(),
                IconButton(
                  icon: Icon(
                    _autoScrollEnabled
                        ? Icons.vertical_align_bottom
                        : Icons.vertical_align_center,
                    size: 16,
                  ),
                  onPressed: () {
                    setState(() {
                      _autoScrollEnabled = !_autoScrollEnabled;
                    });
                    if (_autoScrollEnabled) {
                      _scrollToBottomIfEnabled();
                    }
                  },
                  tooltip: _autoScrollEnabled
                      ? 'Disable auto-scroll'
                      : 'Auto-scroll to bottom',
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(
                    minWidth: 28,
                    minHeight: 28,
                  ),
                ),
                const SizedBox(width: 4),
                IconButton(
                  icon: const Icon(Icons.copy, size: 16),
                  onPressed: _logs.isEmpty ? null : _copyAllLogs,
                  tooltip: 'Copy all logs',
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(
                    minWidth: 28,
                    minHeight: 28,
                  ),
                ),
              ],
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
              child: NotificationListener<ScrollNotification>(
                onNotification: (notification) {
                  // Отключаем автоскролл если пользователь скроллит вверх
                  if (notification is ScrollUpdateNotification) {
                    final maxScroll =
                        _logsScrollController.position.maxScrollExtent;
                    final currentScroll = _logsScrollController.position.pixels;
                    if (maxScroll - currentScroll > 50) {
                      if (_autoScrollEnabled) {
                        setState(() => _autoScrollEnabled = false);
                      }
                    }
                  }
                  return false;
                },
                child: Scrollbar(
                  controller: _logsScrollController,
                  thumbVisibility: true,
                  child: SelectionArea(
                    child: ListView.builder(
                      controller: _logsScrollController,
                      padding: const EdgeInsets.all(8),
                      itemCount: _logs.isEmpty ? 1 : _logs.length,
                      itemBuilder: (context, index) {
                        if (_logs.isEmpty) {
                          return Text(
                            'No output yet...',
                            style: TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 12,
                              color: scheme.outline.withValues(alpha: 0.9),
                            ),
                          );
                        }
                        return _LogLine(line: _logs[index], scheme: scheme);
                      },
                    ),
                  ),
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
