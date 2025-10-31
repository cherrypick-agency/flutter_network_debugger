import 'package:flutter/material.dart';

import '../../application/compose_store.dart';
import '../../domain/models.dart';
import '../../../../../../core/di/di.dart';
import '../../data/compose_repository.dart';

class ComposeLibraryTree extends StatelessWidget {
  final ComposeStore store;
  final void Function(ComposeTemplateDTO) onSelect;
  final String? searchQuery;
  const ComposeLibraryTree({
    super.key,
    required this.store,
    required this.onSelect,
    this.searchQuery,
  });

  @override
  Widget build(BuildContext context) {
    if (store.loading) {
      return const Center(child: CircularProgressIndicator());
    }
    return ListView(
      children: [
        for (final c in store.collections)
          _CollectionNode(
            collection: c,
            store: store,
            onSelect: onSelect,
            searchQuery: (searchQuery ?? '').trim(),
          ),
      ],
    );
  }
}

class _CollectionNode extends StatelessWidget {
  final ComposeCollectionModel collection;
  final ComposeStore store;
  final void Function(ComposeTemplateDTO) onSelect;
  final String searchQuery;
  const _CollectionNode({
    required this.collection,
    required this.store,
    required this.onSelect,
    required this.searchQuery,
  });

  @override
  Widget build(BuildContext context) {
    return ExpansionTile(
      title: Text(collection.name),
      children: [
        _FolderNode(
          collectionId: collection.id,
          folder: collection.root,
          store: store,
          onSelect: onSelect,
          searchQuery: searchQuery,
        ),
      ],
    );
  }
}

class _FolderNode extends StatelessWidget {
  final String collectionId;
  final ComposeFolderModel folder;
  final ComposeStore store;
  final void Function(ComposeTemplateDTO) onSelect;
  final String searchQuery;
  const _FolderNode({
    required this.collectionId,
    required this.folder,
    required this.store,
    required this.onSelect,
    required this.searchQuery,
  });

  @override
  Widget build(BuildContext context) {
    final q = searchQuery.toLowerCase();
    final reqs =
        folder.requests.where((rid) {
          final name = store.requestsById[rid]?.name ?? rid;
          if (q.isEmpty) return true;
          return name.toLowerCase().contains(q);
        }).toList();
    return ExpansionTile(
      title: Text(folder.name),
      children: [
        // Drop area at top -> move to top of this folder
        _DropSpacer(
          onAccept: (rid) async {
            final repo = sl<ComposeRepository>();
            await repo.moveRequest(
              id: rid,
              collectionId: collectionId,
              folderId: folder.id,
              index: 0,
            );
            await store.loadLibrary();
          },
        ),
        for (var i = 0; i < reqs.length; i++) ...[
          _DropSpacer(
            onAccept: (rid) async {
              final repo = sl<ComposeRepository>();
              await repo.moveRequest(
                id: rid,
                collectionId: collectionId,
                folderId: folder.id,
                index: i,
              );
              await store.loadLibrary();
            },
          ),
          _DraggableRequestTile(
            rid: reqs[i],
            title: store.requestsById[reqs[i]]?.name ?? reqs[i],
            onTap: () {
              final t = store.requestsById[reqs[i]];
              if (t != null) onSelect(t);
            },
            menu: _RequestMenu(rid: reqs[i], store: store),
          ),
        ],
        // Bottom drop area -> append
        _DropSpacer(
          onAccept: (rid) async {
            final repo = sl<ComposeRepository>();
            await repo.moveRequest(
              id: rid,
              collectionId: collectionId,
              folderId: folder.id,
              index: -1,
            );
            await store.loadLibrary();
          },
        ),
        for (final f in folder.folders)
          Padding(
            padding: const EdgeInsets.only(left: 8),
            child: _FolderNode(
              collectionId: collectionId,
              folder: f,
              store: store,
              onSelect: onSelect,
              searchQuery: searchQuery,
            ),
          ),
      ],
    );
  }
}

class _DraggableRequestTile extends StatelessWidget {
  const _DraggableRequestTile({
    required this.rid,
    required this.title,
    required this.onTap,
    required this.menu,
  });
  final String rid;
  final String title;
  final VoidCallback onTap;
  final Widget menu;
  @override
  Widget build(BuildContext context) {
    return LongPressDraggable<String>(
      data: rid,
      feedback: Material(
        color: Colors.transparent,
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 280),
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              border: Border.all(
                color: Theme.of(context).colorScheme.outlineVariant,
              ),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Padding(
              padding: const EdgeInsets.all(6),
              child: Text(title, style: Theme.of(context).textTheme.bodySmall),
            ),
          ),
        ),
      ),
      child: ListTile(
        dense: true,
        leading: const Icon(Icons.http),
        title: Text(title),
        trailing: menu,
        onTap: onTap,
      ),
    );
  }
}

