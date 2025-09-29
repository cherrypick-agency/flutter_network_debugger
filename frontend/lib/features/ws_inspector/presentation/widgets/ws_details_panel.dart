import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import 'dart:convert';
import 'package:flutter/services.dart';

String _fmtTime(String ts) {
  try {
    final dt = DateTime.parse(ts).toLocal();
    String two(int v) => v < 10 ? '0$v' : '$v';
    return '${two(dt.hour)}:${two(dt.minute)}:${two(dt.second)}';
  } catch (_) {
    return ts;
  }
}

class WsDetailsPanel extends StatefulWidget {
  const WsDetailsPanel({
    super.key,
    required this.frames,
    required this.events,
    required this.opcodeFilter,
    required this.directionFilter,
    required this.namespaceCtrl,
    required this.onChangeOpcode,
    required this.onChangeDirection,
    required this.hideHeartbeats,
    required this.onToggleHeartbeats,
  });
  final List<dynamic> frames;
  final List<dynamic> events;
  final String opcodeFilter;
  final String directionFilter;
  final TextEditingController namespaceCtrl;
  final void Function(String) onChangeOpcode;
  final void Function(String) onChangeDirection;
  final bool hideHeartbeats;
  final void Function(bool) onToggleHeartbeats;

  @override
  State<WsDetailsPanel> createState() => _WsDetailsPanelState();
}

class _WsDetailsPanelState extends State<WsDetailsPanel> {
  bool _pretty = true;
  bool _tree = false;

  bool _frameMatches(Map<String, dynamic> f) {
    if (widget.opcodeFilter != 'all' && (f['opcode']?.toString() ?? '') != widget.opcodeFilter) return false;
    if (widget.directionFilter != 'all' && (f['direction']?.toString() ?? '') != widget.directionFilter) return false;
    return true;
  }

  bool _eventMatches(Map<String, dynamic> e) {
    final ns = (e['namespace'] ?? '').toString();
    final nsFilter = widget.namespaceCtrl.text.trim();
    if (nsFilter.isNotEmpty && !ns.toLowerCase().contains(nsFilter.toLowerCase())) return false;
    // optional inline event name filter via simple syntax: "ns: foo" already handled above for ns; add name contains support using same field if prefixed with ev=
    final name = (e['event'] ?? e['name'] ?? '').toString();
    if (nsFilter.startsWith('ev=') && nsFilter.length > 3) {
      final ev = nsFilter.substring(3).trim();
      if (ev.isNotEmpty && !name.toLowerCase().contains(ev.toLowerCase())) return false;
    }
    return true;
  }

