import '../../application/stores/sessions_store.dart';
import '../../../filters/application/stores/sessions_filters_store.dart';

// Pure session list filtering for UI column.
List<dynamic> filterVisibleSessions({
  required SessionsStore store,
  required SessionsFiltersStore filters,
  required DateTimeRangeWrapper? selectedRange,
  required Set<String> selectedDomains,
  required Map<String, Map<String, dynamic>> httpMeta,
  required DateTime? since,
  required Set<String> ignoredIds,
  // Quick filters: tag sets
  Set<String> quickTypes = const <String>{},
  Set<String> quickStatusGroups = const <String>{},
}) {
  final src = store.items.toList();
  final filtered = src
      .where((s) {
        if (ignoredIds.contains(s.id)) {
          return false;
        }
        if (since != null) {
          final end = s.closedAt ?? DateTime.now();
          // Show only sessions where end (close time or now) >= since
          if (end.isBefore(since)) {
            return false;
          }
        }
        if (selectedRange != null) {
          final start = s.startedAt;
          final end = s.closedAt ?? s.startedAt;
          if (start == null) {
            return false;
          }
          final inRange =
              !((end != null && end.isBefore(selectedRange.start)) ||
                  start.isAfter(selectedRange.end));
          if (!inRange) {
            return false;
          }
        }
        if (selectedDomains.isNotEmpty) {
          try {
            final host = Uri.parse(s.target).host;
            if (!selectedDomains.contains(host)) {
              return false;
            }
          } catch (_) {}
        }
        final id = s.id;
        final m = (s.httpMeta ?? httpMeta[id]) ?? const {};
        final errorCategory = (m['errorCategory'] ?? '').toString();
        final errorCode = (m['errorCode'] ?? '').toString();
        final status = int.tryParse((m['status'] ?? '0').toString()) ?? 0;
        final hasSessionError = (s.error ?? '').toString().trim().isNotEmpty;
        final hasMetaError = errorCode.trim().isNotEmpty;
        final hasHttpErrorStatus = status >= 400;
        final isError = hasSessionError || hasMetaError || hasHttpErrorStatus;
        final isCancelled =
            errorCategory == 'cancel' ||
            errorCode == 'CANCELED' ||
            errorCode == 'CANCELLED' ||
            errorCode == 'CANCEL';

        if (!filters.showCancelled && isCancelled) {
          return false;
        }
        if (filters.onlyErrors && !isError) {
          return false;
        }

        if (filters.httpMethod != 'any') {
          if ((m['method'] ?? '') != filters.httpMethod) {
            return false;
          }
        }
        if (filters.httpStatus != 'any') {
          if (filters.httpStatus == '2xx' && (status < 200 || status > 299)) {
            return false;
          }
          if (filters.httpStatus == '3xx' && (status < 300 || status > 399)) {
            return false;
          }
          if (filters.httpStatus == '4xx' && (status < 400 || status > 499)) {
            return false;
          }
          if (filters.httpStatus == '5xx' && (status < 500 || status > 599)) {
            return false;
          }
        }
        if (filters.httpMinDurationMs > 0) {
          final d = int.tryParse((m['durationMs'] ?? '0').toString()) ?? 0;
          if (d < filters.httpMinDurationMs) {
            return false;
          }
        }
        if (filters.httpMime.isNotEmpty) {
          final mime = (m['mime'] ?? '').toString().toLowerCase();
          if (!mime.contains(filters.httpMime.toLowerCase())) {
            return false;
          }
        }
        if (filters.headerKey.isNotEmpty) {
          final headers =
              (m['headers'] as Map?)?.map(
                (k, v) => MapEntry(k.toString().toLowerCase(), v.toString()),
              ) ??
              {};
          final hv = headers[filters.headerKey.toLowerCase()] ?? '';
          if (filters.headerVal.isNotEmpty && !hv.contains(filters.headerVal)) {
            return false;
          }
          if (filters.headerVal.isEmpty && hv.isEmpty) {
            return false;
          }
        }
        // Quick status groups: if set — keep only those matching any selected group
        if (quickStatusGroups.isNotEmpty) {
          bool statusOk = false;
          for (final g in quickStatusGroups) {
            if (g == '1xx' && status >= 100 && status <= 199) {
              statusOk = true;
            }
            if (g == '2xx' && status >= 200 && status <= 299) {
              statusOk = true;
            }
            if (g == '3xx' && status >= 300 && status <= 399) {
              statusOk = true;
            }
            if (g == '4xx' && status >= 400 && status <= 499) {
              statusOk = true;
            }
            if (g == '5xx' && status >= 500 && status <= 599) {
              statusOk = true;
            }
          }
          // For WebSocket there may be no status — in this case status filter is not passed
          if (!statusOk) {
            return false;
          }
        }

        // Quick content/protocol types: if set — match any tag
        if (quickTypes.isNotEmpty) {
          final tags = _tagsForSession(s, m);
          if (!quickTypes.any(tags.contains)) {
            return false;
          }
        }

        return true;
      })
      .toList(growable: false);

  if (filters.groupBy == 'none') return filtered;

  String keyFor(dynamic s) {
    try {
      final uri = Uri.parse(s.target as String);
      if (filters.groupBy == 'domain') return uri.host;
      if (filters.groupBy == 'route') {
        return '${uri.host}${uri.path.split('/').take(3).join('/')}';
      }
    } catch (_) {}
    return '';
  }

  filtered.sort((a, b) => keyFor(a).compareTo(keyFor(b)));
  return filtered;
}

