import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../theme/context_ext.dart';
import 'common/copyable_items.dart';

/// Request info tab widget
///
/// Displays:
/// 1. Query Parameters section
/// 2. Headers section (with sensitive data masking)
///
/// Follows SRP - only responsible for displaying request metadata.
class RequestInfoTab extends StatelessWidget {
  const RequestInfoTab({
    super.key,
    required this.queryParams,
    required this.headers,
    this.headersRaw,
  });

  /// Query parameters from URL
  /// Map of parameter name -> list of values
  final Map<String, List<String>> queryParams;

  /// Request headers (may be masked for sensitive data)
  final Map<String, String> headers;

  /// Raw unmasked headers (preferred for sensitive fields like Cookie)
  final Map<String, String>? headersRaw;

  @override
  Widget build(BuildContext context) {
    return CustomScrollView(
      slivers: [
        // Query Parameters section
        if (queryParams.isNotEmpty) ..._buildQueryParamsSection(context),

        const SliverToBoxAdapter(child: SizedBox(height: 8)),

        // Headers section
        ..._buildHeadersSection(context),
      ],
    );
  }

  /// Build query parameters section
  List<Widget> _buildQueryParamsSection(BuildContext context) {
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Query Parameters', style: context.appText.subtitle),
            const SizedBox(height: 4),
            ...queryParams.entries.map(
              (e) => Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: CopyableKeyValueItem(
                  name: e.key,
                  value: e.value.join(', '),
                ),
              ),
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    ];
  }

  /// Copies all headers to clipboard
  void _copyAllHeaders() {
    final raw = headersRaw ?? headers;
    final text = raw.entries.map((e) => '${e.key}: ${e.value}').join('\n');
    Clipboard.setData(ClipboardData(text: text));
  }

  /// Build headers section
  List<Widget> _buildHeadersSection(BuildContext context) {
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text('Headers', style: context.appText.subtitle),
                const SizedBox(width: 6),
                _CopyAllButton(onTap: _copyAllHeaders),
              ],
            ),
            const SizedBox(height: 4),
            ...headers.entries.map(
              (e) => Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: HeaderItem(
                  name: e.key,
                  value: e.value,
                  raw: headersRaw?[e.key],
                ),
              ),
            ),
          ],
        ),
      ),
    ];
  }
}

/// Compact copy button for section headers
class _CopyAllButton extends StatefulWidget {
  const _CopyAllButton({required this.onTap});

  final VoidCallback onTap;

  @override
  State<_CopyAllButton> createState() => _CopyAllButtonState();
}

class _CopyAllButtonState extends State<_CopyAllButton> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: _hover
                ? Theme.of(context).colorScheme.primary.withValues(alpha: 0.12)
                : Colors.transparent,
            borderRadius: BorderRadius.circular(4),
          ),
          child: Icon(
            Icons.copy,
            size: 14,
            color: _hover
                ? Theme.of(context).colorScheme.primary
                : context.appColors.textSecondary,
          ),
        ),
      ),
    );
  }
}
