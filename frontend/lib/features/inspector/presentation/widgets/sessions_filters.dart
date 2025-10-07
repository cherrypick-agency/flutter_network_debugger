import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../../filters/application/stores/sessions_filters_store.dart';
import '../../application/stores/sessions_store.dart';

class SessionsFilters extends StatelessWidget {
  const SessionsFilters({
    super.key,
    required this.targetCtrl,
    required this.httpMethod,
    required this.httpStatus,
    required this.httpMime,
    required this.httpMinDurationMs,
    required this.groupBy,
    required this.headerKeyCtrl,
    required this.headerValCtrl,
    required this.onTargetChanged,
    required this.onTargetSubmitted,
    required this.onHttpMethodChanged,
    required this.onHttpStatusChanged,
    required this.onHttpMimeChanged,
    required this.onHttpMinDurationChanged,
    required this.onGroupByChanged,
    required this.onHeaderKeyChanged,
    required this.onHeaderValChanged,
  });

  final TextEditingController targetCtrl;
  final String httpMethod;
  final String httpStatus;
  final String httpMime;
  final int httpMinDurationMs;
  final String groupBy;
  final TextEditingController headerKeyCtrl;
  final TextEditingController headerValCtrl;

  final VoidCallback onTargetChanged;
  final VoidCallback onTargetSubmitted;
  final ValueChanged<String> onHttpMethodChanged;
  final ValueChanged<String> onHttpStatusChanged;
  final ValueChanged<String> onHttpMimeChanged;
  final ValueChanged<int> onHttpMinDurationChanged;
  final ValueChanged<String> onGroupByChanged;
  final VoidCallback onHeaderKeyChanged;
  final VoidCallback onHeaderValChanged;

  @override
  Widget build(BuildContext context) {
    final filterStore = context.watch<SessionsFiltersStore>();
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        SizedBox(
          width: 300,
          child: TextField(
            style: const TextStyle(fontSize: 12),
            controller: targetCtrl,
            decoration: const InputDecoration(labelText: 'Filter by target'),
            onChanged: (_) {
              filterStore.setTarget(targetCtrl.text);
              onTargetChanged();
            },
            onSubmitted: (_) => onTargetSubmitted(),
          ),
        ),
        DropdownButton<String>(
          value: httpMethod,
          isDense: true,
          iconSize: 18,
          style: TextStyle(
            fontSize: 12,
            color: Theme.of(context).colorScheme.onSurface,
          ),
          items: const [
            DropdownMenuItem(
              value: 'any',
              child: Text('Any method', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'GET',
              child: Text('GET', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'POST',
              child: Text('POST', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'PUT',
              child: Text('PUT', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'DELETE',
              child: Text('DELETE', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'PATCH',
              child: Text('PATCH', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'OPTIONS',
              child: Text('OPTIONS', style: TextStyle(fontSize: 12)),
            ),
          ],
          onChanged: (v) {
            onHttpMethodChanged(v ?? 'any');
            filterStore.setHttpMethod(v ?? 'any');
          },
        ),
        DropdownButton<String>(
          value: httpStatus,
          isDense: true,
          iconSize: 18,
          style: TextStyle(
            fontSize: 12,
            color: Theme.of(context).colorScheme.onSurface,
          ),
          items: const [
            DropdownMenuItem(
              value: 'any',
              child: Text('Any status', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: '2xx',
              child: Text('2xx', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: '3xx',
              child: Text('3xx', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: '4xx',
              child: Text('4xx', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: '5xx',
              child: Text('5xx', style: TextStyle(fontSize: 12)),
            ),
          ],
          onChanged: (v) {
            onHttpStatusChanged(v ?? 'any');
            filterStore.setHttpStatus(v ?? 'any');
          },
        ),
        SizedBox(
          width: 200,
          child: TextField(
            style: const TextStyle(fontSize: 12),
            decoration: const InputDecoration(labelText: 'MIME contains'),
            onChanged: (v) {
              onHttpMimeChanged(v);
              filterStore.setHttpMime(v);
            },
          ),
        ),
        SizedBox(
          width: 120,
          child: TextField(
            style: const TextStyle(fontSize: 12),
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(labelText: 'Min ms'),
            onChanged: (v) {
              final parsed = int.tryParse(v) ?? 0;
              onHttpMinDurationChanged(parsed);
              filterStore.setHttpMinDurationMs(parsed);
            },
          ),
        ),
        DropdownButton<String>(
          value: groupBy,
          isDense: true,
          iconSize: 18,
          style: TextStyle(
            fontSize: 12,
            color: Theme.of(context).colorScheme.onSurface,
          ),
          items: const [
            DropdownMenuItem(
              value: 'none',
              child: Text('No grouping', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'domain',
              child: Text('Group by domain', style: TextStyle(fontSize: 12)),
            ),
            DropdownMenuItem(
              value: 'route',
              child: Text('Group by route', style: TextStyle(fontSize: 12)),
            ),
          ],
          onChanged: (v) {
            final val = v ?? 'none';
            onGroupByChanged(val);
            filterStore.setGroupBy(val);
          },
        ),
        SizedBox(
          width: 160,
          child: TextField(
            style: const TextStyle(fontSize: 12),
            controller: headerKeyCtrl,
            decoration: const InputDecoration(labelText: 'Header key'),
            onChanged: (_) {
              filterStore.setHeaderKey(headerKeyCtrl.text);
              onHeaderKeyChanged();
            },
          ),
        ),
        SizedBox(
          width: 180,
          child: TextField(
            style: const TextStyle(fontSize: 12),
            controller: headerValCtrl,
            decoration: const InputDecoration(labelText: 'Header value'),
            onChanged: (_) {
              filterStore.setHeaderVal(headerValCtrl.text);
              onHeaderValChanged();
            },
          ),
        ),
        Observer(
          builder: (_) {
            final loading = context.watch<SessionsStore>().loading;
            return loading
                ? const SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
                : const SizedBox.shrink();
          },
        ),
      ],
    );
  }
}
