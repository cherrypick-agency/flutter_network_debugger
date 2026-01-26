import 'package:flutter/material.dart';

import 'theme/app_theme.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_web_plugins/url_strategy.dart';
import 'dart:async';
import 'dart:convert';
import 'services/prefs.dart';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:mobx/mobx.dart' as mobx;
import 'features/inspector/application/stores/sessions_store.dart';
import 'features/inspector/application/stores/session_details_store.dart';
import 'features/inspector/application/stores/aggregate_store.dart';
import 'features/inspector/application/stores/home_ui_store.dart';
import 'features/inspector/presentation/widgets/details/details_tabs.dart';
import 'features/inspector/presentation/widgets/timeline/timeline_block.dart';
import 'features/inspector/presentation/widgets/quick_filters_bar.dart';
import 'features/inspector/presentation/widgets/home/header_actions.dart';
import 'features/filters/presentation/widgets/sessions_filters.dart';
import 'features/filters/application/stores/sessions_filters_store.dart';
import 'features/tags/application/stores/tags_store.dart';
import 'features/performance/application/stores/performance_store.dart';
import 'core/di/di.dart' show getIt, setupDI, sl;
import 'core/network/connectivity_banner.dart';
import 'core/notifications/notifications_service.dart';
import 'features/inspector/application/services/monitor_service.dart';
// import removed: http_meta_service not used here
import 'features/inspector/application/services/realtime_inspector_service.dart';

import 'package:app_http_client/application/app_http_client.dart'
    as http_client;
import 'features/hotkeys/presentation/hotkeys_settings_page.dart';
import 'features/landing/presentation/pages/download_page.dart';
import 'features/settings/presentation/settings_page.dart';
import 'features/updates/presentation/pages/updates_page.dart';
import 'features/landing/presentation/pages/integrations_page.dart';
import 'features/compose/presentation/pages/compose_page.dart';
import 'features/scripts/presentation/pages/scripts_page_full.dart';
import 'features/scripts/application/stores/scripts_store.dart';
import 'core/hotkeys/hotkeys_service.dart';
import 'core/hotkeys/keyboard_state_resetter.dart';
import 'features/compiler_management/presentation/pages/compilers_page.dart';
import 'features/compiler_management/presentation/stores/compiler_list_store.dart';
import 'features/compiler_management/presentation/stores/installation_progress_store.dart';
// import removed: debouncer
import 'features/common/notifications/notifications_overlay.dart';
import 'features/breakpoints/presentation/widgets/breakpoints_dialog.dart';
import 'features/mapping/presentation/widgets/mapping_dialog.dart';
import 'features/export_import/presentation/widgets/export_import_dialog.dart';
import 'features/performance/presentation/pages/performance_page.dart';

