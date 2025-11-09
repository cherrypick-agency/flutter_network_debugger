import 'package:flutter/material.dart';
import 'package:animated_tree_view/animated_tree_view.dart';

import '../../application/compose_store.dart';
import '../../domain/models.dart';
import '../../../../../../core/di/di.dart';
import '../../data/compose_repository.dart';

enum _NodeType { root, collection, folder, request }

class _NodeData {
  final _NodeType type;
  final String id; // collectionId | folderId | requestId | 'root'
  final String? parentCollectionId;
  final String title;
  final String? folderId;
  const _NodeData({
    required this.type,
    required this.id,
    required this.title,
    this.parentCollectionId,
    this.folderId,
  });
}

class LibraryAnimatedTree extends StatefulWidget {
  const LibraryAnimatedTree({
    super.key,
    required this.store,
    required this.onSelect,
    this.searchQuery,
  });
  final ComposeStore store;
  final void Function(ComposeTemplateDTO) onSelect;
  final String? searchQuery;

  @override
  State<LibraryAnimatedTree> createState() => _LibraryAnimatedTreeState();
}

class _LibraryAnimatedTreeState extends State<LibraryAnimatedTree> {
  late final TreeNode<_NodeData> _root;
  void _onStoreChanged() {
    if (mounted) {
      setState(() {
        _rebuildTree();
      });
    }
  }

  @override
  void initState() {
    super.initState();
    _root = TreeNode<_NodeData>(
      key: 'root',
      data: const _NodeData(type: _NodeType.root, id: 'root', title: 'root'),
    );
    _rebuildTree();
    widget.store.addListener(_onStoreChanged);
  }

