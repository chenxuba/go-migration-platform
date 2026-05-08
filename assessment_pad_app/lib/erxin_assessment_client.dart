import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';

const String defaultErxinTemplateSummaryPath = String.fromEnvironment(
  'ERXIN_TEMPLATE_SUMMARY_PATH',
  defaultValue: '/api/v1/assessments/erxin/form-template/summary',
);
const String defaultErxinTemplateItemPath = String.fromEnvironment(
  'ERXIN_TEMPLATE_ITEM_PATH',
  defaultValue: '/api/v1/assessments/erxin/form-template/item',
);
const String defaultErxinDraftSavePath = String.fromEnvironment(
  'ERXIN_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/erxin/drafts/save',
);
const String defaultErxinDraftItemSavePath = String.fromEnvironment(
  'ERXIN_DRAFT_ITEM_SAVE_PATH',
  defaultValue: '/api/v1/assessments/erxin/drafts/item/save',
);
const String defaultErxinDraftDetailPath = String.fromEnvironment(
  'ERXIN_DRAFT_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/erxin/drafts/detail',
);
const String defaultErxinDraftsPagePath = String.fromEnvironment(
  'ERXIN_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/erxin/drafts/page',
);
const String defaultErxinDraftSubmitPath = String.fromEnvironment(
  'ERXIN_DRAFT_SUBMIT_PATH',
  defaultValue: '/api/v1/assessments/erxin/drafts/submit',
);
const String defaultErxinRecordsPagePath = String.fromEnvironment(
  'ERXIN_RECORDS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/page',
);
const String defaultErxinRecordCategoryStatsPath = String.fromEnvironment(
  'ERXIN_RECORD_CATEGORY_STATS_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/category-stats',
);
const String defaultErxinRecordDetailPath = String.fromEnvironment(
  'ERXIN_RECORD_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/detail',
);
const String defaultErxinRecordUpdatePath = String.fromEnvironment(
  'ERXIN_RECORD_UPDATE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/update',
);
const String defaultErxinRecordConfigUpdatePath = String.fromEnvironment(
  'ERXIN_RECORD_CONFIG_UPDATE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/config/update',
);
const String defaultErxinRecordReportPdfPath = String.fromEnvironment(
  'ERXIN_RECORD_REPORT_PDF_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/report/pdf',
);
const String defaultErxinRecordReportInterpretationAiPath =
    String.fromEnvironment(
  'ERXIN_RECORD_REPORT_INTERPRETATION_AI_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/report/interpretation/ai',
);
const String defaultErxinRecordReportInterpretationPath =
    String.fromEnvironment(
  'ERXIN_RECORD_REPORT_INTERPRETATION_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/report/interpretation',
);

class ErxinAssessmentLaunchArgs {
  const ErxinAssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = '儿心量表-II',
  });

  final int draftId;
  final int studentId;
  final String studentName;
  final String studentAge;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String scaleName;
}

abstract class ErxinAssessmentClient {
  const ErxinAssessmentClient();

  Future<ErxinTemplateSummary> fetchTemplateSummary(String token);

  Future<ErxinAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  });

  Future<ErxinDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  );

  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  });

  Future<ErxinDraftDetail> fetchDraftDetail(String token, int id);

  Future<ErxinDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  );

  Future<void> submitDraft(String token, int draftId);

  Future<ErxinRecordDetail> fetchRecordDetail(String token, int id);

  Future<ErxinRecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  });

  Future<Uint8List> downloadRecordReportPdf(String token, int id);

  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  );

  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  );
}

