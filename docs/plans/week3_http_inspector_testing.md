# Week 3: HTTP Inspector Testing & Polish

**Goal**: Comprehensive testing, bug fixes, and performance validation
**Duration**: Days 11-15
**Success Criteria**: All tests pass, no regressions, smooth user experience

---

## Day 11: Unit Tests for Extracted Widgets

### Objectives
1. Create test files for all 5 extracted widgets
2. Test widget rendering with various data states
3. Test callback behavior and user interactions
4. Achieve 90%+ code coverage for widget layer

### Test Files to Create

#### `test/features/http_inspector/presentation/widgets/request_action_buttons_test.dart`

```dart
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_proxy/features/http_inspector/presentation/widgets/request_action_buttons.dart';

void main() {
  group('RequestActionButtons', () {
    testWidgets('renders all action buttons', (tester) async {
      bool copyUrlCalled = false;
      bool copyCurlCalled = false;
      bool repeatCalled = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'GET',
              headers: {},
              body: '',
              onCopyUrl: () => copyUrlCalled = true,
              onCopyCurl: () => copyCurlCalled = true,
              onRepeat: () => repeatCalled = true,
            ),
          ),
        ),
      );

      // Verify Copy URL button exists
      expect(find.byIcon(Icons.copy), findsOneWidget);

      // Verify Copy cURL dropdown menu exists
      expect(find.byIcon(Icons.code), findsOneWidget);

      // Verify Repeat button exists
      expect(find.byIcon(Icons.replay), findsOneWidget);
    });

    testWidgets('Copy URL callback works', (tester) async {
      bool called = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'GET',
              headers: {},
              body: '',
              onCopyUrl: () => called = true,
              onCopyCurl: () {},
              onRepeat: () {},
            ),
          ),
        ),
      );

      await tester.tap(find.byIcon(Icons.copy));
      await tester.pump();

      expect(called, true);
    });

    testWidgets('Copy cURL dropdown shows options', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'POST',
              headers: {'Content-Type': 'application/json'},
              body: '{"key": "value"}',
              onCopyUrl: () {},
              onCopyCurl: () {},
              onRepeat: () {},
            ),
          ),
        ),
      );

      // Open dropdown
      await tester.tap(find.byIcon(Icons.code));
      await tester.pumpAndSettle();

      // Verify options
      expect(find.text('Copy cURL'), findsOneWidget);
      expect(find.text('Copy cURL (bash)'), findsOneWidget);
      expect(find.text('Copy cURL (PowerShell)'), findsOneWidget);
    });

    testWidgets('Repeat button disabled when callback is null', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'GET',
              headers: {},
              body: '',
              onCopyUrl: () {},
              onCopyCurl: () {},
              onRepeat: null, // Disabled
            ),
          ),
        ),
      );

      final button = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.replay),
      );
      expect(button.onPressed, null);
    });
  });
}
```

#### `test/features/http_inspector/presentation/widgets/request_body_tab_test.dart`

```dart
void main() {
  group('RequestBodyTab', () {
    testWidgets('renders form data correctly', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestBodyTab(
              formData: {'username': 'test', 'password': 'secret'},
              body: '',
              cookies: [],
            ),
          ),
        ),
      );

      expect(find.text('Form Data'), findsOneWidget);
      expect(find.text('username'), findsOneWidget);
      expect(find.text('test'), findsOneWidget);
    });

    testWidgets('switches between view modes', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestBodyTab(
              formData: {},
              body: '{"key": "value"}',
              cookies: [],
            ),
          ),
        ),
      );

      // Initial mode should be Pretty
      expect(find.byIcon(Icons.data_object), findsOneWidget);

      // Tap Hex mode
      await tester.tap(find.byIcon(Icons.grid_on));
      await tester.pumpAndSettle();

      // Verify hex view rendered
      expect(find.byType(HexBodyRenderer), findsOneWidget);
    });

    testWidgets('displays cookies when provided', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestBodyTab(
              formData: {},
              body: '',
              cookies: [
                {'name': 'session', 'value': 'abc123'},
                {'name': 'token', 'value': 'xyz789'},
              ],
            ),
          ),
        ),
      );

      expect(find.text('Cookies'), findsOneWidget);
      expect(find.text('session'), findsOneWidget);
      expect(find.text('token'), findsOneWidget);
    });
  });
}
```

