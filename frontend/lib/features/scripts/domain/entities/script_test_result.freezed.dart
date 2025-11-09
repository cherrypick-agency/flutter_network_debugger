// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'script_test_result.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TestRequest {

 String get method; String get url; Map<String, List<String>> get headers; String? get body;
/// Create a copy of TestRequest
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TestRequestCopyWith<TestRequest> get copyWith => _$TestRequestCopyWithImpl<TestRequest>(this as TestRequest, _$identity);

  /// Serializes this TestRequest to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TestRequest&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.body, body) || other.body == body));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(headers),body);

@override
String toString() {
  return 'TestRequest(method: $method, url: $url, headers: $headers, body: $body)';
}


}

/// @nodoc
abstract mixin class $TestRequestCopyWith<$Res>  {
  factory $TestRequestCopyWith(TestRequest value, $Res Function(TestRequest) _then) = _$TestRequestCopyWithImpl;
@useResult
$Res call({
 String method, String url, Map<String, List<String>> headers, String? body
});




}
/// @nodoc
class _$TestRequestCopyWithImpl<$Res>
    implements $TestRequestCopyWith<$Res> {
  _$TestRequestCopyWithImpl(this._self, this._then);

  final TestRequest _self;
  final $Res Function(TestRequest) _then;

/// Create a copy of TestRequest
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? method = null,Object? url = null,Object? headers = null,Object? body = freezed,}) {
  return _then(_self.copyWith(
method: null == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String,url: null == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String,headers: null == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,body: freezed == body ? _self.body : body // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [TestRequest].
extension TestRequestPatterns on TestRequest {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TestRequest value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TestRequest() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TestRequest value)  $default,){
final _that = this;
switch (_that) {
case _TestRequest():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TestRequest value)?  $default,){
final _that = this;
switch (_that) {
case _TestRequest() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String method,  String url,  Map<String, List<String>> headers,  String? body)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TestRequest() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.body);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String method,  String url,  Map<String, List<String>> headers,  String? body)  $default,) {final _that = this;
switch (_that) {
case _TestRequest():
return $default(_that.method,_that.url,_that.headers,_that.body);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String method,  String url,  Map<String, List<String>> headers,  String? body)?  $default,) {final _that = this;
switch (_that) {
case _TestRequest() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.body);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TestRequest extends TestRequest {
  const _TestRequest({required this.method, required this.url, final  Map<String, List<String>> headers = const {}, this.body}): _headers = headers,super._();
  factory _TestRequest.fromJson(Map<String, dynamic> json) => _$TestRequestFromJson(json);

@override final  String method;
@override final  String url;
 final  Map<String, List<String>> _headers;
@override@JsonKey() Map<String, List<String>> get headers {
  if (_headers is EqualUnmodifiableMapView) return _headers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(_headers);
}

@override final  String? body;

/// Create a copy of TestRequest
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TestRequestCopyWith<_TestRequest> get copyWith => __$TestRequestCopyWithImpl<_TestRequest>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TestRequestToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _TestRequest&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.body, body) || other.body == body));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(_headers),body);

@override
String toString() {
  return 'TestRequest(method: $method, url: $url, headers: $headers, body: $body)';
}


}

