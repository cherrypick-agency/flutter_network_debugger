import 'package:flutter/foundation.dart';

@immutable
class ComposeHeaderDTO {
  final String key;
  final String value;
  const ComposeHeaderDTO({required this.key, required this.value});
  Map<String, dynamic> toJson() => {'key': key, 'value': value};
  factory ComposeHeaderDTO.fromJson(Map<String, dynamic> json) =>
      ComposeHeaderDTO(
        key: json['key'] as String,
        value: json['value'] as String,
      );
}

@immutable
class ComposeQueryDTO {
  final String key;
  final String value;
  const ComposeQueryDTO({required this.key, required this.value});
  Map<String, dynamic> toJson() => {'key': key, 'value': value};
  factory ComposeQueryDTO.fromJson(Map<String, dynamic> json) =>
      ComposeQueryDTO(
        key: json['key'] as String,
        value: json['value'] as String,
      );
}

@immutable
class ComposeFormFieldDTO {
  final String key;
  final String value;
  const ComposeFormFieldDTO({required this.key, required this.value});
  Map<String, dynamic> toJson() => {'key': key, 'value': value};
  factory ComposeFormFieldDTO.fromJson(Map<String, dynamic> json) =>
      ComposeFormFieldDTO(
        key: json['key'] as String,
        value: json['value'] as String,
      );
}

@immutable
class ComposePartDTO {
  final String name;
  final bool isFile;
  final String? fileName;
  final String? contentType;
  final String? base64; // for file
  final String? value; // for text
  const ComposePartDTO({
    required this.name,
    required this.isFile,
    this.fileName,
    this.contentType,
    this.base64,
    this.value,
  });
  Map<String, dynamic> toJson() => {
    'name': name,
    'isFile': isFile,
    if (fileName != null) 'fileName': fileName,
    if (contentType != null) 'contentType': contentType,
    if (base64 != null) 'base64': base64,
    if (value != null) 'value': value,
  };
  factory ComposePartDTO.fromJson(Map<String, dynamic> json) => ComposePartDTO(
    name: json['name'] as String,
    isFile: (json['isFile'] as bool?) ?? false,
    fileName: json['fileName'] as String?,
    contentType: json['contentType'] as String?,
    base64: json['base64'] as String?,
    value: json['value'] as String?,
  );
}

@immutable
class ComposeAuthDTO {
  final String type; // none|basic|bearer|apiKey
  final String? username;
  final String? password;
  final String? bearerToken;
  final String? apiKey;
  final String? apiKeyHeader;
  const ComposeAuthDTO({
    required this.type,
    this.username,
    this.password,
    this.bearerToken,
    this.apiKey,
    this.apiKeyHeader,
  });
  Map<String, dynamic> toJson() => {
    'type': type,
    if (username != null) 'username': username,
    if (password != null) 'password': password,
    if (bearerToken != null) 'bearerToken': bearerToken,
    if (apiKey != null) 'apiKey': apiKey,
    if (apiKeyHeader != null) 'apiKeyHeader': apiKeyHeader,
  };
  factory ComposeAuthDTO.fromJson(Map<String, dynamic> json) => ComposeAuthDTO(
    type: json['type'] as String,
    username: json['username'] as String?,
    password: json['password'] as String?,
    bearerToken: json['bearerToken'] as String?,
    apiKey: json['apiKey'] as String?,
    apiKeyHeader: json['apiKeyHeader'] as String?,
  );
}

@immutable
class ComposeBodyDTO {
  final String mode; // raw|json|form|multipart
  final String? raw;
  final String? json;
  final List<ComposeFormFieldDTO> form;
  final List<ComposePartDTO> multipart;
  const ComposeBodyDTO({
    required this.mode,
    this.raw,
    this.json,
    this.form = const [],
    this.multipart = const [],
  });
  Map<String, dynamic> toJson() => {
    'mode': mode,
    if (raw != null) 'raw': raw,
    if (json != null) 'json': json,
    if (form.isNotEmpty) 'form': form.map((e) => e.toJson()).toList(),
    if (multipart.isNotEmpty)
      'multipart': multipart.map((e) => e.toJson()).toList(),
  };
  factory ComposeBodyDTO.fromJson(Map<String, dynamic> json) => ComposeBodyDTO(
    mode: json['mode'] as String,
    raw: json['raw'] as String?,
    json: json['json'] as String?,
    form:
        ((json['form'] as List?) ?? const [])
            .map(
              (e) => ComposeFormFieldDTO.fromJson(
                (e as Map).cast<String, dynamic>(),
              ),
            )
            .toList(),
    multipart:
        ((json['multipart'] as List?) ?? const [])
            .map(
              (e) =>
                  ComposePartDTO.fromJson((e as Map).cast<String, dynamic>()),
            )
            .toList(),
  );
}

@immutable
class ComposeTemplateDTO {
  final String id;
  final String name;
  final String method;
  final String url;
  final List<ComposeHeaderDTO> headers;
  final List<ComposeQueryDTO> query;
  final ComposeBodyDTO body;
  final ComposeAuthDTO? auth;
  const ComposeTemplateDTO({
    required this.id,
    required this.name,
    required this.method,
    required this.url,
    required this.headers,
    required this.query,
    required this.body,
    this.auth,
  });
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'method': method,
    'url': url,
    'headers': headers.map((e) => e.toJson()).toList(),
    'query': query.map((e) => e.toJson()).toList(),
    'body': body.toJson(),
    if (auth != null) 'auth': auth!.toJson(),
  };
  factory ComposeTemplateDTO.fromJson(
    Map<String, dynamic> json,
  ) => ComposeTemplateDTO(
    id: json['id'] as String,
    name: json['name'] as String,
    method: json['method'] as String,
    url: json['url'] as String,
    headers:
        ((json['headers'] as List?) ?? const [])
            .map(
              (e) =>
                  ComposeHeaderDTO.fromJson((e as Map).cast<String, dynamic>()),
            )
            .toList(),
    query:
        ((json['query'] as List?) ?? const [])
            .map(
              (e) =>
                  ComposeQueryDTO.fromJson((e as Map).cast<String, dynamic>()),
            )
            .toList(),
    body: ComposeBodyDTO.fromJson(
      (json['body'] as Map).cast<String, dynamic>(),
    ),
    auth:
        json['auth'] == null
            ? null
            : ComposeAuthDTO.fromJson(
              (json['auth'] as Map).cast<String, dynamic>(),
            ),
  );
}
