// Визуальная демонстрация Firebase RTDB дебаггера.
// Делает реальные операции в Firebase и отправляет их в локальный backend.
// Открой фронтенд и смотри как сессии появляются.
//
// Запуск: flutter test test/demo_live_test.dart

import 'dart:convert';
import 'dart:io';
import 'dart:math';

import 'package:firebase_database_debugger/firebase_database_debugger.dart';
import 'package:flutter_test/flutter_test.dart';

const _databaseUrl =
    'https://test-project-9d131-default-rtdb.europe-west1.firebasedatabase.app';

const _debuggerBaseUrl = 'http://127.0.0.1:9092';

class _FirebaseRest {
  final String baseUrl;
  _FirebaseRest(this.baseUrl);

  Future<(int, dynamic)> set(String path, Object? value) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.putUrl(uri);
    req.headers.contentType = ContentType.json;
    req.write(jsonEncode(value));
    final resp = await req.close();
    final body = await utf8.decodeStream(resp);
    client.close(force: true);
    return (resp.statusCode, jsonDecode(body));
  }

  Future<(int, dynamic)> get(String path) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.getUrl(uri);
    final resp = await req.close();
    final body = await utf8.decodeStream(resp);
    client.close(force: true);
    return (resp.statusCode, jsonDecode(body));
  }

  Future<(int, dynamic)> update(String path, Map<String, Object?> value) async {
    final client = HttpClient();
    final uri = Uri.parse('$baseUrl$path.json');
    final req = await client.patchUrl(uri);
    req.headers.contentType = ContentType.json;
    req.write(jsonEncode(value));
    final resp = await req.close();
    final body = await utf8.decodeStream(resp);
    client.close(force: true);
    return (resp.statusCode, jsonDecode(body));
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
}

Future<void> _pause(int seconds) async {
  await Future<void>.delayed(Duration(seconds: seconds));
}

