import 'dart:math' as math;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../theme/context_ext.dart';
import '../../domain/entities/session.dart';

class WaterfallTimeline extends StatefulWidget {
  const WaterfallTimeline({
    super.key,
    required this.sessions,
    this.onIntervalSelected,
    this.onSessionSelected,
    this.autoExtendViewport = true,
    this.height = 140,
    this.padding = const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
    this.initialViewportPadding = const Duration(seconds: 2),
    this.initialRange,
  });

  final List<Session> sessions;
  final ValueChanged<DateTimeRange>? onIntervalSelected;
  final ValueChanged<Session>? onSessionSelected;
  final bool autoExtendViewport;
  final double height;
  final EdgeInsets padding;
  final Duration initialViewportPadding;
  final DateTimeRange? initialRange;

  @override
  State<WaterfallTimeline> createState() => _WaterfallTimelineState();
}

class _WaterfallTimelineState extends State<WaterfallTimeline> with SingleTickerProviderStateMixin {
  final ScrollController _hCtrl = ScrollController();
  final ScrollController _vCtrl = ScrollController();

  late DateTime _viewStart;
  late DateTime _viewEnd;
  bool _followLive = true;

  final double _laneHeight = 18;
  final double _laneGap = 6;
  final double _axisHeight = 22;

  // selection
  bool _dragging = false;
  Offset? _dragStart;
  Offset? _dragCurrent;
  bool _resizingStart = false;
  bool _resizingEnd = false;

  String? _hoverSessionId;
  DateTimeRange? _selectedRange;
  late final AnimationController _pulse;

  @override
  void initState() {
    super.initState();
    _vCtrl.addListener(() { if (mounted) setState(() {}); });
    _pulse = AnimationController(vsync: this, duration: const Duration(seconds: 2))
      ..repeat(reverse: true)
      ..addListener(() { if (mounted) setState(() {}); });
    final now = DateTime.now();
    if (widget.initialRange != null && widget.initialRange!.end.isAfter(widget.initialRange!.start)) {
      _viewStart = widget.initialRange!.start;
      _viewEnd = widget.initialRange!.end;
    } else {
      final times = _timesOf(widget.sessions);
      final start = (times.start ?? now);
      final end = (times.end ?? now).add(widget.initialViewportPadding);
      _viewStart = start;
      _viewEnd = end.isAfter(_viewStart.add(const Duration(milliseconds: 100)))
          ? end
          : _viewStart.add(const Duration(seconds: 10));
    }
  }

  @override
  void didUpdateWidget(covariant WaterfallTimeline oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.autoExtendViewport && widget.sessions.isNotEmpty) {
      final times = _timesOf(widget.sessions);
      final s = times.start;
      final e = times.end;
      if (s != null && s.isBefore(_viewStart)) {
        _viewStart = s.subtract(widget.initialViewportPadding);
      }
      if (e != null && e.isAfter(_viewEnd)) {
        final span = _viewEnd.difference(_viewStart);
        if (_followLive) {
          _viewEnd = e.add(widget.initialViewportPadding);
          _viewStart = _viewEnd.subtract(span);
        } else {
          _viewEnd = e.add(widget.initialViewportPadding);
        }
      }
      if (!_viewEnd.isAfter(_viewStart)) {
        _viewEnd = _viewStart.add(const Duration(seconds: 10));
      }
    }
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  ({DateTime? start, DateTime? end}) _timesOf(List<Session> list) {
    DateTime? s;
    DateTime? e;
    for (final it in list) {
      final st = it.startedAt;
      final en = it.closedAt ?? it.startedAt?.add(const Duration(milliseconds: 1));
      if (st != null) {
        s = (s == null || st.isBefore(s)) ? st : s;
      }
      if (en != null) {
        e = (e == null || en.isAfter(e)) ? en : e;
      }
    }
    return (start: s, end: e);
  }

  double _pxPerMs(double targetWidthMs) {
    // Keep width comfortable (1.6k-5k px) for horizontal scrolling
    final totalMs = _viewEnd.difference(_viewStart).inMilliseconds.toDouble().clamp(1000.0, 3600 * 1000.0);
    final width = math.max(1600.0, math.min(5000.0, totalMs / 8)); // ~8 ms per px by default
    return width / totalMs;
  }

