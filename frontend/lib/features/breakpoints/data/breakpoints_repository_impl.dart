import '../domain/entities/decisions.dart';
import '../domain/entities/intercept_config.dart';
import '../domain/entities/intercept_item.dart';
import '../domain/entities/intercept_rule.dart';
import '../domain/repositories/breakpoints_repository.dart';
import '../../../core/utils/json_cast.dart';
import 'breakpoints_api.dart';

class BreakpointsRepositoryImpl implements BreakpointsRepository {
  BreakpointsRepositoryImpl(this._api);
  final BreakpointsApi _api;

  bool _asBool(dynamic v, {required bool fallback}) {
    return JsonCast.asBool(v, fallback: fallback);
  }

  String? _nonEmptyString(dynamic v) {
    if (v == null) return null;
    final s = v.toString();
    return s.trim().isEmpty ? null : s;
  }

  int _asInt(dynamic v, {required int fallback}) {
    return JsonCast.asInt(v, fallback: fallback);
  }

  List<int> _asIntList(dynamic v) {
    if (v is! List) return const [];
    final out = <int>[];
    for (final e in v) {
      if (e is int) {
        out.add(e);
        continue;
      }
      if (e is num) {
        out.add(e.toInt());
        continue;
      }
      if (e is String) {
        final n = int.tryParse(e);
        if (n != null) out.add(n);
      }
    }
    return out;
  }

  @override
  Future<InterceptConfig> getConfig() async {
    final m = await _api.getConfig();
    final timeoutMs = _asInt(m['timeoutMs'], fallback: 60000);
    final bodyMaxBytes = _asInt(m['bodyMaxBytes'], fallback: 1048576);
    final queueMax = _asInt(m['queueMax'], fallback: 200);
    return InterceptConfig(
      enabled: _asBool(m['enabled'], fallback: false),
      requests: _asBool(m['requests'], fallback: true),
      responses: _asBool(m['responses'], fallback: true),
      timeoutMs: timeoutMs > 0 ? timeoutMs : 60000,
      queueMax: queueMax >= 0 ? queueMax : 200,
      bodyMaxBytes: bodyMaxBytes > 0 ? bodyMaxBytes : 1048576,
      reencode: _asBool(m['reencode'], fallback: true),
      overflow: (m['overflow'] ?? 'auto-continue-oldest').toString(),
    );
  }

  @override
  Future<void> setConfig(InterceptConfig cfg) async {
    await _api.setConfig({
      'enabled': cfg.enabled,
      'requests': cfg.requests,
      'responses': cfg.responses,
      'timeoutMs': cfg.timeoutMs,
      'queueMax': cfg.queueMax,
      'bodyMaxBytes': cfg.bodyMaxBytes,
      'reencode': cfg.reencode,
      'overflow': cfg.overflow,
    });
  }

  @override
  Future<List<InterceptRule>> listRules() async {
    final list = await _api.listRules();
    return list
        .whereType<Map>()
        .map((e) => _mapRule(JsonCast.asMap(e)))
        .toList();
  }

  @override
  Future<void> replaceRules(List<InterceptRule> rules) async {
    await _api.replaceRules(rules.map(_ruleToJson).toList());
  }

  @override
  Future<List<InterceptItem>> listPending({int? limit}) async {
    final list = await _api.listPending(limit: limit);
    return list
        .whereType<Map>()
        .map((e) => _mapItem(JsonCast.asMap(e)))
        .toList();
  }

  @override
  Future<InterceptItem> getItem(String id) async {
    final m = await _api.getItem(id);
    return _mapItem(m);
  }

  @override
  Future<void> continueRequest(String id, RequestDecision decision) async {
    await _api.continueItem(id, {
      'action': decision.action,
      if (decision.method != null) 'method': decision.method,
      if (decision.url != null) 'url': decision.url,
      if (decision.headers != null) 'headers': decision.headers,
      if (decision.bodyBase64 != null) 'bodyBase64': decision.bodyBase64,
    });
  }

  @override
  Future<void> continueResponse(String id, ResponseDecision decision) async {
    await _api.continueItem(id, {
      'action': decision.action,
      if (decision.status != null) 'status': decision.status,
      if (decision.headers != null) 'headers': decision.headers,
      if (decision.bodyBase64 != null) 'bodyBase64': decision.bodyBase64,
    });
  }

  @override
  Future<void> cancel(String id) async {
    await _api.cancel(id);
  }

