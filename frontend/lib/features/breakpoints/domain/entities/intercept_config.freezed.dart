// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'intercept_config.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$InterceptConfig {

 bool get enabled; bool get requests; bool get responses; int get timeoutMs; int get queueMax; int get bodyMaxBytes; bool get reencode; String get overflow;
/// Create a copy of InterceptConfig
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$InterceptConfigCopyWith<InterceptConfig> get copyWith => _$InterceptConfigCopyWithImpl<InterceptConfig>(this as InterceptConfig, _$identity);

  /// Serializes this InterceptConfig to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is InterceptConfig&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.requests, requests) || other.requests == requests)&&(identical(other.responses, responses) || other.responses == responses)&&(identical(other.timeoutMs, timeoutMs) || other.timeoutMs == timeoutMs)&&(identical(other.queueMax, queueMax) || other.queueMax == queueMax)&&(identical(other.bodyMaxBytes, bodyMaxBytes) || other.bodyMaxBytes == bodyMaxBytes)&&(identical(other.reencode, reencode) || other.reencode == reencode)&&(identical(other.overflow, overflow) || other.overflow == overflow));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,enabled,requests,responses,timeoutMs,queueMax,bodyMaxBytes,reencode,overflow);

@override
String toString() {
  return 'InterceptConfig(enabled: $enabled, requests: $requests, responses: $responses, timeoutMs: $timeoutMs, queueMax: $queueMax, bodyMaxBytes: $bodyMaxBytes, reencode: $reencode, overflow: $overflow)';
}


}

