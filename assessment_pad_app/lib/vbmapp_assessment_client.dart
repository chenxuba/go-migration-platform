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
const String defaultVbmappDraftItemSavePath = String.fromEnvironment(
  'VBMAPP_DRAFT_ITEM_SAVE_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/drafts/item/save',
);
const String defaultVbmappSchemaPath = String.fromEnvironment(
  'VBMAPP_SCHEMA_PATH',
  defaultValue: '/api/v1/assessments/vbmapp/schema',
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

  Future<VbmappDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  );

  Future<VbmappDraftDetail> fetchDraftDetail(String token, int id);

  Future<VbmappAssessmentSchema> fetchAssessmentSchema(String token);

  Future<VbmappDraftSubmitResult> submitDraft(String token, int id);
}

class ApiVbmappAssessmentClient extends VbmappAssessmentClient {
  const ApiVbmappAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.draftSavePath = defaultVbmappDraftSavePath,
    this.draftDetailPath = defaultVbmappDraftDetailPath,
    this.draftsPagePath = defaultVbmappDraftsPagePath,
    this.draftSubmitPath = defaultVbmappDraftSubmitPath,
    this.draftItemSavePath = defaultVbmappDraftItemSavePath,
    this.schemaPath = defaultVbmappSchemaPath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String draftSavePath;
  final String draftDetailPath;
  final String draftsPagePath;
  final String draftSubmitPath;
  final String draftItemSavePath;
  final String schemaPath;
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
    return AssessmentDraftPage.fromJson(_mapFrom(data));
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
    return VbmappDraftSaveResult.fromJson(_mapFrom(data));
  }

  @override
  Future<VbmappDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data = await _postJson(
      _uri(educationBaseUrl, draftItemSavePath),
      token,
      payload,
      fallbackMessage: 'VB-MAPP单题证据保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('VB-MAPP单题证据保存返回格式不正确');
    }
    return VbmappDraftDetail.fromJson(_mapFrom(data));
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
    return VbmappDraftDetail.fromJson(_mapFrom(data));
  }

  @override
  Future<VbmappAssessmentSchema> fetchAssessmentSchema(String token) async {
    final Object? data = await _getJson(
      _uri(educationBaseUrl, schemaPath),
      token,
      fallbackMessage: 'VB-MAPP智能题库加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('VB-MAPP智能题库返回格式不正确');
    }
    return VbmappAssessmentSchema.fromJson(_mapFrom(data));
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
    return VbmappDraftSubmitResult.fromJson(_mapFrom(data));
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
      studentName: _textFrom(json['studentName']),
      assessmentDate: _dateTextFrom(json['assessmentDate']),
      examinerName: _textFrom(json['examinerName']),
      status: _textFrom(json['status']),
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
    required this.itemResponses,
  });

  factory VbmappDraftDetail.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> input = _mapFrom(json['input']);
    return VbmappDraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: _textFrom(json['studentName']),
      birthDate: _dateTextFrom(json['birthDate']),
      assessmentDate: _dateTextFrom(json['assessmentDate']),
      examinerName: _textFrom(json['examinerName']),
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
      itemResponses: _itemResponsesFrom(input['itemResponses']),
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
  final Map<String, Map<String, Map<String, dynamic>>> itemResponses;
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
      draftStatus: _textFrom(json['draftStatus']),
    );
  }

  final int draftId;
  final int recordId;
  final String draftStatus;
}

class VbmappAssessmentSchema {
  const VbmappAssessmentSchema({
    required this.scaleVersion,
    required this.itemSchemas,
    required this.materialProfiles,
  });