void main() {
  test('demo: Firebase RTDB -> network-debugger (визуальная проверка)',
      () async {
    final rnd = Random().nextInt(99999);
    final firebase = _FirebaseRest(_databaseUrl);

    // Проверяем Firebase
    final (probeStatus, _) = await firebase
        .set('/nd_e2e_probe', {'ts': DateTime.now().toIso8601String()});
    if (probeStatus != 200) {
      markTestSkipped('Firebase RTDB недоступна (status=$probeStatus)');
      return;
    }

    // Проверяем бэкенд
    bool backendOk = false;
    try {
      final hc = HttpClient();
      final hReq = await hc.getUrl(Uri.parse('$_debuggerBaseUrl/healthz'));
      final hResp = await hReq.close();
      await hResp.drain<void>();
      hc.close(force: true);
      backendOk = hResp.statusCode == 200;
    } catch (_) {}
    if (!backendOk) {
      markTestSkipped('Backend недоступен на $_debuggerBaseUrl');
      return;
    }

    final debugger = FirebaseDatabaseDebugger(
      config: FirebaseDatabaseDebuggerConfig(
        debuggerBaseUrl: _debuggerBaseUrl,
        databaseUrl: _databaseUrl,
        enabled: true,
        // Группируем сессии по префиксу пути:
        // /nd_e2e_test/todos/... и /nd_e2e_test/todos/.../task_1 попадут в одну сессию.
        sessionPathDepth: 2,
        maxBatchFrames: 3,
        flushInterval: const Duration(milliseconds: 400),
      ),
    );

    // ─── Сценарий 1: CRUD пользователя ───
    final userPath = '/nd_e2e_test/users/user_$rnd';

    await _pause(2);
    final userData = {
      'name': 'Иван Петров',
      'email': 'ivan$rnd@example.com',
      'age': 28,
      'city': 'Москва',
      'createdAt': DateTime.now().toUtc().toIso8601String(),
    };
    final (setStatus, _) = await firebase.set(userPath, userData);
    expect(setStatus, 200);
    await debugger.logOperation(
      path: userPath,
      op: 'set',
      direction: 'client->upstream',
      payload: <String, dynamic>{'value': userData},
    );

    await _pause(2);
    final (_, getData) = await firebase.get(userPath);
    await debugger.logOperation(
      path: userPath,
      op: 'get',
      direction: 'client->upstream',
      payload: <String, dynamic>{'exists': true, 'value': getData},
    );

    await _pause(2);
    await firebase.update(userPath, {
      'city': 'Санкт-Петербург',
      'updatedAt': DateTime.now().toUtc().toIso8601String(),
    });
    await debugger.logOperation(
      path: userPath,
      op: 'update',
      direction: 'client->upstream',
      payload: <String, dynamic>{
        'value': {'city': 'Санкт-Петербург'}
      },
    );

    await _pause(1);
    final (_, afterUpd) = await firebase.get(userPath);
    await debugger.logOperation(
      path: userPath,
      op: 'get',
      direction: 'client->upstream',
      payload: <String, dynamic>{'exists': true, 'value': afterUpd},
    );

    // ─── Сценарий 2: Список задач (другая сессия) ───
    final todosPath = '/nd_e2e_test/todos/list_$rnd';

    await _pause(2);
    final todosData = {
      'task_1': {'title': 'Купить молоко', 'done': false, 'priority': 'low'},
      'task_2': {'title': 'Написать тесты', 'done': false, 'priority': 'high'},
      'task_3': {
        'title': 'Деплой на прод',
        'done': false,
        'priority': 'critical'
      },
    };
    await firebase.set(todosPath, todosData);
    await debugger.logOperation(
      path: todosPath,
      op: 'set',
      direction: 'client->upstream',
      payload: <String, dynamic>{'value': todosData},
    );

    await _pause(2);
    await firebase.update('$todosPath/task_1', {'done': true});
    await debugger.logOperation(
      path: '$todosPath/task_1',
      op: 'update',
      direction: 'client->upstream',
      payload: <String, dynamic>{
        'value': {'done': true}
      },
    );

    await _pause(1);
    final (_, todosResult) = await firebase.get(todosPath);
    await debugger.logOperation(
      path: todosPath,
      op: 'get',
      direction: 'client->upstream',
      payload: <String, dynamic>{'exists': true, 'value': todosResult},
    );

    // ─── Сценарий 3: Чат — realtime события ───
    final chatPath = '/nd_e2e_test/chat/room_$rnd';

    await _pause(1);
    await debugger.logOperation(
      path: chatPath,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onValue'},
    );

    final messages = [
      {'from': 'Алиса', 'text': 'Привет! Как дела?'},
      {'from': 'Боб', 'text': 'Нормально, пишу тесты'},
      {'from': 'Алиса', 'text': 'Удачи! Я пошла на обед'},
      {'from': 'Боб', 'text': 'Приятного аппетита!'},
    ];

    for (var i = 0; i < messages.length; i++) {
      await _pause(2);
      final msg = {
        ...messages[i],
        'ts': DateTime.now().toUtc().toIso8601String(),
      };
      await firebase.set('$chatPath/msg_$i', msg);
      await debugger.logOperation(
        path: chatPath,
        op: 'onValue',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'exists': true,
          'value': msg,
          'key': 'msg_$i',
        },
      );
    }

    // ─── Сценарий 4: Ошибка ───
    final errorPath = '/nd_e2e_test/errors/fail_$rnd';
    await _pause(1);
    await debugger.logOperation(
      path: errorPath,
      op: 'set',
      direction: 'client->upstream',
      payload: <String, dynamic>{'value': 'secret_data'},
      ok: false,
      error: 'PERMISSION_DENIED: Permission denied',
    );

    // ─── Удаление (ещё фрейм в сессии user) ───
    await _pause(2);
    await firebase.remove(userPath);
    await debugger.logOperation(
      path: userPath,
      op: 'remove',
      direction: 'client->upstream',
      payload: null,
    );

    // ─── Закрываем сессии ───
    await _pause(2);
    debugger.markSessionClosed(path: userPath);
    debugger.markSessionClosed(path: todosPath);
    debugger.markSessionClosed(path: chatPath);
    debugger.markSessionClosed(path: errorPath, error: 'PERMISSION_DENIED');

    await debugger.flush();
    await debugger.dispose();

    // Чистим Firebase
    await firebase.remove('/nd_e2e_test');
  }, timeout: const Timeout(Duration(minutes: 3)));
}