#### `test/features/http_inspector/presentation/widgets/request_info_tab_test.dart`

```dart
void main() {
  group('RequestInfoTab', () {
    testWidgets('renders query params correctly', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestInfoTab(
              queryParams: {'search': 'flutter', 'limit': '10'},
              headers: {},
            ),
          ),
        ),
      );

      expect(find.text('Query Parameters'), findsOneWidget);
      expect(find.text('search'), findsOneWidget);
      expect(find.text('flutter'), findsOneWidget);
      expect(find.text('limit'), findsOneWidget);
      expect(find.text('10'), findsOneWidget);
    });

    testWidgets('renders headers correctly', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestInfoTab(
              queryParams: {},
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer token123',
              },
            ),
          ),
        ),
      );

      expect(find.text('Headers'), findsOneWidget);
      expect(find.text('Content-Type'), findsOneWidget);
      expect(find.text('application/json'), findsOneWidget);
      expect(find.text('Authorization'), findsOneWidget);
    });

    testWidgets('shows empty state when no data', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestInfoTab(
              queryParams: {},
              headers: {},
            ),
          ),
        ),
      );

      expect(find.text('No query parameters'), findsOneWidget);
      expect(find.text('No headers'), findsOneWidget);
    });
  });
}
```

#### `test/features/http_inspector/presentation/widgets/response_body_tab_test.dart`

```dart
void main() {
  group('ResponseBodyTab', () {
    testWidgets('renders response body', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseBodyTab(
              body: '{"status": "success"}',
              statusCode: 200,
              error: null,
            ),
          ),
        ),
      );

      expect(find.textContaining('status'), findsOneWidget);
      expect(find.textContaining('success'), findsOneWidget);
    });

    testWidgets('displays error banner for 4xx/5xx', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseBodyTab(
              body: '{"error": "Not found"}',
              statusCode: 404,
              error: 'Resource not found',
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.error_outline), findsOneWidget);
      expect(find.text('Error 404'), findsOneWidget);
    });

    testWidgets('switches between view modes', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseBodyTab(
              body: '<html><body>Hello</body></html>',
              statusCode: 200,
              error: null,
            ),
          ),
        ),
      );

      // Switch to Raw mode
      await tester.tap(find.byIcon(Icons.text_fields));
      await tester.pumpAndSettle();

      expect(find.text('<html><body>Hello</body></html>'), findsOneWidget);
    });
  });
}
```

#### `test/features/http_inspector/presentation/widgets/response_info_tab_test.dart`

```dart
void main() {
  group('ResponseInfoTab', () {
    testWidgets('renders response headers', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseInfoTab(
              headers: {
                'Content-Type': 'application/json',
                'Content-Length': '1234',
              },
              security: {},
              cacheInfo: {},
              corsHeaders: {},
            ),
          ),
        ),
      );

      expect(find.text('Response Headers'), findsOneWidget);
      expect(find.text('Content-Type'), findsOneWidget);
      expect(find.text('Content-Length'), findsOneWidget);
    });

    testWidgets('displays security headers', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseInfoTab(
              headers: {},
              security: {
                'Strict-Transport-Security': 'max-age=31536000',
                'X-Frame-Options': 'DENY',
              },
              cacheInfo: {},
              corsHeaders: {},
            ),
          ),
        ),
      );

      expect(find.text('Security Headers'), findsOneWidget);
      expect(find.text('Strict-Transport-Security'), findsOneWidget);
      expect(find.text('X-Frame-Options'), findsOneWidget);
    });

    testWidgets('displays CORS headers', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: ResponseInfoTab(
              headers: {},
              security: {},
              cacheInfo: {},
              corsHeaders: {
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Methods': 'GET, POST',
              },
            ),
          ),
        ),
      );

      expect(find.text('CORS Headers'), findsOneWidget);
      expect(find.text('Access-Control-Allow-Origin'), findsOneWidget);
    });
  });
}
```

