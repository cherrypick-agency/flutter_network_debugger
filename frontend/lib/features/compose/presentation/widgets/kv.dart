import 'package:flutter/foundation.dart';

@immutable
class KvPair {
  final String key;
  final String value;
  const KvPair(this.key, this.value);

  KvPair copyWith({String? key, String? value}) {
    return KvPair(key ?? this.key, value ?? this.value);
  }
}
