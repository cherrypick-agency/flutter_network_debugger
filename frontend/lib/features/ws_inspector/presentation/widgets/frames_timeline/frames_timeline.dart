import 'package:flutter/material.dart';
import 'dart:math' as math;
import '../../../../../theme/context_ext.dart';
import 'frames_timeline_painter.dart';
import 'frames_timeline_highlight_painter.dart';

class FramesTimeline extends StatelessWidget {
  const FramesTimeline({
    super.key,
    required this.frames,
    this.height = 110,
    this.padding = const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
    this.onFrameTap,
    this.onBrushChanged,
    this.onFrameHover,
  });

  // Ожидается список Map: { id, ts(ISO), direction('client->upstream'|'upstream->client'), opcode, size }
  final List<Map<String, dynamic>> frames;
  final double height;
  final EdgeInsets padding;
  final void Function(String frameId)? onFrameTap;
  final void Function(DateTimeRange? range)? onBrushChanged;
  final void Function(String frameId)? onFrameHover;

  @override
  Widget build(BuildContext context) {
    // Находим диапазон времени
    final tsList = <DateTime>[];
    for (final f in frames) {
      final tsStr = (f['ts'] ?? '').toString();
      try { tsList.add(DateTime.parse(tsStr)); } catch (_) {}
    }
    DateTime minTs = tsList.isEmpty ? DateTime.now().subtract(const Duration(seconds: 1)) : tsList.reduce((a, b) => a.isBefore(b) ? a : b);
    DateTime maxTs = tsList.isEmpty ? DateTime.now().add(const Duration(seconds: 1)) : tsList.reduce((a, b) => a.isAfter(b) ? a : b);
    if (!maxTs.isAfter(minTs)) {
      maxTs = minTs.add(const Duration(seconds: 2));
    }

    final colors = context.appColors;
    final style = FramesTimelineStyle(
      axisColor: colors.border,
      textColor: colors.primary,
      binaryColor: colors.warning,
      pingColor: colors.warning,
      pongColor: colors.textSecondary, // Pong серым
      closeColor: colors.danger,
      msgsColor: colors.primary,
      bytesColor: colors.success,
    );

    return LayoutBuilder(builder: (context, constraints) {
      return SizedBox(
        height: height,
        width: double.infinity,
        child: _InteractiveLayer(
          frames: frames,
          minTs: minTs,
          maxTs: maxTs,
          padding: padding,
          painter: FramesTimelinePainter(
            frames: frames,
            minTs: minTs,
            maxTs: maxTs,
            style: style,
            padding: padding,
          ),
          onTapFrame: onFrameTap,
          onBrushChanged: onBrushChanged,
          onHoverFrame: onFrameHover,
        ),
      );
    });
  }
}

class _InteractiveLayer extends StatefulWidget {
  const _InteractiveLayer({
    required this.frames,
    required this.minTs,
    required this.maxTs,
    required this.padding,
    required this.painter,
    this.onTapFrame,
    this.onBrushChanged,
    this.onHoverFrame,
  });

  final List<Map<String, dynamic>> frames;
  final DateTime minTs;
  final DateTime maxTs;
  final EdgeInsets padding;
  final CustomPainter painter;
  final void Function(String frameId)? onTapFrame;
  final void Function(DateTimeRange? range)? onBrushChanged;
  final void Function(String frameId)? onHoverFrame;

  @override
  State<_InteractiveLayer> createState() => _InteractiveLayerState();
}

class _InteractiveLayerState extends State<_InteractiveLayer> {
  Map<String, dynamic>? _hoverFrame;
  Offset? _hoverPos;
  double? _brushStartX;
  double? _brushEndX;
  String? _lastHoverId;
  final GlobalKey<TooltipState> _tooltipKey = GlobalKey<TooltipState>();
  Offset? _tooltipAnchor;

  double _tsToX(DateTime ts, double width) {
    final total = widget.maxTs.difference(widget.minTs).inMilliseconds;
    final dx = ts.difference(widget.minTs).inMilliseconds;
    if (total <= 0) return widget.padding.left;
    final innerW = width - widget.padding.horizontal;
    final t = dx.clamp(0, total) / total;
    return widget.padding.left + innerW * t;
  }

  DateTime _xToTs(double x, double width) {
    final innerW = width - widget.padding.horizontal;
    final clamped = (x - widget.padding.left).clamp(0.0, innerW);
    final total = widget.maxTs.difference(widget.minTs).inMilliseconds;
    final t = innerW <= 0 ? 0.0 : (clamped / innerW);
    final ms = (total * t).toInt();
    return widget.minTs.add(Duration(milliseconds: ms));
  }

  void _handleTap(Offset pos, double width) {
    String? bestId;
    Map<String, dynamic>? bestFrame;
    (bestId, bestFrame) = _findNearest(pos, width);
    setState(() {
      _hoverFrame = bestFrame;
      _hoverPos = pos;
    });
    if (bestId != null && widget.onTapFrame != null && _isClickWithoutBrush()) {
      widget.onTapFrame!(bestId);
    }
  }

