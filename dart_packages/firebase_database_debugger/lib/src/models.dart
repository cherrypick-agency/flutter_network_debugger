class FirebaseIngestRequest {
  FirebaseIngestRequest({
    required this.session,
    required this.frames,
    this.close = false,
    this.error,
  });

  final FirebaseIngestSession session;
  final List<FirebaseIngestFrame> frames;
  final bool close;
  final String? error;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'session': session.toJson(),
      'frames': frames.map((item) => item.toJson()).toList(growable: false),
      'close': close,
      'error': error,
    };
  }
}

class FirebaseIngestSession {
  FirebaseIngestSession({
    required this.id,
    required this.target,
    this.startedAt,
    this.captureId,
  });

  final String id;
  final String target;
  final DateTime? startedAt;
  final String? captureId;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'id': id,
      'target': target,
      if (startedAt != null) 'startedAt': startedAt!.toUtc().toIso8601String(),
      if ((captureId ?? '').trim().isNotEmpty) 'captureId': captureId,
    };
  }
}

class FirebaseIngestFrame {
  FirebaseIngestFrame({
    required this.id,
    required this.ts,
    required this.direction,
    required this.opcode,
    required this.preview,
    this.contentType = 'application/json',
    this.body,
    this.bodyEncoding,
  });

  final String id;
  final DateTime ts;
  final String direction;
  final String opcode;
  final String contentType;
  final String preview;
  final String? body;
  final String? bodyEncoding;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'id': id,
      'ts': ts.toUtc().toIso8601String(),
      'direction': direction,
      'opcode': opcode,
      'contentType': contentType,
      'preview': preview,
      if (body != null) 'body': body,
      if (bodyEncoding != null) 'bodyEncoding': bodyEncoding,
    };
  }
}
