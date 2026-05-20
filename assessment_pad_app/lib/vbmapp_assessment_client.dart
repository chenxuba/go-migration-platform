import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultVbmappDraftSavePath = String.fromEnvironment(
  'VBMAPP_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/save',
);
const String defaultVbmappDraftDetailPath = String.fromEnvironment(
  'VBMAPP_DRAFT_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/detail',
);
const String defaultVbmappDraftsPagePath = String.fromEnvironment(
  'VBMAPP_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/page',
);
const String defaultVbmappDraftSubmitPath = String.fromEnvironment(
  'VBMAPP_DRAFT_SUBMIT_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/submit',
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

  Future<VbmappDraftDetail> fetchDraftDetail(String token, int id);

  Future<VbmappDraftSubmitResult> submitDraft(String token, int id);
}

class ApiVbmappAssessmentClient extends VbmappAssessmentClient {
  const ApiVbmappAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.draftSavePath = defaultVbmappDraftSavePath,
    this.draftDetailPath = defaultVbmappDraftDetailPath,
    this.draftsPagePath = defaultVbmappDraftsPagePath,
    this.draftSubmitPath = defaultVbmappDraftSubmitPath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String draftSavePath;
  final String draftDetailPath;
  final String draftsPagePath;
  final String draftSubmitPath;
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

  @override
  Future<VbmappDraftDetail> fetchDraftDetail(String token, int id) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, draftDetailPath).replace(
        queryParameters: <String, String>{'id': '$id'},
      ),
      token,
      fallbackMessage: 'VB-MAPP草稿详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('VB-MAPP草稿详情返回格式不正确');
    }
    return VbmappDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<VbmappDraftSubmitResult> submitDraft(String token, int id) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftSubmitPath),
      token,
      <String, dynamic>{'id': id},
      fallbackMessage: 'VB-MAPP正式记录提交失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('VB-MAPP提交返回格式不正确');
    }
    return VbmappDraftSubmitResult.fromJson(Map<String, dynamic>.from(data));
  }

  Future<Object?> _getJson(
    Uri uri,
    String token, {
    required String fallbackMessage,
  }) async {
    final http.Response response;
    try {
      response = await _client.get(
        uri,
        headers: <String, String>{
          'Accept': 'application/json',
          'Authorization': 'Bearer $token',
        },
      ).timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw AssessmentScaleApiException('$fallbackMessage，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接VB-MAPP接口：$error');
    }
    return _handleResponse(response, fallbackMessage: fallbackMessage);
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

class VbmappDraftDetail {
  const VbmappDraftDetail({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.milestoneScores,
    required this.barrierScores,
    required this.transitionScores,
  });

  factory VbmappDraftDetail.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> input = _mapFrom(json['input']);
    return VbmappDraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: _dateTextFrom(json['birthDate']),
      assessmentDate: _dateTextFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      milestoneScores: _doubleScoreMapFrom(
        input['milestoneScores'],
        input['milestoneScoreList'],
        'milestoneId',
      ),
      barrierScores: _intScoreMapFrom(
        input['barrierScores'],
        input['barrierScoreList'],
        'barrierCode',
      ),
      transitionScores: _intScoreMapFrom(
        input['transitionScores'],
        input['transitionScoreList'],
        'transitionCode',
      ),
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final Map<String, double> milestoneScores;
  final Map<String, int> barrierScores;
  final Map<String, int> transitionScores;
}

class VbmappDraftSubmitResult {
  const VbmappDraftSubmitResult({
    required this.draftId,
    required this.recordId,
    required this.draftStatus,
  });

  factory VbmappDraftSubmitResult.fromJson(Map<String, dynamic> json) {
    return VbmappDraftSubmitResult(
      draftId: _intFrom(json['draftId']),
      recordId: _intFrom(json['recordId']),
      draftStatus: '${json['draftStatus'] ?? ''}',
    );
  }

  final int draftId;
  final int recordId;
  final String draftStatus;
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

Map<String, dynamic> _mapFrom(Object? raw) {
  if (raw is Map) {
    return Map<String, dynamic>.from(raw);
  }
  if (raw is String && raw.trim().isNotEmpty) {
    final Object? decoded = _decodeResponse(raw);
    if (decoded is Map) {
      return Map<String, dynamic>.from(decoded);
    }
  }
  return <String, dynamic>{};
}

Map<String, double> _doubleScoreMapFrom(
  Object? mapRaw,
  Object? listRaw,
  String codeKey,
) {
  final Map<String, double> out = <String, double>{};
  if (mapRaw is Map) {
    mapRaw.forEach((Object? key, Object? value) {
      final String code = _scoreCodeFrom(key);
      if (code.isNotEmpty) {
        out[code] = _doubleFrom(value);
      }
    });
  }
  if (listRaw is List) {
    for (final Object? raw in listRaw) {
      final Map<String, dynamic> row = _mapFrom(raw);
      final String code = _scoreCodeFrom(row[codeKey] ?? row['code']);
      if (code.isNotEmpty) {
        out[code] = _doubleFrom(row['score']);
      }
    }
  }
  return out;
}

Map<String, int> _intScoreMapFrom(
  Object? mapRaw,
  Object? listRaw,
  String codeKey,
) {
  final Map<String, int> out = <String, int>{};
  if (mapRaw is Map) {
    mapRaw.forEach((Object? key, Object? value) {
      final String code = _scoreCodeFrom(key);
      if (code.isNotEmpty) {
        out[code] = _intFrom(value);
      }
    });
  }
  if (listRaw is List) {
    for (final Object? raw in listRaw) {
      final Map<String, dynamic> row = _mapFrom(raw);
      final String code = _scoreCodeFrom(row[codeKey] ?? row['code']);
      if (code.isNotEmpty) {
        out[code] = _intFrom(row['score']);
      }
    }
  }
  return out;
}

String _scoreCodeFrom(Object? raw) {
  return '${raw ?? ''}'.trim().toUpperCase();
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
