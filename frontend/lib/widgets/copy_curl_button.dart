import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Режим экспорта cURL команды
enum CurlExportMode {
  /// Компактный однострочный формат
  compact,

  /// Многострочный формат с backslash
  multiline,

  /// Многострочный с доп. опциями (--location, --insecure)
  withOptions,
}

/// Кнопка копирования cURL с выпадающим меню форматов.
/// Переиспользуется в Inspector и Compose.
class CopyCurlButton extends StatelessWidget {
  const CopyCurlButton({
    super.key,
    required this.url,
    required this.method,
    required this.headers,
    this.body = '',
    this.form,
    this.onCopied,
    this.showSnackBar = true,
  });

  /// URL запроса
  final String url;

  /// HTTP метод (GET, POST, etc.)
  final String method;

  /// Заголовки запроса
  final Map<String, String> headers;

  /// Тело запроса (raw string)
  final String body;

  /// Form data (для multipart/urlencoded)
  final Map<String, dynamic>? form;

  /// Callback после копирования (опционально)
  final void Function(String curl)? onCopied;

  /// Показывать SnackBar после копирования
  final bool showSnackBar;

  @override
  Widget build(BuildContext context) {
    return MenuAnchor(
      builder: (context, controller, child) {
        return OutlinedButton(
          onPressed: () {
            if (controller.isOpen) {
              controller.close();
            } else {
              controller.open();
            }
          },
          style: OutlinedButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            textStyle: const TextStyle(fontSize: 12),
            minimumSize: Size.zero,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.content_paste, size: 14),
              const SizedBox(width: 4),
              const Text('cURL', style: TextStyle(fontSize: 12)),
              const SizedBox(width: 2),
              Icon(
                controller.isOpen ? Icons.arrow_drop_up : Icons.arrow_drop_down,
                size: 16,
              ),
            ],
          ),
        );
      },
      menuChildren: [
        MenuItemButton(
          leadingIcon: const Icon(Icons.remove, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.compact),
          child: const Text('Compact'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.wrap_text, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.multiline),
          child: const Text('Multiline'),
        ),
        MenuItemButton(
          leadingIcon: const Icon(Icons.settings, size: 16),
          onPressed: () => _copyWithMode(context, CurlExportMode.withOptions),
          child: const Text('With options'),
        ),
      ],
    );
  }

  void _copyWithMode(BuildContext context, CurlExportMode mode) {
    final curl = buildCurl(
      url: url,
      method: method,
      headers: headers,
      body: body,
      form: form,
      mode: mode,
    );
    Clipboard.setData(ClipboardData(text: curl));
    onCopied?.call(curl);

    if (showSnackBar) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('cURL copied')));
    }
  }

  /// Генерация cURL команды из данных запроса.
  /// Статический метод для возможности использования без виджета.
  static String buildCurl({
    required String url,
    required String method,
    required Map<String, String> headers,
    String body = '',
    Map<String, dynamic>? form,
    CurlExportMode mode = CurlExportMode.compact,
  }) {
    if (url.isEmpty) {
      return '# Error: URL is empty';
    }

    // Извлекаем cookies из заголовков
    String? cookieValue;
    final headersToInclude = <String, String>{};
    headers.forEach((k, v) {
      if (k.toLowerCase() == 'cookie') {
        cookieValue = v;
      } else {
        headersToInclude[k] = v;
      }
    });

    // Проверяем нужен ли form-data формат
    final contentType =
        headers['Content-Type'] ?? headers['content-type'] ?? '';
    final useFormData =
        form != null &&
        (contentType.contains('multipart/form-data') ||
            contentType.contains('application/x-www-form-urlencoded'));

    final buffer = StringBuffer();
    final isMultiline =
        mode == CurlExportMode.multiline || mode == CurlExportMode.withOptions;
    final newline = isMultiline ? ' \\\n  ' : ' ';

    buffer.write("curl -X $method$newline'");
    buffer.write(_escapeShellArg(url));
    buffer.write("'");

    // Добавляем заголовки
    headersToInclude.forEach((k, v) {
      // Пропускаем Content-Type и Content-Length для form-data
      if (useFormData &&
          (k.toLowerCase() == 'content-type' ||
              k.toLowerCase() == 'content-length')) {
        return;
      }
      buffer.write("$newline-H '$k: ${_escapeShellArg(v)}'");
    });

    // Добавляем cookie
    if ((cookieValue ?? '').isNotEmpty) {
      buffer.write("$newline--cookie '${_escapeShellArg(cookieValue!)}'");
    }

    // Добавляем body или form data
    if (useFormData) {
      _appendFormData(buffer, newline, form);
    } else if (body.isNotEmpty) {
      buffer.write("$newline--data '${_escapeShellArg(body)}'");
    }

    // Доп. опции для withOptions режима
    if (mode == CurlExportMode.withOptions) {
      buffer.write("$newline--location");
      buffer.write("$newline--insecure");
    }

    buffer.write("$newline--compressed");

    return buffer.toString();
  }

  static void _appendFormData(
    StringBuffer buffer,
    String newline,
    Map<String, dynamic>? form,
  ) {
    final fields =
        (form?['fields'] as List?)?.cast<Map<String, dynamic>>() ?? [];
    final files = (form?['files'] as List?)?.cast<Map<String, dynamic>>() ?? [];

    for (final field in fields) {
      final name = field['name']?.toString() ?? '';
      final value =
          field['valuePreview']?.toString() ?? field['value']?.toString() ?? '';
      if (name.isNotEmpty) {
        buffer.write("$newline-F '$name=${_escapeShellArg(value)}'");
      }
    }

    if (files.isNotEmpty) {
      buffer.write(
        "$newline# Note: Replace file paths with actual local paths",
      );
    }
    for (final file in files) {
      final name = file['name']?.toString() ?? '';
      final filename = file['filename']?.toString() ?? 'file';
      if (name.isNotEmpty) {
        buffer.write("$newline-F '$name=@$filename'");
      }
    }
  }

  static String _escapeShellArg(String arg) {
    return arg.replaceAll("'", "'\\''");
  }
}