  @override
  Widget build(BuildContext context) {
    return Row(children: [
      Expanded(child: _Card(
        title: 'Frames',
        actions: [
          FilterChip(
            label: const Text('Pretty', style: TextStyle(fontSize: 12)),
            selected: _pretty && !_tree,
            onSelected: (v){ setState((){ _pretty = v; if (v) _tree = false; }); },
          ),
          const SizedBox(width: 6),
          FilterChip(
            label: const Text('Tree', style: TextStyle(fontSize: 12)),
            selected: _tree,
            onSelected: (v){ setState((){ _tree = v; if (v) _pretty = false; }); },
          ),
          const SizedBox(width: 6),
          IconButton(
            tooltip: 'Filters',
            icon: const Icon(Icons.filter_list, size: 18),
            onPressed: () => widget._openFilters(context),
          ),
        ],
        child: ListView.builder(
          itemCount: widget.frames.length,
          itemBuilder: (_, i) {
            final f = widget.frames[i] as Map<String, dynamic>;
            if (!_frameMatches(f)) { return const SizedBox.shrink(); }
            final preview = (f['preview'] ?? '').toString();
            final extractedJson = _extractJsonPayload(preview);
            final dir = (f['direction'] ?? '').toString();
            final isDown = dir == 'upstream->client';

            final ts = _fmtTime((f['ts'] ?? '').toString());
            final opcode = (f['opcode'] ?? '').toString();
            final size = (f['size'] ?? 0).toString();
            final isWsPingPong = opcode == 'ping' || opcode == 'pong';
            final isEnginePingPong = opcode == 'text' && (preview == '2' || preview == '3') && size == '1';
            final isHeartbeat = isWsPingPong || isEnginePingPong;
            final icon = Icon(
              isDown ? Icons.south : Icons.north,
              size: isHeartbeat ? 10 : 16,
              color: isDown ? context.appColors.success : context.appColors.primary,
            );
            if (widget.hideHeartbeats && isHeartbeat) {
              return const SizedBox.shrink();
            }
            if (isHeartbeat) {
              final label = isWsPingPong
                  ? opcode.toUpperCase()
                  : (preview == '2' ? 'PING' : 'PONG');
              return ListTile(
                dense: true,
                contentPadding: const EdgeInsets.symmetric(horizontal: 8),
                leading: Row(mainAxisSize: MainAxisSize.min, children: [icon, const SizedBox(width: 6), Text(label)]),
                trailing: Text(ts, style: Theme.of(context).textTheme.labelSmall?.copyWith(color: context.appColors.textSecondary)),
              );
            }
            return ExpansionTile(
              tilePadding: const EdgeInsets.symmetric(horizontal: 8),
              dense: true,
              leading: Row(mainAxisSize: MainAxisSize.min, children: [icon, const SizedBox(width: 6), Text(f['opcode'].toString())]),
              title: Text(preview, maxLines: 1, overflow: TextOverflow.ellipsis, style: context.appText.body),
              subtitle: Row(children: [
                Expanded(child: Text('${f['size']} B', style: Theme.of(context).textTheme.labelSmall?.copyWith(color: context.appColors.textSecondary))),
              ]),
              trailing: Text(ts, style: Theme.of(context).textTheme.labelSmall?.copyWith(color: context.appColors.textSecondary)),
              children: [
                Container(
                  alignment: Alignment.centerLeft,
                  padding: const EdgeInsets.all(8),
                  child: (extractedJson != null)
                      ? (_tree
                          ? JsonViewer(jsonString: extractedJson, forceTree: true)
                          : (_pretty
                              ? JsonViewer(jsonString: extractedJson, forceTree: false)
                              : SelectableText(preview, style: context.appText.monospace)))
                      : SelectableText(preview, style: context.appText.monospace),
                ),
              ],
            );
          },
        ),
      )),
      const VerticalDivider(width: 1),
      SizedBox(width: 200, child: _Card(
        title: 'Events',
        child: ListView.builder(
          itemCount: widget.events.length,
          itemBuilder: (_, i) {
            final e = widget.events[i] as Map<String, dynamic>;
            if (!_eventMatches(e)) { return const SizedBox.shrink(); }
            final args = (e['argsPreview'] ?? '').toString();
            return ExpansionTile(
              tilePadding: const EdgeInsets.symmetric(horizontal: 8),
              dense: true,
              title: Text('${e['namespace']} ${e['event']}', style: context.appText.subtitle),
              subtitle: Text(args, maxLines: 1, overflow: TextOverflow.ellipsis, style: context.appText.body),
              trailing: e['ackId'] != null ? Text('#${e['ackId']}') : null,
              children: [
                Container(alignment: Alignment.centerLeft, padding: const EdgeInsets.all(8), child: JsonViewer(jsonString: args)),
              ],
            );
          },
        ),
      )),
    ]);
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.title, required this.child, this.actions});
  final String title;
  final Widget child;
  final List<Widget>? actions;
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(children: [
              Expanded(child: Text(title, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold))),
              if (actions != null) ...actions!,
            ]),
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

