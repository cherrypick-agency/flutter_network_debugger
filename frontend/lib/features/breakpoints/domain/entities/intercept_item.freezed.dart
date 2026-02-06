// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'intercept_item.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HTTPRequestSnapshot {

 String get method; String get url; Map<String, List<String>> get headers; String? get bodyBase64; bool get bodyTruncated; String? get contentType;
/// Create a copy of HTTPRequestSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HTTPRequestSnapshotCopyWith<HTTPRequestSnapshot> get copyWith => _$HTTPRequestSnapshotCopyWithImpl<HTTPRequestSnapshot>(this as HTTPRequestSnapshot, _$identity);

  /// Serializes this HTTPRequestSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HTTPRequestSnapshot&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64)&&(identical(other.bodyTruncated, bodyTruncated) || other.bodyTruncated == bodyTruncated)&&(identical(other.contentType, contentType) || other.contentType == contentType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(headers),bodyBase64,bodyTruncated,contentType);

@override
String toString() {
  return 'HTTPRequestSnapshot(method: $method, url: $url, headers: $headers, bodyBase64: $bodyBase64, bodyTruncated: $bodyTruncated, contentType: $contentType)';
}


}

/// @nodoc
abstract mixin class $HTTPRequestSnapshotCopyWith<$Res>  {
  factory $HTTPRequestSnapshotCopyWith(HTTPRequestSnapshot value, $Res Function(HTTPRequestSnapshot) _then) = _$HTTPRequestSnapshotCopyWithImpl;
@useResult
$Res call({
 String method, String url, Map<String, List<String>> headers, String? bodyBase64, bool bodyTruncated, String? contentType
});




}
/// @nodoc
class _$HTTPRequestSnapshotCopyWithImpl<$Res>
    implements $HTTPRequestSnapshotCopyWith<$Res> {
  _$HTTPRequestSnapshotCopyWithImpl(this._self, this._then);

  final HTTPRequestSnapshot _self;
  final $Res Function(HTTPRequestSnapshot) _then;

/// Create a copy of HTTPRequestSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? method = null,Object? url = null,Object? headers = null,Object? bodyBase64 = freezed,Object? bodyTruncated = null,Object? contentType = freezed,}) {
  return _then(_self.copyWith(
method: null == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String,url: null == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String,headers: null == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,bodyTruncated: null == bodyTruncated ? _self.bodyTruncated : bodyTruncated // ignore: cast_nullable_to_non_nullable
as bool,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [HTTPRequestSnapshot].
extension HTTPRequestSnapshotPatterns on HTTPRequestSnapshot {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HTTPRequestSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HTTPRequestSnapshot() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HTTPRequestSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _HTTPRequestSnapshot():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HTTPRequestSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _HTTPRequestSnapshot() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String method,  String url,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HTTPRequestSnapshot() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String method,  String url,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)  $default,) {final _that = this;
switch (_that) {
case _HTTPRequestSnapshot():
return $default(_that.method,_that.url,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String method,  String url,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)?  $default,) {final _that = this;
switch (_that) {
case _HTTPRequestSnapshot() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HTTPRequestSnapshot extends HTTPRequestSnapshot {
  const _HTTPRequestSnapshot({this.method = '', this.url = '', final  Map<String, List<String>> headers = const {}, this.bodyBase64, this.bodyTruncated = false, this.contentType}): _headers = headers,super._();
  factory _HTTPRequestSnapshot.fromJson(Map<String, dynamic> json) => _$HTTPRequestSnapshotFromJson(json);

@override@JsonKey() final  String method;
@override@JsonKey() final  String url;
 final  Map<String, List<String>> _headers;
@override@JsonKey() Map<String, List<String>> get headers {
  if (_headers is EqualUnmodifiableMapView) return _headers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(_headers);
}

@override final  String? bodyBase64;
@override@JsonKey() final  bool bodyTruncated;
@override final  String? contentType;

/// Create a copy of HTTPRequestSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HTTPRequestSnapshotCopyWith<_HTTPRequestSnapshot> get copyWith => __$HTTPRequestSnapshotCopyWithImpl<_HTTPRequestSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HTTPRequestSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _HTTPRequestSnapshot&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64)&&(identical(other.bodyTruncated, bodyTruncated) || other.bodyTruncated == bodyTruncated)&&(identical(other.contentType, contentType) || other.contentType == contentType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(_headers),bodyBase64,bodyTruncated,contentType);

@override
String toString() {
  return 'HTTPRequestSnapshot(method: $method, url: $url, headers: $headers, bodyBase64: $bodyBase64, bodyTruncated: $bodyTruncated, contentType: $contentType)';
}


}

/// @nodoc
abstract mixin class _$HTTPRequestSnapshotCopyWith<$Res> implements $HTTPRequestSnapshotCopyWith<$Res> {
  factory _$HTTPRequestSnapshotCopyWith(_HTTPRequestSnapshot value, $Res Function(_HTTPRequestSnapshot) _then) = __$HTTPRequestSnapshotCopyWithImpl;
@override @useResult
$Res call({
 String method, String url, Map<String, List<String>> headers, String? bodyBase64, bool bodyTruncated, String? contentType
});




}
/// @nodoc
class __$HTTPRequestSnapshotCopyWithImpl<$Res>
    implements _$HTTPRequestSnapshotCopyWith<$Res> {
  __$HTTPRequestSnapshotCopyWithImpl(this._self, this._then);

  final _HTTPRequestSnapshot _self;
  final $Res Function(_HTTPRequestSnapshot) _then;

/// Create a copy of HTTPRequestSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? method = null,Object? url = null,Object? headers = null,Object? bodyBase64 = freezed,Object? bodyTruncated = null,Object? contentType = freezed,}) {
  return _then(_HTTPRequestSnapshot(
method: null == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String,url: null == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String,headers: null == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,bodyTruncated: null == bodyTruncated ? _self.bodyTruncated : bodyTruncated // ignore: cast_nullable_to_non_nullable
as bool,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}


/// @nodoc
mixin _$HTTPResponseSnapshot {

 int get status; Map<String, List<String>> get headers; String? get bodyBase64; bool get bodyTruncated; String? get contentType;
/// Create a copy of HTTPResponseSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HTTPResponseSnapshotCopyWith<HTTPResponseSnapshot> get copyWith => _$HTTPResponseSnapshotCopyWithImpl<HTTPResponseSnapshot>(this as HTTPResponseSnapshot, _$identity);

  /// Serializes this HTTPResponseSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HTTPResponseSnapshot&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64)&&(identical(other.bodyTruncated, bodyTruncated) || other.bodyTruncated == bodyTruncated)&&(identical(other.contentType, contentType) || other.contentType == contentType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,status,const DeepCollectionEquality().hash(headers),bodyBase64,bodyTruncated,contentType);

@override
String toString() {
  return 'HTTPResponseSnapshot(status: $status, headers: $headers, bodyBase64: $bodyBase64, bodyTruncated: $bodyTruncated, contentType: $contentType)';
}


}

/// @nodoc
abstract mixin class $HTTPResponseSnapshotCopyWith<$Res>  {
  factory $HTTPResponseSnapshotCopyWith(HTTPResponseSnapshot value, $Res Function(HTTPResponseSnapshot) _then) = _$HTTPResponseSnapshotCopyWithImpl;
@useResult
$Res call({
 int status, Map<String, List<String>> headers, String? bodyBase64, bool bodyTruncated, String? contentType
});




}
/// @nodoc
class _$HTTPResponseSnapshotCopyWithImpl<$Res>
    implements $HTTPResponseSnapshotCopyWith<$Res> {
  _$HTTPResponseSnapshotCopyWithImpl(this._self, this._then);

  final HTTPResponseSnapshot _self;
  final $Res Function(HTTPResponseSnapshot) _then;

/// Create a copy of HTTPResponseSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? status = null,Object? headers = null,Object? bodyBase64 = freezed,Object? bodyTruncated = null,Object? contentType = freezed,}) {
  return _then(_self.copyWith(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int,headers: null == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,bodyTruncated: null == bodyTruncated ? _self.bodyTruncated : bodyTruncated // ignore: cast_nullable_to_non_nullable
as bool,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [HTTPResponseSnapshot].
extension HTTPResponseSnapshotPatterns on HTTPResponseSnapshot {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HTTPResponseSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HTTPResponseSnapshot() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HTTPResponseSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _HTTPResponseSnapshot():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HTTPResponseSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _HTTPResponseSnapshot() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int status,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HTTPResponseSnapshot() when $default != null:
return $default(_that.status,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int status,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)  $default,) {final _that = this;
switch (_that) {
case _HTTPResponseSnapshot():
return $default(_that.status,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int status,  Map<String, List<String>> headers,  String? bodyBase64,  bool bodyTruncated,  String? contentType)?  $default,) {final _that = this;
switch (_that) {
case _HTTPResponseSnapshot() when $default != null:
return $default(_that.status,_that.headers,_that.bodyBase64,_that.bodyTruncated,_that.contentType);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HTTPResponseSnapshot extends HTTPResponseSnapshot {
  const _HTTPResponseSnapshot({this.status = 0, final  Map<String, List<String>> headers = const {}, this.bodyBase64, this.bodyTruncated = false, this.contentType}): _headers = headers,super._();
  factory _HTTPResponseSnapshot.fromJson(Map<String, dynamic> json) => _$HTTPResponseSnapshotFromJson(json);

@override@JsonKey() final  int status;
 final  Map<String, List<String>> _headers;
@override@JsonKey() Map<String, List<String>> get headers {
  if (_headers is EqualUnmodifiableMapView) return _headers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(_headers);
}

@override final  String? bodyBase64;
@override@JsonKey() final  bool bodyTruncated;
@override final  String? contentType;

/// Create a copy of HTTPResponseSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HTTPResponseSnapshotCopyWith<_HTTPResponseSnapshot> get copyWith => __$HTTPResponseSnapshotCopyWithImpl<_HTTPResponseSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HTTPResponseSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _HTTPResponseSnapshot&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.bodyBase64, bodyBase64) || other.bodyBase64 == bodyBase64)&&(identical(other.bodyTruncated, bodyTruncated) || other.bodyTruncated == bodyTruncated)&&(identical(other.contentType, contentType) || other.contentType == contentType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,status,const DeepCollectionEquality().hash(_headers),bodyBase64,bodyTruncated,contentType);

@override
String toString() {
  return 'HTTPResponseSnapshot(status: $status, headers: $headers, bodyBase64: $bodyBase64, bodyTruncated: $bodyTruncated, contentType: $contentType)';
}


}

/// @nodoc
abstract mixin class _$HTTPResponseSnapshotCopyWith<$Res> implements $HTTPResponseSnapshotCopyWith<$Res> {
  factory _$HTTPResponseSnapshotCopyWith(_HTTPResponseSnapshot value, $Res Function(_HTTPResponseSnapshot) _then) = __$HTTPResponseSnapshotCopyWithImpl;
@override @useResult
$Res call({
 int status, Map<String, List<String>> headers, String? bodyBase64, bool bodyTruncated, String? contentType
});




}
/// @nodoc
class __$HTTPResponseSnapshotCopyWithImpl<$Res>
    implements _$HTTPResponseSnapshotCopyWith<$Res> {
  __$HTTPResponseSnapshotCopyWithImpl(this._self, this._then);

  final _HTTPResponseSnapshot _self;
  final $Res Function(_HTTPResponseSnapshot) _then;

/// Create a copy of HTTPResponseSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? status = null,Object? headers = null,Object? bodyBase64 = freezed,Object? bodyTruncated = null,Object? contentType = freezed,}) {
  return _then(_HTTPResponseSnapshot(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int,headers: null == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,bodyBase64: freezed == bodyBase64 ? _self.bodyBase64 : bodyBase64 // ignore: cast_nullable_to_non_nullable
as String?,bodyTruncated: null == bodyTruncated ? _self.bodyTruncated : bodyTruncated // ignore: cast_nullable_to_non_nullable
as bool,contentType: freezed == contentType ? _self.contentType : contentType // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}


/// @nodoc
mixin _$InterceptItem {

 String get id; DateTime get createdAt; DateTime get deadline; String get direction; String get sessionId; String get state; String? get ruleId; HTTPRequestSnapshot? get req; HTTPResponseSnapshot? get res;
/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$InterceptItemCopyWith<InterceptItem> get copyWith => _$InterceptItemCopyWithImpl<InterceptItem>(this as InterceptItem, _$identity);

  /// Serializes this InterceptItem to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is InterceptItem&&(identical(other.id, id) || other.id == id)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.deadline, deadline) || other.deadline == deadline)&&(identical(other.direction, direction) || other.direction == direction)&&(identical(other.sessionId, sessionId) || other.sessionId == sessionId)&&(identical(other.state, state) || other.state == state)&&(identical(other.ruleId, ruleId) || other.ruleId == ruleId)&&(identical(other.req, req) || other.req == req)&&(identical(other.res, res) || other.res == res));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,createdAt,deadline,direction,sessionId,state,ruleId,req,res);

@override
String toString() {
  return 'InterceptItem(id: $id, createdAt: $createdAt, deadline: $deadline, direction: $direction, sessionId: $sessionId, state: $state, ruleId: $ruleId, req: $req, res: $res)';
}


}

/// @nodoc
abstract mixin class $InterceptItemCopyWith<$Res>  {
  factory $InterceptItemCopyWith(InterceptItem value, $Res Function(InterceptItem) _then) = _$InterceptItemCopyWithImpl;
@useResult
$Res call({
 String id, DateTime createdAt, DateTime deadline, String direction, String sessionId, String state, String? ruleId, HTTPRequestSnapshot? req, HTTPResponseSnapshot? res
});


$HTTPRequestSnapshotCopyWith<$Res>? get req;$HTTPResponseSnapshotCopyWith<$Res>? get res;

}
/// @nodoc
class _$InterceptItemCopyWithImpl<$Res>
    implements $InterceptItemCopyWith<$Res> {
  _$InterceptItemCopyWithImpl(this._self, this._then);

  final InterceptItem _self;
  final $Res Function(InterceptItem) _then;

/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? createdAt = null,Object? deadline = null,Object? direction = null,Object? sessionId = null,Object? state = null,Object? ruleId = freezed,Object? req = freezed,Object? res = freezed,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,deadline: null == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as DateTime,direction: null == direction ? _self.direction : direction // ignore: cast_nullable_to_non_nullable
as String,sessionId: null == sessionId ? _self.sessionId : sessionId // ignore: cast_nullable_to_non_nullable
as String,state: null == state ? _self.state : state // ignore: cast_nullable_to_non_nullable
as String,ruleId: freezed == ruleId ? _self.ruleId : ruleId // ignore: cast_nullable_to_non_nullable
as String?,req: freezed == req ? _self.req : req // ignore: cast_nullable_to_non_nullable
as HTTPRequestSnapshot?,res: freezed == res ? _self.res : res // ignore: cast_nullable_to_non_nullable
as HTTPResponseSnapshot?,
  ));
}
/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HTTPRequestSnapshotCopyWith<$Res>? get req {
    if (_self.req == null) {
    return null;
  }

  return $HTTPRequestSnapshotCopyWith<$Res>(_self.req!, (value) {
    return _then(_self.copyWith(req: value));
  });
}/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HTTPResponseSnapshotCopyWith<$Res>? get res {
    if (_self.res == null) {
    return null;
  }

  return $HTTPResponseSnapshotCopyWith<$Res>(_self.res!, (value) {
    return _then(_self.copyWith(res: value));
  });
}
}


/// Adds pattern-matching-related methods to [InterceptItem].
extension InterceptItemPatterns on InterceptItem {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _InterceptItem value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _InterceptItem() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _InterceptItem value)  $default,){
final _that = this;
switch (_that) {
case _InterceptItem():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _InterceptItem value)?  $default,){
final _that = this;
switch (_that) {
case _InterceptItem() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  DateTime createdAt,  DateTime deadline,  String direction,  String sessionId,  String state,  String? ruleId,  HTTPRequestSnapshot? req,  HTTPResponseSnapshot? res)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _InterceptItem() when $default != null:
return $default(_that.id,_that.createdAt,_that.deadline,_that.direction,_that.sessionId,_that.state,_that.ruleId,_that.req,_that.res);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  DateTime createdAt,  DateTime deadline,  String direction,  String sessionId,  String state,  String? ruleId,  HTTPRequestSnapshot? req,  HTTPResponseSnapshot? res)  $default,) {final _that = this;
switch (_that) {
case _InterceptItem():
return $default(_that.id,_that.createdAt,_that.deadline,_that.direction,_that.sessionId,_that.state,_that.ruleId,_that.req,_that.res);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  DateTime createdAt,  DateTime deadline,  String direction,  String sessionId,  String state,  String? ruleId,  HTTPRequestSnapshot? req,  HTTPResponseSnapshot? res)?  $default,) {final _that = this;
switch (_that) {
case _InterceptItem() when $default != null:
return $default(_that.id,_that.createdAt,_that.deadline,_that.direction,_that.sessionId,_that.state,_that.ruleId,_that.req,_that.res);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _InterceptItem extends InterceptItem {
  const _InterceptItem({required this.id, required this.createdAt, required this.deadline, required this.direction, required this.sessionId, this.state = 'PENDING', this.ruleId, this.req, this.res}): super._();
  factory _InterceptItem.fromJson(Map<String, dynamic> json) => _$InterceptItemFromJson(json);

@override final  String id;
@override final  DateTime createdAt;
@override final  DateTime deadline;
@override final  String direction;
@override final  String sessionId;
@override@JsonKey() final  String state;
@override final  String? ruleId;
@override final  HTTPRequestSnapshot? req;
@override final  HTTPResponseSnapshot? res;

/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$InterceptItemCopyWith<_InterceptItem> get copyWith => __$InterceptItemCopyWithImpl<_InterceptItem>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$InterceptItemToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _InterceptItem&&(identical(other.id, id) || other.id == id)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.deadline, deadline) || other.deadline == deadline)&&(identical(other.direction, direction) || other.direction == direction)&&(identical(other.sessionId, sessionId) || other.sessionId == sessionId)&&(identical(other.state, state) || other.state == state)&&(identical(other.ruleId, ruleId) || other.ruleId == ruleId)&&(identical(other.req, req) || other.req == req)&&(identical(other.res, res) || other.res == res));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,createdAt,deadline,direction,sessionId,state,ruleId,req,res);

@override
String toString() {
  return 'InterceptItem(id: $id, createdAt: $createdAt, deadline: $deadline, direction: $direction, sessionId: $sessionId, state: $state, ruleId: $ruleId, req: $req, res: $res)';
}


}

/// @nodoc
abstract mixin class _$InterceptItemCopyWith<$Res> implements $InterceptItemCopyWith<$Res> {
  factory _$InterceptItemCopyWith(_InterceptItem value, $Res Function(_InterceptItem) _then) = __$InterceptItemCopyWithImpl;
@override @useResult
$Res call({
 String id, DateTime createdAt, DateTime deadline, String direction, String sessionId, String state, String? ruleId, HTTPRequestSnapshot? req, HTTPResponseSnapshot? res
});


@override $HTTPRequestSnapshotCopyWith<$Res>? get req;@override $HTTPResponseSnapshotCopyWith<$Res>? get res;

}
/// @nodoc
class __$InterceptItemCopyWithImpl<$Res>
    implements _$InterceptItemCopyWith<$Res> {
  __$InterceptItemCopyWithImpl(this._self, this._then);

  final _InterceptItem _self;
  final $Res Function(_InterceptItem) _then;

/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? createdAt = null,Object? deadline = null,Object? direction = null,Object? sessionId = null,Object? state = null,Object? ruleId = freezed,Object? req = freezed,Object? res = freezed,}) {
  return _then(_InterceptItem(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,deadline: null == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as DateTime,direction: null == direction ? _self.direction : direction // ignore: cast_nullable_to_non_nullable
as String,sessionId: null == sessionId ? _self.sessionId : sessionId // ignore: cast_nullable_to_non_nullable
as String,state: null == state ? _self.state : state // ignore: cast_nullable_to_non_nullable
as String,ruleId: freezed == ruleId ? _self.ruleId : ruleId // ignore: cast_nullable_to_non_nullable
as String?,req: freezed == req ? _self.req : req // ignore: cast_nullable_to_non_nullable
as HTTPRequestSnapshot?,res: freezed == res ? _self.res : res // ignore: cast_nullable_to_non_nullable
as HTTPResponseSnapshot?,
  ));
}

/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HTTPRequestSnapshotCopyWith<$Res>? get req {
    if (_self.req == null) {
    return null;
  }

  return $HTTPRequestSnapshotCopyWith<$Res>(_self.req!, (value) {
    return _then(_self.copyWith(req: value));
  });
}/// Create a copy of InterceptItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HTTPResponseSnapshotCopyWith<$Res>? get res {
    if (_self.res == null) {
    return null;
  }

  return $HTTPResponseSnapshotCopyWith<$Res>(_self.res!, (value) {
    return _then(_self.copyWith(res: value));
  });
}
}

// dart format on
