import 'dart:convert';

/// Base interface for content type detection (SOLID: Interface Segregation)
abstract class ContentDetector {
  /// Check if content matches this detector's format
  bool detect(String content, {String? contentType});

  /// Get display name for this content type
  String get displayName;

  /// Priority for detection (higher = checked first)
  int get priority;
}

/// Detection result with metadata
class DetectionResult {
  final ContentDetector? detector;
  final String content;
  final String? contentType;

  const DetectionResult({
    this.detector,
    required this.content,
    this.contentType,
  });

  bool get isDetected => detector != null;
}

/// Content detector manager (SOLID: Single Responsibility)
class ContentDetectorRegistry {
  final List<ContentDetector> _detectors = [];

  void register(ContentDetector detector) {
    _detectors.add(detector);
    // Sort by priority (higher first)
    _detectors.sort((a, b) => b.priority.compareTo(a.priority));
  }

  void registerAll(List<ContentDetector> detectors) {
    for (final d in detectors) {
      register(d);
    }
  }

  DetectionResult detect(String content, {String? contentType}) {
    for (final detector in _detectors) {
      if (detector.detect(content, contentType: contentType)) {
        return DetectionResult(
          detector: detector,
          content: content,
          contentType: contentType,
        );
      }
    }
    return DetectionResult(content: content, contentType: contentType);
  }

  List<ContentDetector> get all => List.unmodifiable(_detectors);
}

/// JSON content detector
class JsonContentDetector implements ContentDetector {
  @override
  bool detect(String content, {String? contentType}) {
    final ct = (contentType ?? '').toLowerCase();
    if (ct.contains('json') || ct.contains('+json')) return true;

    try {
      jsonDecode(content);
      return true;
    } catch (_) {
      return false;
    }
  }

  @override
  String get displayName => 'JSON';

  @override
  int get priority => 10; // Medium priority
}

/// HTML content detector
class HtmlContentDetector implements ContentDetector {
  @override
  bool detect(String content, {String? contentType}) {
    final ct = (contentType ?? '').toLowerCase();
    if (ct.contains('html')) return true;

    final trimmed = content.trimLeft();
    if (trimmed.isEmpty) return false;

    final lower = trimmed.toLowerCase();
    return lower.startsWith('<!doctype html') ||
        lower.startsWith('<html') ||
        lower.contains('<head>') ||
        lower.contains('<body');
  }

  @override
  String get displayName => 'HTML';

  @override
  int get priority => 8;
}

/// XML content detector
class XmlContentDetector implements ContentDetector {
  @override
  bool detect(String content, {String? contentType}) {
    final ct = (contentType ?? '').toLowerCase();
    if (ct.contains('xml')) return true;

    return content.trimLeft().startsWith('<?xml');
  }

  @override
  String get displayName => 'XML';

  @override
  int get priority => 9;
}
