import 'dart:convert';

import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../core/di/di.dart';
import '../../application/stores/home_ui_store.dart';
import 'package:provider/provider.dart';
import '../../application/stores/aggregate_store.dart';
import '../../domain/entities/session.dart';
import '../../../../widgets/highlighted_url_widget.dart';
import '../../../../widgets/http_method_chip.dart';

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
  // Locally store length for auto-scroll to bottom when adding new items
  int _lastSessionsLen = 0;
  final FocusNode _searchFocus = FocusNode();
  bool _stickToBottom = true; // auto-scroll only if user is at bottom
  // For animated appearance of only new items
  final Set<String> _seenSessionIds = <String>{};

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
    // Domain counters
    final agg = context.watch<AggregateStore>().groups.toList();
    final Map<String, int> domainCounts = <String, int>{};
    if (agg.isNotEmpty) {
      for (final m in agg) {
        final key = (m['key'] ?? '').toString();
        final cnt = int.tryParse((m['count'] ?? '0').toString()) ?? 0;
        if (key.isEmpty) continue;
        domainCounts[key] = cnt;
      }
    } else {
      for (final s in widget.sessions) {
        try {
          final host = Uri.parse((s.target as String)).host;
          if (host.isEmpty) continue;
          domainCounts[host] = (domainCounts[host] ?? 0) + 1;
        } catch (_) {}
      }
    }
    for (final d in widget.selectedDomains) {
      domainCounts.putIfAbsent(d, () => 0);
    }
    final domainKeys = domainCounts.keys.toList()
      ..sort((a, b) => a.toLowerCase().compareTo(b.toLowerCase()));

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
            if (domainKeys.isNotEmpty)
              Badge(
                isLabelVisible: widget.selectedDomains.isNotEmpty,
                label: Text('${widget.selectedDomains.length}'),
                child: IconButton(
                  tooltip: 'Filter by domain',
                  icon: const Icon(Icons.filter_list, size: 18),
                  onPressed: () =>
                      _showDomainFilter(context, domainCounts, domainKeys),
                ),
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
        if (widget.selectedDomains.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Wrap(
              spacing: 4,
              runSpacing: 4,
              children: [
                for (final d in widget.selectedDomains)
                  Chip(
                    label: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(d, style: const TextStyle(fontSize: 10)),
                        const SizedBox(width: 4),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 4,
                            vertical: 1,
                          ),
                          decoration: BoxDecoration(
                            color: Theme.of(
                              context,
                            ).colorScheme.onSurface.withValues(alpha: 0.08),
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            '${domainCounts[d] ?? 0}',
                            style: TextStyle(
                              fontSize: 9,
                              color: Theme.of(
                                context,
                              ).colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ),
                      ],
                    ),
                    deleteIcon: const Icon(Icons.close, size: 12),
                    onDeleted: () => widget.onToggleDomain(d, false),
                    visualDensity: const VisualDensity(
                      horizontal: -4,
                      vertical: -4,
                    ),
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    padding: EdgeInsets.zero,
                    labelPadding: const EdgeInsets.only(left: 4),
                  ),
              ],
            ),
          ),
        const SizedBox(height: 4),
        Expanded(child: _buildSessionsList(context)),
      ],
    );
  }

  void _showDomainFilter(
    BuildContext context,
    Map<String, int> counts,
    List<String> domains,
  ) {
    final localSelected = Set<String>.from(widget.selectedDomains);
    showDialog<void>(
      context: context,
      builder: (_) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          titlePadding: const EdgeInsets.fromLTRB(16, 16, 8, 0),
          contentPadding: const EdgeInsets.symmetric(vertical: 8),
          title: Row(
            children: [
              const Text('Domains'),
              const Spacer(),
              if (localSelected.isNotEmpty)
                TextButton(
                  onPressed: () {
                    for (final d in localSelected.toList()) {
                      widget.onToggleDomain(d, false);
                    }
                    setDialogState(() => localSelected.clear());
                  },
                  child: const Text('Clear all'),
                ),
            ],
          ),
          content: SizedBox(
            width: 380,
            child: ListView(
              shrinkWrap: true,
              children: [
                for (final d in domains)
                  CheckboxListTile(
                    title: Text(d, style: const TextStyle(fontSize: 13)),
                    secondary: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: Theme.of(
                          ctx,
                        ).colorScheme.onSurface.withValues(alpha: 0.08),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        '${counts[d] ?? 0}',
                        style: TextStyle(
                          fontSize: 11,
                          color: Theme.of(ctx).colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ),
                    value: localSelected.contains(d),
                    dense: true,
                    onChanged: (v) {
                      final add = v ?? false;
                      if (add) {
                        localSelected.add(d);
                      } else {
                        localSelected.remove(d);
                      }
                      widget.onToggleDomain(d, add);
                      setDialogState(() {});
                    },
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSessionsList(BuildContext context) {
    if (widget.sessions.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Text(
            'Make a request to see items',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
            textAlign: TextAlign.center,
          ),
        ),
      );
    }
    // Determine the common domain among current visible sessions.
    // If all sessions share the same host — hide the domain in the list below
    // to make the row shorter and easier to read.
    final String? commonHost = (() {
      String? host;
      for (final s in widget.sessions) {
        try {
          final h = Uri.parse((s.target as String)).host;
          if (h.isEmpty)
            return null; // empty/broken URL encountered — don't hide
          host ??= h;
          if (host != h) return null; // different domains — show as is
        } catch (_) {
          return null;
        }
      }
      return host; // same host for all items
    })();

    return ListView.builder(
      controller: widget.sessionsCtrl,
      itemCount: widget.sessions.length,
      itemBuilder: (ctx, i) {
        final s = widget.sessions[i];

        // Auto-scroll to bottom after build if list increased
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
        final corsOk = hasResponse
            ? ((meta['cors']?['ok'] ?? true) == true)
            : true;
        final isWs = (s.kind == 'ws') || (method.isEmpty && (s.kind == null));

        final idStr = (s.id as String);
        final isNew = !_seenSessionIds.contains(idStr);
        _seenSessionIds.add(idStr);

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
            _Appear(
              key: ValueKey('s_row_$idStr'),
              animate: isNew,
              child: MouseRegion(
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
                      color: widget.selectedSessionId == s.id
                          ? Theme.of(
                              context,
                            ).colorScheme.primary.withOpacity(0.06)
                          : null,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Process icon + URL
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Process icon
                            _buildProcessIcon(s.processInfo),
                            // URL
                            Expanded(
                              child: Builder(
                                builder: (context) {
                                  final errCode = (meta['errorCode'] ?? '')
                                      .toString();
                                  final warn =
                                      errCode == 'TIMEOUT' ||
                                      errCode == 'DNS' ||
                                      errCode == 'TLS';
                                  final mark = Theme.of(
                                    context,
                                  ).colorScheme.tertiary;
                                  final child = HighlightedUrlWidget(
                                    url: s.target as String,
                                    searchQuery: widget.sessionSearchCtrl.text,
                                    suppressHost: commonHost,
                                    maxLines: 3,
                                    overflow: TextOverflow.ellipsis,
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
                            ),
                          ],
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
                                  // Process name
                                  if (_shouldShowProcessName(
                                    s.processInfo?.name,
                                  ))
                                    _chip(
                                      s.processInfo!.name,
                                      backgroundColor: Theme.of(
                                        context,
                                      ).colorScheme.primaryContainer,
                                      foregroundColor: Theme.of(
                                        context,
                                      ).colorScheme.onPrimaryContainer,
                                    ),
                                  if (isWs)
                                    _chip(
                                      (s.closedAt == null)
                                          ? 'WS open'
                                          : 'WS closed',
                                      backgroundColor: (s.closedAt == null)
                                          ? Theme.of(context)
                                                .colorScheme
                                                .secondaryContainer
                                                .withOpacity(0.18)
                                          : Theme.of(context).colorScheme.error
                                                .withOpacity(0.12),
                                      foregroundColor: (s.closedAt == null)
                                          ? context.appColors.success
                                          : Theme.of(context).colorScheme.error,
                                    ),
                                  if (isWs && isClosed && hasError)
                                    Tooltip(
                                      message: s.error?.toString() ?? '',
                                      child: _chip(
                                        'ERR',
                                        backgroundColor: Theme.of(context)
                                            .colorScheme
                                            .error
                                            .withValues(alpha: 0.12),
                                        foregroundColor: Theme.of(
                                          context,
                                        ).colorScheme.error,
                                      ),
                                    ),
                                  if (isWs && s.isSocketIo == true)
                                    _chip('socket.io'),
                                  if (!isWs && method.isNotEmpty)
                                    _chip(
                                      method.toUpperCase(),
                                      backgroundColor: getHttpMethodBgColor(
                                        context,
                                        method,
                                      ),
                                      foregroundColor: getHttpMethodFgColor(
                                        context,
                                        method,
                                      ),
                                    ),
                                  if (!isWs && status > 0)
                                    _chip(
                                      'HTTP $status',
                                      backgroundColor: _statusBg(
                                        context,
                                        status,
                                      ),
                                      foregroundColor: _statusFg(
                                        context,
                                        status,
                                      ),
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
                                              (widget.httpMeta[s.id] ??
                                              const {});
                                          final code = (m['errorCode'] ?? '')
                                              .toString();
                                          return code.isNotEmpty ? code : 'ERR';
                                        })(),
                                        backgroundColor: Theme.of(
                                          context,
                                        ).colorScheme.error.withOpacity(0.12),
                                        foregroundColor: Theme.of(
                                          context,
                                        ).colorScheme.error,
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
                                            backgroundColor: Theme.of(
                                              context,
                                            ).colorScheme.surfaceVariant,
                                            foregroundColor: Theme.of(
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
                                      foregroundColor: Theme.of(
                                        context,
                                      ).colorScheme.error,
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
                                    final code = (meta['errorCode'] ?? '')
                                        .toString();
                                    if (code.isEmpty)
                                      return const SizedBox.shrink();
                                    return Text(
                                      'Closed ($code)',
                                      style: Theme.of(context)
                                          .textTheme
                                          .labelSmall
                                          ?.copyWith(
                                            color: Theme.of(
                                              context,
                                            ).colorScheme.error,
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
            ),
            const Divider(height: 1),
          ],
        );
      },
    );
  }

  // ===== helpers (UI formatting) =====
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

  bool _shouldShowProcessName(String? name) {
    final n = (name ?? '').trim().toLowerCase();
    if (n.isEmpty) return false;
    // Hide technical UI/proxy process names
    if (n == 'reverse proxy' || n == 'runner') return false;
    return true;
  }

  String _groupKey(dynamic s) {
    // Duplicate simple grouping by domain/route
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

    // Convert UTC to local time
    final localDt = dt.toLocal();

    // Check if the date is today (by local time)
    final now = DateTime.now();
    final isToday =
        localDt.year == now.year &&
        localDt.month == now.month &&
        localDt.day == now.day;

    final h = localDt.hour.toString().padLeft(2, '0');
    final m = localDt.minute.toString().padLeft(2, '0');
    final s = localDt.second.toString().padLeft(2, '0');
    final time = '$h:$m:$s';

    if (isToday) {
      return time; // Time only
    } else {
      // Date + time (without year)
      const months = [
        'Jan',
        'Feb',
        'Mar',
        'Apr',
        'May',
        'Jun',
        'Jul',
        'Aug',
        'Sep',
        'Oct',
        'Nov',
        'Dec',
      ];
      final monthStr = months[localDt.month - 1];
      final day = localDt.day;
      return '$monthStr $day $time';
    }
  }

  void _onScroll() {
    if (!widget.sessionsCtrl.hasClients) return;
    final pos = widget.sessionsCtrl.position;
    const threshold = 48.0;
    final atBottom = (pos.maxScrollExtent - pos.pixels) <= threshold;
    _stickToBottom = atBottom;
  }

  Widget _buildProcessIcon(ProcessInfo? processInfo) {
    if (processInfo?.icon == null) {
      return const SizedBox.shrink();
    }

    try {
      final iconData = processInfo!.icon!.data;
      final bytes = base64Decode(iconData);

      return Padding(
        padding: const EdgeInsets.only(right: 6),
        child: CircleAvatar(
          radius: 10,
          backgroundColor: Colors.transparent,
          child: ClipOval(
            child: Image.memory(
              bytes,
              width: 20,
              height: 20,
              fit: BoxFit.cover,
              errorBuilder: (_, __, ___) => const Icon(Icons.apps, size: 14),
            ),
          ),
        ),
      );
    } catch (e) {
      return const Padding(
        padding: EdgeInsets.only(right: 6),
        child: CircleAvatar(radius: 10, child: Icon(Icons.apps, size: 14)),
      );
    }
  }
}

class _Appear extends StatelessWidget {
  const _Appear({super.key, required this.child, required this.animate});
  final Widget child;
  final bool animate;
  @override
  Widget build(BuildContext context) {
    if (!animate) return child;
    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: const Duration(milliseconds: 220),
      curve: Curves.easeOutCubic,
      builder: (context, t, child) {
        return Opacity(
          opacity: t,
          child: Transform.translate(
            offset: Offset(0, (1 - t) * 8),
            child: child,
          ),
        );
      },
      child: child,
    );
  }
}

// Simple diagonal strike-through line for chip
class ChipStrikePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final p = Paint()
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
