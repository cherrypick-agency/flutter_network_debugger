/// GraphQL query formatter
///
/// Formats GraphQL queries/mutations/subscriptions with proper indentation
/// and spacing for better readability.
class GraphQLFormatter {
  /// Format a GraphQL query string with proper indentation
  ///
  /// [query] - The GraphQL query string to format
  /// [indent] - Number of spaces per indentation level (default: 2)
  ///
  /// Returns formatted GraphQL query
  static String format(String query, {int indent = 2}) {
    if (query.trim().isEmpty) return '';

    final buffer = StringBuffer();
    int indentLevel = 0;
    bool inString = false;
    String? stringChar;
    bool needIndent = false;
    bool lastWasNewline = false;
    bool lastWasWhitespace = false;

    for (int i = 0; i < query.length; i++) {
      final char = query[i];
      final prevChar = i > 0 ? query[i - 1] : null;

      // Track string literals
      if ((char == '"' || char == "'") && prevChar != '\\') {
        if (!inString) {
          inString = true;
          stringChar = char;
        } else if (char == stringChar) {
          inString = false;
          stringChar = null;
        }
        buffer.write(char);
        needIndent = false;
        lastWasNewline = false;
        lastWasWhitespace = false;
        continue;
      }

      // Don't format inside strings
      if (inString) {
        buffer.write(char);
        continue;
      }

      // Track whitespace
      if (_isWhitespace(char)) {
        lastWasWhitespace = true;
        continue;
      }

      // Add indentation if needed
      if (needIndent) {
        buffer.write('  ' * indentLevel);
        needIndent = false;
        lastWasNewline = false;
      }

      // Add space between identifiers if there was whitespace in source
      // (query GetUser, fragment UserFields, on User)
      if (lastWasWhitespace &&
          buffer.isNotEmpty &&
          !lastWasNewline &&
          _isIdentifierChar(char) &&
          _isIdentifierChar(buffer.toString()[buffer.length - 1])) {
        buffer.write(' ');
      }
      lastWasWhitespace = false;

      switch (char) {
        case '{':
          buffer.write(' {\n');
          indentLevel++;
          needIndent = true;
          lastWasNewline = true;
          break;

        case '}':
          if (!lastWasNewline) {
            buffer.write('\n');
          }
          indentLevel--;
          buffer.write('  ' * indentLevel);
          buffer.write('}');
          needIndent = false;
          lastWasNewline = false;
          break;

        case '(':
          buffer.write('(');
          needIndent = false;
          lastWasNewline = false;
          break;

        case ')':
          buffer.write(')');
          needIndent = false;
          lastWasNewline = false;
          break;

        case ',':
          buffer.write(',\n');
          needIndent = true;
          lastWasNewline = true;
          break;

        case ':':
          buffer.write(': ');
          needIndent = false;
          lastWasNewline = false;
          break;

        case '@':
          if (!lastWasNewline) {
            buffer.write(' ');
          }
          buffer.write('@');
          needIndent = false;
          lastWasNewline = false;
          break;

        default:
          buffer.write(char);
          needIndent = false;
          lastWasNewline = false;
      }
    }

    return buffer.toString().trim();
  }

  static bool _isWhitespace(String char) {
    return char == ' ' || char == '\n' || char == '\r' || char == '\t';
  }

  static bool _isIdentifierChar(String char) {
    final code = char.codeUnitAt(0);
    return (code >= 65 && code <= 90) || // A-Z
        (code >= 97 && code <= 122) || // a-z
        (code >= 48 && code <= 57) || // 0-9
        char == '_' ||
        char == '\$';
  }
}
