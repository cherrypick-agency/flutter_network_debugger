// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'mapping_config.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;
/// @nodoc
mixin _$MappingConfig {

 bool get enabled; int get uploadMaxMB;
/// Create a copy of MappingConfig
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MappingConfigCopyWith<MappingConfig> get copyWith => _$MappingConfigCopyWithImpl<MappingConfig>(this as MappingConfig, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is MappingConfig&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.uploadMaxMB, uploadMaxMB) || other.uploadMaxMB == uploadMaxMB));
}


@override
int get hashCode => Object.hash(runtimeType,enabled,uploadMaxMB);

@override
String toString() {
  return 'MappingConfig(enabled: $enabled, uploadMaxMB: $uploadMaxMB)';
}


}

/// @nodoc
abstract mixin class $MappingConfigCopyWith<$Res>  {
  factory $MappingConfigCopyWith(MappingConfig value, $Res Function(MappingConfig) _then) = _$MappingConfigCopyWithImpl;
@useResult
$Res call({
 bool enabled, int uploadMaxMB
});




}
/// @nodoc
class _$MappingConfigCopyWithImpl<$Res>
    implements $MappingConfigCopyWith<$Res> {
  _$MappingConfigCopyWithImpl(this._self, this._then);

  final MappingConfig _self;
  final $Res Function(MappingConfig) _then;

/// Create a copy of MappingConfig
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? enabled = null,Object? uploadMaxMB = null,}) {
  return _then(_self.copyWith(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,uploadMaxMB: null == uploadMaxMB ? _self.uploadMaxMB : uploadMaxMB // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [MappingConfig].
extension MappingConfigPatterns on MappingConfig {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _MappingConfig value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _MappingConfig() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _MappingConfig value)  $default,){
final _that = this;
switch (_that) {
case _MappingConfig():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _MappingConfig value)?  $default,){
final _that = this;
switch (_that) {
case _MappingConfig() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( bool enabled,  int uploadMaxMB)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _MappingConfig() when $default != null:
return $default(_that.enabled,_that.uploadMaxMB);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( bool enabled,  int uploadMaxMB)  $default,) {final _that = this;
switch (_that) {
case _MappingConfig():
return $default(_that.enabled,_that.uploadMaxMB);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( bool enabled,  int uploadMaxMB)?  $default,) {final _that = this;
switch (_that) {
case _MappingConfig() when $default != null:
return $default(_that.enabled,_that.uploadMaxMB);case _:
  return null;

}
}

}

/// @nodoc


class _MappingConfig extends MappingConfig {
  const _MappingConfig({this.enabled = true, this.uploadMaxMB = 20}): super._();
  

@override@JsonKey() final  bool enabled;
@override@JsonKey() final  int uploadMaxMB;

/// Create a copy of MappingConfig
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MappingConfigCopyWith<_MappingConfig> get copyWith => __$MappingConfigCopyWithImpl<_MappingConfig>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _MappingConfig&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.uploadMaxMB, uploadMaxMB) || other.uploadMaxMB == uploadMaxMB));
}


@override
int get hashCode => Object.hash(runtimeType,enabled,uploadMaxMB);

@override
String toString() {
  return 'MappingConfig(enabled: $enabled, uploadMaxMB: $uploadMaxMB)';
}


}

/// @nodoc
abstract mixin class _$MappingConfigCopyWith<$Res> implements $MappingConfigCopyWith<$Res> {
  factory _$MappingConfigCopyWith(_MappingConfig value, $Res Function(_MappingConfig) _then) = __$MappingConfigCopyWithImpl;
@override @useResult
$Res call({
 bool enabled, int uploadMaxMB
});




}
/// @nodoc
class __$MappingConfigCopyWithImpl<$Res>
    implements _$MappingConfigCopyWith<$Res> {
  __$MappingConfigCopyWithImpl(this._self, this._then);

  final _MappingConfig _self;
  final $Res Function(_MappingConfig) _then;

/// Create a copy of MappingConfig
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? enabled = null,Object? uploadMaxMB = null,}) {
  return _then(_MappingConfig(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,uploadMaxMB: null == uploadMaxMB ? _self.uploadMaxMB : uploadMaxMB // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
