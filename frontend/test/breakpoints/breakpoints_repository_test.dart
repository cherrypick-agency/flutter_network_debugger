import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:app_http_client/application/app_http_client.dart';
import 'package:app_http_client/application/http_method.dart';

import 'package:frontend/features/breakpoints/data/breakpoints_api.dart';
import 'package:frontend/features/breakpoints/data/breakpoints_repository_impl.dart';

void main() {
  group('BreakpointsRepositoryImpl mapping', () {
    late _FakeApi api;
    late BreakpointsRepositoryImpl repo;

    setUp(() {
      api = _FakeApi();
      repo = BreakpointsRepositoryImpl(api);
    });

    test('getConfig maps fields correctly', () async {
      api.config = {
        'enabled': true,
        'requests': true,
        'responses': false,
        'timeoutMs': 15000,
        'queueMax': 42,
        'bodyMaxBytes': 65536,
        'reencode': true,
        'overflow': 'auto-continue-oldest',
      };
      final cfg = await repo.getConfig();
      expect(cfg.enabled, true);
      expect(cfg.requests, true);
      expect(cfg.responses, false);
      expect(cfg.timeoutMs, 15000);
      expect(cfg.queueMax, 42);
      expect(cfg.bodyMaxBytes, 65536);
      expect(cfg.reencode, true);
      expect(cfg.overflow, 'auto-continue-oldest');
    });

    test('listRules maps nested structures', () async {
      api.rules = [
        {
          'id': 'r1',
          'enabled': true,
          'priority': 10,
          'action': 'both',
          'once': false,
          'stopProcessing': true,
          'when': {
            'method': ['POST'],
            'scheme': ['https'],
            'host': {'suffix': '.example.com'},
            'path': {'prefix': '/api/'},
            'contentType': {'prefix': 'application/json'},
            'responseStatus': {'is4xx': true},
            'header': {
              'name': {'equals': 'Authorization'},
              'value': {'prefix': 'Bearer '},
            },
            'bodyContains': 'token',
          },
        },
      ];
      final rules = await repo.listRules();
      expect(rules.length, 1);
      final r = rules.first;
      expect(r.id, 'r1');
      expect(r.enabled, true);
      expect(r.priority, 10);
      expect(r.action, 'both');
      expect(r.when.method, ['POST']);
      expect(r.when.host?.suffix, '.example.com');
      expect(r.when.path?.prefix, '/api/');
      expect(r.when.contentType?.prefix, 'application/json');
      expect(r.when.responseStatus?.is4xx, true);
      expect(r.when.header?.name.equals, 'Authorization');
      expect(r.when.header?.name.suffix, '');
      expect(r.when.header?.name.prefix, '');
      expect(r.when.header?.name.contains, '');
      expect(r.when.header?.name.anyOf, isEmpty);
      expect(r.when.header?.value.prefix, 'Bearer ');
    });

    test('listPending / getItem maps request and response snapshots', () async {
      final now = DateTime.now().toUtc();
      final body = base64Encode(utf8.encode('{"ok":true}'));
      api.pending = [
        {
          'id': 'i1',
          'createdAt': now.toIso8601String(),
          'deadline': now.add(const Duration(seconds: 30)).toIso8601String(),
          'direction': 'request',
          'sessionId': 's1',
          'state': 'PENDING',
          'req': {
            'method': 'POST',
            'url': 'https://api.example.com/v1',
            'headers': {
              'Content-Type': ['application/json'],
              'Authorization': ['Bearer abc'],
            },
            'bodyBase64': body,
            'bodyTruncated': false,
            'contentType': 'application/json',
          },
        },
      ];

      final items = await repo.listPending();
      expect(items.length, 1);
      final it = items.first;
      expect(it.id, 'i1');
      expect(it.direction, 'request');
      expect(it.req?.method, 'POST');
      expect(it.req?.headers['Authorization']?.first, 'Bearer abc');
      expect(it.req?.bodyBase64, body);

      // getItem mirrors the same mapping
      api.item = api.pending.first;
      final it2 = await repo.getItem('i1');
      expect(it2.id, 'i1');
      expect(it2.req?.contentType, 'application/json');
    });
  });
}

class _FakeApi extends BreakpointsApi {
  _FakeApi() : super(_FakeHttp());

  Map<String, dynamic> config = const {};
  List<dynamic> rules = const [];
  List<dynamic> pending = const [];
  dynamic item;

  @override
  Future<Map<String, dynamic>> getConfig() async => config;

  @override
  Future<void> setConfig(Map<String, dynamic> body) async {}

  @override
  Future<List<dynamic>> listRules() async => rules;

  @override
  Future<void> replaceRules(List<dynamic> rules) async {}

  @override
  Future<List<dynamic>> listPending({int? limit}) async => pending;

  @override
  Future<Map<String, dynamic>> getItem(String id) async => _toMap(item);

  Map<String, dynamic> _toMap(dynamic v) {
    if (v is Map<String, dynamic>) return v;
    if (v == null) return <String, dynamic>{};
    throw StateError('Unsupported');
  }
}

class _FakeHttp implements AppHttpClient {
  @override
  String get defaultHost => '';
  @override
  String? get token => null;
  @override
  String? get refreshToken => null;
  @override
  Future<void> clearTokens() async {}
  @override
  Future<Response<T>> delete<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Object? body,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> filesPatch<T>({
    String? host,
    String path = '',
    Map<String, String>? headers,
    Map<String, String>? query,
    required files,
    Map<String, String>? fields,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> filesPost<T>({
    String? host,
    String path = '',
    Map<String, String>? headers,
    Map<String, String>? query,
    required files,
    Map<String, String>? fields,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> get<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
  }) async => throw UnimplementedError();
  @override
  Future<void> init() async {}
  @override
  String get refreshPath => '';
  @override
  Future<void> refresh() async {}
  @override
  Future<void> rememberTokens(String? token, String? refreshToken) async {}
  @override
  Future<Response<T>> patch<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Object? body,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> performRequest<T>({
    String? host,
    String path = '',
    Map<String, dynamic>? query,
    Map<String, dynamic>? headers,
    Object? body,
    required HttpMethod method,
    Map<String, String>? fields,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> post<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Object? body,
  }) async => throw UnimplementedError();
  @override
  Future<Response<T>> put<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Object? body,
  }) async => throw UnimplementedError();
  @override
  Future<List<int>> getBytes({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
  }) async => throw UnimplementedError();
}
