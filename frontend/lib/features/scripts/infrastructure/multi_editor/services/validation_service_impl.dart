import 'package:multi_editor_core/multi_editor_core.dart';

/// Implementation of [ValidationService] for validating names/paths/content
class ValidationServiceImpl implements ValidationService {
  // Allowed patterns for names
  static final RegExp _fileNameRegex = RegExp(r'^[a-zA-Z0-9_.-]+$');
  static final RegExp _folderNameRegex = RegExp(r'^[a-zA-Z0-9_.-]+$');
  static final RegExp _projectNameRegex = RegExp(r'^[a-zA-Z0-9_\\s-]+$');

  // List of valid languages and extensions
  static const List<String> _validLanguages = [
    'dart',
    'javascript',
    'typescript',
    'python',
    'java',
    'cpp',
    'c',
    'go',
    'rust',
    'ruby',
    'php',
    'swift',
    'kotlin',
    'csharp',
    'html',
    'css',
    'json',
    'yaml',
    'markdown',
    'xml',
    'sql',
    'shell',
    'plaintext',
  ];

  static const List<String> _validExtensions = [
    '.dart',
    '.js',
    '.ts',
    '.tsx',
    '.jsx',
    '.py',
    '.java',
    '.cpp',
    '.c',
    '.h',
    '.go',
    '.rs',
    '.rb',
    '.php',
    '.swift',
    '.kt',
    '.cs',
    '.html',
    '.css',
    '.scss',
    '.sass',
    '.json',
    '.yaml',
    '.yml',
    '.md',
    '.xml',
    '.sql',
    '.sh',
    '.txt',
  ];

  @override
  Either<DomainFailure, void> validateFileName(String name) {
    if (name.isEmpty) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'File name cannot be empty',
        ),
      );
    }

    if (name.length > 255) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'File name too long (max 255 characters)',
        ),
      );
    }

    if (!_fileNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Invalid characters (allowed: letters, digits, dot, dash, underscore)',
        ),
      );
    }

    if (name.startsWith('.') || name.startsWith('-')) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'File name must not start with dot or dash',
        ),
      );
    }

    return const Right(null);
  }

  @override
  Either<DomainFailure, void> validateFilePath(String path) {
    if (path.isEmpty) {
      return Left(
        DomainFailure.validationError(
          field: 'path',
          reason: 'Path cannot be empty',
        ),
      );
    }

    if (path.length > 4096) {
      return Left(
        DomainFailure.validationError(
          field: 'path',
          reason: 'Path too long (max 4096 characters)',
        ),
      );
    }

    if (path.contains(RegExp(r'[<>"|?*]'))) {
      return Left(
        DomainFailure.validationError(
          field: 'path',
          reason: 'Path contains invalid characters',
        ),
      );
    }

    return const Right(null);
  }

  @override
  Either<DomainFailure, void> validateFolderName(String name) {
    if (name.isEmpty) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Folder name cannot be empty',
        ),
      );
    }

    if (name.length > 255) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Folder name too long (max 255 characters)',
        ),
      );
    }

    if (!_folderNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Invalid characters (allowed: letters, digits, dot, dash, underscore)',
        ),
      );
    }

    if (name.startsWith('.') || name.startsWith('-')) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Folder name must not start with dot or dash',
        ),
      );
    }

    return const Right(null);
  }

  @override
  Either<DomainFailure, void> validateProjectName(String name) {
    if (name.isEmpty) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Project name cannot be empty',
        ),
      );
    }

    if (name.length < 3) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Project name must be at least 3 characters',
        ),
      );
    }

    if (name.length > 100) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Project name too long (max 100 characters)',
        ),
      );
    }

    if (!_projectNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Invalid characters (allowed: letters, digits, space, dash, underscore)',
        ),
      );
    }

    return const Right(null);
  }

  @override
  Either<DomainFailure, void> validateFileContent(String content) {
    // Simple protection against excessive size
    if (content.length > 10 * 1024 * 1024) {
      return Left(
        DomainFailure.validationError(
          field: 'content',
          reason: 'File content too large (max 10MB)',
        ),
      );
    }

    return const Right(null);
  }

  @override
  bool isValidLanguage(String language) {
    return _validLanguages.contains(language.toLowerCase());
  }

  @override
  bool hasValidExtension(String fileName) {
    final extension = fileName.contains('.')
        ? '.${fileName.split('.').last}'
        : '';
    return _validExtensions.contains(extension.toLowerCase());
  }
}
