import 'package:flutter/material.dart';
import 'package:animated_tree_view/animated_tree_view.dart';

import '../../application/compose_store.dart';
import '../../data/compose_repository.dart';
import '../../../../../../core/di/di.dart';

class FolderSelection {
  final String collectionId;
  final String folderId;
  final String? requestName;
  const FolderSelection({
    required this.collectionId,
    required this.folderId,
    this.requestName,
  });
}

class FolderPickerDialog extends StatefulWidget {
  const FolderPickerDialog({super.key, this.initialRequestName});
  final String? initialRequestName;

  @override
  State<FolderPickerDialog> createState() => _FolderPickerDialogState();
}

class _FolderPickerDialogState extends State<FolderPickerDialog> {
  late final ComposeStore _store;
  late final TreeNode<_FolderNode> _root;
  FolderSelection? _selected;
  late final TextEditingController _nameCtrl;

  @override
  void initState() {
    super.initState();
    _store = sl<ComposeStore>();
    _root = TreeNode<_FolderNode>(key: 'root', data: const _FolderNode.root());
    _nameCtrl = TextEditingController(text: widget.initialRequestName ?? '');
    _rebuild();
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    super.dispose();
  }

  void _rebuild() {
    _root.clear();
    for (final c in _store.collections) {
      final colNode = TreeNode<_FolderNode>(
        key: 'col:${c.id}',
        data: _FolderNode.collection(c.id, c.name, c.root.id),
      );
      _root.add(colNode);
      // Don't show root as a folder node - add its subfolders directly to collection
      for (final subfolder in c.root.folders) {
        _buildFolder(colNode, c.id, subfolder);
      }
    }
    setState(() {});
  }

