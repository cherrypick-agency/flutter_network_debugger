import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import 'common/copyable_items.dart';

/// Response info tab widget
///
/// Displays:
/// 1. Headers section (with sensitive data masking)
/// 2. Security section (TLS info, Cookies summary)
/// 3. Cache & CORS section (Cache status, CORS validation)
///
/// Follows SRP - only responsible for displaying response metadata.
class ResponseInfoTab extends StatelessWidget {
  const ResponseInfoTab({
    super.key,
    required this.headers,
    this.headersRaw,
    this.tlsInfo,
    this.cookieSummary,
    this.cacheInfo,
    this.corsInfo,
  });

  /// Response headers (may be masked for sensitive data)
  final Map<String, String> headers;

  /// Raw unmasked headers
  final Map<String, String>? headersRaw;

  /// TLS information from response preview
  /// Contains: version, cipherSuite, alpn, serverName, peerCertificates
  final Map<String, dynamic>? tlsInfo;

  /// Cookie summary from response preview
  /// Contains: setCookieCount, secure, httpOnly, sameSiteLax, sameSiteStrict, sameSiteNone
  final Map<String, dynamic>? cookieSummary;

  /// Cache metadata
  /// Contains: status (HIT/MISS/REVALIDATED), cache-control, etag, age
  final Map<String, dynamic>? cacheInfo;

  /// CORS metadata
  /// Contains: ok, reason, allowedOrigin, allowedMethods, allowedHeaders, preflight
  final Map<String, dynamic>? corsInfo;

  @override
  Widget build(BuildContext context) {
    return CustomScrollView(
      slivers: [
        // Headers section
        ..._buildHeadersSection(context),

        const SliverToBoxAdapter(child: SizedBox(height: 12)),

        // Security section
        ..._buildSecuritySection(context),

        const SliverToBoxAdapter(child: SizedBox(height: 12)),

        // Cache & CORS section
        ..._buildCacheAndCorsSection(context),
      ],
    );
  }

