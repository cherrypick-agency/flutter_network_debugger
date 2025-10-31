import 'body_view_mode.dart';
import 'content_detector.dart';
import 'jwt_detector.dart';
import 'base64_detector.dart';

/// Analyzes body content and determines available view modes
/// (SOLID: Single Responsibility - only analyzes content)
class BodyContentAnalyzer {
  final ContentDetectorRegistry _registry;
  final JwtContentDetector _jwtDetector;
  final Base64ContentDetector _base64Detector;

  BodyContentAnalyzer()
    : _registry = ContentDetectorRegistry(),
      _jwtDetector = JwtContentDetector(),
      _base64Detector = Base64ContentDetector() {
    // Register detectors in priority order
    _registry.registerAll([
      _jwtDetector,
      JsonContentDetector(),
      HtmlContentDetector(),
      XmlContentDetector(),
      _base64Detector,
    ]);
  }

  /// Analyze content and return available view modes
  ContentAnalysisResult analyze(String content, {String? contentType}) {
    if (content.isEmpty) {
      return ContentAnalysisResult(
        content: content,
        availableModes: {BodyViewMode.raw},
        defaultMode: BodyViewMode.raw,
      );
    }

    final availableModes = <BodyViewMode>{};
    BodyViewMode? primaryMode;
    DetectionResult? primaryDetection;

    // Check JWT first (highest priority)
    if (_jwtDetector.detect(content, contentType: contentType)) {
      availableModes.add(BodyViewMode.jwtDecoded);
      primaryMode ??= BodyViewMode.jwtDecoded;
      primaryDetection = DetectionResult(
        detector: _jwtDetector,
        content: content,
        contentType: contentType,
      );
    }

    // Run standard detection
    final detection = _registry.detect(content, contentType: contentType);

    if (detection.isDetected) {
      final detector = detection.detector!;

      if (detector is JsonContentDetector) {
        availableModes.add(BodyViewMode.pretty);
        availableModes.add(BodyViewMode.jsonTree);
        primaryMode ??= BodyViewMode.pretty;
        primaryDetection ??= detection;
      } else if (detector is HtmlContentDetector) {
        availableModes.add(BodyViewMode.htmlPreview);
        availableModes.add(BodyViewMode.pretty);
        primaryMode ??= BodyViewMode.htmlPreview;
        primaryDetection ??= detection;
      } else if (detector is XmlContentDetector) {
        availableModes.add(BodyViewMode.pretty);
        primaryMode ??= BodyViewMode.pretty;
        primaryDetection ??= detection;
      } else if (detector is Base64ContentDetector) {
        availableModes.add(BodyViewMode.base64Decoded);
        availableModes.add(BodyViewMode.raw);
        primaryMode ??= BodyViewMode.base64Decoded;
        primaryDetection ??= detection;
      }
    }

    // Always offer Raw mode as fallback
    availableModes.add(BodyViewMode.raw);

    // If Pretty mode is available and no primary mode, use it
    if (primaryMode == null && availableModes.contains(BodyViewMode.pretty)) {
      primaryMode = BodyViewMode.pretty;
    }

    // Ultimate fallback
    primaryMode ??= BodyViewMode.raw;

    return ContentAnalysisResult(
      content: content,
      contentType: contentType,
      availableModes: availableModes,
      defaultMode: primaryMode,
      detection: primaryDetection,
      jwtData:
          primaryMode == BodyViewMode.jwtDecoded
              ? _jwtDetector.parse(content)
              : null,
      base64Data:
          primaryMode == BodyViewMode.base64Decoded
              ? _base64Detector.decode(content)
              : null,
    );
  }
}

/// Result of content analysis
class ContentAnalysisResult {
  final String content;
  final String? contentType;
  final Set<BodyViewMode> availableModes;
  final BodyViewMode defaultMode;
  final DetectionResult? detection;
  final JwtData? jwtData;
  final DecodeResult? base64Data;

  const ContentAnalysisResult({
    required this.content,
    this.contentType,
    required this.availableModes,
    required this.defaultMode,
    this.detection,
    this.jwtData,
    this.base64Data,
  });

  bool get isJson => availableModes.contains(BodyViewMode.jsonTree);
  bool get isHtml => availableModes.contains(BodyViewMode.htmlPreview);
  bool get isJwt => availableModes.contains(BodyViewMode.jwtDecoded);
  bool get isBase64 => availableModes.contains(BodyViewMode.base64Decoded);
}
