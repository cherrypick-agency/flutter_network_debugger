import 'package:equatable/equatable.dart';

class SessionAnnotation extends Equatable {
  const SessionAnnotation({
    required this.id,
    required this.sessionId,
    required this.key,
    required this.value,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String sessionId;
  final String key;
  final String value;
  final DateTime createdAt;
  final DateTime updatedAt;

  factory SessionAnnotation.fromJson(Map<String, dynamic> json) {
    return SessionAnnotation(
      id: json['id'] as String,
      sessionId: json['sessionId'] as String,
      key: json['key'] as String,
      value: json['value'] as String,
      createdAt: DateTime.parse(json['createdAt'] as String),
      updatedAt: DateTime.parse(json['updatedAt'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'sessionId': sessionId,
      'key': key,
      'value': value,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
    };
  }

  @override
  List<Object?> get props => [id, sessionId, key, value, createdAt, updatedAt];
}