/// @nodoc
abstract mixin class _$TestRequestCopyWith<$Res> implements $TestRequestCopyWith<$Res> {
  factory _$TestRequestCopyWith(_TestRequest value, $Res Function(_TestRequest) _then) = __$TestRequestCopyWithImpl;
@override @useResult
$Res call({
 String method, String url, Map<String, List<String>> headers, String? body
});




}
/// @nodoc
class __$TestRequestCopyWithImpl<$Res>
    implements _$TestRequestCopyWith<$Res> {
  __$TestRequestCopyWithImpl(this._self, this._then);

  final _TestRequest _self;
  final $Res Function(_TestRequest) _then;

/// Create a copy of TestRequest
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? method = null,Object? url = null,Object? headers = null,Object? body = freezed,}) {
  return _then(_TestRequest(
method: null == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String,url: null == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String,headers: null == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>,body: freezed == body ? _self.body : body // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}


/// @nodoc
mixin _$ModifiedHTTP {

 String? get method; String? get url; Map<String, List<String>>? get headers; String? get body; int? get status;
/// Create a copy of ModifiedHTTP
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ModifiedHTTPCopyWith<ModifiedHTTP> get copyWith => _$ModifiedHTTPCopyWithImpl<ModifiedHTTP>(this as ModifiedHTTP, _$identity);

  /// Serializes this ModifiedHTTP to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ModifiedHTTP&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other.headers, headers)&&(identical(other.body, body) || other.body == body)&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(headers),body,status);

@override
String toString() {
  return 'ModifiedHTTP(method: $method, url: $url, headers: $headers, body: $body, status: $status)';
}


}

/// @nodoc
abstract mixin class $ModifiedHTTPCopyWith<$Res>  {
  factory $ModifiedHTTPCopyWith(ModifiedHTTP value, $Res Function(ModifiedHTTP) _then) = _$ModifiedHTTPCopyWithImpl;
@useResult
$Res call({
 String? method, String? url, Map<String, List<String>>? headers, String? body, int? status
});




}
/// @nodoc
class _$ModifiedHTTPCopyWithImpl<$Res>
    implements $ModifiedHTTPCopyWith<$Res> {
  _$ModifiedHTTPCopyWithImpl(this._self, this._then);

  final ModifiedHTTP _self;
  final $Res Function(ModifiedHTTP) _then;

/// Create a copy of ModifiedHTTP
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? method = freezed,Object? url = freezed,Object? headers = freezed,Object? body = freezed,Object? status = freezed,}) {
  return _then(_self.copyWith(
method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,url: freezed == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String?,headers: freezed == headers ? _self.headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,body: freezed == body ? _self.body : body // ignore: cast_nullable_to_non_nullable
as String?,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

}


/// Adds pattern-matching-related methods to [ModifiedHTTP].
extension ModifiedHTTPPatterns on ModifiedHTTP {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ModifiedHTTP value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ModifiedHTTP() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ModifiedHTTP value)  $default,){
final _that = this;
switch (_that) {
case _ModifiedHTTP():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ModifiedHTTP value)?  $default,){
final _that = this;
switch (_that) {
case _ModifiedHTTP() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String? method,  String? url,  Map<String, List<String>>? headers,  String? body,  int? status)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ModifiedHTTP() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.body,_that.status);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String? method,  String? url,  Map<String, List<String>>? headers,  String? body,  int? status)  $default,) {final _that = this;
switch (_that) {
case _ModifiedHTTP():
return $default(_that.method,_that.url,_that.headers,_that.body,_that.status);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String? method,  String? url,  Map<String, List<String>>? headers,  String? body,  int? status)?  $default,) {final _that = this;
switch (_that) {
case _ModifiedHTTP() when $default != null:
return $default(_that.method,_that.url,_that.headers,_that.body,_that.status);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ModifiedHTTP extends ModifiedHTTP {
  const _ModifiedHTTP({this.method, this.url, final  Map<String, List<String>>? headers, this.body, this.status}): _headers = headers,super._();
  factory _ModifiedHTTP.fromJson(Map<String, dynamic> json) => _$ModifiedHTTPFromJson(json);

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

@override final  String? body;
@override final  int? status;

/// Create a copy of ModifiedHTTP
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ModifiedHTTPCopyWith<_ModifiedHTTP> get copyWith => __$ModifiedHTTPCopyWithImpl<_ModifiedHTTP>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ModifiedHTTPToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ModifiedHTTP&&(identical(other.method, method) || other.method == method)&&(identical(other.url, url) || other.url == url)&&const DeepCollectionEquality().equals(other._headers, _headers)&&(identical(other.body, body) || other.body == body)&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,method,url,const DeepCollectionEquality().hash(_headers),body,status);

@override
String toString() {
  return 'ModifiedHTTP(method: $method, url: $url, headers: $headers, body: $body, status: $status)';
}


}

/// @nodoc
abstract mixin class _$ModifiedHTTPCopyWith<$Res> implements $ModifiedHTTPCopyWith<$Res> {
  factory _$ModifiedHTTPCopyWith(_ModifiedHTTP value, $Res Function(_ModifiedHTTP) _then) = __$ModifiedHTTPCopyWithImpl;
@override @useResult
$Res call({
 String? method, String? url, Map<String, List<String>>? headers, String? body, int? status
});




}
/// @nodoc
class __$ModifiedHTTPCopyWithImpl<$Res>
    implements _$ModifiedHTTPCopyWith<$Res> {
  __$ModifiedHTTPCopyWithImpl(this._self, this._then);

  final _ModifiedHTTP _self;
  final $Res Function(_ModifiedHTTP) _then;

/// Create a copy of ModifiedHTTP
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? method = freezed,Object? url = freezed,Object? headers = freezed,Object? body = freezed,Object? status = freezed,}) {
  return _then(_ModifiedHTTP(
method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,url: freezed == url ? _self.url : url // ignore: cast_nullable_to_non_nullable
as String?,headers: freezed == headers ? _self._headers : headers // ignore: cast_nullable_to_non_nullable
as Map<String, List<String>>?,body: freezed == body ? _self.body : body // ignore: cast_nullable_to_non_nullable
as String?,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}


}


/// @nodoc
mixin _$ScriptTestResult {

 bool get success; String? get error; ModifiedHTTP? get modifiedRequest; ModifiedHTTP? get modifiedResponse; List<String> get logs; int? get durationMs;
/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ScriptTestResultCopyWith<ScriptTestResult> get copyWith => _$ScriptTestResultCopyWithImpl<ScriptTestResult>(this as ScriptTestResult, _$identity);

  /// Serializes this ScriptTestResult to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ScriptTestResult&&(identical(other.success, success) || other.success == success)&&(identical(other.error, error) || other.error == error)&&(identical(other.modifiedRequest, modifiedRequest) || other.modifiedRequest == modifiedRequest)&&(identical(other.modifiedResponse, modifiedResponse) || other.modifiedResponse == modifiedResponse)&&const DeepCollectionEquality().equals(other.logs, logs)&&(identical(other.durationMs, durationMs) || other.durationMs == durationMs));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,success,error,modifiedRequest,modifiedResponse,const DeepCollectionEquality().hash(logs),durationMs);

@override
String toString() {
  return 'ScriptTestResult(success: $success, error: $error, modifiedRequest: $modifiedRequest, modifiedResponse: $modifiedResponse, logs: $logs, durationMs: $durationMs)';
}


}

/// @nodoc
abstract mixin class $ScriptTestResultCopyWith<$Res>  {
  factory $ScriptTestResultCopyWith(ScriptTestResult value, $Res Function(ScriptTestResult) _then) = _$ScriptTestResultCopyWithImpl;
@useResult
$Res call({
 bool success, String? error, ModifiedHTTP? modifiedRequest, ModifiedHTTP? modifiedResponse, List<String> logs, int? durationMs
});


$ModifiedHTTPCopyWith<$Res>? get modifiedRequest;$ModifiedHTTPCopyWith<$Res>? get modifiedResponse;

}
/// @nodoc
class _$ScriptTestResultCopyWithImpl<$Res>
    implements $ScriptTestResultCopyWith<$Res> {
  _$ScriptTestResultCopyWithImpl(this._self, this._then);

  final ScriptTestResult _self;
  final $Res Function(ScriptTestResult) _then;

/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? success = null,Object? error = freezed,Object? modifiedRequest = freezed,Object? modifiedResponse = freezed,Object? logs = null,Object? durationMs = freezed,}) {
  return _then(_self.copyWith(
success: null == success ? _self.success : success // ignore: cast_nullable_to_non_nullable
as bool,error: freezed == error ? _self.error : error // ignore: cast_nullable_to_non_nullable
as String?,modifiedRequest: freezed == modifiedRequest ? _self.modifiedRequest : modifiedRequest // ignore: cast_nullable_to_non_nullable
as ModifiedHTTP?,modifiedResponse: freezed == modifiedResponse ? _self.modifiedResponse : modifiedResponse // ignore: cast_nullable_to_non_nullable
as ModifiedHTTP?,logs: null == logs ? _self.logs : logs // ignore: cast_nullable_to_non_nullable
as List<String>,durationMs: freezed == durationMs ? _self.durationMs : durationMs // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}
/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ModifiedHTTPCopyWith<$Res>? get modifiedRequest {
    if (_self.modifiedRequest == null) {
    return null;
  }

  return $ModifiedHTTPCopyWith<$Res>(_self.modifiedRequest!, (value) {
    return _then(_self.copyWith(modifiedRequest: value));
  });
}/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ModifiedHTTPCopyWith<$Res>? get modifiedResponse {
    if (_self.modifiedResponse == null) {
    return null;
  }

  return $ModifiedHTTPCopyWith<$Res>(_self.modifiedResponse!, (value) {
    return _then(_self.copyWith(modifiedResponse: value));
  });
}
}


/// Adds pattern-matching-related methods to [ScriptTestResult].
extension ScriptTestResultPatterns on ScriptTestResult {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ScriptTestResult value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ScriptTestResult() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ScriptTestResult value)  $default,){
final _that = this;
switch (_that) {
case _ScriptTestResult():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ScriptTestResult value)?  $default,){
final _that = this;
switch (_that) {
case _ScriptTestResult() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( bool success,  String? error,  ModifiedHTTP? modifiedRequest,  ModifiedHTTP? modifiedResponse,  List<String> logs,  int? durationMs)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ScriptTestResult() when $default != null:
return $default(_that.success,_that.error,_that.modifiedRequest,_that.modifiedResponse,_that.logs,_that.durationMs);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( bool success,  String? error,  ModifiedHTTP? modifiedRequest,  ModifiedHTTP? modifiedResponse,  List<String> logs,  int? durationMs)  $default,) {final _that = this;
switch (_that) {
case _ScriptTestResult():
return $default(_that.success,_that.error,_that.modifiedRequest,_that.modifiedResponse,_that.logs,_that.durationMs);}
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( bool success,  String? error,  ModifiedHTTP? modifiedRequest,  ModifiedHTTP? modifiedResponse,  List<String> logs,  int? durationMs)?  $default,) {final _that = this;
switch (_that) {
case _ScriptTestResult() when $default != null:
return $default(_that.success,_that.error,_that.modifiedRequest,_that.modifiedResponse,_that.logs,_that.durationMs);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ScriptTestResult extends ScriptTestResult {
  const _ScriptTestResult({required this.success, this.error, this.modifiedRequest, this.modifiedResponse, final  List<String> logs = const [], this.durationMs}): _logs = logs,super._();
  factory _ScriptTestResult.fromJson(Map<String, dynamic> json) => _$ScriptTestResultFromJson(json);

@override final  bool success;
@override final  String? error;
@override final  ModifiedHTTP? modifiedRequest;
@override final  ModifiedHTTP? modifiedResponse;
 final  List<String> _logs;
@override@JsonKey() List<String> get logs {
  if (_logs is EqualUnmodifiableListView) return _logs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_logs);
}

@override final  int? durationMs;

/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ScriptTestResultCopyWith<_ScriptTestResult> get copyWith => __$ScriptTestResultCopyWithImpl<_ScriptTestResult>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ScriptTestResultToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _ScriptTestResult&&(identical(other.success, success) || other.success == success)&&(identical(other.error, error) || other.error == error)&&(identical(other.modifiedRequest, modifiedRequest) || other.modifiedRequest == modifiedRequest)&&(identical(other.modifiedResponse, modifiedResponse) || other.modifiedResponse == modifiedResponse)&&const DeepCollectionEquality().equals(other._logs, _logs)&&(identical(other.durationMs, durationMs) || other.durationMs == durationMs));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,success,error,modifiedRequest,modifiedResponse,const DeepCollectionEquality().hash(_logs),durationMs);

@override
String toString() {
  return 'ScriptTestResult(success: $success, error: $error, modifiedRequest: $modifiedRequest, modifiedResponse: $modifiedResponse, logs: $logs, durationMs: $durationMs)';
}


}

/// @nodoc
abstract mixin class _$ScriptTestResultCopyWith<$Res> implements $ScriptTestResultCopyWith<$Res> {
  factory _$ScriptTestResultCopyWith(_ScriptTestResult value, $Res Function(_ScriptTestResult) _then) = __$ScriptTestResultCopyWithImpl;
@override @useResult
$Res call({
 bool success, String? error, ModifiedHTTP? modifiedRequest, ModifiedHTTP? modifiedResponse, List<String> logs, int? durationMs
});


@override $ModifiedHTTPCopyWith<$Res>? get modifiedRequest;@override $ModifiedHTTPCopyWith<$Res>? get modifiedResponse;

}
/// @nodoc
class __$ScriptTestResultCopyWithImpl<$Res>
    implements _$ScriptTestResultCopyWith<$Res> {
  __$ScriptTestResultCopyWithImpl(this._self, this._then);

  final _ScriptTestResult _self;
  final $Res Function(_ScriptTestResult) _then;

/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? success = null,Object? error = freezed,Object? modifiedRequest = freezed,Object? modifiedResponse = freezed,Object? logs = null,Object? durationMs = freezed,}) {
  return _then(_ScriptTestResult(
success: null == success ? _self.success : success // ignore: cast_nullable_to_non_nullable
as bool,error: freezed == error ? _self.error : error // ignore: cast_nullable_to_non_nullable
as String?,modifiedRequest: freezed == modifiedRequest ? _self.modifiedRequest : modifiedRequest // ignore: cast_nullable_to_non_nullable
as ModifiedHTTP?,modifiedResponse: freezed == modifiedResponse ? _self.modifiedResponse : modifiedResponse // ignore: cast_nullable_to_non_nullable
as ModifiedHTTP?,logs: null == logs ? _self._logs : logs // ignore: cast_nullable_to_non_nullable
as List<String>,durationMs: freezed == durationMs ? _self.durationMs : durationMs // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ModifiedHTTPCopyWith<$Res>? get modifiedRequest {
    if (_self.modifiedRequest == null) {
    return null;
  }

  return $ModifiedHTTPCopyWith<$Res>(_self.modifiedRequest!, (value) {
    return _then(_self.copyWith(modifiedRequest: value));
  });
}/// Create a copy of ScriptTestResult
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ModifiedHTTPCopyWith<$Res>? get modifiedResponse {
    if (_self.modifiedResponse == null) {
    return null;
  }

  return $ModifiedHTTPCopyWith<$Res>(_self.modifiedResponse!, (value) {
    return _then(_self.copyWith(modifiedResponse: value));
  });
}
}

// dart format on
