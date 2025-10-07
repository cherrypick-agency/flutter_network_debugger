import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../../../../../core/di/di.dart';
import '../../../../application/stores/home_ui_store.dart';
import '../../../../application/stores/sessions_store.dart';

import '../../../widgets/sessions_column.dart';
import '../../../../../../features/filters/application/stores/sessions_filters_store.dart';
import '../../../utils/sessions_filtering.dart';

class SessionsPane extends StatelessWidget {
  const SessionsPane({
    super.key,
    required this.searchCtrl,
    required this.sessionsCtrl,
    required this.onSelectSession,
  });
  final TextEditingController searchCtrl;
  final ScrollController sessionsCtrl;
  final void Function(String id) onSelectSession;

  @override
  Widget build(BuildContext context) {
    return Observer(
      builder: (_) {
        final ui = sl<HomeUiStore>();
        final sessions = filterVisibleSessions(
          store: context.read<SessionsStore>(),
          filters: context.read<SessionsFiltersStore>(),
          selectedRange:
              ui.selectedRange.value == null
                  ? null
                  : DateTimeRangeWrapper(
                    start: ui.selectedRange.value!.start,
                    end: ui.selectedRange.value!.end,
                  ),
          selectedDomains: ui.selectedDomains,
          httpMeta: ui.httpMeta,
          since: ui.since.value,
          ignoredIds: const {},
        );
        return SessionsColumn(
          showSearch: ui.showSearch.value,
          onShowSearchChanged: (v) {
            ui.setShowSearch(v);
            if (!v) searchCtrl.clear();
          },
          sessionSearchCtrl: searchCtrl,
          onSearchPrefsChanged: () {},
          onSearchSubmit: () {},
          selectedDomains: ui.selectedDomains,
          onToggleDomain: (key, add) {
            if (add)
              ui.addDomain(key);
            else
              ui.removeDomain(key);
          },
          sessions: sessions,
          sessionsCtrl: sessionsCtrl,
          groupBy: context.watch<SessionsFiltersStore>().groupBy,
          selectedSessionId: ui.selectedSessionId.value,
          onSelectSession: onSelectSession,
          httpMeta: ui.httpMeta,
        );
      },
    );
  }
}
