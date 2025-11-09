import 'package:get_it/get_it.dart';
import 'package:modules_basis/modules_basis.dart';
import 'package:app_http_client/app_http_client.dart' as module_entry;
import 'package:app_http_client/application/app_http_client.dart';
import '../../features/inspector/data/inspector_repository_impl.dart';
import '../../features/inspector/domain/repositories/inspector_repository.dart';
import '../../features/inspector/application/usecases/list_sessions.dart';
import '../../features/inspector/application/usecases/list_frames.dart';
import '../../features/inspector/application/usecases/list_events.dart';
import '../../features/inspector/application/stores/sessions_store.dart';
import '../../features/inspector/application/stores/session_details_store.dart';
import '../../features/inspector/application/usecases/list_aggregate.dart';
import '../../features/inspector/application/stores/aggregate_store.dart';
import '../notifications/notifications_service.dart';
import '../hotkeys/hotkeys_service.dart';
import '../../features/settings/application/settings_service.dart';
import '../../features/filters/application/stores/sessions_filters_store.dart';
import '../../features/inspector/application/stores/home_ui_store.dart';
import '../../features/inspector/application/services/monitor_service.dart';
import '../realtime/socket_io_service.dart';
import '../../features/inspector/application/services/realtime_inspector_service.dart';
import '../../features/inspector/application/services/http_meta_service.dart';
import '../../features/inspector/application/services/sessions_polling_service.dart';
import '../../features/inspector/application/services/recent_window_service.dart';
import '../../features/compose/data/compose_repository.dart';
import '../../features/compose/application/compose_store.dart';
import '../../features/breakpoints/data/breakpoints_api.dart';
import '../../features/breakpoints/data/breakpoints_repository_impl.dart';
import '../../features/breakpoints/domain/repositories/breakpoints_repository.dart';
import '../../features/breakpoints/application/stores/breakpoints_store.dart';
import '../../features/breakpoints/application/stores/intercept_queue_store.dart';
import '../../features/breakpoints/application/stores/intercept_editor_store.dart';
import '../../features/updates/data/datasources/github_api_datasource.dart';
import '../../features/updates/data/datasources/updates_local_datasource.dart';
import '../../features/updates/data/repositories/updates_repository_impl.dart';
import '../../features/updates/domain/repositories/updates_repository.dart';
import '../../features/updates/application/services/updates_service.dart';
import '../../features/updates/application/stores/updates_store.dart';
import '../../features/scripts/data/services/scripts_api_service.dart';
import '../../features/scripts/data/repositories/scripts_repository_impl.dart';
import '../../features/scripts/domain/repositories/scripts_repository.dart';
import '../../features/scripts/domain/usecases/create_script_usecase.dart';
import '../../features/scripts/domain/usecases/update_script_usecase.dart';
import '../../features/scripts/domain/usecases/delete_script_usecase.dart';
import '../../features/scripts/domain/usecases/get_scripts_usecase.dart';
import '../../features/scripts/domain/usecases/toggle_script_usecase.dart';
import '../../features/scripts/domain/usecases/test_script_usecase.dart';
import '../../features/scripts/domain/usecases/compile_script_usecase.dart';
import '../../features/scripts/domain/usecases/export_script_usecase.dart';
import '../../features/scripts/domain/usecases/import_script_usecase.dart';
import '../../features/scripts/application/stores/scripts_store.dart';
import '../../features/scripts/application/stores/script_editor_store.dart';
import '../../features/compiler_management/data/datasources/compiler_api.dart';
import '../../features/compiler_management/data/repositories/compiler_repository_impl.dart';
import '../../features/compiler_management/domain/repositories/compiler_repository.dart';
import '../../features/compiler_management/domain/usecases/get_compilers_usecase.dart';
import '../../features/compiler_management/domain/usecases/install_compiler_usecase.dart';
import '../../features/compiler_management/domain/usecases/uninstall_compiler_usecase.dart';
import '../../features/compiler_management/domain/usecases/watch_progress_usecase.dart';
import '../../features/compiler_management/presentation/stores/compiler_list_store.dart';
import '../../features/compiler_management/presentation/stores/installation_progress_store.dart';
import '../../features/tags/data/tags_api.dart';
import '../../features/tags/data/tags_repository_impl.dart';
import '../../features/tags/domain/repositories/tags_repository.dart';
import '../../features/tags/application/stores/tags_store.dart';
import '../../features/performance/data/performance_api.dart';
import '../../features/performance/data/performance_repository_impl.dart';
import '../../features/performance/domain/repositories/performance_repository.dart';
import '../../features/performance/application/stores/performance_store.dart';

