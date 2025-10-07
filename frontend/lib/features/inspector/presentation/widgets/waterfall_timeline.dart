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
    this.expandToParent = false,
    this.autoCompressLanes,
    this.padding = const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
    this.initialViewportPadding = const Duration(seconds: 2),
    this.initialRange,
    this.fitAll,
    this.onFitAllChanged,
    this.onIntervalCleared,
    this.hoveredSessionIdExt,
    this.selectedSessionIdExt,
  });

  final List<Session> sessions;
  final ValueChanged<DateTimeRange>? onIntervalSelected;
  final ValueChanged<Session>? onSessionSelected;
  final bool autoExtendViewport;
  final double height;
  final bool expandToParent;

  /// Если true — дорожки будут автоматически ужиматься по высоте,
  /// чтобы все сессии помещались в доступную высоту контейнера.
  /// По умолчанию поведение наследуется от `expandToParent` для обратной совместимости.
  final bool? autoCompressLanes;
  final EdgeInsets padding;
  final Duration initialViewportPadding;
  final DateTimeRange? initialRange;
  final bool? fitAll;
  final ValueChanged<bool>? onFitAllChanged;
  // Колбэк, который дёргаем при сбросе выделения (двойной клик по таймлайну)
  final VoidCallback? onIntervalCleared;
  // Внешняя подсветка по наведению/выбору элемента в списке сессий
  final String? hoveredSessionIdExt;
  final String? selectedSessionIdExt;

  @override
  State<WaterfallTimeline> createState() => _WaterfallTimelineState();
}