  InterceptRule _mapRule(Map<String, dynamic> m) {
    return InterceptRule(
      id: (m['id'] ?? '').toString(),
      enabled: _asBool(m['enabled'], fallback: true),
      priority: _asInt(m['priority'], fallback: 0),
      action: (m['action'] ?? 'both').toString(),
      once: _asBool(m['once'], fallback: false),
      stopProcessing: _asBool(m['stopProcessing'], fallback: true),
      when: _mapWhen(JsonCast.asMap(m['when'])),
    );
  }

  Map<String, dynamic> _ruleToJson(InterceptRule r) => {
    'id': r.id,
    'enabled': r.enabled,
    'priority': r.priority,
    'action': r.action,
    'once': r.once,
    'stopProcessing': r.stopProcessing,
    'when': _whenToJson(r.when),
  };

  InterceptWhen _mapWhen(Map<String, dynamic> m) {
    Map<String, dynamic>? _mapOrNull(dynamic v) {
      final out = JsonCast.asMap(v);
      return out.isEmpty ? null : out;
    }

    RuleStringMatch? _r(Map<String, dynamic>? x) => x == null
        ? null
        : () {
            final equals = _nonEmptyString(x['equals']);
            final prefix = _nonEmptyString(x['prefix']);
            final suffix = _nonEmptyString(x['suffix']);
            final contains = _nonEmptyString(x['contains']);
            final regex = _nonEmptyString(x['regex']);
            final anyOf = JsonCast.asList(
              x['anyOf'],
            ).map(_nonEmptyString).whereType<String>().toList(growable: false);
            final isEmpty =
                equals == null &&
                prefix == null &&
                suffix == null &&
                contains == null &&
                regex == null &&
                anyOf.isEmpty;
            if (isEmpty) return null;
            return RuleStringMatch(
              equals: equals,
              prefix: prefix,
              suffix: suffix,
              contains: contains,
              anyOf: anyOf,
              regex: regex,
            );
          }();
    RuleHeaderMatch? _h(Map<String, dynamic>? x) => x == null
        ? null
        : () {
            final name = _r(_mapOrNull(x['name']));
            final value = _r(_mapOrNull(x['value']));
            if (name == null) return null;
            // If name is empty (or invalid) - don't create phantom header match.
            return RuleHeaderMatch(name: name, value: value);
          }();
    RuleStatusMatch? _s(Map<String, dynamic>? x) => x == null
        ? null
        : () {
            final equals = _asIntList(x['equals']);
            final rawFrom = x['from'];
            final rawTo = x['to'];
            final from = rawFrom == null
                ? null
                : JsonCast.asInt(rawFrom, fallback: 0);
            final to = rawTo == null
                ? null
                : JsonCast.asInt(rawTo, fallback: 0);
            final is4xx = _asBool(x['is4xx'], fallback: false);
            final is5xx = _asBool(x['is5xx'], fallback: false);
            final normalizedFrom = (from == null || from <= 0) ? null : from;
            final normalizedTo = (to == null || to <= 0) ? null : to;
            final isEmpty =
                equals.isEmpty &&
                normalizedFrom == null &&
                normalizedTo == null &&
                !is4xx &&
                !is5xx;
            if (isEmpty) return null;
            return RuleStatusMatch(
              equals: equals,
              from: normalizedFrom,
              to: normalizedTo,
              is4xx: is4xx,
              is5xx: is5xx,
            );
          }();
    return InterceptWhen(
      method: JsonCast.asStringList(m['method']),
      scheme: JsonCast.asStringList(m['scheme']),
      host: _r(_mapOrNull(m['host'])),
      port: _r(_mapOrNull(m['port'])),
      path: _r(_mapOrNull(m['path'])),
      contentType: _r(_mapOrNull(m['contentType'])),
      responseStatus: _s(_mapOrNull(m['responseStatus'])),
      header: _h(_mapOrNull(m['header'])),
      bodyContains: JsonCast.asTrimmedStringOrNull(m['bodyContains']),
    );
  }

  Map<String, dynamic> _whenToJson(InterceptWhen w) {
    final out = <String, dynamic>{};
    if (w.method.isNotEmpty) out['method'] = w.method;
    if (w.scheme.isNotEmpty) out['scheme'] = w.scheme;

    if (w.host != null) {
      final m = _rToJson(w.host!);
      if (m.isNotEmpty) out['host'] = m;
    }
    if (w.port != null) {
      final m = _rToJson(w.port!);
      if (m.isNotEmpty) out['port'] = m;
    }
    if (w.path != null) {
      final m = _rToJson(w.path!);
      if (m.isNotEmpty) out['path'] = m;
    }
    if (w.contentType != null) {
      final m = _rToJson(w.contentType!);
      if (m.isNotEmpty) out['contentType'] = m;
    }
    if (w.responseStatus != null) {
      final m = _sToJson(w.responseStatus!);
      if (m.isNotEmpty) out['responseStatus'] = m;
    }
    if (w.header != null) {
      final m = _hToJson(w.header!);
      final name = (m['name'] as Map?)?.cast<String, dynamic>() ?? const {};
      final value = (m['value'] as Map?)?.cast<String, dynamic>();
      if (name.isNotEmpty || (value != null && value.isNotEmpty)) {
        out['header'] = m;
      }
    }
    if (w.bodyContains != null && w.bodyContains!.trim().isNotEmpty) {
      out['bodyContains'] = w.bodyContains;
    }
    return out;
  }

