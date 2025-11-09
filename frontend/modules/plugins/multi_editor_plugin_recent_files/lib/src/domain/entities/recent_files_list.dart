import 'package:freezed_annotation/freezed_annotation.dart';
import '../value_objects/recent_file_entry.dart';

part 'recent_files_list.freezed.dart';
part 'recent_files_list.g.dart';

@freezed
sealed class RecentFilesList with _$RecentFilesList {
  const RecentFilesList._();

  const factory RecentFilesList({
    @Default([]) List<RecentFileEntry> entries,
    @Default(10) int maxEntries,
  }) = _RecentFilesList;

  factory RecentFilesList.fromJson(Map<String, dynamic> json) =>
      _$RecentFilesListFromJson(json);

  factory RecentFilesList.create({int maxEntries = 10}) {
    if (maxEntries < 1 || maxEntries > 50) {
      throw ArgumentError('maxEntries must be between 1 and 50');
    }
    return RecentFilesList(maxEntries: maxEntries);
  }

  RecentFilesList addFile(RecentFileEntry entry) {
    final existingIndex = entries.indexWhere((e) => e.fileId == entry.fileId);

    List<RecentFileEntry> newEntries;
    if (existingIndex >= 0) {
      newEntries = List<RecentFileEntry>.from(entries);
      newEntries.removeAt(existingIndex);
      newEntries.insert(0, entry.touch());
    } else {
      newEntries = [entry, ...entries];
      if (newEntries.length > maxEntries) {
        newEntries = newEntries.take(maxEntries).toList();
      }
    }

    return copyWith(entries: newEntries);
  }

  RecentFilesList removeFile(String fileId) {
    final newEntries = entries.where((e) => e.fileId != fileId).toList();
    return copyWith(entries: newEntries);
  }

  RecentFilesList clear() => copyWith(entries: []);

  List<RecentFileEntry> get sortedByRecent =>
      List<RecentFileEntry>.from(entries);

  bool contains(String fileId) => entries.any((e) => e.fileId == fileId);

  int get count => entries.length;

  bool get isEmpty => entries.isEmpty;

  bool get isNotEmpty => entries.isNotEmpty;

  bool get isFull => entries.length >= maxEntries;
}
