import '../domain/entities/decisions.dart';
import '../domain/entities/intercept_config.dart';
import '../domain/entities/intercept_item.dart';
import '../domain/entities/intercept_rule.dart';
import '../domain/repositories/breakpoints_repository.dart';
import 'breakpoints_api.dart';

class BreakpointsRepositoryImpl implements BreakpointsRepository {
  BreakpointsRepositoryImpl(this._api);
  final BreakpointsApi _api;

  @override
  Future<InterceptConfig> getConfig() async {
    final m = await _api.getConfig();
    return InterceptConfig(
      enabled: (m['enabled'] ?? false) as bool,
      requests: (m['requests'] ?? true) as bool,
      responses: (m['responses'] ?? true) as bool,
      timeoutMs: (m['timeoutMs'] ?? 60000) as int,
      queueMax: (m['queueMax'] ?? 200) as int,
      bodyMaxBytes: (m['bodyMaxBytes'] ?? 1048576) as int,
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
      priority: (m['priority'] ?? 0) as int,
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
    RuleStringMatch? _r(Map<String, dynamic>? x) =>
        x == null
            ? null
            : RuleStringMatch(
              equals: x['equals'] as String?,
              prefix: x['prefix'] as String?,
              suffix: x['suffix'] as String?,
              contains: x['contains'] as String?,
              anyOf: (x['anyOf'] as List<dynamic>?)?.cast<String>() ?? const [],
              regex: x['regex'] as String?,
            );
    RuleHeaderMatch? _h(Map<String, dynamic>? x) =>
        x == null
            ? null
            : RuleHeaderMatch(
              name: _r((x['name'] as Map<String, dynamic>?))!,
              value: _r(x['value'] as Map<String, dynamic>?),
            );
    RuleStatusMatch? _s(Map<String, dynamic>? x) =>
        x == null
            ? null
            : RuleStatusMatch(
              equals: (x['equals'] as List<dynamic>?)?.cast<int>() ?? const [],
              from: x['from'] as int?,
              to: x['to'] as int?,
              is4xx: (x['is4xx'] ?? false) as bool,
              is5xx: (x['is5xx'] ?? false) as bool,
            );
    return InterceptWhen(
      method: (m['method'] as List<dynamic>?)?.cast<String>() ?? const [],
      scheme: (m['scheme'] as List<dynamic>?)?.cast<String>() ?? const [],
      host: _r(m['host'] as Map<String, dynamic>?),
      port: _r(m['port'] as Map<String, dynamic>?),
      path: _r(m['path'] as Map<String, dynamic>?),
      contentType: _r(m['contentType'] as Map<String, dynamic>?),
      responseStatus: _s(m['responseStatus'] as Map<String, dynamic>?),
      header: _h(m['header'] as Map<String, dynamic>?),
      bodyContains: m['bodyContains'] as String?,
    );
  }

  Map<String, dynamic> _whenToJson(InterceptWhen w) => {
    if (w.method.isNotEmpty) 'method': w.method,
    if (w.scheme.isNotEmpty) 'scheme': w.scheme,
    if (w.host != null) 'host': _rToJson(w.host!),
    if (w.port != null) 'port': _rToJson(w.port!),
    if (w.path != null) 'path': _rToJson(w.path!),
    if (w.contentType != null) 'contentType': _rToJson(w.contentType!),
    if (w.responseStatus != null) 'responseStatus': _sToJson(w.responseStatus!),
    if (w.header != null) 'header': _hToJson(w.header!),
    if (w.bodyContains != null) 'bodyContains': w.bodyContains,
  };

  Map<String, dynamic> _rToJson(RuleStringMatch r) => {
    if (r.equals != null) 'equals': r.equals,
    if (r.prefix != null) 'prefix': r.prefix,
    if (r.suffix != null) 'suffix': r.suffix,
    if (r.contains != null) 'contains': r.contains,
    if (r.anyOf.isNotEmpty) 'anyOf': r.anyOf,
    if (r.regex != null) 'regex': r.regex,
  };

  Map<String, dynamic> _hToJson(RuleHeaderMatch h) => {
    'name': _rToJson(h.name),
    if (h.value != null) 'value': _rToJson(h.value!),
  };

  Map<String, dynamic> _sToJson(RuleStatusMatch s) => {
    if (s.equals.isNotEmpty) 'equals': s.equals,
    if (s.from != null) 'from': s.from,
    if (s.to != null) 'to': s.to,
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
        status: (rs['status'] ?? 0) as int,
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
    if (v is Map<String, dynamic>) {
      return v.map(
        (key, value) => MapEntry(key, (value as List<dynamic>).cast<String>()),
      );
    }
    return <String, List<String>>{};
  }
}
