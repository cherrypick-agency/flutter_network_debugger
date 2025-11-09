// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'script_config.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ScriptConfig {

 int? get timeoutMs; int? get memoryLimitMB; List<String> get allowedHosts;
/// Create a copy of ScriptConfig
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ScriptConfigCopyWith<ScriptConfig> get copyWith => _$ScriptConfigCopyWithImpl<ScriptConfig>(this as ScriptConfig, _$identity);

  /// Serializes this ScriptConfig to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ScriptConfig&&(identical(other.timeoutMs, timeoutMs) || other.timeoutMs == timeoutMs)&&(identical(other.memoryLimitMB, memoryLimitMB) || other.memoryLimitMB == memoryLimitMB)&&const DeepCollectionEquality().equals(other.allowedHosts, allowedHosts));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,timeoutMs,memoryLimitMB,const DeepCollectionEquality().hash(allowedHosts));

@override
String toString() {
  return 'ScriptConfig(timeoutMs: $timeoutMs, memoryLimitMB: $memoryLimitMB, allowedHosts: $allowedHosts)';
}


}

/// @nodoc
abstract mixin class $ScriptConfigCopyWith<$Res>  {
  factory $ScriptConfigCopyWith(ScriptConfig value, $Res Function(ScriptConfig) _then) = _$ScriptConfigCopyWithImpl;
@useResult
$Res call({
 int? timeoutMs, int? memoryLimitMB, List<String> allowedHosts
});




}
/// @nodoc
class _$ScriptConfigCopyWithImpl<$Res>
    implements $ScriptConfigCopyWith<$Res> {
  _$ScriptConfigCopyWithImpl(this._self, this._then);

  final ScriptConfig _self;
  final $Res Function(ScriptConfig) _then;

/// Create a copy of ScriptConfig
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? timeoutMs = freezed,Object? memoryLimitMB = freezed,Object? allowedHosts = null,}) {
  return _then(_self.copyWith(
timeoutMs: freezed == timeoutMs ? _self.timeoutMs : timeoutMs // ignore: cast_nullable_to_non_nullable
as int?,memoryLimitMB: freezed == memoryLimitMB ? _self.memoryLimitMB : memoryLimitMB // ignore: cast_nullable_to_non_nullable
as int?,allowedHosts: null == allowedHosts ? _self.allowedHosts : allowedHosts // ignore: cast_nullable_to_non_nullable
as List<String>,
  ));
}

}


/// Adds pattern-matching-related methods to [ScriptConfig].
extension ScriptConfigPatterns on ScriptConfig {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ScriptConfig value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ScriptConfig() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ScriptConfig value)  $default,){
final _that = this;
switch (_that) {
case _ScriptConfig():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ScriptConfig value)?  $default,){
final _that = this;
switch (_that) {
case _ScriptConfig() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int? timeoutMs,  int? memoryLimitMB,  List<String> allowedHosts)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ScriptConfig() when $default != null:
return $default(_that.timeoutMs,_that.memoryLimitMB,_that.allowedHosts);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int? timeoutMs,  int? memoryLimitMB,  List<String> allowedHosts)  $default,) {final _that = this;
switch (_that) {
case _ScriptConfig():
return $default(_that.timeoutMs,_that.memoryLimitMB,_that.allowedHosts);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int? timeoutMs,  int? memoryLimitMB,  List<String> allowedHosts)?  $default,) {final _that = this;
switch (_that) {
case _ScriptConfig() when $default != null:
return $default(_that.timeoutMs,_that.memoryLimitMB,_that.allowedHosts);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ScriptConfig extends ScriptConfig {
  const _ScriptConfig({this.timeoutMs, this.memoryLimitMB, final  List<String> allowedHosts = const []}): _allowedHosts = allowedHosts,super._();
  factory _ScriptConfig.fromJson(Map<String, dynamic> json) => _$ScriptConfigFromJson(json);

@override final  int? timeoutMs;
@override final  int? memoryLimitMB;
 final  List<String> _allowedHosts;
@override@JsonKey() List<String> get allowedHosts {
  if (_allowedHosts is EqualUnmodifiableListView) return _allowedHosts;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_allowedHosts);
}


/// Create a copy of ScriptConfig
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ScriptConfigCopyWith<_ScriptConfig> get copyWith => __$ScriptConfigCopyWithImpl<_ScriptConfig>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ScriptConfigToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ScriptConfig&&(identical(other.timeoutMs, timeoutMs) || other.timeoutMs == timeoutMs)&&(identical(other.memoryLimitMB, memoryLimitMB) || other.memoryLimitMB == memoryLimitMB)&&const DeepCollectionEquality().equals(other._allowedHosts, _allowedHosts));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,timeoutMs,memoryLimitMB,const DeepCollectionEquality().hash(_allowedHosts));

@override
String toString() {
  return 'ScriptConfig(timeoutMs: $timeoutMs, memoryLimitMB: $memoryLimitMB, allowedHosts: $allowedHosts)';
}


}

/// @nodoc
abstract mixin class _$ScriptConfigCopyWith<$Res> implements $ScriptConfigCopyWith<$Res> {
  factory _$ScriptConfigCopyWith(_ScriptConfig value, $Res Function(_ScriptConfig) _then) = __$ScriptConfigCopyWithImpl;
@override @useResult
$Res call({
 int? timeoutMs, int? memoryLimitMB, List<String> allowedHosts
});




}
/// @nodoc
class __$ScriptConfigCopyWithImpl<$Res>
    implements _$ScriptConfigCopyWith<$Res> {
  __$ScriptConfigCopyWithImpl(this._self, this._then);

  final _ScriptConfig _self;
  final $Res Function(_ScriptConfig) _then;

/// Create a copy of ScriptConfig
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? timeoutMs = freezed,Object? memoryLimitMB = freezed,Object? allowedHosts = null,}) {
  return _then(_ScriptConfig(
timeoutMs: freezed == timeoutMs ? _self.timeoutMs : timeoutMs // ignore: cast_nullable_to_non_nullable
as int?,memoryLimitMB: freezed == memoryLimitMB ? _self.memoryLimitMB : memoryLimitMB // ignore: cast_nullable_to_non_nullable
as int?,allowedHosts: null == allowedHosts ? _self._allowedHosts : allowedHosts // ignore: cast_nullable_to_non_nullable
as List<String>,
  ));
}


}

// dart format on
