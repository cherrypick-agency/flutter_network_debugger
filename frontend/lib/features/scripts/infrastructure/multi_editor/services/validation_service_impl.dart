import 'package:multi_editor_core/multi_editor_core.dart';

/// Реализация [ValidationService] для проверки имён/путей/контента
class ValidationServiceImpl implements ValidationService {
  // Разрешённые паттерны для имён
  static final RegExp _fileNameRegex = RegExp(r'^[a-zA-Z0-9_.-]+$');
  static final RegExp _folderNameRegex = RegExp(r'^[a-zA-Z0-9_.-]+$');
  static final RegExp _projectNameRegex = RegExp(r'^[a-zA-Z0-9_\\s-]+$');

  // Список допустимых языков и расширений
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
          reason: 'Имя файла не может быть пустым',
        ),
      );
    }

    if (name.length > 255) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Слишком длинное имя файла (макс. 255 символов)',
        ),
      );
    }

    if (!_fileNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Недопустимые символы (разрешены буквы/цифры, точка, дефис, подчёркивание)',
        ),
      );
    }

    if (name.startsWith('.') || name.startsWith('-')) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Имя файла не должно начинаться с точки или дефиса',
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
          reason: 'Путь не может быть пустым',
        ),
      );
    }

    if (path.length > 4096) {
      return Left(
        DomainFailure.validationError(
          field: 'path',
          reason: 'Слишком длинный путь (макс. 4096 символов)',
        ),
      );
    }

    if (path.contains(RegExp(r'[<>"|?*]'))) {
      return Left(
        DomainFailure.validationError(
          field: 'path',
          reason: 'Путь содержит недопустимые символы',
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
          reason: 'Имя папки не может быть пустым',
        ),
      );
    }

    if (name.length > 255) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Слишком длинное имя папки (макс. 255 символов)',
        ),
      );
    }

    if (!_folderNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Недопустимые символы (разрешены буквы/цифры, точка, дефис, подчёркивание)',
        ),
      );
    }

    if (name.startsWith('.') || name.startsWith('-')) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Имя папки не должно начинаться с точки или дефиса',
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
          reason: 'Имя проекта не может быть пустым',
        ),
      );
    }

    if (name.length < 3) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Имя проекта должно быть не короче 3 символов',
        ),
      );
    }

    if (name.length > 100) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason: 'Слишком длинное имя проекта (макс. 100 символов)',
        ),
      );
    }

    if (!_projectNameRegex.hasMatch(name)) {
      return Left(
        DomainFailure.validationError(
          field: 'name',
          reason:
              'Недопустимые символы (разрешены буквы/цифры, пробел, дефис, подчёркивание)',
        ),
      );
    }

    return const Right(null);
  }

  @override
  Either<DomainFailure, void> validateFileContent(String content) {
    // Простая защита от чрезмерного объёма
    if (content.length > 10 * 1024 * 1024) {
      return Left(
        DomainFailure.validationError(
          field: 'content',
          reason: 'Содержимое файла слишком большое (макс. 10MB)',
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
