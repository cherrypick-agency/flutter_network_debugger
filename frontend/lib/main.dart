import 'package:flutter/material.dart';

import 'theme/app_theme.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_web_plugins/url_strategy.dart';
import 'dart:async';
import 'services/prefs.dart';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'features/inspector/application/stores/sessions_store.dart';
import 'features/inspector/application/stores/session_details_store.dart';
import 'features/inspector/application/stores/aggregate_store.dart';
import 'features/inspector/application/stores/home_ui_store.dart';
import 'features/inspector/presentation/widgets/details/details_tabs.dart';
import 'features/inspector/presentation/widgets/timeline/timeline_block.dart';
import 'features/inspector/presentation/widgets/home/header_actions.dart';
import 'features/filters/presentation/widgets/sessions_filters.dart';
import 'features/filters/application/stores/sessions_filters_store.dart';
import 'core/di/di.dart';
import 'core/network/connectivity_banner.dart';
import 'features/inspector/application/services/monitor_service.dart';
import 'features/inspector/application/services/http_meta_service.dart';
import 'features/inspector/application/services/sessions_polling_service.dart'
    as features_inspector_application_services;

import 'package:app_http_client/application/app_http_client.dart'
    as http_client;
import 'features/hotkeys/presentation/hotkeys_settings_page.dart';
import 'features/landing/presentation/pages/download_page.dart';
import 'features/settings/presentation/settings_page.dart';
import 'core/hotkeys/hotkeys_service.dart';
import 'core/utils/debouncer.dart';
import 'features/common/notifications/notifications_overlay.dart';

