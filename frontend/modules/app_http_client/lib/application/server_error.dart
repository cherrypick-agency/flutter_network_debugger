// Server error payload and codes from backend

class ServerErrorPayload {
  final String code;
  final String message;
  final Map<String, dynamic>? details;

  const ServerErrorPayload({
    required this.code,
    required this.message,
    this.details,
  });
}

enum ServerErrorCode {
  notFound,
  missingTarget,
  invalidTarget,
  upstreamError,
  sessionCreateFailed,
  sessionsListFailed,
  sessionGetFailed,
  framesListFailed,
  eventsListFailed,
  httpListFailed,
  streamUnsupported,
  unauthorized,
  badJson,
  validation,
  conflict,
  badParam,
  invalidUrl,
  badBase64,
  invalidStatus,
  unknown,
}

ServerErrorCode parseServerErrorCode(String? code) {
  switch ((code ?? '').toUpperCase()) {
    case 'NOT_FOUND':
      return ServerErrorCode.notFound;
    case 'MISSING_TARGET':
      return ServerErrorCode.missingTarget;
    case 'INVALID_TARGET':
      return ServerErrorCode.invalidTarget;
    case 'UPSTREAM_ERROR':
      return ServerErrorCode.upstreamError;
    case 'SESSION_CREATE_FAILED':
      return ServerErrorCode.sessionCreateFailed;
    case 'SESSIONS_LIST_FAILED':
      return ServerErrorCode.sessionsListFailed;
    case 'SESSION_GET_FAILED':
      return ServerErrorCode.sessionGetFailed;
    case 'FRAMES_LIST_FAILED':
      return ServerErrorCode.framesListFailed;
    case 'EVENTS_LIST_FAILED':
      return ServerErrorCode.eventsListFailed;
    case 'HTTP_LIST_FAILED':
      return ServerErrorCode.httpListFailed;
    case 'STREAM_UNSUPPORTED':
      return ServerErrorCode.streamUnsupported;
    case 'UNAUTHORIZED':
      return ServerErrorCode.unauthorized;
    case 'BAD_JSON':
      return ServerErrorCode.badJson;
    case 'VALIDATION':
      return ServerErrorCode.validation;
    case 'CONFLICT':
      return ServerErrorCode.conflict;
    case 'BAD_PARAM':
      return ServerErrorCode.badParam;
    case 'INVALID_URL':
      return ServerErrorCode.invalidUrl;
    case 'BAD_BASE64':
      return ServerErrorCode.badBase64;
    case 'INVALID_STATUS':
      return ServerErrorCode.invalidStatus;
    default:
      return ServerErrorCode.unknown;
  }
}
