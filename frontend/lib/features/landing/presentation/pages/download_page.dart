import 'package:flutter/material.dart';
import '../../utils/os_detect.dart' as osd;
import '../../utils/open_url.dart' as nav;

class DownloadPage extends StatelessWidget {
  const DownloadPage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final os = osd.detectOS();
    final arch = osd.detectArch();

    void dl(String file) {
      nav.openUrl('assets/downloads/$file');
    }

    Widget card(String title, List<Widget> children) {
      return Card(
        elevation: 1,
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              ...children,
            ],
          ),
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Downloads')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: ListView(
          children: [
            card('Рекомендуемая сборка', [
              // Для Safari на macOS точную архитектуру получаем асинхронно
              FutureBuilder<String?>(
                future: osd.detectArchPrecise(),
                builder: (context, snap) {
                  final effectiveArch = (snap.data ?? arch);
                  final note =
                      os == 'mac'
                          ? ' (${osd.macArchLabel(effectiveArch)})'
                          : '';

                  String pickWith(String name) {
                    if (os == 'win')
                      return '${name}_windows_${effectiveArch}.zip';
                    if (os == 'mac')
                      return '${name}_darwin_${effectiveArch}.tar.gz';
                    return '${name}_linux_${effectiveArch}.tar.gz';
                  }

                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Определено: $os/$effectiveArch$note',
                        style: theme.textTheme.bodySmall,
                      ),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          ElevatedButton(
                            onPressed: () => dl(pickWith('network-debugger')),
                            child: const Text('Скачать network-debugger'),
                          ),
                          OutlinedButton(
                            onPressed:
                                () => dl(pickWith('network-debugger-web')),
                            child: const Text('Скачать network-debugger-web'),
                          ),
                        ],
                      ),
                    ],
                  );
                },
              ),
            ]),
            const SizedBox(height: 12),
            card('Выбрать вручную', [
              _Section(
                title: 'Windows (.zip)',
                items: const [
                  'network-debugger_windows_amd64.zip',
                  'network-debugger_windows_386.zip',
                  'network-debugger_windows_arm64.zip',
                  'network-debugger-web_windows_amd64.zip',
                  'network-debugger-web_windows_386.zip',
                  'network-debugger-web_windows_arm64.zip',
                ],
                onTap: (f) => dl(f),
              ),
              const SizedBox(height: 8),
              _Section(
                title: 'macOS (.tar.gz)',
                items: const [
                  'network-debugger_darwin_amd64.tar.gz',
                  'network-debugger_darwin_arm64.tar.gz',
                  'network-debugger-web_darwin_amd64.tar.gz',
                  'network-debugger-web_darwin_arm64.tar.gz',
                ],
                onTap: (f) => dl(f),
              ),
              const SizedBox(height: 8),
              _Section(
                title: 'Linux (.tar.gz)',
                items: const [
                  'network-debugger_linux_amd64.tar.gz',
                  'network-debugger_linux_arm64.tar.gz',
                  'network-debugger-web_linux_amd64.tar.gz',
                  'network-debugger-web_linux_arm64.tar.gz',
                ],
                onTap: (f) => dl(f),
              ),
            ]),
            const SizedBox(height: 12),
            card('Как запускать', [
              const SelectableText('''
tar -xzf network-debugger_<os>_<arch>.tar.gz
./network-debugger

tar -xzf network-debugger-web_<os>_<arch>.tar.gz
./network-debugger-web
              '''),
            ]),
          ],
        ),
      ),
    );
  }
}

class _Section extends StatelessWidget {
  const _Section({
    required this.title,
    required this.items,
    required this.onTap,
  });
  final String title;
  final List<String> items;
  final void Function(String) onTap;
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 4),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            for (final f in items)
              OutlinedButton(
                onPressed:
                    () => onTap(
                      'assets/downloads/$f'.split('assets/downloads/').last,
                    ),
                child: Text(f, style: const TextStyle(fontSize: 12)),
              ),
          ],
        ),
      ],
    );
  }
}
