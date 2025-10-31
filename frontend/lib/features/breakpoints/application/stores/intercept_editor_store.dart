import 'dart:convert';
import 'package:flutter/foundation.dart';
import '../../domain/entities/intercept_item.dart';
import '../../domain/entities/decisions.dart';
import '../../domain/repositories/breakpoints_repository.dart';

class InterceptEditorStore extends ChangeNotifier {
  InterceptEditorStore(this._repo);
  final BreakpointsRepository _repo;

  InterceptItem? _item;
  String _headersError = '';
  String _bodyJsonError = '';

  InterceptItem? get item => _item;
  String get headersError => _headersError;
  String get bodyJsonError => _bodyJsonError;

  void setItem(InterceptItem? it) {
    _item = it;
    _headersError = '';
    _bodyJsonError = '';
    notifyListeners();
  }

  void validateJson(String? text) {
    _bodyJsonError = '';
    if (text == null || text.trim().isEmpty) {
      notifyListeners();
      return;
    }
    try {
      json.decode(text);
    } catch (e) {
      _bodyJsonError = 'Некорректный JSON';
    }
    notifyListeners();
  }

  Future<void> continueRequest({
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? bodyBase64,
    bool drop = false,
  }) async {
    if (_item == null) return;
    await _repo.continueRequest(
      _item!.id,
      RequestDecision(
        action: drop ? 'drop' : 'continue',
        method: method,
        url: url,
        headers: headers,
        bodyBase64: bodyBase64,
      ),
    );
  }

  Future<void> continueResponse({
    int? status,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) async {
    if (_item == null) return;
    await _repo.continueResponse(
      _item!.id,
      ResponseDecision(
        action: 'continue',
        status: status,
        headers: headers,
        bodyBase64: bodyBase64,
      ),
    );
  }

  Future<void> cancel() async {
    if (_item == null) return;
    await _repo.cancel(_item!.id);
  }
}
