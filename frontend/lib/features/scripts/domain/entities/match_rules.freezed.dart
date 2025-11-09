// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'match_rules.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$MatchRules {

 List<String> get methods; String? get pathPattern; String? get hostPattern; PatternType get patternType;
/// Create a copy of MatchRules
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MatchRulesCopyWith<MatchRules> get copyWith => _$MatchRulesCopyWithImpl<MatchRules>(this as MatchRules, _$identity);

  /// Serializes this MatchRules to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is MatchRules&&const DeepCollectionEquality().equals(other.methods, methods)&&(identical(other.pathPattern, pathPattern) || other.pathPattern == pathPattern)&&(identical(other.hostPattern, hostPattern) || other.hostPattern == hostPattern)&&(identical(other.patternType, patternType) || other.patternType == patternType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(methods),pathPattern,hostPattern,patternType);

@override
String toString() {
  return 'MatchRules(methods: $methods, pathPattern: $pathPattern, hostPattern: $hostPattern, patternType: $patternType)';
}


}

/// @nodoc
abstract mixin class $MatchRulesCopyWith<$Res>  {
  factory $MatchRulesCopyWith(MatchRules value, $Res Function(MatchRules) _then) = _$MatchRulesCopyWithImpl;
@useResult
$Res call({
 List<String> methods, String? pathPattern, String? hostPattern, PatternType patternType
});




}
/// @nodoc
class _$MatchRulesCopyWithImpl<$Res>
    implements $MatchRulesCopyWith<$Res> {
  _$MatchRulesCopyWithImpl(this._self, this._then);

  final MatchRules _self;
  final $Res Function(MatchRules) _then;

/// Create a copy of MatchRules
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? methods = null,Object? pathPattern = freezed,Object? hostPattern = freezed,Object? patternType = null,}) {
  return _then(_self.copyWith(
methods: null == methods ? _self.methods : methods // ignore: cast_nullable_to_non_nullable
as List<String>,pathPattern: freezed == pathPattern ? _self.pathPattern : pathPattern // ignore: cast_nullable_to_non_nullable
as String?,hostPattern: freezed == hostPattern ? _self.hostPattern : hostPattern // ignore: cast_nullable_to_non_nullable
as String?,patternType: null == patternType ? _self.patternType : patternType // ignore: cast_nullable_to_non_nullable
as PatternType,
  ));
}

}


/// Adds pattern-matching-related methods to [MatchRules].
extension MatchRulesPatterns on MatchRules {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _MatchRules value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _MatchRules() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _MatchRules value)  $default,){
final _that = this;
switch (_that) {
case _MatchRules():
return $default(_that);}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _MatchRules value)?  $default,){
final _that = this;
switch (_that) {
case _MatchRules() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<String> methods,  String? pathPattern,  String? hostPattern,  PatternType patternType)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _MatchRules() when $default != null:
return $default(_that.methods,_that.pathPattern,_that.hostPattern,_that.patternType);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<String> methods,  String? pathPattern,  String? hostPattern,  PatternType patternType)  $default,) {final _that = this;
switch (_that) {
case _MatchRules():
return $default(_that.methods,_that.pathPattern,_that.hostPattern,_that.patternType);}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<String> methods,  String? pathPattern,  String? hostPattern,  PatternType patternType)?  $default,) {final _that = this;
switch (_that) {
case _MatchRules() when $default != null:
return $default(_that.methods,_that.pathPattern,_that.hostPattern,_that.patternType);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _MatchRules extends MatchRules {
  const _MatchRules({final  List<String> methods = const [], this.pathPattern, this.hostPattern, this.patternType = PatternType.wildcard}): _methods = methods,super._();
  factory _MatchRules.fromJson(Map<String, dynamic> json) => _$MatchRulesFromJson(json);

 final  List<String> _methods;
@override@JsonKey() List<String> get methods {
  if (_methods is EqualUnmodifiableListView) return _methods;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_methods);
}

@override final  String? pathPattern;
@override final  String? hostPattern;
@override@JsonKey() final  PatternType patternType;

/// Create a copy of MatchRules
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MatchRulesCopyWith<_MatchRules> get copyWith => __$MatchRulesCopyWithImpl<_MatchRules>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$MatchRulesToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _MatchRules&&const DeepCollectionEquality().equals(other._methods, _methods)&&(identical(other.pathPattern, pathPattern) || other.pathPattern == pathPattern)&&(identical(other.hostPattern, hostPattern) || other.hostPattern == hostPattern)&&(identical(other.patternType, patternType) || other.patternType == patternType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_methods),pathPattern,hostPattern,patternType);

@override
String toString() {
  return 'MatchRules(methods: $methods, pathPattern: $pathPattern, hostPattern: $hostPattern, patternType: $patternType)';
}


}

/// @nodoc
abstract mixin class _$MatchRulesCopyWith<$Res> implements $MatchRulesCopyWith<$Res> {
  factory _$MatchRulesCopyWith(_MatchRules value, $Res Function(_MatchRules) _then) = __$MatchRulesCopyWithImpl;
@override @useResult
$Res call({
 List<String> methods, String? pathPattern, String? hostPattern, PatternType patternType
});




}
/// @nodoc
class __$MatchRulesCopyWithImpl<$Res>
    implements _$MatchRulesCopyWith<$Res> {
  __$MatchRulesCopyWithImpl(this._self, this._then);

  final _MatchRules _self;
  final $Res Function(_MatchRules) _then;

/// Create a copy of MatchRules
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? methods = null,Object? pathPattern = freezed,Object? hostPattern = freezed,Object? patternType = null,}) {
  return _then(_MatchRules(
methods: null == methods ? _self._methods : methods // ignore: cast_nullable_to_non_nullable
as List<String>,pathPattern: freezed == pathPattern ? _self.pathPattern : pathPattern // ignore: cast_nullable_to_non_nullable
as String?,hostPattern: freezed == hostPattern ? _self.hostPattern : hostPattern // ignore: cast_nullable_to_non_nullable
as String?,patternType: null == patternType ? _self.patternType : patternType // ignore: cast_nullable_to_non_nullable
as PatternType,
  ));
}


}

// dart format on
