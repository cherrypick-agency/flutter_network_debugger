import '../../domain/entities/compiler_info.dart';
import '../../domain/entities/compiler_status.dart';

class CompilerInfoModel {
  final String language;
  final String version;
  final String status;
  final String installedPath;
  final int size;
  final int downloadSize;
  final String error;

  CompilerInfoModel({
    required this.language,
    required this.version,
    required this.status,
    required this.installedPath,
    required this.size,
    required this.downloadSize,
    required this.error,
  });

  factory CompilerInfoModel.fromJson(Map<String, dynamic> json) {
    return CompilerInfoModel(
      language: json['language'] as String? ?? '',
      version: json['version'] as String? ?? '',
      status: json['status'] as String? ?? 'not_installed',
      installedPath: json['installedPath'] as String? ?? '',
      size: json['size'] as int? ?? 0,
      downloadSize: json['downloadSize'] as int? ?? 0,
      error: json['error'] as String? ?? '',
    );
  }

  CompilerInfo toEntity() {
    return CompilerInfo(
      language: language,
      version: version,
      status: CompilerStatus.fromString(status),
      installedPath: installedPath,
      size: size,
      downloadSize: downloadSize,
      error: error,
    );
  }
}
