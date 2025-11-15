import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/http_inspector/presentation/widgets/request_action_buttons.dart';

void main() {
  group('RequestActionButtons', () {
    testWidgets('renders all action buttons', (tester) async {
      bool copyCurlCalled = false;
      bool repeatCalled = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'GET',
              headers: const {},
              body: '',
              onCopyCurl: (curl) => copyCurlCalled = true,
              onRepeat: () => repeatCalled = true,
            ),
          ),
        ),
      );

      // Verify buttons exist
      expect(find.text('Copy as cURL'), findsOneWidget);
      expect(find.text('Repeat'), findsOneWidget);
    });

    testWidgets('Repeat button not shown when URL is empty', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: '', // Empty URL
              method: 'GET',
              headers: const {},
              body: '',
              onCopyCurl: (curl) {},
              onRepeat: () {},
            ),
          ),
        ),
      );

      // Button should not be rendered when URL is empty
      expect(find.text('Repeat'), findsNothing);
    });

    testWidgets('generates correct compact cURL command', (tester) async {
      String? generatedCurl;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api/users',
              method: 'POST',
              headers: const {'Content-Type': 'application/json'},
              body: '{"name": "test"}',
              onCopyCurl: (curl) => generatedCurl = curl,
            ),
          ),
        ),
      );

      // Tap the cURL button to open menu
      await tester.tap(find.text('Copy as cURL'));
      await tester.pumpAndSettle();

      // Tap Compact option
      await tester.tap(find.text('Compact'));
      await tester.pump();

      expect(generatedCurl, isNotNull);
      expect(generatedCurl, contains('curl -X POST'));
      expect(generatedCurl, contains('https://example.com/api/users'));
      expect(generatedCurl, contains('Content-Type: application/json'));
      expect(generatedCurl, contains('{"name": "test"}'));
    });

    testWidgets('shows loading indicator when isLoading is true', (
      tester,
    ) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: RequestActionButtons(
              url: 'https://example.com/api',
              method: 'GET',
              headers: const {},
              body: '',
              onCopyCurl: (curl) {},
              onRepeat: () {},
              isLoading: true,
            ),
          ),
        ),
      );

      // Should show loading indicator instead of Repeat button
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Repeat'), findsNothing);
    });
  });
}
