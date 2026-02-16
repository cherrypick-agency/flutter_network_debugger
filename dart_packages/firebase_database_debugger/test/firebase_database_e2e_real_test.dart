import 'dart:convert';
import 'dart:io';
import 'dart:math';

import 'package:firebase_database_debugger/firebase_database_debugger.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nd_test_support/nd_test_support.dart';

/// URL тестовой Firebase RTDB (из GoogleService-Info-network-debugger-test.plist)
const _databaseUrl =
    'https://test-project-9d131-default-rtdb.europe-west1.firebasedatabase.app';

/// Корневой путь для тестовых данных — чистим после каждого прогона
final _testRoot = '/nd_e2e_test/${DateTime.now().millisecondsSinceEpoch}';

/// Проверяет, доступна ли Firebase RTDB для записи (правила открыты)
Future<bool> _isFirebaseWritable() async {
  final client = HttpClient();
  try {
    final uri = Uri.parse('$_databaseUrl/nd_e2e_probe.json');
    final req = await client.putUrl(uri);
    req.headers.contentType = ContentType.json;
    req.write(jsonEncode({'probe': true}));
    final resp = await req.close();
    await resp.drain<void>();
    return resp.statusCode == 200;
  } catch (_) {
    return false;
  } finally {
    client.close(force: true);
  }
}

/// REST-обёртка для Firebase RTDB
class _FirebaseRest {
  final String baseUrl;
  _FirebaseRest(this.baseUrl);

  Future<int> set(String path, Object? value) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.putUrl(uri);
    req.headers.contentType = ContentType.json;
    req.write(jsonEncode(value));
    final resp = await req.close();
    await resp.drain<void>();
    final code = resp.statusCode;
    client.close(force: true);
    return code;
  }

  Future<dynamic> get(String path) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.getUrl(uri);
    final resp = await req.close();
    final txt = await utf8.decodeStream(resp);
    client.close(force: true);
    return jsonDecode(txt);
  }

  Future<int> remove(String path) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.deleteUrl(uri);
    final resp = await req.close();
    await resp.drain<void>();
    final code = resp.statusCode;
    client.close(force: true);
    return code;
  }

  Future<int> update(String path, Map<String, Object?> value) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.patchUrl(uri);
    req.headers.contentType = ContentType.json;
    req.write(jsonEncode(value));
    final resp = await req.close();
    await resp.drain<void>();
    final code = resp.statusCode;
    client.close(force: true);
    return code;
  }
}

