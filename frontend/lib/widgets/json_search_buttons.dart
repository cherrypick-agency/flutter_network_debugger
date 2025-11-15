import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../core/hotkeys/hotkeys_service.dart';
import '../core/di/di.dart';
import 'common_search_bar.dart';
import 'json_search_controller.dart';

/// Search and copy buttons for JSON viewer
/// Can be positioned externally using Stack/Positioned
class JsonSearchButtons extends StatefulWidget {
  const JsonSearchButtons({
    super.key,
    required this.controller,
    required this.jsonString,
    this.focusNode,
  });

  final JsonSearchController controller;
  final String jsonString;
  final FocusNode? focusNode;

  @override
  State<JsonSearchButtons> createState() => _JsonSearchButtonsState();
}

class _JsonSearchButtonsState extends State<JsonSearchButtons> {
  late final TextEditingController _searchCtrl;
  late final FocusNode _searchFocusNode;

  @override
  void initState() {
    super.initState();
    _searchCtrl = TextEditingController();
    _searchFocusNode = widget.focusNode ?? FocusNode();

    // Listen to controller changes
    widget.controller.addListener(_onControllerChange);
  }

  @override
  void dispose() {
    widget.controller.removeListener(_onControllerChange);
    _searchCtrl.dispose();
    if (widget.focusNode == null) {
      _searchFocusNode.dispose();
    }
    super.dispose();
  }

  void _onControllerChange() {
    // Sync text field with controller
    if (_searchCtrl.text != widget.controller.searchText) {
      _searchCtrl.text = widget.controller.searchText;
    }
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    if (widget.controller.showSearch) {
      return _buildSearchBar(context);
    } else {
      return _buildSearchButton(context);
    }
  }

  Widget _buildSearchButton(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            iconSize: 18,
            padding: const EdgeInsets.all(4),
            constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
            tooltip: 'Copy JSON',
            onPressed: () {
              Clipboard.setData(ClipboardData(text: widget.jsonString));
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('JSON copied to clipboard'),
                  duration: Duration(seconds: 1),
                ),
              );
            },
            icon: const Icon(Icons.copy, size: 18),
          ),
          const SizedBox(width: 4),
          IconButton(
            iconSize: 18,
            padding: const EdgeInsets.all(4),
            constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
            tooltip: 'Search',
            onPressed: () {
              widget.controller.openSearch();
              WidgetsBinding.instance.addPostFrameCallback((_) {
                _searchFocusNode.requestFocus();
              });
            },
            icon: const Icon(Icons.search, size: 18),
          ),
        ],
      ),
    );
  }

  Widget _buildSearchBar(BuildContext context) {
    final countText = widget.controller.hasMatches
        ? '${widget.controller.focusedIndex + 1}/${widget.controller.matchCount}'
        : '0/0';

    final hk = sl<HotkeysService>();
    void closeSearch() {
      widget.controller.closeSearch();
      _searchFocusNode.unfocus();
    }

    final handlers = hk.buildHandlers({
      'jsonSearch.next': widget.controller.nextMatch,
      'jsonSearch.prev': widget.controller.prevMatch,
      'jsonSearch.close': closeSearch,
    });

    return CallbackShortcuts(
      bindings: handlers,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 400),
        child: CommonSearchBar(
          controller: _searchCtrl,
          focusNode: _searchFocusNode,
          countText: countText,
          matchCase: widget.controller.matchCase,
          wholeWord: widget.controller.wholeWord,
          useRegex: widget.controller.useRegex,
          canNavigate: widget.controller.hasMatches,
          onChanged: () {
            widget.controller.updateSearchText(_searchCtrl.text);
          },
          onNext: widget.controller.nextMatch,
          onPrev: widget.controller.prevMatch,
          onClose: closeSearch,
          onToggleMatchCase: widget.controller.toggleMatchCase,
          onToggleWholeWord: widget.controller.toggleWholeWord,
          onToggleRegex: widget.controller.toggleRegex,
        ),
      ),
    );
  }
}