  int _currentTickStepMs() {
    final totalMs = _viewEnd.difference(_viewStart).inMilliseconds;
    if (totalMs <= 5000) return 250;
    if (totalMs <= 15000) return 500;
    if (totalMs <= 60000) return 1000;
    if (totalMs <= 5 * 60000) return 5000;
    return 10000;
  }

  DateTime _snapToTick(DateTime t) {
    final step = _currentTickStepMs();
    final base = _viewStart.millisecondsSinceEpoch;
    final delta = t.millisecondsSinceEpoch - base;
    final snapped = (delta / step).round() * step;
    return DateTime.fromMillisecondsSinceEpoch(base + snapped);
  }

  double _timeToX(DateTime t, double pxPerMs) {
    final ms = t.difference(_viewStart).inMilliseconds.toDouble();
    return widget.padding.left + ms * pxPerMs;
  }

  DateTime _xToTime(double x, double pxPerMs) {
    final rel = (x - widget.padding.left) / pxPerMs;
    final ms = rel.clamp(0, double.infinity).round();
    return _viewStart.add(Duration(milliseconds: ms));
  }

  @override
  Widget build(BuildContext context) {
    return CallbackShortcuts(
      bindings: <ShortcutActivator, VoidCallback>{
        const SingleActivator(LogicalKeyboardKey.equal): () => _zoom(0.8),
        const SingleActivator(LogicalKeyboardKey.minus): () => _zoom(1.25),
        const SingleActivator(LogicalKeyboardKey.arrowLeft): () => _pan(-0.2),
        const SingleActivator(LogicalKeyboardKey.arrowRight): () => _pan(0.2),
        const SingleActivator(LogicalKeyboardKey.escape): () => setState(() {
              _dragStart = null;
              _dragCurrent = null;
            }),
      },
      child: Focus(
        autofocus: false,
        child: SizedBox(
          height: widget.height,
          child: LayoutBuilder(
            builder: (context, constraints) {
              final pxPerMs =
                  _pxPerMs(_viewEnd.difference(_viewStart).inMilliseconds.toDouble());
              final totalMs =
                  _viewEnd.difference(_viewStart).inMilliseconds.toDouble();
              final contentWidth =
                  (totalMs * pxPerMs).clamp(800.0, 100000.0);

              final items = _computeLayoutItems(widget.sessions);
              final lanes = items.isEmpty
                  ? 1
                  : (items.map((e) => e.lane).reduce(math.max) + 1);
              final lanesHeight = lanes * (_laneHeight + _laneGap);
              final contentHeight =
                  _axisHeight + lanesHeight + widget.padding.vertical;
              final vOff = _vCtrl.hasClients ? _vCtrl.offset : 0.0;

              return Scrollbar(
                controller: _hCtrl,
                thumbVisibility: true,
                child: SingleChildScrollView(
                  controller: _hCtrl,
                  scrollDirection: Axis.horizontal,
                  child: Scrollbar(
                    controller: _vCtrl,
                    thumbVisibility: true,
                    child: SingleChildScrollView(
                      controller: _vCtrl,
                      scrollDirection: Axis.vertical,
                      child: SizedBox(
                        width: contentWidth + widget.padding.horizontal,
                        height: math.max(widget.height, contentHeight),
                        child: Listener(
                          onPointerDown: (ev) {
                            if (ev.kind == PointerDeviceKind.mouse &&
                                ev.buttons != kPrimaryMouseButton) return;
                            setState(() {
                              // detect edge resize if selection exists
                              _resizingStart = false; _resizingEnd = false;
                              final pxPer = _pxPerMs(_viewEnd.difference(_viewStart).inMilliseconds.toDouble());
                              if (_selectedRange != null) {
                                final sx = _timeToX(_selectedRange!.start, pxPer);
                                final ex = _timeToX(_selectedRange!.end, pxPer);
                                if ((ev.localPosition.dx - sx).abs() <= 6) { _resizingStart = true; }
                                if ((ev.localPosition.dx - ex).abs() <= 6) { _resizingEnd = true; }
                              }
                              if (!_resizingStart && !_resizingEnd) {
                                _dragging = true;
                                _dragStart = ev.localPosition;
                                _dragCurrent = ev.localPosition;
                              }
                            });
                          },
                          onPointerMove: (ev) {
                            setState(() {
                              final pxPer = _pxPerMs(_viewEnd.difference(_viewStart).inMilliseconds.toDouble());
                              if (_resizingStart && _selectedRange != null) {
                                final t = _xToTime(ev.localPosition.dx, pxPer);
                                if (t.isBefore(_selectedRange!.end)) { _selectedRange = DateTimeRange(start: t, end: _selectedRange!.end); }
                                if (widget.onIntervalSelected != null) widget.onIntervalSelected!(_selectedRange!);
                                return;
                              }
                              if (_resizingEnd && _selectedRange != null) {
                                final t = _xToTime(ev.localPosition.dx, pxPer);
                                if (t.isAfter(_selectedRange!.start)) { _selectedRange = DateTimeRange(start: _selectedRange!.start, end: t); }
                                if (widget.onIntervalSelected != null) widget.onIntervalSelected!(_selectedRange!);
                                return;
                              }
                              if (_dragging) {
                                _dragCurrent = ev.localPosition;
                              }
                            });
                          },
                          onPointerUp: (ev) {
                            setState(() {
                              if (_resizingStart || _resizingEnd) { _resizingStart = false; _resizingEnd = false; return; }
                              if (_dragging) {
                                _dragging = false;
                                final range = _selectionRange(_pxPerMs(_viewEnd.difference(_viewStart).inMilliseconds.toDouble()));
                                if (range != null) {
                                  _selectedRange = range;
                                  if (widget.onIntervalSelected != null) widget.onIntervalSelected!(range);
                                }
                                _dragStart = null;
                                _dragCurrent = null;
                              }
                            });
                          },
                          child: GestureDetector(
                            behavior: HitTestBehavior.opaque,
                            onDoubleTap: () { setState(() { _selectedRange = null; }); },
                            child: Stack(
                              children: [
                                Positioned(
                                   left: 0,
                                   right: 0,
                                  top: -vOff,
                                  height: widget.padding.top + _axisHeight,
                                  child: CustomPaint(
                                    painter: _GridPainter(
                                      start: _viewStart,
                                      end: _viewEnd,
                                      pxPerMs: pxPerMs,
                                      axisHeight: _axisHeight,
                                      padding: widget.padding,
                                      colorScheme: Theme.of(context).colorScheme,
                                      textStyle: Theme.of(context)
                                          .textTheme
                                          .labelSmall
                                          ?.copyWith(
                                              color: context
                                                  .appColors
                                                  .textSecondary),
                                      showMilliseconds:
                                          _viewEnd.difference(_viewStart).inMilliseconds <=
                                              5000,
                                    ),
                                  ),
                                ),
                                Positioned.fill(
                                  child: CustomPaint(
                                    painter: _GridPainter(
                                      start: _viewStart,
                                      end: _viewEnd,
                                      pxPerMs: pxPerMs,
                                      axisHeight: _axisHeight,
                                      padding: widget.padding,
                                      colorScheme: Theme.of(context).colorScheme,
                                      textStyle: Theme.of(context)
                                          .textTheme
                                          .labelSmall
                                          ?.copyWith(
                                              color: context
                                                  .appColors
                                                  .textSecondary),
                                      showMilliseconds:
                                          _viewEnd.difference(_viewStart).inMilliseconds <=
                                              5000,
                                    ),
                                  ),
                                ),
                                ...items.map((it) {
                                  final left =
                                      widget.padding.left + (it.startMs * pxPerMs);
                                  final width =
                                      math.max(2.0, (it.durationMs * pxPerMs));
                                  final top = widget.padding.top +
                                      _axisHeight +
                                      it.lane * (_laneHeight + _laneGap);
                                  final isHover = _hoverSessionId == it.session.id;
                                  final baseColor = _barColor(context, it.session);
                                  final method = (it.session.httpMeta?['method'] ?? '').toString();
                                  final tuned = _applyMethodSaturation(baseColor, method);
                                  final pulse = (it.session.kind == 'ws' && it.session.closedAt == null) ? (0.6 + 0.4 * _pulse.value) : 1.0;
                                  final baseAlpha = isHover ? 0.95 : 0.75;
                                  final color = tuned.withOpacity((baseAlpha * pulse).clamp(0.0, 1.0));
                                  final borderColor = baseColor;
                                  final tooltip = _sessionLabel(it.session);
                                  return Positioned(
                                    left: left,
                                    top: top,
                                    width: width,
                                    height: _laneHeight,
                                    child: MouseRegion(
                                      onEnter: (_) => setState(() {
                                        _hoverSessionId = it.session.id;
                                      }),
                                      onExit: (_) => setState(() {
                                        _hoverSessionId = null;
                                      }),
                                      cursor: SystemMouseCursors.click,
                                      child: Tooltip(
                                        message: tooltip,
                                        waitDuration:
                                            const Duration(milliseconds: 300),
                                        child: GestureDetector(
                                          behavior: HitTestBehavior.opaque,
                                          onDoubleTap: () {},
                                          onTap: () {
                                            widget.onSessionSelected?.call(it.session);
                                          },
                                          child: Container(
                                            decoration: BoxDecoration(
                                              color: color,
                                              borderRadius:
                                                  BorderRadius.circular(4),
                                              border: Border.all(
                                                  color: borderColor, width: 1),
                                            ),
                                            child: width >= 80
                                                ? Padding(
                                                    padding: const EdgeInsets
                                                        .symmetric(
                                                        horizontal: 4),
                                                    child: Align(
                                                      alignment:
                                                          Alignment.centerLeft,
                                                      child: Text(
                                                        _barLabel(it.session),
                                                        overflow:
                                                            TextOverflow.fade,
                                                        softWrap: false,
                                                        style: Theme.of(context)
                                                            .textTheme
                                                            .labelSmall
                                                            ?.copyWith(
                                                              color: ThemeData
                                                                          .estimateBrightnessForColor(
                                                                              color) ==
                                                                      Brightness
                                                                          .dark
                                                                  ? Colors.white
                                                                  : Colors.black,
                                                            ),
                                                      ),
                                                    ),
                                                  )
                                                : null,
                                          ),
                                        ),
                                      ),
                                    ),
                                  );
                                }).toList(),
                                if (_selectedRange != null) ...[
                                  Builder(builder: (context){
                                    final pxPer = _pxPerMs(_viewEnd.difference(_viewStart).inMilliseconds.toDouble());
                                    final l = _timeToX(_selectedRange!.start, pxPer);
                                    final r = _timeToX(_selectedRange!.end, pxPer);
                                    final top = widget.padding.top + _axisHeight;
                                    return Stack(children: [
                                      AnimatedPositioned(duration: const Duration(milliseconds: 180), curve: Curves.easeOutCubic, left: l, right: contentWidth + widget.padding.horizontal - r, top: top, bottom: widget.padding.bottom, child: IgnorePointer(child: Container(color: Theme.of(context).colorScheme.primary.withOpacity(0.10)))) ,
                                      AnimatedPositioned(duration: const Duration(milliseconds: 180), curve: Curves.easeOutCubic, left: l - 3, top: top, bottom: widget.padding.bottom, child: MouseRegion(cursor: SystemMouseCursors.resizeColumn, child: Container(width: 6, color: Theme.of(context).colorScheme.primary.withOpacity(0.35)))),
                                      AnimatedPositioned(duration: const Duration(milliseconds: 180), curve: Curves.easeOutCubic, left: r - 3, top: top, bottom: widget.padding.bottom, child: MouseRegion(cursor: SystemMouseCursors.resizeColumn, child: Container(width: 6, color: Theme.of(context).colorScheme.primary.withOpacity(0.35)))),
                                    ]);
                                  }),
                                ],
                                if (_dragging &&
                                    _dragStart != null &&
                                    _dragCurrent != null)
                                  _buildSelectionRect(pxPerMs),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildSelectionRect(double pxPerMs) {
    final start = _dragStart!;
    final cur = _dragCurrent!;
    final l = math.min(start.dx, cur.dx);
    final r = math.max(start.dx, cur.dx);
    final top = widget.padding.top + _axisHeight;
    return Positioned(
      left: l,
      top: top,
      bottom: widget.padding.bottom,
      width: r - l,
      child: IgnorePointer(
        child: Container(
          color: Theme.of(context).colorScheme.primary.withOpacity(0.12),
        ),
      ),
    );
  }

  DateTimeRange? _selectionRange(double pxPerMs) {
    if (_dragStart == null || _dragCurrent == null) return null;
    final l = math.min(_dragStart!.dx, _dragCurrent!.dx) - widget.padding.left;
    final r = math.max(_dragStart!.dx, _dragCurrent!.dx) - widget.padding.left;
    final startMs = (l / pxPerMs).clamp(0, double.infinity);
    final endMs = (r / pxPerMs).clamp(0, double.infinity);
    final s = _viewStart.add(Duration(milliseconds: startMs.round()));
    final e = _viewStart.add(Duration(milliseconds: endMs.round()));
    if (!e.isAfter(s)) return null;
    return DateTimeRange(start: s, end: e);
  }

  String _sessionLabel(Session s) {
    final uri = _safeUri(s.target);
    final String pathStr = (uri != null && uri.path.isNotEmpty) ? uri.path : s.target;
    final host = uri?.host ?? '';
    final method = (s.httpMeta != null ? (s.httpMeta!['method']?.toString() ?? '') : '');
    final kind = s.kind ?? (method.isEmpty ? 'ws' : 'http');
    final durMs = _durationOf(s).inMilliseconds;
    final dur = '${durMs}ms';
    if (method.isNotEmpty) {
      return '$method $host$pathStr — $dur';
    }
    return '${kind.toUpperCase()} $host$pathStr — $dur';
  }

  Uri? _safeUri(String v) { try { return Uri.parse(v); } catch (_) { return null; } }

  Duration _durationOf(Session s) {
    final start = s.startedAt ?? DateTime.now();
    final end = s.closedAt ?? DateTime.now();
    final d = end.difference(start);
    if (d.inMilliseconds <= 0) return const Duration(milliseconds: 1);
    return d;
  }

  List<_LayoutItem> _computeLayoutItems(List<Session> sessions) {
    final items = <_LayoutItem>[];
    final sorted = sessions.where((s) => s.startedAt != null).toList()
      ..sort((a, b) => a.startedAt!.compareTo(b.startedAt!));

    final laneEnds = <DateTime>[];

    for (final s in sorted) {
      final start = s.startedAt!;
      final end = s.closedAt ?? start.add(const Duration(milliseconds: 1));
      // render only visible (overlapping viewport)
      if (!end.isAfter(_viewStart) || !start.isBefore(_viewEnd)) {
        continue;
      }
      int lane = 0;
      bool placed = false;
      for (var i = 0; i < laneEnds.length; i++) {
        if (!start.isBefore(laneEnds[i])) { lane = i; laneEnds[i] = end; placed = true; break; }
      }
      if (!placed) { lane = laneEnds.length; laneEnds.add(end); }

      final startMs = start.difference(_viewStart).inMilliseconds.toDouble();
      final endMs = end.difference(_viewStart).inMilliseconds.toDouble();
      final durationMs = math.max(1.0, endMs - startMs);
      items.add(_LayoutItem(session: s, lane: lane, startMs: startMs, durationMs: durationMs));
    }
    return items;
  }

  void _zoom(double factor) {
    setState(() {
      final span = _viewEnd.difference(_viewStart);
      final centerMs = _viewStart.millisecondsSinceEpoch + span.inMilliseconds / 2;
      final newSpanMs = (span.inMilliseconds * factor).clamp(200.0, 6 * 60 * 1000.0);
      final newStartMs = (centerMs - newSpanMs / 2).round();
      final newEndMs = (centerMs + newSpanMs / 2).round();
      _viewStart = DateTime.fromMillisecondsSinceEpoch(newStartMs);
      _viewEnd = DateTime.fromMillisecondsSinceEpoch(newEndMs);
    });
  }

  void _pan(double portion) {
    setState(() {
      final spanMs = _viewEnd.difference(_viewStart).inMilliseconds;
      final delta = (spanMs * portion).round();
      _viewStart = _viewStart.add(Duration(milliseconds: delta));
      _viewEnd = _viewEnd.add(Duration(milliseconds: delta));
    });
  }

  Color _barColor(BuildContext context, Session s) {
    final cs = Theme.of(context).colorScheme;
    final status = int.tryParse((s.httpMeta?['status'] ?? '').toString()) ?? 0;
    if (s.kind == 'ws' || (s.kind == null && (s.httpMeta?['method'] ?? '').toString().isEmpty)) {
      return cs.primary;
    }
    if (status >= 500) return cs.error;
    if (status >= 400) return cs.tertiary;
    if (status >= 300) return cs.primary;
    if (status >= 200) return Colors.green;
    return cs.surfaceTint;
  }

  Color _applyMethodSaturation(Color base, String method) {
    final m = method.toUpperCase();
    double factor = 1.0;
    if (m == 'GET') factor = 0.85;
    else if (m == 'POST') factor = 1.0;
    else if (m == 'PUT') factor = 0.95;
    else if (m == 'PATCH') factor = 0.95;
    else if (m == 'DELETE') factor = 1.1;
    final hsl = HSLColor.fromColor(base);
    final newSat = (hsl.saturation * factor).clamp(0.0, 1.0);
    return hsl.withSaturation(newSat).toColor();
  }

  String _barLabel(Session s) {
    final method = (s.httpMeta?['method'] ?? '').toString();
    final status = (s.httpMeta?['status'] ?? '').toString();
    if (method.isNotEmpty) {
      return status.isNotEmpty ? '$method $status' : method;
    }
    final durMs = _durationOf(s).inMilliseconds;
    return '${durMs}ms';
  }
}

class _LayoutItem {
  _LayoutItem({required this.session, required this.lane, required this.startMs, required this.durationMs});
  final Session session;
  final int lane;
  final double startMs;
  final double durationMs;
}

class _GridPainter extends CustomPainter {
  _GridPainter({
    required this.start,
    required this.end,
    required this.pxPerMs,
    required this.axisHeight,
    required this.padding,
    required this.colorScheme,
    required this.textStyle,
    this.showMilliseconds = false,
  });
  final DateTime start;
  final DateTime end;
  final double pxPerMs;
  final double axisHeight;
  final EdgeInsets padding;
  final ColorScheme colorScheme;
  final TextStyle? textStyle;
  final bool showMilliseconds;

  @override
  void paint(Canvas canvas, Size size) {
    final left = padding.left;
    final top = padding.top;
    final axisBottom = top + axisHeight;
    final totalMs = end.difference(start).inMilliseconds;

    final gridPaint = Paint()
      ..color = colorScheme.outlineVariant.withOpacity(0.5)
      ..strokeWidth = 1;

    final minor = _tickStep(totalMs);
    final major = minor * 5;

    // grid lines
    for (int ms = 0; ms <= totalMs; ms += minor) {
      final dx = left + ms * pxPerMs;
      final isMajor = ms % major == 0;
      final p = Offset(dx, axisBottom);
      final p2 = Offset(dx, size.height - padding.bottom);
      canvas.drawLine(p, p2, gridPaint..color = colorScheme.outlineVariant.withOpacity(isMajor ? 0.7 : 0.25));
      if (isMajor) {
        final t = _formatTick(start.add(Duration(milliseconds: ms)));
        final tp = TextPainter(
          text: TextSpan(text: t, style: textStyle),
          textDirection: TextDirection.ltr,
        )..layout();
        tp.paint(canvas, Offset(dx + 2, top + 2));
      }
    }

    // axis baseline
    final axisPaint = Paint()
      ..color = colorScheme.outline
      ..strokeWidth = 1.5;
    canvas.drawLine(Offset(left, axisBottom), Offset(size.width - padding.right, axisBottom), axisPaint);
  }

  int _tickStep(int totalMs) {
    // choose step based on span
    if (totalMs <= 5000) return 250; // 0.25s
    if (totalMs <= 15000) return 500; // 0.5s
    if (totalMs <= 60000) return 1000; // 1s
    if (totalMs <= 5 * 60000) return 5000; // 5s
    return 10000; // 10s
  }

  String _formatTick(DateTime t) {
    String two(int v) => v < 10 ? '0$v' : '$v';
    if (showMilliseconds) {
      final ms = t.millisecond.toString().padLeft(3, '0');
      return '${two(t.hour)}:${two(t.minute)}:${two(t.second)}.$ms';
    }
    return '${two(t.hour)}:${two(t.minute)}:${two(t.second)}';
  }

  @override
  bool shouldRepaint(covariant _GridPainter old) {
    return old.start != start || old.end != end || old.pxPerMs != pxPerMs || old.textStyle != textStyle;
  }
}