  factory VbmappAssessmentSchema.fromJson(Map<String, dynamic> json) {
    final Map<String, VbmappItemResponseSchema> itemSchemas =
        <String, VbmappItemResponseSchema>{};
    void collect(Object? raw) {
      if (raw is! List) {
        return;
      }
      for (final Object? itemRaw in raw) {
        final VbmappItemResponseSchema schema =
            VbmappItemResponseSchema.fromJson(_mapFrom(itemRaw));
        if (schema.itemCode.isNotEmpty) {
          itemSchemas[_schemaKey(schema.moduleCode, schema.itemCode)] = schema;
        }
      }
    }

    collect(json['milestoneResponseSchemas']);
    collect(json['barrierResponseSchemas']);
    collect(json['transitionResponseSchemas']);

    final Map<String, VbmappMaterialProfile> materialProfiles =
        <String, VbmappMaterialProfile>{};
    final Object? profilesRaw = json['responseMaterialProfiles'];
    if (profilesRaw is Map) {
      profilesRaw.forEach((Object? key, Object? value) {
        final String id = _textFrom(key);
        if (id.isNotEmpty) {
          materialProfiles[id] =
              VbmappMaterialProfile.fromJson(_mapFrom(value));
        }
      });
    }

    return VbmappAssessmentSchema(
      scaleVersion: _textFrom(json['scaleVersion']),
      itemSchemas: itemSchemas,
      materialProfiles: materialProfiles,
    );
  }

  final String scaleVersion;
  final Map<String, VbmappItemResponseSchema> itemSchemas;
  final Map<String, VbmappMaterialProfile> materialProfiles;

  VbmappItemResponseSchema? schemaFor(String moduleCode, String itemCode) {
    return itemSchemas[_schemaKey(moduleCode, itemCode)];
  }
}

class VbmappItemResponseSchema {
  const VbmappItemResponseSchema({
    required this.moduleCode,
    required this.itemCode,
    required this.uiPattern,
    required this.recordDepth,
    required this.materialProfileId,
    required this.whyRecord,
    required this.evidenceTargets,
    required this.qualityChecks,
    required this.scoreStrategy,
    required this.onePointCriteria,
    required this.halfPointCriteria,
  });

  factory VbmappItemResponseSchema.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> scoreEvidence = _mapFrom(json['scoreEvidence']);
    final Map<String, dynamic> autoCompletion =
        _mapFrom(json['autoCompletion']);
    return VbmappItemResponseSchema(
      moduleCode: _textFrom(json['moduleCode']),
      itemCode: _scoreCodeFrom(
        json['milestoneId'] ?? json['barrierCode'] ?? json['transitionCode'],
      ),
      uiPattern: _textFrom(json['uiPattern']),
      recordDepth: _textFrom(json['recordDepth']),
      materialProfileId: _textFrom(json['materialProfileId']),
      whyRecord: _textFrom(json['whyRecord']),
      evidenceTargets: _stringListFrom(json['evidenceTargets']),
      qualityChecks: _stringListFrom(json['qualityChecks']),
      scoreStrategy: _textFrom(autoCompletion['scoreStrategy']),
      onePointCriteria: _textFrom(scoreEvidence['onePointCriteria']),
      halfPointCriteria: _textFrom(scoreEvidence['halfPointCriteria']),
    );
  }

  final String moduleCode;
  final String itemCode;
  final String uiPattern;
  final String recordDepth;
  final String materialProfileId;
  final String whyRecord;
  final List<String> evidenceTargets;
  final List<String> qualityChecks;
  final String scoreStrategy;
  final String onePointCriteria;
  final String halfPointCriteria;
}

class VbmappMaterialProfile {
  const VbmappMaterialProfile({
    required this.label,
    required this.suggestedTypes,
    this.recommendedMaterials = const <VbmappMaterialSuggestion>[],
    this.quickPicks = const <String>[],
    required this.preparationChecks,
  });

  factory VbmappMaterialProfile.fromJson(Map<String, dynamic> json) {
    return VbmappMaterialProfile(
      label: _textFrom(json['label']),
      suggestedTypes: _stringListFrom(json['suggestedTypes']),
      recommendedMaterials:
          _materialSuggestionsFrom(json['recommendedMaterials']),
      quickPicks: _stringListFrom(
        json['quickPicks'] ?? json['recommendedWords'],
      ),
      preparationChecks: _stringListFrom(json['preparationChecks']),
    );
  }

  final String label;
  final List<String> suggestedTypes;
  final List<VbmappMaterialSuggestion> recommendedMaterials;
  final List<String> quickPicks;
  final List<String> preparationChecks;

