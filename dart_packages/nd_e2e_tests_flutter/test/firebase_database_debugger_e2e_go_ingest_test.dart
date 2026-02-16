import 'package:firebase_database_debugger/firebase_database_debugger.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nd_test_support/nd_test_support.dart';

void main() {
  group('e2e (firebase_database_debugger) -> Go ingest', () {
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      proxy = await GoNetworkDebuggerProcess.start();
      await clearSessions(proxy!.apiBase);
    });

    tearDown(() async {
      await proxy?.stop();
    });

    test('sends firebase ingest and session appears in API', () async {
      final debugger = FirebaseDatabaseDebugger(
        config: FirebaseDatabaseDebuggerConfig(
          debuggerBaseUrl: proxy!.apiBase.toString(),
          databaseUrl: 'https://demo-default-rtdb.firebaseio.com',
          enabled: true,
          maxBatchFrames: 10,
        ),
      );
      try {
        await debugger.logOperation(
          path: '/users/123',
          op: 'set',
          direction: 'client->upstream',
          payload: <String, dynamic>{
            'value': <String, dynamic>{'name': 'Bob'}
          },
        );
        await debugger.flush();

        final sessions = await listSessions(proxy!.apiBase, types: 'firebase');
        expect(
          sessions.any((item) {
            final kind = (item['kind'] ?? '').toString();
            final target = (item['target'] ?? '').toString();
            return kind == 'firebase_database' &&
                target.contains('demo-default-rtdb.firebaseio.com');
          }),
          isTrue,
        );
      } finally {
        await debugger.dispose();
      }
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
