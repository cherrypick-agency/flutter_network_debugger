import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class MappingActivityPanel extends StatelessWidget {
  const MappingActivityPanel({super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.history, size: 40, color: context.appColors.textSecondary),
          const SizedBox(height: 8),
          Text('No activity yet', style: context.appText.subtitle),
          const SizedBox(height: 4),
          Text(
            'Recent mapping hits and changes will appear here.',
            style: context.appText.body,
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
