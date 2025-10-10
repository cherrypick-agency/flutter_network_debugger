import 'package:flutter/material.dart';

class TimelineSettingsButton extends StatelessWidget {
  const TimelineSettingsButton({super.key, required this.getFit, required this.setFit, required this.getCrop, required this.setCrop});
  final bool Function() getFit;
  final void Function(bool) setFit;
  final bool Function() getCrop;
  final void Function(bool) setCrop;
  @override
  Widget build(BuildContext context) {
    return IconButton(
      tooltip: 'Timeline settings',
      icon: const Icon(Icons.settings, size: 16),
      iconSize: 16,
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
      onPressed: () async {
        final res = await showDialog<({bool fit, bool crop})>(
          context: context,
          builder: (ctx) {
            bool localFit = getFit();
            bool localCrop = getCrop();
            return AlertDialog(
              title: const Text('Timeline Settings'),
              content: StatefulBuilder(
                builder: (ctx, setState) {
                  return Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      CheckboxListTile(
                        value: localFit,
                        onChanged: (v){ setState(()=> localFit = v ?? localFit); },
                        title: const Text('Fit all'),
                        controlAffinity: ListTileControlAffinity.leading,
                        contentPadding: EdgeInsets.zero,
                      ),
                      CheckboxListTile(
                        value: localCrop,
                        onChanged: (v){ setState(()=> localCrop = v ?? localCrop); },
                        title: const Text('Crop to recent window'),
                        subtitle: const Text('Ограничить таймлайн окном последних N минут'),
                        controlAffinity: ListTileControlAffinity.leading,
                        contentPadding: EdgeInsets.zero,
                      ),
                    ],
                  );
                },
              ),
              actions: [
                TextButton(onPressed: ()=> Navigator.of(ctx).pop(null), child: const Text('Cancel')),
                ElevatedButton(onPressed: ()=> Navigator.of(ctx).pop((fit: localFit, crop: localCrop)), child: const Text('Apply')),
              ],
            );
          },
        );
        if (res != null) {
          setFit(res.fit);
          setCrop(res.crop);
        }
      },
    );
  }
}