### Testing Checklist
- [ ] All 5 widget test files created
- [ ] Each widget tested with empty/null data
- [ ] Each widget tested with populated data
- [ ] All callbacks tested
- [ ] View mode switching tested
- [ ] Error states tested
- [ ] Run `flutter test` - all tests pass
- [ ] Code coverage ≥ 90%

### Rollback Strategy
If tests reveal critical bugs in widgets:
1. Document failing tests
2. Create bug fix branches
3. Fix one widget at a time
4. Re-run tests until green

---

## Day 12: Integration Tests for Tab Navigation

### Objectives
1. Test Request/Response tab switching
2. Test state persistence across tab changes
3. Test interaction between widgets and parent panel
4. Verify no memory leaks in TabController disposal

### Integration Test File

#### `test/features/http_inspector/presentation/widgets/http_details_panel_integration_test.dart`

```dart
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_proxy/features/http_inspector/presentation/widgets/http_details_panel.dart';

void main() {
  group('HttpDetailsPanel Integration', () {
    testWidgets('switches between Request tabs', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: HttpDetailsPanel(
              sessionId: 'test-session',
              frameId: 'test-frame',
              data: {
                'request': {
                  'method': 'POST',
                  'url': 'https://api.example.com/users',
                  'headers': {'Content-Type': 'application/json'},
                  'body': '{"name": "John"}',
                },
                'response': {
                  'status': 200,
                  'headers': {},
                  'body': '{"id": 1}',
                },
              },
            ),
          ),
        ),
      );

      // Initially on Body tab
      expect(find.text('Body'), findsNWidgets(2)); // Request + Response
      expect(find.text('Info'), findsNWidgets(2));

      // Tap Request Info tab
      final requestInfoTab = find.text('Info').first;
      await tester.tap(requestInfoTab);
      await tester.pumpAndSettle();

      // Verify Info tab content displayed
      expect(find.text('Headers'), findsOneWidget);
      expect(find.text('Content-Type'), findsOneWidget);

      // Switch back to Body tab
      final requestBodyTab = find.text('Body').first;
      await tester.tap(requestBodyTab);
      await tester.pumpAndSettle();

      // Verify Body content displayed
      expect(find.textContaining('name'), findsOneWidget);
    });

    testWidgets('switches between Response tabs', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: HttpDetailsPanel(
              sessionId: 'test-session',
              frameId: 'test-frame',
              data: {
                'request': {
                  'method': 'GET',
                  'url': 'https://api.example.com/health',
                  'headers': {},
                  'body': '',
                },
                'response': {
                  'status': 200,
                  'headers': {
                    'Cache-Control': 'no-cache',
                    'Content-Type': 'application/json',
                  },
                  'body': '{"status": "ok"}',
                },
              },
            ),
          ),
        ),
      );

      // Tap Response Info tab
      final responseInfoTab = find.text('Info').last;
      await tester.tap(responseInfoTab);
      await tester.pumpAndSettle();

      // Verify Info tab content
      expect(find.text('Response Headers'), findsOneWidget);
      expect(find.text('Cache-Control'), findsOneWidget);
    });

    testWidgets('action buttons work from Request header', (tester) async {
      bool copyUrlCalled = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: HttpDetailsPanel(
              sessionId: 'test-session',
              frameId: 'test-frame',
              data: {
                'request': {
                  'method': 'GET',
                  'url': 'https://api.example.com/test',
                  'headers': {},
                  'body': '',
                },
                'response': null,
              },
              onCopyUrl: () => copyUrlCalled = true,
            ),
          ),
        ),
      );

      // Find and tap Copy URL button in Request header
      await tester.tap(find.byIcon(Icons.copy).first);
      await tester.pump();

      expect(copyUrlCalled, true);
    });

    testWidgets('view mode persists across tab switches', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: HttpDetailsPanel(
              sessionId: 'test-session',
              frameId: 'test-frame',
              data: {
                'request': {
                  'method': 'POST',
                  'url': 'https://api.example.com/data',
                  'headers': {},
                  'body': '{"test": true}',
                },
                'response': {
                  'status': 200,
                  'headers': {},
                  'body': '{"result": "success"}',
                },
              },
            ),
          ),
        ),
      );

      // Switch Request body to Hex mode
      await tester.tap(find.byIcon(Icons.grid_on).first);
      await tester.pumpAndSettle();

      expect(find.byType(HexBodyRenderer), findsOneWidget);

      // Switch to Info tab
      await tester.tap(find.text('Info').first);
      await tester.pumpAndSettle();

      // Switch back to Body tab
      await tester.tap(find.text('Body').first);
      await tester.pumpAndSettle();

      // Verify still in Hex mode
      expect(find.byType(HexBodyRenderer), findsOneWidget);
    });

    testWidgets('disposes TabControllers properly', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: HttpDetailsPanel(
              sessionId: 'test-session',
              frameId: 'test-frame',
              data: {
                'request': {'method': 'GET', 'url': '', 'headers': {}, 'body': ''},
                'response': null,
              },
            ),
          ),
        ),
      );

      // Trigger disposal
      await tester.pumpWidget(const MaterialApp(home: SizedBox()));

      // No assertion errors should occur
      expect(tester.takeException(), isNull);
    });
  });
}
```

