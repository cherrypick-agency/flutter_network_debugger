// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sessions_filters_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$SessionsFiltersStore on _SessionsFiltersStore, Store {
  late final _$targetAtom = Atom(
    name: '_SessionsFiltersStore.target',
    context: context,
  );

  @override
  String get target {
    _$targetAtom.reportRead();
    return super.target;
  }

  @override
  set target(String value) {
    _$targetAtom.reportWrite(value, super.target, () {
      super.target = value;
    });
  }

  late final _$httpMethodAtom = Atom(
    name: '_SessionsFiltersStore.httpMethod',
    context: context,
  );

  @override
  String get httpMethod {
    _$httpMethodAtom.reportRead();
    return super.httpMethod;
  }

  @override
  set httpMethod(String value) {
    _$httpMethodAtom.reportWrite(value, super.httpMethod, () {
      super.httpMethod = value;
    });
  }

  late final _$httpStatusAtom = Atom(
    name: '_SessionsFiltersStore.httpStatus',
    context: context,
  );

  @override
  String get httpStatus {
    _$httpStatusAtom.reportRead();
    return super.httpStatus;
  }

  @override
  set httpStatus(String value) {
    _$httpStatusAtom.reportWrite(value, super.httpStatus, () {
      super.httpStatus = value;
    });
  }

  late final _$httpMimeAtom = Atom(
    name: '_SessionsFiltersStore.httpMime',
    context: context,
  );

  @override
  String get httpMime {
    _$httpMimeAtom.reportRead();
    return super.httpMime;
  }

  @override
  set httpMime(String value) {
    _$httpMimeAtom.reportWrite(value, super.httpMime, () {
      super.httpMime = value;
    });
  }

  late final _$httpMinDurationMsAtom = Atom(
    name: '_SessionsFiltersStore.httpMinDurationMs',
    context: context,
  );

  @override
  int get httpMinDurationMs {
    _$httpMinDurationMsAtom.reportRead();
    return super.httpMinDurationMs;
  }

  @override
  set httpMinDurationMs(int value) {
    _$httpMinDurationMsAtom.reportWrite(value, super.httpMinDurationMs, () {
      super.httpMinDurationMs = value;
    });
  }

  late final _$groupByAtom = Atom(
    name: '_SessionsFiltersStore.groupBy',
    context: context,
  );

  @override
  String get groupBy {
    _$groupByAtom.reportRead();
    return super.groupBy;
  }

  @override
  set groupBy(String value) {
    _$groupByAtom.reportWrite(value, super.groupBy, () {
      super.groupBy = value;
    });
  }

  late final _$headerKeyAtom = Atom(
    name: '_SessionsFiltersStore.headerKey',
    context: context,
  );

  @override
  String get headerKey {
    _$headerKeyAtom.reportRead();
    return super.headerKey;
  }

  @override
  set headerKey(String value) {
    _$headerKeyAtom.reportWrite(value, super.headerKey, () {
      super.headerKey = value;
    });
  }

  late final _$headerValAtom = Atom(
    name: '_SessionsFiltersStore.headerVal',
    context: context,
  );

  @override
  String get headerVal {
    _$headerValAtom.reportRead();
    return super.headerVal;
  }

  @override
  set headerVal(String value) {
    _$headerValAtom.reportWrite(value, super.headerVal, () {
      super.headerVal = value;
    });
  }

  Computed<bool>? _$hasActiveComputed;

  @override
  bool get hasActive =>
      (_$hasActiveComputed ??= Computed<bool>(
            () => super.hasActive,
            name: '_SessionsFiltersStore.hasActive',
          ))
          .value;

  @override
  String toString() {
    return '''
target: ${target},
httpMethod: ${httpMethod},
httpStatus: ${httpStatus},
httpMime: ${httpMime},
httpMinDurationMs: ${httpMinDurationMs},
groupBy: ${groupBy},
headerKey: ${headerKey},
headerVal: ${headerVal},
hasActive: ${hasActive}
    ''';
  }
}
