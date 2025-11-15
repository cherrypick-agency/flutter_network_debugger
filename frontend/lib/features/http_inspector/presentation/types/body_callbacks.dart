library;

/// Common callback type definitions for HTTP Inspector widgets
///
/// This file contains shared typedef declarations used across multiple
/// presentation layer widgets to avoid code duplication.

/// Callback for cURL command generation and copying
///
/// Takes the generated cURL command string and performs an action
/// (typically copying to clipboard or showing in a dialog)
typedef CurlCallback = void Function(String curl);
