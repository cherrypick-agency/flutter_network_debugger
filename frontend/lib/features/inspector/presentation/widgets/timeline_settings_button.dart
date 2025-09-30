import 'package:flutter/material.dart';

class TimelineSettingsButton extends StatelessWidget {
  const TimelineSettingsButton({super.key, required this.getFit, required this.setFit});
  final bool Function() getFit;
  final void Function(bool) setFit;
  @override
  Widget build(BuildContext context) {
    return IconButton(
      tooltip: 'Timeline settings',
      icon: const Icon(Icons.settings),
      onPressed: () async {
        final fit = await showDialog<bool>(
          context: context,
          builder: (ctx) {
            bool localFit = getFit();
            return AlertDialog(
              title: const Text('Timeline Settings'),
              content: StatefulBuilder(
                builder: (ctx, setState) {
                  return CheckboxListTile(
                    value: localFit,
                    onChanged: (v){ setState(()=> localFit = v ?? localFit); },
                    title: const Text('Fit all'),
                    controlAffinity: ListTileControlAffinity.leading,
                    contentPadding: EdgeInsets.zero,
                  );
                },
              ),
              actions: [
                TextButton(onPressed: ()=> Navigator.of(ctx).pop(null), child: const Text('Cancel')),
                ElevatedButton(onPressed: ()=> Navigator.of(ctx).pop(localFit), child: const Text('Apply')),
              ],
            );
          },
        );
        if (fit != null) setFit(fit);
      },
    );
  }
}


