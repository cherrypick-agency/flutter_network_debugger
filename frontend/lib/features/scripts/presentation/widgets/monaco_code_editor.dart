import 'package:flutter/material.dart';
import 'package:flutter_code_editor/flutter_code_editor.dart';
import 'package:flutter_highlight/themes/vs.dart';
import 'package:flutter_highlight/themes/vs2015.dart';
import 'package:highlight/languages/dart.dart';
import 'package:highlight/languages/javascript.dart';
import 'package:highlight/languages/typescript.dart';
import 'package:highlight/languages/rust.dart';
import 'package:highlight/languages/go.dart';
import 'package:highlight/languages/cpp.dart';

/// Code Editor wrapper widget
/// Provides code editing with syntax highlighting for multiple languages
class MonacoCodeEditor extends StatefulWidget {
  final String code;
  final String language;
  final ValueChanged<String>? onChanged;
  final bool readOnly;

  const MonacoCodeEditor({
    super.key,
    required this.code,
    required this.language,
    this.onChanged,
    this.readOnly = false,
  });

  @override
  State<MonacoCodeEditor> createState() => _MonacoCodeEditorState();
}

class _MonacoCodeEditorState extends State<MonacoCodeEditor> {
  late final CodeController _controller;

  @override
  void initState() {
    super.initState();
    _controller = CodeController(
      text: widget.code,
      language: _getHighlightMode(widget.language),
    );

    if (widget.onChanged != null) {
      _controller.addListener(() {
        if (_controller.text != widget.code) {
          widget.onChanged!(_controller.text);
        }
      });
    }
  }

  @override
  void didUpdateWidget(MonacoCodeEditor oldWidget) {
    super.didUpdateWidget(oldWidget);

    // Update text only if it changed externally (not from user input)
    if (widget.code != oldWidget.code && widget.code != _controller.text) {
      _controller.text = widget.code;
    }

    // Update language if changed
    if (widget.language != oldWidget.language) {
      _controller.language = _getHighlightMode(widget.language);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final brightness = Theme.of(context).brightness;
    final theme = brightness == Brightness.dark ? vs2015Theme : vsTheme;

    return CodeTheme(
      data: CodeThemeData(styles: theme),
      child: SingleChildScrollView(
        child: CodeField(
          controller: _controller,
          enabled: !widget.readOnly,
          textStyle: const TextStyle(fontFamily: 'monospace', fontSize: 14),
        ),
      ),
    );
  }

  /// Map language name to highlight mode
  dynamic _getHighlightMode(String language) {
    return switch (language.toLowerCase()) {
      'dart' => dart,
      'javascript' => javascript,
      'typescript' => typescript,
      'rust' => rust,
      'go' => go,
      'c' => cpp, // C uses cpp highlighting
      'cpp' => cpp,
      _ => null,
    };
  }
}
