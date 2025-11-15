import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/byte_selection.dart';
import '../models/hex_config.dart';
import '../formatting/hex_formatter.dart';
import 'hex_header.dart';
import 'hex_row.dart';
import 'hex_controls.dart';

/// A hexadecimal viewer widget for displaying binary data with advanced features.
///
/// This widget provides a professional hex editor-style view of binary data with:
/// - Virtualized scrolling for handling large files efficiently
/// - Byte selection with Shift+click for range selection
/// - Copy to clipboard in hex, text, or binary format
/// - Configurable byte grouping and display options
/// - Color-coded bytes (null bytes, ASCII, control chars)
/// - Dark and light theme support
///
/// Example:
/// ```dart
/// HexViewer(
///   data: myBinaryData,
///   config: HexConfig(
///     bytesPerLine: 16,
///     groupSize: 8,
///     colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
///   ),
///   onSelectionChanged: (selection) {
///     print('Selected: ${selection?.length} bytes');
///   },
/// )
/// ```
///
/// See also:
/// - [HexConfig] for configuration options
/// - [ByteSelection] for selection information
/// - [ByteColorScheme] for color customization
class HexViewer extends StatefulWidget {
  /// Binary data to display as hexadecimal.
  ///
  /// This can be any [List<int>] including file contents, network packets,
  /// or any binary data. The widget efficiently handles large data sets
  /// using virtualized scrolling.
  final List<int> data;

  /// Display configuration including bytes per line, grouping, and colors.
  ///
  /// Defaults to [HexConfig()] with standard settings (16 bytes per line,
  /// 8-byte grouping). See [HexConfig] for all available options.
  final HexConfig config;

  /// Callback invoked when the user selects or deselects bytes.
  ///
  /// Returns [null] when selection is cleared, or a [ByteSelection] object
  /// containing the start offset, end offset, and selected bytes.
  ///
  /// Example:
  /// ```dart
  /// onSelectionChanged: (selection) {
  ///   if (selection != null) {
  ///     print('Selected ${selection.length} bytes from '
  ///           '${selection.start} to ${selection.end}');
  ///   }
  /// }
  /// ```
  final ValueChanged<ByteSelection?>? onSelectionChanged;

  /// Optional custom monospace font family for hex and ASCII display.
  ///
  /// If not provided, uses the default monospace font from the theme.
  /// Recommended fonts: 'Courier New', 'Monaco', 'Consolas', 'Roboto Mono'.
  final String? monospaceFontFamily;

  /// Creates a hexadecimal viewer widget.
  ///
  /// The [data] parameter is required and contains the binary data to display.
  /// All other parameters are optional with sensible defaults.
  const HexViewer({
    super.key,
    required this.data,
    this.config = const HexConfig(),
    this.onSelectionChanged,
    this.monospaceFontFamily,
  });

  @override
  State<HexViewer> createState() => _HexViewerState();

  /// Builds slivers for displaying hex data in a CustomScrollView without state.
  ///
  /// This is a static method that creates read-only hex viewer slivers without
  /// interactive features like selection or copy. Use this when you need to embed
  /// hex data in a CustomScrollView without nested scroll conflicts.
  ///
  /// Returns an empty list if [data] is empty.
  ///
  /// Example:
  /// ```dart
  /// CustomScrollView(
  ///   slivers: [
  ///     ...HexViewer.buildStaticSlivers(
  ///       context: context,
  ///       data: myData,
  ///       config: HexConfig(),
  ///     ),
  ///   ],
  /// )
  /// ```
  static List<Widget> buildStaticSlivers({
    required BuildContext context,
    required List<int> data,
    required HexConfig config,
    String? monospaceFontFamily,
  }) {
    if (data.isEmpty) {
      return [
        SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Text(
                'No data to display',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
              ),
            ),
          ),
        ),
      ];
    }

    final monoStyle = TextStyle(
      fontFamily: monospaceFontFamily ?? 'monospace',
      fontSize: 13,
    );

    final lineCount = HexFormatter.calculateLineCount(
      data.length,
      config.bytesPerLine,
    );

    return [
      // Header
      SliverToBoxAdapter(
        child: HexHeader(config: config, monoStyle: monoStyle),
      ),

      const SliverToBoxAdapter(child: Divider(height: 1)),

      // Hex rows as SliverList (virtualized, no scroll conflict)
      SliverList(
        delegate: SliverChildBuilderDelegate((context, index) {
          final offset = index * config.bytesPerLine;
          final line = HexFormatter.formatLine(data, offset, config);

          return HexRow(
            line: line,
            config: config,
            selection: null, // Read-only mode, no selection
            monoStyle: monoStyle,
            onBytePressed: null, // Read-only mode, no interaction
          );
        }, childCount: lineCount),
      ),
    ];
  }
}

class _HexViewerState extends State<HexViewer> {
  ByteSelection? _selection;
  final ScrollController _scrollController = ScrollController();
  HexConfig? _runtimeConfig;
  int? _selectionAnchor;

  HexConfig get _effectiveConfig => _runtimeConfig ?? widget.config;

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _handleBytePressed(int offset, {bool isShiftHeld = false}) {
    if (!_effectiveConfig.enableSelection) return;

    setState(() {
      if (isShiftHeld && _selectionAnchor != null) {
        // Extend selection
        _selection = ByteSelection.range(_selectionAnchor!, offset);
      } else {
        // New selection
        _selectionAnchor = offset;
        _selection = ByteSelection.single(offset);
      }
    });

    widget.onSelectionChanged?.call(_selection);
  }

