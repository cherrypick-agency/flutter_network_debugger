import 'package:multi_editor_plugins/multi_editor_plugins.dart';

/// Collection of pure Dart code snippets for autocomplete.
///
/// Contains only core Dart language constructs, without any Flutter dependencies.
class DartSnippets {
  DartSnippets._();

  // ============================================================================
  // BASIC DART LANGUAGE CONSTRUCTS
  // ============================================================================

  static const main = SnippetData(
    prefix: 'main',
    label: 'main function',
    body: 'void main() {\n  \${1:// code}\n}',
    description: 'Main entry point',
  );

  static const mainAsync = SnippetData(
    prefix: 'maina',
    label: 'async main function',
    body: 'Future<void> main() async {\n  \${1:// code}\n}',
    description: 'Async main entry point',
  );

  static const classSnippet = SnippetData(
    prefix: 'class',
    label: 'class',
    body: 'class \${1:ClassName} {\n  \${2:// fields and methods}\n}',
    description: 'Class declaration',
  );

  static const abstractClass = SnippetData(
    prefix: 'absclass',
    label: 'abstract class',
    body: 'abstract class \${1:ClassName} {\n  \${2:// abstract methods}\n}',
    description: 'Abstract class declaration',
  );

  static const enumSnippet = SnippetData(
    prefix: 'enum',
    label: 'enum',
    body: 'enum \${1:EnumName} {\n  \${2:value1},\n  \${3:value2},\n}',
    description: 'Enum declaration',
  );

  static const ifStatement = SnippetData(
    prefix: 'if',
    label: 'if statement',
    body: 'if (\${1:condition}) {\n  \${2:// code}\n}',
    description: 'If statement',
  );

  static const ifElseStatement = SnippetData(
    prefix: 'ifelse',
    label: 'if-else statement',
    body:
        'if (\${1:condition}) {\n  \${2:// code}\n} else {\n  \${3:// code}\n}',
    description: 'If-else statement',
  );

  static const forLoop = SnippetData(
    prefix: 'for',
    label: 'for loop',
    body:
        'for (var \${1:i} = 0; \${1:i} < \${2:length}; \${1:i}++) {\n  \${3:// code}\n}',
    description: 'For loop',
  );

  static const forInLoop = SnippetData(
    prefix: 'forin',
    label: 'for-in loop',
    body: 'for (var \${1:item} in \${2:collection}) {\n  \${3:// code}\n}',
    description: 'For-in loop',
  );

  static const whileLoop = SnippetData(
    prefix: 'while',
    label: 'while loop',
    body: 'while (\${1:condition}) {\n  \${2:// code}\n}',
    description: 'While loop',
  );

  static const tryCatch = SnippetData(
    prefix: 'try',
    label: 'try-catch',
    body:
        'try {\n  \${1:// code}\n} catch (\${2:e}) {\n  \${3:// error handling}\n}',
    description: 'Try-catch block',
  );

  static const tryCatchFinally = SnippetData(
    prefix: 'tryf',
    label: 'try-catch-finally',
    body:
        'try {\n  \${1:// code}\n} catch (\${2:e}) {\n  \${3:// error handling}\n} finally {\n  \${4:// cleanup}\n}',
    description: 'Try-catch-finally block',
  );

  static const asyncFunction = SnippetData(
    prefix: 'funa',
    label: 'async function',
    body: 'Future<\${1:void}> \${2:functionName}() async {\n  \${3:// code}\n}',
    description: 'Async function',
  );

  static const function = SnippetData(
    prefix: 'fun',
    label: 'function',
    body: '\${1:void} \${2:functionName}() {\n  \${3:// code}\n}',
    description: 'Function declaration',
  );

  // ============================================================================
  // COMPLETE SNIPPET LIST
  // ============================================================================

  /// All available pure Dart snippets (14 total).
  static const List<SnippetData> all = [
    main,
    mainAsync,
    classSnippet,
    abstractClass,
    enumSnippet,
    ifStatement,
    ifElseStatement,
    forLoop,
    forInLoop,
    whileLoop,
    tryCatch,
    tryCatchFinally,
    asyncFunction,
    function,
  ];
}
