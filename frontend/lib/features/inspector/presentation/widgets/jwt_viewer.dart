import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import '../utils/jwt_detector.dart';

/// Widget for displaying decoded JWT token
class JwtViewer extends StatelessWidget {
  final String token;
  final JwtData? jwtData;

  const JwtViewer({required this.token, this.jwtData, super.key});

  @override
  Widget build(BuildContext context) {
    final jwt = jwtData ?? JwtContentDetector().parse(token);

    if (jwt == null) {
      return SelectableText(token, style: context.appText.monospace);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        _buildValidityStatus(context, jwt),
        const SizedBox(height: 12),
        _buildMetadataChips(context, jwt),
        const SizedBox(height: 12),
        _JwtSection(
          title: 'Header',
          json: jwt.headerRaw,
          onCopy: () => _copyToClipboard(jwt.headerRaw),
        ),
        const SizedBox(height: 12),
        _JwtSection(
          title: 'Payload',
          json: jwt.payloadRaw,
          onCopy: () => _copyToClipboard(jwt.payloadRaw),
        ),
        const SizedBox(height: 12),
        _JwtSection(
          title: 'Signature',
          text: jwt.signature,
          onCopy: () => _copyToClipboard(jwt.signature),
          isSignature: true,
        ),
      ],
    );
  }

  Widget _buildValidityStatus(BuildContext context, JwtData jwt) {
    final isValid = jwt.isValid;
    final expiresAt = jwt.expiresAt;
    final now = DateTime.now();

    Color statusColor;
    String statusText;
    IconData statusIcon;

    if (expiresAt == null) {
      statusColor = Colors.grey;
      statusText = 'No expiration';
      statusIcon = Icons.info_outline;
    } else if (now.isAfter(expiresAt)) {
      statusColor = Theme.of(context).colorScheme.error;
      statusText = 'Expired';
      statusIcon = Icons.error_outline;
    } else if (isValid) {
      final timeLeft = jwt.timeUntilExpiration;
      if (timeLeft != null && timeLeft.inHours < 1) {
        statusColor = Colors.orange;
        statusText = 'Expires in ${timeLeft.inMinutes}m';
        statusIcon = Icons.warning_amber;
      } else {
        statusColor = Colors.green;
        statusText = 'Valid';
        statusIcon = Icons.check_circle_outline;
      }
    } else {
      statusColor = Theme.of(context).colorScheme.error;
      statusText = 'Invalid';
      statusIcon = Icons.cancel_outlined;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: statusColor.withOpacity(0.1),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: statusColor.withOpacity(0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(statusIcon, size: 16, color: statusColor),
          const SizedBox(width: 6),
          Text(
            statusText,
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: statusColor,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMetadataChips(BuildContext context, JwtData jwt) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        if (jwt.algorithm != null) _chip(context, 'Alg: ${jwt.algorithm}'),
        if (jwt.tokenType != null) _chip(context, 'Type: ${jwt.tokenType}'),
        if (jwt.issuer != null) _chip(context, 'Issuer: ${jwt.issuer}'),
        if (jwt.subject != null) _chip(context, 'Subject: ${jwt.subject}'),
        if (jwt.issuedAt != null)
          _chip(context, 'Issued: ${_formatTimestamp(jwt.issuedAt!)}'),
        if (jwt.expiresAt != null)
          _chip(context, 'Expires: ${_formatTimestamp(jwt.expiresAt!)}'),
      ],
    );
  }

  Widget _chip(BuildContext context, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceVariant,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(text, style: Theme.of(context).textTheme.labelSmall),
    );
  }

  String _formatTimestamp(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);

    if (diff.isNegative) {
      final absDiff = diff.abs();
      if (absDiff.inDays > 0) return 'in ${absDiff.inDays}d';
      if (absDiff.inHours > 0) return 'in ${absDiff.inHours}h';
      return 'in ${absDiff.inMinutes}m';
    } else {
      if (diff.inDays > 0) return '${diff.inDays}d ago';
      if (diff.inHours > 0) return '${diff.inHours}h ago';
      return '${diff.inMinutes}m ago';
    }
  }

  void _copyToClipboard(String text) {
    Clipboard.setData(ClipboardData(text: text));
  }
}

class _JwtSection extends StatelessWidget {
  final String title;
  final String? json;
  final String? text;
  final VoidCallback onCopy;
  final bool isSignature;

  const _JwtSection({
    required this.title,
    this.json,
    this.text,
    required this.onCopy,
    this.isSignature = false,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(title, style: context.appText.subtitle),
            const SizedBox(width: 8),
            TextButton.icon(
              onPressed: onCopy,
              icon: const Icon(Icons.copy, size: 14),
              label: const Text('Copy'),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            border: Border.all(color: Theme.of(context).dividerColor),
            borderRadius: BorderRadius.circular(4),
          ),
          child:
              isSignature
                  ? SelectableText(text ?? '', style: context.appText.monospace)
                  : JsonViewer(jsonString: json ?? '{}', forceTree: false),
        ),
      ],
    );
  }
}