class ApiErxinAssessmentClient extends ErxinAssessmentClient {
  const ApiErxinAssessmentClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.templateSummaryPath = defaultErxinTemplateSummaryPath,
    this.templateItemPath = defaultErxinTemplateItemPath,
    this.draftSavePath = defaultErxinDraftSavePath,
    this.draftItemSavePath = defaultErxinDraftItemSavePath,
    this.draftDetailPath = defaultErxinDraftDetailPath,
    this.draftsPagePath = defaultErxinDraftsPagePath,
    this.draftSubmitPath = defaultErxinDraftSubmitPath,
    this.recordDetailPath = defaultErxinRecordDetailPath,
    this.recordUpdatePath = defaultErxinRecordUpdatePath,
    this.recordConfigUpdatePath = defaultErxinRecordConfigUpdatePath,
    this.recordReportPdfPath = defaultErxinRecordReportPdfPath,
    this.recordReportInterpretationPath =
        defaultErxinRecordReportInterpretationPath,
    this.recordReportInterpretationAiPath =
        defaultErxinRecordReportInterpretationAiPath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String templateSummaryPath;
  final String templateItemPath;
  final String draftSavePath;
  final String draftItemSavePath;
  final String draftDetailPath;
  final String draftsPagePath;
  final String draftSubmitPath;
  final String recordDetailPath;
  final String recordUpdatePath;
  final String recordConfigUpdatePath;
  final String recordReportPdfPath;
  final String recordReportInterpretationPath;
  final String recordReportInterpretationAiPath;
  final http.Client? httpClient;

  @override
  Future<ErxinTemplateSummary> fetchTemplateSummary(String token) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.get(
      _uri(educationBaseUrl, templateSummaryPath),
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表模板加载失败',
    );
    if (data == null) {
      return ErxinTemplateSummary.empty;
    }
    return ErxinTemplateSummary.fromJson(
        Map<String, dynamic>.from(data as Map));
  }

  @override
  Future<ErxinAssessmentItem> fetchTemplateItem(
    String token, {
    required int itemNo,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      templateItemPath,
      <String, String>{'itemNo': '$itemNo'},
    );
    final http.Response response = await client.get(
      uri,
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '题目说明加载失败',
    );
    if (data == null) {
      return ErxinAssessmentItem.empty;
    }
    return ErxinAssessmentItem.fromJson(Map<String, dynamic>.from(data as Map));
  }

  @override
  Future<ErxinDraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.post(
      _uri(educationBaseUrl, draftSavePath),
      headers: _jsonHeaders(token),
      body: jsonEncode(payload),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表草稿保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿保存返回格式不正确');
    }
    return ErxinDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AssessmentDraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.post(
      _uri(educationBaseUrl, draftsPagePath),
      headers: _jsonHeaders(token),
      body: jsonEncode(<String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'ERXIN2',
          if (studentId > 0) 'studentId': studentId,
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      }),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表草稿列表加载失败',
    );
    if (data is! Map) {
      return AssessmentDraftPage.empty;
    }
    return AssessmentDraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinDraftDetail> fetchDraftDetail(String token, int id) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      draftDetailPath,
      <String, String>{'id': '$id'},
    );
    final http.Response response = await client.get(
      uri,
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表草稿详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('草稿详情返回格式不正确');
    }
    return ErxinDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinDraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.post(
      _uri(educationBaseUrl, draftItemSavePath),
      headers: _jsonHeaders(token),
      body: jsonEncode(payload),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表单题保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('单题保存返回格式不正确');
    }
    return ErxinDraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.post(
      _uri(educationBaseUrl, draftSubmitPath),
      headers: _jsonHeaders(token),
      body: jsonEncode(<String, int>{'id': draftId}),
    );
    _handleResponse(
      response,
      fallbackMessage: '儿心量表正式记录提交失败',
    );
  }

  @override
  Future<ErxinRecordDetail> fetchRecordDetail(String token, int id) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      recordDetailPath,
      <String, String>{'id': '$id'},
    );
    final http.Response response = await client.get(
      uri,
      headers: _authHeaders(token),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表记录详情加载失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('测评记录详情返回格式不正确');
    }
    return ErxinRecordDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinRecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  }) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response = await client.post(
      _uri(educationBaseUrl, recordConfigUpdatePath),
      headers: _jsonHeaders(token),
      body: jsonEncode(<String, dynamic>{
        'id': id,
        'examinerName': examinerName.trim(),
        'assessmentDate': assessmentDate.trim(),
      }),
    );
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '儿心量表评估配置保存失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('评估配置返回格式不正确');
    }
    return ErxinRecordDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Uint8List> downloadRecordReportPdf(String token, int id) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      recordReportPdfPath,
      <String, String>{'id': '$id'},
    );
    final http.Response response;
    try {
      response = await client
          .get(
            uri,
            headers: _authHeaders(token),
          )
          .timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const AssessmentScaleApiException('评估报告PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接评估报告PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      final Object? decoded = response.body.trim().isEmpty
          ? null
          : jsonDecode(utf8.decode(response.bodyBytes));
      throw AssessmentScaleApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final Object? decoded = response.body.trim().isEmpty
          ? null
          : jsonDecode(utf8.decode(response.bodyBytes));
      throw AssessmentScaleApiException(
        _messageFromPayload(decoded) ?? '评估报告PDF加载失败',
      );
    }
    return response.bodyBytes;
  }

  @override
  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final Uri uri = _uri(
      educationBaseUrl,
      recordReportInterpretationPath,
      <String, String>{'id': '$id'},
    );
    final http.Response response;
    try {
      response = await client
          .get(uri, headers: _authHeaders(token))
          .timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const AssessmentScaleApiException('报告解读读取超时，请稍后重试');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接报告解读接口：$error');
    }
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '报告解读读取失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final http.Response response;
    try {
      response = await client
          .post(
            _uri(educationBaseUrl, recordReportInterpretationAiPath),
            headers: _jsonHeaders(token),
            body: jsonEncode(<String, int>{'id': id}),
          )
          .timeout(const Duration(seconds: 190));
    } on TimeoutException {
      throw const AssessmentScaleApiException('报告解读生成超时，请稍后重试');
    } on Object catch (error) {
      throw AssessmentScaleApiException('无法连接报告解读接口：$error');
    }
    final Object? data = _handleResponse(
      response,
      fallbackMessage: '报告解读生成失败',
    );
    if (data is! Map) {
      throw const AssessmentScaleApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }
}

