import 'package:firebase_database/firebase_database.dart';
import 'package:firebase_database_debugger/firebase_database_debugger.dart';
import 'package:flutter/foundation.dart';

Future<void> main() async {
  final db = FirebaseDatabase.instance;
  final debugger = FirebaseDatabaseDebugger(
    config: FirebaseDatabaseDebuggerConfig(
      enabled: kDebugMode,
    ),
  );

  final usersRef = debugger.ref(db.ref('users/alice'));
  await usersRef.set({'name': 'Alice', 'age': 30});
  await usersRef.update({'age': 31});
  await usersRef.get();

  await debugger.dispose();
}
