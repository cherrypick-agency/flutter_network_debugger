import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart' as ws_io;
import 'dart:io' as io;
import 'dart:convert';
import 'theme/app_theme.dart';
import 'theme/context_ext.dart';
import 'features/http_inspector/presentation/widgets/http_details_panel.dart';
import 'features/ws_inspector/presentation/widgets/ws_details_panel.dart';
import 'package:flutter/services.dart';
import 'dart:async';
import 'services/prefs.dart';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'features/inspector/application/stores/sessions_store.dart';
import 'features/inspector/application/stores/session_details_store.dart';
import 'features/inspector/application/stores/aggregate_store.dart';
import 'features/inspector/presentation/widgets/waterfall_timeline.dart';
import 'features/inspector/presentation/widgets/waterfall_timeline_fullscreen.dart';
import 'features/inspector/presentation/widgets/timeline_settings_button.dart';
import 'core/di/di.dart';
import 'core/notifications/notifications_service.dart';
import 'core/notifications/notification_snackbar.dart';
import 'core/network/connectivity_banner.dart';
import 'core/notifications/notification.dart';
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;
import 'widgets/strike_painter.dart';
import 'features/hotkeys/presentation/hotkeys_settings_page.dart';
import 'core/hotkeys/hotkeys_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await setupDI(baseUrl: 'http://localhost:9091');
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});
  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  ThemeMode _mode = ThemeMode.system;

  @override
  void initState() {
    super.initState();
    _loadTheme();
  }

  Future<void> _loadTheme() async {
    final m = await PrefsService().loadThemeModeString();
    setState(() {
      _mode = _fromString(m);
    });
  }

  ThemeMode _fromString(String s) {
    switch (s) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }

  Future<void> _toggleTheme() async {
    setState(() {
      if (_mode == ThemeMode.light)
        _mode = ThemeMode.dark;
      else if (_mode == ThemeMode.dark)
        _mode = ThemeMode.system;
      else
        _mode = ThemeMode.light;
    });
    await PrefsService().saveThemeModeString(
      _mode == ThemeMode.light
          ? 'light'
          : _mode == ThemeMode.dark
          ? 'dark'
          : 'system',
    );
  }

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        Provider<SessionsStore>.value(value: sl<SessionsStore>()),
        Provider<SessionDetailsStore>.value(value: sl<SessionDetailsStore>()),
        Provider<AggregateStore>.value(value: sl<AggregateStore>()),
      ],
      child: Builder(
        builder: (context) {
          return MaterialApp(
            theme: buildLightTheme(),
            darkTheme: buildDarkTheme(),
            themeMode: _mode,
            routes: {'/hotkeys': (_) => const HotkeysSettingsPage()},
            home: MyHomePage(onToggleTheme: _toggleTheme),
          );
        },
      ),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, this.onToggleTheme});
  final Future<void> Function()? onToggleTheme;
  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final TextEditingController _sessionSearchCtrl = TextEditingController();
  final TextEditingController _sessionTargetFilterCtrl =
      TextEditingController();
  final TextEditingController _namespaceFilterCtrl = TextEditingController();
  WebSocketChannel? _monitor;
  final List<String> _monitorLog = [];
  // sessions from store
  String? _selectedSessionId;
  List<dynamic> _frames = [];
  List<dynamic> _events = [];
  // legacy fields removed
  bool _loadingSessions = false;
  // legacy loaders removed
  final ScrollController _framesCtrl = ScrollController();
  final ScrollController _eventsCtrl = ScrollController();
  String _opcodeFilter = 'all';
  String _directionFilter = 'all';
  Timer? _pollTimer;
  Timer? _sessionsReloadDebounce;
  final FocusNode _searchFocus = FocusNode();
  // HTTP quick-filters
  String _httpMethodFilter = 'any';
  String _httpStatusFilter = 'any';
  String _httpMimeFilter = '';
  int _httpMinDurationMs = 0;
  // Cached HTTP meta by session
  final Map<String, Map<String, dynamic>> _httpMeta = {};
  String _groupBy = 'none';
  DateTimeRange? _selectedRange;
  final TextEditingController _headerKeyCtrl = TextEditingController();
  final TextEditingController _headerValCtrl = TextEditingController();
  bool _showFilters = false;
  final Set<String> _selectedDomains = <String>{};
  bool _hideHeartbeats = false;
  bool _wfFitAll = true;
  // Показывать только сессии, начавшиеся после очистки
  DateTime? _since;

  @override
  void initState() {
    super.initState();
    _connectMonitor();
    _restorePrefs();
    _restoreMonitorLog();
    _loadSessions();
    _framesCtrl.addListener(_onFramesScroll);
    _eventsCtrl.addListener(_onEventsScroll);
  }

  @override
  void dispose() {
    _monitor?.sink.close();
    _pollTimer?.cancel();
    _sessionsReloadDebounce?.cancel();
    _framesCtrl.dispose();
    _eventsCtrl.dispose();
    _searchFocus.dispose();
    super.dispose();
  }

  void _connectMonitor() {
    // Без логирования ошибок: UI уже показывает статус
    final wsUrl = _wsURL(
      sl<http_client.AppHttpClient>().defaultHost,
      '/_api/v1/monitor/ws',
    );
    Future<void>(() async {
      try {
        final sock = await io.WebSocket.connect(
          wsUrl,
        ).timeout(const Duration(seconds: 3));
        final ch = ws_io.IOWebSocketChannel(sock);
        setState(() {
          _monitor = ch;
        });
        _monitor!.stream.listen(
          (msg) {
            final s = msg.toString();
            setState(() {
              _monitorLog.insert(0, s);
            });
            _persistMonitorLogDebounced();
            try {
              final Map<String, dynamic> ev = jsonDecode(s);
              final t = (ev['type'] ?? '').toString();
              if (t == 'session_started' || t == 'session_ended') {
                // Если только что чистили — дадим UI стабилизироваться и не дергать загрузку
                if (_loadingSessions) return;
                _scheduleSessionsReload();
              }
              if (t == 'frame_added' ||
                  t == 'event_added' ||
                  t == 'sio_probe') {
                final sid = (ev['id'] ?? '').toString();
                if (_selectedSessionId != null && sid == _selectedSessionId) {
                  _tickRefresh();
                }
              }
            } catch (_) {}
          },
          onError: (_) {},
          onDone: () {
            // silent reconnect later
          },
        );
      } catch (_) {
        // swallow; визуальный индикатор уже показывает отсутствие связи
      }
    });
  }

  Timer? _persistDebounce;
  void _persistMonitorLogDebounced() {
    _persistDebounce?.cancel();
    _persistDebounce = Timer(const Duration(milliseconds: 400), () async {
      await PrefsService().saveMonitorLog(_monitorLog);
    });
  }

  Future<void> _restoreMonitorLog() async {
    final saved = await PrefsService().loadMonitorLog();
    if (saved.isNotEmpty) {
      setState(() {
        _monitorLog
          ..clear()
          ..addAll(saved);
      });
    }
  }

  void _scheduleSessionsReload() {
    _sessionsReloadDebounce?.cancel();
    _sessionsReloadDebounce = Timer(const Duration(milliseconds: 300), () {
      if (!_loadingSessions) {
        _loadSessions();
      }
    });
  }

  Future<void> _clearAllSessions() async {
    final ok = await showDialog<bool>(
      context: context,
      builder:
          (ctx) => AlertDialog(
            title: const Text('Clear all sessions?'),
            content: const Text(
              'This will remove all sessions from backend and UI.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(false),
                child: const Text('Cancel'),
              ),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: Theme.of(context).colorScheme.error,
                ),
                onPressed: () => Navigator.of(ctx).pop(true),
                child: const Text('Clear'),
              ),
            ],
          ),
    );
    if (ok != true) return;
    try {
      // ignore: invalid_use_of_protected_member
      final client = sl.get<Object>();
      bool cleared = false;
      try {
        await (client as dynamic).delete(path: '/_api/v1/sessions');
        cleared = true;
      } catch (_) {}
      if (!cleared) {
        try {
          await (client as dynamic).delete(path: '/api/sessions');
          cleared = true;
        } catch (_) {}
      }
      if (!cleared) {
        // fallback: iteratively delete known sessions
        final items = context.read<SessionsStore>().items.toList();
        for (final s in items) {
          try {
            await (client as dynamic).delete(path: '/_api/v1/sessions/${s.id}');
          } catch (_) {}
          try {
            await (client as dynamic).delete(path: '/api/sessions/${s.id}');
          } catch (_) {}
        }
      }
    } catch (_) {}

    // reset local state immediately
    setState(() {
      _selectedSessionId = null;
      _frames.clear();
      _events.clear();
      _selectedRange = null;
      _wfFitAll = true; // let timeline start from now and fit all
      _since = DateTime.now().toUtc(); // водораздел для UI
    });
    try {
      await PrefsService().saveSince(_since!);
    } catch (_) {}
    // clear stores immediately for instant UI reset
    try {
      context.read<SessionsStore>().clear();
    } catch (_) {}
    try {
      context.read<AggregateStore>().clear();
    } catch (_) {}
    // сразу перезагрузим агрегатор доменов, чтобы счётчики обнулились мгновенно
    try {
      await context.read<AggregateStore>().load(groupBy: 'domain');
    } catch (_) {}
    // temporarily suppress auto-reload caused by monitor events
    _sessionsReloadDebounce?.cancel();
    _loadingSessions = true;
    await Future.delayed(const Duration(milliseconds: 300));
    _loadingSessions = false;
    // final reload to confirm empty state
    await _loadSessions();
  }

  String _wsURL(String base, String path) {
    var b = base;
    if (b.startsWith('http://')) {
      b = 'ws://${b.substring(7)}';
    }
    if (b.startsWith('https://')) {
      b = 'wss://${b.substring(8)}';
    }
    if (!path.startsWith('/')) path = '/$path';
    if (b.endsWith('/')) {
      b = b.substring(0, b.length - 1);
    }
    return '$b$path';
  }

  Future<void> _loadSessions() async {
    final store = context.read<SessionsStore>();
    final q = _sessionSearchCtrl.text.trim();
    final target = _sessionTargetFilterCtrl.text.trim();
    await store.load(q: q, target: target);
    _suckMetaFromSessions();
  }

  // ignore: unused_element
  Future<void> _warmupHttpMeta() async {
    // limit to first 50 to avoid overload
    final list = context.read<SessionsStore>().items.take(50).toList();
    final client = sl.get<Object>();
    for (final s in list) {
      final id = s.id;
      if (_httpMeta.containsKey(id)) continue;
      try {
        // ignore: invalid_use_of_protected_member
        final r = await (client as dynamic).get(
          path: '/api/sessions/$id/frames',
          query: {'limit': '2'},
        );
        final m = <String, dynamic>{};
        final data =
            (r.data is Map<String, dynamic>)
                ? (r.data as Map<String, dynamic>)
                : {};
        final items =
            (data['items'] as List<dynamic>).cast<Map<String, dynamic>>();
        Map<String, dynamic>? req;
        Map<String, dynamic>? resp;
        DateTime? tReq;
        DateTime? tResp;
        for (final f in items) {
          final p = f['preview']?.toString() ?? '';
          try {
            final mp = jsonDecode(p) as Map<String, dynamic>;
            if (mp['type'] == 'http_request') {
              req = mp;
              tReq = DateTime.tryParse(f['ts']?.toString() ?? '');
            }
            if (mp['type'] == 'http_response') {
              resp = mp;
              tResp = DateTime.tryParse(f['ts']?.toString() ?? '');
            }
          } catch (_) {}
        }
        if (req != null) m['method'] = (req['method'] ?? '').toString();
        if (resp != null) {
          m['status'] = (resp['status'] ?? '').toString();
          final headers =
              (resp['headers'] as Map?)?.map(
                (k, v) => MapEntry(k.toString(), v.toString()),
              ) ??
              {};
          final ctEntry = headers.entries.firstWhere(
            (e) => e.key.toLowerCase() == 'content-type',
            orElse: () => const MapEntry('', ''),
          );
          m['mime'] = ctEntry.value;
          final upg =
              headers.entries
                  .firstWhere(
                    (e) => e.key.toLowerCase() == 'upgrade',
                    orElse: () => const MapEntry('', ''),
                  )
                  .value;
          m['streaming'] =
              (m['mime']?.toString().contains('text/event-stream') ?? false) ||
              (upg.toString().toLowerCase() == 'websocket');
          m['headers'] = headers;
        }
        if (tReq != null && tResp != null) {
          m['durationMs'] = tResp.difference(tReq).inMilliseconds;
        }
        if (m.isNotEmpty) {
          setState(() {
            _httpMeta[id] = m;
          });
        }
      } catch (_) {}
    }
  }

  Future<void> _restorePrefs() async {
    final data = await PrefsService().load();
    setState(() {
      _sessionSearchCtrl.text = data['q']!;
      _sessionTargetFilterCtrl.text = data['targetFilter']!;
      _opcodeFilter = data['opcode']!;
      _directionFilter = data['direction']!;
      _namespaceFilterCtrl.text = data['namespace']!;
      _httpMethodFilter = data['httpMethod'] ?? 'any';
      _httpStatusFilter = data['httpStatus'] ?? 'any';
      _httpMimeFilter = data['httpMime'] ?? '';
      _httpMinDurationMs = int.tryParse(data['httpMinDuration'] ?? '0') ?? 0;
      _groupBy = data['groupBy'] ?? 'none';
      _headerKeyCtrl.text = data['headerKey'] ?? '';
      _headerValCtrl.text = data['headerVal'] ?? '';
    });
    // restore since-ts if any
    try {
      _since = await PrefsService().loadSince();
    } catch (_) {}
  }

  Future<void> _savePrefs() async {
    await PrefsService().save(
      baseUrl: sl<http_client.AppHttpClient>().defaultHost,
      targetWs: '',
      q: _sessionSearchCtrl.text,
      targetFilter: _sessionTargetFilterCtrl.text,
      opcode: _opcodeFilter,
      direction: _directionFilter,
      namespace: _namespaceFilterCtrl.text,
      httpMethod: _httpMethodFilter,
      httpStatus: _httpStatusFilter,
      httpMime: _httpMimeFilter,
      httpMinDurationMs: _httpMinDurationMs,
      groupBy: _groupBy,
      headerKey: _headerKeyCtrl.text,
      headerVal: _headerValCtrl.text,
    );
  }

  Future<void> _loadDetails(String id) async {
    final details = context.read<SessionDetailsStore>();
    await details.open(id);
    _startAutoRefresh();
  }

  void _onFramesScroll() {
    if (_framesCtrl.position.pixels >=
        _framesCtrl.position.maxScrollExtent - 200) {
      context.read<SessionDetailsStore>().loadMoreFrames();
    }
  }

  void _onEventsScroll() {
    if (_eventsCtrl.position.pixels >=
        _eventsCtrl.position.maxScrollExtent - 200) {
      context.read<SessionDetailsStore>().loadMoreEvents();
    }
  }

  // no-op retained for compatibility with shortcuts (kept for future keyboard shortcuts)

  void _startAutoRefresh() {
    _pollTimer?.cancel();
    if (_selectedSessionId == null) return;
    _pollTimer = Timer.periodic(
      const Duration(seconds: 2),
      (_) => _tickRefresh(),
    );
  }

  Future<void> _tickRefresh() async {
    if (_selectedSessionId == null) return;
    try {
      await Future.wait([
        context.read<SessionDetailsStore>().loadMoreFrames(),
        context.read<SessionDetailsStore>().loadMoreEvents(),
      ]);
    } catch (_) {}
  }

  // фильтры перенесены в WsDetailsPanel

  Future<void> _deleteSelected() async {
    if (_selectedSessionId == null) return;
    final id = _selectedSessionId!;
    // ignore: invalid_use_of_protected_member
    final client = sl.get<Object>();
    try {
      await (client as dynamic).delete(path: '/_api/v1/sessions/$id');
    } catch (_) {}
    setState(() {
      _selectedSessionId = null;
      _frames.clear();
      _events.clear();
    });
    await _loadSessions();
  }

  @override
  Widget build(BuildContext context) {
    final hk = sl<HotkeysService>();
    final globalHandlers = hk.buildHandlers({
      'sessions.refresh': _loadSessions,
      'sessions.refresh.ctrl': _loadSessions,
      'sessions.focusSearch': () {
        _searchFocus.requestFocus();
      },
      'sessions.delete': _deleteSelected,
    });
    return CallbackShortcuts(
      bindings: globalHandlers,
      child: Focus(
        autofocus: true,
        child: Scaffold(
          /*
      appBar: AppBar(
        // title: const Text('go-proxy Console'), 
        actions: [
        IconButton(onPressed: (){
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Switch theme in system settings (current: ${Theme.of(context).brightness.name})')));
        }, icon: const Icon(Icons.brightness_6))
      ]),
      */
          body: Stack(
            children: [
              Column(
                children: [
                  // Connectivity banner outside of padded content to avoid outer gaps
                  ConnectivityBanner(
                    baseUrl: () => sl<http_client.AppHttpClient>().defaultHost,
                  ),
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Column(
                        children: [
                          // Top controls
                          Theme(
                            data: Theme.of(context).copyWith(
                              elevatedButtonTheme: ElevatedButtonThemeData(
                                style: ButtonStyle(
                                  textStyle: MaterialStatePropertyAll(
                                    const TextStyle(fontSize: 12),
                                  ),
                                  padding: MaterialStatePropertyAll(
                                    const EdgeInsets.symmetric(
                                      horizontal: 12,
                                      vertical: 8,
                                    ),
                                  ),
                                ),
                              ),
                              textButtonTheme: TextButtonThemeData(
                                style: ButtonStyle(
                                  textStyle: MaterialStatePropertyAll(
                                    const TextStyle(fontSize: 12),
                                  ),
                                  padding: MaterialStatePropertyAll(
                                    const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 6,
                                    ),
                                  ),
                                ),
                              ),
                            ),
                            child: Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: [
                                // ElevatedButton(onPressed: _loadSessions, child: const Text('Refresh')),
                                IconButton(
                                  onPressed: () {
                                    setState(() {
                                      _showFilters = !_showFilters;
                                    });
                                  },
                                  tooltip: 'Filters',
                                  icon: const Icon(Icons.filter_list),
                                ),
                                IconButton(
                                  onPressed: widget.onToggleTheme,
                                  tooltip: 'Theme',
                                  icon: const Icon(Icons.brightness_6),
                                ),
                                IconButton(
                                  onPressed: () {
                                    Navigator.of(context).pushNamed('/hotkeys');
                                  },
                                  tooltip: 'Hotkeys',
                                  icon: const Icon(Icons.keyboard),
                                ),
                              ],
                            ),
                          ),
                          // Waterfall timeline
                          Padding(
                            padding: const EdgeInsets.only(top: 8),
                            child: Stack(
                              children: [
                                Observer(
                                  builder: (_) {
                                    final raw =
                                        context
                                            .watch<SessionsStore>()
                                            .items
                                            .toList();
                                    final sessions =
                                        _since == null
                                            ? raw
                                            : raw
                                                .where((s) {
                                                  final st = s.startedAt;
                                                  return st == null ||
                                                      !st.isBefore(_since!);
                                                })
                                                .toList(growable: false);
                                    return WaterfallTimeline(
                                      sessions: sessions,
                                      fitAll: _wfFitAll,
                                      onFitAllChanged: (v) {
                                        setState(() {
                                          _wfFitAll = v;
                                        });
                                      },
                                      onIntervalSelected: (range) {
                                        setState(() {
                                          _selectedRange = range;
                                        });
                                      },
                                      onSessionSelected: (s) {
                                        setState(() {
                                          _selectedSessionId = s.id;
                                        });
                                        _loadDetails(s.id);
                                      },
                                      initialRange:
                                          _since == null
                                              ? null
                                              : DateTimeRange(
                                                start: _since!,
                                                end: DateTime.now(),
                                              ),
                                    );
                                  },
                                ),
                                Positioned(
                                  right: 0,
                                  bottom: 0,
                                  child: Container(
                                    decoration: BoxDecoration(
                                      color: Theme.of(
                                        context,
                                      ).colorScheme.surface.withOpacity(0.45),
                                      borderRadius: BorderRadius.circular(10),
                                      border: Border.all(
                                        color:
                                            Theme.of(
                                              context,
                                            ).colorScheme.outlineVariant,
                                        width: 1,
                                      ),
                                    ),
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 4,
                                      vertical: 2,
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          tooltip: 'Clear all sessions',
                                          iconSize: 16,
                                          padding: EdgeInsets.zero,
                                          constraints: const BoxConstraints(
                                            minWidth: 28,
                                            minHeight: 28,
                                          ),
                                          icon: Icon(
                                            Icons.delete_forever,
                                            size: 16,
                                            color:
                                                Theme.of(
                                                  context,
                                                ).colorScheme.error,
                                          ),
                                          onPressed: () async {
                                            await _clearAllSessions();
                                          },
                                        ),
                                        SizedBox(
                                          width: 28,
                                          height: 28,
                                          child: FittedBox(
                                            child: TimelineSettingsButton(
                                              getFit: () => _wfFitAll,
                                              setFit: (v) {
                                                setState(() {
                                                  _wfFitAll = v;
                                                });
                                              },
                                            ),
                                          ),
                                        ),
                                        IconButton(
                                          tooltip: 'Open fullscreen',
                                          iconSize: 16,
                                          padding: EdgeInsets.zero,
                                          constraints: const BoxConstraints(
                                            minWidth: 28,
                                            minHeight: 28,
                                          ),
                                          icon: const Icon(
                                            Icons.fullscreen,
                                            size: 16,
                                          ),
                                          onPressed: () async {
                                            final res = await Navigator.of(
                                              context,
                                            ).push<dynamic>(
                                              MaterialPageRoute(
                                                builder:
                                                    (_) =>
                                                        WaterfallTimelineFullscreenPage(
                                                          initialRange:
                                                              _selectedRange,
                                                        ),
                                              ),
                                            );
                                            if (res is String) {
                                              setState(() {
                                                _selectedSessionId = res;
                                              });
                                              _loadDetails(res);
                                            }
                                          },
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 8),
                          if (_showFilters)
                            Theme(
                              data: Theme.of(context).copyWith(
                                inputDecorationTheme:
                                    const InputDecorationTheme(
                                      isDense: true,
                                      contentPadding: EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 4,
                                      ),
                                      labelStyle: TextStyle(fontSize: 12),
                                    ),
                              ),
                              child: Wrap(
                                spacing: 8,
                                runSpacing: 8,
                                children: [
                                  SizedBox(
                                    width: 300,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      focusNode: _searchFocus,
                                      controller: _sessionSearchCtrl,
                                      decoration: const InputDecoration(
                                        labelText: 'Search sessions (q)',
                                      ),
                                      onChanged: (_) {
                                        _savePrefs();
                                        _scheduleSessionsReload();
                                      },
                                      onSubmitted: (_) {
                                        _savePrefs();
                                        _loadSessions();
                                      },
                                    ),
                                  ),
                                  SizedBox(
                                    width: 300,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      controller: _sessionTargetFilterCtrl,
                                      decoration: const InputDecoration(
                                        labelText: 'Filter by target',
                                      ),
                                      onChanged: (_) {
                                        _savePrefs();
                                        _scheduleSessionsReload();
                                      },
                                      onSubmitted: (_) {
                                        _savePrefs();
                                        _loadSessions();
                                      },
                                    ),
                                  ),
                                  // HTTP quick filters (client-side)
                                  DropdownButton<String>(
                                    value: _httpMethodFilter,
                                    isDense: true,
                                    iconSize: 18,
                                    style: TextStyle(
                                      fontSize: 12,
                                      color:
                                          Theme.of(
                                            context,
                                          ).colorScheme.onSurface,
                                    ),
                                    items: const [
                                      DropdownMenuItem(
                                        value: 'any',
                                        child: Text(
                                          'Any method',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'GET',
                                        child: Text(
                                          'GET',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'POST',
                                        child: Text(
                                          'POST',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'PUT',
                                        child: Text(
                                          'PUT',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'DELETE',
                                        child: Text(
                                          'DELETE',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'PATCH',
                                        child: Text(
                                          'PATCH',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'OPTIONS',
                                        child: Text(
                                          'OPTIONS',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                    ],
                                    onChanged: (v) {
                                      setState(() {
                                        _httpMethodFilter = v ?? 'any';
                                      });
                                    },
                                  ),
                                  DropdownButton<String>(
                                    value: _httpStatusFilter,
                                    isDense: true,
                                    iconSize: 18,
                                    style: TextStyle(
                                      fontSize: 12,
                                      color:
                                          Theme.of(
                                            context,
                                          ).colorScheme.onSurface,
                                    ),
                                    items: const [
                                      DropdownMenuItem(
                                        value: 'any',
                                        child: Text(
                                          'Any status',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: '2xx',
                                        child: Text(
                                          '2xx',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: '3xx',
                                        child: Text(
                                          '3xx',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: '4xx',
                                        child: Text(
                                          '4xx',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: '5xx',
                                        child: Text(
                                          '5xx',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                    ],
                                    onChanged: (v) {
                                      setState(() {
                                        _httpStatusFilter = v ?? 'any';
                                      });
                                    },
                                  ),
                                  SizedBox(
                                    width: 200,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      decoration: const InputDecoration(
                                        labelText: 'MIME contains',
                                      ),
                                      onChanged: (v) {
                                        setState(() {
                                          _httpMimeFilter = v;
                                        });
                                      },
                                    ),
                                  ),
                                  SizedBox(
                                    width: 120,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      keyboardType: TextInputType.number,
                                      decoration: const InputDecoration(
                                        labelText: 'Min ms',
                                      ),
                                      onChanged: (v) {
                                        setState(() {
                                          _httpMinDurationMs =
                                              int.tryParse(v) ?? 0;
                                        });
                                      },
                                    ),
                                  ),
                                  DropdownButton<String>(
                                    value: _groupBy,
                                    isDense: true,
                                    iconSize: 18,
                                    style: TextStyle(
                                      fontSize: 12,
                                      color:
                                          Theme.of(
                                            context,
                                          ).colorScheme.onSurface,
                                    ),
                                    items: const [
                                      DropdownMenuItem(
                                        value: 'none',
                                        child: Text(
                                          'No grouping',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'domain',
                                        child: Text(
                                          'Group by domain',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                      DropdownMenuItem(
                                        value: 'route',
                                        child: Text(
                                          'Group by route',
                                          style: TextStyle(fontSize: 12),
                                        ),
                                      ),
                                    ],
                                    onChanged: (v) {
                                      setState(() {
                                        _groupBy = v ?? 'none';
                                      });
                                    },
                                  ),
                                  SizedBox(
                                    width: 160,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      controller: _headerKeyCtrl,
                                      decoration: const InputDecoration(
                                        labelText: 'Header key',
                                      ),
                                      onChanged: (_) {
                                        _scheduleSessionsReload();
                                      },
                                    ),
                                  ),
                                  SizedBox(
                                    width: 180,
                                    child: TextField(
                                      style: const TextStyle(fontSize: 12),
                                      controller: _headerValCtrl,
                                      decoration: const InputDecoration(
                                        labelText: 'Header value',
                                      ),
                                      onChanged: (_) {
                                        _scheduleSessionsReload();
                                      },
                                    ),
                                  ),
                                  Observer(
                                    builder: (_) {
                                      final loading =
                                          context
                                              .watch<SessionsStore>()
                                              .loading;
                                      return loading
                                          ? const SizedBox(
                                            width: 24,
                                            height: 24,
                                            child: CircularProgressIndicator(
                                              strokeWidth: 2,
                                            ),
                                          )
                                          : const SizedBox.shrink();
                                    },
                                  ),
                                ],
                              ),
                            ),
                          const SizedBox(height: 12),
                          Expanded(
                            child: Row(
                              children: [
                                SizedBox(
                                  width: 360,
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        'Sessions',
                                        style: context.appText.title,
                                      ),
                                      const SizedBox(height: 6),
                                      // Domain chips under Sessions header as Wrap (max 3 rows, then scroll)
                                      Observer(
                                        builder: (_) {
                                          final agg =
                                              context.watch<AggregateStore>();
                                          if (!agg.loading &&
                                              agg.groups.isEmpty) {
                                            Future.microtask(
                                              () => agg.load(groupBy: 'domain'),
                                            );
                                          }
                                          return ConstrainedBox(
                                            constraints: const BoxConstraints(
                                              maxHeight: 48.0,
                                            ),
                                            child: SingleChildScrollView(
                                              child: Wrap(
                                                spacing: 6,
                                                runSpacing: 4,
                                                children: [
                                                  for (final g in agg.groups)
                                                    Builder(
                                                      builder: (context) {
                                                        final key =
                                                            (g['key'] ?? '')
                                                                .toString();
                                                        final selected =
                                                            _selectedDomains
                                                                .contains(key);
                                                        return Transform.scale(
                                                          scale: 0.69,
                                                          alignment:
                                                              Alignment
                                                                  .centerLeft,
                                                          child: ChoiceChip(
                                                            label: Text(
                                                              '$key (${g['count']})',
                                                              style:
                                                                  const TextStyle(
                                                                    fontSize:
                                                                        12,
                                                                  ),
                                                            ),
                                                            selected: selected,
                                                            visualDensity:
                                                                const VisualDensity(
                                                                  horizontal:
                                                                      -3,
                                                                  vertical: -3,
                                                                ),
                                                            materialTapTargetSize:
                                                                MaterialTapTargetSize
                                                                    .shrinkWrap,
                                                            onSelected: (_) {
                                                              setState(() {
                                                                if (selected) {
                                                                  _selectedDomains
                                                                      .remove(
                                                                        key,
                                                                      );
                                                                } else {
                                                                  _selectedDomains
                                                                      .add(key);
                                                                }
                                                              });
                                                            },
                                                          ),
                                                        );
                                                      },
                                                    ),
                                                ],
                                              ),
                                            ),
                                          );
                                        },
                                      ),
                                      const SizedBox(height: 8),
                                      Expanded(
                                        child: Observer(
                                          builder: (_) {
                                            final sessions = _visibleSessions();
                                            return ListView.builder(
                                              itemCount: sessions.length,
                                              itemBuilder: (ctx, i) {
                                                final s = sessions[i];
                                                final showHeader =
                                                    _groupBy != 'none' &&
                                                    (i == 0 ||
                                                        _groupKey(
                                                              sessions[i - 1],
                                                            ) !=
                                                            _groupKey(s));
                                                final header = _groupKey(s);
                                                // derive http meta or ws kind
                                                final meta =
                                                    (s.httpMeta ??
                                                        _httpMeta[s.id]) ??
                                                    const {};
                                                final method =
                                                    (meta['method'] ?? '')
                                                        .toString();
                                                final status =
                                                    int.tryParse(
                                                      (meta['status'] ?? '')
                                                          .toString(),
                                                    ) ??
                                                    0;
                                                final durationMs =
                                                    int.tryParse(
                                                      (meta['durationMs'] ?? '')
                                                          .toString(),
                                                    ) ??
                                                    -1;
                                                final cacheStatus =
                                                    (meta['cache']?['status'] ??
                                                            '')
                                                        .toString();
                                                // Пока нет ответа (status==0) не показываем CORS, рисуем loader
                                                final hasResponse = status > 0;
                                                final isClosed =
                                                    s.closedAt != null;
                                                final hasError =
                                                    (s.error ?? '').isNotEmpty;
                                                final corsOk =
                                                    hasResponse
                                                        ? ((meta['cors']?['ok'] ??
                                                                false) ==
                                                            true)
                                                        : true;
                                                final isWs =
                                                    (s.kind == 'ws') ||
                                                    (method.isEmpty &&
                                                        (s.kind == null));
                                                return Column(
                                                  crossAxisAlignment:
                                                      CrossAxisAlignment.start,
                                                  children: [
                                                    if (showHeader)
                                                      Padding(
                                                        padding:
                                                            const EdgeInsets.fromLTRB(
                                                              8,
                                                              12,
                                                              8,
                                                              4,
                                                            ),
                                                        child: Text(
                                                          header,
                                                          style: Theme.of(
                                                                context,
                                                              )
                                                              .textTheme
                                                              .labelMedium
                                                              ?.copyWith(
                                                                fontWeight:
                                                                    FontWeight
                                                                        .bold,
                                                              ),
                                                        ),
                                                      ),
                                                    InkWell(
                                                      onTap: () {
                                                        setState(() {
                                                          _selectedSessionId =
                                                              s.id;
                                                        });
                                                        _loadDetails(
                                                          _selectedSessionId!,
                                                        );
                                                      },
                                                      child: Container(
                                                        padding:
                                                            const EdgeInsets.symmetric(
                                                              horizontal: 8,
                                                              vertical: 8,
                                                            ),
                                                        decoration: BoxDecoration(
                                                          color:
                                                              _selectedSessionId ==
                                                                      s.id
                                                                  ? Theme.of(
                                                                        context,
                                                                      )
                                                                      .colorScheme
                                                                      .primary
                                                                      .withOpacity(
                                                                        0.06,
                                                                      )
                                                                  : null,
                                                          borderRadius:
                                                              BorderRadius.circular(
                                                                6,
                                                              ),
                                                        ),
                                                        child: Column(
                                                          crossAxisAlignment:
                                                              CrossAxisAlignment
                                                                  .start,
                                                          children: [
                                                            // URL (подсветка для проблемных транспортных ошибок: TIMEOUT/DNS/TLS)
                                                            Builder(
                                                              builder: (
                                                                context,
                                                              ) {
                                                                final errCode =
                                                                    (meta['errorCode'] ??
                                                                            '')
                                                                        .toString();
                                                                final warn =
                                                                    errCode ==
                                                                        'TIMEOUT' ||
                                                                    errCode ==
                                                                        'DNS' ||
                                                                    errCode ==
                                                                        'TLS';
                                                                final mark =
                                                                    Theme.of(
                                                                          context,
                                                                        )
                                                                        .colorScheme
                                                                        .tertiary;
                                                                final child = Text(
                                                                  s.target,
                                                                  maxLines: 3,
                                                                  overflow:
                                                                      TextOverflow
                                                                          .ellipsis,
                                                                  style: Theme.of(
                                                                        context,
                                                                      )
                                                                      .textTheme
                                                                      .bodyMedium
                                                                      ?.copyWith(
                                                                        fontFamily:
                                                                            'monospace',
                                                                      ),
                                                                );
                                                                if (!warn)
                                                                  return child;
                                                                return Container(
                                                                  padding:
                                                                      const EdgeInsets.symmetric(
                                                                        horizontal:
                                                                            4,
                                                                        vertical:
                                                                            2,
                                                                      ),
                                                                  decoration: BoxDecoration(
                                                                    color: mark
                                                                        .withOpacity(
                                                                          0.06,
                                                                        ),
                                                                    borderRadius:
                                                                        BorderRadius.circular(
                                                                          4,
                                                                        ),
                                                                    border: Border(
                                                                      left: BorderSide(
                                                                        color:
                                                                            mark,
                                                                        width:
                                                                            2,
                                                                      ),
                                                                    ),
                                                                  ),
                                                                  child: child,
                                                                );
                                                              },
                                                            ),
                                                            const SizedBox(
                                                              height: 4,
                                                            ),
                                                            Row(
                                                              crossAxisAlignment:
                                                                  CrossAxisAlignment
                                                                      .start,
                                                              children: [
                                                                Expanded(
                                                                  child: Wrap(
                                                                    spacing: 6,
                                                                    runSpacing:
                                                                        4,
                                                                    children: [
                                                                      if (isWs)
                                                                        _chip(
                                                                          (s.closedAt ==
                                                                                  null)
                                                                              ? 'WS open'
                                                                              : 'WS closed',
                                                                          backgroundColor:
                                                                              (s.closedAt ==
                                                                                      null)
                                                                                  ? Theme.of(
                                                                                    context,
                                                                                  ).colorScheme.secondaryContainer.withOpacity(
                                                                                    0.18,
                                                                                  )
                                                                                  : Theme.of(
                                                                                    context,
                                                                                  ).colorScheme.error.withOpacity(
                                                                                    0.12,
                                                                                  ),
                                                                          foregroundColor:
                                                                              (s.closedAt ==
                                                                                      null)
                                                                                  ? context.appColors.success
                                                                                  : Theme.of(
                                                                                    context,
                                                                                  ).colorScheme.error,
                                                                        ),
                                                                      if (!isWs &&
                                                                          method
                                                                              .isNotEmpty)
                                                                        _chip(
                                                                          method
                                                                              .toUpperCase(),
                                                                          backgroundColor:
                                                                              Theme.of(
                                                                                context,
                                                                              ).colorScheme.surfaceVariant,
                                                                          foregroundColor:
                                                                              Theme.of(
                                                                                context,
                                                                              ).colorScheme.onSurfaceVariant,
                                                                        ),
                                                                      if (!isWs &&
                                                                          status >
                                                                              0)
                                                                        _chip(
                                                                          'HTTP $status',
                                                                          backgroundColor: _statusBg(
                                                                            status,
                                                                          ),
                                                                          foregroundColor: _statusFg(
                                                                            status,
                                                                          ),
                                                                        ),
                                                                      if (!isWs &&
                                                                          !hasResponse &&
                                                                          isClosed &&
                                                                          hasError)
                                                                        Tooltip(
                                                                          message:
                                                                              s.error!,
                                                                          child: _chip(
                                                                            (() {
                                                                              final m =
                                                                                  (_httpMeta[s.id] ??
                                                                                          const {})
                                                                                      as Map<
                                                                                        String,
                                                                                        dynamic
                                                                                      >;
                                                                              final code =
                                                                                  (m['errorCode'] ??
                                                                                          '')
                                                                                      .toString();
                                                                              return code.isNotEmpty
                                                                                  ? code
                                                                                  : 'ERR';
                                                                            })(),
                                                                            backgroundColor: Theme.of(
                                                                              context,
                                                                            ).colorScheme.error.withOpacity(
                                                                              0.12,
                                                                            ),
                                                                            foregroundColor:
                                                                                Theme.of(
                                                                                  context,
                                                                                ).colorScheme.error,
                                                                          ),
                                                                        ),
                                                                      if (!isWs &&
                                                                          !hasResponse &&
                                                                          !isClosed)
                                                                        const SizedBox(
                                                                          width:
                                                                              14,
                                                                          height:
                                                                              14,
                                                                          child: CircularProgressIndicator(
                                                                            strokeWidth:
                                                                                2,
                                                                          ),
                                                                        ),
                                                                      if (!isWs &&
                                                                          durationMs >=
                                                                              0)
                                                                        _chip(
                                                                          '${durationMs} ms',
                                                                          backgroundColor: _durationBg(
                                                                            durationMs,
                                                                          ),
                                                                          foregroundColor: _durationFg(
                                                                            durationMs,
                                                                          ),
                                                                        ),
                                                                      if (!isWs &&
                                                                          cacheStatus
                                                                              .isNotEmpty)
                                                                        (cacheStatus.toUpperCase() ==
                                                                                'MISS')
                                                                            ? _chipStrike(
                                                                              'cache',
                                                                              backgroundColor:
                                                                                  Theme.of(
                                                                                    context,
                                                                                  ).colorScheme.surfaceVariant,
                                                                              foregroundColor:
                                                                                  Theme.of(
                                                                                    context,
                                                                                  ).colorScheme.onSurfaceVariant,
                                                                            )
                                                                            : _chip(
                                                                              'cache: ${cacheStatus.toUpperCase()}',
                                                                            ),
                                                                      if (!isWs &&
                                                                          !corsOk)
                                                                        _chip(
                                                                          'CORS',
                                                                          backgroundColor: Theme.of(
                                                                            context,
                                                                          ).colorScheme.error.withOpacity(
                                                                            0.12,
                                                                          ),
                                                                          foregroundColor:
                                                                              Theme.of(
                                                                                context,
                                                                              ).colorScheme.error,
                                                                        ),
                                                                    ],
                                                                  ),
                                                                ),
                                                                const SizedBox(
                                                                  width: 8,
                                                                ),
                                                                Column(
                                                                  crossAxisAlignment:
                                                                      CrossAxisAlignment
                                                                          .end,
                                                                  children: [
                                                                    Text(
                                                                      _formatTimeHMSSafe(
                                                                        s.startedAt,
                                                                      ),
                                                                      textAlign:
                                                                          TextAlign
                                                                              .right,
                                                                      style:
                                                                          Theme.of(
                                                                            context,
                                                                          ).textTheme.bodySmall,
                                                                    ),
                                                                    if (s.closedAt !=
                                                                        null)
                                                                      (() {
                                                                        final code =
                                                                            (meta['errorCode'] ??
                                                                                    '')
                                                                                .toString();
                                                                        if (code
                                                                            .isEmpty)
                                                                          return const SizedBox.shrink();
                                                                        return Text(
                                                                          'Closed ($code)',
                                                                          style: Theme.of(
                                                                            context,
                                                                          ).textTheme.labelSmall?.copyWith(
                                                                            color:
                                                                                Theme.of(
                                                                                  context,
                                                                                ).colorScheme.error,
                                                                          ),
                                                                        );
                                                                      })(),
                                                                  ],
                                                                ),
                                                              ],
                                                            ),
                                                          ],
                                                        ),
                                                      ),
                                                    ),
                                                    const Divider(height: 1),
                                                  ],
                                                );
                                              },
                                            );
                                          },
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const VerticalDivider(width: 1),
                                // если есть selectedSessionId, то отображаем details panel
                                if (_selectedSessionId != null)
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Expanded(
                                          child: Builder(
                                            builder: (context) {
                                              // Determine selected session kind for dynamic tabs
                                              bool selIsWs = true;
                                              bool selIsHttp =
                                                  true; // fallback both
                                              if (_selectedSessionId != null) {
                                                final items =
                                                    context
                                                        .watch<SessionsStore>()
                                                        .items
                                                        .toList();
                                                Map<String, dynamic>? meta;
                                                String? kind;
                                                for (final s in items) {
                                                  if (s.id ==
                                                      _selectedSessionId) {
                                                    meta =
                                                        (s.httpMeta ??
                                                                _httpMeta[s.id])
                                                            ?.cast<
                                                              String,
                                                              dynamic
                                                            >();
                                                    kind = s.kind;
                                                    break;
                                                  }
                                                }
                                                final method =
                                                    (meta?['method'] ?? '')
                                                        .toString();
                                                final isWs =
                                                    (kind == 'ws') ||
                                                    (method.isEmpty &&
                                                        kind == null);
                                                selIsWs = isWs;
                                                selIsHttp = !isWs;
                                              }

                                              // tabs/views are not needed because TabBar is built inline

                                              return DefaultTabController(
                                                length:
                                                    (() {
                                                      if (selIsWs && selIsHttp)
                                                        return 2;
                                                      if (selIsWs || selIsHttp)
                                                        return 1;
                                                      return 2;
                                                    })(),
                                                child: Column(
                                                  children: [
                                                    Builder(
                                                      builder: (context) {
                                                        if (selIsWs &&
                                                            selIsHttp) {
                                                          return const TabBar(
                                                            tabs: [
                                                              Tab(
                                                                text:
                                                                    'WebSocket',
                                                              ),
                                                              Tab(text: 'HTTP'),
                                                            ],
                                                          );
                                                        }
                                                        // Hide TabBar if only one
                                                        return const SizedBox.shrink();
                                                      },
                                                    ),
                                                    Expanded(
                                                      child: Observer(
                                                        builder: (_) {
                                                          final details =
                                                              context
                                                                  .watch<
                                                                    SessionDetailsStore
                                                                  >();
                                                          final frames =
                                                              details.frames
                                                                  .map(
                                                                    (f) => {
                                                                      'id':
                                                                          f.id,
                                                                      'ts':
                                                                          f.ts.toIso8601String(),
                                                                      'direction':
                                                                          f.direction,
                                                                      'opcode':
                                                                          f.opcode,
                                                                      'size':
                                                                          f.size,
                                                                      'preview':
                                                                          f.preview,
                                                                    },
                                                                  )
                                                                  .toList();
                                                          final events =
                                                              details.events
                                                                  .map(
                                                                    (e) => {
                                                                      'id':
                                                                          e.id,
                                                                      'ts':
                                                                          e.ts.toIso8601String(),
                                                                      'namespace':
                                                                          e.namespace,
                                                                      'event':
                                                                          e.event,
                                                                      'ackId':
                                                                          e.ackId,
                                                                      'argsPreview':
                                                                          e.argsPreview,
                                                                    },
                                                                  )
                                                                  .toList();

                                                          if (selIsWs &&
                                                              selIsHttp) {
                                                            return TabBarView(
                                                              children: [
                                                                WsDetailsPanel(
                                                                  frames:
                                                                      frames,
                                                                  events:
                                                                      events,
                                                                  opcodeFilter:
                                                                      _opcodeFilter,
                                                                  directionFilter:
                                                                      _directionFilter,
                                                                  namespaceCtrl:
                                                                      _namespaceFilterCtrl,
                                                                  onChangeOpcode: (
                                                                    v,
                                                                  ) {
                                                                    setState(() {
                                                                      _opcodeFilter =
                                                                          v;
                                                                    });
                                                                    _savePrefs();
                                                                  },
                                                                  onChangeDirection: (
                                                                    v,
                                                                  ) {
                                                                    setState(() {
                                                                      _directionFilter =
                                                                          v;
                                                                    });
                                                                    _savePrefs();
                                                                  },
                                                                  hideHeartbeats:
                                                                      _hideHeartbeats,
                                                                  onToggleHeartbeats: (
                                                                    v,
                                                                  ) {
                                                                    setState(() {
                                                                      _hideHeartbeats =
                                                                          v;
                                                                    });
                                                                    _savePrefs();
                                                                  },
                                                                ),
                                                                HttpDetailsPanel(
                                                                  sessionId:
                                                                      _selectedSessionId,
                                                                  frames:
                                                                      frames,
                                                                  httpMeta:
                                                                      _httpMeta[_selectedSessionId]
                                                                          as Map<
                                                                            String,
                                                                            dynamic
                                                                          >?,
                                                                ),
                                                              ],
                                                            );
                                                          }
                                                          if (selIsWs) {
                                                            return WsDetailsPanel(
                                                              frames: frames,
                                                              events: events,
                                                              opcodeFilter:
                                                                  _opcodeFilter,
                                                              directionFilter:
                                                                  _directionFilter,
                                                              namespaceCtrl:
                                                                  _namespaceFilterCtrl,
                                                              onChangeOpcode: (
                                                                v,
                                                              ) {
                                                                setState(() {
                                                                  _opcodeFilter =
                                                                      v;
                                                                });
                                                                _savePrefs();
                                                              },
                                                              onChangeDirection: (
                                                                v,
                                                              ) {
                                                                setState(() {
                                                                  _directionFilter =
                                                                      v;
                                                                });
                                                                _savePrefs();
                                                              },
                                                              hideHeartbeats:
                                                                  _hideHeartbeats,
                                                              onToggleHeartbeats: (
                                                                v,
                                                              ) {
                                                                setState(() {
                                                                  _hideHeartbeats =
                                                                      v;
                                                                });
                                                                _savePrefs();
                                                              },
                                                            );
                                                          }
                                                          // selIsHttp
                                                          return HttpDetailsPanel(
                                                            sessionId:
                                                                _selectedSessionId,
                                                            frames: frames,
                                                            httpMeta:
                                                                _httpMeta[_selectedSessionId]
                                                                    as Map<
                                                                      String,
                                                                      dynamic
                                                                    >?,
                                                          );
                                                        },
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              );
                                            },
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                /* SizedBox(
                  width: 320,
                  child: _Card(
                    title: 'Monitor',
                    child: ListView.builder(
                      reverse: true,
                      itemCount: _monitorLog.length,
                      itemBuilder: (_, i) => Text(_monitorLog[i], style: context.appText.monospace),
                    ),
                  ),
                ),*/
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),

              // Global notifications listener (overlay)
              Positioned.fill(
                child: IgnorePointer(
                  child: StreamBuilder(
                    stream: sl<NotificationsService>().stream,
                    builder: (context, snapshot) {
                      if (!snapshot.hasData) return const SizedBox.shrink();
                      final n = snapshot.data as NotificationMessage;
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        NotificationSnackbar.show(context, n);
                      });
                      return const SizedBox.shrink();
                    },
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  List<dynamic> _visibleSessions() {
    final src = context.read<SessionsStore>().items.toList();
    final filtered = src
        .where((s) {
          // hide sessions started before last clear
          if (_since != null) {
            final st = s.startedAt;
            if (st != null && st.isBefore(_since!)) return false;
          }
          // time range filter (if selected)
          if (_selectedRange != null) {
            final start = s.startedAt;
            final end = s.closedAt ?? s.startedAt;
            if (start == null) return false;
            final inRange =
                !(end != null && end.isBefore(_selectedRange!.start)) &&
                !start.isAfter(_selectedRange!.end);
            if (!inRange) return false;
          }
          // domain selection filter (if any chip selected)
          if (_selectedDomains.isNotEmpty) {
            try {
              final host = Uri.parse(s.target as String).host;
              if (!_selectedDomains.contains(host)) return false;
            } catch (_) {}
          }
          final id = s.id;
          final m = (s.httpMeta ?? _httpMeta[id]) ?? const {};
          // method filter
          if (_httpMethodFilter != 'any') {
            if ((m['method'] ?? '') != _httpMethodFilter) return false;
          }
          // status class filter
          if (_httpStatusFilter != 'any') {
            final st = int.tryParse((m['status'] ?? '0').toString()) ?? 0;
            if (_httpStatusFilter == '2xx' && (st < 200 || st > 299))
              return false;
            if (_httpStatusFilter == '3xx' && (st < 300 || st > 399))
              return false;
            if (_httpStatusFilter == '4xx' && (st < 400 || st > 499))
              return false;
            if (_httpStatusFilter == '5xx' && (st < 500 || st > 599))
              return false;
          }
          if (_httpMinDurationMs > 0) {
            final d = int.tryParse((m['durationMs'] ?? '0').toString()) ?? 0;
            if (d < _httpMinDurationMs) return false;
          }
          if (_httpMimeFilter.isNotEmpty) {
            final mime = (m['mime'] ?? '').toString().toLowerCase();
            if (!mime.contains(_httpMimeFilter.toLowerCase())) return false;
          }
          // header key/value filter
          if (_headerKeyCtrl.text.isNotEmpty) {
            final headers =
                (m['headers'] as Map?)?.map(
                  (k, v) => MapEntry(k.toString().toLowerCase(), v.toString()),
                ) ??
                {};
            final hv = headers[_headerKeyCtrl.text.toLowerCase()] ?? '';
            if (_headerValCtrl.text.isNotEmpty &&
                !hv.contains(_headerValCtrl.text))
              return false;
            if (_headerValCtrl.text.isEmpty && hv.isEmpty) return false;
          }
          return true;
        })
        .toList(growable: false);

    if (_groupBy == 'none') return filtered;

    // simple grouping order by key
    String keyFor(dynamic s) {
      try {
        final uri = Uri.parse(s.target as String);
        if (_groupBy == 'domain') return uri.host;
        if (_groupBy == 'route')
          return '${uri.host}${uri.path.split('/').take(3).join('/')}';
      } catch (_) {}
      return '';
    }

    filtered.sort((a, b) => keyFor(a).compareTo(keyFor(b)));
    return filtered;
  }

  void _suckMetaFromSessions() {
    final src = context.read<SessionsStore>().items.toList();
    for (final s in src) {
      final meta = s.httpMeta;
      if (meta != null && meta.isNotEmpty) {
        _httpMeta[s.id] = Map<String, dynamic>.from(meta);
      }
    }
  }

  Widget _buildHttpMetaChips(String id) {
    final m = _httpMeta[id] ?? const {};
    if (m.isEmpty) return const SizedBox.shrink();
    final List<Widget> chips = [];
    if (m['method'] != null) chips.add(_chip(m['method'].toString()));
    if (m['status'] != null) chips.add(_chip(m['status'].toString()));
    if (m['durationMs'] != null) chips.add(_chip('${m['durationMs']} ms'));
    if ((m['streaming'] ?? false) == true) chips.add(_chip('streaming'));
    return Wrap(spacing: 6, children: chips);
  }

  Widget _chip(String text, {Color? backgroundColor, Color? foregroundColor}) {
    final bg = backgroundColor ?? Theme.of(context).colorScheme.surfaceVariant;
    final fg =
        foregroundColor ?? Theme.of(context).colorScheme.onSurfaceVariant;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        text,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(color: fg),
      ),
    );
  }

  Widget _chipStrike(
    String text, {
    Color? backgroundColor,
    Color? foregroundColor,
  }) {
    final base = _chip(
      text,
      backgroundColor: backgroundColor,
      foregroundColor: foregroundColor,
    );
    return Stack(
      alignment: Alignment.centerLeft,
      children: [
        base,
        Positioned.fill(
          child: IgnorePointer(
            child: CustomPaint(painter: ChipStrikePainter()),
          ),
        ),
      ],
    );
  }

  Color _statusBg(int st) {
    final cs = Theme.of(context).colorScheme;
    if (st >= 500) return cs.error.withOpacity(0.12);
    if (st >= 400) return cs.tertiary.withOpacity(0.12);
    if (st >= 300) return cs.primary.withOpacity(0.12);
    return Colors.green.withOpacity(0.12);
  }

  Color _statusFg(int st) {
    final cs = Theme.of(context).colorScheme;
    if (st >= 500) return cs.error;
    if (st >= 400) return cs.tertiary;
    if (st >= 300) return cs.primary;
    return Colors.green;
  }

  Color _durationBg(int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green.withOpacity(0.12);
    if (ms < 1000) return cs.tertiary.withOpacity(0.12);
    return cs.error.withOpacity(0.12);
  }

  Color _durationFg(int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green;
    if (ms < 1000) return cs.tertiary;
    return cs.error;
  }

  String _groupKey(dynamic s) {
    if (_groupBy == 'none') return '';
    try {
      final uri = Uri.parse(s.target as String);
      if (_groupBy == 'domain') return uri.host;
      if (_groupBy == 'route')
        return '${uri.host}${uri.path.split('/').take(3).join('/')}';
    } catch (_) {}
    return '';
  }

  String _formatTimeHMSSafe(DateTime? dt) {
    if (dt == null) return '';
    final h = dt.hour.toString().padLeft(2, '0');
    final m = dt.minute.toString().padLeft(2, '0');
    final s = dt.second.toString().padLeft(2, '0');
    return '$h:$m:$s';
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.title, required this.child});
  final String title;
  final Widget child;
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: Theme.of(
                context,
              ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

// moved TimelineSettingsButton to features/inspector/presentation/widgets/timeline_settings_button.dart