class _WaterfallTimelineState extends State<WaterfallTimeline>
    with SingleTickerProviderStateMixin {
  final ScrollController _hCtrl = ScrollController();
  final ScrollController _vCtrl = ScrollController();

  late DateTime _viewStart;
  late DateTime _viewEnd;
  late DateTime _axisStart;
  bool _followLive = true;
  bool _autoFitAll = true;

  final double _laneHeight = 18;
  static const double _maxLaneHeight = 48;
  final double _laneGap = 6;
  final double _axisHeight = 22;

  // selection
  bool _dragging = false;
  Offset? _dragStart;
  Offset? _dragCurrent;
  bool _resizingStart = false;
  bool _resizingEnd = false;
  bool _suppressSelectionOnce = false;

  String? _hoverSessionId;
  DateTimeRange? _selectedRange;
  late final AnimationController _pulse;
  // zoom state
  DateTime? _scaleAnchor;

  bool get _isFitAll =>
      widget.onFitAllChanged != null ? (widget.fitAll ?? false) : _autoFitAll;
  void _setFitAll(bool v) {
    if (widget.onFitAllChanged != null) {
      widget.onFitAllChanged!(v);
    } else {
      _autoFitAll = v;
    }
  }

  @override
  void initState() {
    super.initState();
    _vCtrl.addListener(() {
      if (mounted) setState(() {});
    });
    _pulse =
        AnimationController(
            vsync: this,
            duration: const Duration(milliseconds: 220),
          )
          ..repeat(reverse: true)
          ..addListener(() {
            if (mounted) setState(() {});
          });
    final now = DateTime.now();
    if (widget.initialRange != null &&
        widget.initialRange!.end.isAfter(widget.initialRange!.start)) {
      _viewStart = widget.initialRange!.start;
      _viewEnd = widget.initialRange!.end;
    } else {
      final times = _timesOf(widget.sessions);
      final start = (times.start ?? now);
      final end = (times.end ?? now);
      _viewStart = start;
      _viewEnd =
          end.isAfter(_viewStart.add(const Duration(milliseconds: 100)))
              ? end
              : _viewStart.add(const Duration(seconds: 10));
    }
    _axisStart = _viewStart;
  }

  @override
  void didUpdateWidget(covariant WaterfallTimeline oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.autoExtendViewport && widget.sessions.isNotEmpty) {
      final times = _timesOf(widget.sessions);
      final s = times.start;
      final e = times.end;
      if (_isFitAll) {
        if (s != null) _viewStart = s;
        if (e != null) _viewEnd = e;
      } else {
        if (s != null && s.isBefore(_viewStart)) {
          _viewStart = s;
        }
        if (e != null && e.isAfter(_viewEnd)) {
          final span = _viewEnd.difference(_viewStart);
          if (_followLive) {
            _viewEnd = e;
            _viewStart = _viewEnd.subtract(span);
            // auto-scroll to rightmost
            if (_hCtrl.hasClients) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                if (!_hCtrl.hasClients) {
                  return;
                }
                _hCtrl.animateTo(
                  _hCtrl.position.maxScrollExtent,
                  duration: const Duration(milliseconds: 120),
                  curve: Curves.easeOut,
                );
              });
            }
          } else {
            _viewEnd = e;
          }
        }
      }
      if (!_viewEnd.isAfter(_viewStart)) {
        _viewEnd = _viewStart.add(const Duration(seconds: 10));
      }
      _clampViewport();
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
      // Для активных сессий берём текущий момент как правую границу,
      // чтобы таймлайн имел актуальный dataEnd и полоса росла анимированно
      final en = it.closedAt ?? (st != null ? DateTime.now() : null);
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
    final totalMs = _viewEnd
        .difference(_viewStart)
        .inMilliseconds
        .toDouble()
        .clamp(1000.0, 3600 * 1000.0);
    final width = math.max(
      1600.0,
      math.min(5000.0, totalMs / 8),
    ); // ~8 ms per px by default
    return width / totalMs;
  }

  void _clampViewport() {
    final bounds = _timesOf(widget.sessions);
    final ds = bounds.start ?? _viewStart;
    final de = bounds.end ?? _viewEnd;
    final maxEnd = de;
    // shift window inside [ds, maxEnd] while preserving span
    var start = _viewStart;
    var end = _viewEnd;
    final minSpanMs = 200;
    if (end.isBefore(start.add(const Duration(milliseconds: 1)))) {
      end = start.add(Duration(milliseconds: minSpanMs));
    }
    if (start.isBefore(ds)) {
      final delta = ds.difference(start);
      start = ds;
      end = end.add(delta);
    }
    if (end.isAfter(maxEnd)) {
      final delta = end.difference(maxEnd);
      end = maxEnd;
      start = start.subtract(delta);
    }
    // ensure min span remains and within bounds
    if (!end.isAfter(start)) {
      end = start.add(Duration(milliseconds: minSpanMs));
      if (end.isAfter(maxEnd)) {
        end = maxEnd;
        start =
            (end.subtract(Duration(milliseconds: minSpanMs)).isBefore(ds))
                ? ds
                : end.subtract(Duration(milliseconds: minSpanMs));
      }
    }
    _viewStart = start;
    _viewEnd = end;
  }

  double _timeToX(DateTime t, double pxPerMs) {
    final ms = t.difference(_axisStart).inMilliseconds.toDouble();
    return widget.padding.left + ms * pxPerMs;
  }

  DateTime _xToTime(double x, double pxPerMs) {
    final rel = (x - widget.padding.left) / pxPerMs;
    final ms = rel.clamp(0, double.infinity).round();
    return _axisStart.add(Duration(milliseconds: ms));
  }

  @override
  Widget build(BuildContext context) {
    return CallbackShortcuts(
      bindings: <ShortcutActivator, VoidCallback>{
        SingleActivator(LogicalKeyboardKey.equal): () => _zoom(0.8),
        SingleActivator(LogicalKeyboardKey.minus): () => _zoom(1.25),
        SingleActivator(LogicalKeyboardKey.arrowLeft): () => _pan(-0.2),
        SingleActivator(LogicalKeyboardKey.arrowRight): () => _pan(0.2),
        SingleActivator(LogicalKeyboardKey.escape):
            () => setState(() {
              _dragStart = null;
              _dragCurrent = null;
            }),
      },
      child: Focus(
        autofocus: false,
        child: SizedBox(
          height: widget.expandToParent ? null : widget.height,
          child: LayoutBuilder(
            builder: (context, constraints) {
              final dataBounds = _timesOf(widget.sessions);
              final dataStart = dataBounds.start ?? _viewStart;
              final dataEnd = dataBounds.end ?? _viewEnd;
              final fullMs = (dataEnd
                      .difference(dataStart)
                      .inMilliseconds
                      .toDouble())
                  .clamp(1.0, 86400000.0);
              final minContent = math.max(
                constraints.maxWidth - widget.padding.horizontal,
                800.0,
              );
              const double _fitRightTailPx = 12.0;
              final double fitWidthPx = (minContent - _fitRightTailPx).clamp(
                1.0,
                minContent,
              );
              // fitAll всегда уважаем, даже при активном прогрессе
              final bool useFitAll = _isFitAll;
              final double pxPerMs =
                  useFitAll
                      ? (fullMs > 0 ? (fitWidthPx / fullMs) : 1.0)
                      : _pxPerMs(
                        _viewEnd
                            .difference(_viewStart)
                            .inMilliseconds
                            .toDouble(),
                      );

              final items = _computeLayoutItems(widget.sessions);
              final lanes =
                  items.isEmpty
                      ? 1
                      : (items.map((e) => e.lane).reduce(math.max) + 1);
              // динамическая высота дорожек: чем выше сам таймлайн, тем выше полосы
              // если включено авто-сжатие (autoCompressLanes == true) — подгоняем под доступную высоту
              double laneHeight = _laneHeight;
              final bool compressLanes =
                  (widget.autoCompressLanes ?? widget.expandToParent);
              if (compressLanes && constraints.hasBoundedHeight && lanes > 0) {
                final available = (constraints.maxHeight -
                        widget.padding.vertical -
                        _axisHeight)
                    .clamp(0.0, double.infinity);
                final candidate = (available / lanes) - _laneGap;
                laneHeight =
                    candidate.clamp(_laneHeight, _maxLaneHeight).toDouble();
              }
              final lanesHeight = lanes * (laneHeight + _laneGap);
              final contentHeight =
                  _axisHeight + lanesHeight + widget.padding.vertical;
              final vOff = _vCtrl.hasClients ? _vCtrl.offset : 0.0;
              final contentWidth =
                  useFitAll
                      ? minContent
                      : (fullMs * pxPerMs).clamp(minContent, 2000000.0);
              // Keep mapping origin consistent with painting
              _axisStart = dataStart;

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
                        height: math.max(
                          (widget.expandToParent
                              ? (constraints.hasBoundedHeight
                                  ? constraints.maxHeight
                                  : widget.height)
                              : widget.height),
                          contentHeight,
                        ),
                        child: GestureDetector(
                          onScaleStart: (d) {
                            // Не отключаем Fit All на старте жеста; только когда реально меняется scale
                            final dx = d.localFocalPoint.dx;
                            final pxPer =
                                pxPerMs; // согласованный мэппинг как при рендере
                            _scaleAnchor = _xToTime(dx, pxPer);
                          },
                          onScaleUpdate: (d) {
                            if (_scaleAnchor == null) return;
                            if (d.scale == 1.0) return;
                            // Отключаем Fit All только при реальном зуме
                            _setFitAll(false);
                            final raw = (1 / d.scale);
                            final spanMs =
                                _viewEnd.difference(_viewStart).inMilliseconds;
                            int div;
                            if (spanMs <= 2000)
                              div = 8; // супер-мягкий при близком зуме
                            else if (spanMs <= 5000)
                              div = 6;
                            else if (spanMs <= 15000)
                              div = 5;
                            else if (spanMs <= 60000)
                              div = 4;
                            else
                              div = 3;
                            final factor = 1 + (raw - 1) / div;
                            _applyZoomAround(_scaleAnchor!, factor);
                          },
                          onScaleEnd: (_) {
                            _scaleAnchor = null;
                          },
                          child: Listener(
                            onPointerDown: (ev) {
                              if (ev.kind == PointerDeviceKind.mouse &&
                                  ev.buttons != kPrimaryMouseButton)
                                return;
                              if (_suppressSelectionOnce) {
                                _suppressSelectionOnce = false;
                                return;
                              }
                              setState(() {
                                _resizingStart = false;
                                _resizingEnd = false;
                                final pxPer =
                                    pxPerMs; // использовать ту же шкалу, что и рендер
                                if (_selectedRange != null) {
                                  final sx = _timeToX(
                                    _selectedRange!.start,
                                    pxPer,
                                  );
                                  final ex = _timeToX(
                                    _selectedRange!.end,
                                    pxPer,
                                  );
                                  if ((ev.localPosition.dx - sx).abs() <= 6) {
                                    _resizingStart = true;
                                  }
                                  if ((ev.localPosition.dx - ex).abs() <= 6) {
                                    _resizingEnd = true;
                                  }
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
                                final pxPer =
                                    pxPerMs; // согласованная конверсия
                                if (_resizingStart && _selectedRange != null) {
                                  final t = _xToTime(
                                    ev.localPosition.dx,
                                    pxPer,
                                  );
                                  if (t.isBefore(_selectedRange!.end)) {
                                    _selectedRange = DateTimeRange(
                                      start: t,
                                      end: _selectedRange!.end,
                                    );
                                  }
                                  if (widget.onIntervalSelected != null)
                                    widget.onIntervalSelected!(_selectedRange!);
                                  return;
                                }
                                if (_resizingEnd && _selectedRange != null) {
                                  final t = _xToTime(
                                    ev.localPosition.dx,
                                    pxPer,
                                  );
                                  if (t.isAfter(_selectedRange!.start)) {
                                    _selectedRange = DateTimeRange(
                                      start: _selectedRange!.start,
                                      end: t,
                                    );
                                  }
                                  if (widget.onIntervalSelected != null)
                                    widget.onIntervalSelected!(_selectedRange!);
                                  return;
                                }
                                if (_dragging) {
                                  _dragCurrent = ev.localPosition;
                                }
                              });
                            },
                            onPointerUp: (ev) {
                              if (_suppressSelectionOnce) {
                                _suppressSelectionOnce = false;
                                return;
                              }
                              setState(() {
                                if (_resizingStart || _resizingEnd) {
                                  _resizingStart = false;
                                  _resizingEnd = false;
                                  return;
                                }
                                if (_dragging) {
                                  _dragging = false;
                                  final range = _selectionRange(pxPerMs);
                                  if (range != null) {
                                    _selectedRange = range;
                                    if (widget.onIntervalSelected != null)
                                      widget.onIntervalSelected!(range);
                                  }
                                  _dragStart = null;
                                  _dragCurrent = null;
                                }
                              });
                            },
                            child: GestureDetector(
                              behavior: HitTestBehavior.opaque,
                              onTapUp: (d) {
                                // Зум в точку клика по таймлайну (кроме полос сессий — они перехватят событие)
                                _setFitAll(false);
                                final anchor = _xToTime(
                                  d.localPosition.dx,
                                  pxPerMs,
                                );
                                _applyZoomAround(anchor, 0.8); // приближение
                              },
                              onDoubleTap: () {
                                setState(() {
                                  _selectedRange = null;
                                });
                                // Сообщаем наверх, что выделение сброшено,
                                // чтобы внешний фильтр по интервалу тоже очистился
                                widget.onIntervalCleared?.call();
                              },
                              child: Stack(
                                clipBehavior: Clip.hardEdge,
                                children: [
                                  // Sticky axis overlay (painted across full data extent)
                                  Positioned(
                                    left: 0,
                                    right: 0,
                                    top: -vOff,
                                    height: widget.padding.top + _axisHeight,
                                    child: CustomPaint(
                                      painter: _GridPainter(
                                        start: dataStart,
                                        end: dataEnd,
                                        pxPerMs: pxPerMs,
                                        axisHeight: _axisHeight,
                                        padding: widget.padding,
                                        colorScheme:
                                            Theme.of(context).colorScheme,
                                        textStyle: Theme.of(
                                          context,
                                        ).textTheme.labelSmall?.copyWith(
                                          color:
                                              context.appColors.textSecondary,
                                        ),
                                        showMilliseconds:
                                            _viewEnd
                                                .difference(_viewStart)
                                                .inMilliseconds <=
                                            5000,
                                      ),
                                    ),
                                  ),
                                  // Background grid
                                  Positioned.fill(
                                    child: CustomPaint(
                                      painter: _GridPainter(
                                        start: dataStart,
                                        end: dataEnd,
                                        pxPerMs: pxPerMs,
                                        axisHeight: _axisHeight,
                                        padding: widget.padding,
                                        colorScheme:
                                            Theme.of(context).colorScheme,
                                        textStyle: Theme.of(
                                          context,
                                        ).textTheme.labelSmall?.copyWith(
                                          color:
                                              context.appColors.textSecondary,
                                        ),
                                        showMilliseconds:
                                            _viewEnd
                                                .difference(_viewStart)
                                                .inMilliseconds <=
                                            5000,
                                      ),
                                    ),
                                  ),
                                  // Bars
                                  ...items.map((it) {
                                    final ts =
                                        it.session.startedAt ?? dataStart;
                                    final startMs =
                                        ts
                                            .difference(dataStart)
                                            .inMilliseconds
                                            .toDouble();
                                    // Для активных — правая граница = now
                                    final endTs =
                                        it.session.closedAt ?? DateTime.now();
                                    final durMs = (endTs
                                            .difference(ts)
                                            .inMilliseconds
                                            .toDouble())
                                        .clamp(1.0, double.infinity);
                                    double left =
                                        widget.padding.left +
                                        (startMs * pxPerMs);
                                    double width = math.max(
                                      2.0,
                                      (durMs * pxPerMs),
                                    );
                                    // clip to content bounds to prevent overflow on extreme zoom
                                    final double contentLeft =
                                        widget.padding.left;
                                    final double contentRight =
                                        widget.padding.left + contentWidth;
                                    if (left < contentLeft) {
                                      final delta = contentLeft - left;
                                      width -= delta;
                                      left = contentLeft;
                                    }
                                    if (left + width > contentRight) {
                                      width = contentRight - left;
                                    }
                                    if (width <= 0) {
                                      return const SizedBox.shrink();
                                    }
                                    final top =
                                        widget.padding.top +
                                        _axisHeight +
                                        it.lane * (laneHeight + _laneGap);
                                    final isSelectedExt =
                                        (widget.selectedSessionIdExt != null &&
                                            widget.selectedSessionIdExt ==
                                                it.session.id);
                                    final isHoverExt =
                                        (widget.hoveredSessionIdExt != null &&
                                            widget.hoveredSessionIdExt ==
                                                it.session.id);
                                    final isHover =
                                        _hoverSessionId == it.session.id ||
                                        isHoverExt ||
                                        isSelectedExt;
                                    final baseColor = _barColor(
                                      context,
                                      it.session,
                                    );
                                    final method =
                                        (it.session.httpMeta?['method'] ?? '')
                                            .toString();
                                    final tuned = _applyMethodSaturation(
                                      baseColor,
                                      method,
                                    );
                                    final isActive =
                                        it.session.closedAt == null;
                                    final pulse =
                                        isActive
                                            ? (0.6 + 0.4 * _pulse.value)
                                            : 1.0;
                                    final baseAlpha =
                                        isHover
                                            ? (isSelectedExt ? 1.0 : 0.95)
                                            : 0.75;
                                    final color = tuned.withOpacity(
                                      (baseAlpha * pulse).clamp(0.0, 1.0),
                                    );
                                    final borderColor = baseColor;
                                    final tooltip = _sessionLabel(it.session);
                                    return Positioned(
                                      left: left,
                                      top: top,
                                      width: width,
                                      height: laneHeight,
                                      child: Listener(
                                        onPointerDown: (_) {
                                          _suppressSelectionOnce = true;
                                        },
                                        child: MouseRegion(
                                          onEnter:
                                              (_) => setState(() {
                                                _hoverSessionId = it.session.id;
                                              }),
                                          onExit:
                                              (_) => setState(() {
                                                _hoverSessionId = null;
                                              }),
                                          cursor: SystemMouseCursors.click,
                                          child: Tooltip(
                                            message: tooltip,
                                            waitDuration: const Duration(
                                              milliseconds: 300,
                                            ),
                                            child: GestureDetector(
                                              behavior: HitTestBehavior.opaque,
                                              onDoubleTap: () {},
                                              onTap: () {
                                                widget.onSessionSelected?.call(
                                                  it.session,
                                                );
                                              },
                                              child: Stack(
                                                children: [
                                                  // background bar
                                                  Container(
                                                    decoration: BoxDecoration(
                                                      color: color,
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                            4,
                                                          ),
                                                      border: Border.all(
                                                        color: borderColor,
                                                        width:
                                                            isSelectedExt
                                                                ? 2
                                                                : (isHover
                                                                    ? 1.3
                                                                    : 1),
                                                      ),
                                                    ),
                                                  ),
                                                  // label overlay when wide enough
                                                  if (width >= 80)
                                                    Padding(
                                                      padding:
                                                          const EdgeInsets.symmetric(
                                                            horizontal: 4,
                                                          ),
                                                      child: Align(
                                                        alignment:
                                                            Alignment
                                                                .centerLeft,
                                                        child: Text(
                                                          _barLabel(it.session),
                                                          overflow:
                                                              TextOverflow.fade,
                                                          softWrap: false,
                                                          style: Theme.of(
                                                            context,
                                                          ).textTheme.labelSmall?.copyWith(
                                                            color:
                                                                ThemeData.estimateBrightnessForColor(
                                                                          color,
                                                                        ) ==
                                                                        Brightness
                                                                            .dark
                                                                    ? Colors
                                                                        .white
                                                                    : Colors
                                                                        .black,
                                                          ),
                                                        ),
                                                      ),
                                                    ),
                                                  // убрали точку; пульсация самой полосы ускорена выше
                                                ],
                                              ),
                                            ),
                                          ),
                                        ),
                                      ),
                                    );
                                  }).toList(),
                                  // Persistent selection overlay
                                  if (_selectedRange != null) ...[
                                    Builder(
                                      builder: (context) {
                                        final pxPer = pxPerMs;
                                        final l =
                                            widget.padding.left +
                                            (_selectedRange!.start
                                                    .difference(dataStart)
                                                    .inMilliseconds
                                                    .toDouble() *
                                                pxPer);
                                        final r =
                                            widget.padding.left +
                                            (_selectedRange!.end
                                                    .difference(dataStart)
                                                    .inMilliseconds
                                                    .toDouble() *
                                                pxPer);
                                        final top =
                                            widget.padding.top + _axisHeight;
                                        return Stack(
                                          children: [
                                            AnimatedPositioned(
                                              duration: const Duration(
                                                milliseconds: 180,
                                              ),
                                              curve: Curves.easeOutCubic,
                                              left: l,
                                              right:
                                                  contentWidth +
                                                  widget.padding.horizontal -
                                                  r,
                                              top: top,
                                              bottom: widget.padding.bottom,
                                              child: IgnorePointer(
                                                child: Container(
                                                  color: Theme.of(context)
                                                      .colorScheme
                                                      .primary
                                                      .withOpacity(0.10),
                                                ),
                                              ),
                                            ),
                                            AnimatedPositioned(
                                              duration: const Duration(
                                                milliseconds: 180,
                                              ),
                                              curve: Curves.easeOutCubic,
                                              left: l - 3,
                                              top: top,
                                              bottom: widget.padding.bottom,
                                              child: MouseRegion(
                                                cursor:
                                                    SystemMouseCursors
                                                        .resizeColumn,
                                                child: Container(
                                                  width: 6,
                                                  color: Theme.of(context)
                                                      .colorScheme
                                                      .primary
                                                      .withOpacity(0.35),
                                                ),
                                              ),
                                            ),
                                            AnimatedPositioned(
                                              duration: const Duration(
                                                milliseconds: 180,
                                              ),
                                              curve: Curves.easeOutCubic,
                                              left: r - 3,
                                              top: top,
                                              bottom: widget.padding.bottom,
                                              child: MouseRegion(
                                                cursor:
                                                    SystemMouseCursors
                                                        .resizeColumn,
                                                child: Container(
                                                  width: 6,
                                                  color: Theme.of(context)
                                                      .colorScheme
                                                      .primary
                                                      .withOpacity(0.35),
                                                ),
                                              ),
                                            ),
                                          ],
                                        );
                                      },
                                    ),
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
    final s = _axisStart.add(Duration(milliseconds: startMs.round()));
    final e = _axisStart.add(Duration(milliseconds: endMs.round()));
    if (!e.isAfter(s)) return null;
    return DateTimeRange(start: s, end: e);
  }

  String _sessionLabel(Session s) {
    final uri = _safeUri(s.target);
    final String pathStr =
        (uri != null && uri.path.isNotEmpty) ? uri.path : s.target;
    final host = uri?.host ?? '';
    final method =
        (s.httpMeta != null ? (s.httpMeta!['method']?.toString() ?? '') : '');
    final kind = s.kind ?? (method.isEmpty ? 'ws' : 'http');
    final dur = _formatDurationBrief(_durationOf(s));
    if (method.isNotEmpty) {
      return '$method $host$pathStr — $dur';
    }
    return '${kind.toUpperCase()} $host$pathStr — $dur';
  }

  Uri? _safeUri(String v) {
    try {
      return Uri.parse(v);
    } catch (_) {
      return null;
    }
  }

  Duration _durationOf(Session s) {
    final start = s.startedAt ?? DateTime.now();
    final end = s.closedAt ?? DateTime.now();
    final d = end.difference(start);
    if (d.inMilliseconds <= 0) return const Duration(milliseconds: 1);
    return d;
  }

  String _formatDurationBrief(Duration d) {
    final totalMs = d.inMilliseconds;
    if (totalMs < 1000) {
      return '${totalMs}ms';
    }
    final secs = totalMs / 1000.0;
    if (secs < 10) {
      return '${secs.toStringAsFixed(1)}s';
    }
    if (secs < 60) {
      return '${secs.round()}s';
    }
    final totalSeconds = d.inSeconds;
    final minutes = totalSeconds ~/ 60;
    final secondsR = totalSeconds % 60;
    if (minutes < 60) {
      return secondsR > 0 ? '${minutes}m ${secondsR}s' : '${minutes}m';
    }
    final hours = minutes ~/ 60;
    final minutesR = minutes % 60;
    return minutesR > 0 ? '${hours}h ${minutesR}m' : '${hours}h';
  }

  List<_LayoutItem> _computeLayoutItems(List<Session> sessions) {
    final items = <_LayoutItem>[];
    final sorted =
        sessions.where((s) => s.startedAt != null).toList()
          ..sort((a, b) => a.startedAt!.compareTo(b.startedAt!));

    final laneEnds = <DateTime>[];

    for (final s in sorted) {
      final start = s.startedAt!;
      final end = s.closedAt ?? DateTime.now();
      // Do not filter by viewport; keep all items to retain scrollability at any zoom
      int lane = 0;
      bool placed = false;
      for (var i = 0; i < laneEnds.length; i++) {
        final lastEnd = laneEnds[i];
        if (!start.isBefore(lastEnd)) {
          lane = i;
          laneEnds[i] = end;
          placed = true;
          break;
        }
      }
      if (!placed) {
        lane = laneEnds.length;
        laneEnds.add(end);
      }

      final startMs = start.difference(_viewStart).inMilliseconds.toDouble();
      final endMs = end.difference(_viewStart).inMilliseconds.toDouble();
      final durationMs = math.max(1.0, endMs - startMs);
      items.add(
        _LayoutItem(
          session: s,
          lane: lane,
          startMs: startMs,
          durationMs: durationMs,
        ),
      );
    }
    return items;
  }

  void _zoom(double factor) {
    setState(() {
      _setFitAll(false);
      final span = _viewEnd.difference(_viewStart);
      final centerMs =
          _viewStart.millisecondsSinceEpoch + span.inMilliseconds / 2;
      final newSpanMs = (span.inMilliseconds * factor).clamp(
        200.0,
        6 * 60 * 1000.0,
      );
      final newStartMs = (centerMs - newSpanMs / 2).round();
      final newEndMs = (centerMs + newSpanMs / 2).round();
      _viewStart = DateTime.fromMillisecondsSinceEpoch(newStartMs);
      _viewEnd = DateTime.fromMillisecondsSinceEpoch(newEndMs);
      _clampViewport();
    });
  }

  void _pan(double portion) {
    setState(() {
      _setFitAll(false);
      final spanMs = _viewEnd.difference(_viewStart).inMilliseconds;
      final delta = (spanMs * portion).round();
      _viewStart = _viewStart.add(Duration(milliseconds: delta));
      _viewEnd = _viewEnd.add(Duration(milliseconds: delta));
      _clampViewport();
    });
  }

  Color _barColor(BuildContext context, Session s) {
    final cs = Theme.of(context).colorScheme;
    final status = int.tryParse((s.httpMeta?['status'] ?? '').toString()) ?? 0;
    if (s.kind == 'ws' ||
        (s.kind == null && (s.httpMeta?['method'] ?? '').toString().isEmpty)) {
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
    if (m == 'GET')
      factor = 0.85;
    else if (m == 'POST')
      factor = 1.0;
    else if (m == 'PUT')
      factor = 0.95;
    else if (m == 'PATCH')
      factor = 0.95;
    else if (m == 'DELETE')
      factor = 1.1;
    final hsl = HSLColor.fromColor(base);
    final newSat = (hsl.saturation * factor).clamp(0.0, 1.0);
    return hsl.withSaturation(newSat).toColor();
  }

  String _barLabel(Session s) {
    final method = (s.httpMeta?['method'] ?? '').toString();
    final status = (s.httpMeta?['status'] ?? '').toString();
    final dur = _formatDurationBrief(_durationOf(s));
    if (method.isNotEmpty) {
      final left = status.isNotEmpty ? '$method $status' : method;
      return '$left — $dur';
    }
    return dur;
  }

  void _applyZoomAround(DateTime anchor, double factor) {
    setState(() {
      final span = _viewEnd.difference(_viewStart);
      final anchorMs = anchor.millisecondsSinceEpoch.toDouble();
      final startMs = _viewStart.millisecondsSinceEpoch.toDouble();
      final endMs = _viewEnd.millisecondsSinceEpoch.toDouble();
      final anchorT = ((anchorMs - startMs) / (endMs - startMs)).clamp(
        0.0,
        1.0,
      );
      final newSpanMs = (span.inMilliseconds * factor).clamp(
        200.0,
        6 * 60 * 1000.0,
      );
      final newStartMs = anchorMs - anchorT * newSpanMs;
      final newEndMs = newStartMs + newSpanMs;
      _viewStart = DateTime.fromMillisecondsSinceEpoch(newStartMs.round());
      _viewEnd = DateTime.fromMillisecondsSinceEpoch(newEndMs.round());
      _clampViewport();
      // if live, keep rightmost scroll
      if (_followLive && _hCtrl.hasClients) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!_hCtrl.hasClients) {
            return;
          }
          _hCtrl.animateTo(
            _hCtrl.position.maxScrollExtent,
            duration: const Duration(milliseconds: 100),
            curve: Curves.easeOut,
          );
        });
      }
    });
  }
}

class _LayoutItem {
  _LayoutItem({
    required this.session,
    required this.lane,
    required this.startMs,
    required this.durationMs,
  });
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

    final gridPaint =
        Paint()
          ..color = colorScheme.outlineVariant.withOpacity(0.5)
          ..strokeWidth = 1;

    // adaptive grid/labels
    const double minGridPx = 24; // minimal px between grid lines
    const double minLabelPx = 72; // minimal px between labels
    final int minor = _chooseStepMs(totalMs, pxPerMs, minGridPx);
    final int labelStep = _chooseStepMs(totalMs, pxPerMs, minLabelPx);

    // grid lines (minor with emphasis on label ticks)
    for (int ms = 0; ms <= totalMs; ms += minor) {
      final dx = left + ms * pxPerMs;
      final isLabelTick = (ms % labelStep == 0);
      final p = Offset(dx, axisBottom);
      final p2 = Offset(dx, size.height - padding.bottom);
      canvas.drawLine(
        p,
        p2,
        gridPaint
          ..color = colorScheme.outlineVariant.withOpacity(
            isLabelTick ? 0.7 : 0.25,
          ),
      );
    }

    // labels without overlap
    double lastRight = -1e9;
    final usableRight = size.width - padding.right;
    for (int ms = 0; ms <= totalMs; ms += labelStep) {
      final dx = left + ms * pxPerMs;
      final t = _formatTick(start.add(Duration(milliseconds: ms)));
      final tp = TextPainter(
        text: TextSpan(text: t, style: textStyle),
        textDirection: TextDirection.ltr,
      )..layout();
      final x = dx + 2;
      final right = x + tp.width;
      if (x > lastRight + 6 && right <= usableRight) {
        tp.paint(canvas, Offset(x, top + 2));
        lastRight = right;
      }
    }

    // axis baseline
    final axisPaint =
        Paint()
          ..color = colorScheme.outline
          ..strokeWidth = 1.5;
    canvas.drawLine(
      Offset(left, axisBottom),
      Offset(size.width - padding.right, axisBottom),
      axisPaint,
    );
  }

  int _chooseStepMs(int totalMs, double pxPerMs, double minPx) {
    if (totalMs <= 0) return 1;
    const List<int> steps = [
      1,
      2,
      5,
      10,
      20,
      50,
      100,
      200,
      250,
      500,
      1000,
      2000,
      5000,
      10000,
      15000,
      30000,
      60000,
      120000,
      300000,
      600000,
    ];
    for (final s in steps) {
      if (pxPerMs * s >= minPx) return s;
    }
    return steps.last;
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
    return old.start != start ||
        old.end != end ||
        old.pxPerMs != pxPerMs ||
        old.textStyle != textStyle;
  }
}
