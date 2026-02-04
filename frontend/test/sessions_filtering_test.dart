import 'package:flutter_test/flutter_test.dart';
import 'package:mobx/mobx.dart';
import 'package:frontend/features/inspector/application/stores/sessions_store.dart';
import 'package:frontend/features/inspector/application/usecases/list_sessions.dart';
import 'package:frontend/features/inspector/domain/repositories/inspector_repository.dart';
import 'package:frontend/features/inspector/domain/entities/session.dart';
import 'package:frontend/features/inspector/domain/entities/frame.dart';
import 'package:frontend/features/inspector/domain/entities/event.dart';
import 'package:frontend/features/filters/application/stores/sessions_filters_store.dart';
import 'package:frontend/features/inspector/presentation/utils/sessions_filtering.dart';

class _FakeRepo implements InspectorRepository {
  @override
  Future<List<Session>> listSessions({
    String? q,
    String? target,
    List<String>? tags,
  }) async => [];
  @override
  Future<List<Map<String, dynamic>>> aggregateSessions({
    String groupBy = 'domain',
  }) async => [];
  @override
  Future<List<EventEntity>> listEvents(
    String sessionId, {
    String? from,
    int limit = 100,
  }) async => <EventEntity>[];
  @override
  Future<List<Frame>> listFrames(
    String sessionId, {
    String? from,
    int limit = 100,
  }) async => <Frame>[];

  @override
  Future<String> exportHAR({
    List<String>? sessionIds,
    bool includeBodies = true,
    bool includeSensitive = false,
    bool minify = false,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<Map<String, dynamic>> importHAR(
    String harJson, {
    required String mode,
  }) async {
    throw UnimplementedError();
  }
}

void main() {
  late SessionsStore store;
  late SessionsFiltersStore filters;
  late Map<String, Map<String, dynamic>> httpMeta;

  setUp(() {
    store = SessionsStore(ListSessionsUseCase(_FakeRepo()));
    filters = SessionsFiltersStore();
    httpMeta = {};

    final now = DateTime.now().toUtc();
    final s1 = Session(
      id: '1',
      target: 'https://a.com/path',
      startedAt: now.subtract(const Duration(minutes: 10)),
      closedAt: now.subtract(const Duration(minutes: 9)),
      kind: 'http',
      httpMeta: {
        'method': 'GET',
        'status': 200,
        'mime': 'application/json',
        'durationMs': 300,
        'headers': {'x-req-id': 'abc123'},
      },
    );
    final s2 = Session(
      id: '2',
      target: 'https://b.com/path',
      startedAt: now.subtract(const Duration(minutes: 5)),
      closedAt: now.subtract(const Duration(minutes: 4)),
      kind: 'http',
      httpMeta: {'method': 'POST', 'status': 404, 'durationMs': 50},
    );
    final s3 = Session(
      id: '3',
      target: 'https://a.com/api',
      startedAt: now.subtract(const Duration(minutes: 2)),
      closedAt: now.subtract(const Duration(minutes: 1)),
      kind: 'http',
      httpMeta: {
        'method': 'GET',
        'status': 503,
        'durationMs': 1200,
        'mime': 'text/html',
      },
    );
    final s4 = Session(
      id: '4',
      target: 'https://c.com/x',
      startedAt: now.subtract(const Duration(minutes: 20)),
      closedAt: now.subtract(const Duration(minutes: 19)),
      kind: 'http',
      httpMeta: null,
    );
    httpMeta['4'] = {
      'method': 'PUT',
      'status': 201,
      'durationMs': 700,
      'mime': 'text/plain',
    };

    store.items = ObservableList.of([s1, s2, s3, s4]);
  });

  test('filter by httpMethod=GET', () {
    // Arrange
    filters.setHttpMethod('GET');

    // Act
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );

    // Assert
    final ids = res.map((e) => e.id).toList();
    expect(ids, containsAll(<String>['1', '3']));
    expect(ids, isNot(contains('2')));
  });

  test('filter by httpStatus=4xx', () {
    filters.setHttpStatus('4xx');
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id), ['2']);
  });

  test('filter by minimum duration', () {
    filters.setHttpMinDurationMs(600);
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id).toList(), containsAll(<String>['3', '4']));
    expect(res.map((e) => e.id), isNot(contains('1')));
  });

  test('filter by mime contains json', () {
    filters.setHttpMime('json');
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id).toList(), ['1']);
  });

  test('filter by headers (key is required, value is optional)', () {
    filters
      ..setHeaderKey('x-req-id')
      ..setHeaderVal('abc');
    var res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id).toList(), ['1']);

    // If value doesn't match - result is empty
    filters.setHeaderVal('zzz');
    res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res, isEmpty);

    // If value is empty but key exists - returns only if header is present
    filters.setHeaderVal('');
    res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id).toList(), ['1']);
  });

  test('selectedDomains filters by target domain', () {
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: {'a.com'},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    expect(res.map((e) => e.id).toList(), containsAll(<String>['1', '3']));
    expect(res.map((e) => e.id), isNot(contains('2')));
  });

  test('since excludes sessions closed before threshold', () {
    final since = DateTime.now().toUtc().subtract(const Duration(minutes: 8));
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: since,
      ignoredIds: <String>{},
    );
    // s1 was closed 9 minutes ago - should be filtered out
    expect(res.map((e) => e.id), isNot(contains('1')));
  });

  test('selectedRange: includes sessions intersecting the range', () {
    final now = DateTime.now().toUtc();
    final range = DateTimeRangeWrapper(
      start: now.subtract(const Duration(minutes: 6)),
      end: now.subtract(const Duration(minutes: 3)),
    );
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: range,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    // s2 (5..4 min ago) falls within range, while s1 (10..9) does not
    final ids = res.map((e) => e.id).toList();
    expect(ids, contains('2'));
    expect(ids, isNot(contains('1')));
  });

  test('groupBy=domain sorts by domain', () {
    filters.setGroupBy('domain');
    final res = filterVisibleSessions(
      store: store,
      filters: filters,
      selectedRange: null,
      selectedDomains: <String>{},
      httpMeta: httpMeta,
      since: null,
      ignoredIds: <String>{},
    );
    final targets = res.map((e) => e.target).toList();
    expect(targets.first, startsWith('https://a.com'));
    expect(targets.last, startsWith('https://c.com'));
  });
}
