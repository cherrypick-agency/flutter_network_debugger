import '../domain/entities/decisions.dart';
import '../domain/entities/intercept_config.dart';
import '../domain/entities/intercept_item.dart';
import '../domain/entities/intercept_rule.dart';
import '../domain/repositories/breakpoints_repository.dart';
import 'breakpoints_api.dart';

class BreakpointsRepositoryImpl implements BreakpointsRepository {
  BreakpointsRepositoryImpl(this._api);
  final BreakpointsApi _api;

  String? _nonEmptyString(dynamic v) {
    if (v == null) return null;
    final s = v.toString();
    return s.trim().isEmpty ? null : s;
  }

  int _asInt(dynamic v, {required int fallback}) {
    if (v is int) return v;
    if (v is num) return v.toInt();
    if (v is String) return int.tryParse(v) ?? fallback;
    return fallback;
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
    return InterceptConfig(
      enabled: (m['enabled'] ?? false) as bool,
      requests: (m['requests'] ?? true) as bool,
      responses: (m['responses'] ?? true) as bool,
      timeoutMs: _asInt(m['timeoutMs'], fallback: 60000),
      queueMax: _asInt(m['queueMax'], fallback: 200),
      bodyMaxBytes: _asInt(m['bodyMaxBytes'], fallback: 1048576),
      reencode: (m['reencode'] ?? true) as bool,
      overflow: (m['overflow'] ?? 'auto-continue-oldest') as String,
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
    return list.map((e) => _mapRule(e as Map<String, dynamic>)).toList();
  }

  @override
  Future<void> replaceRules(List<InterceptRule> rules) async {
    await _api.replaceRules(rules.map(_ruleToJson).toList());
  }

  @override
  Future<List<InterceptItem>> listPending({int? limit}) async {
    final list = await _api.listPending(limit: limit);
    return list.map((e) => _mapItem(e as Map<String, dynamic>)).toList();
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
      id: (m['id'] ?? '') as String,
      enabled: (m['enabled'] ?? true) as bool,
      priority: _asInt(m['priority'], fallback: 0),
      action: (m['action'] ?? 'both') as String,
      once: (m['once'] ?? false) as bool,
      stopProcessing: (m['stopProcessing'] ?? true) as bool,
      when: _mapWhen(
        (m['when'] ?? const <String, dynamic>{}) as Map<String, dynamic>,
      ),
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
    RuleStringMatch? _r(Map<String, dynamic>? x) => x == null
        ? null
        : () {
            final equals = _nonEmptyString(x['equals']);
            final prefix = _nonEmptyString(x['prefix']);
            final suffix = _nonEmptyString(x['suffix']);
            final contains = _nonEmptyString(x['contains']);
            final regex = _nonEmptyString(x['regex']);
            final anyOf =
                (x['anyOf'] as List<dynamic>?)
                    ?.map((e) => _nonEmptyString(e))
                    .whereType<String>()
                    .toList(growable: false) ??
                const <String>[];
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
            final name = _r((x['name'] as Map<String, dynamic>?));
            final value = _r(x['value'] as Map<String, dynamic>?);
            if (name == null) return null;
            // Если name пустой (или невалидный) — не создаём фантомный header match.
            return RuleHeaderMatch(name: name, value: value);
          }();
    RuleStatusMatch? _s(Map<String, dynamic>? x) => x == null
        ? null
        : () {
            final equals = _asIntList(x['equals']);
            final from = (x['from'] is num)
                ? (x['from'] as num).toInt()
                : (x['from'] as int?);
            final to = (x['to'] is num)
                ? (x['to'] as num).toInt()
                : (x['to'] as int?);
            final is4xx = (x['is4xx'] ?? false) as bool;
            final is5xx = (x['is5xx'] ?? false) as bool;
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
      method: (m['method'] as List<dynamic>?)?.cast<String>() ?? const [],
      scheme: (m['scheme'] as List<dynamic>?)?.cast<String>() ?? const [],
      host: _r(m['host'] as Map<String, dynamic>?),
      port: _r(m['port'] as Map<String, dynamic>?),
      path: _r(m['path'] as Map<String, dynamic>?),
      contentType: _r(m['contentType'] as Map<String, dynamic>?),
      responseStatus: _s(m['responseStatus'] as Map<String, dynamic>?),
      header: _h(m['header'] as Map<String, dynamic>?),
      bodyContains: (m['bodyContains'] as String?)?.trim().isEmpty == true
          ? null
          : (m['bodyContains'] as String?),
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
    HTTPRequestSnapshot? req;
    final rq = m['req'] as Map<String, dynamic>?;
    if (rq != null) {
      req = HTTPRequestSnapshot(
        method: (rq['method'] ?? '') as String,
        url: (rq['url'] ?? '') as String,
        headers: _headersMap(rq['headers']),
        bodyBase64: rq['bodyBase64'] as String?,
        bodyTruncated: (rq['bodyTruncated'] ?? false) as bool,
        contentType: rq['contentType'] as String?,
      );
    }
    HTTPResponseSnapshot? res;
    final rs = m['res'] as Map<String, dynamic>?;
    if (rs != null) {
      res = HTTPResponseSnapshot(
        status: _asInt(rs['status'], fallback: 0),
        headers: _headersMap(rs['headers']),
        bodyBase64: rs['bodyBase64'] as String?,
        bodyTruncated: (rs['bodyTruncated'] ?? false) as bool,
        contentType: rs['contentType'] as String?,
      );
    }
    return InterceptItem(
      id: (m['id'] ?? '') as String,
      createdAt:
          DateTime.tryParse((m['createdAt'] ?? '') as String) ??
          DateTime.now().toUtc(),
      deadline:
          DateTime.tryParse((m['deadline'] ?? '') as String) ??
          DateTime.now().toUtc(),
      direction: (m['direction'] ?? 'request') as String,
      sessionId: (m['sessionId'] ?? '') as String,
      state: (m['state'] ?? 'PENDING') as String,
      ruleId: m['ruleId'] as String?,
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
