# Week 1: HTTP Inspector Widgets Creation

**Goal**: Extract HTTP Inspector UI components into separate, reusable widget files

**Duration**: 5 days (1 widget per day)

**Success Criteria**:
- 5 new widget files created
- Each widget compiles without errors
- Each widget tested in isolation with mock data
- No dependencies on parent state (data passed via parameters)
- Clean separation of concerns

---

## Day 1: Request Action Buttons Widget

### Objective
Create standalone widget for Request action buttons (Copy URL, Copy as cURL with dropdown, Repeat)

### File to Create
`lib/features/http_inspector/presentation/widgets/request_action_buttons.dart`

### Code to Extract
**From**: `http_details_panel.dart` lines 630-695

### Widget Structure
```dart
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class RequestActionButtons extends StatelessWidget {
  const RequestActionButtons({
    super.key,
    required this.req,
    required this.url,
    required this.normalizedUrl,
    required this.loadingFetch,
    required this.onRepeat,
    required this.buildCurl,
  });

  final Map<String, dynamic>? req;
  final String url;
  final String? normalizedUrl;
  final bool loadingFetch;
  final VoidCallback onRepeat;
  final String Function(Map<String, dynamic>, CurlExportMode) buildCurl;

  @override
  Widget build(BuildContext context) {
    if (url.isEmpty) return const SizedBox.shrink();

    return Wrap(
      spacing: 8,
      children: [
        // Copy URL button
        TextButton.icon(
          onPressed: () {
            Clipboard.setData(ClipboardData(text: normalizedUrl ?? url));
          },
          icon: const Icon(Icons.link, size: 16),
          label: const Text('Copy URL'),
        ),

        // Copy as cURL dropdown menu
        MenuAnchor(
          builder: (context, controller, child) {
            return TextButton.icon(
              onPressed: () {
                if (req != null) {
                  final curl = buildCurl(req!, CurlExportMode.compact);
                  Clipboard.setData(ClipboardData(text: curl));
                }
              },
              onLongPress: () {
                if (controller.isOpen) {
                  controller.close();
                } else {
                  controller.open();
                }
              },
              icon: const Icon(Icons.content_paste, size: 16),
              label: const Text('Copy as cURL'),
            );
          },
          menuChildren: [
            MenuItemButton(
              leadingIcon: const Icon(Icons.remove, size: 16),
              child: const Text('Compact'),
              onPressed: () {
                if (req != null) {
                  final curl = buildCurl(req!, CurlExportMode.compact);
                  Clipboard.setData(ClipboardData(text: curl));
                }
              },
            ),
            MenuItemButton(
              leadingIcon: const Icon(Icons.wrap_text, size: 16),
              child: const Text('Multiline'),
              onPressed: () {
                if (req != null) {
                  final curl = buildCurl(req!, CurlExportMode.multiline);
                  Clipboard.setData(ClipboardData(text: curl));
                }
              },
            ),
            MenuItemButton(
              leadingIcon: const Icon(Icons.settings, size: 16),
              child: const Text('With options'),
              onPressed: () {
                if (req != null) {
                  final curl = buildCurl(req!, CurlExportMode.withOptions);
                  Clipboard.setData(ClipboardData(text: curl));
                }
              },
            ),
          ],
        ),

        // Repeat button
        if (loadingFetch)
          const SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        else
          TextButton.icon(
            onPressed: onRepeat,
            icon: const Icon(Icons.refresh, size: 16),
            label: const Text('Repeat'),
          ),
      ],
    );
  }
}

// Required enum (if not already exported)
enum CurlExportMode { compact, multiline, withOptions }
```

### Testing Checklist
- [ ] Widget file created
- [ ] Imports all dependencies
- [ ] Compiles without errors
- [ ] Copy URL button works
- [ ] Copy as cURL dropdown opens
- [ ] All 3 cURL modes copy to clipboard
- [ ] Repeat button shows loading state
- [ ] Repeat callback fires
- [ ] Null handling for `req` works

### Parameters Required from Parent
- `Map<String, dynamic>? req` - Request data
- `String url` - Raw URL
- `String? normalizedUrl` - Cleaned URL with normalized query params
- `bool loadingFetch` - Loading state for Repeat button
- `VoidCallback onRepeat` - Callback for Repeat action
- `String Function(Map<String, dynamic>, CurlExportMode) buildCurl` - cURL builder