  void _handleCopy(CopyFormat format) {
    if (_selection == null) return;

    final bytes = _selection!.getBytes(widget.data);
    String text;

    switch (format) {
      case CopyFormat.hex:
        text = HexFormatter.formatSelectionAsHex(bytes);
      case CopyFormat.binary:
        text = HexFormatter.formatSelectionAsBinary(bytes);
      case CopyFormat.text:
        text = HexFormatter.formatSelectionAsText(bytes);
    }

    Clipboard.setData(ClipboardData(text: text));

    // Show snackbar feedback
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Copied ${bytes.length} bytes as ${format.name}'),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  void _handleConfigChange(HexConfig newConfig) {
    setState(() {
      _runtimeConfig = newConfig;
    });
  }

  int _calculateLineCount() {
    return HexFormatter.calculateLineCount(
      widget.data.length,
      _effectiveConfig.bytesPerLine,
    );
  }

  @override
  Widget build(BuildContext context) {
    if (widget.data.isEmpty) {
      return Center(
        child: Text(
          'No data to display',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
      );
    }

    final monoStyle = TextStyle(
      fontFamily: widget.monospaceFontFamily ?? 'monospace',
      fontSize: 13,
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Controls
        HexControls(
          config: _effectiveConfig,
          selection: _selection,
          onCopy: _handleCopy,
          onConfigChanged: _handleConfigChange,
        ),

        const Divider(height: 1),

        // Header
        HexHeader(config: _effectiveConfig, monoStyle: monoStyle),

        const Divider(height: 1),

        // Virtualized list
        Expanded(
          child: Scrollbar(
            controller: _scrollController,
            thumbVisibility: true,
            child: ListView.builder(
              controller: _scrollController,
              itemCount: _calculateLineCount(),
              itemBuilder: (context, index) {
                final offset = index * _effectiveConfig.bytesPerLine;
                final line = HexFormatter.formatLine(
                  widget.data,
                  offset,
                  _effectiveConfig,
                );

                return HexRow(
                  line: line,
                  config: _effectiveConfig,
                  selection: _selection,
                  monoStyle: monoStyle,
                  onBytePressed: (byteOffset, {bool isShiftHeld = false}) {
                    _handleBytePressed(byteOffset, isShiftHeld: isShiftHeld);
                  },
                );
              },
            ),
          ),
        ),

        // Status bar
        if (_selection != null)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              border: Border(
                top: BorderSide(color: Theme.of(context).dividerColor),
              ),
            ),
            child: Row(
              children: [
                Text(
                  'Selected: ${_selection!.length} bytes '
                  '(0x${_selection!.startOffset.toRadixString(16).toUpperCase()} - '
                  '0x${_selection!.endOffset.toRadixString(16).toUpperCase()})',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
              ],
            ),
          ),
      ],
    );
  }

  /// Renders hex data as slivers for use in CustomScrollView.
  ///
  /// Returns a list of sliver widgets that can be inserted into a CustomScrollView.
  /// This allows the hex viewer to integrate seamlessly with parent scrolling
  /// without creating nested scroll conflicts.
  ///
  /// Use this method when you want hex data to be part of a larger scrollable area
  /// rather than having its own independent scroll view.
  List<Widget> buildAsSlivers() {
    if (widget.data.isEmpty) {
      return [
        SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Text(
                'No data to display',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
              ),
            ),
          ),
        ),
      ];
    }

    final monoStyle = TextStyle(
      fontFamily: widget.monospaceFontFamily ?? 'monospace',
      fontSize: 13,
    );

    return [
      // Controls
      SliverToBoxAdapter(
        child: HexControls(
          config: _effectiveConfig,
          selection: _selection,
          onCopy: _handleCopy,
          onConfigChanged: _handleConfigChange,
        ),
      ),

      // Divider
      const SliverToBoxAdapter(child: Divider(height: 1)),

      // Header
      SliverToBoxAdapter(
        child: HexHeader(config: _effectiveConfig, monoStyle: monoStyle),
      ),

      // Divider
      const SliverToBoxAdapter(child: Divider(height: 1)),

      // Hex rows as SliverList (virtualized, no scroll conflict)
      SliverList(
        delegate: SliverChildBuilderDelegate((context, index) {
          final offset = index * _effectiveConfig.bytesPerLine;
          final line = HexFormatter.formatLine(
            widget.data,
            offset,
            _effectiveConfig,
          );

          return HexRow(
            line: line,
            config: _effectiveConfig,
            selection: _selection,
            monoStyle: monoStyle,
            onBytePressed: (byteOffset, {bool isShiftHeld = false}) {
              _handleBytePressed(byteOffset, isShiftHeld: isShiftHeld);
            },
          );
        }, childCount: _calculateLineCount()),
      ),

      // Status bar (if selection exists)
      if (_selection != null)
        SliverToBoxAdapter(
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              border: Border(
                top: BorderSide(color: Theme.of(context).dividerColor),
              ),
            ),
            child: Row(
              children: [
                Text(
                  'Selected: ${_selection!.length} bytes '
                  '(0x${_selection!.startOffset.toRadixString(16).toUpperCase()} - '
                  '0x${_selection!.endOffset.toRadixString(16).toUpperCase()})',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
              ],
            ),
          ),
        ),
    ];
  }
}

/// Copy format options
enum CopyFormat { hex, binary, text }