import 'features/inspector/presentation/pages/home/widgets/sessions_pane.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  if (kIsWeb) {
    setUrlStrategy(const HashUrlStrategy());
  }
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
  bool _themeToggled = false;

  @override
  void initState() {
    super.initState();
    _loadTheme();
  }

  Future<void> _loadTheme() async {
    final m = await PrefsService().loadThemeModeString();
    if (!mounted) return;
    if (_themeToggled)
      return; // не затираем выбор пользователя, если он уже нажал
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
        // если система светлая — переключаемся сразу в тёмную (и наоборот), чтобы был видимый эффект
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
      ],
      child: Builder(
        builder: (context) {
          return MaterialApp(
            theme: buildLightTheme(),
            darkTheme: buildDarkTheme(),
            themeMode: _mode,
            routes: {
              '/hotkeys': (_) => const HotkeysSettingsPage(),
              '/settings': (_) => const SettingsPage(),
              '/download': (_) => const DownloadPage(),
            },
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
  final TextEditingController _namespaceFilterCtrl = TextEditingController();

  // sessions from store
  bool _loadingSessions = false;
  final ScrollController _framesCtrl = ScrollController();
  final ScrollController _eventsCtrl = ScrollController();
  // Скролл сессий: если пользователь на самом дне — липнем к низу при новых элементах
  final ScrollController _sessionsCtrl = ScrollController();
  Timer? _pollTimer;
  Debouncer _sessionsReloadDebounce = Debouncer(
    const Duration(milliseconds: 300),
  );
  Debouncer _detailsRefreshDebounce = Debouncer(
    const Duration(milliseconds: 150),
  );
  // Фоновый пуллинг списка сессий как запасной канал обновления

  final FocusNode _searchFocus = FocusNode();
  MonitorListener? _monitorListener;

  @override
  void initState() {
    super.initState();
    _connectMonitor();
    _restorePrefs().then((_) {
      // После восстановления сохранённых фильтров сразу подгружаем список
      _loadSessions();
    });

    _framesCtrl.addListener(_onFramesScroll);
    _eventsCtrl.addListener(_onEventsScroll);
    _sessionsCtrl.addListener(_onSessionsScroll);
    _sessionSearchCtrl.addListener(() {
      sl<HomeUiStore>().setSessionSearchQuery(_sessionSearchCtrl.text);
      _scheduleSessionsReload();
    });
    // Фоновая подзагрузка через сервис
    sl<features_inspector_application_services.SessionsPollingService>().start(
      onTick: () async {
        if (!mounted) return;
        final ui = sl<HomeUiStore>();
        if (!ui.isRecording.value) return;
        final store = context.read<SessionsStore>();
        if (!store.loading && !_loadingSessions) {
          await _loadSessions();
        }
      },
    );
  }

  @override
  void dispose() {
    final monitor = sl<MonitorService>();
    if (_monitorListener != null) {
      monitor.removeListener(_monitorListener!);
    }
    monitor.dispose();
    _pollTimer?.cancel();
    _sessionsReloadDebounce.dispose();
    sl<features_inspector_application_services.SessionsPollingService>().stop();
    _framesCtrl.dispose();
    _eventsCtrl.dispose();
    _sessionsCtrl.dispose();
    _searchFocus.dispose();
    super.dispose();
  }

  void _connectMonitor() {
    final monitor = sl<MonitorService>();
    final listener = (Map<String, dynamic> ev) {
      try {
        final ui = sl<HomeUiStore>();
        final t = (ev['type'] ?? '').toString();
        if (t == 'session_started' || t == 'session_ended') {
          if (!ui.isRecording.value) return;
          if (_loadingSessions) return;
          _scheduleSessionsReload();
          Future.microtask(() {
            try {
              context.read<AggregateStore>().load(groupBy: 'domain');
            } catch (_) {}
          });
        }
        if (!ui.isRecording.value && t == 'session_started') {
          return; // paused: не подхватываем обновления
        }
        if (t == 'frame_added' || t == 'event_added' || t == 'sio_probe') {
          final sid = (ev['id'] ?? '').toString();
          if (ui.selectedSessionId.value != null &&
              sid == ui.selectedSessionId.value) {
            _tickRefresh();
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

  void _scheduleSessionsReload() {
    _sessionsReloadDebounce.run(() {
      final ui = sl<HomeUiStore>();
      if (!ui.isRecording.value) return;
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
    // сразу перезагрузим агрегатор доменов, чтобы счётчики обнулились мгновенно
    try {
      await context.read<AggregateStore>().load(groupBy: 'domain');
    } catch (_) {}
    // temporarily suppress auto-reload caused by monitor events
    _sessionsReloadDebounce.dispose();
    _loadingSessions = true;
    await Future.delayed(const Duration(milliseconds: 300));
    _loadingSessions = false;
    // final reload to confirm empty state
    await _loadSessions();
  }

  Future<void> _loadSessions() async {
    final store = context.read<SessionsStore>();
    final q = sl<HomeUiStore>().sessionSearchQuery.value.trim();
    final target = context.read<SessionsFiltersStore>().target.trim();
    await store.load(q: q, target: target);
    _suckMetaFromSessions();
    // fire and forget lightweight warmup to enrich httpMeta
    // ignore: discarded_futures
    sl<HttpMetaService>().warmup(limit: 50);
    try {
      await context.read<AggregateStore>().load(groupBy: 'domain');
    } catch (_) {}
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
    // восстановим фильтры в Store
    final f = context.read<SessionsFiltersStore>();
    f.setTarget(data['targetFilter'] ?? '');
    f.setHttpMethod(data['httpMethod'] ?? 'any');
    f.setHttpStatus(data['httpStatus'] ?? 'any');
    f.setHttpMime(data['httpMime'] ?? '');
    f.setHttpMinDurationMs(int.tryParse(data['httpMinDuration'] ?? '0') ?? 0);
    f.setGroupBy(data['groupBy'] ?? 'none');
    f.setHeaderKey(data['headerKey'] ?? '');
    f.setHeaderVal(data['headerVal'] ?? '');
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
    final ui = sl<HomeUiStore>();
    if (ui.selectedSessionId.value == null) return;
    _pollTimer = Timer.periodic(
      const Duration(seconds: 2),
      (_) => _tickRefresh(),
    );
  }

  Future<void> _tickRefresh() async {
    final ui = sl<HomeUiStore>();
    if (ui.selectedSessionId.value == null) return;
    _detailsRefreshDebounce.run(() async {
      try {
        await Future.wait([
          context.read<SessionDetailsStore>().loadMoreFrames(),
          context.read<SessionDetailsStore>().loadMoreEvents(),
        ]);
      } catch (_) {}
    });
  }

  void _onSessionsScroll() {
    if (!_sessionsCtrl.hasClients) return;
    final pos = _sessionsCtrl.position;
    // Считаем "внизу", если остался небольшой хвост (для стабилизации на резайзах)
    (pos.maxScrollExtent - pos.pixels) < 48;
  }

  // фильтры перенесены в WsDetailsPanel

  Future<void> _deleteSelected() async {
    final ui = sl<HomeUiStore>();
    if (ui.selectedSessionId.value == null) return;
    final id = ui.selectedSessionId.value!;
    // ignore: invalid_use_of_protected_member
    final client = sl.get<Object>();
    try {
      await (client as dynamic).delete(path: '/_api/v1/sessions/$id');
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
                                  isRecording: ui.isRecording.value,
                                  onToggleRecording: () async {
                                    // Toggle on backend via capture API
                                    final newVal = !ui.isRecording.value;
                                    try {
                                      final client =
                                          sl.get<Object>() as dynamic;
                                      await client.post(
                                        path: '/_api/v1/capture',
                                        body: {
                                          'action': newVal ? 'start' : 'stop',
                                        },
                                      );
                                    } catch (_) {}
                                    ui.setIsRecording(newVal);
                                    try {
                                      await PrefsService().saveIsRecording(
                                        newVal,
                                      );
                                    } catch (_) {}
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
                                child:
                                    ui.showTimeline.value
                                        ? TimelineBlock(
                                          since: ui.since.value,
                                          wfFitAll: ui.wfFitAll.value,
                                          onFitAllChanged:
                                              (v) => ui.setWfFitAll(v),
                                          onSelectSession: (id) {
                                            ui.setSelectedSessionId(id);
                                            _loadDetails(id);
                                          },
                                          onClearAllSessions: _clearAllSessions,
                                          selectedRange: ui.selectedRange.value,
                                          onRangeChanged:
                                              (range) =>
                                                  ui.setSelectedRange(range),
                                          onRangeCleared:
                                              () => ui.clearSelectedRange(),
                                        )
                                        : const SizedBox.shrink(),
                              );
                            },
                          ),
                          const SizedBox(height: 8),
                          Observer(
                            builder: (_) {
                              final show = sl<HomeUiStore>().showFilters.value;
                              return AnimatedSize(
                                duration: const Duration(milliseconds: 220),
                                curve: Curves.easeInOutCubic,
                                alignment: Alignment.topCenter,
                                child:
                                    show
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
                                // если есть selectedSessionId, то отображаем details panel
                                Observer(
                                  builder: (_) {
                                    final has =
                                        sl<HomeUiStore>()
                                            .selectedSessionId
                                            .value !=
                                        null;
                                    if (!has) return const SizedBox.shrink();
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
                                                if (sl<HomeUiStore>()
                                                        .selectedSessionId
                                                        .value !=
                                                    null) {
                                                  final items =
                                                      context
                                                          .watch<
                                                            SessionsStore
                                                          >()
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
                                                        if (selIsWs &&
                                                            selIsHttp)
                                                          return 2;
                                                        if (selIsWs ||
                                                            selIsHttp)
                                                          return 1;
                                                        return 2;
                                                      })(),
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
                                                                  'id': f.id,
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
                                                                  'id': e.id,
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
                                                      return DetailsTabs(
                                                        showWs: selIsWs,
                                                        showHttp: selIsHttp,
                                                        frames:
                                                            frames
                                                                .cast<
                                                                  Map<
                                                                    String,
                                                                    dynamic
                                                                  >
                                                                >(),
                                                        events:
                                                            events
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
                                                        onToggleHeartbeats: (
                                                          v,
                                                        ) {
                                                          sl<HomeUiStore>()
                                                              .setHideHeartbeats(
                                                                v,
                                                              );
                                                          _savePrefs();
                                                        },
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

  void _suckMetaFromSessions() {
    final src = context.read<SessionsStore>().items.toList();
    for (final s in src) {
      final meta = s.httpMeta;
      if (meta != null && meta.isNotEmpty) {
        sl<HomeUiStore>().httpMeta[s.id] = Map<String, dynamic>.from(meta);
      }
    }
  }
}
