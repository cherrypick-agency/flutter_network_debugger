import '../../domain/entities/installer_result.dart';

/// Stub implementation for web platform
Future<InstallerResult> openInstallerImpl(String filePath) async {
  return InstallerResult.failure(
    'Installer opening not supported on web platform',
  );
}