  @override
  void didUpdateWidget(covariant LibraryAnimatedTree oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.store != widget.store) {
      oldWidget.store.removeListener(_onStoreChanged);
      widget.store.addListener(_onStoreChanged);
      if (mounted) {
        setState(() {
          _rebuildTree();
        });
      }
    }
    if (oldWidget.store == widget.store &&
        (oldWidget.searchQuery ?? '') != (widget.searchQuery ?? '')) {
      if (mounted) {
        setState(() {
          _rebuildTree();
        });
      }
    }
  }

  @override
  void dispose() {
    widget.store.removeListener(_onStoreChanged);
    super.dispose();
  }

  void _rebuildTree() {
    _root.clear();
    final store = widget.store;
    final q = (widget.searchQuery ?? '').trim().toLowerCase();

    for (final col in store.collections) {
      final colNode = TreeNode<_NodeData>(
        key: 'col:${col.id}',
        data: _NodeData(
          type: _NodeType.collection,
          id: col.id,
          title: col.name,
          parentCollectionId: col.id,
        ),
      );
      _root.add(colNode);
      // Don't show root as a folder node - add its contents directly to collection
      _buildRootContents(colNode, col.root, store, q, col.id);
    }
  }

  bool _matches(String name, String q) {
    if (q.isEmpty) return true;
    return name.toLowerCase().contains(q);
  }

  void _buildRootContents(
    TreeNode<_NodeData> colNode,
    ComposeFolderModel root,
    ComposeStore store,
    String q,
    String collectionId,
  ) {
    // Add requests from root directly to collection node
    for (final rid in root.requests) {
      final tpl = store.requestsById[rid];
      if (tpl == null) continue;
      final title = tpl.name.isNotEmpty ? tpl.name : rid;
      if (!_matches(title, q)) continue;

      final requestNode = TreeNode<_NodeData>(
        key: 'req:$rid',
        data: _NodeData(
          type: _NodeType.request,
          id: rid,
          title: title,
          parentCollectionId: collectionId,
          folderId: root.id,
        ),
      );
      colNode.add(requestNode);
    }

    // Add subfolders from root directly to collection node
    for (final subfolder in root.folders) {
      _buildFolder(colNode, subfolder, store, q, collectionId);
    }
  }

  bool _buildFolder(
    TreeNode<_NodeData> parent,
    ComposeFolderModel folder,
    ComposeStore store,
    String q,
    String collectionId,
  ) {
    // Build children first to decide visibility of this folder on search
    final folderNode = TreeNode<_NodeData>(
      key: 'fld:${folder.id}',
      data: _NodeData(
        type: _NodeType.folder,
        id: folder.id,
        title: folder.name,
        parentCollectionId: collectionId,
        folderId: folder.id,
      ),
    );
    bool anyMatch = _matches(folder.name, q);

    // Requests
    for (final rid in folder.requests) {
      final tpl = store.requestsById[rid];
      if (tpl == null) continue;
      final title = tpl.name.isNotEmpty ? tpl.name : rid;
      final requestNode = TreeNode<_NodeData>(
        key: 'req:$rid',
        data: _NodeData(
          type: _NodeType.request,
          id: rid,
          title: title,
          parentCollectionId: collectionId,
          folderId: folder.id,
        ),
      );
      final match = _matches(title, q);
      if (match) {
        anyMatch = true;
        folderNode.add(requestNode);
      }
    }

    // Subfolders
    for (final sub in folder.folders) {
      final subHas = _buildFolder(folderNode, sub, store, q, collectionId);
      anyMatch = anyMatch || subHas;
    }

    if (anyMatch) {
      parent.add(folderNode);
      return true;
    }
    return false;
  }

  @override
  Widget build(BuildContext context) {
    if (widget.store.loading) {
      return const Center(child: CircularProgressIndicator(strokeWidth: 2));
    }
    if (widget.store.collections.isEmpty) {
      return _buildEmptyState(context);
    }
    return TreeView.simpleTyped<_NodeData, TreeNode<_NodeData>>(
      tree: _root,
      padding: EdgeInsets.zero,
      showRootNode: false,
      expansionBehavior: ExpansionBehavior.none,
      builder: (context, node) {
        final d = node.data!;
        switch (d.type) {
          case _NodeType.root:
            return const SizedBox.shrink();
          case _NodeType.collection:
            return _collectionTile(context, node);
          case _NodeType.folder:
            return _folderTile(context, node);
          case _NodeType.request:
            return _requestTile(context, node);
        }
      },
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.folder_off,
              size: 64,
              color: Theme.of(context).colorScheme.outline,
            ),
            const SizedBox(height: 16),
            Text(
              'Библиотека пуста',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              'Создайте первую коллекцию',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 24),
            FilledButton.icon(
              onPressed: () async {
                final name = await _promptText(
                  context,
                  title: 'Название коллекции',
                );
                if (name == null || name.isEmpty) return;
                final repo = sl<ComposeRepository>();
                final collectionId =
                    'col-${DateTime.now().microsecondsSinceEpoch}';
                final rootFolderId =
                    'fld-${DateTime.now().microsecondsSinceEpoch}';
                await repo.upsertCollection({
                  'id': collectionId,
                  'name': name,
                  'root': {
                    'id': rootFolderId,
                    'name': '',
                    'requests': <String>[],
                    'folders': <Map<String, dynamic>>[],
                  },
                });
                await widget.store.loadLibrary();
              },
              icon: const Icon(Icons.add),
              label: const Text('Создать коллекцию'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _collectionTile(BuildContext context, TreeNode<_NodeData> node) {
    final vd = const VisualDensity(horizontal: -4, vertical: -4);
    return Row(
      children: [
        const Icon(Icons.folder_copy, size: 16),
        const SizedBox(width: 6),
        Expanded(child: Text(node.data!.title)),
        IconButton(
          tooltip: 'New folder',
          icon: const Icon(Icons.add, size: 18),
          visualDensity: vd,
          onPressed: () async {
            final name = await _promptText(context, title: 'New folder name');
            if (name == null || name.isEmpty) return;
            final repo = sl<ComposeRepository>();
            final store = widget.store;
            final col = store.collections.firstWhere(
              (c) => c.id == node.data!.id,
              orElse: () => store.collections.first,
            );
            final root = col.root.toJson();
            _folderAdd(root, col.root.id, {
              'id': 'fld-${DateTime.now().microsecondsSinceEpoch}',
              'name': name,
              'requests': <String>[],
              'folders': <Map<String, dynamic>>[],
            });
            await repo.upsertCollection({
              'id': col.id,
              'name': col.name,
              'root': root,
            });
            await store.loadLibrary();
          },
        ),
        PopupMenuButton<String>(
          tooltip: 'Collection menu',
          itemBuilder: (ctx) => const [
            PopupMenuItem(value: 'rename', child: Text('Rename')),
            PopupMenuItem(value: 'delete', child: Text('Delete')),
          ],
          onSelected: (v) async {
            final repo = sl<ComposeRepository>();
            final store = widget.store;
            final col = store.collections.firstWhere(
              (c) => c.id == node.data!.id,
              orElse: () => store.collections.first,
            );
            if (v == 'rename') {
              final name = await _promptText(
                context,
                title: 'Rename collection',
                initial: col.name,
              );
              if (name == null || name.isEmpty) return;
              await repo.upsertCollection({
                'id': col.id,
                'name': name,
                'root': col.root.toJson(),
              });
              await store.loadLibrary();
            } else if (v == 'delete') {
              final ok = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: const Text('Confirm'),
                  content: const Text('Delete collection?'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.of(ctx).pop(false),
                      child: const Text('Cancel'),
                    ),
                    FilledButton(
                      onPressed: () => Navigator.of(ctx).pop(true),
                      child: const Text('OK'),
                    ),
                  ],
                ),
              );
              if (ok == true) {
                await repo.deleteCollection(col.id);
                await store.loadLibrary();
              }
            }
          },
        ),
      ],
    );
  }

  Widget _folderTile(BuildContext context, TreeNode<_NodeData> node) {
    final vd = const VisualDensity(horizontal: -4, vertical: -4);
    return Row(
      children: [
        const Icon(Icons.folder, size: 16),
        const SizedBox(width: 6),
        Expanded(child: Text(node.data!.title)),
        IconButton(
          tooltip: 'New subfolder',
          icon: const Icon(Icons.add, size: 18),
          visualDensity: vd,
          onPressed: () async {
            final name = await _promptText(context, title: 'New folder name');
            if (name == null || name.isEmpty) return;
            final repo = sl<ComposeRepository>();
            final store = widget.store;
            final col = store.collections.firstWhere(
              (c) => c.id == node.data!.parentCollectionId,
              orElse: () => store.collections.first,
            );
            final root = col.root.toJson();
            _folderAdd(root, node.data!.id, {
              'id': 'fld-${DateTime.now().microsecondsSinceEpoch}',
              'name': name,
              'requests': <String>[],
              'folders': <Map<String, dynamic>>[],
            });
            await repo.upsertCollection({
              'id': col.id,
              'name': col.name,
              'root': root,
            });
            await store.loadLibrary();
          },
        ),
        PopupMenuButton<String>(
          tooltip: 'Folder menu',
          itemBuilder: (ctx) => const [
            PopupMenuItem(value: 'rename', child: Text('Rename')),
            PopupMenuItem(value: 'delete', child: Text('Delete')),
          ],
          onSelected: (v) async {
            final repo = sl<ComposeRepository>();
            final store = widget.store;
            final col = store.collections.firstWhere(
              (c) => c.id == node.data!.parentCollectionId,
              orElse: () => store.collections.first,
            );
            final root = col.root.toJson();
            if (v == 'rename') {
              final name = await _promptText(
                context,
                title: 'Rename folder',
                initial: node.data!.title,
              );
              if (name == null || name.isEmpty) return;
              _folderRename(root, node.data!.id, name);
              await repo.upsertCollection({
                'id': col.id,
                'name': col.name,
                'root': root,
              });
              await store.loadLibrary();
            } else if (v == 'delete') {
              final ok = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: const Text('Confirm'),
                  content: const Text('Delete folder and its contents?'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.of(ctx).pop(false),
                      child: const Text('Cancel'),
                    ),
                    FilledButton(
                      onPressed: () => Navigator.of(ctx).pop(true),
                      child: const Text('OK'),
                    ),
                  ],
                ),
              );
              if (ok == true) {
                _folderDelete(root, node.data!.id);
                await repo.upsertCollection({
                  'id': col.id,
                  'name': col.name,
                  'root': root,
                });
                await store.loadLibrary();
              }
            }
          },
        ),
      ],
    );
  }

  Widget _requestTile(BuildContext context, TreeNode<_NodeData> node) {
    final store = widget.store;
    return Row(
      children: [
        const Icon(Icons.http, size: 16),
        const SizedBox(width: 6),
        Expanded(
          child: InkWell(
            onTap: () {
              final tpl = store.requestsById[node.data!.id];
              if (tpl != null) widget.onSelect(tpl);
            },
            child: Text(node.data!.title),
          ),
        ),
        PopupMenuButton<String>(
          itemBuilder: (ctx) => const [
            PopupMenuItem(value: 'rename', child: Text('Rename')),
            PopupMenuItem(value: 'delete', child: Text('Delete')),
          ],
          onSelected: (v) async {
            final repo = sl<ComposeRepository>();
            final tpl = store.requestsById[node.data!.id];
            if (tpl == null) return;
            if (v == 'rename') {
              final name = await _promptText(
                context,
                title: 'Rename request',
                initial: tpl.name,
              );
              if (name == null || name.isEmpty) return;
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
            } else if (v == 'delete') {
              await store.deleteTemplate(tpl.id);
            }
          },
        ),
      ],
    );
  }
}

