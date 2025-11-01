import 'package:flutter/foundation.dart';

import '../../compose/data/compose_repository.dart';
import '../../compose/domain/models.dart';

class ComposeStore extends ChangeNotifier {
  final ComposeRepository _repo;
  ComposeStore(this._repo);

  bool loading = false;
  Map<String, ComposeTemplateDTO> requestsById = {};
  List<ComposeCollectionModel> collections = [];

  ComposeTemplateDTO? selected;

  Future<void> loadLibrary() async {
    loading = true;
    notifyListeners();
    try {
      final data = await _repo.library();
      requestsById = {};
      final reqs =
          (data['requestsById'] as Map?)?.cast<String, dynamic>() ?? {};
      reqs.forEach((k, v) {
        requestsById[k] = ComposeTemplateDTO.fromJson(
          v as Map<String, dynamic>,
        );
      });
      final cols = (data['collections'] as List?) ?? const [];
      collections =
          cols
              .map(
                (e) => ComposeCollectionModel.fromJson(
                  (e as Map).cast<String, dynamic>(),
                ),
              )
              .toList();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  void selectTemplate(ComposeTemplateDTO t) {
    selected = t;
    notifyListeners();
  }

  Future<void> saveTemplate(ComposeTemplateDTO t) async {
    await _repo.upsertRequest(t);
    await loadLibrary();
  }

  Future<void> deleteTemplate(String id) async {
    await _repo.deleteRequest(id);
    await loadLibrary();
  }
}

class ComposeCollectionModel {
  final String id;
  final String name;
  final ComposeFolderModel root;
  ComposeCollectionModel({
    required this.id,
    required this.name,
    required this.root,
  });
  factory ComposeCollectionModel.fromJson(Map<String, dynamic> json) {
    return ComposeCollectionModel(
      id: json['id'] as String,
      name: json['name'] as String,
      root: ComposeFolderModel.fromJson(
        (json['root'] as Map).cast<String, dynamic>(),
      ),
    );
  }
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'root': root.toJson(),
  };
}

class ComposeFolderModel {
  final String id;
  final String name;
  final List<String> requests;
  final List<ComposeFolderModel> folders;
  ComposeFolderModel({
    required this.id,
    required this.name,
    required this.requests,
    required this.folders,
  });
  factory ComposeFolderModel.fromJson(Map<String, dynamic> json) {
    return ComposeFolderModel(
      id: json['id'] as String,
      name: json['name'] as String,
      requests: (json['requests'] as List?)?.cast<String>() ?? const [],
      folders:
          ((json['folders'] as List?) ?? const [])
              .map(
                (e) => ComposeFolderModel.fromJson(
                  (e as Map).cast<String, dynamic>(),
                ),
              )
              .toList(),
    );
  }
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'requests': requests,
    'folders': folders.map((e) => e.toJson()).toList(),
  };
}
