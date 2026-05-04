import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

const String defaultLoginApiBaseUrl = String.fromEnvironment(
  'LOGIN_API_BASE_URL',
  defaultValue: 'http://127.0.0.1:8081',
);
const String defaultLoginApiPath = String.fromEnvironment(
  'LOGIN_API_PATH',
  defaultValue: '/api/v1/auth/login',
);
const String defaultLoginInstitutionsPath = String.fromEnvironment(
  'LOGIN_INSTITUTIONS_PATH',
  defaultValue: '/api/v1/auth/login-institutions',
);
const String defaultLoginTenantDomain = String.fromEnvironment(
  'LOGIN_TENANT_DOMAIN',
  defaultValue: '',
);
const String defaultQrLoginUrl = String.fromEnvironment(
  'LOGIN_QR_URL',
  defaultValue: '',
);

class AuthException implements Exception {
  const AuthException(this.message);

  final String message;

  @override
  String toString() => message;
}

class LoginResult {
  const LoginResult({
    required this.token,
    required this.loginType,
    required this.raw,
    this.tenantId = '',
    this.orgId,
  });

  factory LoginResult.fromJson(Map<String, dynamic> json) {
    return LoginResult(
      token: '${json['token'] ?? ''}',
      loginType: '${json['loginType'] ?? ''}',
      tenantId: '${json['tenantId'] ?? ''}',
      orgId: _intOrNull(json['orgId']),
      raw: json,
    );
  }

  final String token;
  final String loginType;
  final String tenantId;
  final int? orgId;
  final Map<String, dynamic> raw;
}

class InstitutionLoginOption {
  const InstitutionLoginOption({
    required this.userId,
    required this.instId,
    required this.orgName,
    required this.loginName,
    required this.nickName,
    required this.mobile,
    required this.admin,
    required this.institutionReadonly,
    this.logo = '',
    this.institutionStatus = '',
  });

  factory InstitutionLoginOption.fromJson(Map<String, dynamic> json) {
    return InstitutionLoginOption(
      userId: _intOrNull(json['userId']) ?? 0,
      instId: _intOrNull(json['instId']) ?? 0,
      orgName: '${json['orgName'] ?? ''}',
      loginName: '${json['loginName'] ?? ''}',
      nickName: '${json['nickName'] ?? ''}',
      mobile: '${json['mobile'] ?? ''}',
      logo: '${json['logo'] ?? ''}',
      admin: json['admin'] == true,
      institutionStatus: '${json['institutionStatus'] ?? ''}',
      institutionReadonly: json['institutionReadonly'] == true,
    );
  }

  final int userId;
  final int instId;
  final String orgName;
  final String loginName;
  final String nickName;
  final String mobile;
  final String logo;
  final bool admin;
  final String institutionStatus;
  final bool institutionReadonly;
}

abstract interface class AuthClient {
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  });

  Future<LoginResult> login({
    required String username,
    required String password,
    InstitutionLoginOption? institution,
  });

  Uri buildQrLoginUri(String nonce);
}

class IamAuthClient implements AuthClient {
  const IamAuthClient({
    this.baseUrl = defaultLoginApiBaseUrl,
    this.loginPath = defaultLoginApiPath,
    this.institutionsPath = defaultLoginInstitutionsPath,
    this.tenantDomain = defaultLoginTenantDomain,
    this.qrLoginUrl = defaultQrLoginUrl,
  });

  final String baseUrl;
  final String loginPath;
  final String institutionsPath;
  final String tenantDomain;
  final String qrLoginUrl;

  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
    final Map<String, Object> payload = <String, Object>{
      'identifier': identifier,
      'loginType': 2,
    };
    if (password.trim().isNotEmpty) {
      payload['password'] = password;
    }
    final Object? data = await _postJson(institutionsPath, payload);
    if (data is! List) {
      return <InstitutionLoginOption>[];
    }
    return data
        .whereType<Map>()
        .map((Map item) => InstitutionLoginOption.fromJson(
              Map<String, dynamic>.from(item),
            ))
        .toList();
  }

  @override
  Future<LoginResult> login({
    required String username,
    required String password,
    InstitutionLoginOption? institution,
  }) async {
    final Map<String, Object> payload = <String, Object>{
      'username': username,
      'password': password,
      'loginType': 2,
      'type': 'account',
    };
    if (institution != null) {
      payload['institutionId'] = institution.instId;
      payload['userId'] = institution.userId;
    }

    final Object? data = await _postJson(loginPath, payload);
    if (data is! Map) {
      throw const AuthException('登录接口返回格式不正确');
    }
    final LoginResult result = LoginResult.fromJson(
      Map<String, dynamic>.from(data),
    );
    if (result.token.trim().isEmpty) {
      throw const AuthException('登录接口未返回有效 token');
    }
    return result;
  }

  @override
  Uri buildQrLoginUri(String nonce) {
    final Uri base = Uri.parse(
      qrLoginUrl.trim().isNotEmpty
          ? qrLoginUrl.trim()
          : _uri('/institution/').toString(),
    );
    return base.replace(
      queryParameters: <String, String>{
        ...base.queryParameters,
        'qrLoginToken': nonce,
        'source': 'assessment-pad',
      },
    );
  }

  Future<Object?> _postJson(String path, Map<String, Object> payload) async {
    final http.Response response;
    try {
      response = await http
          .post(
            _uri(path),
            headers: <String, String>{
              'Accept': 'application/json',
              'Content-Type': 'application/json; charset=utf-8',
              if (tenantDomain.trim().isNotEmpty)
                'X-Tenant-Domain': tenantDomain.trim(),
            },
            body: jsonEncode(payload),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const AuthException('登录接口响应超时，请检查网络');
    } on Object catch (error) {
      throw AuthException('无法连接登录接口：$error');
    }

    final Object? decoded = _decodeResponse(response.body);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw AuthException(_messageFromPayload(decoded) ?? '登录失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw AuthException(_messageFromPayload(envelope) ?? '登录失败');
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }

  Uri _uri(String path) {
    final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
    final String normalizedPath = path.startsWith('/') ? path : '/$path';
    return Uri.parse('$trimmedBase$normalizedPath');
  }
}

Object? _decodeResponse(String body) {
  if (body.trim().isEmpty) {
    return null;
  }
  try {
    return jsonDecode(body);
  } on FormatException {
    return body;
  }
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  if (payload is String && payload.trim().isNotEmpty) {
    return payload;
  }
  return null;
}

int? _intOrNull(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse('$value');
}
