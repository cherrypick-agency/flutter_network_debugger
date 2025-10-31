import 'package:flutter/foundation.dart';

@immutable
class InterceptRule {
  const InterceptRule({
    required this.id,
    required this.enabled,
    required this.priority,
    required this.action, // request|response|both
    required this.once,
    required this.stopProcessing,
    required this.when,
  });
  final String id;
  final bool enabled;
  final int priority;
  final String action;
  final bool once;
  final bool stopProcessing;
  final InterceptWhen when;
}

@immutable
class InterceptWhen {
  const InterceptWhen({
    this.method = const [],
    this.scheme = const [],
    this.host,
    this.port,
    this.path,
    this.contentType,
    this.responseStatus,
    this.header,
    this.bodyContains,
  });
  final List<String> method;
  final List<String> scheme;
  final RuleStringMatch? host;
  final RuleStringMatch? port;
  final RuleStringMatch? path;
  final RuleStringMatch? contentType;
  final RuleStatusMatch? responseStatus;
  final RuleHeaderMatch? header;
  final String? bodyContains;
}

@immutable
class RuleStringMatch {
  const RuleStringMatch({
    this.equals,
    this.prefix,
    this.suffix,
    this.contains,
    this.anyOf = const [],
    this.regex,
  });
  final String? equals;
  final String? prefix;
  final String? suffix;
  final String? contains;
  final List<String> anyOf;
  final String? regex;
}

@immutable
class RuleHeaderMatch {
  const RuleHeaderMatch({required this.name, this.value});
  final RuleStringMatch name;
  final RuleStringMatch? value;
}

@immutable
class RuleStatusMatch {
  const RuleStatusMatch({
    this.equals = const [],
    this.from,
    this.to,
    this.is4xx = false,
    this.is5xx = false,
  });
  final List<int> equals;
  final int? from;
  final int? to;
  final bool is4xx;
  final bool is5xx;
}
