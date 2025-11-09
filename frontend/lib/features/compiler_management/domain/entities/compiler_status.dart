import 'package:freezed_annotation/freezed_annotation.dart';

part 'compiler_status.freezed.dart';
part 'compiler_status.g.dart';

@freezed
sealed class CompilerStatus with _$CompilerStatus {
  const factory CompilerStatus.installed() = _Installed;
  const factory CompilerStatus.installing() = _Installing;
  const factory CompilerStatus.notInstalled() = _NotInstalled;

  factory CompilerStatus.fromJson(Map<String, dynamic> json) =>
      _$CompilerStatusFromJson(json);

  factory CompilerStatus.fromString(String status) {
    switch (status.toLowerCase()) {
      case 'installed':
        return const CompilerStatus.installed();
      case 'installing':
        return const CompilerStatus.installing();
      case 'not_installed':
      case 'notinstalled':
        return const CompilerStatus.notInstalled();
      default:
        return const CompilerStatus.notInstalled();
    }
  }
}
