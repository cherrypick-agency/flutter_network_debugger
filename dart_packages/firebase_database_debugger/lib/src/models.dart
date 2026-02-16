/// Запрос на отправку фреймов в ingest API дебаггера.
///
/// Содержит информацию о сессии, список фреймов и флаг закрытия.
class FirebaseIngestRequest {
  FirebaseIngestRequest({
    required this.session,
    required this.frames,
    this.close = false,
    this.error,
  });

  /// Сессия, к которой относятся фреймы.
  final FirebaseIngestSession session;

  /// Список фреймов (операций), накопленных за текущий батч.
  final List<FirebaseIngestFrame> frames;

  /// Если `true`, сессия будет закрыта на стороне дебаггера.
  final bool close;

  /// Текст ошибки при закрытии сессии (если есть).
  final String? error;

  /// Сериализация в JSON для отправки по HTTP.
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'session': session.toJson(),
      'frames': frames.map((item) => item.toJson()).toList(growable: false),
      'close': close,
      'error': error,
    };
  }
}

/// Метаданные сессии отладки Firebase Realtime Database.
///
/// Сессия группирует фреймы по целевому пути (target) в базе данных.
class FirebaseIngestSession {
  FirebaseIngestSession({
    required this.id,
    required this.target,
    this.startedAt,
    this.captureId,
  });

  /// Уникальный идентификатор сессии, генерируется на основе target + run-id.
  final String id;

  /// Полный URL пути в Firebase, например `https://my-db.firebaseio.com/users`.
  final String target;

  /// Время начала сессии (UTC). Может быть `null`, если не задано явно.
  final DateTime? startedAt;

  /// Идентификатор захвата (capture) в дебаггере.
  final String? captureId;

  /// Сериализация в JSON для отправки по HTTP.
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'id': id,
      'target': target,
      if (startedAt != null) 'startedAt': startedAt!.toUtc().toIso8601String(),
      if ((captureId ?? '').trim().isNotEmpty) 'captureId': captureId,
    };
  }
}

/// Один фрейм (запись об операции) в сессии отладки.
///
/// Содержит направление, тип операции, превью данных и опциональное тело.
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

  /// Уникальный идентификатор фрейма, например `fr-set-1700000000-1`.
  final String id;

  /// Временная метка операции (UTC).
  final DateTime ts;

  /// Направление: `client->upstream` или `upstream->client`.
  final String direction;

  /// Тип фрейма, обычно `text`.
  final String opcode;

  /// MIME-тип содержимого, по умолчанию `application/json`.
  final String contentType;

  /// Краткое превью данных операции в формате JSON.
  final String preview;

  /// Полное тело фрейма (base64), если превью превысило лимит.
  final String? body;

  /// Кодировка [body], например `base64`. `null` если тело не задано.
  final String? bodyEncoding;

  /// Сериализация в JSON для отправки по HTTP.
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
