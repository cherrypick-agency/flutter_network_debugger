import '../../application/stores/sessions_store.dart';
import '../../../filters/application/stores/sessions_filters_store.dart';

// Чистая фильтрация списка сессий под UI-колонку.
List<dynamic> filterVisibleSessions({
  required SessionsStore store,
  required SessionsFiltersStore filters,
  required DateTimeRangeWrapper? selectedRange,
  required Set<String> selectedDomains,
  required Map<String, Map<String, dynamic>> httpMeta,
  required DateTime? since,
  required Set<String> ignoredIds,
  // Быстрые фильтры: наборы тегов
  Set<String> quickTypes = const <String>{},
  Set<String> quickStatusGroups = const <String>{},
}) {
  final src = store.items.toList();
  final filtered = src
      .where((s) {
        if (ignoredIds.contains(s.id)) return false;
        if (since != null) {
          final end = s.closedAt ?? DateTime.now();
          // Показываем только сессии, у которых конец \(закрытие или now\) >= since
          if (end.isBefore(since)) return false;
        }
        if (selectedRange != null) {
          final start = s.startedAt;
          final end = s.closedAt ?? s.startedAt;
          if (start == null) return false;
          final inRange =
              !((end != null && end.isBefore(selectedRange.start)) ||
                  start.isAfter(selectedRange.end));
          if (!inRange) return false;
        }
        if (selectedDomains.isNotEmpty) {
          try {
            final host = Uri.parse(s.target).host;
            if (!selectedDomains.contains(host)) return false;
          } catch (_) {}
        }
        final id = s.id;
        final m = (s.httpMeta ?? httpMeta[id]) ?? const {};
        if (filters.httpMethod != 'any') {
          if ((m['method'] ?? '') != filters.httpMethod) return false;
        }
        if (filters.httpStatus != 'any') {
          final st = int.tryParse((m['status'] ?? '0').toString()) ?? 0;
          if (filters.httpStatus == '2xx' && (st < 200 || st > 299))
            return false;
          if (filters.httpStatus == '3xx' && (st < 300 || st > 399))
            return false;
          if (filters.httpStatus == '4xx' && (st < 400 || st > 499))
            return false;
          if (filters.httpStatus == '5xx' && (st < 500 || st > 599))
            return false;
        }
        if (filters.httpMinDurationMs > 0) {
          final d = int.tryParse((m['durationMs'] ?? '0').toString()) ?? 0;
          if (d < filters.httpMinDurationMs) return false;
        }
        if (filters.httpMime.isNotEmpty) {
          final mime = (m['mime'] ?? '').toString().toLowerCase();
          if (!mime.contains(filters.httpMime.toLowerCase())) return false;
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
          if (filters.headerVal.isEmpty && hv.isEmpty) return false;
        }
        // Быстрые группы статусов: если заданы — оставляем только попадающие в любую выбранную группу
        if (quickStatusGroups.isNotEmpty) {
          final st = int.tryParse((m['status'] ?? '0').toString()) ?? 0;
          bool statusOk = false;
          for (final g in quickStatusGroups) {
            if (g == '1xx' && st >= 100 && st <= 199) statusOk = true;
            if (g == '2xx' && st >= 200 && st <= 299) statusOk = true;
            if (g == '3xx' && st >= 300 && st <= 399) statusOk = true;
            if (g == '4xx' && st >= 400 && st <= 499) statusOk = true;
            if (g == '5xx' && st >= 500 && st <= 599) statusOk = true;
          }
          // Для WebSocket у статуса может не быть — в этом случае не проходим статусный фильтр
          if (!statusOk) return false;
        }

        // Быстрые типы контента/протокола: если заданы — совпадение по любому тегу
        if (quickTypes.isNotEmpty) {
          final tags = _tagsForSession(s, m);
          if (!quickTypes.any(tags.contains)) return false;
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

// Набор коротких тегов по сессии: протоколы (http, https, ws) и типы контента
Set<String> _tagsForSession(dynamic s, Map<String, dynamic> meta) {
  final tags = <String>{};
  try {
    final uri = Uri.parse((s.target as String));
    final scheme = uri.scheme.toLowerCase();
    if (scheme == 'https') tags.add('https');
    if (scheme == 'http') tags.add('http');
    if (scheme == 'ws' || scheme == 'wss') tags.add('ws');
    // Явная метка kind для надёжности
    final kind = (s.kind as String?);
    if (kind == 'ws') tags.add('ws');
  } catch (_) {}

  final mime = (meta['mime'] ?? '').toString().toLowerCase();
  if (mime.isNotEmpty) {
    if (mime.contains('json')) tags.add('json');
    if (mime.contains('x-www-form-urlencoded') ||
        mime.contains('multipart/form-data'))
      tags.add('form');
    if (mime.contains('xml')) tags.add('xml');
    if (mime.contains('javascript') || mime.endsWith('/js')) tags.add('js');
    if (mime.contains('css')) tags.add('css');
    if (mime.contains('graphql')) tags.add('graphql');
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

  // Грубая эвристика для GraphQL по пути
  try {
    final uri = Uri.parse((s.target as String));
    final path = uri.path.toLowerCase();
    if (path.contains('graphql')) tags.add('graphql');
  } catch (_) {}

  // Если ничего не распознали и это не websocket — считаем "other"
  if (tags.isEmpty || (tags.length == 1 && tags.contains('http'))) {
    // http без mime тоже можно считать other, но оставим общий флаг
    tags.add('other');
  }
  return tags;
}

// Обёртка над DateTimeRange, чтобы не тянуть material в util (оставляем совместимость типов)
class DateTimeRangeWrapper {
  DateTimeRangeWrapper({required this.start, required this.end});
  final DateTime start;
  final DateTime end;
}
