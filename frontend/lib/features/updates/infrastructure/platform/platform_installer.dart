// Domain entity import
import '../../domain/entities/installer_result.dart';

// Conditional imports (must be before declarations)
import 'platform_installer_stub.dart'
    if (dart.library.io) 'platform_installer_io.dart';

export '../../domain/entities/installer_result.dart';

/// Открывает установщик на текущей платформе
Future<InstallerResult> openInstaller(String filePath) async {
  return openInstallerImpl(filePath);
}
