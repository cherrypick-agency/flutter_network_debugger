import 'package:freezed_annotation/freezed_annotation.dart';
import 'compiler_status.dart';

part 'compiler_info.freezed.dart';
part 'compiler_info.g.dart';

@freezed
sealed class CompilerInfo with _$CompilerInfo {
  const CompilerInfo._();

  const factory CompilerInfo({
    required String language,
    required String version,
    required CompilerStatus status,
    required String installedPath,
    required int size,
    required int downloadSize,
    @Default('') String error,
  }) = _CompilerInfo;

  factory CompilerInfo.fromJson(Map<String, dynamic> json) =>
      _$CompilerInfoFromJson(json);

  bool get isInstalled =>
      status.maybeWhen(installed: () => true, orElse: () => false);
  bool get isInstalling =>
      status.maybeWhen(installing: () => true, orElse: () => false);
  bool get isNotInstalled =>
      status.maybeWhen(notInstalled: () => true, orElse: () => false);

  String get displaySize {
    if (isInstalled && size > 0) {
      return _formatBytes(size);
    } else if (downloadSize > 0) {
      return _formatBytes(downloadSize);
    }
    return 'Unknown';
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }
}
