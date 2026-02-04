/// Installer opening result
class InstallerResult {
  final bool success;
  final String? errorMessage;
  final String? instructions; // May contain instructions for Linux

  const InstallerResult({
    required this.success,
    this.errorMessage,
    this.instructions,
  });

  factory InstallerResult.success([String? instructions]) {
    return InstallerResult(success: true, instructions: instructions);
  }

  factory InstallerResult.failure(String message) {
    return InstallerResult(success: false, errorMessage: message);
  }
}
