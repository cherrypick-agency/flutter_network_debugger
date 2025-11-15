import 'package:flutter/material.dart';

/// Controller for managing JSON search state
/// Can be shared between JsonViewer and external search UI
class JsonSearchController extends ChangeNotifier {
  bool _showSearch = false;
  String _searchText = '';
  bool _matchCase = false;
  bool _wholeWord = false;
  bool _useRegex = false;
  int _focusedIndex = 0;
  List<GlobalKey> _matchKeys = const [];

  // Getters
  bool get showSearch => _showSearch;
  String get searchText => _searchText;
  bool get matchCase => _matchCase;
  bool get wholeWord => _wholeWord;
  bool get useRegex => _useRegex;
  int get focusedIndex => _focusedIndex;
  List<GlobalKey> get matchKeys => _matchKeys;

  int get matchCount => _matchKeys.length;
  bool get hasMatches => _matchKeys.isNotEmpty;

  /// Open search bar
  void openSearch() {
    _showSearch = true;
    notifyListeners();
  }

  /// Close search bar and clear state
  void closeSearch() {
    _showSearch = false;
    _searchText = '';
    _focusedIndex = 0;
    _matchKeys = const [];
    notifyListeners();
  }

  /// Update search text
  void updateSearchText(String text) {
    _searchText = text;
    _focusedIndex = 0; // Reset to first match
    notifyListeners();
  }

  /// Toggle match case
  void toggleMatchCase() {
    _matchCase = !_matchCase;
    notifyListeners();
  }

  /// Toggle whole word
  void toggleWholeWord() {
    _wholeWord = !_wholeWord;
    notifyListeners();
  }

  /// Toggle regex
  void toggleRegex() {
    _useRegex = !_useRegex;
    notifyListeners();
  }

  /// Update match keys (called by JsonViewer after rebuild)
  void updateMatchKeys(List<GlobalKey> keys) {
    _matchKeys = keys;
    // Ensure focusedIndex is within bounds
    if (_focusedIndex >= keys.length && keys.isNotEmpty) {
      _focusedIndex = keys.length - 1;
    }
    notifyListeners();
  }

  /// Go to next match
  void nextMatch() {
    if (_matchKeys.isEmpty) return;
    _focusedIndex = (_focusedIndex + 1) % _matchKeys.length;
    _scrollToFocused();
    notifyListeners();
  }

  /// Go to previous match
  void prevMatch() {
    if (_matchKeys.isEmpty) return;
    _focusedIndex = (_focusedIndex - 1 + _matchKeys.length) % _matchKeys.length;
    _scrollToFocused();
    notifyListeners();
  }

  /// Scroll to focused match
  void _scrollToFocused() {
    if (_focusedIndex < _matchKeys.length) {
      final key = _matchKeys[_focusedIndex];
      final context = key.currentContext;
      if (context != null) {
        Scrollable.ensureVisible(
          context,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeInOut,
          alignment: 0.2, // 20% from top (~100-150px offset)
        );
      }
    }
  }
}