  Map<String, dynamic> _rToJson(RuleStringMatch r) {
    final out = <String, dynamic>{};
    final equals = _nonEmptyString(r.equals);
    final prefix = _nonEmptyString(r.prefix);
    final suffix = _nonEmptyString(r.suffix);
    final contains = _nonEmptyString(r.contains);
    final regex = _nonEmptyString(r.regex);
    final anyOf = r.anyOf
        .map((e) => _nonEmptyString(e))
        .whereType<String>()
        .toList(growable: false);
    if (equals != null) out['equals'] = equals;
    if (prefix != null) out['prefix'] = prefix;
    if (suffix != null) out['suffix'] = suffix;
    if (contains != null) out['contains'] = contains;
    if (anyOf.isNotEmpty) out['anyOf'] = anyOf;
    if (regex != null) out['regex'] = regex;
    return out;
  }

  Map<String, dynamic> _hToJson(RuleHeaderMatch h) => {
    'name': _rToJson(h.name),
    if (h.value != null) 'value': _rToJson(h.value!),
  };

  Map<String, dynamic> _sToJson(RuleStatusMatch s) => {
    if (s.equals.isNotEmpty) 'equals': s.equals,
    if (s.from != null && s.from! > 0) 'from': s.from,
    if (s.to != null && s.to! > 0) 'to': s.to,
    if (s.is4xx) 'is4xx': true,
    if (s.is5xx) 'is5xx': true,
  };

  InterceptItem _mapItem(Map<String, dynamic> m) {
    String? _stringOrNull(dynamic v) => v is String ? v : null;

    HTTPRequestSnapshot? req;
    final rqAny = m['req'];
    if (rqAny is Map) {
      final rq = JsonCast.asMap(rqAny);
      req = HTTPRequestSnapshot(
        method: JsonCast.asString(rq['method'], fallback: ''),
        url: JsonCast.asString(rq['url'], fallback: ''),
        headers: _headersMap(rq['headers']),
        bodyBase64: _stringOrNull(rq['bodyBase64']),
        bodyTruncated: _asBool(rq['bodyTruncated'], fallback: false),
        contentType: _stringOrNull(rq['contentType']),
      );
    }
    HTTPResponseSnapshot? res;
    final rsAny = m['res'];
    if (rsAny is Map) {
      final rs = JsonCast.asMap(rsAny);
      res = HTTPResponseSnapshot(
        status: _asInt(rs['status'], fallback: 0),
        headers: _headersMap(rs['headers']),
        bodyBase64: _stringOrNull(rs['bodyBase64']),
        bodyTruncated: _asBool(rs['bodyTruncated'], fallback: false),
        contentType: _stringOrNull(rs['contentType']),
      );
    }
    return InterceptItem(
      id: JsonCast.asString(m['id'], fallback: ''),
      createdAt:
          DateTime.tryParse(JsonCast.asString(m['createdAt'], fallback: '')) ??
          DateTime.now().toUtc(),
      deadline:
          DateTime.tryParse(JsonCast.asString(m['deadline'], fallback: '')) ??
          DateTime.now().toUtc(),
      direction: JsonCast.asString(m['direction'], fallback: 'request'),
      sessionId: JsonCast.asString(m['sessionId'], fallback: ''),
      state: JsonCast.asString(m['state'], fallback: 'PENDING'),
      ruleId: JsonCast.asTrimmedStringOrNull(m['ruleId']),
      req: req,
      res: res,
    );
  }

  Map<String, List<String>> _headersMap(dynamic v) {
    if (v is! Map) return <String, List<String>>{};
    final out = <String, List<String>>{};
    for (final entry in v.entries) {
      final key = entry.key?.toString();
      if (key == null || key.isEmpty) continue;
      final value = entry.value;
      if (value == null) {
        out[key] = const <String>[];
        continue;
      }
      if (value is List) {
        out[key] = value.map((e) => e.toString()).toList(growable: false);
        continue;
      }
      if (value is String) {
        out[key] = <String>[value];
        continue;
      }
      out[key] = <String>[value.toString()];
    }
    return out;
  }
}
