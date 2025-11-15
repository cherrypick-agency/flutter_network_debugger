# Week 2: HTTP Inspector Integration

**Goal**: Integrate Week 1 widgets into main HTTP details panel with tabs

**Duration**: 5 days

**Success Criteria**:
- Main `http_details_panel.dart` refactored to use new widgets
- Request and Response have working Body/Info tabs
- All buttons moved to card headers
- No functionality lost
- File reduced from ~2373 lines to ~1500 lines

---

## Day 1: Enhance _Card Widget & Add Actions

### Objective
Update `_Card` widget to display action buttons in header row alongside title

### File to Modify
`lib/features/http_inspector/presentation/widgets/http_details_panel.dart`

### Changes Required

#### 1. Update _Card Widget (lines 2014-2067)
**Current code** already has tabs support, need to finalize actions:

```dart
class _Card extends StatelessWidget {
  const _Card({
    required this.title,
    required this.child,
    this.actions,           // ← Already exists (line 2018)
    this.tabController,     // ← Already exists (line 2019)
    this.tabs,              // ← Already exists (line 2020)
  });

  final String title;
  final Widget child;
  final List<Widget>? actions;
  final TabController? tabController;
  final List<Tab>? tabs;

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
            // MODIFY THIS ROW (lines 2038-2050)
            Row(
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (actions != null) ...[
                  const Spacer(),  // ← Push actions to the right
                  ...actions!,
                ],
              ],
            ),
            if (tabs != null && tabController != null) ...[
              const SizedBox(height: 8),
              TabBar(
                controller: tabController,
                tabs: tabs!,
                labelColor: Theme.of(context).colorScheme.primary,
                indicatorSize: TabBarIndicatorSize.tab,
              ),
            ],
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}
```

**Key Change**: Add `const Spacer()` before `...actions!` to push buttons to right side of header

#### 2. Update _ResponseCard Widget (lines 2070-2162)

Add same support for tabs and actions:

```dart
class _ResponseCard extends StatelessWidget {
  const _ResponseCard({
    required this.resp,
    required this.durationMs,
    required this.respTtfbMs,
    required this.respTotalMs,
    required this.child,
    this.tabController,      // ← ADD THIS
    this.tabs,               // ← ADD THIS
  });

  final Map<String, dynamic>? resp;
  final int? durationMs;
  final int? respTtfbMs;
  final int? respTotalMs;
  final Widget child;
  final TabController? tabController;  // ← ADD THIS
  final List<Tab>? tabs;               // ← ADD THIS

  @override
  Widget build(BuildContext context) {
    final status = (resp?['status'] ?? 0) as int;

    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Title row with status and timing
            Row(
              children: [
                Text(
                  'Response',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (status > 0) ...[
                  const SizedBox(width: 8),
                  _StatusChip(status: status),
                ],
                if (durationMs != null) ...[
                  const SizedBox(width: 8),
                  _DurationChip(durationMs: durationMs!),
                ],
                // Add timing info chips...
              ],
            ),
            // ADD TABS SUPPORT (same as _Card)
            if (tabs != null && tabController != null) ...[
              const SizedBox(height: 8),
              TabBar(
                controller: tabController,
                tabs: tabs!,
                labelColor: Theme.of(context).colorScheme.primary,
                indicatorSize: TabBarIndicatorSize.tab,
              ),
            ],
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}
```

### Testing Checklist
- [ ] `_Card` widget shows actions in header
- [ ] Actions aligned to right with Spacer
- [ ] Tabs render below title row
- [ ] `_ResponseCard` supports tabs
- [ ] No visual regressions
- [ ] dart format passes
- [ ] flutter analyze passes

### Rollback Point
```bash
git add lib/features/http_inspector/presentation/widgets/http_details_panel.dart
git commit -m "feat: add actions and tabs support to Card widgets"
```

---

## Day 2: Extract Helper Methods & Typedefs

### Objective
Create typedef aliases for callbacks and prepare helper extraction methods

### File to Modify
`lib/features/http_inspector/presentation/widgets/http_details_panel.dart`

### Changes Required

#### 1. Add Typedefs (after imports, before class)

```dart
// Callback type aliases for body rendering
typedef BodyAnalyzer = ContentAnalysisResult Function(
  String body,
  String? contentType,
);

typedef BodyViewChipsBuilder = Widget Function({
  required String body,
  required BodyViewController controller,
  required ContentAnalysisResult analysis,
  String? title,
  String? contentType,
  String? baseUrl,
  String? frameId,
  int? bodySize,
});

typedef BodyContentSliverRenderer = List<Widget> Function({
  required String body,
  required BodyViewController controller,
  required ContentAnalysisResult analysis,
  required String? contentType,
  required bool isRequest,
  String? baseUrl,
  String? frameId,
  int? bodySize,
});

typedef SecurityRowsBuilder = List<Widget> Function(
  Map<String, dynamic>? resp,
  Map<String, String> headers,
);
```

