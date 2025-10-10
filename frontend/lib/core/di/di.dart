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
import '../../features/inspector/application/services/http_meta_service.dart';
import '../../features/inspector/application/services/sessions_polling_service.dart';
import '../../features/inspector/application/services/recent_window_service.dart';

final sl = GetIt.instance;

Future<void> setupDI({required String baseUrl}) async {
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
  sl.registerLazySingleton<HttpMetaService>(() => HttpMetaService());
  sl.registerLazySingleton<SessionsPollingService>(
    () => SessionsPollingService(),
  );
  sl.registerLazySingleton<RecentWindowService>(() => RecentWindowService());
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
}
