import 'package:flutter/material.dart';
import 'package:animated_tree_view/animated_tree_view.dart';

import '../../application/compose_store.dart';
import '../../data/compose_repository.dart';
import '../../../../../../core/di/di.dart';

class FolderSelection {
  final String collectionId;
  final String folderId;
  const FolderSelection({required this.collectionId, required this.folderId});
}

class FolderPickerDialog extends StatefulWidget {
  const FolderPickerDialog({super.key});

  @override
  State<FolderPickerDialog> createState() => _FolderPickerDialogState();
}

class _FolderPickerDialogState extends State<FolderPickerDialog> {
  late final ComposeStore _store;
  late final TreeNode<_FolderNode> _root;
  FolderSelection? _selected;

  @override
  void initState() {
    super.initState();
    _store = sl<ComposeStore>();
    _root = TreeNode<_FolderNode>(key: 'root', data: const _FolderNode.root());
    _rebuild();
  }

  void _rebuild() {
    _root.clear();
    for (final c in _store.collections) {
      final colNode = TreeNode<_FolderNode>(
        key: 'col:${c.id}',
        data: _FolderNode.collection(c.id, c.name),
      );
      _root.add(colNode);
      _buildFolder(colNode, c.id, c.root);
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

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Выберите папку'),
      content: SizedBox(
        width: 420,
        height: 380,
        child: TreeView.simpleTyped<_FolderNode, TreeNode<_FolderNode>>(
          tree: _root,
          builder: (context, node) {
            final d = node.data!;
            if (d.isRoot) return const SizedBox.shrink();
            final isFolder = d.kind == _FolderKind.folder;
            final selected =
                _selected?.collectionId == d.collectionId &&
                _selected?.folderId == d.folderId;
            return InkWell(
              onTap:
                  isFolder
                      ? () {
                        setState(() {
                          _selected = FolderSelection(
                            collectionId: d.collectionId!,
                            folderId: d.folderId!,
                          );
                        });
                      }
                      : null,
              child: Row(
                children: [
                  Icon(
                    isFolder ? Icons.folder : Icons.folder_copy,
                    size: 16,
                    color:
                        isFolder
                            ? Theme.of(context).colorScheme.primary
                            : Theme.of(context).iconTheme.color,
                  ),
                  const SizedBox(width: 6),
                  Expanded(child: Text(d.title)),
                  if (!isFolder)
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
                        await _store.loadLibrary();
                        _rebuild();
                      },
                    )
                  else ...[
                    Radio<String>(
                      value: d.folderId!,
                      groupValue: _selected?.folderId,
                      onChanged: (_) {
                        setState(() {
                          _selected = FolderSelection(
                            collectionId: d.collectionId!,
                            folderId: d.folderId!,
                          );
                        });
                      },
                    ),
                    PopupMenuButton<String>(
                      tooltip: 'Folder menu',
                      itemBuilder:
                          (ctx) => const [
                            PopupMenuItem(
                              value: 'new',
                              child: Text('New subfolder'),
                            ),
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
                        final root = col.root.toJson();
                        if (v == 'new') {
                          final name = await _promptText(
                            context,
                            title: 'New folder name',
                          );
                          if (name == null || name.isEmpty) return;
                          _folderAdd(root, d.folderId!, {
                            'id':
                                'fld-${DateTime.now().microsecondsSinceEpoch}',
                            'name': name,
                            'requests': <String>[],
                            'folders': <Map<String, dynamic>>[],
                          });
                        } else if (v == 'rename') {
                          final name = await _promptText(
                            context,
                            title: 'Rename folder',
                            initial: d.title,
                          );
                          if (name == null || name.isEmpty) return;
                          _folderRename(root, d.folderId!, name);
                        } else if (v == 'delete') {
                          final ok = await showDialog<bool>(
                            context: context,
                            builder:
                                (ctx) => AlertDialog(
                                  title: const Text('Confirm'),
                                  content: const Text(
                                    'Delete folder and its contents?',
                                  ),
                                  actions: [
                                    TextButton(
                                      onPressed:
                                          () => Navigator.of(ctx).pop(false),
                                      child: const Text('Cancel'),
                                    ),
                                    FilledButton(
                                      onPressed:
                                          () => Navigator.of(ctx).pop(true),
                                      child: const Text('OK'),
                                    ),
                                  ],
                                ),
                          );
                          if (ok != true) return;
                          _folderDelete(root, d.folderId!);
                        }
                        await repo.upsertCollection({
                          'id': col.id,
                          'name': col.name,
                          'root': root,
                        });
                        await _store.loadLibrary();
                        _rebuild();
                      },
                    ),
                  ],
                ],
              ),
            );
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed:
              _selected == null
                  ? null
                  : () => Navigator.of(context).pop(_selected),
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
  const _FolderNode.collection(String colId, String title)
    : this._(_FolderKind.collection, title, colId, null);
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
    builder:
        (ctx) => AlertDialog(
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
