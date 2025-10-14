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
  // recent window controls
  bool _recentEnabled = false;
  final TextEditingController _recentMinutesCtrl = TextEditingController(
    text: '5',
  );

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
      _recentEnabled = (data['recentWindowEnabled'] ?? 'false') == 'true';
      _recentMinutesCtrl.text = (data['recentWindowMinutes'] ?? '5');
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
      final minutes = int.tryParse(_recentMinutesCtrl.text.trim()) ?? 5;
      await SettingsService().saveRecentWindow(
        enabled: _recentEnabled,
        minutes: minutes.clamp(1, 1440),
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
        child: LayoutBuilder(
          builder: (context, c) {
            final twoCols = c.maxWidth >= 900;
            final left = <Widget>[
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                dense: true,
                value: _enabled,
                onChanged: (v) {
                  setState(() {
                    _enabled = (v ?? false);
                  });
                },
                title: const Text('Response delay'),
                subtitle: const Text(
                  'Artificial response delay for all proxied requests',
                ),
              ),
              TextField(
                controller: _delayCtrl,
                enabled: _enabled,
                decoration: const InputDecoration(
                  labelText: 'Response delay (ms or range)',
                  hintText: 'e.g.: 1000 or 1000-3000',
                  helperText: 'Leave empty or 0 to disable',
                  isDense: true,
                ),
              ),
            ];
            final right = <Widget>[
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                dense: true,
                value: _recentEnabled,
                onChanged: (v) => setState(() => _recentEnabled = (v ?? false)),
                title: const Text('Show only last N minutes'),
                subtitle: const Text(
                  'Show only last N minutes in Sessions and Timeline',
                ),
              ),
              TextField(
                controller: _recentMinutesCtrl,
                enabled: _recentEnabled,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                  labelText: 'N minutes',
                  hintText: 'e.g.: 5',
                  isDense: true,
                ),
              ),
            ];
            if (!twoCols) {
              return SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [...left, const SizedBox(height: 16), ...right],
                ),
              );
            }
            return Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: left,
                  ),
                ),
                const SizedBox(width: 24),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: right,
                  ),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}
