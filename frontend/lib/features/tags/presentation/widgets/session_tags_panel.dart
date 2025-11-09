import 'package:flutter/material.dart';

import 'annotations_editor.dart';
import 'tags_editor.dart';

class SessionTagsPanel extends StatelessWidget {
  const SessionTagsPanel({required this.sessionId, super.key});

  final String sessionId;

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      child: Column(
        children: [
          const TabBar(
            tabs: [
              Tab(text: 'Tags', icon: Icon(Icons.label, size: 16)),
              Tab(text: 'Annotations', icon: Icon(Icons.note, size: 16)),
            ],
          ),
          Expanded(
            child: TabBarView(
              children: [
                SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: TagsEditor(sessionId: sessionId),
                ),
                SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: AnnotationsEditor(sessionId: sessionId),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
