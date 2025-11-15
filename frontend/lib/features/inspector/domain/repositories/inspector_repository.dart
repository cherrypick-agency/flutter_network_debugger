import '../entities/session.dart';
import '../entities/frame.dart';
import '../entities/event.dart';

abstract class InspectorRepository {
  Future<List<Session>> listSessions({
    String? q,
    String? target,
    List<String>? tags,
  });
  Future<List<Frame>> listFrames(
    String sessionId, {
    String? from,
    int limit = 100,
  });
  Future<List<EventEntity>> listEvents(
    String sessionId, {
    String? from,
    int limit = 100,
  });
  Future<List<Map<String, dynamic>>> aggregateSessions({
    String groupBy = 'domain',
  });

  /// Export sessions to HAR format
  /// Returns the URL to download the HAR file
  Future<String> exportHAR({
    List<String>? sessionIds,
    bool includeBodies = true,
    bool includeSensitive = false,
    bool minify = false,
  });

  /// Import HAR file
  /// Returns statistics about the import operation
  Future<Map<String, dynamic>> importHAR(
    String harJson, {
    required String mode,
  });
}
