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
  // Already resolved earlier — pass through as is
  if (e is ResolvedErrorMessage) {
    return e;
  }
  // Plain text
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
    final msg = e.messageFromServer.isNotEmpty
        ? e.messageFromServer
        : _defaultMessageForCode(c);
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
    // Form a more understandable brief description:
    // "HTTP 404 POST /_api/v1/sessions/{id}/tags - Not Found"
    String makeDescription() {
      final method = req.method;
      final url = req.uri.toString();
      String pathOrUrl = url;
      try {
        final u = Uri.parse(url);
        pathOrUrl = u.path.isNotEmpty ? u.path : url;
      } catch (_) {}
      final statusCode = resp?.statusCode;
      final statusMsg = resp?.statusMessage ?? '';
      final baseMsg = e.message.trim();
      final parts = <String>[];
      if (statusCode != null) {
        parts.add(
          'HTTP $statusCode${statusMsg.isNotEmpty ? ' $statusMsg' : ''}',
        );
      }
      parts.add('$method $pathOrUrl');
      if (baseMsg.isNotEmpty) {
        parts.add('- $baseMsg');
      }
      return parts.join(' ');
    }

    return ResolvedErrorMessage(
      title: 'Network error',
      description: makeDescription().isNotEmpty
          ? makeDescription()
          : (e.message.isNotEmpty ? e.message : 'Request failed.'),
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
    case ServerErrorCode.unauthorized:
      return 'Unauthorized';
    case ServerErrorCode.badJson:
      return 'Invalid JSON';
    case ServerErrorCode.validation:
      return 'Validation error';
    case ServerErrorCode.conflict:
      return 'Conflict';
    case ServerErrorCode.badParam:
      return 'Bad parameter';
    case ServerErrorCode.invalidUrl:
      return 'Invalid URL';
    case ServerErrorCode.badBase64:
      return 'Invalid body encoding';
    case ServerErrorCode.invalidStatus:
      return 'Invalid status code';
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
    case ServerErrorCode.unauthorized:
      return 'Authentication required or token invalid.';
    case ServerErrorCode.badJson:
      return 'Invalid JSON in request body.';
    case ServerErrorCode.validation:
      return 'Validation failed. Check the input values.';
    case ServerErrorCode.conflict:
      return 'Item already finalized or modified.';
    case ServerErrorCode.badParam:
      return 'Invalid query parameter value.';
    case ServerErrorCode.invalidUrl:
      return 'URL must be valid http or https.';
    case ServerErrorCode.badBase64:
      return 'Invalid base64-encoded body.';
    case ServerErrorCode.invalidStatus:
      return 'Status code must be between 100 and 599.';
    case ServerErrorCode.unknown:
      return 'Request failed.';
  }
}
