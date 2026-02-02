import 'package:flutter/material.dart';
import '../../../../../theme/context_ext.dart';

class MappingRuleRemoteSection extends StatelessWidget {
  const MappingRuleRemoteSection({
    super.key,
    required this.targetURLTemplateController,
    required this.saving,
    required this.preserveHost,
    required this.onPreserveHostChanged,
    this.targetURLTemplateErrorText,
    this.onTargetURLTemplateChanged,
  });

  final TextEditingController targetURLTemplateController;
  final bool saving;
  final bool preserveHost;
  final ValueChanged<bool> onPreserveHostChanged;
  final String? targetURLTemplateErrorText;
  final ValueChanged<String>? onTargetURLTemplateChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(
          controller: targetURLTemplateController,
          enabled: !saving,
          decoration: InputDecoration(
            labelText: 'Target URL template',
            hintText: 'https://example.com{path}',
            errorText: targetURLTemplateErrorText,
          ),
          onChanged: onTargetURLTemplateChanged,
        ),
        const SizedBox(height: 6),
        Text(
          'Токены: {path} {host} {hostname} {url} {scheme} {port} {portWithColon}',
          style: context.appText.body.copyWith(
            color: context.appColors.textSecondary,
          ),
        ),
        const SizedBox(height: 8),
        SwitchListTile(
          dense: true,
          contentPadding: EdgeInsets.zero,
          title: const Text('Preserve Host header'),
          value: preserveHost,
          onChanged: saving ? null : onPreserveHostChanged,
        ),
      ],
    );
  }
}