#### 2. Create Helper Methods for Request Data Extraction

Add method to extract and prepare Request data (before `_buildRequest`):

```dart
/// Extract and prepare all Request data for tabs
({
  Map<String, String> headers,
  Map<String, String> headersRaw,
  String body,
  String url,
  String? normalizedUrl,
  Map<String, List<String>> queryParams,
  List<Map<String, String>> cookies,
  String contentType,
  bool hasFormObj,
  bool isFormCt,
  bool hideRawBody,
}) _extractRequestData(Map<String, dynamic>? req) {
  if (req == null) {
    return (
      headers: {},
      headersRaw: {},
      body: '',
      url: '',
      normalizedUrl: null,
      queryParams: {},
      cookies: [],
      contentType: '',
      hasFormObj: false,
      isFormCt: false,
      hideRawBody: false,
    );
  }

  // Extract all data (move code from lines 572-622)
  final headers = ((req['headers'] ?? <String, dynamic>{}) as Map)
      .cast<String, dynamic>()
      .map((k, v) => MapEntry(k, v.toString()));

  // ... rest of extraction logic

  return (
    headers: headers,
    headersRaw: headersRaw,
    body: body,
    url: url,
    normalizedUrl: normalizedUrl,
    queryParams: qp,
    cookies: reqCookies,
    contentType: ctHeader,
    hasFormObj: hasFormObj,
    isFormCt: isFormCt,
    hideRawBody: hideRawBody,
  );
}
```

#### 3. Create Helper Methods for Response Data Extraction

```dart
/// Extract and prepare all Response data for tabs
({
  Map<String, String> headers,
  Map<String, String> headersRaw,
  String body,
  String contentType,
  String? errorMessage,
  Map<String, dynamic> cacheMetadata,
  Map<String, dynamic> corsMetadata,
}) _extractResponseData(Map<String, dynamic>? resp) {
  // Extract all data (move code from lines 880-916)
  // ...
}
```

### Testing Checklist
- [ ] All typedefs compile
- [ ] Helper methods compile
- [ ] Helper methods return correct data
- [ ] No functionality changed (just extraction)
- [ ] dart format passes
- [ ] flutter analyze passes

### Rollback Point
```bash
git commit -m "refactor: extract data preparation helpers"
```

---

## Day 3: Refactor _buildRequest with Tabs

### Objective
Replace `_buildRequest` CustomScrollView with TabBarView using Week 1 widgets

### File to Modify
`lib/features/http_inspector/presentation/widgets/http_details_panel.dart`

### Changes Required

#### 1. Import Week 1 Widgets (at top of file)

```dart
import 'request_action_buttons.dart';
import 'request_body_tab.dart';
import 'request_info_tab.dart';
```

#### 2. Refactor _buildRequest Method (lines 507-813)

**BEFORE** (current):
```dart
Widget _buildRequest(...) {
  if (req == null) { /* null handling */ }

  // 300 lines of data extraction and UI building
  return CustomScrollView(
    slivers: [
      // Buttons
      // Query Params
      // Form Data
      // Body
      // Headers
    ],
  );
}
```

**AFTER** (new):
```dart
Widget _buildRequest(
  BuildContext context,
  Map<String, dynamic>? req,
  Map<String, dynamic>? reqFrame,
) {
  if (req == null) {
    // Keep existing null handling (lines 512-575)
    // ...
    return /* existing placeholder */;
  }

  // Extract data using helper
  final data = _extractRequestData(req);

  // Return TabBarView with new widgets
  return TabBarView(
    controller: _reqTabController,
    children: [
      // Body Tab
      RequestBodyTab(
        req: req,
        reqFrame: reqFrame,
        body: data.body,
        fullBody: _fullReqBody,
        isBodyLoading: _reqBodyLoading,
        controller: _reqViewController,
        contentType: data.contentType,
        hasFormObj: data.hasFormObj,
        isFormCt: data.isFormCt,
        cookies: data.cookies,
        analyzeBody: _analyzeRequestBody,
        buildBodyViewChips: _buildBodyViewChips,
        renderBodyContentAsSliver: _renderBodyContentAsSliver,
      ),

      // Info Tab
      RequestInfoTab(
        queryParams: data.queryParams,
        headers: data.headers,
        headersRaw: data.headersRaw,
      ),
    ],
  );
}
```

#### 3. Update Card Usage (lines 486-489)

