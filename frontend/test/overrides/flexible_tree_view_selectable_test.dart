import 'package:flexible_tree_view/flexible_tree_view.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('FlexibleTreeView wraps tree in SelectionArea', (tester) async {
    final nodes = [
      TreeNode<String>(
        data: 'root',
        expanded: true,
        children: [TreeNode<String>(data: 'child')],
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: FlexibleTreeView<String>(
            nodes: nodes,
            nodeItemBuilder: (context, node) => Text(node.data),
          ),
        ),
      ),
    );

    expect(find.byType(SelectionArea), findsOneWidget);
    expect(find.text('root'), findsOneWidget);
    expect(find.text('child'), findsOneWidget);
  });
}