  /// Build headers section
  List<Widget> _buildHeadersSection(BuildContext context) {
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Headers', style: context.appText.subtitle),
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

  /// Build security section (TLS + Cookies)
  List<Widget> _buildSecuritySection(BuildContext context) {
    final widgets = <Widget>[];

    widgets.add(Text('Security', style: context.appText.subtitle));
    widgets.add(const SizedBox(height: 6));

    // TLS summary
    if (tlsInfo != null && tlsInfo!.isNotEmpty) {
      widgets.addAll(_buildTlsInfo(context));
    }

    // Cookie summary
    if (cookieSummary != null && cookieSummary!.isNotEmpty) {
      widgets.addAll(_buildCookieSummary(context));
    }

    // If no security info, show placeholder
    if (tlsInfo == null && cookieSummary == null) {
      widgets.add(SelectableText('—', style: context.appText.monospace));
    }

    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: widgets,
        ),
      ),
    ];
  }

  /// Build TLS information
  List<Widget> _buildTlsInfo(BuildContext context) {
    final widgets = <Widget>[];

    // TLS version and cipher
    widgets.add(
      SelectableText(
        'TLS: ${tlsInfo!['version'] ?? ''} ${tlsInfo!['cipherSuite'] ?? ''}  ALPN: ${tlsInfo!['alpn'] ?? ''}',
        style: context.appText.monospace,
      ),
    );

    // Server Name Indication (SNI)
    if ((tlsInfo!['serverName'] ?? '').toString().isNotEmpty) {
      widgets.add(
        SelectableText(
          'SNI: ${tlsInfo!['serverName']}',
          style: context.appText.monospace,
        ),
      );
    }

    // Certificates
    final certs =
        (tlsInfo!['peerCertificates'] as List?)?.cast<Map<String, dynamic>>() ??
        const [];
    if (certs.isNotEmpty) {
      widgets.add(const SizedBox(height: 4));
      widgets.add(
        Text('Certificates', style: Theme.of(context).textTheme.labelLarge),
      );
      for (final cert in certs) {
        widgets.add(
          Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: SelectableText(
              '- ${cert['subject']} | Issuer: ${cert['issuer']} | NotAfter: ${cert['notAfter']}',
              style: context.appText.monospace,
            ),
          ),
        );
      }
      widgets.add(const SizedBox(height: 6));
    }

    return widgets;
  }

  /// Build cookie summary
  List<Widget> _buildCookieSummary(BuildContext context) {
    final widgets = <Widget>[];

    widgets.add(Text('Cookies', style: Theme.of(context).textTheme.labelLarge));
    widgets.add(
      SelectableText(
        'Set-Cookie: ${cookieSummary!['setCookieCount']} | '
        'Secure: ${cookieSummary!['secure']} | '
        'HttpOnly: ${cookieSummary!['httpOnly']} | '
        'SameSite Lax/Strict/None: ${cookieSummary!['sameSiteLax']}/${cookieSummary!['sameSiteStrict']}/${cookieSummary!['sameSiteNone']}',
        style: context.appText.monospace,
      ),
    );

    return widgets;
  }

  /// Build Cache & CORS section
  List<Widget> _buildCacheAndCorsSection(BuildContext context) {
    final widgets = <Widget>[];

    widgets.add(Text('Cache & CORS', style: context.appText.subtitle));
    widgets.add(const SizedBox(height: 6));

    // Status chips
    widgets.add(
      Wrap(
        spacing: 8,
        runSpacing: 4,
        children: [
          if (cacheInfo?['status'] != null)
            _chip(context, 'cache: ${cacheInfo!['status']}'),
          if (corsInfo != null)
            _chip(context, corsInfo!['ok'] == true ? 'CORS OK' : 'CORS Fail'),
          if (headers['Vary'] != null || headers['vary'] != null)
            _chip(context, 'Vary: ${headers['Vary'] ?? headers['vary']}'),
        ],
      ),
    );
    widgets.add(const SizedBox(height: 6));

    // Cache details table
    if (cacheInfo != null && cacheInfo!.isNotEmpty) {
      widgets.add(Text('Cache', style: Theme.of(context).textTheme.labelLarge));
      widgets.add(const SizedBox(height: 4));
      widgets.addAll(
        _kvList(cacheInfo!).map(
          (e) => Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: CopyableKeyValueItem(name: e.key, value: e.value),
          ),
        ),
      );
      widgets.add(const SizedBox(height: 8));
    }

    // CORS details table
    if (corsInfo != null && corsInfo!.isNotEmpty) {
      widgets.add(Text('CORS', style: Theme.of(context).textTheme.labelLarge));
      widgets.add(const SizedBox(height: 4));
      widgets.addAll(
        _kvList({
          'ok': (corsInfo!['ok'] ?? false).toString(),
          if (corsInfo!['reason'] != null) 'reason': corsInfo!['reason'],
          'allowedOrigin': corsInfo!['allowedOrigin'] ?? '',
          'allowedMethods': (corsInfo!['allowedMethods'] ?? []).toString(),
          'allowedHeaders': (corsInfo!['allowedHeaders'] ?? []).toString(),
          'preflight': (corsInfo!['preflight'] == true).toString(),
        }).map(
          (e) => Padding(
            padding: const EdgeInsets.only(bottom: 2),
            child: CopyableKeyValueItem(name: e.key, value: e.value),
          ),
        ),
      );
    }

    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: widgets,
        ),
      ),
    ];
  }

  /// Convert map to list of key-value entries (helper)
  List<MapEntry<String, String>> _kvList(Map<String, dynamic> m) {
    return m.entries
        .map(
          (e) => MapEntry(
            e.key,
            e.value is List ? (e.value as List).join(', ') : e.value.toString(),
          ),
        )
        .toList();
  }

  /// Build chip widget
  Widget _chip(BuildContext context, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(text, style: Theme.of(context).textTheme.labelSmall),
    );
  }
}