### Testing Checklist
- [ ] Request tab switching works
- [ ] Response tab switching works
- [ ] Action buttons accessible in all tabs
- [ ] View mode state persists
- [ ] TabController disposal works
- [ ] No memory leaks detected
- [ ] Run `flutter test test/features/http_inspector/`
- [ ] All integration tests pass

### Rollback Strategy
If integration issues found:
1. Isolate failing test scenario
2. Check TabController lifecycle
3. Verify callback prop drilling
4. Fix and re-test

---

## Day 13: End-to-End Testing & Performance

### Objectives
1. Test with real API responses (large payloads)
2. Performance test with 100+ HTTP frames
3. Memory usage validation
4. Scroll performance in tabs

### E2E Test Scenarios

#### Scenario 1: Large JSON Response (1MB+)
```dart
testWidgets('handles 1MB JSON response', (tester) async {
  // Generate large JSON
  final largeJson = {
    'data': List.generate(10000, (i) => {
      'id': i,
      'name': 'User $i',
      'email': 'user$i@example.com',
      'metadata': {
        'created': DateTime.now().toIso8601String(),
        'tags': ['tag1', 'tag2', 'tag3'],
      },
    }),
  };

  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: HttpDetailsPanel(
          sessionId: 'perf-test',
          frameId: 'large-json',
          data: {
            'request': {
              'method': 'GET',
              'url': 'https://api.example.com/users',
              'headers': {},
              'body': '',
            },
            'response': {
              'status': 200,
              'headers': {'Content-Type': 'application/json'},
              'body': jsonEncode(largeJson),
            },
          },
        ),
      ),
    ),
  );

  // Should render without freezing
  await tester.pumpAndSettle(const Duration(seconds: 5));

  // Verify JSON rendered
  expect(find.textContaining('data'), findsOneWidget);

  // Switch to Hex mode
  await tester.tap(find.byIcon(Icons.grid_on).last);
  await tester.pumpAndSettle();

  // Should handle hex view
  expect(find.byType(HexBodyRenderer), findsOneWidget);
});
```