class ErxinReportInterpretation {
  const ErxinReportInterpretation({
    required this.title,
    required this.generatedBy,
    required this.summary,
    required this.domainAnalysis,
    required this.suggestions,
    required this.notes,
    this.model = '',
    this.generatedAt = '',
  });

  factory ErxinReportInterpretation.fromJson(Map<String, dynamic> json) {
    return ErxinReportInterpretation(
      title: '${json['title'] ?? ''}',
      model: '${json['model'] ?? ''}',
      generatedBy: '${json['generatedBy'] ?? ''}',
      generatedAt: '${json['generatedAt'] ?? ''}',
      summary: '${json['summary'] ?? ''}',
      domainAnalysis: _stringListFrom(json['domainAnalysis']),
      suggestions: _stringListFrom(json['suggestions']),
      notes: _stringListFrom(json['notes']),
    );
  }

  static const ErxinReportInterpretation empty = ErxinReportInterpretation(
    title: '',
    generatedBy: '',
    summary: '',
    domainAnalysis: <String>[],
    suggestions: <String>[],
    notes: <String>[],
  );

  final String title;
  final String model;
  final String generatedBy;
  final String generatedAt;
  final String summary;
  final List<String> domainAnalysis;
  final List<String> suggestions;
  final List<String> notes;

  bool get isEmpty =>
      summary.trim().isEmpty &&
      domainAnalysis.isEmpty &&
      suggestions.isEmpty &&
      notes.isEmpty;
}

class ErxinTemplateSummary {
  const ErxinTemplateSummary({
    required this.templateCode,
    required this.title,
    required this.scaleCode,
    required this.scaleVersion,
    required this.itemCount,
    required this.domains,
    required this.ageGroups,
  });

  factory ErxinTemplateSummary.fromJson(Map<String, dynamic> json) {
    return ErxinTemplateSummary(
      templateCode: '${json['templateCode'] ?? ''}',
      title: '${json['title'] ?? ''}',
      scaleCode: '${json['scaleCode'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      itemCount: _intFrom(json['itemCount']),
      domains: _listFrom(json['domains'])
          .map((Map<String, dynamic> item) => ErxinDomain.fromJson(item))
          .toList(),
      ageGroups: _listFrom(json['ageGroups'])
          .map((Map<String, dynamic> item) => ErxinAgeGroup.fromJson(item))
          .toList(),
    );
  }

