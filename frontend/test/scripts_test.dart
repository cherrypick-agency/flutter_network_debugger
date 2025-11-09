import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/scripts/domain/entities/script.dart';
import 'package:frontend/features/scripts/domain/entities/match_rules.dart';

void main() {
  test('Can create MatchRules instance', () {
    final rules = MatchRules(
      methods: ['GET', 'POST'],
      pathPattern: '/api/*',
      patternType: PatternType.wildcard,
    );

    expect(rules.methods, ['GET', 'POST']);
    expect(rules.pathPattern, '/api/*');
  });

  test('Can create Script instance', () {
    final script = Script(
      id: '123',
      name: 'Test Script',
      runtime: ScriptRuntime.dart,
      code: 'test code',
      language: 'dart',
      triggerType: TriggerType.request,
    );

    expect(script.name, 'Test Script');
    expect(script.runtime, ScriptRuntime.dart);
  });
}
