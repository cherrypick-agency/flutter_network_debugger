import 'dart:io';
import 'package:path/path.dart' as path;
import '../../domain/entities/installer_result.dart';

/// Desktop implementation for opening installers
Future<InstallerResult> openInstallerImpl(String filePath) async {
  try {
    if (!File(filePath).existsSync()) {
      return InstallerResult.failure('Installer file not found: $filePath');
    }

    final fileName = path.basename(filePath);
    final extension = path.extension(fileName).toLowerCase();

    // macOS
    if (Platform.isMacOS) {
      if (extension == '.dmg') {
        final result = await Process.run('open', [filePath]);
        if (result.exitCode == 0) {
          return InstallerResult.success(
            'Opening DMG file.\n\n'
            'Installation steps:\n'
            '1. Drag the app to the Applications folder\n'
            '2. If you see "unidentified developer" warning:\n'
            '   • Right-click the app in Applications\n'
            '   • Select "Open" from the menu\n'
            '   • Click "Open" in the dialog\n\n'
            'This bypasses macOS Gatekeeper for this app.',
          );
        } else {
          return InstallerResult.failure(
            'Failed to open DMG: ${result.stderr}',
          );
        }
      } else {
        return InstallerResult.failure(
          'Unsupported file type for macOS: $extension',
        );
      }
    }

    // Windows
    if (Platform.isWindows) {
      if (extension == '.msi') {
        // Start MSI installer in detached mode (we can't verify if it actually started)
        await Process.start('msiexec', [
          '/i',
          filePath,
        ], mode: ProcessStartMode.detached);
        return InstallerResult.success(
          'Starting MSI installer.\n\n'
          'Installation steps:\n'
          '1. Windows may ask for Administrator permission (UAC)\n'
          '   • Click "Yes" to allow installation\n'
          '2. Follow the installation wizard\n'
          '3. The app will be installed to Program Files\n\n'
          'Note: If the installer doesn\'t start, you may need to:\n'
          '• Right-click the downloaded file\n'
          '• Select "Run as administrator"',
        );
      } else if (extension == '.zip') {
        // Open containing folder
        final dir = path.dirname(filePath);
        await Process.run('explorer', [dir]);
        return InstallerResult.success(
          'Opening folder with installer. Please extract and run the installer.',
        );
      } else {
        return InstallerResult.failure(
          'Unsupported file type for Windows: $extension',
        );
      }
    }

    // Linux
    if (Platform.isLinux) {
      if (extension == '.deb') {
        // Try to open with system package manager
        final result = await Process.run('xdg-open', [filePath]);
        if (result.exitCode == 0) {
          return InstallerResult.success(
            'Opening with system package manager.\n\n'
            'Installation steps:\n'
            '1. Enter your password when prompted\n'
            '2. Review the package information\n'
            '3. Click "Install" to proceed\n\n'
            'Alternative installation:\n'
            'Open terminal and run:\n'
            'sudo dpkg -i $fileName\n'
            'sudo apt-get install -f  # Fix dependencies if needed',
          );
        } else {
          // Fallback: provide instructions
          return InstallerResult.success(
            'To install, run in terminal:\n\n'
            'sudo dpkg -i $fileName\n'
            'sudo apt-get install -f  # Fix dependencies if needed\n\n'
            'Or double-click the file to open with Software Center.',
          );
        }
      } else if (extension == '.appimage') {
        // Check if FUSE is available (required for AppImage)
        final fuseCheck = await Process.run('which', ['fusermount']);
        final hasFuse = fuseCheck.exitCode == 0;

        if (!hasFuse) {
          return InstallerResult.failure(
            'AppImage requires FUSE to run.\n\n'
            'Install FUSE:\n'
            'Ubuntu/Debian: sudo apt-get install fuse libfuse2\n'
            'Fedora: sudo dnf install fuse fuse-libs\n'
            'Arch: sudo pacman -S fuse2\n\n'
            'After installing FUSE, restart your system and try again.',
          );
        }

        // Make executable and open
        final makeExecutable = await Process.run('chmod', ['+x', filePath]);
        if (makeExecutable.exitCode == 0) {
          // Start AppImage in detached mode (we can't verify if it actually started)
          await Process.start(filePath, [], mode: ProcessStartMode.detached);
          return InstallerResult.success(
            'Starting AppImage...\n\n'
            'AppImage is a portable format:\n'
            '• No installation needed\n'
            '• Can be moved to any location\n'
            '• Recommended: Move to ~/Applications or /opt\n\n'
            'To make it accessible from menu:\n'
            'Right-click → Properties → Permissions → Allow executing as program',
          );
        } else {
          return InstallerResult.success(
            'To run AppImage:\n\n'
            'chmod +x $fileName\n'
            './$fileName\n\n'
            'Or right-click → Properties → Permissions → Allow executing as program',
          );
        }
      } else if (extension == '.gz' && fileName.endsWith('.tar.gz')) {
        // Open containing folder
        final dir = path.dirname(filePath);
        await Process.run('xdg-open', [dir]);
        return InstallerResult.success(
          'Opening folder with archive.\n\n'
          'To extract and install:\n'
          'tar -xzf $fileName\n'
          'cd [extracted-folder]\n'
          'cat README.md  # Read installation instructions\n\n'
          'Common installation steps:\n'
          './configure\n'
          'make\n'
          'sudo make install',
        );
      } else {
        return InstallerResult.failure(
          'Unsupported file type for Linux: $extension',
        );
      }
    }

    return InstallerResult.failure(
      'Unsupported platform: ${Platform.operatingSystem}',
    );
  } catch (e) {
    return InstallerResult.failure('Error opening installer: $e');
  }
}
