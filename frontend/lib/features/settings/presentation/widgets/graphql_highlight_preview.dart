import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:highlight_selectable/highlight_selectable.dart';
import 'package:highlight_selectable/theme_map.dart';
import '../../application/settings_service.dart';

class GraphqlHighlightPreview extends StatefulWidget {
  const GraphqlHighlightPreview({super.key});

  @override
  State<GraphqlHighlightPreview> createState() =>
      _GraphqlHighlightPreviewState();
}

class _GraphqlHighlightPreviewState extends State<GraphqlHighlightPreview> {
  // Тестовый GraphQL-запрос
  static const String _sample = '''query GetUser(\$id: ID!) {
  user(id: \$id) {
    id
    name
    posts(limit: 5) {
      id
      title
    }
  }
}''';

  // Тестовый JSON-запрос GraphQL (оставлен как пример формата в HTTP body)

  static const String _sampleHtml = '''<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Example</title>
  </head>
  <body>
    <h1>Hello</h1>
    <p class="lead">Sample HTML snippet</p>
  </body>
</html>''';

  static const String _sampleXml = '''<?xml version="1.0" encoding="UTF-8"?>
<note>
  <to>User</to>
  <from>System</from>
  <heading>Reminder</heading>
  <body>XML example</body>
</note>''';

  static const String _sampleJs = r'''function greet(name) {
  console.log(`Hi, ${name}!`);
}
greet('World');''';

  static const String _sampleTs = '''type User = { id: string; name: string };
const u: User = { id: '1', name: 'Alice' };
console.log(u);''';

  static const String _sampleCss = '''body {
  background: #222;
  color: #eee;
}
.btn {
  padding: 8px 12px;
  border-radius: 6px;
}''';

  late String _lightTheme;
  late String _darkTheme;
  bool _selectable = true;
  String _lang = 'graphql';
  String _code = _sample;

