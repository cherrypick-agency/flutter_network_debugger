import 'dart:convert';
import 'dart:io';

Future<void> clearSessions(Uri apiBase) async {
  final client = HttpClient();
  final req = await client.deleteUrl(apiBase.resolve('/_api/v1/sessions'));
  final resp = await req.close();
  await resp.drain<void>();
  client.close(force: true);
  if (resp.statusCode != 204) {
    throw StateError('clearSessions failed: ${resp.statusCode}');
  }
}

Future<List<Map<String, dynamic>>> listSessions(
  Uri apiBase, {
  required String types,
  int limit = 50,
}) async {
  final client = HttpClient();
  final req = await client.getUrl(
    apiBase.resolve('/_api/v1/sessions?limit=$limit&types=$types'),
  );
  final resp = await req.close();
  final txt = await utf8.decodeStream(resp);
  client.close(force: true);

  final decoded = jsonDecode(txt) as Map<String, dynamic>;
  final items = (decoded['items'] as List).cast<Map>();
  return items
      .map((e) => e.map((k, v) => MapEntry(k.toString(), v)))
      .toList(growable: false);
}