import 'features/inspector/presentation/pages/home/widgets/sessions_pane.dart';
import 'theme/font_scale.dart';
import 'theme/visual_density_notifier.dart';
import 'features/custom_fonts/application/font_service.dart';
// import 'package:http_debugger/http_debugger.dart';
import 'core/desktop/desktop_bootstrap.dart';
import 'dart:io' show exit;
import 'dart:async' show StreamController;
import 'package:dio/dio.dart' show CancelToken;
import 'package:package_info_plus/package_info_plus.dart';
import 'features/updates/application/services/updates_service.dart';
import 'features/updates/domain/entities/update_info.dart';
import 'features/updates/presentation/widgets/update_dialog.dart';
import 'features/updates/presentation/widgets/download_progress_dialog.dart';
import 'features/updates/infrastructure/platform/platform_installer.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // if (kDebugMode) {
  //   HttpDebugger.enable();
  // }
  if (kIsWeb) {
    setUrlStrategy(const HashUrlStrategy());
  }

  // Для desktop - покажем bootstrap app который запросит конфигурацию
  // Для web - сразу setupDI и запускаем
  if (!kIsWeb && DesktopBootstrap.isDesktop()) {
    runApp(const BootstrapApp());
  } else {
    final packageInfo = await PackageInfo.fromPlatform();
    await setupDI(
      baseUrl: 'http://localhost:9092',
      githubOwner: 'cherrypick-agency',
      githubRepo: 'flutter_network_debugger',
      currentVersion: packageInfo.version,
    );

    // Initialize custom fonts service
    await FontService().initialize();

    runApp(const MyApp());
  }
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});
  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  ThemeMode _mode = ThemeMode.system;
  bool _themeToggled = false;
  double _fontScale = 1.0;
  String? _fontFamily;
  VisualDensity _visualDensity = VisualDensity.standard;
  late final VoidCallback _fontListener;

  @override
  void initState() {
    super.initState();
    _loadTheme();
    _loadFontScale();
    _loadVisualDensity();
    FontScale.value.addListener(() {
      if (!mounted) return;
      setState(() {
        _fontScale = FontScale.value.value;
      });
    });
    VisualDensityNotifier.value.addListener(() {
      if (!mounted) return;
      setState(() {
        _visualDensity = VisualDensityNotifier.value.value;
      });
    });
    // Listen to custom font changes
    _fontListener = () {
      if (!mounted) return;
      setState(() {
        _fontFamily = FontService().getCurrentFontFamily();
      });
    };
    FontService().currentFont.addListener(_fontListener);
    // Set initial font family
    _fontFamily = FontService().getCurrentFontFamily();
  }

  Future<void> _loadTheme() async {
    final m = await PrefsService().loadThemeModeString();
    if (!mounted) return;
    if (_themeToggled)
      return; // don't overwrite user's choice if they already clicked
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
    _themeToggled = true;
    setState(() {
      if (_mode == ThemeMode.system) {
        // if system is light — switch to dark immediately (and vice versa) for visible effect
        final system =
            WidgetsBinding.instance.platformDispatcher.platformBrightness;
        _mode = (system == Brightness.light) ? ThemeMode.dark : ThemeMode.light;
      } else if (_mode == ThemeMode.light) {
        _mode = ThemeMode.dark;
      } else if (_mode == ThemeMode.dark) {
        _mode = ThemeMode.system;
      } else {
        _mode = ThemeMode.light;
      }
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
        Provider<SessionsFiltersStore>.value(value: sl<SessionsFiltersStore>()),
        Provider<HomeUiStore>.value(value: sl<HomeUiStore>()),
        Provider<TagsStore>.value(value: sl<TagsStore>()),
        Provider<PerformanceStore>.value(value: sl<PerformanceStore>()),
      ],
      child: Builder(
        builder: (context) {
          return MaterialApp(
            theme: buildLightTheme(
              fontFamily: _fontFamily,
              visualDensity: _visualDensity,
            ),
            darkTheme: buildDarkTheme(
              fontFamily: _fontFamily,
              visualDensity: _visualDensity,
            ),
            themeMode: _mode,
            builder: (context, child) {
              final mq = MediaQuery.of(context);
              // Применяем глобальный масштаб текста
              return MediaQuery(
                data: mq.copyWith(textScaler: TextScaler.linear(_fontScale)),
                child: KeyboardStateResetter(child: child!),
              );
            },
            routes: {
              '/hotkeys': (_) => const HotkeysSettingsPage(),
              '/settings': (_) => const SettingsPage(),
              '/updates': (_) => const UpdatesPage(),
              '/download': (_) => const DownloadPage(),
              '/integrations': (_) => const IntegrationsPage(),
              '/compose': (_) => const ComposePage(),
              '/scripts': (_) => ScriptsPageFull(store: sl<ScriptsStore>()),
              '/compilers': (_) => CompilersPage(
                store: sl<CompilerListStore>(),
                progressStore: sl<InstallationProgressStore>(),
              ),
              '/performance': (_) => const PerformancePage(),
            },
            home: MyHomePage(onToggleTheme: _toggleTheme),
          );
        },
      ),
    );
  }

  Future<void> _loadFontScale() async {
    try {
      final s = await PrefsService().load();
      final v = double.tryParse((s['fontScale'] ?? '1.0').toString()) ?? 1.0;
      setState(() {
        _fontScale = v;
      });
      FontScale.value.value = v;
    } catch (_) {}
  }

  Future<void> _loadVisualDensity() async {
    try {
      final density = await PrefsService().loadVisualDensity();
      if (!mounted) return;
      final vd = VisualDensityNotifier.fromString(density);
      setState(() {
        _visualDensity = vd;
      });
      VisualDensityNotifier.value.value = vd;
    } catch (_) {}
  }

  @override
  void dispose() {
    FontService().currentFont.removeListener(_fontListener);
    super.dispose();
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
  final TextEditingController _namespaceFilterCtrl = TextEditingController();

  // sessions from store
  final ScrollController _framesCtrl = ScrollController();
  final ScrollController _eventsCtrl = ScrollController();
  // Sessions scroll: if user is at the bottom — stick to bottom when new items arrive
  final ScrollController _sessionsCtrl = ScrollController();
  Timer? _pollTimer;
  // removed: _detailsRefreshDebounce (SSE обновляет детали)
  // Background polling of sessions list as backup update channel

  final FocusNode _searchFocus = FocusNode();
  MonitorListener? _monitorListener;
  final List<mobx.ReactionDisposer> _reactions = [];

  @override
  void initState() {
    super.initState();
    _connectMonitor();

    // Setup synchronous listeners first
    _framesCtrl.addListener(_onFramesScroll);
    _eventsCtrl.addListener(_onEventsScroll);
    _sessionsCtrl.addListener(_onSessionsScroll);

    // Move ALL async initialization to post-frame callback
    // This ensures widget tree is fully built before accessing context
    // Prevents deadlock between setState() and context.read()
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _restorePrefs().then((_) async {
        // Синхронизируем реальное состояние записи с бэкендом (перезапишет префы при расхождении)
        await _syncRecordingStateFromBackend();
        // Setup search controller listener AFTER preferences are restored
        // This prevents cascade of resubscribe calls during initialization
        _sessionSearchCtrl.addListener(() {
          sl<HomeUiStore>().setSessionSearchQuery(_sessionSearchCtrl.text);
          // ignore: discarded_futures
          sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
        });

        // После восстановления фильтров сразу подписываемся на realtime
        // ignore: discarded_futures
        sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
        // Ресабскрайб при изменении ключевых фильтров
        final ui = sl<HomeUiStore>();
        final f = context.read<SessionsFiltersStore>();
        _reactions.add(
          mobx.reaction((_) => ui.captureScope.value, (_) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
        _reactions.add(
          mobx.reaction((_) => ui.quickTypes.toList(growable: false), (_) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
        _reactions.add(
          mobx.reaction((_) => ui.quickStatusGroups.toList(growable: false), (
            _,
          ) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
        _reactions.add(
          mobx.reaction((_) => f.target, (_) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
        _reactions.add(
          mobx.reaction((_) => ui.includePaused.value, (_) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
        _reactions.add(
          mobx.reaction((_) => f.selectedTags.toList(growable: false), (_) {
            // ignore: discarded_futures
            sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          }),
        );
      });
    });
  }

  Future<void> _syncRecordingStateFromBackend() async {
    try {
      final ui = sl<HomeUiStore>();
      final client = sl<http_client.AppHttpClient>();
      final res = await client.get<Map<String, dynamic>>(
        path: '/_api/v1/capture',
      );
      final data = (res.data is Map<String, dynamic>)
          ? res.data as Map<String, dynamic>
          : null;
      final rec = data?['recording'];
      if (rec is bool) {
        ui.setIsRecording(rec);
        try {
          await PrefsService().saveIsRecording(rec);
        } catch (_) {}
      }
    } catch (_) {}
  }

  @override
  void dispose() {
    final monitor = sl<MonitorService>();
    if (_monitorListener != null) {
      monitor.removeListener(_monitorListener!);
    }
    monitor.dispose();
    _pollTimer?.cancel();
    _framesCtrl.dispose();
    _eventsCtrl.dispose();
    _sessionsCtrl.dispose();
    _searchFocus.dispose();
    for (final d in _reactions) {
      try {
        d();
      } catch (_) {}
    }
    super.dispose();
  }

  void _connectMonitor() {
    final monitor = sl<MonitorService>();
    final listener = (Map<String, dynamic> ev) {
      try {
        final ui = sl<HomeUiStore>();
        final t = (ev['type'] ?? '').toString();
        if (t == 'sessions_cleared') {
          // мгновенно очищаем стора и мету, чтобы список не восстанавливался от фонового поллинга
          try {
            context.read<SessionsStore>().clear();
          } catch (_) {}
          try {
            context.read<AggregateStore>().clear();
          } catch (_) {}
          ui.setSelectedSessionId(null);
          ui.setSelectedRange(null);
          ui.setWfFitAll(true);
          // выровняем since на сейчас, чтобы повторная подгрузка не схватила старые
          ui.setSince(DateTime.now().toUtc());
          // realtime: пересоберём подписку
          // ignore: discarded_futures
          sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
          return;
        }
        // список обновляется по сокету — перезагрузка не нужна
        if (!ui.isRecording.value && t == 'session_started') {
          return; // paused: don't pick up updates
        }
        // frames/events приходят по SSE, доп. поллинг не нужен
        if (t == 'session_error') {
          // Display user-friendly error notification
          final errorData = ev['error'] as Map<String, dynamic>?;
          if (errorData != null) {
            final code = errorData['code']?.toString() ?? 'UNKNOWN_ERROR';
            final message =
                errorData['message']?.toString() ?? 'Unknown error occurred';
            final target = errorData['target']?.toString() ?? '';
            final method = errorData['method']?.toString() ?? '';
            final sessionId = (ev['id'] ?? '').toString();

            // Build user-friendly title based on error code
            String title = 'Proxy Error';
            if (code == 'CONNECTION_CLOSED') {
              title = 'Connection Closed';
            } else if (code == 'SERVER_UNAVAILABLE') {
              title = 'Server Unavailable';
            } else if (code == 'TIMEOUT') {
              title = 'Request Timeout';
            } else if (code == 'DNS_ERROR') {
              title = 'Domain Not Found';
            } else if (code == 'TLS_ERROR') {
              title = 'Certificate Error';
            }

            // Truncate long URLs for readability
            final displayTarget = target.length > 50
                ? '${target.substring(0, 47)}...'
                : target;

            // Handle empty or short session IDs
            final sessionDisplay = sessionId.isEmpty
                ? 'Unknown'
                : (sessionId.length >= 8
                      ? sessionId.substring(0, 8)
                      : sessionId);

            sl<NotificationsService>().error(
              title,
              '$method $displayTarget\n$message\nSession: $sessionDisplay...',
            );
            // realtime обновит список автоматически
          }
        }
      } catch (_) {}
    };
    monitor.addListener(listener);
    _monitorListener = listener;
    // fire and forget
    // ignore: discarded_futures
    monitor.connect();
  }

  // removed: _scheduleSessionsReload (REST поллинг заменён на сокеты)

  Future<void> _clearAllSessions() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
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
      final client = sl<http_client.AppHttpClient>();
      bool cleared = false;
      try {
        await client.delete(path: '/_api/v1/sessions');
        cleared = true;
      } catch (_) {}
      if (!cleared) {
        try {
          await client.delete(path: '/api/sessions');
          cleared = true;
        } catch (_) {}
      }
      if (!cleared) {
        // fallback: iteratively delete known sessions
        final items = context.read<SessionsStore>().items.toList();
        for (final s in items) {
          try {
            await client.delete(path: '/_api/v1/sessions/${s.id}');
          } catch (_) {}
          try {
            await client.delete(path: '/api/sessions/${s.id}');
          } catch (_) {}
        }
      }
    } catch (_) {}

    // reset local state immediately
    final ui = sl<HomeUiStore>();
    ui.setSelectedSessionId(null);
    ui.setSelectedRange(null);
    ui.setWfFitAll(true);
    ui.setSince(DateTime.now().toUtc());
    try {
      if (ui.since.value != null) {
        await PrefsService().saveSince(ui.since.value!);
      }
    } catch (_) {}
    // clear stores immediately for instant UI reset
    try {
      context.read<SessionsStore>().clear();
    } catch (_) {}
    try {
      context.read<AggregateStore>().clear();
    } catch (_) {}
    // immediately reload domain aggregator so counters reset instantly
    try {
      await context.read<AggregateStore>().load(groupBy: 'domain');
    } catch (_) {}
    // realtime заново подтянет состояние
    await _loadSessions();
  }

  Future<void> _loadSessions() async {
    // realtime подписка заменяет REST-загрузку списка/агрегатов
    // ignore: discarded_futures
    sl<RealtimeInspectorService>().resubscribeWithCurrentFilters();
  }

  Future<void> _restorePrefs() async {
    final data = await PrefsService().load();
    final ui = sl<HomeUiStore>();
    setState(() {
      _sessionSearchCtrl.text = data['q']!;
      ui.setSessionSearchQuery(_sessionSearchCtrl.text);
      ui.setOpcodeFilter(data['opcode'] ?? 'all');
      ui.setDirectionFilter(data['direction'] ?? 'all');
      _namespaceFilterCtrl.text = data['namespace']!;
    });
    // restore filters in Store
    final f = context.read<SessionsFiltersStore>();
    f.setTarget(data['targetFilter'] ?? '');
    f.setHttpMethod(data['httpMethod'] ?? 'any');
    f.setHttpStatus(data['httpStatus'] ?? 'any');
    f.setHttpMime(data['httpMime'] ?? '');
    f.setHttpMinDurationMs(int.tryParse(data['httpMinDuration'] ?? '0') ?? 0);
    f.setGroupBy(data['groupBy'] ?? 'none');
    f.setHeaderKey(data['headerKey'] ?? '');
    f.setHeaderVal(data['headerVal'] ?? '');
    // restore selected tags
    try {
      final tagsJson = data['selectedTags'];
      if (tagsJson != null && tagsJson.isNotEmpty) {
        f.setSelectedTags(List<String>.from(jsonDecode(tagsJson) as List));
      }
    } catch (_) {}
    // restore since-ts if any
    try {
      ui.setSince(await PrefsService().loadSince());
    } catch (_) {}
    // restore recording state
    try {
      ui.setIsRecording(await PrefsService().loadIsRecording());
    } catch (_) {}
  }

  Future<void> _savePrefs() async {
    final f = context.read<SessionsFiltersStore>();
    final ui = sl<HomeUiStore>();
    await PrefsService().save(
      baseUrl: sl<http_client.AppHttpClient>().defaultHost,
      targetWs: '',
      q: ui.sessionSearchQuery.value,
      targetFilter: f.target,
      opcode: ui.opcodeFilter.value,
      direction: ui.directionFilter.value,
      namespace: _namespaceFilterCtrl.text,
      httpMethod: f.httpMethod,
      httpStatus: f.httpStatus,
      httpMime: f.httpMime,
      httpMinDurationMs: f.httpMinDurationMs,
      groupBy: f.groupBy,
      headerKey: f.headerKey,
      headerVal: f.headerVal,
      selectedTags: jsonEncode(f.selectedTags.toList()),
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
    // отключено: детали обновляет SSE-stream
    _pollTimer?.cancel();
  }

  // removed: _tickRefresh (SSE обновляет детали)

  void _onSessionsScroll() {
    if (!_sessionsCtrl.hasClients) return;
    final pos = _sessionsCtrl.position;
    // Consider "at bottom" if there's a small tail left (for resize stabilization)
    (pos.maxScrollExtent - pos.pixels) < 48;
  }

  // filters moved to WsDetailsPanel

  Future<void> _deleteSelected() async {
    final ui = sl<HomeUiStore>();
    if (ui.selectedSessionId.value == null) return;
    final id = ui.selectedSessionId.value!;
    final client = sl<http_client.AppHttpClient>();
    try {
      await client.delete(path: '/_api/v1/sessions/$id');
    } catch (_) {}
    ui.setSelectedSessionId(null);
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
        // title: const Text('network-debugger Console'), 
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
                            child: Observer(
                              builder: (_) {
                                final ui = sl<HomeUiStore>();
                                return HeaderActions(
                                  showFilters: ui.showFilters.value,
                                  onToggleFilters: () {
                                    ui.toggleShowFilters();
                                  },
                                  onToggleTheme: widget.onToggleTheme,
                                  onOpenHotkeys: () {
                                    Navigator.of(context).pushNamed('/hotkeys');
                                  },
                                  onOpenSettings: () {
                                    Navigator.of(
                                      context,
                                    ).pushNamed('/settings');
                                  },
                                  onOpenUpdates: () {
                                    Navigator.of(context).pushNamed('/updates');
                                  },
                                  onOpenIntegrations: () {
                                    Navigator.of(
                                      context,
                                    ).pushNamed('/integrations');
                                  },
                                  onOpenCompose: () {
                                    Navigator.of(context).pushNamed('/compose');
                                  },
                                  onOpenScripts: () {
                                    Navigator.of(context).pushNamed('/scripts');
                                  },
                                  onOpenBreakpoints: () async {
                                    await showGeneralDialog(
                                      context: context,
                                      barrierDismissible: true,
                                      barrierLabel: 'Breakpoints',
                                      pageBuilder: (ctx, _, __) {
                                        return const BreakpointsDialog();
                                      },
                                    );
                                  },
                                  onOpenMapping: () async {
                                    await showGeneralDialog(
                                      context: context,
                                      barrierDismissible: true,
                                      barrierLabel: 'Mapping',
                                      pageBuilder: (ctx, _, __) {
                                        return const MappingDialog();
                                      },
                                    );
                                  },
                                  onOpenExportImport: () async {
                                    final sessionsStore = sl<SessionsStore>();
                                    await showExportImportDialog(
                                      context,
                                      visibleSessionIds: sessionsStore.items
                                          .map((s) => s.id)
                                          .toList(),
                                      visibleSessionsCount:
                                          sessionsStore.items.length,
                                      totalSessionsCount:
                                          sessionsStore.items.length,
                                    );
                                  },
                                  // onOpenPerformance: () {
                                  //   Navigator.of(
                                  //     context,
                                  //   ).pushNamed('/performance');
                                  // },
                                  isRecording: ui.isRecording.value,
                                  onToggleRecording: () async {
                                    // Toggle on backend via capture API
                                    final newVal = !ui.isRecording.value;
                                    // ignore: avoid_print
                                    print(
                                      '[ToggleRecording] click; current=${ui.isRecording.value} -> newVal=$newVal',
                                    );
                                    bool? backendRec;
                                    final client =
                                        sl<http_client.AppHttpClient>();
                                    try {
                                      // ignore: avoid_print
                                      print(
                                        '[ToggleRecording] POST /_api/v1/capture action=${newVal ? 'start' : 'stop'}',
                                      );
                                      await client.post(
                                        path: '/_api/v1/capture',
                                        body: {
                                          'action': newVal ? 'start' : 'stop',
                                        },
                                      );
                                    } catch (e) {
                                      // ignore: avoid_print
                                      print(
                                        '[ToggleRecording] POST failed: $e',
                                      );
                                    }
                                    // Verify actual state from backend
                                    try {
                                      // ignore: avoid_print
                                      print(
                                        '[ToggleRecording] GET /_api/v1/capture ...',
                                      );
                                      final res = await client
                                          .get<Map<String, dynamic>>(
                                            path: '/_api/v1/capture',
                                          );
                                      final data =
                                          (res.data is Map<String, dynamic>)
                                          ? res.data as Map<String, dynamic>
                                          : null;
                                      // ignore: avoid_print
                                      print(
                                        '[ToggleRecording] GET data: $data (type=${res.data.runtimeType})',
                                      );
                                      final rec = data?['recording'];
                                      if (rec is bool) {
                                        backendRec = rec;
                                      }
                                    } catch (e) {
                                      // ignore: avoid_print
                                      print('[ToggleRecording] GET failed: $e');
                                    }
                                    final effectiveRec = backendRec ?? newVal;
                                    // ignore: avoid_print
                                    print(
                                      '[ToggleRecording] effectiveRec=$effectiveRec (backendRec=$backendRec)',
                                    );
                                    ui.setIsRecording(effectiveRec);
                                    // При остановке записи скрываем «непривязанные» (новые) сессии
                                    if (!ui.isRecording.value) {
                                      ui.setIncludePaused(false);
                                      ui.setPausedSince(DateTime.now().toUtc());
                                    } else {
                                      ui.setPausedSince(null);
                                    }
                                    try {
                                      await PrefsService().saveIsRecording(
                                        ui.isRecording.value,
                                      );
                                    } catch (e) {
                                      // ignore: avoid_print
                                      print(
                                        '[ToggleRecording] saveIsRecording failed: $e',
                                      );
                                    }
                                    // ignore: avoid_print
                                    print(
                                      '[ToggleRecording] done; ui.isRecording=${ui.isRecording.value}',
                                    );
                                    await _loadSessions();
                                  },
                                  themeMode:
                                      (context
                                          .findAncestorStateOfType<
                                            _MyAppState
                                          >()
                                          ?._mode) ??
                                      ThemeMode.system,
                                  timelineVisible: ui.showTimeline.value,
                                  onToggleTimeline: () {
                                    ui.setShowTimeline(!ui.showTimeline.value);
                                  },
                                );
                              },
                            ),
                          ),
                          // Waterfall timeline (animated)
                          Observer(
                            builder: (_) {
                              final ui = sl<HomeUiStore>();
                              return AnimatedSize(
                                duration: const Duration(milliseconds: 220),
                                curve: Curves.easeInOutCubic,
                                alignment: Alignment.topCenter,
                                child: ui.showTimeline.value
                                    ? TimelineBlock(
                                        since: ui.since.value,
                                        wfFitAll: ui.wfFitAll.value,
                                        onFitAllChanged: (v) =>
                                            ui.setWfFitAll(v),
                                        onSelectSession: (id) {
                                          ui.setSelectedSessionId(id);
                                          _loadDetails(id);
                                        },
                                        onClearAllSessions: _clearAllSessions,
                                        selectedRange: ui.selectedRange.value,
                                        onRangeChanged: (range) =>
                                            ui.setSelectedRange(range),
                                        onRangeCleared: () =>
                                            ui.clearSelectedRange(),
                                      )
                                    : const SizedBox.shrink(),
                              );
                            },
                          ),
                          const SizedBox(height: 6),
                          // Быстрая панель фильтров по типам/статусам
                          const QuickFiltersBar(),
                          const SizedBox(height: 8),
                          Observer(
                            builder: (_) {
                              final show = sl<HomeUiStore>().showFilters.value;
                              return AnimatedSize(
                                duration: const Duration(milliseconds: 220),
                                curve: Curves.easeInOutCubic,
                                alignment: Alignment.topCenter,
                                child: show
                                    ? Theme(
                                        data: Theme.of(context).copyWith(
                                          inputDecorationTheme:
                                              const InputDecorationTheme(
                                                isDense: true,
                                                contentPadding:
                                                    EdgeInsets.symmetric(
                                                      horizontal: 8,
                                                      vertical: 4,
                                                    ),
                                                labelStyle: TextStyle(
                                                  fontSize: 12,
                                                ),
                                              ),
                                        ),
                                        child: SessionsFilters(
                                          onApply: () async {
                                            await _savePrefs();
                                            await _loadSessions();
                                          },
                                        ),
                                      )
                                    : const SizedBox.shrink(),
                              );
                            },
                          ),
                          const SizedBox(height: 12),
                          Expanded(
                            child: Row(
                              children: [
                                SizedBox(
                                  width: 360,
                                  child: SessionsPane(
                                    searchCtrl: _sessionSearchCtrl,
                                    sessionsCtrl: _sessionsCtrl,
                                    onSelectSession: (id) {
                                      sl<HomeUiStore>().setSelectedSessionId(
                                        id,
                                      );
                                      _loadDetails(id);
                                    },
                                  ),
                                ),
                                const VerticalDivider(width: 1),
                                // if there's selectedSessionId, show details panel
                                Observer(
                                  builder: (_) {
                                    final has =
                                        sl<HomeUiStore>()
                                            .selectedSessionId
                                            .value !=
                                        null;
                                    if (!has) {
                                      return const Expanded(
                                        child: _SessionPlaceholder(
                                          key: ValueKey('session_placeholder'),
                                        ),
                                      );
                                    }
                                    return Expanded(
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
                                                bool wsClosed = false;
                                                DateTime? wsClosedAt;
                                                if (sl<HomeUiStore>()
                                                        .selectedSessionId
                                                        .value !=
                                                    null) {
                                                  final items = context
                                                      .watch<SessionsStore>()
                                                      .items
                                                      .toList();
                                                  Map<String, dynamic>? meta;
                                                  String? kind;
                                                  for (final s in items) {
                                                    if (s.id ==
                                                        sl<HomeUiStore>()
                                                            .selectedSessionId
                                                            .value) {
                                                      meta =
                                                          (s.httpMeta ??
                                                                  sl<
                                                                        HomeUiStore
                                                                      >()
                                                                      .httpMeta[s
                                                                      .id])
                                                              ?.cast<
                                                                String,
                                                                dynamic
                                                              >();
                                                      kind = s.kind;
                                                      wsClosed =
                                                          s.closedAt != null;
                                                      wsClosedAt = s.closedAt;
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
                                                  length: (() {
                                                    if (selIsWs && selIsHttp)
                                                      return 2;
                                                    if (selIsWs || selIsHttp)
                                                      return 1;
                                                    return 2;
                                                  })(),
                                                  child: Observer(
                                                    builder: (_) {
                                                      final details = context
                                                          .watch<
                                                            SessionDetailsStore
                                                          >();
                                                      final frames = details
                                                          .frames
                                                          .map(
                                                            (f) => {
                                                              'id': f.id,
                                                              'ts': f.ts
                                                                  .toIso8601String(),
                                                              'direction':
                                                                  f.direction,
                                                              'opcode':
                                                                  f.opcode,
                                                              'size': f.size,
                                                              'preview':
                                                                  f.preview,
                                                            },
                                                          )
                                                          .toList();
                                                      final events = details
                                                          .events
                                                          .map(
                                                            (e) => {
                                                              'id': e.id,
                                                              'ts': e.ts
                                                                  .toIso8601String(),
                                                              'namespace':
                                                                  e.namespace,
                                                              'event': e.event,
                                                              'ackId': e.ackId,
                                                              'argsPreview':
                                                                  e.argsPreview,
                                                            },
                                                          )
                                                          .toList();
                                                      return DetailsTabs(
                                                        showWs: selIsWs,
                                                        showHttp: selIsHttp,
                                                        frames: frames
                                                            .cast<
                                                              Map<
                                                                String,
                                                                dynamic
                                                              >
                                                            >(),
                                                        events: events
                                                            .cast<
                                                              Map<
                                                                String,
                                                                dynamic
                                                              >
                                                            >(),
                                                        selectedSessionId:
                                                            sl<HomeUiStore>()
                                                                .selectedSessionId
                                                                .value,
                                                        httpMeta:
                                                            sl<HomeUiStore>()
                                                                .httpMeta[sl<
                                                                  HomeUiStore
                                                                >()
                                                                .selectedSessionId
                                                                .value],
                                                        opcodeFilter:
                                                            sl<HomeUiStore>()
                                                                .opcodeFilter
                                                                .value,
                                                        directionFilter:
                                                            sl<HomeUiStore>()
                                                                .directionFilter
                                                                .value,
                                                        namespaceCtrl:
                                                            _namespaceFilterCtrl,
                                                        onChangeOpcode: (v) {
                                                          sl<HomeUiStore>()
                                                              .setOpcodeFilter(
                                                                v,
                                                              );
                                                          _savePrefs();
                                                        },
                                                        onChangeDirection: (v) {
                                                          sl<HomeUiStore>()
                                                              .setDirectionFilter(
                                                                v,
                                                              );
                                                          _savePrefs();
                                                        },
                                                        hideHeartbeats:
                                                            sl<HomeUiStore>()
                                                                .hideHeartbeats
                                                                .value,
                                                        onToggleHeartbeats: (v) {
                                                          sl<HomeUiStore>()
                                                              .setHideHeartbeats(
                                                                v,
                                                              );
                                                          _savePrefs();
                                                        },
                                                        wsClosed: wsClosed,
                                                        wsClosedAt: wsClosedAt,
                                                      );
                                                    },
                                                  ),
                                                );
                                              },
                                            ),
                                          ),
                                        ],
                                      ),
                                    );
                                  },
                                ),
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
              const NotificationsOverlay(),
            ],
          ),
        ),
      ),
    );
  }

  // removed: _suckMetaFromSessions (meta приходит сразу в init/upsert)
}

class _SessionPlaceholder extends StatelessWidget {
  const _SessionPlaceholder({super.key});
  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Icon(Icons.find_in_page, size: 72, color: cs.onSurfaceVariant),
            const SizedBox(height: 12),
            Text(
              'No session selected',
              style: tt.titleLarge,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 6),
            Text(
              'Pick a session on the left to see its details here.',
              style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

/// Helper widget для инициализации внутри MaterialApp context
class _BootstrapInitializer extends StatefulWidget {
  final Future<void> Function(BuildContext) onInitialize;
  final VoidCallback onInitialized;

  const _BootstrapInitializer({
    required this.onInitialize,
    required this.onInitialized,
  });

  @override
  State<_BootstrapInitializer> createState() => _BootstrapInitializerState();
}

class _BootstrapInitializerState extends State<_BootstrapInitializer> {
  @override
  void initState() {
    super.initState();
    // Показываем dialog после первого build когда MaterialApp context доступен
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      if (!mounted) return;
      await widget.onInitialize(context);
      widget.onInitialized();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(48),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface.withValues(alpha: 0.9),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 20,
            spreadRadius: 5,
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            Icons.network_check,
            size: 64,
            color: Theme.of(context).colorScheme.primary,
          ),
          const SizedBox(height: 16),
          Text(
            'Network Debugger',
            style: Theme.of(
              context,
            ).textTheme.headlineMedium?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Text(
            'Initializing...',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      ),
    );
  }
}

/// Bootstrap приложение для desktop платформ
/// Показывает startup dialog, запускает Go сервер, затем показывает MyApp
class BootstrapApp extends StatefulWidget {
  const BootstrapApp({super.key});

  @override
  State<BootstrapApp> createState() => _BootstrapAppState();
}

class _BootstrapAppState extends State<BootstrapApp> {
  bool _initialized = false;

  @override
  void initState() {
    super.initState();
    // Инициализация будет происходить внутри MaterialApp context
    // через BootstrapInitializer widget
  }

  Future<void> _initialize(BuildContext context) async {
    if (!mounted) return;

    final apiPort = await DesktopBootstrap.bootstrap(context);

    if (apiPort == null) {
      // Пользователь отменил запуск или произошла ошибка
      // Закрываем приложение
      if (mounted) {
        // ignore: use_build_context_synchronously
        await showDialog(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Text('Startup Cancelled'),
            content: const Text('Application will now exit.'),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.of(ctx).pop();
                  // Exit app
                  exit(0);
                },
                child: const Text('OK'),
              ),
            ],
          ),
        );
      }
      return;
    }

    // Инициализируем DI с правильным портом
    PackageInfo packageInfo;
    try {
      packageInfo = await PackageInfo.fromPlatform();
    } catch (e) {
      // Fallback если package_info не сработал
      packageInfo = PackageInfo(
        appName: 'Network Debugger',
        packageName: 'com.example.app',
        version: '1.0.0',
        buildNumber: '1',
      );
    }

    await setupDI(
      baseUrl: 'http://localhost:$apiPort',
      githubOwner: 'cherrypick-agency',
      githubRepo: 'flutter_network_debugger',
      currentVersion: packageInfo.version,
    );

    // Initialize custom fonts service
    await FontService().initialize();

    if (mounted) {
      setState(() {
        _initialized = true;
      });

      // Проверяем обновления после успешной инициализации
      _checkForUpdates();
    }
  }

  /// Проверяет наличие обновлений и показывает диалог если доступны
  Future<void> _checkForUpdates() async {
    if (!mounted) return;

    try {
      final updatesService = getIt<UpdatesService>();
      final updateInfo = await updatesService.checkForUpdates();

      if (updateInfo != null && mounted) {
        // ignore: use_build_context_synchronously
        final result = await showUpdateDialog(context, updateInfo);

        switch (result) {
          case UpdateDialogResult.download:
            // Запускаем загрузку с прогрессом
            await _downloadAndInstall(updateInfo);
            break;
          case UpdateDialogResult.skip:
            final updatesService = getIt<UpdatesService>();
            await updatesService.skipVersion(updateInfo.version);
            break;
          case UpdateDialogResult.viewAllReleases:
            // Открываем страницу со всеми релизами
            // ignore: use_build_context_synchronously
            Navigator.of(context).pushNamed('/updates');
            break;
          case UpdateDialogResult.remindLater:
          case null:
            // Ничего не делаем
            break;
        }
      }
    } catch (e) {
      // Тихо игнорируем ошибки проверки обновлений
      // чтобы не мешать работе приложения
    }
  }

  /// Загружает и устанавливает обновление
  Future<void> _downloadAndInstall(UpdateInfo updateInfo) async {
    if (!mounted) return;

    final progressController = StreamController<DownloadProgress>();
    final cancelToken = CancelToken();

    try {
      // Показываем диалог прогресса
      // ignore: use_build_context_synchronously
      final updatesService = getIt<UpdatesService>();
      final downloadFuture = updatesService.downloadUpdate(
        updateInfo,
        progressController: progressController,
        cancelToken: cancelToken,
      );

      // ignore: use_build_context_synchronously
      final downloadDialogFuture = showDownloadProgressDialog(
        context,
        progressStream: progressController.stream,
        onCancel: () => cancelToken.cancel('User cancelled'),
      );

      try {
        // Ждем завершения загрузки
        final filePath = await downloadFuture;

        // Ждем закрытия диалога прогресса (диалог может еще использовать stream)
        final downloadResult = await downloadDialogFuture;

        if (downloadResult == DownloadDialogResult.cancelled) {
          // Пользователь отменил загрузку
          return;
        }

        if (downloadResult == DownloadDialogResult.error || filePath == null) {
          // Ошибка загрузки
          if (mounted) {
            // ignore: use_build_context_synchronously
            await showDialog(
              context: context,
              builder: (ctx) => AlertDialog(
                title: const Row(
                  children: [
                    Icon(Icons.error_outline, color: Colors.red),
                    SizedBox(width: 12),
                    Text('Download Failed'),
                  ],
                ),
                content: const Text(
                  'Failed to download update. Please try again later or download manually from GitHub.',
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.of(ctx).pop(),
                    child: const Text('OK'),
                  ),
                ],
              ),
            );
          }
          return;
        }

        // Загрузка успешна - показываем диалог установки
        if (mounted) {
          // ignore: use_build_context_synchronously
          final shouldInstall = await showDialog<bool>(
            context: context,
            builder: (ctx) => AlertDialog(
              title: const Row(
                children: [
                  Icon(Icons.check_circle, color: Colors.green),
                  SizedBox(width: 12),
                  Text('Download Complete'),
                ],
              ),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Update has been downloaded successfully!'),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.blue.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: Colors.blue.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.info_outline,
                          color: Colors.blue,
                          size: 20,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Version ${updateInfo.version}',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Colors.blue,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),
                  const Text(
                    'Would you like to open the installer now?',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(false),
                  child: const Text('Later'),
                ),
                ElevatedButton.icon(
                  onPressed: () => Navigator.of(ctx).pop(true),
                  icon: const Icon(Icons.install_desktop),
                  label: const Text('Install Now'),
                ),
              ],
            ),
          );

          if (shouldInstall == true) {
            // Открываем установщик
            final installerResult = await openInstaller(filePath);

            if (mounted) {
              if (installerResult.success) {
                // Показываем инструкции если есть
                if (installerResult.instructions != null) {
                  // ignore: use_build_context_synchronously
                  await showDialog(
                    context: context,
                    builder: (ctx) => AlertDialog(
                      title: const Text('Installation Instructions'),
                      content: SingleChildScrollView(
                        child: Text(installerResult.instructions!),
                      ),
                      actions: [
                        ElevatedButton(
                          onPressed: () => Navigator.of(ctx).pop(),
                          child: const Text('OK'),
                        ),
                      ],
                    ),
                  );
                }
              } else {
                // Показываем ошибку
                // ignore: use_build_context_synchronously
                await showDialog(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Row(
                      children: [
                        Icon(Icons.error_outline, color: Colors.red),
                        SizedBox(width: 12),
                        Text('Installation Error'),
                      ],
                    ),
                    content: Text(
                      installerResult.errorMessage ??
                          'Failed to open installer',
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(ctx).pop(),
                        child: const Text('OK'),
                      ),
                    ],
                  ),
                );
              }
            }
          }
        }
      } finally {
        // CRITICAL: Всегда закрываем stream после того как диалог закрылся
        // Это предотвращает memory leak
        if (!progressController.isClosed) {
          await progressController.close();
        }
      }
    } catch (e) {
      // Показываем ошибку пользователю
      if (mounted) {
        // ignore: use_build_context_synchronously
        await showDialog(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Text('Error'),
            content: Text('Failed to download update: $e'),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('OK'),
              ),
            ],
          ),
        );
      }
    } finally {
      // Финальная очистка - закрываем stream если еще не закрыт
      // (на случай если что-то пошло не так в catch блоке)
      if (!progressController.isClosed) {
        await progressController.close();
      }
    }
  }

  @override
  void dispose() {
    // Останавливаем сервер при закрытии приложения
    DesktopBootstrap.shutdown();
    // Освобождаем ресурсы UpdatesService
    try {
      getIt<UpdatesService>().dispose();
    } catch (e) {
      // Ignore if not registered
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Всегда оборачиваем в MaterialApp для Material виджетов
    if (!_initialized) {
      // Показываем splash screen пока инициализируемся
      return MaterialApp(
        theme: buildLightTheme(),
        darkTheme: buildDarkTheme(),
        home: Scaffold(
          body: Container(
            decoration: const BoxDecoration(
              image: DecorationImage(
                image: NetworkImage('https://i.imgur.com/MdXfv74.png'),
                fit: BoxFit.cover,
              ),
            ),
            child: Center(
              child: _BootstrapInitializer(
                onInitialize: _initialize,
                onInitialized: () {
                  if (mounted) {
                    setState(() {
                      _initialized = true;
                    });
                  }
                },
              ),
            ),
          ),
        ),
      );
    }

    // После инициализации показываем основное приложение
    return const MyApp();
  }
}
