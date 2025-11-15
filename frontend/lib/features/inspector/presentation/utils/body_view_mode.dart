import 'package:flutter/material.dart';

/// View modes for displaying HTTP body content
enum BodyViewMode {
  /// Pretty-printed with syntax highlighting
  pretty,

  /// JSON tree view (expandable/collapsible)
  jsonTree,

  /// Rendered HTML preview
  htmlPreview,

  /// Decoded Base64 content
  base64Decoded,

  /// JWT token viewer with header/payload/signature
  jwtDecoded,

  /// Raw text without formatting
  raw,

  /// Hexadecimal viewer
  hex,
}

/// Controller for managing body view state (SOLID: Single Responsibility)
class BodyViewController {
  BodyViewMode _currentMode = BodyViewMode.pretty;
  final Set<BodyViewMode> _availableModes = {};

  BodyViewController();

  /// Get current view mode
  BodyViewMode get current => _currentMode;

  /// Get all available modes for current content
  Set<BodyViewMode> get available => Set.unmodifiable(_availableModes);

  /// Set available modes based on content detection
  void setAvailableModes(Set<BodyViewMode> modes) {
    _availableModes.clear();
    _availableModes.addAll(modes);

    // If current mode is not available, switch to first available
    if (!_availableModes.contains(_currentMode)) {
      _currentMode = _availableModes.isNotEmpty
          ? _availableModes.first
          : BodyViewMode.raw;
    }
  }

  /// Switch to specific view mode
  void switchTo(BodyViewMode mode) {
    if (_availableModes.contains(mode)) {
      _currentMode = mode;
    }
  }

  /// Check if mode is available
  bool isAvailable(BodyViewMode mode) => _availableModes.contains(mode);

  /// Check if mode is currently active
  bool isActive(BodyViewMode mode) => _currentMode == mode;

  /// Get display label for mode
  static String getLabel(BodyViewMode mode) {
    switch (mode) {
      case BodyViewMode.pretty:
        return 'Pretty';
      case BodyViewMode.jsonTree:
        return 'Tree';
      case BodyViewMode.htmlPreview:
        return 'HTML';
      case BodyViewMode.base64Decoded:
        return 'Decoded';
      case BodyViewMode.jwtDecoded:
        return 'JWT';
      case BodyViewMode.raw:
        return 'Raw';
      case BodyViewMode.hex:
        return 'Hex';
    }
  }

  /// Get icon for mode
  static IconData? getIcon(BodyViewMode mode) {
    switch (mode) {
      case BodyViewMode.pretty:
        return null;
      case BodyViewMode.jsonTree:
        return null;
      case BodyViewMode.htmlPreview:
        return null;
      case BodyViewMode.base64Decoded:
        return null;
      case BodyViewMode.jwtDecoded:
        return null;
      case BodyViewMode.raw:
        return null;
      case BodyViewMode.hex:
        return null;
    }
  }
}