  (String?, Map<String, dynamic>?) _findNearest(Offset pos, double width) {
    String? bestId;
    double bestDist = double.infinity;
    Map<String, dynamic>? bestFrame;
    for (var i = 0; i < widget.frames.length; i++) {
      final f = widget.frames[i];
      final tsStr = (f['ts'] ?? '').toString();
      DateTime? ts;
      try { ts = DateTime.parse(tsStr); } catch (_) {}
      if (ts == null) continue;
      final x = _tsToX(ts, width);
      final dx = (x - pos.dx).abs();
      if (dx < bestDist) {
        bestDist = dx;
        bestId = (f['id'] ?? '${ts.toIso8601String()}/$i').toString();
        bestFrame = f;
      }
    }
    return (bestId, bestFrame);
  }

  bool _isClickWithoutBrush() {
    if (_brushStartX == null && _brushEndX == null) return true;
    if (_brushStartX != null && _brushEndX != null) {
      final dx = (_brushEndX! - _brushStartX!).abs();
      return dx < 3;
    }
    return true;
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(builder: (context, c) {
      return MouseRegion(
        onHover: (e) {
          final r = _findNearest(e.localPosition, c.maxWidth);
          setState(() {
            _hoverPos = e.localPosition;
            _hoverFrame = r.$2;
            if (_hoverFrame != null) {
              final centerY = c.maxHeight / 2;
              final laneOffset = math.min(12.0, (c.maxHeight - widget.padding.vertical) / 4);
              final dir = (_hoverFrame!['direction'] ?? '').toString();
              final y = dir == 'upstream->client' ? (centerY - laneOffset) : (centerY + laneOffset);
              _tooltipAnchor = Offset(e.localPosition.dx, (y - 12).clamp(0.0, c.maxHeight));
            } else {
              _tooltipAnchor = null;
            }
          });
          if (widget.onHoverFrame != null && r.$1 != null && r.$1 != _lastHoverId) {
            _lastHoverId = r.$1;
            widget.onHoverFrame!(r.$1!);
          }
          if (_tooltipKey.currentState != null) {
            _tooltipKey.currentState!.ensureTooltipVisible();
          }
        },
        onExit: (_) { setState(() { _hoverFrame = null; _hoverPos = null; _tooltipAnchor = null; }); },
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTapDown: (d) => _handleTap(d.localPosition, c.maxWidth),
          onPanStart: (d) {
            setState(() {
              _brushStartX = d.localPosition.dx;
              _brushEndX = d.localPosition.dx;
            });
          },
          onPanUpdate: (d) {
            setState(() { _brushEndX = d.localPosition.dx; });
          },
          onPanEnd: (_) {
            if (widget.onBrushChanged != null && _brushStartX != null && _brushEndX != null) {
              final start = _xToTs(_brushStartX!, c.maxWidth);
              final end = _xToTs(_brushEndX!, c.maxWidth);
              final range = start.isBefore(end)
                  ? DateTimeRange(start: start, end: end)
                  : DateTimeRange(start: end, end: start);
              widget.onBrushChanged!(range);
            }
          },
          onDoubleTap: () {
            setState(() { _brushStartX = null; _brushEndX = null; });
            if (widget.onBrushChanged != null) { widget.onBrushChanged!(null); }
          },
          child: Stack(
            fit: StackFit.expand,
            children: [
              Positioned.fill(child: CustomPaint(painter: widget.painter)),
              if (_hoverFrame != null)
                Positioned.fill(
                  child: CustomPaint(
                    painter: FramesTimelineHighlightPainter(
                      hoverFrame: _hoverFrame,
                      minTs: widget.minTs,
                      maxTs: widget.maxTs,
                      padding: widget.padding,
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              if (_hoverFrame != null && _tooltipAnchor != null)
                Positioned(
                  left: _tooltipAnchor!.dx,
                  top: _tooltipAnchor!.dy,
                  child: Tooltip(
                    key: _tooltipKey,
                    triggerMode: TooltipTriggerMode.manual,
                    waitDuration: Duration.zero,
                    showDuration: const Duration(milliseconds: 800),
                    preferBelow: false,
                    verticalOffset: 0,
                    message: _buildTooltipMessage(_hoverFrame!),
                    child: const SizedBox(width: 1, height: 1),
                  ),
                ),
              if (_brushStartX != null && _brushEndX != null)
                Positioned(
                  left: math.min(_brushStartX!, _brushEndX!),
                  top: 0,
                  width: (_brushEndX! - _brushStartX!).abs(),
                  height: c.maxHeight,
                  child: Container(color: Theme.of(context).colorScheme.primary.withOpacity(0.12)),
                ),
              if (_hoverFrame != null && _hoverPos != null)
                const SizedBox.shrink(),
            ],
          ),
        ),
      );
    });
  }

  String _buildTooltipMessage(Map<String, dynamic> f) {
    final opcode = (f['opcode'] ?? '').toString();
    final dir = (f['direction'] ?? '').toString();
    final size = (f['size'] ?? '').toString();
    final ts = (f['ts'] ?? '').toString();
    final preview = (f['preview'] ?? '').toString();
    final sb = StringBuffer();
    sb.write('$dir  ·  $opcode  ·  ${size}B\n');
    sb.write(ts);
    if (preview.isNotEmpty) {
      sb.write('\n');
      sb.write(preview.length > 120 ? preview.substring(0, 120) + '…' : preview);
    }
    return sb.toString();
  }
}


