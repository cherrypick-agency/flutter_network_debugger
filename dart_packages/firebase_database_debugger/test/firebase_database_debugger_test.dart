import 'dart:convert';

import 'package:firebase_database_debugger/firebase_database_debugger.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  test('uses stable short session id and sends ingest payload', () async {
    late Uri capturedUri;
    late Map<String, String> capturedHeaders;
    late Map<String, dynamic> capturedBody;

    final client = MockClient((http.Request request) async {
      capturedUri = request.url;
      capturedHeaders = request.headers;
      capturedBody = jsonDecode(request.body) as Map<String, dynamic>;
      return http.Response('', 204);
    });

    final debugger = FirebaseDatabaseDebugger(
      config: FirebaseDatabaseDebuggerConfig(
        debuggerBaseUrl: 'http://localhost:9092',
        databaseUrl: 'https://demo-default-rtdb.firebaseio.com',
        adminToken: 'secret',
        maxBatchFrames: 100,
      ),
      httpClient: client,
    );

    await debugger.logOperation(
      path: '/very/long/path/that/should/not/be/used/as/session/id/directly',
      op: 'set',
      direction: 'client->upstream',
      payload: <String, dynamic>{'value': 'ok'},
    );
    await debugger.flush();
    await debugger.dispose();

    expect(
      capturedUri.toString(),
      'http://localhost:9092/_api/v1/ingest/firebase_database',
    );
    expect(capturedHeaders['X-Admin-Token'], 'secret');

    final session = capturedBody['session'] as Map<String, dynamic>;
    final sessionId = (session['id'] as String?) ?? '';
    expect(sessionId.startsWith('fb-'), isTrue);
    expect(sessionId.length, lessThanOrEqualTo(128));

    final frames = (capturedBody['frames'] as List<dynamic>);
    expect(frames.length, 1);
    final frame = frames.first as Map<String, dynamic>;
    expect(frame['direction'], 'client->upstream');
    expect(frame['opcode'], 'text');
    expect((frame['preview'] as String).contains('"type":"firebase_database"'),
        isTrue);
  });

  test('spills large preview into body base64', () async {
    late Map<String, dynamic> capturedBody;
    final client = MockClient((http.Request request) async {
      capturedBody = jsonDecode(request.body) as Map<String, dynamic>;
      return http.Response('', 204);
    });

    final debugger = FirebaseDatabaseDebugger(
      config: FirebaseDatabaseDebuggerConfig(
        debuggerBaseUrl: 'http://localhost:9092',
        previewBodyThresholdBytes: 64,
      ),
      httpClient: client,
    );

    await debugger.logOperation(
      path: '/users/123',
      op: 'set',
      direction: 'client->upstream',
      payload: <String, dynamic>{'value': List.filled(6000, 'x').join()},
    );
    await debugger.flush();
    await debugger.dispose();

    final frames = capturedBody['frames'] as List<dynamic>;
    final frame = frames.first as Map<String, dynamic>;
    expect(frame['bodyEncoding'], 'base64');
    expect((frame['body'] as String).isNotEmpty, isTrue);
    expect((frame['preview'] as String).contains('"bodySpilled":true'), isTrue);
  });
}
