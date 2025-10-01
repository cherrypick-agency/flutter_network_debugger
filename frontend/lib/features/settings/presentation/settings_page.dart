import 'package:flutter/material.dart';
import '../application/settings_service.dart';
import '../../../services/prefs.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});
  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  final TextEditingController _delayCtrl = TextEditingController();
  bool _enabled = false;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _delayCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final data = await PrefsService().load();
    setState(() {
      _enabled = (data['respDelayEnabled'] ?? 'false') == 'true';
      _delayCtrl.text = data['respDelayValue'] ?? '';
    });
  }

  Future<void> _save() async {
    setState(() {
      _saving = true;
    });
    try {
      await SettingsService().saveResponseDelay(
        enabled: _enabled,
        value: _delayCtrl.text,
      );
      if (mounted) Navigator.of(context).pop();
    } finally {
      if (mounted)
        setState(() {
          _saving = false;
        });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child:
                _saving
                    ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                    : const Text('Save'),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            CheckboxListTile(
              contentPadding: EdgeInsets.zero,
              dense: true,
              value: _enabled,
              onChanged: (v) {
                setState(() {
                  _enabled = v ?? false;
                });
              },
              title: const Text('Response delay'),
              subtitle: const Text(
                'Искусственная задержка ответа для всех проксируемых запросов',
              ),
            ),
            TextField(
              controller: _delayCtrl,
              enabled: _enabled,
              decoration: const InputDecoration(
                labelText: 'Response delay (ms or range)',
                hintText: 'например: 1000 или 1000-3000',
                helperText: 'Оставьте пустым или 0, чтобы отключить',
                isDense: true,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
