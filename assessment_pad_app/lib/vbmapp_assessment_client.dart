import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultVbmappDraftSavePath = String.fromEnvironment(
  'VBMAPP_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/save',
);
const String defaultVbmappDraftsPagePath = String.fromEnvironment(
  'VBMAPP_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/page',
);

abstract class VbmappAssessmentClient {
  const VbmappAssessmentClient();

  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  });

  Future<VbmappDraftSaveResult> saveDraft(
    String token,
    Map<String, dynamic> payload,
  );
}

class ApiVbmappAssessmentClient extends VbmappAssessmentClient {
  const ApiVbmappAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.draftSavePath = defaultVbmappDraftSavePath,
    this.draftsPagePath = defaultVbmappDraftsPagePath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String draftSavePath;
  final String draftsPagePath;
  final http.Client? httpClient;

  static final http.Client _sharedHttpClient = http.Client();

  http.Client get _client => httpClient ?? _sharedHttpClient;

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'VBMAPP',
          if (studentId > 0) 'studentId': studentId,
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      },
      fallbackMessage: 'VB-MAPP草稿列表加载失败',
    );
    if (data is! Map) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<VbmappDraftSaveResult> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftSavePath),
      token,
      payload,
      fallbackMessage: 'VB-MAPP草稿保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('VB-MAPP草稿保存返回格式不正确');
    }
    return VbmappDraftSaveResult.fromJson(Map<String, dynamic>.from(data));
  }

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> body, {
    required String fallbackMessage,
  }) async {
    final http.Response response;
    try {
      response = await _client
          .post(
            uri,
            headers: <String, String>{
              'Accept': 'application/json',
              'Content-Type': 'application/json',
              'Authorization': 'Bearer $token',
            },
            body: jsonEncode(body),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw AssessmentScaleApiException('$fallbackMessage，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接VB-MAPP接口：$error');
    }
    return _handleResponse(response, fallbackMessage: fallbackMessage);
  }
}

class VbmappDraftSaveResult {
  const VbmappDraftSaveResult({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.assessmentDate,
    required this.examinerName,
    required this.status,
    required this.answeredItemCount,
    required this.completionPercent,
  });

  factory VbmappDraftSaveResult.fromJson(Map<String, dynamic> json) {
    return VbmappDraftSaveResult(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentDate: _dateTextFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      status: '${json['status'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String assessmentDate;
  final String examinerName;
  final String status;
  final int answeredItemCount;
  final double completionPercent;
}

Uri _uri(String baseUrl, String path) {
  final String trimmedBase = baseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return Uri.parse('$trimmedBase$normalizedPath');
}

Object? _handleResponse(
  http.Response response, {
  required String fallbackMessage,
}) {
  final Object? decoded = _decodeResponse(response.body);
  if (response.statusCode == 401) {
    throw AssessmentScaleApiException(
      _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
      unauthorized: true,
    );
  }
  if (response.statusCode < 200 || response.statusCode >= 300) {
    throw AssessmentScaleApiException(
      _messageFromPayload(decoded) ?? fallbackMessage,
    );
  }
  if (decoded is Map) {
    final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
    if (envelope['success'] == false) {
      throw AssessmentScaleApiException(
        _messageFromPayload(envelope) ?? fallbackMessage,
      );
    }
    if (envelope.containsKey('data')) {
      return envelope['data'];
    }
  }
  return decoded;
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

String _dateTextFrom(Object? raw) {
  final String value = '${raw ?? ''}'.trim();
  if (value.length >= 10) {
    return value.substring(0, 10);
  }
  return value;
}

int _intFrom(Object? raw) {
  if (raw is int) {
    return raw;
  }
  if (raw is num) {
    return raw.toInt();
  }
  return int.tryParse('${raw ?? ''}') ?? 0;
}

double _doubleFrom(Object? raw) {
  if (raw is num) {
    return raw.toDouble();
  }
  return double.tryParse('${raw ?? ''}') ?? 0;
}
