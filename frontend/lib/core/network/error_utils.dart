import 'package:app_http_client/application/app_http_exception.dart';
import 'package:app_http_client/application/server_error.dart';

class ResolvedErrorMessage {
  final String title;
  final String description;
  final ServerErrorCode? code;
  final Map<String, dynamic>? details;
  final String? raw;
  final String? stack;

  const ResolvedErrorMessage({
    required this.title,
    required this.description,
    this.code,
    this.details,
    this.raw,
    this.stack,
  });
}

ResolvedErrorMessage resolveErrorMessage(Object e, [StackTrace? stackTrace]) {
  // Уже резолвлено ранее — пробрасываем как есть
  if (e is ResolvedErrorMessage) {
    return e;
  }
  // Простой текст
  if (e is String) {
    return ResolvedErrorMessage(
      title: 'Error',
      description: e,
      code: ServerErrorCode.unknown,
      details: {'stack': stackTrace?.toString()},
    );
  }
  if (e is AppHttpServerException) {
    final c = e.code;
    final msg = e.messageFromServer.isNotEmpty ? e.messageFromServer : _defaultMessageForCode(c);
    final req = e.requestOptions;
    final resp = e.response;
    final mergedDetails = <String, dynamic>{
      if (e.serverError.details != null) ...e.serverError.details!,
      if (resp?.statusCode != null) 'statusCode': resp!.statusCode,
      'method': req.method,
      'url': req.uri.toString(),
    };
    return ResolvedErrorMessage(
      title: _titleForCode(c),
      description: msg,
      code: c,
      details: mergedDetails,
      raw: e.toString(),
      stack: stackTrace?.toString(),
    );
  }
  if (e is AppHttp401Exception) {
    return const ResolvedErrorMessage(
      title: 'Unauthorized',
      description: 'Authentication required or token expired.',
      code: ServerErrorCode.unknown,
    );
  }
  if (e is AppHttpException) {
    final req = e.requestOptions;
    final resp = e.response;
    return ResolvedErrorMessage(
      title: 'Network error',
      description: e.message.isNotEmpty ? e.message : 'Request failed.',
      code: ServerErrorCode.unknown,
      details: {
        'method': req.method,
        'url': req.uri.toString(),
        if (resp?.statusCode != null) 'statusCode': resp!.statusCode,
      },
      raw: e.toString(),
      stack: stackTrace?.toString(),
    );
  }
  return ResolvedErrorMessage(
    title: 'Error',
    description: 'Unexpected error. ${e.toString()}',
    details: {
      'stack': stackTrace?.toString(),
      'message': e.toString(),
      'runtimeType': e.runtimeType.toString(),
    },
    code: ServerErrorCode.unknown,
  );
}

String _titleForCode(ServerErrorCode c) {
  switch (c) {
    case ServerErrorCode.notFound:
      return 'Not found';
    case ServerErrorCode.missingTarget:
      return 'Missing target';
    case ServerErrorCode.invalidTarget:
      return 'Invalid target';
    case ServerErrorCode.upstreamError:
      return 'Upstream error';
    case ServerErrorCode.sessionCreateFailed:
      return 'Session error';
    case ServerErrorCode.sessionsListFailed:
    case ServerErrorCode.sessionGetFailed:
    case ServerErrorCode.framesListFailed:
    case ServerErrorCode.eventsListFailed:
    case ServerErrorCode.httpListFailed:
      return 'Data error';
    case ServerErrorCode.streamUnsupported:
      return 'Streaming unsupported';
    case ServerErrorCode.unknown:
      return 'Error';
  }
}

String _defaultMessageForCode(ServerErrorCode c) {
  switch (c) {
    case ServerErrorCode.notFound:
      return 'Resource not found.';
    case ServerErrorCode.missingTarget:
      return 'Query parameter "target" is required.';
    case ServerErrorCode.invalidTarget:
      return 'Target URL is invalid.';
    case ServerErrorCode.upstreamError:
      return 'Upstream service error.';
    case ServerErrorCode.sessionCreateFailed:
      return 'Failed to create session.';
    case ServerErrorCode.sessionsListFailed:
      return 'Failed to list sessions.';
    case ServerErrorCode.sessionGetFailed:
      return 'Failed to get session.';
    case ServerErrorCode.framesListFailed:
      return 'Failed to list frames.';
    case ServerErrorCode.eventsListFailed:
      return 'Failed to list events.';
    case ServerErrorCode.httpListFailed:
      return 'Failed to list HTTP requests.';
    case ServerErrorCode.streamUnsupported:
      return 'Stream is not supported for this resource.';
    case ServerErrorCode.unknown:
      return 'Request failed.';
  }
}


