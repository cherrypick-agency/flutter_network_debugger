import 'package:equatable/equatable.dart';

class SessionTag extends Equatable {
  const SessionTag({
    required this.id,
    required this.sessionId,
    required this.tagName,
    required this.createdAt,
  });

  final String id;
  final String sessionId;
  final String tagName;
  final DateTime createdAt;

  factory SessionTag.fromJson(Map<String, dynamic> json) {
    return SessionTag(
      id: json['id'] as String,
      sessionId: json['sessionId'] as String,
      tagName: json['tagName'] as String,
      createdAt: DateTime.parse(json['createdAt'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'sessionId': sessionId,
      'tagName': tagName,
      'createdAt': createdAt.toIso8601String(),
    };
  }

  @override
  List<Object?> get props => [id, sessionId, tagName, createdAt];
}