#### Scenario 2: Binary File Response (Image)
```dart
testWidgets('handles binary image response', (tester) async {
  // Simulate PNG bytes
  final pngBytes = List.generate(50000, (i) => i % 256);

  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: ResponseBodyTab(
          body: String.fromCharCodes(pngBytes),
          statusCode: 200,
          error: null,
          bodySize: pngBytes.length,
        ),
      ),
    ),
  );

  await tester.pumpAndSettle();

  // Switch to Hex view for binary
  await tester.tap(find.byIcon(Icons.grid_on));
  await tester.pumpAndSettle();

  // Verify hex viewer loaded
  expect(find.byType(HexBodyRenderer), findsOneWidget);
});
```

#### Scenario 3: Rapid Tab Switching
```dart
testWidgets('handles rapid tab switching', (tester) async {
  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: HttpDetailsPanel(
          sessionId: 'rapid-test',
          frameId: 'frame-1',
          data: {
            'request': {
              'method': 'POST',
              'url': 'https://api.example.com/data',
              'headers': {'Content-Type': 'application/json'},
              'body': '{"data": "test"}',
            },
            'response': {
              'status': 200,
              'headers': {'Content-Type': 'application/json'},
              'body': '{"result": "ok"}',
            },
          },
        ),
      ),
    ),
  );

  // Rapidly switch tabs 10 times
  for (int i = 0; i < 10; i++) {
    await tester.tap(find.text('Info').first);
    await tester.pump();
    await tester.tap(find.text('Body').first);
    await tester.pump();
  }

  await tester.pumpAndSettle();

  // No errors should occur
  expect(tester.takeException(), isNull);
});
```

### Performance Benchmarks

Create `test/performance/http_inspector_benchmark.dart`:

```dart
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('benchmark tab switching performance', (tester) async {
    final stopwatch = Stopwatch()..start();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HttpDetailsPanel(
            sessionId: 'benchmark',
            frameId: 'frame-1',
            data: _generateTestData(),
          ),
        ),
      ),
    );

    stopwatch.stop();
    print('Initial render: ${stopwatch.elapsedMilliseconds}ms');

    stopwatch.reset();
    stopwatch.start();

    await tester.tap(find.text('Info').first);
    await tester.pumpAndSettle();

    stopwatch.stop();
    print('Tab switch: ${stopwatch.elapsedMilliseconds}ms');

    // Assert performance targets
    expect(stopwatch.elapsedMilliseconds, lessThan(100)); // <100ms
  });
}

Map<String, dynamic> _generateTestData() {
  return {
    'request': {
      'method': 'GET',
      'url': 'https://api.example.com/data',
      'headers': Map.fromEntries(
        List.generate(50, (i) => MapEntry('X-Custom-$i', 'value-$i')),
      ),
      'body': jsonEncode({
        'items': List.generate(1000, (i) => {'id': i, 'name': 'Item $i'}),
      }),
    },
    'response': {
      'status': 200,
      'headers': Map.fromEntries(
        List.generate(30, (i) => MapEntry('X-Response-$i', 'value-$i')),
      ),
      'body': jsonEncode({
        'results': List.generate(1000, (i) => {'id': i, 'value': i * 2}),
      }),
    },
  };
}
```

### Testing Checklist
- [ ] Test with 1MB+ JSON responses
- [ ] Test with binary file responses
- [ ] Test rapid tab switching (no crashes)
- [ ] Scroll performance in all view modes
- [ ] Memory usage < 100MB for 100 frames
- [ ] Tab switch latency < 100ms
- [ ] No frame drops during scrolling
- [ ] Run performance benchmarks

### Rollback Strategy
If performance issues found:
1. Profile with DevTools
2. Identify bottlenecks (build vs paint)
3. Add `const` constructors where possible
4. Consider lazy loading for large payloads

---

## Day 14: Bug Fixes from Testing

### Objectives
1. Fix all bugs discovered in Days 11-13
2. Address edge cases
3. Improve error handling
4. Code cleanup and optimization

### Expected Bug Categories