  List<String> get quickPickLabels {
    final List<String> values = <String>[
      for (final VbmappMaterialSuggestion material in recommendedMaterials)
        material.name,
      ...quickPicks,
    ];
    final Set<String> seen = <String>{};
    return values.where((String value) {
      final String normalized = value.trim();
      if (normalized.isEmpty || seen.contains(normalized)) {
        return false;
      }
      seen.add(normalized);
      return true;
    }).toList(growable: false);
  }
}

class VbmappMaterialSuggestion {
  const VbmappMaterialSuggestion({
    required this.id,
    required this.name,
    required this.type,
  });

  factory VbmappMaterialSuggestion.fromJson(Map<String, dynamic> json) {
    return VbmappMaterialSuggestion(
      id: _textFrom(json['id']),
      name: _textFrom(json['name'] ?? json['label'] ?? json['materialName']),
      type: _textFrom(json['type'] ?? json['materialType']),
    );
  }

  final String id;
  final String name;
  final String type;
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
    final Map<String, dynamic> envelope = _mapFrom(decoded);
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
    final Map<String, dynamic> out = <String, dynamic>{};
    raw.forEach((Object? key, Object? value) {
      final String normalizedKey = _textFrom(key);
      if (normalizedKey.isNotEmpty) {
        out[normalizedKey] = value;
      }
    });
    return out;
  }
  if (raw is String && raw.trim().isNotEmpty) {
    final Object? decoded = _decodeResponse(raw);
    if (decoded is Map) {
      return _mapFrom(decoded);
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

Map<String, Map<String, Map<String, dynamic>>> _itemResponsesFrom(
  Object? raw,
) {
  final Map<String, Map<String, Map<String, dynamic>>> out =
      <String, Map<String, Map<String, dynamic>>>{};
  if (raw is! Map) {
    return out;
  }
  raw.forEach((Object? moduleKey, Object? moduleRaw) {
    final String moduleCode = _textFrom(moduleKey);
    if (moduleCode.isEmpty || moduleRaw is! Map) {
      return;
    }
    moduleRaw.forEach((Object? itemKey, Object? itemRaw) {
      final String itemCode = _scoreCodeFrom(itemKey);
      if (itemCode.isEmpty) {
        return;
      }
      out.putIfAbsent(
              moduleCode, () => <String, Map<String, dynamic>>{})[itemCode] =
          _mapFrom(itemRaw);
    });
  });
  return out;
}

List<String> _stringListFrom(Object? raw) {
  if (raw is! List) {
    return const <String>[];
  }
  return raw
      .map(_textFrom)
      .where((String value) => value.isNotEmpty)
      .toList(growable: false);
}

List<VbmappMaterialSuggestion> _materialSuggestionsFrom(Object? raw) {
  if (raw is! List) {
    return const <VbmappMaterialSuggestion>[];
  }
  return raw
      .map((Object? value) {
        if (value is Map) {
          return VbmappMaterialSuggestion.fromJson(_mapFrom(value));
        }
        final String name = _textFrom(value);
        return VbmappMaterialSuggestion(id: '', name: name, type: '');
      })
      .where((VbmappMaterialSuggestion value) => value.name.isNotEmpty)
      .toList(growable: false);
}

String _scoreCodeFrom(Object? raw) {
  return _textFrom(raw).toUpperCase();
}

String _textFrom(Object? raw) {
  if (raw == null) {
    return '';
  }
  if (raw is String) {
    return raw.trim();
  }
  if (raw is num || raw is bool) {
    return '$raw'.trim();
  }
  return '$raw'.trim();
}

String _schemaKey(String moduleCode, String itemCode) {
  return '${_textFrom(moduleCode).toLowerCase()}::${_scoreCodeFrom(itemCode)}';
}

String _dateTextFrom(Object? raw) {
  final String value = _textFrom(raw);
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
  return int.tryParse(_textFrom(raw)) ?? 0;
}

double _doubleFrom(Object? raw) {
  if (raw is num) {
    return raw.toDouble();
  }
  return double.tryParse(_textFrom(raw)) ?? 0;
}
