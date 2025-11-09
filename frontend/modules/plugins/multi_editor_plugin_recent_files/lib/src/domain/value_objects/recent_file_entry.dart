import 'package:freezed_annotation/freezed_annotation.dart';

part 'recent_file_entry.freezed.dart';
part 'recent_file_entry.g.dart';

@freezed
sealed class RecentFileEntry with _$RecentFileEntry {
  const RecentFileEntry._();

  const factory RecentFileEntry({
    required String fileId,
    required String fileName,
    required String filePath,
    required DateTime lastOpened,
  }) = _RecentFileEntry;

  factory RecentFileEntry.fromJson(Map<String, dynamic> json) =>
      _$RecentFileEntryFromJson(json);

  factory RecentFileEntry.create({
    required String fileId,
    required String fileName,
    required String filePath,
  }) {
    return RecentFileEntry(
      fileId: fileId,
      fileName: fileName,
      filePath: filePath,
      lastOpened: DateTime.now(),
    );
  }

  RecentFileEntry touch() => copyWith(lastOpened: DateTime.now());

  String get displayName => fileName;

  Duration timeSince(DateTime now) => now.difference(lastOpened);

  /// Format relative time in Russian (e.g., "10 секунд назад", "3 минуты назад")
  String get formattedTime {
    final duration = timeSince(DateTime.now());

    if (duration.inSeconds < 60) {
      final seconds = duration.inSeconds;
      return '$seconds ${_secondsWord(seconds)} назад';
    } else if (duration.inMinutes < 60) {
      final minutes = duration.inMinutes;
      return '$minutes ${_minutesWord(minutes)} назад';
    } else if (duration.inHours < 24) {
      final hours = duration.inHours;
      return '$hours ${_hoursWord(hours)} назад';
    } else {
      final days = duration.inDays;
      return '$days ${_daysWord(days)} назад';
    }
  }

  String _secondsWord(int n) {
    if (n % 10 == 1 && n % 100 != 11) return 'секунду';
    if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) {
      return 'секунды';
    }
    return 'секунд';
  }

  String _minutesWord(int n) {
    if (n % 10 == 1 && n % 100 != 11) return 'минуту';
    if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) {
      return 'минуты';
    }
    return 'минут';
  }

  String _hoursWord(int n) {
    if (n % 10 == 1 && n % 100 != 11) return 'час';
    if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) {
      return 'часа';
    }
    return 'часов';
  }

  String _daysWord(int n) {
    if (n % 10 == 1 && n % 100 != 11) return 'день';
    if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) {
      return 'дня';
    }
    return 'дней';
  }
}