### Integration Point
Will be used in `_Card` widget's `actions` parameter:
```dart
_Card(
  title: 'Request',
  actions: [
    RequestActionButtons(
      req: req,
      url: url,
      normalizedUrl: normalizedUrl,
      loadingFetch: _loadingFetch,
      onRepeat: () => _refetchOriginalRequest(context, req),
      buildCurl: _buildCurl,
    ),
  ],
  ...
)
```

---

## Day 2: Request Body Tab Widget

### Objective
Create widget for Request body tab (Form Data, Body with view modes, Cookies)

### File to Create
`lib/features/http_inspector/presentation/widgets/request_body_tab.dart`

### Code to Extract
**From**: `http_details_panel.dart`
- Lines 739-747: Form Data section
- Lines 749-790: Body with view mode chips
- Lines 720-737: Cookies section

### Widget Structure
```dart
import 'package:flutter/material.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';
import 'form_data_view.dart';

class RequestBodyTab extends StatelessWidget {
  const RequestBodyTab({
    super.key,
    required this.req,
    required this.reqFrame,
    required this.body,
    required this.fullBody,
    required this.isBodyLoading,
    required this.controller,
    required this.contentType,
    required this.hasFormObj,
    required this.isFormCt,
    required this.cookies,
    required this.analyzeBody,
    required this.buildBodyViewChips,
    required this.renderBodyContentAsSliver,
  });

  final Map<String, dynamic>? req;
  final Map<String, dynamic>? reqFrame;
  final String body;
  final String? fullBody;
  final bool isBodyLoading;
  final BodyViewController controller;
  final String contentType;
  final bool hasFormObj;
  final bool isFormCt;
  final List<Map<String, String>> cookies;

  // Callbacks
  final ContentAnalysisResult Function(String, String?) analyzeBody;
  final Widget Function({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    String? title,
    String? contentType,
    String? baseUrl,
    String? frameId,
    int? bodySize,
  }) buildBodyViewChips;
  final List<Widget> Function({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    required String? contentType,
    required bool isRequest,
    String? baseUrl,
    String? frameId,
    int? bodySize,
  }) renderBodyContentAsSliver;

  @override
  Widget build(BuildContext context) {
    final hideRawBody = hasFormObj || isFormCt;

    return CustomScrollView(
      slivers: [
        // Form Data section
        if (hasFormObj) ...[
          SliverToBoxAdapter(
            child: FormDataView(
              form: (req?['form'] is Map)
                  ? (req!['form'] as Map).cast<String, dynamic>()
                  : null,
              contentType: contentType,
              rawBody: body,
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 6)),
        ],

        // Body content with view chips
        if (body.isNotEmpty && !hideRawBody) ...[
          ...() {
            final bodyForAnalysis = fullBody ?? body;
            final analysis = analyzeBody(bodyForAnalysis, contentType);
            return [
              // Header with view chips
              SliverToBoxAdapter(
                child: buildBodyViewChips(
                  body: body,
                  controller: controller,
                  analysis: analysis,
                  title: 'Request Body',
                  contentType: contentType,
                  frameId: reqFrame?['frame']?['id']?.toString(),
                  bodySize: reqFrame?['frame']?['size'] as int?,
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 6)),
              // Body content
              ...renderBodyContentAsSliver(
                body: body,
                controller: controller,
                analysis: analysis,
                contentType: contentType,
                isRequest: true,
                frameId: reqFrame?['frame']?['id']?.toString(),
                bodySize: reqFrame?['frame']?['size'] as int?,
              ),
            ];
          }(),
        ],

        // Cookies section
        if (cookies.isNotEmpty) ...[
          const SliverToBoxAdapter(child: SizedBox(height: 8)),
          SliverToBoxAdapter(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Cookies', style: Theme.of(context).textTheme.titleSmall),
                const SizedBox(height: 4),
                ...cookies.map(
                  (c) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 2),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          flex: 2,
                          child: SelectableText(
                            c['name'] ?? '',
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          flex: 3,
                          child: SelectableText(
                            c['value'] ?? '',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }
}
```

### Testing Checklist
- [ ] Widget file created
- [ ] Compiles without errors
- [ ] Form Data displays when present
- [ ] Body view chips render
- [ ] All body view modes work (Pretty, Raw, Hex, etc.)
- [ ] Cookies display when present
- [ ] hideRawBody logic works (hides body when form data shown)
- [ ] Loading state shows spinner
- [ ] Null handling works

