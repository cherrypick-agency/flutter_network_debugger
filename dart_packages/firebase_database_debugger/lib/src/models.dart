/// Request to send frames to the debugger ingest API.
///
/// Contains session info, frame list, and close flag.
class FirebaseIngestRequest {
  FirebaseIngestRequest({
    required this.session,
    required this.frames,
    this.close = false,
    this.error,
  });

  /// Session the frames belong to.
  final FirebaseIngestSession session;

  /// List of frames (operations) accumulated in the current batch.
  final List<FirebaseIngestFrame> frames;

  /// If `true`, the session will be closed on the debugger side.
  final bool close;

  /// Error message when closing the session (if any).
  final String? error;

  /// Serializes to JSON for HTTP transport.
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'session': session.toJson(),
      'frames': frames.map((item) => item.toJson()).toList(growable: false),
      'close': close,
      'error': error,
    };
  }
}

/// Metadata for a Firebase Realtime Database debug session.
///
/// A session groups frames by their target path in the database.
class FirebaseIngestSession {
  FirebaseIngestSession({
    required this.id,
    required this.target,
    this.startedAt,
    this.captureId,
  });

  /// Unique session identifier, generated from target + run-id.
  final String id;

  /// Full Firebase path URL, e.g. `https://my-db.firebaseio.com/users`.
  final String target;

  /// Session start time (UTC). May be `null` if not set explicitly.
  final DateTime? startedAt;

  /// Capture ID in the debugger.
  final String? captureId;

  /// Serializes to JSON for HTTP transport.
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'id': id,
      'target': target,
      if (startedAt != null) 'startedAt': startedAt!.toUtc().toIso8601String(),
      if ((captureId ?? '').trim().isNotEmpty) 'captureId': captureId,
    };
  }
}

/// A single frame (operation record) in a debug session.
///
/// Contains direction, operation type, data preview, and optional body.
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

  /// Unique frame identifier, e.g. `fr-set-1700000000-1`.
  final String id;

  /// Operation timestamp (UTC).
  final DateTime ts;

  /// Direction: `client->upstream` or `upstream->client`.
  final String direction;

  /// Frame type, usually `text`.
  final String opcode;

  /// Content MIME type, defaults to `application/json`.
  final String contentType;

  /// Short JSON preview of the operation data.
  final String preview;

  /// Full frame body (base64) when preview exceeds the limit.
  final String? body;

  /// Encoding for [body], e.g. `base64`. `null` if body is not set.
  final String? bodyEncoding;

  /// Serializes to JSON for HTTP transport.
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
