import 'package:freezed_annotation/freezed_annotation.dart';

part 'file_statistics.freezed.dart';
part 'file_statistics.g.dart';

@freezed
sealed class FileStatistics with _$FileStatistics {
  const FileStatistics._();

  const factory FileStatistics({
    required String fileId,
    required int lines,
    required int characters,
    required int words,
    required int bytes,
    required DateTime calculatedAt,
  }) = _FileStatistics;

  factory FileStatistics.fromJson(Map<String, dynamic> json) =>
      _$FileStatisticsFromJson(json);

  factory FileStatistics.calculate(String fileId, String content) {
    final lines = content.split('\n').length;
    final characters = content.length;
    final words = content.trim().isEmpty
        ? 0
        : content.trim().split(RegExp(r'\s+')).length;
    final bytes = content.length;

    return FileStatistics(
      fileId: fileId,
      lines: lines,
      characters: characters,
      words: words,
      bytes: bytes,
      calculatedAt: DateTime.now(),
    );
  }

  String get displayText => '$lines lines, $characters chars, $words words';

  bool get isEmpty => lines == 0 && characters == 0;
}
