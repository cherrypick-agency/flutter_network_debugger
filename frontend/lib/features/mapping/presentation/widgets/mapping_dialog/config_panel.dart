import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../core/widgets/inline_status_banner.dart';
import '../../../application/stores/mapping_config_store.dart';
import '../../../domain/mapping_config.dart';

class MappingConfigPanel extends StatefulWidget {
  const MappingConfigPanel({super.key});

  @override
  State<MappingConfigPanel> createState() => _MappingConfigPanelState();
}

class _MappingConfigPanelState extends State<MappingConfigPanel> {
  late final TextEditingController _uploadCtrl;
  late final FocusNode _uploadFocus;
  late final MappingConfigStore _store;
  int _lastValidUploadMaxMB = 20;

  @override
  void initState() {
    super.initState();
    _uploadFocus = FocusNode();
    _uploadCtrl = TextEditingController(text: '20');
    _store = context.read<MappingConfigStore>();
    _store.addListener(_syncFromStore);
    _syncFromStore();

    _uploadFocus.addListener(() {
      if (_uploadFocus.hasFocus) return;
      final n = _parsePositiveInt(_uploadCtrl.text);
      if (n != null) {
        _lastValidUploadMaxMB = n;
        return;
      }
      _uploadCtrl.text = _lastValidUploadMaxMB.toString();
      setState(() {});
    });
  }

  void _syncFromStore() {
    final cfg = _store.config;
    if (cfg == null) return;
    if (_uploadFocus.hasFocus) return;
    final desired = cfg.uploadMaxMB.toString();
    if (_uploadCtrl.text != desired) {
      _uploadCtrl.text = desired;
    }
    if (cfg.uploadMaxMB > 0) _lastValidUploadMaxMB = cfg.uploadMaxMB;
  }

  @override
  void dispose() {
    _store.removeListener(_syncFromStore);
    _uploadFocus.dispose();
    _uploadCtrl.dispose();
    super.dispose();
  }

  int? _parsePositiveInt(String v) {
    final n = int.tryParse(v.trim());
    if (n == null || n <= 0) return null;
    return n;
  }

  @override
  Widget build(BuildContext context) {
    final store = context.watch<MappingConfigStore>();
    final cfg = store.config;
    final canEdit = !store.loading && cfg != null;
    if (store.loading && cfg == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final err = store.lastError;
    final uploadErr =
        _uploadCtrl.text.trim().isEmpty ||
            _parsePositiveInt(_uploadCtrl.text) != null
        ? null
        : 'Must be > 0';

    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          InlineStatusBanner(
            loading: store.loading && cfg != null,
            loadingText: 'Refreshing config…',
            errorText: err,
            onRetry: () => store.load(),
          ),
          Row(
            children: [
              SizedBox(
                width: 220,
                child: SwitchListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Enabled'),
                  value: cfg?.enabled ?? true,
                  onChanged: !canEdit
                      ? null
                      : (v) async {
                          try {
                            final c = MappingConfig(
                              enabled: v,
                              uploadMaxMB:
                                  _parsePositiveInt(_uploadCtrl.text) ??
                                  _lastValidUploadMaxMB,
                            );
                            await store.save(c);
                          } catch (e) {
                            sl<NotificationsService>().error(
                              'Save failed',
                              e.toString(),
                            );
                          }
                        },
                ),
              ),
              const Spacer(),
              IconButton(
                tooltip: 'Refresh',
                onPressed: () async {
                  try {
                    await store.load();
                    final newCfg = store.config;
                    _uploadCtrl.text = (newCfg?.uploadMaxMB ?? 20).toString();
                  } catch (e) {
                    sl<NotificationsService>().error('Failed', e.toString());
                  }
                },
                icon: const Icon(Icons.refresh),
              ),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            width: 220,
            child: TextField(
              controller: _uploadCtrl,
              focusNode: _uploadFocus,
              keyboardType: TextInputType.number,
              textInputAction: TextInputAction.done,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              onSubmitted: !canEdit
                  ? null
                  : (_) async {
                      try {
                        final val =
                            _parsePositiveInt(_uploadCtrl.text) ??
                            _lastValidUploadMaxMB;
                        final c = MappingConfig(
                          enabled: cfg.enabled,
                          uploadMaxMB: val,
                        );
                        await store.save(c);
                      } catch (e) {
                        sl<NotificationsService>().error(
                          'Save failed',
                          e.toString(),
                        );
                      }
                    },
              decoration: InputDecoration(
                labelText: 'Upload max (MB)',
                errorText: uploadErr,
              ),
            ),
          ),
          const SizedBox(height: 12),
          ElevatedButton.icon(
            onPressed: !canEdit
                ? null
                : () async {
                    try {
                      final val =
                          _parsePositiveInt(_uploadCtrl.text) ??
                          _lastValidUploadMaxMB;
                      final c = MappingConfig(
                        enabled: cfg.enabled,
                        uploadMaxMB: val,
                      );
                      await store.save(c);
                    } catch (e) {
                      sl<NotificationsService>().error(
                        'Save failed',
                        e.toString(),
                      );
                    }
                  },
            icon: const Icon(Icons.save),
            label: const Text('Save'),
          ),
        ],
      ),
    );
  }
}
