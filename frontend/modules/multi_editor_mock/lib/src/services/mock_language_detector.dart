import 'package:multi_editor_core/multi_editor_core.dart';

class MockLanguageDetector implements LanguageDetector {
  static final Map<String, String> _extensionToLanguage = {
    '.dart': 'dart',
    '.js': 'javascript',
    '.jsx': 'javascript',
    '.ts': 'typescript',
    '.tsx': 'typescript',
    '.py': 'python',
    '.java': 'java',
    '.cpp': 'cpp',
    '.cc': 'cpp',
    '.cxx': 'cpp',
    '.c': 'c',
    '.h': 'c',
    '.hpp': 'cpp',
    '.go': 'go',
    '.rs': 'rust',
    '.rb': 'ruby',
    '.php': 'php',
    '.swift': 'swift',
    '.kt': 'kotlin',
    '.kts': 'kotlin',
    '.cs': 'csharp',
    '.html': 'html',
    '.htm': 'html',
    '.css': 'css',
    '.scss': 'css',
    '.sass': 'css',
    '.json': 'json',
    '.yaml': 'yaml',
    '.yml': 'yaml',
    '.md': 'markdown',
    '.markdown': 'markdown',
    '.xml': 'xml',
    '.sql': 'sql',
    '.sh': 'shell',
    '.bash': 'shell',
    '.zsh': 'shell',
    '.txt': 'plaintext',
  };

  static final Map<String, String> _languageToExtension = {
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
  };

  @override
  String detectFromFileName(String fileName) {
    if (fileName.contains('.')) {
      final extension = '.${fileName.split('.').last}';
      return detectFromExtension(extension);
    }

    // Check for common files without extensions
    final lowerName = fileName.toLowerCase();
    if (lowerName == 'dockerfile') return 'dockerfile';
    if (lowerName == 'makefile') return 'makefile';
    if (lowerName == 'rakefile') return 'ruby';
    if (lowerName == 'gemfile') return 'ruby';
    if (lowerName == 'readme') return 'markdown';

    return 'plaintext';
  }

  @override
  String detectFromExtension(String extension) {
    final normalized = extension.toLowerCase();
    return _extensionToLanguage[normalized] ?? 'plaintext';
  }

  @override
  String detectFromContent(String content) {
    if (content.isEmpty) {
      return 'plaintext';
    }

    // Check shebang for scripts
    if (content.startsWith('#!')) {
      final firstLine = content.split('\n').first.toLowerCase();
      if (firstLine.contains('python')) return 'python';
      if (firstLine.contains('ruby')) return 'ruby';
      if (firstLine.contains('node')) return 'javascript';
      if (firstLine.contains('bash') || firstLine.contains('sh'))
        return 'shell';
    }

    // Check for common language patterns
    if (content.contains('import \'package:') ||
        content.contains('import "package:')) {
      return 'dart';
    }

    if (content.contains('<?php')) {
      return 'php';
    }

    if (content.contains('<!DOCTYPE html>') || content.contains('<html')) {
      return 'html';
    }

    if (content.contains('{') && content.contains('}')) {
      // Could be JSON
      try {
        if (content.trimLeft().startsWith('{') ||
            content.trimLeft().startsWith('[')) {
          return 'json';
        }
      } catch (_) {
        // Not JSON
      }
    }

    if (content.contains('class ') && content.contains('def ')) {
      return 'python';
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

    // Default to plaintext if can't detect
    return 'plaintext';
  }

  @override
  List<String> getSupportedLanguages() {
    return _extensionToLanguage.values.toSet().toList()..sort();
  }

  @override
  String getFileExtension(String language) {
    final normalized = language.toLowerCase();
    return _languageToExtension[normalized] ?? '.txt';
  }
}