final sl = GetIt.instance;
final getIt = sl; // Alias for consistency

Future<void> setupDI({
  required String baseUrl,
  required String githubOwner,
  required String githubRepo,
  required String currentVersion,
}) async {
  // init http module (как в qovo_flutter)
  final container = ContainerDI(sl);
  // tokens storage внутри модуля; baseURL как лямбда
  final module = module_entry.AppHttpClientModule(
    () => baseUrl,
    (_) {},
    container,
  );
  await module.execute();
  // TODO Injectable
  // Repository
  sl.registerLazySingleton<InspectorRepository>(
    () => InspectorRepositoryImpl(sl<AppHttpClient>()),
  );
  // Use cases
  sl.registerLazySingleton<ListSessionsUseCase>(
    () => ListSessionsUseCase(sl<InspectorRepository>()),
  );
  sl.registerLazySingleton<ListFramesUseCase>(
    () => ListFramesUseCase(sl<InspectorRepository>()),
  );
  sl.registerLazySingleton<ListEventsUseCase>(
    () => ListEventsUseCase(sl<InspectorRepository>()),
  );
  sl.registerLazySingleton<ListAggregateUseCase>(
    () => ListAggregateUseCase(sl<InspectorRepository>()),
  );
  // Stores
  sl.registerLazySingleton<SessionsStore>(
    () => SessionsStore(sl<ListSessionsUseCase>()),
  );
  sl.registerLazySingleton<SessionDetailsStore>(
    () => SessionDetailsStore(sl<ListFramesUseCase>(), sl<ListEventsUseCase>()),
  );
  sl.registerLazySingleton<AggregateStore>(
    () => AggregateStore(sl<ListAggregateUseCase>()),
  );
  // UI store
  sl.registerLazySingleton<HomeUiStore>(() => HomeUiStore());
  // Services
  sl.registerLazySingleton<MonitorService>(() => MonitorService());
  sl.registerLazySingleton<SocketIoService>(() => SocketIoService());
  sl.registerLazySingleton<RealtimeInspectorService>(
    () => RealtimeInspectorService(),
  );
  sl.registerLazySingleton<HttpMetaService>(() => HttpMetaService());
  sl.registerLazySingleton<SessionsPollingService>(
    () => SessionsPollingService(),
  );
  sl.registerLazySingleton<RecentWindowService>(() => RecentWindowService());
  // Compose repository
  sl.registerLazySingleton<ComposeRepository>(
    () => ComposeRepository(sl<AppHttpClient>()),
  );
  sl.registerLazySingleton<ComposeStore>(
    () => ComposeStore(sl<ComposeRepository>()),
  );
  // Filters store
  sl.registerLazySingleton<SessionsFiltersStore>(() => SessionsFiltersStore());
  // Notifications
  sl.registerLazySingleton<NotificationsService>(() => NotificationsService());
  // Hotkeys
  final hk = HotkeysService();
  await hk.init();
  sl.registerSingleton<HotkeysService>(hk);
  // Settings service
  sl.registerLazySingleton<SettingsService>(() => SettingsService());
  // Первичная синхронизация настроек (задержка ответа)
  // ignore: unawaited_futures
  sl<SettingsService>().syncPrefsToBackend();
  // Recent window init
  // ignore: unawaited_futures
  sl<RecentWindowService>().initFromPrefs();

  // Breakpoints feature
  sl.registerLazySingleton<BreakpointsApi>(
    () => BreakpointsApi(sl<AppHttpClient>()),
  );
  sl.registerLazySingleton<BreakpointsRepository>(
    () => BreakpointsRepositoryImpl(sl<BreakpointsApi>()),
  );
  sl.registerLazySingleton<BreakpointsStore>(
    () => BreakpointsStore(sl<BreakpointsRepository>()),
  );
  sl.registerLazySingleton<InterceptQueueStore>(
    () => InterceptQueueStore(sl<BreakpointsRepository>()),
  );
  sl.registerLazySingleton<InterceptEditorStore>(
    () => InterceptEditorStore(sl<BreakpointsRepository>()),
  );

  // Updates feature
  sl.registerLazySingleton<GitHubApiDataSource>(
    () => GitHubApiDataSource(githubOwner: githubOwner, githubRepo: githubRepo),
  );
  sl.registerLazySingleton<UpdatesLocalDataSource>(
    () => UpdatesLocalDataSource(),
  );
  sl.registerLazySingleton<UpdatesRepository>(
    () => UpdatesRepositoryImpl(
      githubApi: sl<GitHubApiDataSource>(),
      localStorage: sl<UpdatesLocalDataSource>(),
      currentVersion: currentVersion,
    ),
  );
  sl.registerLazySingleton<UpdatesService>(
    () => UpdatesService(repository: sl<UpdatesRepository>()),
  );
  sl.registerLazySingleton<UpdatesStore>(
    () => UpdatesStore(sl<UpdatesService>()),
  );

  // Scripts feature
  sl.registerLazySingleton<ScriptsApiService>(
    () => ScriptsApiService(sl<AppHttpClient>()),
  );
  sl.registerLazySingleton<ScriptsRepository>(
    () => ScriptsRepositoryImpl(sl<ScriptsApiService>()),
  );
  // Use cases
  sl.registerLazySingleton<GetScriptsUseCase>(
    () => GetScriptsUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<CreateScriptUseCase>(
    () => CreateScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<UpdateScriptUseCase>(
    () => UpdateScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<DeleteScriptUseCase>(
    () => DeleteScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<ToggleScriptUseCase>(
    () => ToggleScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<TestScriptUseCase>(
    () => TestScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<CompileScriptUseCase>(
    () => CompileScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<ExportScriptUseCase>(
    () => ExportScriptUseCase(sl<ScriptsRepository>()),
  );
  sl.registerLazySingleton<ImportScriptUseCase>(
    () => ImportScriptUseCase(sl<ScriptsRepository>()),
  );
  // Stores
  sl.registerLazySingleton<ScriptsStore>(
    () => ScriptsStore(
      getScriptsUseCase: sl<GetScriptsUseCase>(),
      createUseCase: sl<CreateScriptUseCase>(),
      updateUseCase: sl<UpdateScriptUseCase>(),
      deleteUseCase: sl<DeleteScriptUseCase>(),
      toggleUseCase: sl<ToggleScriptUseCase>(),
      compileUseCase: sl<CompileScriptUseCase>(),
      exportUseCase: sl<ExportScriptUseCase>(),
      importUseCase: sl<ImportScriptUseCase>(),
    ),
  );
  sl.registerFactory<ScriptEditorStore>(
    () => ScriptEditorStore(testUseCase: sl<TestScriptUseCase>()),
  );

  // Compiler Management feature
  sl.registerLazySingleton<CompilerApi>(() => CompilerApi(baseUrl: baseUrl));
  sl.registerLazySingleton<CompilerRepository>(
    () => CompilerRepositoryImpl(sl<CompilerApi>()),
  );
  // Use cases
  sl.registerLazySingleton<GetCompilersUseCase>(
    () => GetCompilersUseCase(sl<CompilerRepository>()),
  );
  sl.registerLazySingleton<InstallCompilerUseCase>(
    () => InstallCompilerUseCase(sl<CompilerRepository>()),
  );
  sl.registerLazySingleton<UninstallCompilerUseCase>(
    () => UninstallCompilerUseCase(sl<CompilerRepository>()),
  );
  sl.registerLazySingleton<WatchProgressUseCase>(
    () => WatchProgressUseCase(sl<CompilerRepository>()),
  );
  // Stores
  sl.registerLazySingleton<CompilerListStore>(
    () => CompilerListStore(
      getCompilers: sl<GetCompilersUseCase>(),
      installCompiler: sl<InstallCompilerUseCase>(),
      uninstallCompiler: sl<UninstallCompilerUseCase>(),
    ),
  );
  sl.registerLazySingleton<InstallationProgressStore>(
    () => InstallationProgressStore(watchProgress: sl<WatchProgressUseCase>()),
  );

  // Tags feature
  sl.registerLazySingleton<TagsApi>(() => TagsApi(sl<AppHttpClient>()));
  sl.registerLazySingleton<TagsRepository>(
    () => TagsRepositoryImpl(sl<TagsApi>()),
  );
  sl.registerLazySingleton<TagsStore>(() => TagsStore(sl<TagsRepository>()));

  // Performance feature
  sl.registerLazySingleton<PerformanceApi>(
    () => PerformanceApi(sl<AppHttpClient>()),
  );
  sl.registerLazySingleton<PerformanceRepository>(
    () => PerformanceRepositoryImpl(sl<PerformanceApi>()),
  );
  sl.registerLazySingleton<PerformanceStore>(
    () => PerformanceStore(sl<PerformanceRepository>()),
  );
}