#### Category 1: State Management
**Potential Issues**:
- TabController index out of sync
- View mode state lost on widget rebuild
- Memory leaks from undisposed controllers

**Fixes**:
```dart
// Ensure TabController properly initialized
@override
void didUpdateWidget(HttpDetailsPanel oldWidget) {
  super.didUpdateWidget(oldWidget);

  if (oldWidget.frameId != widget.frameId) {
    // Reset tab indices to 0 on new frame
    _reqTabController.index = 0;
    _respTabController.index = 0;
  }
}

// Verify disposal in all code paths
@override
void dispose() {
  _reqTabController.dispose();
  _respTabController.dispose();
  _reqViewController.dispose();
  _respViewController.dispose();
  super.dispose();
}
```

#### Category 2: Edge Cases
**Potential Issues**:
- Null response handling
- Empty body rendering
- Missing headers display

**Fixes**:
```dart
// Handle null response gracefully
if (widget.data['response'] == null) {
  return const Center(
    child: Text('Waiting for response...'),
  );
}

// Handle empty body
final body = widget.data['request']?['body'] ?? '';
if (body.isEmpty) {
  return const Center(
    child: Text('No request body'),
  );
}
```

#### Category 3: Performance
**Potential Issues**:
- Rebuilding entire panel on tab switch
- Large JSON parsing on every render
- Hex viewer recreating state

**Fixes**:
```dart
// Memoize expensive computations
late final Map<String, dynamic>? _parsedRequestBody;

@override
void initState() {
  super.initState();
  final bodyStr = widget.data['request']?['body'] ?? '';
  try {
    _parsedRequestBody = jsonDecode(bodyStr);
  } catch (_) {
    _parsedRequestBody = null;
  }
}

// Use const constructors
const _Card({
  required this.title,
  required this.child,
  this.actions,
  this.tabController,
  this.tabs,
});
```

### Bug Fix Workflow
1. Create issue for each bug found
2. Write failing test case
3. Implement fix
4. Verify test passes
5. Manual smoke test
6. Update documentation

### Testing Checklist
- [ ] All failing tests now pass
- [ ] Edge cases handled
- [ ] Error messages user-friendly
- [ ] Performance targets met
- [ ] No new regressions introduced
- [ ] Code coverage maintained ≥ 90%
- [ ] Run full test suite
- [ ] Manual testing on real data

### Rollback Strategy
If critical bug cannot be fixed:
1. Revert specific commit
2. Document blocker
3. Create alternative approach
4. Re-implement with safeguards

---

## Day 15: Final Testing & Acceptance

### Objectives
1. End-to-end manual testing
2. Regression testing against original features
3. Documentation review
4. Final acceptance checklist

### Manual Test Plan

#### Test Case 1: Basic Request/Response Viewing
1. Open HTTP inspector
2. Select HTTP frame
3. Verify Request section displays:
   - Method, URL, Headers in Info tab
   - Body content in Body tab
   - Action buttons in header (Copy URL, Copy cURL, Repeat)
4. Verify Response section displays:
   - Status code, Headers in Info tab
   - Body content in Body tab
5. Switch between tabs multiple times
6. ✅ No UI glitches or freezes

#### Test Case 2: View Mode Switching
1. Select frame with JSON body
2. In Request Body tab, click Pretty view
3. Click Raw view - should show unformatted JSON
4. Click Tree view - should show expandable tree
5. Click Hex view - should show hex dump
6. Switch to Info tab and back to Body tab
7. ✅ View mode should persist

#### Test Case 3: Query Parameters
1. Select frame with URL containing query params
2. Go to Request Info tab
3. ✅ Query params displayed in table
4. ✅ Each param has key and value

#### Test Case 4: Copy Functions
1. Click Copy URL button
2. ✅ URL copied to clipboard
3. Click Copy cURL dropdown
4. Select "Copy cURL (bash)"
5. ✅ cURL command copied
6. Paste in terminal - should work