**BEFORE**:
```dart
_Card(
  title: 'Request',
  child: _buildRequest(context, req, reqFrame),
),
```

**AFTER**:
```dart
() {
  final data = _extractRequestData(req);
  return _Card(
    title: 'Request',
    actions: req != null ? [
      RequestActionButtons(
        req: req,
        url: data.url,
        normalizedUrl: data.normalizedUrl,
        loadingFetch: _loadingFetch,
        onRepeat: () => _refetchOriginalRequest(context, req),
        buildCurl: _buildCurl,
      ),
    ] : null,
    tabController: _reqTabController,
    tabs: const [
      Tab(text: 'Body'),
      Tab(text: 'Info'),
    ],
    child: _buildRequest(context, req, reqFrame),
  );
}(),
```

### Testing Checklist
- [ ] Request tabs render
- [ ] Body tab shows Form Data + Body + Cookies
- [ ] Info tab shows Query Params + Headers
- [ ] Tab switching works smoothly
- [ ] Action buttons in header work
- [ ] Copy URL works
- [ ] Copy cURL dropdown works
- [ ] Repeat button works
- [ ] All view mode chips work
- [ ] No visual regressions
- [ ] dart format passes
- [ ] flutter analyze passes

### Rollback Point
```bash
git commit -m "refactor: convert Request to tabs with new widgets"
```

---

## Day 4: Refactor _buildResponse with Tabs

### Objective
Replace `_buildResponse` CustomScrollView with TabBarView using Week 1 widgets

### File to Modify
`lib/features/http_inspector/presentation/widgets/http_details_panel.dart`

### Changes Required

#### 1. Import Week 1 Widgets

```dart
import 'response_body_tab.dart';
import 'response_info_tab.dart';
```

#### 2. Refactor _buildResponse Method (lines 815-1087)

**AFTER** (new):
```dart
Widget _buildResponse(
  BuildContext context,
  Map<String, dynamic>? resp,
  Map<String, dynamic>? respFrame,
) {
  if (resp == null) {
    // Keep existing null handling
    // ...
    return /* existing placeholder */;
  }

  // Extract data using helper
  final data = _extractResponseData(resp);

  // Compute cache and CORS metadata (keep existing logic)
  final cache = _computeCacheMeta(data.headers);
  final cors = _computeCorsMeta(data.headers);

  // Get URL for baseUrl
  final req = _findByType(widget.frames, 'http_request');
  final url = (req?['url'] ?? '').toString();

  // Return TabBarView with new widgets
  return TabBarView(
    controller: _respTabController,
    children: [
      // Body Tab
      ResponseBodyTab(
        resp: resp,
        respFrame: respFrame,
        body: data.body,
        fullBody: _fullRespBody,
        isBodyLoading: _respBodyLoading,
        controller: _respViewController,
        contentType: data.contentType,
        baseUrl: url,
        errorMessage: data.errorMessage,
        analyzeBody: _analyzeResponseBody,
        buildBodyViewChips: _buildBodyViewChips,
        renderBodyContentAsSliver: _renderBodyContentAsSliver,
      ),

      // Info Tab
      ResponseInfoTab(
        headers: data.headers,
        headersRaw: data.headersRaw,
        resp: resp,
        cacheMetadata: cache,
        corsMetadata: cors,
        buildSecurityRows: _securityRows,
      ),
    ],
  );
}
```

#### 3. Update Card Usage (lines 492-499)

**AFTER**:
```dart
_ResponseCard(
  resp: resp,
  durationMs: durationMs,
  respTtfbMs: _respTtfbMs,
  respTotalMs: _respTotalMs,
  tabController: _respTabController,
  tabs: const [
    Tab(text: 'Body'),
    Tab(text: 'Info'),
  ],
  child: _buildResponse(context, resp, respFrame),
),
```

### Testing Checklist
- [ ] Response tabs render
- [ ] Body tab shows body content
- [ ] Info tab shows Headers + Security + Cache/CORS
- [ ] Tab switching works
- [ ] Status and timing chips show in header
- [ ] Error banner shows when status == 0
- [ ] All view mode chips work
- [ ] Security rows display TLS and cookies
- [ ] Cache chips display
- [ ] CORS chips display
- [ ] No visual regressions
- [ ] dart format passes
- [ ] flutter analyze passes

### Rollback Point
```bash
git commit -m "refactor: convert Response to tabs with new widgets"
```

---

## Day 5: Final Integration & Cleanup

### Objective
Clean up unused code, verify all functionality, fix edge cases

### Tasks

#### 1. Remove Unused Code from http_details_panel.dart

