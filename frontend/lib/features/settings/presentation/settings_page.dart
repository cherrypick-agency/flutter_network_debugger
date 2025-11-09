import 'package:flutter/material.dart';
import '../application/settings_service.dart';
import '../../../services/prefs.dart';
import '../application/throttle_service.dart';
import '../../../theme/font_scale.dart';
import 'widgets/graphql_highlight_preview.dart';
import '../../custom_fonts/presentation/providers/font_provider.dart';
import '../../custom_fonts/presentation/widgets/custom_font_selector.dart';

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
  final TextEditingController _adminTokenCtrl = TextEditingController();

  // throttling profiles
  bool _thEnabled = false;
  int _thDown = 0;
  int _thUp = 0;
  int _thLoss = 0;
  bool _thOffline = false;
  List<ThrottleProfileDTO> _profiles = const [];
  String? _selectedProfileId; // custom profile selection
  double _fontScale = 1.0;

  // proxy runtime
  bool _fwdEnabled = true;
  int _fwdPort = 9091;
  bool _socksEnabled = false;
  int _socksPort = 8889;
  String _socksAuthMode = 'none';
  final TextEditingController _socksUserCtrl = TextEditingController();
  final TextEditingController _socksPassCtrl = TextEditingController();

  // custom fonts
  late final FontProvider _fontProvider;

  Future<void> _showAddProfileDialog(BuildContext context) async {
    final nameCtrl = TextEditingController();
    final downCtrl = TextEditingController(text: '$_thDown');
    final upCtrl = TextEditingController(text: '$_thUp');
    final lossCtrl = TextEditingController(text: '$_thLoss');
    bool offline = _thOffline;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          title: const Text('Add profile'),
          content: SizedBox(
            width: 420,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: nameCtrl,
                  decoration: const InputDecoration(labelText: 'Name'),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: downCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Down (kbit/s)',
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: TextField(
                        controller: upCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Up (kbit/s)',
                        ),
                      ),
                    ),
                  ],
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: lossCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Packet loss %',
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: StatefulBuilder(
                        builder: (context, setDialogState) {
                          return CheckboxListTile(
                            contentPadding: EdgeInsets.zero,
                            dense: true,
                            value: offline,
                            onChanged: (v) {
                              setDialogState(() {
                                offline = (v ?? false);
                              });
                            },
                            title: const Text('Offline'),
                          );
                        },
                      ),
                    ),
                  ],
                ),
              ],
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
    if (ok != true) return;
    final name = nameCtrl.text.trim();
    if (name.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Profile name is required')),
        );
      }
      return;
    }
    final down = int.tryParse(downCtrl.text.trim()) ?? 0;
    final up = int.tryParse(upCtrl.text.trim()) ?? 0;
    final loss = (int.tryParse(lossCtrl.text.trim()) ?? 0).clamp(0, 100);
    try {
      final saved = await ThrottleService().upsert(
        ThrottleProfileDTO(
          name: name,
          downKbps: down,
          upKbps: up,
          packetLossPct: loss,
          latencyMs: 0,
          latencyJitter: 0,
          offline: offline,
        ),
      );
      final ps = await ThrottleService().listProfiles();
      if (!mounted) return;
      setState(() {
        _profiles = ps;
        // Only set selected if saved.id is not null and exists in the list
        if (saved.id != null && ps.any((p) => p.id == saved.id)) {
          _selectedProfileId = saved.id;
        } else {
          _selectedProfileId = null;
        }
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Profile "$name" saved (total: ${ps.length}, with ID: ${ps.where((p) => p.id != null).length})',
            ),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Failed to save profile: $e')));
      }
    }
  }

  @override
  void initState() {
    super.initState();
    _fontProvider = FontProvider();
    _load();
  }

  @override
  void dispose() {
    _delayCtrl.dispose();
    _adminTokenCtrl.dispose();
    _socksUserCtrl.dispose();
    _socksPassCtrl.dispose();
    _fontProvider.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final data = await PrefsService().load();
    setState(() {
      _enabled = (data['respDelayEnabled'] ?? 'false') == 'true';
      _delayCtrl.text = data['respDelayValue'] ?? '';
      _recentEnabled = (data['recentWindowEnabled'] ?? 'false') == 'true';
      _recentMinutesCtrl.text = (data['recentWindowMinutes'] ?? '5');
      _adminTokenCtrl.text = data['adminToken'] ?? '';
    });

    // load throttling current + profiles
    try {
      final t = await ThrottleService().current();
      final ps = await ThrottleService().listProfiles();
      setState(() {
        _thEnabled = (t['enabled'] ?? false) as bool;
        _thDown = (t['downKbps'] ?? 0) as int;
        _thUp = (t['upKbps'] ?? 0) as int;
        _thLoss = (t['packetLossPct'] ?? 0) as int;
        _thOffline = (t['offline'] ?? false) as bool;
        _profiles = ps;
        // Reset selected profile if it doesn't exist in the loaded profiles
        if (_selectedProfileId != null &&
            !ps.any((p) => p.id == _selectedProfileId)) {
          _selectedProfileId = null;
        }
      });
    } catch (_) {}
    // load font scale
    try {
      _fontScale = await PrefsService().loadFontScale();
      FontScale.value.value = _fontScale;
    } catch (_) {}

    // load proxy config
    try {
      final pc = await SettingsService().fetchProxyConfig();
      setState(() {
        _fwdEnabled = (pc['forward']?['enabled'] ?? true) as bool;
        _fwdPort = (pc['forward']?['port'] ?? 9091) as int;
        _socksEnabled = (pc['socks']?['enabled'] ?? false) as bool;
        _socksPort = (pc['socks']?['port'] ?? 8889) as int;
        _socksAuthMode = (pc['socks']?['authMode'] ?? 'none') as String;
      });
    } catch (_) {}
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
      await PrefsService().saveAdminToken(_adminTokenCtrl.text);
      // сохранить и применить порты прокси
      await SettingsService().saveProxyConfig(
        forwardEnabled: _fwdEnabled,
        forwardPort: _fwdPort,
        socksEnabled: _socksEnabled,
        socksPort: _socksPort,
        authMode: _socksAuthMode,
        user: _socksUserCtrl.text.trim(),
        pass: _socksPassCtrl.text.trim(),
      );
      // применить throttling настройки
      await ThrottleService().apply(
        enabled: _thEnabled,
        downKbps: _thOffline ? 0 : _thDown,
        upKbps: _thOffline ? 0 : _thUp,
        packetLossPct: _thOffline ? 0 : _thLoss,
        offline: _thOffline,
      );
      if (mounted) Navigator.of(context).pop();
    } finally {
      if (mounted)
        setState(() {
          _saving = false;
        });
    }
  }

  void _resetResponseDelay() {
    setState(() {
      _enabled = false;
      _delayCtrl.text = '';
    });
  }

  void _resetRecent() {
    setState(() {
      _recentEnabled = false;
      _recentMinutesCtrl.text = '5';
    });
  }

  void _resetProxy() {
    setState(() {
      _fwdEnabled = true;
      _fwdPort = 8888;
      _socksEnabled = false;
      _socksPort = 8889;
      _socksAuthMode = 'none';
      _socksUserCtrl.text = '';
      _socksPassCtrl.text = '';
    });
  }

  void _resetNetwork() {
    setState(() {
      _thEnabled = false;
      _thDown = 0;
      _thUp = 0;
      _thLoss = 0;
      _thOffline = false;
      _enabled = false;
      _delayCtrl.text = '';
    });
  }

  void _resetAdmin() {
    setState(() {
      _adminTokenCtrl.text = '';
    });
  }

  void _resetAll() {
    _resetResponseDelay();
    _resetRecent();
    _resetProxy();
    _resetNetwork();
    _resetAdmin();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
        actions: [
          TextButton(
            onPressed: _saving ? null : _resetAll,
            child: const Text('Reset'),
          ),
          TextButton(
            onPressed: _saving ? null : _save,
            child: _saving
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
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(
                            'Network throttling',
                            style: Theme.of(context).textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          TextButton(
                            onPressed: _resetNetwork,
                            child: const Text('Reset'),
                          ),
                        ],
                      ),
                      SwitchListTile(
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                        value: _thEnabled,
                        onChanged: (v) => setState(() => _thEnabled = (v)),
                        title: const Text('Enable throttling'),
                      ),
                      Row(
                        children: [
                          Expanded(
                            child: TextField(
                              enabled: _thEnabled && !_thOffline,
                              keyboardType: TextInputType.number,
                              decoration: const InputDecoration(
                                labelText: 'Down (kbit/s)',
                                isDense: true,
                              ),
                              controller:
                                  TextEditingController(text: '$_thDown')
                                    ..selection = TextSelection.collapsed(
                                      offset: '$_thDown'.length,
                                    ),
                              onChanged: (v) =>
                                  _thDown = int.tryParse(v.trim()) ?? 0,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: TextField(
                              enabled: _thEnabled && !_thOffline,
                              keyboardType: TextInputType.number,
                              decoration: const InputDecoration(
                                labelText: 'Up (kbit/s)',
                                isDense: true,
                              ),
                              controller: TextEditingController(text: '$_thUp')
                                ..selection = TextSelection.collapsed(
                                  offset: '$_thUp'.length,
                                ),
                              onChanged: (v) =>
                                  _thUp = int.tryParse(v.trim()) ?? 0,
                            ),
                          ),
                        ],
                      ),
                      Row(
                        children: [
                          Expanded(
                            child: TextField(
                              enabled: _thEnabled && !_thOffline,
                              keyboardType: TextInputType.number,
                              decoration: const InputDecoration(
                                labelText: 'Packet loss %',
                                isDense: true,
                              ),
                              controller:
                                  TextEditingController(text: '$_thLoss')
                                    ..selection = TextSelection.collapsed(
                                      offset: '$_thLoss'.length,
                                    ),
                              onChanged: (v) => _thLoss =
                                  int.tryParse(v.trim())?.clamp(0, 100) ?? 0,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: CheckboxListTile(
                              contentPadding: EdgeInsets.zero,
                              dense: true,
                              title: const Text('Simulate offline'),
                              value: _thOffline,
                              onChanged: (v) =>
                                  setState(() => _thOffline = (v ?? false)),
                            ),
                          ),
                        ],
                      ),
                      Row(
                        children: [
                          const Text(
                            'Quick presets:',
                            style: TextStyle(fontWeight: FontWeight.w500),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              children: [
                                OutlinedButton(
                                  onPressed: () async {
                                    await ThrottleService().apply(
                                      enabled: true,
                                      downKbps: 400,
                                      upKbps: 400,
                                      packetLossPct: 0,
                                      offline: false,
                                    );
                                    setState(() {
                                      _thEnabled = true;
                                      _thDown = 400;
                                      _thUp = 400;
                                      _thLoss = 0;
                                      _thOffline = false;
                                    });
                                  },
                                  child: const Text('3G'),
                                ),
                                OutlinedButton(
                                  onPressed: () async {
                                    await ThrottleService().apply(
                                      enabled: true,
                                      downKbps: 1500,
                                      upKbps: 1500,
                                      packetLossPct: 0,
                                      offline: false,
                                    );
                                    setState(() {
                                      _thEnabled = true;
                                      _thDown = 1500;
                                      _thUp = 1500;
                                      _thLoss = 0;
                                      _thOffline = false;
                                    });
                                  },
                                  child: const Text('Slow 4G'),
                                ),
                                OutlinedButton(
                                  onPressed: () async {
                                    await ThrottleService().apply(
                                      enabled: true,
                                      downKbps: 3000,
                                      upKbps: 3000,
                                      packetLossPct: 0,
                                      offline: false,
                                    );
                                    setState(() {
                                      _thEnabled = true;
                                      _thDown = 3000;
                                      _thUp = 3000;
                                      _thLoss = 0;
                                      _thOffline = false;
                                    });
                                  },
                                  child: const Text('Fast 4G'),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const Divider(height: 24),
                      Row(
                        children: [
                          Text(
                            'Saved profiles (${_profiles.where((p) => p.id != null).length}):',
                            style: const TextStyle(fontWeight: FontWeight.w500),
                          ),
                          const SizedBox(width: 12),
                          if (_profiles.where((p) => p.id != null).isNotEmpty)
                            DropdownButton<String>(
                              key: ValueKey(
                                _profiles
                                    .where((p) => p.id != null)
                                    .map((p) => p.id)
                                    .join(','),
                              ),
                              hint: const Text('Select profile'),
                              value:
                                  _profiles
                                      .where((p) => p.id != null)
                                      .any((p) => p.id == _selectedProfileId)
                                  ? _selectedProfileId
                                  : null,
                              items: _profiles
                                  .where((p) => p.id != null)
                                  .map(
                                    (p) => DropdownMenuItem(
                                      value: p.id!,
                                      child: Text(p.name),
                                    ),
                                  )
                                  .toList(),
                              onChanged: (id) {
                                if (id == null) {
                                  setState(() => _selectedProfileId = null);
                                  return;
                                }
                                final p = _profiles.firstWhere(
                                  (e) => e.id == id,
                                  orElse: () => _profiles.first,
                                );
                                setState(() {
                                  _selectedProfileId = id;
                                  _thEnabled = true;
                                  _thDown = p.downKbps;
                                  _thUp = p.upKbps;
                                  _thLoss = p.packetLossPct;
                                  _thOffline = p.offline;
                                });
                              },
                            )
                          else
                            const Text(
                              'No saved profiles',
                              style: TextStyle(
                                fontStyle: FontStyle.italic,
                                color: Colors.grey,
                              ),
                            ),
                          IconButton(
                            tooltip: 'Add new profile',
                            onPressed: () => _showAddProfileDialog(context),
                            icon: const Icon(Icons.add_circle_outline),
                            iconSize: 20,
                          ),
                          if (_selectedProfileId != null)
                            IconButton(
                              tooltip: 'Delete selected profile',
                              onPressed: () async {
                                final id = _selectedProfileId!;
                                await ThrottleService().deleteProfile(id);
                                final ps = await ThrottleService()
                                    .listProfiles();
                                if (!mounted) return;
                                setState(() {
                                  _profiles = ps;
                                  _selectedProfileId = null;
                                });
                              },
                              icon: const Icon(Icons.delete_outline),
                              iconSize: 20,
                            ),
                        ],
                      ),
                      const Divider(height: 24),
                      Row(
                        children: [
                          Text(
                            'Response delay',
                            style: Theme.of(context).textTheme.titleSmall
                                ?.copyWith(fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          TextButton(
                            onPressed: _resetResponseDelay,
                            child: const Text('Reset'),
                          ),
                        ],
                      ),
                      CheckboxListTile(
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                        value: _enabled,
                        onChanged: (v) {
                          setState(() {
                            _enabled = (v ?? false);
                          });
                        },
                        title: const Text('Enable artificial delay'),
                        subtitle: const Text('Applies to all proxied requests'),
                      ),
                      TextField(
                        controller: _delayCtrl,
                        enabled: _enabled,
                        decoration: const InputDecoration(
                          labelText: 'Delay (ms or range)',
                          hintText: 'e.g.: 1000 or 1000-3000',
                          isDense: true,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ];
            final right = <Widget>[
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(
                            'Typography',
                            style: Theme.of(context).textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          TextButton(
                            onPressed: () async {
                              setState(() {
                                _fontScale = 1.0;
                              });
                              FontScale.value.value = 1.0;
                              await SettingsService().saveFontScale(1.0);
                            },
                            child: const Text('Reset'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          const Text('Font size'),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Slider(
                              value: _fontScale,
                              min: 0.8,
                              max: 1.6,
                              divisions: 16,
                              label: _fontScale.toStringAsFixed(2),
                              onChanged: (v) {
                                setState(() {
                                  _fontScale = double.parse(
                                    v.toStringAsFixed(2),
                                  );
                                });
                                FontScale.value.value = _fontScale;
                              },
                              onChangeEnd: (v) async {
                                await SettingsService().saveFontScale(v);
                              },
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                      CustomFontSelector(provider: _fontProvider),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Code highlighting',
                        style: Theme.of(context).textTheme.titleMedium
                            ?.copyWith(fontWeight: FontWeight.w600),
                      ),
                      const SizedBox(height: 8),
                      const GraphqlHighlightPreview(),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      CheckboxListTile(
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                        value: _recentEnabled,
                        onChanged: (v) =>
                            setState(() => _recentEnabled = (v ?? false)),
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
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(
                            'Proxy',
                            style: Theme.of(context).textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          TextButton(
                            onPressed: _resetProxy,
                            child: const Text('Reset'),
                          ),
                        ],
                      ),
                      SwitchListTile(
                        dense: true,
                        contentPadding: EdgeInsets.zero,
                        value: _fwdEnabled,
                        onChanged: (v) => setState(() => _fwdEnabled = v),
                        title: const Text('Enable HTTP forward proxy'),
                      ),
                      TextField(
                        enabled: _fwdEnabled,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'HTTP proxy port',
                          isDense: true,
                        ),
                        controller: TextEditingController(text: '$_fwdPort')
                          ..selection = TextSelection.collapsed(
                            offset: '$_fwdPort'.length,
                          ),
                        onChanged: (v) => _fwdPort =
                            int.tryParse(v.trim())?.clamp(0, 65535) ?? 0,
                      ),
                      const SizedBox(height: 8),
                      SwitchListTile(
                        dense: true,
                        contentPadding: EdgeInsets.zero,
                        value: _socksEnabled,
                        onChanged: (v) => setState(() => _socksEnabled = v),
                        title: const Text('Enable SOCKS5 proxy'),
                      ),
                      TextField(
                        enabled: _socksEnabled,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'SOCKS5 port',
                          isDense: true,
                        ),
                        controller: TextEditingController(text: '$_socksPort')
                          ..selection = TextSelection.collapsed(
                            offset: '$_socksPort'.length,
                          ),
                        onChanged: (v) => _socksPort =
                            int.tryParse(v.trim())?.clamp(0, 65535) ?? 0,
                      ),
                      DropdownButtonFormField<String>(
                        value: _socksAuthMode,
                        decoration: const InputDecoration(
                          isDense: true,
                          labelText: 'SOCKS auth',
                        ),
                        items: const [
                          DropdownMenuItem(value: 'none', child: Text('none')),
                          DropdownMenuItem(
                            value: 'userpass',
                            child: Text('user/pass'),
                          ),
                        ],
                        onChanged: !_socksEnabled
                            ? null
                            : (v) =>
                                  setState(() => _socksAuthMode = v ?? 'none'),
                      ),
                      if (_socksEnabled && _socksAuthMode == 'userpass') ...[
                        TextField(
                          controller: _socksUserCtrl,
                          decoration: const InputDecoration(
                            labelText: 'User',
                            isDense: true,
                          ),
                        ),
                        TextField(
                          controller: _socksPassCtrl,
                          decoration: const InputDecoration(
                            labelText: 'Password',
                            isDense: true,
                          ),
                          obscureText: true,
                        ),
                      ],
                    ],
                  ),
                ),
              ),
              // Admin token moved to the bottom card
              const SizedBox.shrink(),
            ];
            // Admin token card at bottom
            right.add(const SizedBox(height: 12));
            right.add(
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(
                            'Admin token',
                            style: Theme.of(context).textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                          ),
                          const Spacer(),
                          TextButton(
                            onPressed: _resetAdmin,
                            child: const Text('Reset'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: _adminTokenCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Admin token (X-Admin-Token)',
                          hintText:
                              'Optional token to protect sensitive endpoints',
                          isDense: true,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
            if (!twoCols) {
              return SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [...left, const SizedBox(height: 16), ...right],
                ),
              );
            }
            return SingleChildScrollView(
              child: Row(
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
              ),
            );
          },
        ),
      ),
    );
  }
}