### Parameters Required from Parent
(See widget structure above)

### Integration Point
Will be used in `TabBarView` for Request:
```dart
TabBarView(
  controller: _reqTabController,
  children: [
    RequestBodyTab(...),
    RequestInfoTab(...),
  ],
)
```

---

## Day 3: Request Info Tab Widget

### Objective
Create widget for Request info tab (Query Params, Headers)

### File to Create
`lib/features/http_inspector/presentation/widgets/request_info_tab.dart`

### Code to Extract
**From**: `http_details_panel.dart`
- Lines 699-718: Query Params section
- Lines 792-810: Headers section

### Widget Structure
```dart
import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';

class RequestInfoTab extends StatelessWidget {
  const RequestInfoTab({
    super.key,
    required this.queryParams,
    required this.headers,
    required this.headersRaw,
  });

  final Map<String, List<String>> queryParams;
  final Map<String, String> headers;
  final Map<String, String> headersRaw;

  @override
  Widget build(BuildContext context) {
    return CustomScrollView(
      slivers: [
        // Query Params section
        if (queryParams.isNotEmpty) ...[
          SliverToBoxAdapter(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Query Params', style: context.appText.subtitle),
                const SizedBox(height: 4),
                ...queryParams.entries.map(
                  (e) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 2),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          flex: 2,
                          child: SelectableText(
                            e.key,
                            style: context.appText.bodySmall.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          flex: 3,
                          child: SelectableText(
                            e.value.join(', '),
                            style: context.appText.bodySmall,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 12)),
        ],

        // Headers section
        SliverToBoxAdapter(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Headers', style: context.appText.subtitle),
              const SizedBox(height: 4),
              ...headers.entries.map(
                (e) => _CopyableKeyValueItem(
                  name: e.key,
                  value: e.value,
                  raw: headersRaw[e.key],
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

// Helper widget for copyable header rows (extract from parent if not already separate)
class _CopyableKeyValueItem extends StatefulWidget {
  const _CopyableKeyValueItem({
    required this.name,
    required this.value,
    this.raw,
  });

  final String name;
  final String value;
  final String? raw;

  @override
  State<_CopyableKeyValueItem> createState() => _CopyableKeyValueItemState();
}

class _CopyableKeyValueItemState extends State<_CopyableKeyValueItem> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 2),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 2,
              child: SelectableText(
                widget.name,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              flex: 3,
              child: SelectableText(
                widget.value,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
            if (_hover && widget.raw != null)
              IconButton(
                icon: const Icon(Icons.copy, size: 14),
                onPressed: () {
                  Clipboard.setData(ClipboardData(text: widget.raw!));
                },
                tooltip: 'Copy raw value',
              ),
          ],
        ),
      ),
    );
  }
}
```

### Testing Checklist
- [ ] Widget file created
- [ ] Compiles without errors
- [ ] Query Params display correctly
- [ ] Headers display correctly
- [ ] Hover on header shows copy button
- [ ] Copy raw value works
- [ ] Empty query params handled (section hidden)
- [ ] Masked headers show masked values

### Parameters Required from Parent
- `Map<String, List<String>> queryParams` - Parsed query parameters
- `Map<String, String> headers` - Display headers (may be masked)
- `Map<String, String> headersRaw` - Raw unmasked headers

### Integration Point
Used in Request TabBarView (see Day 2)

---

## Day 4: Response Body Tab Widget

### Objective
Create widget for Response body tab

### File to Create
`lib/features/http_inspector/presentation/widgets/response_body_tab.dart`

### Code to Extract
**From**: `http_details_panel.dart`
- Lines 930-940: Error banner
- Lines 942-984: Body with view mode chips

