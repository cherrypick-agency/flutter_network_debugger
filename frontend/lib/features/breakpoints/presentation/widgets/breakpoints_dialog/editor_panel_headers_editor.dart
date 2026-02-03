import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../../compose/presentation/widgets/key_value_editor.dart';
import '../../../../compose/presentation/widgets/kv.dart';

class EditorPanelHeadersEditor extends StatelessWidget {
  const EditorPanelHeadersEditor({
    super.key,
    required this.headers,
    required this.onChanged,
  });

  final List<KvPair> headers;
  final ValueChanged<List<KvPair>> onChanged;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: context.appColors.border),
        borderRadius: BorderRadius.circular(8),
      ),
      child: KeyValueEditor(
        items: headers,
        onChanged: onChanged,
        labelKey: 'Header',
      ),
    );
  }
}
