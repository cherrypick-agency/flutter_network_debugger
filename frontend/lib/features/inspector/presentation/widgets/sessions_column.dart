import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../core/di/di.dart';
import '../../application/stores/home_ui_store.dart';

class SessionsColumn extends StatefulWidget {
  const SessionsColumn({
    super.key,
    required this.showSearch,
    required this.onShowSearchChanged,
    required this.sessionSearchCtrl,
    required this.onSearchPrefsChanged,
    required this.onSearchSubmit,
    required this.selectedDomains,
    required this.onToggleDomain,
    required this.sessions,
    required this.sessionsCtrl,
    required this.groupBy,
    required this.selectedSessionId,
    required this.onSelectSession,
    required this.httpMeta,
  });

  final bool showSearch;
  final ValueChanged<bool> onShowSearchChanged;
  final TextEditingController sessionSearchCtrl;
  final VoidCallback onSearchPrefsChanged;
  final VoidCallback onSearchSubmit;

  final Set<String> selectedDomains;
  final void Function(String key, bool selected) onToggleDomain;

  final List<dynamic> sessions;
  final ScrollController sessionsCtrl;
  final String groupBy;
  final String? selectedSessionId;
  final void Function(String id) onSelectSession;
  final Map<String, Map<String, dynamic>> httpMeta;

  @override
  State<SessionsColumn> createState() => _SessionsColumnState();
}

class _SessionsColumnState extends State<SessionsColumn> {
  // Локально храним длину для автопрокрутки к низу при добавлении новых элементов
  int _lastSessionsLen = 0;
  final FocusNode _searchFocus = FocusNode();
  bool _stickToBottom = true; // автопрокрутка только если пользователь у низа

  @override
  void initState() {
    super.initState();
    widget.sessionsCtrl.addListener(_onScroll);
  }

