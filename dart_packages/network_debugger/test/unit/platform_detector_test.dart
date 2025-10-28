import 'package:test/test.dart';
import 'package:network_debugger/src/platform_detector.dart';

void main() {
  group('PlatformDetector', () {
    test('getPlatformIdentifier returns valid format', () {
      final identifier = PlatformDetector.getPlatformIdentifier();
      expect(
        identifier,
        matches(RegExp(r'^(darwin|windows|linux)_(amd64|arm64|386)$')),
      );
    });

    test('getBinaryName returns correct format', () {
      final binaryName = PlatformDetector.getBinaryName();
      expect(binaryName, contains('network-debugger-web_'));
      expect(binaryName, contains(PlatformDetector.getPlatformIdentifier()));
    });

    test('getArchiveName returns correct extension', () {
      final archiveName = PlatformDetector.getArchiveName();
      final hasValidExtension =
          archiveName.endsWith('.tar.gz') || archiveName.endsWith('.zip');
      expect(hasValidExtension, isTrue);
    });

    test('isSupported returns true for current platform', () {
      expect(PlatformDetector.isSupported(), isTrue);
    });

    test('getPlatformDescription returns non-empty string', () {
      final description = PlatformDetector.getPlatformDescription();
      expect(description, isNotEmpty);
      expect(description, contains(PlatformDetector.getPlatformIdentifier()));
    });

    test('getFileExtension matches platform', () {
      final ext = PlatformDetector.getFileExtension();
      final identifier = PlatformDetector.getPlatformIdentifier();

      if (identifier.startsWith('windows')) {
        expect(ext, equals('.exe'));
      } else {
        expect(ext, equals(''));
      }
    });

    test('getArchiveExtension matches platform', () {
      final ext = PlatformDetector.getArchiveExtension();
      final identifier = PlatformDetector.getPlatformIdentifier();

      if (identifier.startsWith('windows')) {
        expect(ext, equals('.zip'));
      } else {
        expect(ext, equals('.tar.gz'));
      }
    });
  });
}