#### Test Case 5: Large Payloads
1. Select frame with >1MB response
2. ✅ Loading indicator shown initially
3. ✅ Content renders without freeze
4. Switch to Hex view
5. ✅ Hex viewer loads progressively
6. ✅ Scroll performance smooth

#### Test Case 6: Error Responses
1. Select frame with 404 status
2. ✅ Error banner shown in Response Body tab
3. ✅ Error message clear and helpful

#### Test Case 7: Binary Content
1. Select frame with image response
2. Switch to Hex view
3. ✅ Hex data displayed correctly
4. ✅ Can click/select hex bytes

### Regression Checklist

Verify no breakage of existing features:

- [ ] WebSocket inspector still works
- [ ] Frame search still works
- [ ] Frame filtering still works
- [ ] Copy/Repeat buttons still work
- [ ] Fullscreen mode still works
- [ ] Dark/Light theme switching works
- [ ] Window resize doesn't break layout
- [ ] Multiple sessions loadable
- [ ] Session switching works

### Final Acceptance Criteria

✅ **Code Quality**
- [ ] All tests pass (`flutter test`)
- [ ] No analyzer warnings (`flutter analyze`)
- [ ] Code coverage ≥ 90%
- [ ] Code formatted (`dart format .`)
- [ ] No TODO comments in production code

✅ **Performance**
- [ ] Initial render < 500ms
- [ ] Tab switch < 100ms
- [ ] Handles 100+ HTTP frames
- [ ] Memory usage < 100MB
- [ ] No frame drops during scroll

✅ **Functionality**
- [ ] All 7 manual test cases pass
- [ ] No regressions found
- [ ] Edge cases handled gracefully
- [ ] Error messages helpful

✅ **Code Organization**
- [ ] Main file reduced from 2373 → <1500 lines
- [ ] 5 widgets extracted to separate files
- [ ] Code reusable and maintainable
- [ ] Clear separation of concerns

✅ **Documentation**
- [ ] Widget documentation complete
- [ ] Callback typedefs documented
- [ ] Edge cases documented
- [ ] Known limitations documented

### Final Deliverables

1. **Updated Files**:
   - `http_details_panel.dart` (refactored)
   - `request_action_buttons.dart` (new)
   - `request_body_tab.dart` (new)
   - `request_info_tab.dart` (new)
   - `response_body_tab.dart` (new)
   - `response_info_tab.dart` (new)

2. **Test Files**:
   - Unit tests for all 5 widgets
   - Integration test for panel
   - Performance benchmarks

3. **Documentation**:
   - Updated CHANGELOG.md
   - Widget API documentation
   - Migration guide (if needed)

### Success Metrics

- ✅ 0 critical bugs
- ✅ 0 regressions
- ✅ All tests green
- ✅ Performance targets met
- ✅ Code review approved

### Rollback Strategy

If final acceptance fails:
1. Document all failing criteria
2. Prioritize critical vs nice-to-have
3. Fix critical issues only
4. Re-run acceptance checklist
5. If still failing, revert to previous stable version

---

## Week 3 Summary

### Days 11-15 Overview
- **Day 11**: Unit tests for extracted widgets
- **Day 12**: Integration tests for tab navigation
- **Day 13**: E2E testing and performance validation
- **Day 14**: Bug fixes from testing
- **Day 15**: Final manual testing and acceptance

### Success Criteria
- All 37 previously identified issues verified fixed
- No new regressions introduced
- Performance benchmarks met
- Code quality standards maintained
- User acceptance achieved

### Post-Week 3 Actions
1. Merge to main branch
2. Tag release (e.g., v0.2.0)
3. Update documentation
4. Monitor for user feedback
5. Plan next iteration

---

## Notes

- Run `flutter test --coverage` to generate coverage report
- Use `flutter run --profile` to profile performance
- DevTools for memory leak detection
- Keep test data realistic (use actual API responses)
- Test on both web and desktop platforms
- Consider accessibility testing (screen reader, keyboard navigation)