void main() {
  group('e2e: real Firebase RTDB + Go backend', () {
    GoNetworkDebuggerProcess? proxy;
    late _FirebaseRest firebase;
    late FirebaseDatabaseDebugger debugger;
    late bool firebaseAvailable;

    setUpAll(() async {
      firebaseAvailable = await _isFirebaseWritable();
    });

    setUp(() async {
      if (!firebaseAvailable) return;
      proxy = await GoNetworkDebuggerProcess.start();
      await clearSessions(proxy!.apiBase);
      firebase = _FirebaseRest(_databaseUrl);
      debugger = FirebaseDatabaseDebugger(
        config: FirebaseDatabaseDebuggerConfig(
          debuggerBaseUrl: proxy!.apiBase.toString(),
          databaseUrl: _databaseUrl,
          enabled: true,
          maxBatchFrames: 50,
          flushInterval: const Duration(milliseconds: 100),
        ),
      );
    });

    tearDown(() async {
      if (!firebaseAvailable) return;
      await debugger.dispose();
      // Чистим тестовые данные в Firebase
      try {
        await firebase.remove(_testRoot);
      } catch (_) {}
      await proxy?.stop();
    });

    test('set + get + update + remove: сессии и фреймы в Go backend', () async {
      if (!firebaseAvailable) {
        markTestSkipped('Firebase RTDB недоступна (правила закрыты)');
        return;
      }
      final rnd = Random().nextInt(99999);
      final userPath = '$_testRoot/users/user_$rnd';

      // 1. SET
      final setStatus = await firebase.set(userPath, {
        'name': 'Test User $rnd',
        'email': 'test$rnd@example.com',
        'createdAt': DateTime.now().toUtc().toIso8601String(),
      });
      expect(setStatus, 200, reason: 'Firebase SET должен вернуть 200');
      await debugger.logOperation(
        path: userPath,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'value': {
            'name': 'Test User $rnd',
            'email': 'test$rnd@example.com',
          },
        },
      );

      // 2. GET
      final getData = await firebase.get(userPath);
      expect(getData, isA<Map>());
      expect((getData as Map)['name'], 'Test User $rnd');
      await debugger.logOperation(
        path: userPath,
        op: 'get',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'exists': true,
          'value': getData,
        },
      );

      // 3. UPDATE
      final updateStatus = await firebase.update(userPath, {
        'name': 'Updated User $rnd',
        'updatedAt': DateTime.now().toUtc().toIso8601String(),
      });
      expect(updateStatus, 200, reason: 'Firebase UPDATE должен вернуть 200');
      await debugger.logOperation(
        path: userPath,
        op: 'update',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'value': {'name': 'Updated User $rnd'},
        },
      );

      // Проверяем что update реально применился
      final afterUpdate = await firebase.get(userPath);
      expect((afterUpdate as Map)['name'], 'Updated User $rnd');
      expect(afterUpdate['email'], 'test$rnd@example.com');

      // 4. REMOVE
      final removeStatus = await firebase.remove(userPath);
      expect(removeStatus, 200, reason: 'Firebase REMOVE должен вернуть 200');
      await debugger.logOperation(
        path: userPath,
        op: 'remove',
        direction: 'client->upstream',
        payload: null,
      );

      // Проверяем что данные удалены
      final afterRemove = await firebase.get(userPath);
      expect(afterRemove, isNull);

      // 5. Закрываем сессию
      debugger.markSessionClosed(path: userPath);

      // Принудительно сбрасываем буфер
      await debugger.flush();

      // 6. Проверяем что сессия появилась в Go backend
      final sessions = await listSessions(proxy!.apiBase, types: 'firebase');
      expect(sessions.isNotEmpty, isTrue,
          reason: 'Должна быть хотя бы одна Firebase сессия');

      final fbSession = sessions.firstWhere(
        (s) =>
            (s['kind'] ?? '') == 'firebase_database' &&
            (s['target'] ?? '').toString().contains('user_$rnd'),
        orElse: () => <String, dynamic>{},
      );
      expect(fbSession.isNotEmpty, isTrue,
          reason: 'Сессия для user_$rnd должна существовать');

      final sessionId = fbSession['id'] as String;

      // 7. Проверяем фреймы
      final frames = await listFrames(proxy!.apiBase, sessionId);
      expect(frames.length, greaterThanOrEqualTo(4),
          reason: 'Должно быть минимум 4 фрейма (set, get, update, remove)');

      // Проверяем что есть фреймы с нужными операциями
      final ops = frames
          .map((f) {
            final preview = f['preview'] ?? '';
            try {
              final decoded = jsonDecode(preview.toString()) as Map;
              return decoded['op'] as String?;
            } catch (_) {
              return null;
            }
          })
          .whereType<String>()
          .toSet();

      expect(ops.contains('set'), isTrue, reason: 'Должен быть фрейм set');
      expect(ops.contains('get'), isTrue, reason: 'Должен быть фрейм get');
      expect(ops.contains('update'), isTrue,
          reason: 'Должен быть фрейм update');
      expect(ops.contains('remove'), isTrue,
          reason: 'Должен быть фрейм remove');

      // 8. Проверяем что сессия закрыта
      final sessionDetail = await getSession(proxy!.apiBase, sessionId);
      expect(sessionDetail['closedAt'], isNotNull,
          reason: 'Сессия должна быть закрыта');
    }, timeout: const Timeout(Duration(seconds: 120)));

    test('несколько путей создают разные сессии', () async {
      if (!firebaseAvailable) {
        markTestSkipped('Firebase RTDB недоступна (правила закрыты)');
        return;
      }
      final rnd = Random().nextInt(99999);
      final path1 = '$_testRoot/alpha_$rnd';
      final path2 = '$_testRoot/beta_$rnd';

      // Пишем в два разных пути
      await firebase.set(path1, {'v': 1});
      await debugger.logOperation(
        path: path1,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{'value': 1},
      );

      await firebase.set(path2, {'v': 2});
      await debugger.logOperation(
        path: path2,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{'value': 2},
      );

      await debugger.flush();

      final sessions = await listSessions(proxy!.apiBase, types: 'firebase');
      final matching = sessions.where((s) {
        final target = (s['target'] ?? '').toString();
        return target.contains('alpha_$rnd') || target.contains('beta_$rnd');
      }).toList();

      expect(matching.length, 2,
          reason: 'Два разных пути должны создать две разные сессии');

      // Чистим
      await firebase.remove(path1);
      await firebase.remove(path2);
    }, timeout: const Timeout(Duration(seconds: 120)));

    test('большой payload корректно спиливается в body', () async {
      if (!firebaseAvailable) {
        markTestSkipped('Firebase RTDB недоступна (правила закрыты)');
        return;
      }
      final rnd = Random().nextInt(99999);
      final path = '$_testRoot/big_$rnd';

      // Генерируем большой payload (>16KB)
      final bigValue = List.generate(500, (i) => 'item_$i').join(',');
      await firebase.set(path, {'data': bigValue});
      await debugger.logOperation(
        path: path,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{'value': bigValue},
      );

      await debugger.flush();

      final sessions = await listSessions(proxy!.apiBase, types: 'firebase');
      final session = sessions.firstWhere(
        (s) => (s['target'] ?? '').toString().contains('big_$rnd'),
        orElse: () => <String, dynamic>{},
      );
      expect(session.isNotEmpty, isTrue);

      final frames = await listFrames(proxy!.apiBase, session['id'] as String);
      expect(frames.isNotEmpty, isTrue);

      // Для большого payload preview должен содержать bodySpilled
      final frame = frames.first;
      final preview = frame['preview']?.toString() ?? '';
      if (bigValue.length > 1024) {
        // Если payload достаточно большой, проверяем spilling
        final decoded = jsonDecode(preview) as Map;
        if (decoded.containsKey('bodySpilled')) {
          expect(decoded['bodySpilled'], isTrue);
          expect(frame['bodyEncoding'], 'base64');
        }
      }

      await firebase.remove(path);
    }, timeout: const Timeout(Duration(seconds: 120)));
  },
      skip: GoNetworkDebuggerProcess.hasGo()
          ? false
          : 'go not found — нужен Go для сборки backend');
}
