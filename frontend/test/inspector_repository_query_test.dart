import 'package:app_http_client/application/app_http_client.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/inspector/data/inspector_repository_impl.dart';

class _RecordingHttpClient implements AppHttpClient {
  String? lastPath;
  Map<String, String>? lastQuery;

  @override
  Future<Response<T>> get<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
  }) async {
    lastPath = path;
    lastQuery = query;
    return Response<T>(
      requestOptions: RequestOptions(path: path),
      data: <String, dynamic>{'items': <Map<String, dynamic>>[]} as T,
    );
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

void main() {
  test('serializes all-captures, paused and GraphQL filters', () async {
    final client = _RecordingHttpClient();
    final repository = InspectorRepositoryImpl(client);

    final sessions = await repository.listSessions(
      captureScope: 'all',
      includePaused: true,
      types: {'graphql'},
      statusGroups: {'5xx'},
    );

    expect(sessions, isEmpty);
    expect(client.lastPath, '/_api/v1/sessions');
    expect(client.lastQuery, {
      'limit': '1000',
      'captures': 'all',
      'includeUnassigned': 'true',
      'types': 'graphql',
      'status': '5xx',
      'scan': 'graphql',
    });
  });

  test('uses current capture without paused sessions or GraphQL scan', () async {
    final client = _RecordingHttpClient();
    final repository = InspectorRepositoryImpl(client);

    await repository.listSessions(types: {'http'}, statusGroups: {'4xx'});

    expect(client.lastQuery, {
      'limit': '1000',
      'captureId': 'current',
      'types': 'http',
      'status': '4xx',
    });
  });

  test('does not trigger deep scan for no selected types', () async {
    final client = _RecordingHttpClient();
    final repository = InspectorRepositoryImpl(client);

    await repository.listSessions();

    expect(client.lastQuery, {'limit': '1000', 'captureId': 'current'});
  });
}