### Widget Structure
```dart
import 'package:flutter/material.dart';
import '../../../inspector/presentation/utils/body_view_mode.dart';
import '../../../inspector/presentation/utils/body_content_analyzer.dart';

class ResponseBodyTab extends StatelessWidget {
  const ResponseBodyTab({
    super.key,
    required this.resp,
    required this.respFrame,
    required this.body,
    required this.fullBody,
    required this.isBodyLoading,
    required this.controller,
    required this.contentType,
    required this.baseUrl,
    required this.errorMessage,
    required this.analyzeBody,
    required this.buildBodyViewChips,
    required this.renderBodyContentAsSliver,
  });

  final Map<String, dynamic>? resp;
  final Map<String, dynamic>? respFrame;
  final String body;
  final String? fullBody;
  final bool isBodyLoading;
  final BodyViewController controller;
  final String contentType;
  final String baseUrl;
  final String? errorMessage;

  // Callbacks (same as RequestBodyTab)
  final ContentAnalysisResult Function(String, String?) analyzeBody;
  final Widget Function({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    String? title,
    String? contentType,
    String? baseUrl,
    String? frameId,
    int? bodySize,
  }) buildBodyViewChips;
  final List<Widget> Function({
    required String body,
    required BodyViewController controller,
    required ContentAnalysisResult analysis,
    required String? contentType,
    required bool isRequest,
    String? baseUrl,
    String? frameId,
    int? bodySize,
  }) renderBodyContentAsSliver;

  @override
  Widget build(BuildContext context) {
    final status = (resp?['status'] ?? 0) as int;

    return CustomScrollView(
      slivers: [
        // Error banner
        if (status == 0 && errorMessage != null && errorMessage!.isNotEmpty)
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(
                'Transport Error: $errorMessage',
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                  color: Theme.of(context).colorScheme.error,
                ),
              ),
            ),
          ),

        // Body content
        if (body.isNotEmpty) ...[
          ...() {
            final bodyForAnalysis = fullBody ?? body;
            final analysis = analyzeBody(bodyForAnalysis, contentType);
            return [
              // Header with view chips
              SliverToBoxAdapter(
                child: Row(
                  children: [
                    Text('Body', style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(width: 12),
                    Expanded(
                      child: buildBodyViewChips(
                        body: body,
                        controller: controller,
                        analysis: analysis,
                        title: 'Response Body',
                        contentType: contentType,
                        baseUrl: baseUrl.isNotEmpty ? baseUrl : null,
                        frameId: respFrame?['frame']?['id']?.toString(),
                        bodySize: respFrame?['frame']?['size'] as int?,
                      ),
                    ),
                  ],
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 6)),
              // Body content
              ...renderBodyContentAsSliver(
                body: body,
                controller: controller,
                analysis: analysis,
                contentType: contentType,
                isRequest: false,
                baseUrl: baseUrl.isNotEmpty ? baseUrl : null,
                frameId: respFrame?['frame']?['id']?.toString(),
                bodySize: respFrame?['frame']?['size'] as int?,
              ),
            ];
          }(),
        ],
      ],
    );
  }
}
```

### Testing Checklist
- [ ] Widget file created
- [ ] Compiles without errors
- [ ] Error banner shows when status == 0
- [ ] Body view chips render
- [ ] All body view modes work
- [ ] Empty body handled
- [ ] Loading state works
- [ ] Null handling works

### Parameters Required from Parent
(See widget structure above)

### Integration Point
Used in Response TabBarView

---

## Day 5: Response Info Tab Widget

### Objective
Create widget for Response info tab (Headers, Security, Cache & CORS)

### File to Create
`lib/features/http_inspector/presentation/widgets/response_info_tab.dart`

### Code to Extract
**From**: `http_details_panel.dart`
- Lines 987-1005: Headers section
- Lines 1007-1016: Security section
- Lines 1019-1085: Cache & CORS section

