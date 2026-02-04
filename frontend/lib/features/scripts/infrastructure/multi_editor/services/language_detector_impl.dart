import 'package:multi_editor_core/multi_editor_core.dart';

/// Implementation of [LanguageDetector] with extension to Monaco language mapping
class LanguageDetectorImpl implements LanguageDetector {
  /// Map of extensions (without dot) -> Monaco language
  static const Map<String, String> _extensionMap = {
    // dart
    'dart': 'dart',

    // js/ts
    'js': 'javascript',
    'jsx': 'javascript',
    'ts': 'typescript',
    'tsx': 'typescript',
    'mjs': 'javascript',
    'cjs': 'javascript',

    // web
    'html': 'html',
    'htm': 'html',
    'css': 'css',
    'scss': 'scss',
    'sass': 'sass',
    'less': 'less',

    // data
    'json': 'json',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'toml': 'toml',

    // markup/docs
    'md': 'markdown',
    'markdown': 'markdown',
    'rst': 'rst',

    // systems
    'rs': 'rust',
    'go': 'go',
    'c': 'c',
    'cpp': 'cpp',
    'cc': 'cpp',
    'cxx': 'cpp',
    'h': 'c',
    'hpp': 'cpp',
    'hxx': 'cpp',

    // other
    'py': 'python',
    'java': 'java',
    'kt': 'kotlin',
    'swift': 'swift',
    'rb': 'ruby',
    'php': 'php',
    'sh': 'shell',
    'bash': 'shell',
    'zsh': 'shell',
    'sql': 'sql',
    'lua': 'lua',
    'r': 'r',

    // config
    'ini': 'ini',
    'conf': 'ini',
    'cfg': 'ini',
    'env': 'shell',

    // build/package
    'gradle': 'gradle',
    'dockerfile': 'dockerfile',
    'makefile': 'makefile',
  };

  /// Reverse mapping language -> extension (with dot)
  static const Map<String, String> _languageToExtension = {
    'dart': '.dart',
    'javascript': '.js',
    'typescript': '.ts',
    'python': '.py',
    'java': '.java',
    'cpp': '.cpp',
    'c': '.c',
    'go': '.go',
    'rust': '.rs',
    'ruby': '.rb',
    'php': '.php',
    'swift': '.swift',
    'kotlin': '.kt',
    'csharp': '.cs',
    'html': '.html',
    'css': '.css',
    'json': '.json',
    'yaml': '.yaml',
    'markdown': '.md',
    'xml': '.xml',
    'sql': '.sql',
    'shell': '.sh',
    'plaintext': '.txt',
    'ini': '.ini',
    'gradle': '.gradle',
    'dockerfile': 'Dockerfile',
    'makefile': 'Makefile',
  };

  @override
  String detectFromFileName(String fileName) {
    final lastDotIndex = fileName.lastIndexOf('.');
    if (lastDotIndex == -1 || lastDotIndex == fileName.length - 1) {
      return _detectSpecialNames(fileName);
    }
    final ext = fileName.substring(lastDotIndex + 1).toLowerCase();
    return _extensionMap[ext] ?? 'plaintext';
  }

  /// Handle files without extension and common names
  String _detectSpecialNames(String fileName) {
    final lowerName = fileName.toLowerCase();

    if (lowerName == 'dockerfile') return 'dockerfile';
    if (lowerName == 'makefile') return 'makefile';
    if (lowerName == 'cmakelists.txt') return 'cmake';

    return 'plaintext';
  }

  @override
  String detectFromExtension(String extension) {
    final normalized = extension.startsWith('.')
        ? extension.substring(1).toLowerCase()
        : extension.toLowerCase();
    return _extensionMap[normalized] ?? 'plaintext';
  }

  @override
  String detectFromContent(String content) {
    if (content.isEmpty) return 'plaintext';

    // Shebang
    if (content.startsWith('#!')) {
      final firstLine = content.split('\n').first.toLowerCase();
      if (firstLine.contains('python')) return 'python';
      if (firstLine.contains('ruby')) return 'ruby';
      if (firstLine.contains('node')) return 'javascript';
      if (firstLine.contains('bash') || firstLine.contains('sh')) {
        return 'shell';
      }
    }

    // Simple heuristics
    if (content.contains('import \'package:') ||
        content.contains('import "package:')) {
      return 'dart';
    }
    if (content.contains('<?php')) return 'php';
    if (content.contains('<!DOCTYPE html>') || content.contains('<html')) {
      return 'html';
    }
    if (content.trimLeft().startsWith('{') ||
        content.trimLeft().startsWith('[')) {
      return 'json';
    }
    if (content.contains('function ') ||
        content.contains('const ') ||
        content.contains('let ') ||
        content.contains('=>')) {
      return 'javascript';
    }
    if (content.contains('public class ') ||
        content.contains('private class ')) {
      return 'java';
    }

    return 'plaintext';
  }

  @override
  List<String> getSupportedLanguages() {
    return _extensionMap.values.toSet().toList()..sort();
  }

  @override
  String getFileExtension(String language) {
    final normalized = language.toLowerCase();
    return _languageToExtension[normalized] ?? '.txt';
  }
}