  static const ErxinTemplateSummary empty = ErxinTemplateSummary(
    templateCode: '',
    title: '',
    scaleCode: 'ERXIN2',
    scaleVersion: '',
    itemCount: 0,
    domains: <ErxinDomain>[],
    ageGroups: <ErxinAgeGroup>[],
  );

  final String templateCode;
  final String title;
  final String scaleCode;
  final String scaleVersion;
  final int itemCount;
  final List<ErxinDomain> domains;
  final List<ErxinAgeGroup> ageGroups;
}

class ErxinDomain {
  const ErxinDomain({
    required this.domainCode,
    required this.domainName,
    required this.sortNo,
  });

  factory ErxinDomain.fromJson(Map<String, dynamic> json) {
    return ErxinDomain(
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      sortNo: _intFrom(json['sortNo']),
    );
  }

  final String domainCode;
  final String domainName;
  final int sortNo;
}

class ErxinAgeGroup {
  const ErxinAgeGroup({
    required this.ageMonth,
    required this.title,
    required this.items,
  });

  factory ErxinAgeGroup.fromJson(Map<String, dynamic> json) {
    return ErxinAgeGroup(
      ageMonth: _intFrom(json['ageMonth']),
      title: '${json['title'] ?? ''}',
      items: _listFrom(json['items'])
          .map((Map<String, dynamic> item) => ErxinItemSummary.fromJson(item))
          .toList(),
    );
  }

  final int ageMonth;
  final String title;
  final List<ErxinItemSummary> items;
}

class ErxinItemSummary {
  const ErxinItemSummary({
    required this.itemNo,
    required this.itemTitle,
    required this.testItem,
    required this.ageMonth,
    required this.domainCode,
    required this.domainName,
    required this.parentReportAllowed,
    required this.attentionIfFailed,
  });

  factory ErxinItemSummary.fromJson(Map<String, dynamic> json) {
    return ErxinItemSummary(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      ageMonth: _intFrom(json['ageMonth']),
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      parentReportAllowed: json['parentReportAllowed'] == true,
      attentionIfFailed: json['attentionIfFailed'] == true,
    );
  }

  final int itemNo;
  final String itemTitle;
  final String testItem;
  final int ageMonth;
  final String domainCode;
  final String domainName;
  final bool parentReportAllowed;
  final bool attentionIfFailed;
}

class ErxinAssessmentItem extends ErxinItemSummary {
  const ErxinAssessmentItem({
    required super.itemNo,
    required super.itemTitle,
    required super.testItem,
    required super.ageMonth,
    required super.domainCode,
    required super.domainName,
    required super.parentReportAllowed,
    required super.attentionIfFailed,
    required this.method,
    required this.passCriteria,
  });

  factory ErxinAssessmentItem.fromJson(Map<String, dynamic> json) {
    return ErxinAssessmentItem(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      ageMonth: _intFrom(json['ageMonth']),
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      parentReportAllowed: json['parentReportAllowed'] == true,
      attentionIfFailed: json['attentionIfFailed'] == true,
      method: '${json['method'] ?? ''}',
      passCriteria: '${json['passCriteria'] ?? ''}',
    );
  }

  static const ErxinAssessmentItem empty = ErxinAssessmentItem(
    itemNo: 0,
    itemTitle: '',
    testItem: '',
    ageMonth: 0,
    domainCode: '',
    domainName: '',
    parentReportAllowed: false,
    attentionIfFailed: false,
    method: '',
    passCriteria: '',
  );

  final String method;
  final String passCriteria;
}

class ErxinDraftDetail {
  const ErxinDraftDetail({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.answeredItemCount,
    required this.completionPercent,
    required this.updatedTime,
    required this.progress,
    required this.input,
  });

  factory ErxinDraftDetail.fromJson(Map<String, dynamic> json) {
    return ErxinDraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      examinerName: '${json['examinerName'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      updatedTime: '${json['updatedTime'] ?? ''}',
      progress: ErxinDraftProgress.fromJson(_mapFrom(json['progress'])),
      input: ErxinDraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final int answeredItemCount;
  final double completionPercent;
  final String updatedTime;
  final ErxinDraftProgress progress;
  final ErxinDraftInput input;
}

class ErxinDraftInput {
  const ErxinDraftInput({
    this.studentId = 0,
    this.studentName = '',
    this.examinerName = '',
    this.remark = '',
    this.birthDate = '',
    this.assessmentDate = '',
    required this.itemPasses,
    required this.itemRemarks,
  });