### Widget Structure
```dart
import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';

class ResponseInfoTab extends StatelessWidget {
  const ResponseInfoTab({
    super.key,
    required this.headers,
    required this.headersRaw,
    required this.resp,
    required this.cacheMetadata,
    required this.corsMetadata,
    required this.buildSecurityRows,
  });

  final Map<String, String> headers;
  final Map<String, String> headersRaw;
  final Map<String, dynamic>? resp;
  final Map<String, dynamic> cacheMetadata;
  final Map<String, dynamic> corsMetadata;
  final List<Widget> Function(Map<String, dynamic>?, Map<String, String>) buildSecurityRows;

  @override
  Widget build(BuildContext context) {
    return CustomScrollView(
      slivers: [
        // Headers
        SliverToBoxAdapter(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Headers', style: context.appText.subtitle),
              const SizedBox(height: 4),
              ...headers.entries.map(
                (e) => _CopyableKeyValueItem(
                  name: e.key,
                  value: e.value,
                  raw: headersRaw[e.key],
                ),
              ),
            ],
          ),
        ),
        const SliverToBoxAdapter(child: SizedBox(height: 12)),

        // Security
        SliverToBoxAdapter(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Security', style: context.appText.subtitle),
              const SizedBox(height: 6),
              ...buildSecurityRows(resp, headers),
            ],
          ),
        ),
        const SliverToBoxAdapter(child: SizedBox(height: 12)),

        // Cache & CORS
        SliverToBoxAdapter(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Cache & CORS', style: context.appText.subtitle),
              const SizedBox(height: 6),
              Wrap(
                spacing: 8,
                runSpacing: 4,
                children: [
                  if (cacheMetadata['status'] != null)
                    _chip(context, 'cache: ${cacheMetadata['status']}'),
                  _chip(context, corsMetadata['ok'] == true ? 'CORS OK' : 'CORS Fail'),
                  if ((headers['Vary'] ?? headers['vary']) != null)
                    _chip(context, 'Vary: ${(headers['Vary'] ?? headers['vary'])}'),
                ],
              ),
              const SizedBox(height: 6),
            ],
          ),
        ),

        // Cache table
        if (cacheMetadata.isNotEmpty) ..._buildCacheTable(context),

        // CORS table
        if (corsMetadata.isNotEmpty) ..._buildCorsTable(context),
      ],
    );
  }

  Widget _chip(BuildContext context, String text) {
    return Chip(
      label: Text(text),
      labelStyle: Theme.of(context).textTheme.bodySmall,
      padding: EdgeInsets.zero,
      visualDensity: VisualDensity.compact,
    );
  }

  List<Widget> _buildCacheTable(BuildContext context) {
    // Extract from parent (lines ~1040-1070)
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 8),
            Text('Cache Details', style: context.appText.subtitle),
            const SizedBox(height: 4),
            // Table with cache metadata...
          ],
        ),
      ),
    ];
  }

  List<Widget> _buildCorsTable(BuildContext context) {
    // Extract from parent (lines ~1070-1085)
    return [
      SliverToBoxAdapter(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 8),
            Text('CORS Details', style: context.appText.subtitle),
            const SizedBox(height: 4),
            // Table with CORS metadata...
          ],
        ),
      ),
    ];
  }
}

// Reuse _CopyableKeyValueItem from Day 3
```

### Testing Checklist
- [ ] Widget file created
- [ ] Compiles without errors
- [ ] Headers display
- [ ] Security rows display (TLS, cookies)
- [ ] Cache chips display
- [ ] CORS chips display
- [ ] Cache table displays when metadata exists
- [ ] CORS table displays when metadata exists
- [ ] Null handling works

### Parameters Required from Parent
- `Map<String, String> headers` - Response headers
- `Map<String, String> headersRaw` - Raw headers
- `Map<String, dynamic>? resp` - Full response (for security info)
- `Map<String, dynamic> cacheMetadata` - Computed cache metadata
- `Map<String, dynamic> corsMetadata` - Computed CORS metadata
- `List<Widget> Function(...) buildSecurityRows` - Security rows builder

### Integration Point
Used in Response TabBarView

---

## Week 1 Completion Criteria

### Code Quality
- [ ] All 5 widget files created
- [ ] All widgets compile without errors
- [ ] No lint warnings
- [ ] Consistent code style (formatted with `dart format`)
- [ ] Proper null handling in all widgets

### Functionality
- [ ] Each widget works in isolation with mock data
- [ ] All callbacks fire correctly
- [ ] No missing UI elements from original
- [ ] No visual regressions

### Documentation
- [ ] Each widget has clear doc comments
- [ ] Parameters documented with inline comments
- [ ] Integration points noted

### Testing
- [ ] Manual testing with real HTTP session data
- [ ] Edge cases tested (null data, empty lists, etc.)
- [ ] No console errors

---

## Rollback Strategy

If any widget creation fails:
1. Delete the problematic widget file
2. Document the issue
3. Review plan for that day
4. Adjust approach and retry

**Safe Point**: After each day, commit working code:
```bash
git add lib/features/http_inspector/presentation/widgets/<widget_name>.dart
git commit -m "feat: extract <widget_name> widget"
```

---

## Next Week Preview

Week 2 will integrate these widgets into the main `http_details_panel.dart`:
- Update `_Card` widget to accept actions and tabs
- Refactor `_buildRequest` to use new widgets
- Refactor `_buildResponse` to use new widgets
- Wire up all callbacks
