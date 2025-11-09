import '../repositories/font_repository.dart';

/// Use case for removing a custom font
/// Follows Single Responsibility Principle
class RemoveCustomFontUseCase {
  final FontRepository repository;

  RemoveCustomFontUseCase({required this.repository});

  /// Execute the use case
  /// Removes custom font data and metadata from storage
  /// Note: Flutter doesn't support unloading fonts at runtime,
  /// so the font will remain registered until app restart
  Future<void> execute() async {
    await repository.removeCustomFont();
  }
}