  @override
  void initState() {
    super.initState();
    _lightTheme = 'a11y-light';
    _darkTheme = 'a11y-dark';
    // загрузим сохранённые темы с бэкенда, если есть
    SettingsService().fetchRuntime().then((data) {
      if (!mounted) return;
      final savedLight = (data['highlightThemeLight'] ?? '').toString();
      final savedDark = (data['highlightThemeDark'] ?? '').toString();
      final legacy = (data['highlightTheme'] ?? '').toString();
      final candidateLight =
          savedLight.isNotEmpty
              ? savedLight
              : (legacy.isNotEmpty ? legacy : _lightTheme);
      final candidateDark =
          savedDark.isNotEmpty
              ? savedDark
              : (legacy.isNotEmpty ? legacy : _darkTheme);
      setState(() {
        _lightTheme =
            themeMap.containsKey(candidateLight) ? candidateLight : _lightTheme;
        _darkTheme =
            themeMap.containsKey(candidateDark) ? candidateDark : _darkTheme;
      });
    });
    if (!themeMap.containsKey(_lightTheme)) _lightTheme = themeMap.keys.first;
    if (!themeMap.containsKey(_darkTheme)) _darkTheme = themeMap.keys.first;
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final active = isDark ? _darkTheme : _lightTheme;
    final themeStyles = themeMap[active]!;
    final bgColor = themeStyles['root']?.backgroundColor;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          runSpacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            const Text('Light theme:'),
            DropdownButton<String>(
              value: _lightTheme,
              items:
                  themeMap.keys
                      .map(
                        (k) =>
                            DropdownMenuItem<String>(value: k, child: Text(k)),
                      )
                      .toList(),
              onChanged: (v) async {
                final next = v ?? _lightTheme;
                setState(() => _lightTheme = next);
                await SettingsService().saveHighlightThemeLight(_lightTheme);
              },
            ),
            const SizedBox(width: 12),
            const Text('Dark theme:'),
            DropdownButton<String>(
              value: _darkTheme,
              items:
                  themeMap.keys
                      .map(
                        (k) =>
                            DropdownMenuItem<String>(value: k, child: Text(k)),
                      )
                      .toList(),
              onChanged: (v) async {
                final next = v ?? _darkTheme;
                setState(() => _darkTheme = next);
                await SettingsService().saveHighlightThemeDark(_darkTheme);
              },
            ),
            const SizedBox(width: 12),
            Switch(
              value: _selectable,
              onChanged: (v) => setState(() => _selectable = v),
            ),
            const Text('Selectable'),
            const SizedBox(width: 16),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Text('Language:'),
                const SizedBox(width: 8),
                DropdownButton<String>(
                  value: _lang,
                  items: const [
                    DropdownMenuItem(value: 'graphql', child: Text('GraphQL')),
                    DropdownMenuItem(value: 'html', child: Text('HTML')),
                    DropdownMenuItem(value: 'xml', child: Text('XML')),
                    DropdownMenuItem(
                      value: 'javascript',
                      child: Text('JavaScript'),
                    ),
                    DropdownMenuItem(
                      value: 'typescript',
                      child: Text('TypeScript'),
                    ),
                    DropdownMenuItem(value: 'css', child: Text('CSS')),
                  ],
                  onChanged: (v) {
                    final next = v ?? _lang;
                    setState(() {
                      _lang = next;
                      _code = switch (next) {
                        'html' => _sampleHtml,
                        'xml' => _sampleXml,
                        'javascript' => _sampleJs,
                        'typescript' => _sampleTs,
                        'css' => _sampleCss,
                        _ => _sample,
                      };
                    });
                  },
                ),
              ],
            ),
          ],
        ),
        const SizedBox(height: 8),
        const SizedBox(height: 12),
        const SizedBox(height: 8),
        Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Theme.of(context).dividerColor),
            color: bgColor, // фон совпадает с темой подсветки
          ),
          clipBehavior: Clip.antiAlias,
          child: Stack(
            children: [
              Padding(
                padding: const EdgeInsets.all(12),
                child: HighlightSelectable(
                  key: ValueKey(
                    'hs:${active}|${_lang}|${_selectable}|${_code.hashCode}',
                  ),
                  _code,
                  language: _lang,
                  theme: themeStyles,
                  selectable: _selectable,
                  showCopyButton: false,
                  showEditButton: false,
                  onChanged: (v) => setState(() => _code = v),
                ),
              ),
              Positioned(
                top: 4,
                right: 4,
                child: Row(
                  children: [
                    IconButton(
                      tooltip: 'Copy',
                      visualDensity: VisualDensity.compact,
                      onPressed:
                          () => Clipboard.setData(ClipboardData(text: _code)),
                      icon: const Icon(Icons.copy, size: 18),
                    ),
                    IconButton(
                      tooltip: 'Edit',
                      visualDensity: VisualDensity.compact,
                      onPressed: () async {
                        final ctrl = TextEditingController(text: _code);
                        final ok = await showDialog<bool>(
                          context: context,
                          builder: (ctx) {
                            return AlertDialog(
                              title: const Text('Edit code'),
                              content: SizedBox(
                                width: 720,
                                child: TextField(
                                  controller: ctrl,
                                  maxLines: 20,
                                  minLines: 8,
                                  style: Theme.of(context).textTheme.bodyMedium
                                      ?.copyWith(fontFamily: 'monospace'),
                                ),
                              ),
                              actions: [
                                TextButton(
                                  onPressed: () => Navigator.of(ctx).pop(false),
                                  child: const Text('Cancel'),
                                ),
                                ElevatedButton(
                                  onPressed: () => Navigator.of(ctx).pop(true),
                                  child: const Text('Save'),
                                ),
                              ],
                            );
                          },
                        );
                        if (ok == true) setState(() => _code = ctrl.text);
                      },
                      icon: const Icon(Icons.edit, size: 18),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