extension on WsDetailsPanel {
  void _openFilters(BuildContext context) {
    showModalBottomSheet(context: context, builder: (_) {
      String localOpcode = opcodeFilter;
      String localDirection = directionFilter;
      bool localHideHeartbeats = hideHeartbeats;
      return StatefulBuilder(builder: (context, setState) {
        return Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('WebSocket filters', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              Row(children: [
                const Text('Opcode:'),
                const SizedBox(width: 8),
                DropdownButton<String>(
                  value: localOpcode,
                  style: TextStyle(color: Theme.of(context).colorScheme.onSurface),
                  items: const [
                    DropdownMenuItem(value: 'all', child: Text('Any')),
                    DropdownMenuItem(value: 'text', child: Text('Text')),
                    DropdownMenuItem(value: 'binary', child: Text('Binary')),
                    DropdownMenuItem(value: 'ping', child: Text('Ping')),
                    DropdownMenuItem(value: 'pong', child: Text('Pong')),
                    DropdownMenuItem(value: 'close', child: Text('Close')),
                  ],
                  onChanged: (v){ setState((){ localOpcode = v ?? 'all'; }); },
                ),
                const SizedBox(width: 16),
                const Text('Direction:'),
                const SizedBox(width: 8),
                DropdownButton<String>(
                  value: localDirection,
                  style: TextStyle(color: Theme.of(context).colorScheme.onSurface),
                  items: const [
                    DropdownMenuItem(value: 'all', child: Text('Any')),
                    DropdownMenuItem(value: 'client->upstream', child: Text('client->upstream')),
                    DropdownMenuItem(value: 'upstream->client', child: Text('upstream->client')),
                  ],
                  onChanged: (v){ setState((){ localDirection = v ?? 'all'; }); },
                ),
              ]),
              const SizedBox(height: 8),
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                dense: true,
                title: const Text('Hide heartbeats (ping/pong)'),
                value: localHideHeartbeats,
                onChanged: (v){ setState((){ localHideHeartbeats = v ?? false; }); },
              ),
              const SizedBox(height: 8),
              TextField(controller: namespaceCtrl, decoration: const InputDecoration(labelText: 'Namespace contains (or ev=eventName)'),),
              const SizedBox(height: 12),
              Align(
                alignment: Alignment.centerRight,
                child: ElevatedButton(onPressed: (){
                  onChangeOpcode(localOpcode);
                  onChangeDirection(localDirection);
                  onToggleHeartbeats(localHideHeartbeats);
                  Navigator.of(context).pop();
                }, child: const Text('Apply')),
              )
            ],
          ),
        );
      });
    });
  }
}

bool _isJsonLocal(String s) { try { jsonDecode(s); return true; } catch (_) { return false; } }

// Некоторые фреймы содержат обёртку протокола (socket.io) вида '42/namespace,[...]' или '2'/'3'
// Попробуем безопасно извлечь JSON часть, если она есть
String? _extractJsonPayload(String preview) {
  final trimmed = preview.trim();
  if (trimmed.isEmpty) return null;
  // чистый JSON
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    if (_isJsonLocal(trimmed)) return trimmed;
  }
  // socket.io payload: цифры/код + опциональный namespace + запятая + JSON-массив/объект
  final idxBrace = trimmed.indexOf('[');
  final idxBraceObj = trimmed.indexOf('{');
  int idx = -1;
  if (idxBrace >= 0 && idxBraceObj >= 0) {
    idx = idxBrace < idxBraceObj ? idxBrace : idxBraceObj;
  } else {
    idx = idxBrace >= 0 ? idxBrace : idxBraceObj;
  }
  if (idx > 0) {
    final candidate = trimmed.substring(idx);
    if (_isJsonLocal(candidate)) return candidate;
  }
  return null;
}

class _JsonToggleRow extends StatefulWidget {
  const _JsonToggleRow({required this.json});
  final String json;
  @override
  State<_JsonToggleRow> createState() => _JsonToggleRowState();
}

class _JsonToggleRowState extends State<_JsonToggleRow> {
  bool pretty = true;
  bool tree = false;
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(spacing: 8, children: [
          FilterChip(label: const Text('Pretty', style: TextStyle(fontSize: 12)), selected: pretty && !tree, onSelected: (v){ setState(() { tree = false; pretty = true; }); }),
          FilterChip(label: const Text('Tree', style: TextStyle(fontSize: 12)), selected: tree, onSelected: (v){ setState(() { tree = v; pretty = !v; }); }),
          TextButton.icon(
            style: TextButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6), textStyle: const TextStyle(fontSize: 12)),
            onPressed: (){ Clipboard.setData(ClipboardData(text: widget.json)); },
            icon: const Icon(Icons.copy, size: 16),
            label: const Text('Copy'),
          ),
        ]),
        const SizedBox(height: 6),
        // Контент
        if (tree)
          JsonViewer(jsonString: widget.json, forceTree: true)
        else
          JsonViewer(jsonString: widget.json),
      ],
    );
  }
}


