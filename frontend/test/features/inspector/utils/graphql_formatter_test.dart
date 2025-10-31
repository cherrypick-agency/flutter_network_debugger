import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/inspector/presentation/utils/graphql_formatter.dart';

void main() {
  group('GraphQLFormatter', () {
    test('formats simple query', () {
      const input = 'query{user{name}}';
      final result = GraphQLFormatter.format(input);

      expect(result, equals('query {\n  user {\n    name\n  }\n}'));
    });

    test('formats query with arguments', () {
      const input = 'query{user(id:1){name,email}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('user(id: 1)'));
      expect(result, contains('name,'));
      expect(result, contains('email'));
    });

    test('formats nested fields', () {
      const input = 'query{user{name,posts{title,body}}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('user {'));
      expect(result, contains('  name,'));
      expect(result, contains('  posts {'));
      expect(result, contains('    title,'));
      expect(result, contains('    body'));
    });

    test('formats mutation', () {
      const input = 'mutation{createUser(name:"John"){id,name}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('mutation {'));
      expect(result, contains('createUser(name: "John")'));
      expect(result, contains('id,'));
      expect(result, contains('name'));
    });

    test('formats query with operation name', () {
      const input = 'query GetUser{user(id:1){name}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('query GetUser {'));
    });

    test('formats query with directives', () {
      const input = 'query{user{name @include(if:true),email}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('@include(if: true)'));
    });

    test('preserves string literals', () {
      const input = 'query{user(name:"John Doe"){email}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('"John Doe"'));
    });

    test('handles complex minified query', () {
      const input =
          'query{user(id:1){name,email,posts{title,body,comments{author,text}}}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('query {'));
      expect(result, contains('  user(id: 1) {'));
      expect(result, contains('    name,'));
      expect(result, contains('    posts {'));
      expect(result, contains('      title,'));
      expect(result, contains('      comments {'));
      expect(result, contains('        author,'));
    });

    test('handles query with variables', () {
      const input = r'query GetUser($id:ID!){user(id:$id){name}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains(r'GetUser($id: ID!)'));
    });

    test('handles subscription', () {
      const input = 'subscription{messageAdded{id,text,author}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('subscription {'));
      expect(result, contains('messageAdded {'));
    });

    test('handles empty query', () {
      const input = '';
      final result = GraphQLFormatter.format(input);

      expect(result, equals(''));
    });

    test('handles whitespace-only query', () {
      const input = '   \n  \t  ';
      final result = GraphQLFormatter.format(input);

      expect(result, equals(''));
    });

    test('handles query with fragments', () {
      const input =
          'query{user{...UserFields}}fragment UserFields on User{name,email}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('...UserFields'));
      expect(result, contains('fragment UserFields on User'));
    });

    test('handles already formatted query', () {
      const input = '''query {
  user {
    name
  }
}''';
      final result = GraphQLFormatter.format(input);

      // Should produce similar formatting
      expect(result, contains('query {'));
      expect(result, contains('user {'));
      expect(result, contains('name'));
    });

    test('handles multiple arguments', () {
      const input = 'query{users(first:10,offset:5,order:DESC){name}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('users(first: 10,'));
      expect(result, contains('offset: 5,'));
      expect(result, contains('order: DESC)'));
    });

    test('handles inline fragments', () {
      const input = 'query{search{...on User{name}...on Post{title}}}';
      final result = GraphQLFormatter.format(input);

      expect(result, contains('...on User'));
      expect(result, contains('...on Post'));
    });
  });
}