Future<String?> _promptText(
  BuildContext context, {
  required String title,
  String? initial,
}) async {
  final ctrl = TextEditingController(text: initial ?? '');
  return showDialog<String>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
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
}

bool _folderAdd(
  Map<String, dynamic> folder,
  String parentId,
  Map<String, dynamic> newFolder,
) {
  if (folder['id'] == parentId) {
    final List<dynamic> list = (folder['folders'] as List?) ?? <dynamic>[];
    list.add(newFolder);
    folder['folders'] = list;
    return true;
  }
  final List<dynamic> subs = (folder['folders'] as List?) ?? <dynamic>[];
  for (final s in subs) {
    if (_folderAdd((s as Map).cast<String, dynamic>(), parentId, newFolder)) {
      return true;
    }
  }
  return false;
}

bool _folderRename(Map<String, dynamic> folder, String id, String name) {
  if (folder['id'] == id) {
    folder['name'] = name;
    return true;
  }
  final List<dynamic> subs = (folder['folders'] as List?) ?? <dynamic>[];
  for (final s in subs) {
    if (_folderRename((s as Map).cast<String, dynamic>(), id, name))
      return true;
  }
  return false;
}

bool _folderDelete(Map<String, dynamic> folder, String id) {
  final List<dynamic> subs = (folder['folders'] as List?) ?? <dynamic>[];
  for (int i = 0; i < subs.length; i++) {
    final m = (subs[i] as Map).cast<String, dynamic>();
    if (m['id'] == id) {
      subs.removeAt(i);
      folder['folders'] = subs;
      return true;
    }
    if (_folderDelete(m, id)) return true;
  }
  return false;
}
