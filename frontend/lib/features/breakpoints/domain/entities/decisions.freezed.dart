// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'decisions.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RequestDecision {

 String get action; String? get method; String? get url; Map<String, List<String>>? get headers; String? get bodyBase64;
/// Create a copy of RequestDecision
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RequestDecisionCopyWith<RequestDecision> get copyWith => _$RequestDecisionCopyWithImpl<RequestDecision>(this as RequestDecision, _$identity);

  /// Serializes this RequestDecision to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RequestDecision&&(identical(other.action, action) || other.action == action)&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,action,method,url,const DeepCollectionEquality().hash(headers),bodyBase64);

@override
String toString() {
  return 'RequestDecision(action: $action, method: $method, url: $url, headers: $headers, bodyBase64: $bodyBase64)';
}


}

/// @nodoc
abstract mixin class $RequestDecisionCopyWith<$Res>  {
  factory $RequestDecisionCopyWith(RequestDecision value, $Res Function(RequestDecision) _then) = _$RequestDecisionCopyWithImpl;
@useResult
$Res call({
 String action, String? method, String? url, Map<String, List<String>>? headers, String? bodyBase64
});




}
/// @nodoc
class _$RequestDecisionCopyWithImpl<$Res>
    implements $RequestDecisionCopyWith<$Res> {
  _$RequestDecisionCopyWithImpl(this._self, this._then);

  final RequestDecision _self;
  final $Res Function(RequestDecision) _then;

/// Create a copy of RequestDecision
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? action = null,Object? method = freezed,Object? url = freezed,Object? headers = freezed,Object? bodyBase64 = freezed,}) {
  return _then(_self.copyWith(
action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,url: freezed == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String?,headers: freezed == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [RequestDecision].
extension RequestDecisionPatterns on RequestDecision {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RequestDecision value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RequestDecision() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RequestDecision value)  $default,){
final _that = this;
switch (_that) {
case _RequestDecision():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RequestDecision value)?  $default,){
final _that = this;
switch (_that) {
case _RequestDecision() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String action,  String? method,  String? url,  Map<String, List<String>>? headers,  String? bodyBase64)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RequestDecision() when $default != null:
return $default(_that.action,_that.method,_that.url,_that.headers,_that.bodyBase64);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String action,  String? method,  String? url,  Map<String, List<String>>? headers,  String? bodyBase64)  $default,) {final _that = this;
switch (_that) {
case _RequestDecision():
return $default(_that.action,_that.method,_that.url,_that.headers,_that.bodyBase64);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String action,  String? method,  String? url,  Map<String, List<String>>? headers,  String? bodyBase64)?  $default,) {final _that = this;
switch (_that) {
case _RequestDecision() when $default != null:
return $default(_that.action,_that.method,_that.url,_that.headers,_that.bodyBase64);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RequestDecision extends RequestDecision {
  const _RequestDecision({required this.action, this.method, this.url, final  Map<String, List<String>>? headers, this.bodyBase64}): _headers = headers,super._();
  factory _RequestDecision.fromJson(Map<String, dynamic> json) => _$RequestDecisionFromJson(json);

@override final  String action;
@override final  String? method;
@override final  String? url;
 final  Map<String, List<String>>? _headers;
@override Map<String, List<String>>? get headers {
  final value = _headers;
  if (value == null) return null;
  if (_headers is EqualUnmodifiableMapView) return _headers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(value);
}

@override final  String? bodyBase64;

/// Create a copy of RequestDecision
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RequestDecisionCopyWith<_RequestDecision> get copyWith => __$RequestDecisionCopyWithImpl<_RequestDecision>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RequestDecisionToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _RequestDecision&&(identical(other.action, action) || other.action == action)&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,action,method,url,const DeepCollectionEquality().hash(_headers),bodyBase64);

@override
String toString() {
  return 'RequestDecision(action: $action, method: $method, url: $url, headers: $headers, bodyBase64: $bodyBase64)';
}


}

/// @nodoc
abstract mixin class _$RequestDecisionCopyWith<$Res> implements $RequestDecisionCopyWith<$Res> {
  factory _$RequestDecisionCopyWith(_RequestDecision value, $Res Function(_RequestDecision) _then) = __$RequestDecisionCopyWithImpl;
@override @useResult
$Res call({
 String action, String? method, String? url, Map<String, List<String>>? headers, String? bodyBase64
});




}
/// @nodoc
class __$RequestDecisionCopyWithImpl<$Res>
    implements _$RequestDecisionCopyWith<$Res> {
  __$RequestDecisionCopyWithImpl(this._self, this._then);

  final _RequestDecision _self;
  final $Res Function(_RequestDecision) _then;

/// Create a copy of RequestDecision
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? action = null,Object? method = freezed,Object? url = freezed,Object? headers = freezed,Object? bodyBase64 = freezed,}) {
  return _then(_RequestDecision(
action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,url: freezed == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String?,headers: freezed == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}


/// @nodoc
mixin _$ResponseDecision {

 String get action; int? get status; Map<String, List<String>>? get headers; String? get bodyBase64;
/// Create a copy of ResponseDecision
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ResponseDecisionCopyWith<ResponseDecision> get copyWith => _$ResponseDecisionCopyWithImpl<ResponseDecision>(this as ResponseDecision, _$identity);

  /// Serializes this ResponseDecision to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ResponseDecision&&(identical(other.action, action) || other.action == action)&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,action,status,const DeepCollectionEquality().hash(headers),bodyBase64);

@override
String toString() {
  return 'ResponseDecision(action: $action, status: $status, headers: $headers, bodyBase64: $bodyBase64)';
}


}

/// @nodoc
abstract mixin class $ResponseDecisionCopyWith<$Res>  {
  factory $ResponseDecisionCopyWith(ResponseDecision value, $Res Function(ResponseDecision) _then) = _$ResponseDecisionCopyWithImpl;
@useResult
$Res call({
 String action, int? status, Map<String, List<String>>? headers, String? bodyBase64
});




}
/// @nodoc
class _$ResponseDecisionCopyWithImpl<$Res>
    implements $ResponseDecisionCopyWith<$Res> {
  _$ResponseDecisionCopyWithImpl(this._self, this._then);

  final ResponseDecision _self;
  final $Res Function(ResponseDecision) _then;

/// Create a copy of ResponseDecision
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? action = null,Object? status = freezed,Object? headers = freezed,Object? bodyBase64 = freezed,}) {
  return _then(_self.copyWith(
action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int?,headers: freezed == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [ResponseDecision].
extension ResponseDecisionPatterns on ResponseDecision {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ResponseDecision value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ResponseDecision() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ResponseDecision value)  $default,){
final _that = this;
switch (_that) {
case _ResponseDecision():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ResponseDecision value)?  $default,){
final _that = this;
switch (_that) {
case _ResponseDecision() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String action,  int? status,  Map<String, List<String>>? headers,  String? bodyBase64)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ResponseDecision() when $default != null:
return $default(_that.action,_that.status,_that.headers,_that.bodyBase64);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String action,  int? status,  Map<String, List<String>>? headers,  String? bodyBase64)  $default,) {final _that = this;
switch (_that) {
case _ResponseDecision():
return $default(_that.action,_that.status,_that.headers,_that.bodyBase64);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String action,  int? status,  Map<String, List<String>>? headers,  String? bodyBase64)?  $default,) {final _that = this;
switch (_that) {
case _ResponseDecision() when $default != null:
return $default(_that.action,_that.status,_that.headers,_that.bodyBase64);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ResponseDecision extends ResponseDecision {
  const _ResponseDecision({required this.action, this.status, final  Map<String, List<String>>? headers, this.bodyBase64}): _headers = headers,super._();
  factory _ResponseDecision.fromJson(Map<String, dynamic> json) => _$ResponseDecisionFromJson(json);

@override final  String action;
@override final  int? status;
 final  Map<String, List<String>>? _headers;
@override Map<String, List<String>>? get headers {
  final value = _headers;
  if (value == null) return null;
  if (_headers is EqualUnmodifiableMapView) return _headers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(value);
}

@override final  String? bodyBase64;

/// Create a copy of ResponseDecision
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ResponseDecisionCopyWith<_ResponseDecision> get copyWith => __$ResponseDecisionCopyWithImpl<_ResponseDecision>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ResponseDecisionToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ResponseDecision&&(identical(other.action, action) || other.action == action)&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,action,status,const DeepCollectionEquality().hash(_headers),bodyBase64);

@override
String toString() {
  return 'ResponseDecision(action: $action, status: $status, headers: $headers, bodyBase64: $bodyBase64)';
}


}

/// @nodoc
abstract mixin class _$ResponseDecisionCopyWith<$Res> implements $ResponseDecisionCopyWith<$Res> {
  factory _$ResponseDecisionCopyWith(_ResponseDecision value, $Res Function(_ResponseDecision) _then) = __$ResponseDecisionCopyWithImpl;
@override @useResult
$Res call({
 String action, int? status, Map<String, List<String>>? headers, String? bodyBase64
});




}
/// @nodoc
class __$ResponseDecisionCopyWithImpl<$Res>
    implements _$ResponseDecisionCopyWith<$Res> {
  __$ResponseDecisionCopyWithImpl(this._self, this._then);

  final _ResponseDecision _self;
  final $Res Function(_ResponseDecision) _then;

/// Create a copy of ResponseDecision
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? action = null,Object? status = freezed,Object? headers = freezed,Object? bodyBase64 = freezed,}) {
  return _then(_ResponseDecision(
action: null == action ? _self.action : action // ignore: cast_nullable_to_non_nullable
as String,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int?,headers: freezed == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
