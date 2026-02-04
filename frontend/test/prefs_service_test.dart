import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:frontend/services/prefs.dart';

void main() {
  setUp(() async {
    // Clear mock before each test
    SharedPreferences.setMockInitialValues(<String, Object>{});
  });

  test('save/load saves only passed fields and applies defaults', () async {
    // Arrange
    final prefs = PrefsService();

    // Act
    await prefs.save(
      baseUrl: 'http://x',
      targetWs: 'ws://y',
      q: 'abc',
      httpMinDurationMs: 250,
      groupBy: 'domain',
      respDelayEnabled: true,
      respDelayValue: '1000-2000',
    );
    final loaded = await prefs.load();

    // Assert
    expect(loaded['baseUrl'], 'http://x');
    expect(loaded['targetWs'], 'ws://y');
    expect(loaded['q'], 'abc');
    expect(
      loaded['httpMethod'],
      'any',
      reason: 'not passed - should be default',
    );
    expect(
      loaded['httpStatus'],
      'any',
      reason: 'default if not explicitly saved',
    );
    expect(loaded['httpMinDuration'], '250');
    expect(loaded['groupBy'], 'domain');
    expect(loaded['respDelayEnabled'], 'true');
    expect(loaded['respDelayValue'], '1000-2000');
  });

  test(
    'saveThemeModeString/loadThemeModeString: default is system and saves value',
    () async {
      // Arrange
      final prefs = PrefsService();

      // Act / Assert
      expect(await prefs.loadThemeModeString(), 'system');
      await prefs.saveThemeModeString('dark');
      expect(await prefs.loadThemeModeString(), 'dark');
    },
  );

  test(
    'saveSince/loadSince: correct ISO-UTC format and garbage handling',
    () async {
      // Arrange
      final prefs = PrefsService();
      final ts = DateTime.utc(2024, 1, 2, 3, 4, 5);

      // Act
      await prefs.saveSince(ts);
      final loaded = await prefs.loadSince();

      // Assert
      expect(loaded, isNotNull);
      expect(loaded!.toUtc(), ts);

      // Garbage value - should return null
      final sp = await SharedPreferences.getInstance();
      await sp.setString('clear_since_ts', 'n/a');
      expect(await prefs.loadSince(), isNull);
    },
  );

  test('saveMonitorLog/loadMonitorLog: truncates to 500 lines', () async {
    // Arrange
    final prefs = PrefsService();
    final big = List<String>.generate(800, (i) => 'line_$i');

    // Act
    await prefs.saveMonitorLog(big);
    final got = await prefs.loadMonitorLog();

    // Assert
    expect(got.length, 500);
    expect(got.first, 'line_0');
    expect(got.last, 'line_499');
  });

  test(
    'saveIsRecording/loadIsRecording: default is true and saves false',
    () async {
      // Arrange
      final prefs = PrefsService();

      // Act / Assert
      expect(await prefs.loadIsRecording(), isTrue);
      await prefs.saveIsRecording(false);
      expect(await prefs.loadIsRecording(), isFalse);
    },
  );

  test('saveRecentWindow: minutes is clamped to [1, 1440]', () async {
    // Arrange
    final prefs = PrefsService();

    // Act
    await prefs.saveRecentWindow(enabled: true, minutes: 0);
    await prefs.saveRecentWindow(enabled: true, minutes: 20000);

    // Assert
    expect(await prefs.loadRecentWindowEnabled(), isTrue);
    // Take the last save as the true state
    expect(await prefs.loadRecentWindowMinutes(), 1440);
  });

  test(
    'saveVisualDensity/loadVisualDensity: default is standard and saves value',
    () async {
      // Arrange
      final prefs = PrefsService();

      // Act / Assert default
      expect(await prefs.loadVisualDensity(), 'standard');

      // Save compact
      await prefs.saveVisualDensity('compact');
      expect(await prefs.loadVisualDensity(), 'compact');

      // Save comfortable
      await prefs.saveVisualDensity('comfortable');
      expect(await prefs.loadVisualDensity(), 'comfortable');
    },
  );
}