/// @nodoc
abstract mixin class $InterceptConfigCopyWith<$Res>  {
  factory $InterceptConfigCopyWith(InterceptConfig value, $Res Function(InterceptConfig) _then) = _$InterceptConfigCopyWithImpl;
@useResult
$Res call({
 bool enabled, bool requests, bool responses, int timeoutMs, int queueMax, int bodyMaxBytes, bool reencode, String overflow
});




}
/// @nodoc
class _$InterceptConfigCopyWithImpl<$Res>
    implements $InterceptConfigCopyWith<$Res> {
  _$InterceptConfigCopyWithImpl(this._self, this._then);

  final InterceptConfig _self;
  final $Res Function(InterceptConfig) _then;

/// Create a copy of InterceptConfig
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? enabled = null,Object? requests = null,Object? responses = null,Object? timeoutMs = null,Object? queueMax = null,Object? bodyMaxBytes = null,Object? reencode = null,Object? overflow = null,}) {
  return _then(_self.copyWith(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,requests: null == requests ? _self.requests : requests // ignore: cast_nullable_to_non_nullable
as bool,responses: null == responses ? _self.responses : responses // ignore: cast_nullable_to_non_nullable
as bool,timeoutMs: null == timeoutMs ? _self.timeoutMs : timeoutMs // ignore: cast_nullable_to_non_nullable
as int,queueMax: null == queueMax ? _self.queueMax : queueMax // ignore: cast_nullable_to_non_nullable
as int,bodyMaxBytes: null == bodyMaxBytes ? _self.bodyMaxBytes : bodyMaxBytes // ignore: cast_nullable_to_non_nullable
as int,reencode: null == reencode ? _self.reencode : reencode // ignore: cast_nullable_to_non_nullable
as bool,overflow: null == overflow ? _self.overflow : overflow // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [InterceptConfig].
extension InterceptConfigPatterns on InterceptConfig {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _InterceptConfig value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _InterceptConfig() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _InterceptConfig value)  $default,){
final _that = this;
switch (_that) {
case _InterceptConfig():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _InterceptConfig value)?  $default,){
final _that = this;
switch (_that) {
case _InterceptConfig() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( bool enabled,  bool requests,  bool responses,  int timeoutMs,  int queueMax,  int bodyMaxBytes,  bool reencode,  String overflow)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _InterceptConfig() when $default != null:
return $default(_that.enabled,_that.requests,_that.responses,_that.timeoutMs,_that.queueMax,_that.bodyMaxBytes,_that.reencode,_that.overflow);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( bool enabled,  bool requests,  bool responses,  int timeoutMs,  int queueMax,  int bodyMaxBytes,  bool reencode,  String overflow)  $default,) {final _that = this;
switch (_that) {
case _InterceptConfig():
return $default(_that.enabled,_that.requests,_that.responses,_that.timeoutMs,_that.queueMax,_that.bodyMaxBytes,_that.reencode,_that.overflow);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( bool enabled,  bool requests,  bool responses,  int timeoutMs,  int queueMax,  int bodyMaxBytes,  bool reencode,  String overflow)?  $default,) {final _that = this;
switch (_that) {
case _InterceptConfig() when $default != null:
return $default(_that.enabled,_that.requests,_that.responses,_that.timeoutMs,_that.queueMax,_that.bodyMaxBytes,_that.reencode,_that.overflow);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _InterceptConfig extends InterceptConfig {
  const _InterceptConfig({this.enabled = false, this.requests = true, this.responses = true, this.timeoutMs = 60000, this.queueMax = 200, this.bodyMaxBytes = 1048576, this.reencode = true, this.overflow = 'auto-continue-oldest'}): super._();
  factory _InterceptConfig.fromJson(Map<String, dynamic> json) => _$InterceptConfigFromJson(json);

@override@JsonKey() final  bool enabled;
@override@JsonKey() final  bool requests;
@override@JsonKey() final  bool responses;
@override@JsonKey() final  int timeoutMs;
@override@JsonKey() final  int queueMax;
@override@JsonKey() final  int bodyMaxBytes;
@override@JsonKey() final  bool reencode;
@override@JsonKey() final  String overflow;

/// Create a copy of InterceptConfig
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$InterceptConfigCopyWith<_InterceptConfig> get copyWith => __$InterceptConfigCopyWithImpl<_InterceptConfig>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$InterceptConfigToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _InterceptConfig&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.requests, requests) || other.requests == requests)&&(identical(other.responses, responses) || other.responses == responses)&&(identical(other.timeoutMs, timeoutMs) || other.timeoutMs == timeoutMs)&&(identical(other.queueMax, queueMax) || other.queueMax == queueMax)&&(identical(other.bodyMaxBytes, bodyMaxBytes) || other.bodyMaxBytes == bodyMaxBytes)&&(identical(other.reencode, reencode) || other.reencode == reencode)&&(identical(other.overflow, overflow) || other.overflow == overflow));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,enabled,requests,responses,timeoutMs,queueMax,bodyMaxBytes,reencode,overflow);

@override
String toString() {
  return 'InterceptConfig(enabled: $enabled, requests: $requests, responses: $responses, timeoutMs: $timeoutMs, queueMax: $queueMax, bodyMaxBytes: $bodyMaxBytes, reencode: $reencode, overflow: $overflow)';
}


}

/// @nodoc
abstract mixin class _$InterceptConfigCopyWith<$Res> implements $InterceptConfigCopyWith<$Res> {
  factory _$InterceptConfigCopyWith(_InterceptConfig value, $Res Function(_InterceptConfig) _then) = __$InterceptConfigCopyWithImpl;
@override @useResult
$Res call({
 bool enabled, bool requests, bool responses, int timeoutMs, int queueMax, int bodyMaxBytes, bool reencode, String overflow
});




}
/// @nodoc
class __$InterceptConfigCopyWithImpl<$Res>
    implements _$InterceptConfigCopyWith<$Res> {
  __$InterceptConfigCopyWithImpl(this._self, this._then);

  final _InterceptConfig _self;
  final $Res Function(_InterceptConfig) _then;

/// Create a copy of InterceptConfig
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? enabled = null,Object? requests = null,Object? responses = null,Object? timeoutMs = null,Object? queueMax = null,Object? bodyMaxBytes = null,Object? reencode = null,Object? overflow = null,}) {
  return _then(_InterceptConfig(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,requests: null == requests ? _self.requests : requests // ignore: cast_nullable_to_non_nullable
as bool,responses: null == responses ? _self.responses : responses // ignore: cast_nullable_to_non_nullable
as bool,timeoutMs: null == timeoutMs ? _self.timeoutMs : timeoutMs // ignore: cast_nullable_to_non_nullable
as int,queueMax: null == queueMax ? _self.queueMax : queueMax // ignore: cast_nullable_to_non_nullable
as int,bodyMaxBytes: null == bodyMaxBytes ? _self.bodyMaxBytes : bodyMaxBytes // ignore: cast_nullable_to_non_nullable
as int,reencode: null == reencode ? _self.reencode : reencode // ignore: cast_nullable_to_non_nullable
as bool,overflow: null == overflow ? _self.overflow : overflow // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