class _DropSpacer extends StatelessWidget {
  const _DropSpacer({required this.onAccept});
  final Future<void> Function(String rid) onAccept;
  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return DragTarget<String>(
      builder: (context, candidate, rejected) {
        final hovering = candidate.isNotEmpty;
        return Container(
          height: 8,
          margin: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            border: Border.all(
              color: hovering ? cs.primary : cs.outlineVariant,
            ),
            borderRadius: BorderRadius.circular(4),
          ),
        );
      },
      onWillAccept: (data) => data != null && data.isNotEmpty,
      onAccept: (data) async {
        await onAccept(data);
      },
    );
  }
}

class _RequestMenu extends StatelessWidget {
  const _RequestMenu({required this.rid, required this.store});
  final String rid;
  final ComposeStore store;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<String>(
      onSelected: (v) async {
        if (v == 'rename') {
          final tpl = store.requestsById[rid];
          if (tpl == null) return;
          final ctrl = TextEditingController(text: tpl.name);
          final name = await showDialog<String>(
            context: context,
            builder:
                (ctx) => AlertDialog(
                  title: const Text('Rename'),
                  content: TextField(controller: ctrl, autofocus: true),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.of(ctx).pop(null),
                      child: const Text('Cancel'),
                    ),
                    FilledButton(
                      onPressed: () => Navigator.of(ctx).pop(ctrl.text.trim()),
                      child: const Text('Save'),
                    ),
                  ],
                ),
          );
          if (name != null && name.isNotEmpty) {
            final repo = sl<ComposeRepository>();
            await repo.upsertRequest(
              ComposeTemplateDTO(
                id: tpl.id,
                name: name,
                method: tpl.method,
                url: tpl.url,
                headers: tpl.headers,
                query: tpl.query,
                body: tpl.body,
                auth: tpl.auth,
              ),
            );
            await store.loadLibrary();
          }
        }
        if (v == 'duplicate') {
          final tpl = store.requestsById[rid];
          if (tpl == null) return;
          final repo = sl<ComposeRepository>();
          final copy = ComposeTemplateDTO(
            id: 'tpl-${DateTime.now().microsecondsSinceEpoch}',
            name: '${tpl.name} (copy)',
            method: tpl.method,
            url: tpl.url,
            headers: tpl.headers,
            query: tpl.query,
            body: tpl.body,
            auth: tpl.auth,
          );
          await repo.upsertRequest(copy);
          await store.loadLibrary();
        }
        if (v == 'delete') {
          await store.deleteTemplate(rid);
        }
        if (v == 'move') {
          final tpl = store.requestsById[rid];
          if (tpl == null) return;
          final collections = store.collections;
          String? targetCol =
              collections.isNotEmpty ? collections.first.id : null;
          String? targetFolder;
          final res = await showDialog<({String col, String folder})>(
            context: context,
            builder: (ctx) {
              final foldersByCol = <String, List<ComposeFolderModel>>{
                for (final c in collections) c.id: _flattenFolders(c.root),
              };
              targetFolder = foldersByCol[targetCol]?.first.id;
              return StatefulBuilder(
                builder: (ctx, setS) {
                  final colFolders =
                      foldersByCol[targetCol] ?? const <ComposeFolderModel>[];
                  targetFolder ??=
                      colFolders.isNotEmpty ? colFolders.first.id : null;
                  return AlertDialog(
                    title: const Text('Move to'),
                    content: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        DropdownButtonFormField<String>(
                          value: targetCol,
                          items: [
                            for (final c in collections)
                              DropdownMenuItem(
                                value: c.id,
                                child: Text(c.name),
                              ),
                          ],
                          onChanged: (v) {
                            setS(() {
                              targetCol = v;
                              targetFolder = foldersByCol[targetCol]?.first.id;
                            });
                          },
                        ),
                        const SizedBox(height: 8),
                        DropdownButtonFormField<String>(
                          value: targetFolder,
                          items: [
                            for (final f in colFolders)
                              DropdownMenuItem(
                                value: f.id,
                                child: Text(f.name),
                              ),
                          ],
                          onChanged: (v) {
                            setS(() {
                              targetFolder = v;
                            });
                          },
                        ),
                      ],
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(ctx).pop(null),
                        child: const Text('Cancel'),
                      ),
                      FilledButton(
                        onPressed:
                            () => Navigator.of(
                              ctx,
                            ).pop((col: targetCol!, folder: targetFolder!)),
                        child: const Text('Move'),
                      ),
                    ],
                  );
                },
              );
            },
          );
          if (res != null) {
            final repo = sl<ComposeRepository>();
            await repo.moveRequest(
              id: rid,
              collectionId: res.col,
              folderId: res.folder,
              index: -1,
            );
            await store.loadLibrary();
          }
        }
      },
      itemBuilder:
          (ctx) => const [
            PopupMenuItem(value: 'rename', child: Text('Rename')),
            PopupMenuItem(value: 'duplicate', child: Text('Duplicate')),
            PopupMenuItem(value: 'move', child: Text('Move to...')),
            PopupMenuItem(value: 'delete', child: Text('Delete')),
          ],
    );
  }
}

List<ComposeFolderModel> _flattenFolders(ComposeFolderModel root) {
  final out = <ComposeFolderModel>[];
  void walk(ComposeFolderModel f) {
    out.add(f);
    for (final c in f.folders) {
      walk(c);
    }
  }

  walk(root);
  return out;
}