// Set of short tags for session: protocols (http, https, ws) and content types
Set<String> _tagsForSession(dynamic s, Map<String, dynamic> meta) {
  final tags = <String>{};
  try {
    final uri = Uri.parse((s.target as String));
    final scheme = uri.scheme.toLowerCase();
    if (scheme == 'https') {
      tags.add('https');
    }
    if (scheme == 'http') {
      tags.add('http');
    }
    if (scheme == 'ws' || scheme == 'wss') {
      tags.add('ws');
    }
    // Explicit kind label for reliability
    final kind = (s.kind as String?);
    if (kind == 'ws') {
      tags.add('ws');
    }
  } catch (_) {}

  final mime = (meta['mime'] ?? '').toString().toLowerCase();
  if (mime.isNotEmpty) {
    if (mime.contains('json')) {
      tags.add('json');
    }
    if (mime.contains('x-www-form-urlencoded') ||
        mime.contains('multipart/form-data')) {
      tags.add('form');
    }
    if (mime.contains('xml')) {
      tags.add('xml');
    }
    if (mime.contains('javascript') || mime.endsWith('/js')) {
      tags.add('js');
    }
    if (mime.contains('css')) {
      tags.add('css');
    }
    if (mime.contains('graphql')) {
      tags.add('graphql');
    }
    if (mime.contains('html') || mime.contains('pdf') || mime.contains('rtf')) {
      tags.add('document');
    }
    if (mime.startsWith('image/') ||
        mime.startsWith('audio/') ||
        mime.startsWith('video/') ||
        mime.startsWith('font/')) {
      tags.add('media');
    }
  }

  // Rough heuristic for GraphQL by path
  try {
    final uri = Uri.parse((s.target as String));
    final path = uri.path.toLowerCase();
    if (path.contains('graphql')) {
      tags.add('graphql');
    }
  } catch (_) {}

  // If nothing recognized and not websocket — consider "other"
  if (tags.isEmpty || (tags.length == 1 && tags.contains('http'))) {
    // http without mime can also be considered other, but we keep the common flag
    tags.add('other');
  }
  return tags;
}

// Wrapper over DateTimeRange to avoid pulling material into util (keeping type compatibility)
class DateTimeRangeWrapper {
  DateTimeRangeWrapper({required this.start, required this.end});
  final DateTime start;
  final DateTime end;
}