  @override
  void dispose() {
    widget.sessionsCtrl.removeListener(_onScroll);
    _searchFocus.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant SessionsColumn oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!oldWidget.showSearch && widget.showSearch) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _searchFocus.requestFocus();
      });
    }
    if (oldWidget.sessionsCtrl != widget.sessionsCtrl) {
      oldWidget.sessionsCtrl.removeListener(_onScroll);
      widget.sessionsCtrl.addListener(_onScroll);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text('Sessions', style: context.appText.title),
            const SizedBox(width: 8),
            if (!widget.showSearch)
              IconButton(
                tooltip: 'Search',
                icon: const Icon(Icons.search, size: 18),
                onPressed: () => widget.onShowSearchChanged(true),
              )
            else
              IconButton(
                tooltip: 'Close search',
                icon: const Icon(Icons.close, size: 18),
                onPressed: () => widget.onShowSearchChanged(false),
              ),
            if (widget.showSearch)
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: TextField(
                    style: const TextStyle(fontSize: 12),
                    controller: widget.sessionSearchCtrl,
                    autofocus: true,
                    focusNode: _searchFocus,
                    decoration: const InputDecoration(
                      isDense: true,
                      hintText: 'Search sessions...',
                    ),
                    onChanged: (_) => widget.onSearchPrefsChanged(),
                    onSubmitted: (_) => widget.onSearchSubmit(),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 6),
        // Домены инлайн (до 3 рядов), далее скролл — только из отфильтрованных сессий
        Builder(
          builder: (_) {
            // Собираем счётчики доменов из текущего списка сессий
            final Map<String, int> counts = <String, int>{};
            for (final s in widget.sessions) {
              try {
                final host = Uri.parse((s.target as String)).host;
                if (host.isEmpty) continue;
                counts[host] = (counts[host] ?? 0) + 1;
              } catch (_) {}
            }
            // Гарантируем наличие выбранных доменов даже если они обнулились по другим фильтрам
            for (final d in widget.selectedDomains) {
              counts.putIfAbsent(d, () => 0);
            }
            // Стабильная сортировка по имени домена
            final domains =
                counts.keys.toList()
                  ..sort((a, b) => a.toLowerCase().compareTo(b.toLowerCase()));
            final labels = [for (final d in domains) '$d (${counts[d]})'];

            const double spacing = 6.0;
            const double runSpacing = 4.0;
            const double horizPad = 6.0;
            const double fontSize = 10.0;
            const double chipHPaddingTotal = horizPad * 2;
            const double rowHeight = 24.0;

            return LayoutBuilder(
              builder: (context, c) {
                double maxW = c.maxWidth.isFinite ? c.maxWidth : 320.0;
                final textStyle =
                    Theme.of(
                      context,
                    ).textTheme.labelSmall?.copyWith(fontSize: fontSize) ??
                    const TextStyle(fontSize: fontSize);

                final chipWidths = <double>[];
                for (final label in labels) {
                  final tp = TextPainter(
                    text: TextSpan(text: label, style: textStyle),
                    textDirection: TextDirection.ltr,
                  )..layout();
                  chipWidths.add(tp.size.width + chipHPaddingTotal);
                }

                int rows = 1;
                double lineW = 0;
                for (final cw in chipWidths) {
                  final add = lineW == 0 ? cw : cw + spacing;
                  if (lineW + add <= maxW) {
                    lineW += add;
                  } else {
                    rows++;
                    lineW = cw;
                  }
                }
                final int visibleRows = rows.clamp(0, 3);
                final double maxHeight =
                    visibleRows > 0
                        ? (rowHeight * visibleRows +
                            runSpacing * (visibleRows - 1))
                        : 0;

                Widget wrap = Wrap(
                  spacing: spacing,
                  runSpacing: runSpacing,
                  children: [
                    for (var i = 0; i < domains.length; i++)
                      Builder(
                        builder: (context) {
                          final key = domains[i];
                          final selected = widget.selectedDomains.contains(key);
                          return ChoiceChip(
                            label: Text(labels[i], style: textStyle),
                            labelPadding: const EdgeInsets.symmetric(
                              horizontal: horizPad,
                            ),
                            selected: selected,
                            visualDensity: const VisualDensity(
                              horizontal: -4,
                              vertical: -4,
                            ),
                            materialTapTargetSize:
                                MaterialTapTargetSize.shrinkWrap,
                            onSelected:
                                (_) => widget.onToggleDomain(key, !selected),
                          );
                        },
                      ),
                  ],
                );

                if (rows <= 3) {
                  return wrap;
                }
                return SizedBox(
                  width: double.infinity,
                  height: maxHeight,
                  child: Scrollbar(
                    thumbVisibility: false,
                    child: SingleChildScrollView(child: wrap),
                  ),
                );
              },
            );
          },
        ),
        const SizedBox(height: 8),
        Expanded(child: _buildSessionsList(context)),
      ],
    );
  }

  Widget _buildSessionsList(BuildContext context) {
    return ListView.builder(
      controller: widget.sessionsCtrl,
      itemCount: widget.sessions.length,
      itemBuilder: (ctx, i) {
        final s = widget.sessions[i];

        // Автопрокрутка к низу после билда, если список увеличился
        if (i == widget.sessions.length - 1 &&
            widget.sessions.length >= _lastSessionsLen &&
            _stickToBottom) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!widget.sessionsCtrl.hasClients) return;
            final max = widget.sessionsCtrl.position.maxScrollExtent;
            if (max > 0) {
              widget.sessionsCtrl.animateTo(
                max,
                duration: const Duration(milliseconds: 120),
                curve: Curves.easeOut,
              );
            }
          });
          _lastSessionsLen = widget.sessions.length;
        }

        final showHeader =
            widget.groupBy != 'none' &&
            (i == 0 || _groupKey(widget.sessions[i - 1]) != _groupKey(s));
        final header = _groupKey(s);

        final meta = (s.httpMeta ?? widget.httpMeta[s.id]) ?? const {};
        final method = (meta['method'] ?? '').toString();
        final status = int.tryParse((meta['status'] ?? '').toString()) ?? 0;
        final durationMs =
            int.tryParse((meta['durationMs'] ?? '').toString()) ?? -1;
        final cacheStatus = (meta['cache']?['status'] ?? '').toString();
        final hasResponse = status > 0;
        final isClosed = s.closedAt != null;
        final hasError = (s.error ?? '').toString().isNotEmpty;
        final corsOk =
            hasResponse ? ((meta['cors']?['ok'] ?? false) == true) : true;
        final isWs = (s.kind == 'ws') || (method.isEmpty && (s.kind == null));

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (showHeader)
              Padding(
                padding: const EdgeInsets.fromLTRB(8, 12, 8, 4),
                child: Text(
                  header,
                  style: Theme.of(context).textTheme.labelMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            MouseRegion(
              onEnter: (_) {
                sl<HomeUiStore>().setHoveredSessionId(s.id);
              },
              onExit: (_) {
                sl<HomeUiStore>().setHoveredSessionId(null);
              },
              child: InkWell(
                onTap: () => widget.onSelectSession(s.id),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    color:
                        widget.selectedSessionId == s.id
                            ? Theme.of(
                              context,
                            ).colorScheme.primary.withOpacity(0.06)
                            : null,
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // URL
                      Builder(
                        builder: (context) {
                          final errCode = (meta['errorCode'] ?? '').toString();
                          final warn =
                              errCode == 'TIMEOUT' ||
                              errCode == 'DNS' ||
                              errCode == 'TLS';
                          final mark = Theme.of(context).colorScheme.tertiary;
                          final child = _buildHighlightedUrl(
                            context,
                            s.target,
                            widget.sessionSearchCtrl.text,
                          );
                          if (!warn) return child;
                          return Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 4,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: mark.withOpacity(0.06),
                              borderRadius: BorderRadius.circular(4),
                              border: Border(
                                left: BorderSide(color: mark, width: 2),
                              ),
                            ),
                            child: child,
                          );
                        },
                      ),
                      const SizedBox(height: 4),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Wrap(
                              spacing: 6,
                              runSpacing: 4,
                              children: [
                                if (isWs)
                                  _chip(
                                    (s.closedAt == null)
                                        ? 'WS open'
                                        : 'WS closed',
                                    backgroundColor:
                                        (s.closedAt == null)
                                            ? Theme.of(context)
                                                .colorScheme
                                                .secondaryContainer
                                                .withOpacity(0.18)
                                            : Theme.of(context)
                                                .colorScheme
                                                .error
                                                .withOpacity(0.12),
                                    foregroundColor:
                                        (s.closedAt == null)
                                            ? context.appColors.success
                                            : Theme.of(
                                              context,
                                            ).colorScheme.error,
                                  ),
                                if (!isWs && method.isNotEmpty)
                                  _chip(
                                    method.toUpperCase(),
                                    backgroundColor:
                                        Theme.of(
                                          context,
                                        ).colorScheme.surfaceVariant,
                                    foregroundColor:
                                        Theme.of(
                                          context,
                                        ).colorScheme.onSurfaceVariant,
                                  ),
                                if (!isWs && status > 0)
                                  _chip(
                                    'HTTP $status',
                                    backgroundColor: _statusBg(context, status),
                                    foregroundColor: _statusFg(context, status),
                                  ),
                                if (!isWs &&
                                    !hasResponse &&
                                    isClosed &&
                                    hasError)
                                  Tooltip(
                                    message: s.error?.toString() ?? '',
                                    child: _chip(
                                      (() {
                                        final m =
                                            (widget.httpMeta[s.id] ?? const {});
                                        final code =
                                            (m['errorCode'] ?? '').toString();
                                        return code.isNotEmpty ? code : 'ERR';
                                      })(),
                                      backgroundColor: Theme.of(
                                        context,
                                      ).colorScheme.error.withOpacity(0.12),
                                      foregroundColor:
                                          Theme.of(context).colorScheme.error,
                                    ),
                                  ),
                                if (!isWs && !hasResponse && !isClosed)
                                  const SizedBox(
                                    width: 14,
                                    height: 14,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  ),
                                if (!isWs && durationMs >= 0)
                                  _chip(
                                    '${durationMs} ms',
                                    backgroundColor: _durationBg(
                                      context,
                                      durationMs,
                                    ),
                                    foregroundColor: _durationFg(
                                      context,
                                      durationMs,
                                    ),
                                  ),
                                if (!isWs && cacheStatus.isNotEmpty)
                                  (cacheStatus.toUpperCase() == 'MISS')
                                      ? _chipStrike(
                                        'cache',
                                        backgroundColor:
                                            Theme.of(
                                              context,
                                            ).colorScheme.surfaceVariant,
                                        foregroundColor:
                                            Theme.of(
                                              context,
                                            ).colorScheme.onSurfaceVariant,
                                      )
                                      : _chip(
                                        'cache: ${cacheStatus.toUpperCase()}',
                                      ),
                                if (!isWs && !corsOk)
                                  _chip(
                                    'CORS',
                                    backgroundColor: Theme.of(
                                      context,
                                    ).colorScheme.error.withOpacity(0.12),
                                    foregroundColor:
                                        Theme.of(context).colorScheme.error,
                                  ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                _formatTimeHMSSafe(s.startedAt as DateTime?),
                                textAlign: TextAlign.right,
                                style: Theme.of(context).textTheme.bodySmall,
                              ),
                              if (s.closedAt != null)
                                (() {
                                  final code =
                                      (meta['errorCode'] ?? '').toString();
                                  if (code.isEmpty)
                                    return const SizedBox.shrink();
                                  return Text(
                                    'Closed ($code)',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.labelSmall?.copyWith(
                                      color:
                                          Theme.of(context).colorScheme.error,
                                    ),
                                  );
                                })(),
                            ],
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const Divider(height: 1),
          ],
        );
      },
    );
  }

  // ===== helpers (UI форматирование) =====
  Widget _buildHighlightedUrl(BuildContext context, String text, String query) {
    final baseStyle =
        Theme.of(
          context,
        ).textTheme.bodyMedium?.copyWith(fontFamily: 'monospace') ??
        const TextStyle(fontFamily: 'monospace');
    final q = query.trim();
    if (q.isEmpty) {
      return Text(
        text,
        maxLines: 3,
        overflow: TextOverflow.ellipsis,
        style: baseStyle,
      );
    }
    final String src = text;
    final String srcLower = src.toLowerCase();
    final String qLower = q.toLowerCase();
    int start = 0;
    final spans = <InlineSpan>[];
    // Жёлтая подсветка как в других поисках
    final Color hl = context.appColors.warning.withValues(alpha: 0.35);
    while (true) {
      final idx = srcLower.indexOf(qLower, start);
      if (idx < 0) {
        if (start < src.length) {
          spans.add(TextSpan(text: src.substring(start), style: baseStyle));
        }
        break;
      }
      if (idx > start) {
        spans.add(TextSpan(text: src.substring(start, idx), style: baseStyle));
      }
      spans.add(
        TextSpan(
          text: src.substring(idx, idx + q.length),
          style: baseStyle.copyWith(backgroundColor: hl),
        ),
      );
      start = idx + q.length;
    }
    return RichText(
      text: TextSpan(children: spans),
      maxLines: 3,
      overflow: TextOverflow.ellipsis,
    );
  }

  Widget _chip(String text, {Color? backgroundColor, Color? foregroundColor}) {
    final bg = backgroundColor ?? Theme.of(context).colorScheme.surfaceVariant;
    final fg =
        foregroundColor ?? Theme.of(context).colorScheme.onSurfaceVariant;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        text,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(color: fg),
      ),
    );
  }

  Widget _chipStrike(
    String text, {
    Color? backgroundColor,
    Color? foregroundColor,
  }) {
    final base = _chip(
      text,
      backgroundColor: backgroundColor,
      foregroundColor: foregroundColor,
    );
    return Stack(
      alignment: Alignment.centerLeft,
      children: [
        base,
        Positioned.fill(
          child: IgnorePointer(
            child: CustomPaint(painter: ChipStrikePainter()),
          ),
        ),
      ],
    );
  }

  Color _statusBg(BuildContext context, int st) {
    final cs = Theme.of(context).colorScheme;
    if (st >= 500) return cs.error.withOpacity(0.12);
    if (st >= 400) return cs.tertiary.withOpacity(0.12);
    if (st >= 300) return cs.primary.withOpacity(0.12);
    return Colors.green.withOpacity(0.12);
  }

  Color _statusFg(BuildContext context, int st) {
    final cs = Theme.of(context).colorScheme;
    if (st >= 500) return cs.error;
    if (st >= 400) return cs.tertiary;
    if (st >= 300) return cs.primary;
    return Colors.green;
  }

  Color _durationBg(BuildContext context, int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green.withOpacity(0.12);
    if (ms < 1000) return cs.tertiary.withOpacity(0.12);
    return cs.error.withOpacity(0.12);
  }

  Color _durationFg(BuildContext context, int ms) {
    final cs = Theme.of(context).colorScheme;
    if (ms < 300) return Colors.green;
    if (ms < 1000) return cs.tertiary;
    return cs.error;
  }

  String _groupKey(dynamic s) {
    // Дублируем простую группировку по домену/роуту
    try {
      final uri = Uri.parse(s.target as String);
      if (widget.groupBy == 'domain') return uri.host;
      if (widget.groupBy == 'route')
        return '${uri.host}${uri.path.split('/').take(3).join('/')}';
    } catch (_) {}
    return '';
  }

  String _formatTimeHMSSafe(DateTime? dt) {
    if (dt == null) return '';
    final h = dt.hour.toString().padLeft(2, '0');
    final m = dt.minute.toString().padLeft(2, '0');
    final s = dt.second.toString().padLeft(2, '0');
    return '$h:$m:$s';
  }

  void _onScroll() {
    if (!widget.sessionsCtrl.hasClients) return;
    final pos = widget.sessionsCtrl.position;
    const threshold = 48.0;
    final atBottom = (pos.maxScrollExtent - pos.pixels) <= threshold;
    _stickToBottom = atBottom;
  }
}

// Простая диагональная зачёркивающая линия для чипа
class ChipStrikePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final p =
        Paint()
          ..color = const Color(0xFF9E9E9E)
          ..strokeWidth = 1.4
          ..style = PaintingStyle.stroke;
    final start = const Offset(2, 2);
    final end = Offset(size.width - 2, size.height - 2);
    canvas.drawLine(start, end, p);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