Now that widgets are extracted, remove:
- Lines 623-697: Button slivers (moved to `RequestActionButtons`)
- Lines 699-737: Query Params and Cookies slivers (moved to tabs)
- Lines 739-790: Form Data and Body slivers (moved to `RequestBodyTab`)
- Lines 792-810: Headers sliver (moved to `RequestInfoTab`)
- Lines 930-984: Response body slivers (moved to `ResponseBodyTab`)
- Lines 987-1085: Response info slivers (moved to `ResponseInfoTab`)

**Expected line reduction**: ~2373 → ~1500 lines

#### 2. Verify All Imports

Ensure all new widget imports are at top:
```dart
import 'request_action_buttons.dart';
import 'request_body_tab.dart';
import 'request_info_tab.dart';
import 'response_body_tab.dart';
import 'response_info_tab.dart';
```

#### 3. Test Edge Cases

- [ ] Null request (shows placeholder)
- [ ] Null response (shows placeholder)
- [ ] Empty body (Body tab shows nothing)
- [ ] No query params (Info tab hides section)
- [ ] No cookies (Body tab hides section)
- [ ] CONNECT method (shows special placeholder)
- [ ] Transport error (shows error banner)
- [ ] Loading state (shows spinner)
- [ ] Masked headers (Copy shows raw value)
- [ ] All view modes (Pretty, Raw, Hex, JWT, HTML, GraphQL)
- [ ] Form data display
- [ ] Cache metadata table
- [ ] CORS metadata table
- [ ] TLS info display
- [ ] Cookie info display

#### 4. Performance Check

- [ ] No lag when switching tabs
- [ ] View mode changes are instant
- [ ] Copy operations don't freeze UI
- [ ] Repeat request shows modal without delay
- [ ] Large bodies (>1MB) render smoothly
- [ ] Hex viewer interactive in normal mode

#### 5. Code Quality

- [ ] Run `dart format` on all files
- [ ] Run `flutter analyze` - fix all warnings
- [ ] Check for unused imports
- [ ] Verify all typedefs used
- [ ] Remove dead code
- [ ] Add doc comments where missing

### Testing Checklist
- [ ] All edge cases pass
- [ ] No performance regressions
- [ ] No visual regressions
- [ ] Code quality checks pass
- [ ] Main file reduced to ~1500 lines
- [ ] All 5 new widgets working

### Final Rollback Point
```bash
git add .
git commit -m "refactor(http-inspector): complete tabs integration

- Extract 5 new widgets for Request and Response sections
- Add tabs (Body/Info) to Request and Response
- Move action buttons to card headers
- Reduce main file from 2373 to ~1500 lines
- All functionality preserved, no regressions

BREAKING CHANGE: None - internal refactor only"
```

---

## Week 2 Completion Criteria

### Code Structure
- [ ] Main file reduced from 2373 → ~1500 lines
- [ ] 5 new widget files (from Week 1) integrated
- [ ] Clear separation of concerns
- [ ] No code duplication

### Functionality
- [ ] All original features work
- [ ] Request tabs (Body/Info) work
- [ ] Response tabs (Body/Info) work
- [ ] Action buttons in headers work
- [ ] All callbacks fire correctly
- [ ] No data loss or display errors

### User Experience
- [ ] Tabs switch smoothly
- [ ] No visual regressions
- [ ] Performance same or better
- [ ] All interactions work (copy, repeat, view modes)

### Code Quality
- [ ] All files formatted
- [ ] No lint warnings
- [ ] Proper null handling
- [ ] Type-safe callbacks (using typedefs)

---

## Risk Mitigation

### Potential Issues

1. **Callback Hell**: Too many function parameters
   - **Solution**: Use typedefs for clarity
   - **Solution**: Group related callbacks into objects if needed

2. **State Synchronization**: Tabs and view modes out of sync
   - **Solution**: Keep `BodyViewController` in parent state
   - **Solution**: Pass controllers explicitly to children

3. **Performance**: Multiple rebuilds on tab switch
   - **Solution**: Use `const` constructors where possible
   - **Solution**: Memoize expensive computations

4. **Data Flow**: Child widgets need parent state
   - **Solution**: Pass data explicitly via parameters
   - **Solution**: Use callbacks for state updates (not direct setState)

### Rollback Strategy

If integration breaks:
1. **Immediate**: Revert to last commit
2. **Identify**: Which day's changes caused issue
3. **Fix**: Address specific issue in isolation
4. **Re-test**: Validate fix with comprehensive tests
5. **Continue**: Resume from fixed point

---

## Next Week Preview

Week 3 will focus on comprehensive testing:
- Unit tests for edge cases
- Integration tests for data flow
- End-to-end tests for user workflows
- Performance testing with large datasets
- Bug fixes and polish