  factory ErxinDraftInput.fromJson(Map<String, dynamic> json) {
    final Map<int, bool> itemPasses = <int, bool>{};
    for (final MapEntry<String, dynamic> entry
        in _mapFrom(json['itemPasses']).entries) {
      final int itemNo = _intFrom(entry.key);
      if (itemNo > 0 && entry.value is bool) {
        itemPasses[itemNo] = entry.value as bool;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemPassList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final Object? passed = item['passed'];
      if (itemNo > 0 && passed is bool) {
        itemPasses[itemNo] = passed;
      }
    }

    final Map<int, String> itemRemarks = <int, String>{};
    for (final MapEntry<String, dynamic> entry
        in _mapFrom(json['itemRemarks']).entries) {
      final int itemNo = _intFrom(entry.key);
      final String remark = '${entry.value ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemRemarkList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final String remark = '${item['remark'] ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }
    for (final Object? raw in _rawListFrom(json['itemPassList'])) {
      if (raw is! Map) {
        continue;
      }
      final Map<String, dynamic> item = Map<String, dynamic>.from(raw);
      final int itemNo = _intFrom(item['itemNo']);
      final String remark = '${item['remark'] ?? ''}'.trim();
      if (itemNo > 0 && remark.isNotEmpty) {
        itemRemarks[itemNo] = remark;
      }
    }

    return ErxinDraftInput(
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      remark: '${json['remark'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      itemPasses: itemPasses,
      itemRemarks: itemRemarks,
    );
  }

  static const ErxinDraftInput empty = ErxinDraftInput(
    studentId: 0,
    studentName: '',
    examinerName: '',
    remark: '',
    birthDate: '',
    assessmentDate: '',
    itemPasses: <int, bool>{},
    itemRemarks: <int, String>{},
  );

  final int studentId;
  final String studentName;
  final String examinerName;
  final String remark;
  final String birthDate;
  final String assessmentDate;
  final Map<int, bool> itemPasses;
  final Map<int, String> itemRemarks;
}

class ErxinRecordDetail {
  const ErxinRecordDetail({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.assessmentCode,
    required this.assessmentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.updatedTime,
    this.studentGender = '',
    this.studentAvatar = '',
    this.studentPhone = '',
    this.scaleCategory = '',
    this.scaleVersion = '',
    this.ageYears = 0,
    this.ageMonths = 0,
    this.ageDays = 0,
    this.normAgeMonths = 0,
    this.assessmentSequence = 0,
    this.createdTime = '',
    this.remark = '',
    required this.input,
  });

  factory ErxinRecordDetail.fromJson(Map<String, dynamic> json) {
    return ErxinRecordDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      studentGender: '${json['studentGender'] ?? ''}',
      studentAvatar: '${json['studentAvatar'] ?? ''}',
      studentPhone: '${json['studentPhone'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      scaleCategory: '${json['scaleCategory'] ?? ''}',
      scaleVersion: '${json['scaleVersion'] ?? ''}',
      birthDate: _dateOnlyFrom(json['birthDate']),
      assessmentDate: _dateOnlyFrom(json['assessmentDate']),
      ageYears: _intFrom(json['ageYears']),
      ageMonths: _intFrom(json['ageMonths']),
      ageDays: _intFrom(json['ageDays']),
      normAgeMonths: _intFrom(json['normAgeMonths']),
      assessmentSequence: _intFrom(json['assessmentSequence']),
      examinerName: '${json['examinerName'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
      remark: '${json['remark'] ?? ''}',
      input: ErxinDraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String studentGender;
  final String studentAvatar;
  final String studentPhone;
  final String assessmentCode;
  final String assessmentName;
  final String scaleCategory;
  final String scaleVersion;
  final String birthDate;
  final String assessmentDate;
  final int ageYears;
  final int ageMonths;
  final int ageDays;
  final int normAgeMonths;
  final int assessmentSequence;
  final String examinerName;
  final String createdTime;
  final String updatedTime;
  final String remark;
  final ErxinDraftInput input;
}

class ErxinDraftProgress {
  const ErxinDraftProgress({
    required this.itemCount,
    required this.answeredItemCount,
    required this.missingItemCount,
    required this.completionPercent,
    required this.complete,
    required this.canScore,
    required this.missingItemNos,
  });

  factory ErxinDraftProgress.fromJson(Map<String, dynamic> json) {
    return ErxinDraftProgress(
      itemCount: _intFrom(json['itemCount']),
      answeredItemCount: _intFrom(json['answeredItemCount']),
      missingItemCount: _intFrom(json['missingItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      complete: json['complete'] == true,
      canScore: json['canScore'] == true,
      missingItemNos: _rawListFrom(json['missingItemNos'])
          .map(_intFrom)
          .where((int itemNo) => itemNo > 0)
          .toList(),
    );
  }

  static const ErxinDraftProgress empty = ErxinDraftProgress(
    itemCount: 0,
    answeredItemCount: 0,
    missingItemCount: 0,
    completionPercent: 0,
    complete: false,
    canScore: false,
    missingItemNos: <int>[],
  );

  final int itemCount;
  final int answeredItemCount;
  final int missingItemCount;
  final double completionPercent;
  final bool complete;
  final bool canScore;
  final List<int> missingItemNos;
}

Uri _uri(String base, String path, [Map<String, String>? query]) {
  final Uri baseUri = Uri.parse(base);
  final String normalizedPath = path.startsWith('/') ? path : '/$path';
  return baseUri.replace(
    path: normalizedPath,
    queryParameters: query == null || query.isEmpty ? null : query,
  );
}

Map<String, String> _authHeaders(String token) {
  return <String, String>{
    'Authorization': 'Bearer $token',
    'Accept': 'application/json',
  };
}

Map<String, String> _jsonHeaders(String token) {
  return <String, String>{
    ..._authHeaders(token),
    'Content-Type': 'application/json',
  };
}

Object? _handleResponse(
  http.Response response, {
  required String fallbackMessage,
}) {
  final Object? decoded = response.body.trim().isEmpty
      ? null
      : jsonDecode(utf8.decode(response.bodyBytes));
  if (response.statusCode == 401 || response.statusCode == 403) {
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
  if (decoded is Map<String, dynamic>) {
    final bool success = decoded['success'] == true;
    if (success) {
      return decoded['data'];
    }
    throw AssessmentScaleApiException(
      _messageFromPayload(decoded) ?? fallbackMessage,
    );
  }
  return decoded;
}

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is List) {
    return value
        .whereType<Map>()
        .map((Map item) => Map<String, dynamic>.from(item))
        .toList();
  }
  return <Map<String, dynamic>>[];
}

List<Object?> _rawListFrom(Object? value) {
  if (value is List) {
    return value;
  }
  return <Object?>[];
}

List<String> _stringListFrom(Object? value) {
  if (value is List) {
    return value
        .map((Object? item) => '${item ?? ''}'.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  return <String>[];
}

Map<String, dynamic> _mapFrom(Object? value) {
  if (value is Map) {
    return Map<String, dynamic>.from(value);
  }
  return <String, dynamic>{};
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  if (value is String) {
    return int.tryParse(value.trim()) ?? 0;
  }
  return 0;
}

double _doubleFrom(Object? value) {
  if (value is double) {
    return value;
  }
  if (value is int) {
    return value.toDouble();
  }
  if (value is num) {
    return value.toDouble();
  }
  if (value is String) {
    return double.tryParse(value.trim()) ?? 0;
  }
  return 0;
}

String _dateOnlyFrom(Object? value) {
  final String text = '${value ?? ''}'.trim();
  if (text.isEmpty) {
    return '';
  }
  final RegExpMatch? match =
      RegExp(r'^(\d{4})[-/](\d{1,2})[-/](\d{1,2})').firstMatch(text);
  if (match != null) {
    final String year = match.group(1)!;
    final String month = match.group(2)!.padLeft(2, '0');
    final String day = match.group(3)!.padLeft(2, '0');
    return '$year-$month-$day';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    return text;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')}';
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map<String, dynamic>) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  return null;
}