  void _buildFolder(
    TreeNode<_FolderNode> parent,
    String colId,
    ComposeFolderModel f,
  ) {
    final node = TreeNode<_FolderNode>(
      key: 'fld:${f.id}',
      data: _FolderNode.folder(colId, f.id, f.name),
    );
    parent.add(node);
    for (final sub in f.folders) {
      _buildFolder(node, colId, sub);
    }
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.folder_off,
            size: 64,
            color: Colors.grey.withValues(alpha: 0.5),
          ),
          const SizedBox(height: 16),
          Text(
            'No collections yet',
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 8),
          Text(
            'Create your first collection',
            style: Theme.of(
              context,
            ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
          ),
        ],
      ),
    );
  }

  Future<void> _onCreateCollection() async {
    final name = await _promptText(
      context,
      title: 'Collection name',
      initial: 'My Collection',
    );

    if (name == null || name.isEmpty) return;

    final repo = sl<ComposeRepository>();
    final collectionId = 'col-${DateTime.now().microsecondsSinceEpoch}';
    final rootFolderId = 'fld-${DateTime.now().microsecondsSinceEpoch}';

    // Create collection with root folder
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

    // Reload library and rebuild tree
    await _store.loadLibrary();
    _rebuild();

    // Automatically select created collection
    setState(() {
      _selected = FolderSelection(
        collectionId: collectionId,
        folderId: rootFolderId,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Select folder'),
      content: SizedBox(
        width: 420,
        height: 440,
        child: _store.collections.isEmpty
            ? _buildEmptyState()
            : Column(
                children: [
                  TextField(
                    controller: _nameCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Request name',
                      hintText: 'e.g., Get user by id',
                    ),
                  ),
                  const SizedBox(height: 8),
                  Expanded(
                    child: TreeView.simpleTyped<_FolderNode, TreeNode<_FolderNode>>(
                      tree: _root,
                      showRootNode: false,
                      expansionBehavior: ExpansionBehavior.none,
                      builder: (context, node) {
                        final d = node.data!;
                        if (d.isRoot) return const SizedBox.shrink();
                        final isFolder = d.kind == _FolderKind.folder;
                        final selected =
                            _selected?.collectionId == d.collectionId &&
                            _selected?.folderId == d.folderId;
                        return Container(
                          color: selected
                              ? Theme.of(context).colorScheme.primaryContainer
                                    .withValues(alpha: 0.3)
                              : null,
                          child: InkWell(
                            onTap: () {
                              setState(() {
                                _selected = FolderSelection(
                                  collectionId: d.collectionId!,
                                  folderId: d.folderId!,
                                  requestName: _nameCtrl.text.trim(),
                                );
                              });
                            },
                            child: Row(
                              children: [
                                Icon(
                                  isFolder ? Icons.folder : Icons.folder_copy,
                                  size: 16,
                                  color: isFolder
                                      ? Theme.of(context).colorScheme.primary
                                      : Theme.of(context).iconTheme.color,
                                ),
                                const SizedBox(width: 6),
                                Expanded(child: Text(d.title)),
                                IconButton(
                                  tooltip: 'New folder',
                                  icon: const Icon(Icons.add, size: 16),
                                  onPressed: () async {
                                    final name = await _promptText(
                                      context,
                                      title: 'New folder name',
                                    );
                                    if (name == null || name.isEmpty) return;
                                    final repo = sl<ComposeRepository>();
                                    final col = _store.collections.firstWhere(
                                      (c) => c.id == d.collectionId,
                                      orElse: () => _store.collections.first,
                                    );
                                    final root = col.root.toJson();
                                    // Add to the current folder (for both collection and folder nodes)
                                    _folderAdd(root, d.folderId!, {
                                      'id':
                                          'fld-${DateTime.now().microsecondsSinceEpoch}',
                                      'name': name,
                                      'requests': <String>[],
                                      'folders': <Map<String, dynamic>>[],
                                    });
                                    await repo.upsertCollection({
                                      'id': col.id,
                                      'name': col.name,
                                      'root': root,
                                    });
                                    await _store.loadLibrary();
                                    _rebuild();
                                  },
                                ),
                                Radio<String>(
                                  value: d.folderId!,
                                  groupValue: _selected?.folderId,
                                  onChanged: (_) {
                                    setState(() {
                                      _selected = FolderSelection(
                                        collectionId: d.collectionId!,
                                        folderId: d.folderId!,
                                        requestName: _nameCtrl.text.trim(),
                                      );
                                    });
                                  },
                                ),
                                PopupMenuButton<String>(
                                  tooltip: isFolder
                                      ? 'Folder menu'
                                      : 'Collection menu',
                                  itemBuilder: (ctx) => const [
                                    PopupMenuItem(
                                      value: 'rename',
                                      child: Text('Rename'),
                                    ),
                                    PopupMenuItem(
                                      value: 'delete',
                                      child: Text('Delete'),
                                    ),
                                  ],
                                  onSelected: (v) async {
                                    final repo = sl<ComposeRepository>();
                                    final col = _store.collections.firstWhere(
                                      (c) => c.id == d.collectionId,
                                      orElse: () => _store.collections.first,
                                    );
                                    if (v == 'rename') {
                                      final name = await _promptText(
                                        context,
                                        title: isFolder
                                            ? 'Rename folder'
                                            : 'Rename collection',
                                        initial: d.title,
                                      );
                                      if (name == null || name.isEmpty) return;
                                      if (isFolder) {
                                        final root = col.root.toJson();
                                        _folderRename(root, d.folderId!, name);
                                        await repo.upsertCollection({
                                          'id': col.id,
                                          'name': col.name,
                                          'root': root,
                                        });
                                      } else {
                                        await repo.upsertCollection({
                                          'id': col.id,
                                          'name': name,
                                          'root': col.root.toJson(),
                                        });
                                      }
                                    } else if (v == 'delete') {
                                      final ok = await showDialog<bool>(
                                        context: context,
                                        builder: (ctx) => AlertDialog(
                                          title: const Text('Confirm'),
                                          content: Text(
                                            isFolder
                                                ? 'Delete folder and its contents?'
                                                : 'Delete collection?',
                                          ),
                                          actions: [
                                            TextButton(
                                              onPressed: () =>
                                                  Navigator.of(ctx).pop(false),
                                              child: const Text('Cancel'),
                                            ),
                                            FilledButton(
                                              onPressed: () =>
                                                  Navigator.of(ctx).pop(true),
                                              child: const Text('OK'),
                                            ),
                                          ],
                                        ),
                                      );
                                      if (ok != true) return;
                                      if (isFolder) {
                                        final root = col.root.toJson();
                                        _folderDelete(root, d.folderId!);
                                        await repo.upsertCollection({
                                          'id': col.id,
                                          'name': col.name,
                                          'root': root,
                                        });
                                      } else {
                                        await repo.deleteCollection(col.id);
                                      }
                                    }
                                    await _store.loadLibrary();
                                    _rebuild();
                                  },
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],
              ),
      ),
      actions: [
        TextButton.icon(
          onPressed: _onCreateCollection,
          icon: const Icon(Icons.create_new_folder, size: 18),
          label: const Text('Create collection'),
        ),
        const Spacer(),
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _selected == null
              ? null
              : () {
                  final res = FolderSelection(
                    collectionId: _selected!.collectionId,
                    folderId: _selected!.folderId,
                    requestName: _nameCtrl.text.trim(),
                  );
                  Navigator.of(context).pop(res);
                },
          child: const Text('Save'),
        ),
      ],
    );
  }
}

enum _FolderKind { root, collection, folder }

class _FolderNode {
  final _FolderKind kind;
  final String title;
  final String? collectionId;
  final String? folderId;

  const _FolderNode._(this.kind, this.title, this.collectionId, this.folderId);
  const _FolderNode.root() : this._(_FolderKind.root, 'root', null, null);
  const _FolderNode.collection(String colId, String title, String folderId)
    : this._(_FolderKind.collection, title, colId, folderId);
  const _FolderNode.folder(String colId, String folderId, String title)
    : this._(_FolderKind.folder, title, colId, folderId);

  bool get isRoot => kind == _FolderKind.root;
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
