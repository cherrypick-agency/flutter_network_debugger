import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:hex_viewer/hex_viewer.dart';
import 'package:http/http.dart' as http;

/// Renders HTTP body content as hexadecimal view
///
/// This widget handles conversion from string body to bytes
/// and displays them using the HexViewer component.
///
/// For large bodies (>50KB preview limit), it can fetch full data from API.
class HexBodyRenderer extends StatefulWidget {
  final String body;
  final String? sessionId;
  final String? frameId;
  final int? bodySize;
  final int bytesPerLine;
  final int groupSize;
  final String? apiBaseUrl;

  const HexBodyRenderer({
    super.key,
    required this.body,
    this.sessionId,
    this.frameId,
    this.bodySize,
    this.bytesPerLine = 16,
    this.groupSize = 8,
    this.apiBaseUrl,
  });

  @override
  State<HexBodyRenderer> createState() => _HexBodyRendererState();
}

class _HexBodyRendererState extends State<HexBodyRenderer> {
  List<int>? _fullData;
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    if (_shouldLoadFromApi()) {
      _loadFullData();
    }
  }

  @override
  void didUpdateWidget(HexBodyRenderer oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Reset and reload if frame data changed
    if (oldWidget.sessionId != widget.sessionId ||
        oldWidget.frameId != widget.frameId ||
        oldWidget.bodySize != widget.bodySize) {
      setState(() {
        _fullData = null;
        _loading = false;
        _error = null;
      });
      if (_shouldLoadFromApi()) {
        _loadFullData();
      }
    }
  }

  @override
  void dispose() {
    // Clear large data buffer
    _fullData = null;
    super.dispose();
  }

  bool _shouldLoadFromApi() {
    // Load from API if:
    // 1. We have session/frame IDs
    // 2. Body size is larger than current body length (indicating truncation)
    return widget.sessionId != null &&
        widget.frameId != null &&
        widget.bodySize != null &&
        widget.bodySize! > utf8.encode(widget.body).length;
  }

  Future<void> _loadFullData() async {
    if (!mounted) return;

    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final baseUrl = widget.apiBaseUrl ?? 'http://localhost:8080';
      final url =
          '$baseUrl/api/v1/sessions/${widget.sessionId}/frames/${widget.frameId}/body';

      final response = await http
          .get(Uri.parse(url))
          .timeout(
            const Duration(seconds: 30),
            onTimeout: () {
              throw TimeoutException('Request timed out after 30 seconds');
            },
          );

      if (!mounted) return;

      if (response.statusCode == 200) {
        setState(() {
          _fullData = response.bodyBytes;
          _loading = false;
        });
      } else if (response.statusCode == 413) {
        setState(() {
          _error = 'Body too large (exceeds 100MB limit)';
          _loading = false;
        });
      } else if (response.statusCode == 404) {
        setState(() {
          _error = 'Frame body not found';
          _loading = false;
        });
      } else {
        setState(() {
          _error = 'Failed to load body: HTTP ${response.statusCode}';
          _loading = false;
        });
      }
    } catch (e) {
      if (!mounted) return;

      setState(() {
        _error = 'Network error: $e';
        _loading = false;
      });
    }
  }

  List<int> _encodeBodySafely(String body) {
    // Check if all characters are in Latin1 range (0-255)
    // This preserves original byte values for binary data
    if (body.runes.every((rune) => rune <= 0xFF)) {
      return latin1.encode(body);
    }

    // Fallback to UTF-8 for text with non-Latin1 characters
    // This handles regular UTF-8 text correctly
    return utf8.encode(body);
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            Text(
              'Loading full body data...',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            Text(
              '${widget.bodySize} bytes',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      );
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.error_outline,
              size: 48,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(height: 16),
            Text(
              'Error loading hex data',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              _error!,
              style: Theme.of(context).textTheme.bodySmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            ElevatedButton.icon(
              onPressed: _loadFullData,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    // Use loaded data if available, otherwise encode body string
    final bytes = _fullData ?? _encodeBodySafely(widget.body);

    return HexViewer(
      data: bytes,
      config: HexConfig(
        bytesPerLine: widget.bytesPerLine,
        groupSize: widget.groupSize,
        colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
      ),
    );
  }

  /// Renders hex data as slivers for use in CustomScrollView.
  ///
  /// Returns a list of sliver widgets that can be inserted into a CustomScrollView.
  /// This eliminates nested scroll conflicts by integrating with the parent scroll view.
  ///
  /// Handles three states:
  /// - Loading: Shows progress indicator while fetching full body data
  /// - Error: Shows error message with retry button
  /// - Success: Returns HexViewer slivers with the data
  List<Widget> buildAsSlivers() {
    // Loading state
    if (_loading) {
      return [
        SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(32.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(),
                  const SizedBox(height: 16),
                  Text(
                    'Loading full body data...',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                  if (widget.bodySize != null)
                    Text(
                      '${widget.bodySize} bytes',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                    ),
                ],
              ),
            ),
          ),
        ),
      ];
    }

    // Error state
    if (_error != null) {
      return [
        SliverToBoxAdapter(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(32.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.error_outline,
                    size: 48,
                    color: Theme.of(context).colorScheme.error,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Error loading hex data',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _error!,
                    style: Theme.of(context).textTheme.bodySmall,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: _loadFullData,
                    icon: const Icon(Icons.refresh),
                    label: const Text('Retry'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ];
    }

    // Success state - use HexViewer's static sliver builder
    final bytes = _fullData ?? _encodeBodySafely(widget.body);
    final hexConfig = HexConfig(
      bytesPerLine: widget.bytesPerLine,
      groupSize: widget.groupSize,
      colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
    );

    // Use static method to build slivers without widget state
    // This avoids nested scrolling issues
    return HexViewer.buildStaticSlivers(
      context: context,
      data: bytes,
      config: hexConfig,
    );
  }
}
