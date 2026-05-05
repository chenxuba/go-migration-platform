import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'home_client.dart';

const String defaultPep3TemplateSummaryPath = String.fromEnvironment(
  'PEP3_TEMPLATE_SUMMARY_PATH',
  defaultValue: '/api/v1/assessments/pep3/form-template/summary',
);
const String defaultPep3TemplateItemPath = String.fromEnvironment(
  'PEP3_TEMPLATE_ITEM_PATH',
  defaultValue: '/api/v1/assessments/pep3/form-template/item',
);
const String defaultPep3DraftSavePath = String.fromEnvironment(
  'PEP3_DRAFT_SAVE_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/save',
);
const String defaultPep3DraftItemSavePath = String.fromEnvironment(
  'PEP3_DRAFT_ITEM_SAVE_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/item/save',
);
const String defaultPep3DraftDetailPath = String.fromEnvironment(
  'PEP3_DRAFT_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/detail',
);
const String defaultPep3DraftsPagePath = String.fromEnvironment(
  'PEP3_DRAFTS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/page',
);
const String defaultPep3DraftSubmitPath = String.fromEnvironment(
  'PEP3_DRAFT_SUBMIT_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/submit',
);
const String defaultPep3CaregiverInvitePath = String.fromEnvironment(
  'PEP3_CAREGIVER_INVITE_PATH',
  defaultValue: '/api/v1/assessments/pep3/drafts/caregiver-report/invite',
);
const String defaultPep3RecordsPagePath = String.fromEnvironment(
  'PEP3_RECORDS_PAGE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/page',
);
const String defaultPep3RecordDetailPath = String.fromEnvironment(
  'PEP3_RECORD_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/detail',
);

class Pep3AssessmentLaunchArgs {
  const Pep3AssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = 'PEP-3',
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

class Pep3ApiException implements Exception {
  const Pep3ApiException(this.message, {this.unauthorized = false});

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

class Pep3ScoreOption {
  const Pep3ScoreOption({
    required this.value,
    required this.label,
    required this.description,
  });

  factory Pep3ScoreOption.fromJson(Map<String, dynamic> json) {
    return Pep3ScoreOption(
      value: _intFrom(json['value']),
      label: '${json['label'] ?? ''}',
      description: '${json['description'] ?? ''}',
    );
  }

  final int value;
  final String label;
  final String description;
}

class Pep3TemplateSummary {
  const Pep3TemplateSummary({
    required this.title,
    required this.itemCount,
    required this.scoreOptions,
    required this.itemGroups,
  });

  factory Pep3TemplateSummary.fromJson(Map<String, dynamic> json) {
    return Pep3TemplateSummary(
      title: '${json['title'] ?? 'PEP-3'}',
      itemCount: _intFrom(json['itemCount']),
      scoreOptions: _listFrom(json['scoreOptions'])
          .map(Pep3ScoreOption.fromJson)
          .toList(),
      itemGroups: _listFrom(json['itemGroups'])
          .map(Pep3ItemGroupSummary.fromJson)
          .toList(),
    );
  }

  static const Pep3TemplateSummary empty = Pep3TemplateSummary(
    title: 'PEP-3',
    itemCount: 0,
    scoreOptions: <Pep3ScoreOption>[],
    itemGroups: <Pep3ItemGroupSummary>[],
  );

  final String title;
  final int itemCount;
  final List<Pep3ScoreOption> scoreOptions;
  final List<Pep3ItemGroupSummary> itemGroups;

  List<Pep3ItemSummary> get allItems =>
      itemGroups.expand((Pep3ItemGroupSummary group) => group.items).toList();
}

class Pep3ItemGroupSummary {
  const Pep3ItemGroupSummary({
    required this.groupCode,
    required this.title,
    required this.bookletPageNo,
    required this.startItemNo,
    required this.endItemNo,
    required this.items,
  });

  factory Pep3ItemGroupSummary.fromJson(Map<String, dynamic> json) {
    return Pep3ItemGroupSummary(
      groupCode: '${json['groupCode'] ?? ''}',
      title: '${json['title'] ?? ''}',
      bookletPageNo: _intFrom(json['bookletPageNo']),
      startItemNo: _intFrom(json['startItemNo']),
      endItemNo: _intFrom(json['endItemNo']),
      items: _listFrom(json['items']).map(Pep3ItemSummary.fromJson).toList(),
    );
  }

  final String groupCode;
  final String title;
  final int bookletPageNo;
  final int startItemNo;
  final int endItemNo;
  final List<Pep3ItemSummary> items;

  String get displayTitle {
    if (bookletPageNo > 0 && startItemNo > 0 && endItemNo > 0) {
      return '第 $bookletPageNo 页 $startItemNo-$endItemNo题';
    }
    if (title.trim().isNotEmpty) {
      return title.trim();
    }
    return '记录册页面';
  }

  String get key =>
      groupCode.trim().isNotEmpty ? groupCode.trim() : 'page_$bookletPageNo';
}

class Pep3ItemSummary {
  const Pep3ItemSummary({
    required this.itemNo,
    required this.itemTitle,
    required this.testItem,
    required this.domainCode,
    required this.domainName,
  });

  factory Pep3ItemSummary.fromJson(Map<String, dynamic> json) {
    return Pep3ItemSummary(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
    );
  }

  final int itemNo;
  final String itemTitle;
  final String testItem;
  final String domainCode;
  final String domainName;

  String get displayTitle => _displayItemTitle(itemTitle, testItem);
}

class Pep3AssessmentItem extends Pep3ItemSummary {
  const Pep3AssessmentItem({
    required super.itemNo,
    required super.itemTitle,
    required super.testItem,
    required super.domainCode,
    required super.domainName,
    required this.materials,
    required this.materialImages,
    required this.method,
    required this.guidance,
    required this.guidanceVideo,
    required this.standard,
    required this.scoreOptions,
    required this.recordFields,
  });

  factory Pep3AssessmentItem.fromJson(Map<String, dynamic> json) {
    return Pep3AssessmentItem(
      itemNo: _intFrom(json['itemNo']),
      itemTitle: '${json['itemTitle'] ?? ''}',
      testItem: '${json['testItem'] ?? ''}',
      domainCode: '${json['domainCode'] ?? ''}',
      domainName: '${json['domainName'] ?? ''}',
      materials: '${json['materials'] ?? ''}',
      materialImages: _rawListFrom(json['materialImages'])
          .map((Object? item) => '$item'.trim())
          .where((String item) => item.isNotEmpty)
          .toList(),
      method: '${json['method'] ?? ''}',
      guidance: '${json['guidance'] ?? ''}',
      guidanceVideo: '${json['guidanceVideo'] ?? ''}',
      standard: '${json['standard'] ?? ''}',
      scoreOptions: _listFrom(json['scoreOptions'])
          .map(Pep3ScoreOption.fromJson)
          .toList(),
      recordFields: _listFrom(json['recordFields'])
          .map(Pep3RecordField.fromJson)
          .toList(),
    );
  }

  static const Pep3AssessmentItem empty = Pep3AssessmentItem(
    itemNo: 0,
    itemTitle: '',
    testItem: '',
    domainCode: '',
    domainName: '',
    materials: '',
    materialImages: <String>[],
    method: '',
    guidance: '',
    guidanceVideo: '',
    standard: '',
    scoreOptions: <Pep3ScoreOption>[],
    recordFields: <Pep3RecordField>[],
  );

  final String materials;
  final List<String> materialImages;
  final String method;
  final String guidance;
  final String guidanceVideo;
  final String standard;
  final List<Pep3ScoreOption> scoreOptions;
  final List<Pep3RecordField> recordFields;
}

class Pep3RecordField {
  const Pep3RecordField({
    required this.key,
    required this.label,
    required this.fieldType,
    required this.displayType,
    required this.required,
    required this.placeholder,
    required this.options,
  });

  factory Pep3RecordField.fromJson(Map<String, dynamic> json) {
    return Pep3RecordField(
      key: '${json['key'] ?? ''}',
      label: '${json['label'] ?? ''}',
      fieldType: '${json['fieldType'] ?? ''}',
      displayType: '${json['displayType'] ?? ''}',
      required: json['required'] == true,
      placeholder: '${json['placeholder'] ?? ''}',
      options: _listFrom(json['options'])
          .map(Pep3RecordFieldOption.fromJson)
          .toList(),
    );
  }

  final String key;
  final String label;
  final String fieldType;
  final String displayType;
  final bool required;
  final String placeholder;
  final List<Pep3RecordFieldOption> options;
}

class Pep3RecordFieldOption {
  const Pep3RecordFieldOption({required this.value, required this.label});

  factory Pep3RecordFieldOption.fromJson(Map<String, dynamic> json) {
    return Pep3RecordFieldOption(
      value: '${json['value'] ?? ''}',
      label: '${json['label'] ?? ''}',
    );
  }

  final String value;
  final String label;
}

class Pep3DraftPage {
  const Pep3DraftPage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  factory Pep3DraftPage.fromJson(Map<String, dynamic> json) {
    return Pep3DraftPage(
      items: _listFrom(json['items']).map(Pep3DraftSummary.fromJson).toList(),
      total: _intFrom(json['total']),
      current: _intFrom(json['current']),
      size: _intFrom(json['size']),
    );
  }

  final List<Pep3DraftSummary> items;
  final int total;
  final int current;
  final int size;
}

class Pep3DraftSummary {
  const Pep3DraftSummary({
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
  });

  factory Pep3DraftSummary.fromJson(Map<String, dynamic> json) {
    return Pep3DraftSummary(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      updatedTime: '${json['updatedTime'] ?? ''}',
      progress: Pep3DraftProgress.fromJson(_mapFrom(json['progress'])),
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
  final Pep3DraftProgress progress;
}

class Pep3DraftDetail extends Pep3DraftSummary {
  const Pep3DraftDetail({
    required super.id,
    required super.studentId,
    required super.studentName,
    required super.birthDate,
    required super.assessmentDate,
    required super.examinerName,
    required super.answeredItemCount,
    required super.completionPercent,
    required super.updatedTime,
    required super.progress,
    required this.input,
  });

  factory Pep3DraftDetail.fromJson(Map<String, dynamic> json) {
    return Pep3DraftDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      answeredItemCount: _intFrom(json['answeredItemCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      updatedTime: '${json['updatedTime'] ?? ''}',
      progress: Pep3DraftProgress.fromJson(_mapFrom(json['progress'])),
      input: Pep3DraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final Pep3DraftInput input;
}

class Pep3DraftProgress {
  const Pep3DraftProgress({
    required this.itemCount,
    required this.answeredItemCount,
    required this.missingItemCount,
    required this.rawScoreCount,
    required this.caregiverRawScoreCount,
    required this.completionPercent,
    required this.complete,
    required this.canScore,
    required this.missingItemNos,
  });

  factory Pep3DraftProgress.fromJson(Map<String, dynamic> json) {
    return Pep3DraftProgress(
      itemCount: _intFrom(json['itemCount']),
      answeredItemCount: _intFrom(json['answeredItemCount']),
      missingItemCount: _intFrom(json['missingItemCount']),
      rawScoreCount: _intFrom(json['rawScoreCount']),
      caregiverRawScoreCount: _intFrom(json['caregiverRawScoreCount']),
      completionPercent: _doubleFrom(json['completionPercent']),
      complete: json['complete'] == true,
      canScore: json['canScore'] == true,
      missingItemNos: _rawListFrom(json['missingItemNos'])
          .map(_intFrom)
          .where((int item) => item > 0)
          .toList(),
    );
  }

  static const Pep3DraftProgress empty = Pep3DraftProgress(
    itemCount: 0,
    answeredItemCount: 0,
    missingItemCount: 0,
    rawScoreCount: 0,
    caregiverRawScoreCount: 0,
    completionPercent: 0,
    complete: false,
    canScore: false,
    missingItemNos: <int>[],
  );

  final int itemCount;
  final int answeredItemCount;
  final int missingItemCount;
  final int rawScoreCount;
  final int caregiverRawScoreCount;
  final double completionPercent;
  final bool complete;
  final bool canScore;
  final List<int> missingItemNos;
}

class Pep3DraftInput {
  const Pep3DraftInput({
    required this.studentId,
    required this.studentName,
    required this.examinerName,
    required this.birthDate,
    required this.assessmentDate,
    required this.remark,
    required this.allowMissingItems,
    required this.itemScores,
    required this.itemRecordValues,
  });

  factory Pep3DraftInput.fromJson(Map<String, dynamic> json) {
    final Map<int, int> itemScores = <int, int>{};
    _mapFrom(json['itemScores']).forEach((String key, dynamic value) {
      final int itemNo = int.tryParse(key) ?? 0;
      final int score = _intFrom(value);
      if (itemNo > 0 && _isValidScore(score)) {
        itemScores[itemNo] = score;
      }
    });
    for (final Map<String, dynamic> item in _listFrom(json['itemScoreList'])) {
      final int itemNo = _intFrom(item['itemNo']);
      final int score = _intFrom(item['score']);
      if (itemNo > 0 && _isValidScore(score)) {
        itemScores[itemNo] = score;
      }
    }

    final Map<int, Map<String, dynamic>> recordValues =
        <int, Map<String, dynamic>>{};
    _mapFrom(json['itemRecordValues']).forEach((String key, dynamic value) {
      final int itemNo = int.tryParse(key) ?? 0;
      if (itemNo > 0 && value is Map) {
        recordValues[itemNo] = Map<String, dynamic>.from(value);
      }
    });
    for (final Map<String, dynamic> item
        in _listFrom(json['itemRecordValueList'])) {
      final int itemNo = _intFrom(item['itemNo']);
      final String fieldKey = '${item['fieldKey'] ?? ''}'.trim();
      if (itemNo > 0 && fieldKey.isNotEmpty) {
        recordValues.putIfAbsent(itemNo, () => <String, dynamic>{});
        recordValues[itemNo]![fieldKey] = item['value'];
      }
    }

    return Pep3DraftInput(
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      remark: '${json['remark'] ?? ''}',
      allowMissingItems: json['allowMissingItems'] != false,
      itemScores: itemScores,
      itemRecordValues: recordValues,
    );
  }

  static const Pep3DraftInput empty = Pep3DraftInput(
    studentId: 0,
    studentName: '',
    examinerName: '',
    birthDate: '',
    assessmentDate: '',
    remark: '',
    allowMissingItems: true,
    itemScores: <int, int>{},
    itemRecordValues: <int, Map<String, dynamic>>{},
  );

  final int studentId;
  final String studentName;
  final String examinerName;
  final String birthDate;
  final String assessmentDate;
  final String remark;
  final bool allowMissingItems;
  final Map<int, int> itemScores;
  final Map<int, Map<String, dynamic>> itemRecordValues;
}

class Pep3CaregiverInvite {
  const Pep3CaregiverInvite({
    required this.miniProgramCodeDataUrl,
    required this.qrCodeValue,
    required this.wechatUrlLink,
    required this.miniProgramPath,
    required this.url,
  });

  factory Pep3CaregiverInvite.fromJson(Map<String, dynamic> json) {
    return Pep3CaregiverInvite(
      miniProgramCodeDataUrl: '${json['miniProgramCodeDataUrl'] ?? ''}',
      qrCodeValue: '${json['qrCodeValue'] ?? ''}',
      wechatUrlLink: '${json['wechatUrlLink'] ?? ''}',
      miniProgramPath: '${json['miniProgramPath'] ?? ''}',
      url: '${json['url'] ?? ''}',
    );
  }

  final String miniProgramCodeDataUrl;
  final String qrCodeValue;
  final String wechatUrlLink;
  final String miniProgramPath;
  final String url;

  String get qrValue {
    for (final String value in <String>[
      qrCodeValue,
      wechatUrlLink,
      miniProgramPath,
      url,
    ]) {
      if (value.trim().isNotEmpty) {
        return value.trim();
      }
    }
    return '';
  }
}

class Pep3RecordPage {
  const Pep3RecordPage({
    required this.items,
    required this.total,
    required this.current,
    required this.size,
  });

  factory Pep3RecordPage.fromJson(Map<String, dynamic> json) {
    return Pep3RecordPage(
      items: _listFrom(json['items']).map(Pep3RecordSummary.fromJson).toList(),
      total: _intFrom(json['total']),
      current: _intFrom(json['current']),
      size: _intFrom(json['size']),
    );
  }

  final List<Pep3RecordSummary> items;
  final int total;
  final int current;
  final int size;
}

class Pep3RecordSummary {
  const Pep3RecordSummary({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.assessmentCode,
    required this.assessmentName,
    required this.birthDate,
    required this.assessmentDate,
    required this.examinerName,
    required this.updatedTime,
  });

  factory Pep3RecordSummary.fromJson(Map<String, dynamic> json) {
    return Pep3RecordSummary(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
    );
  }

  final int id;
  final int studentId;
  final String studentName;
  final String assessmentCode;
  final String assessmentName;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String updatedTime;
}

class Pep3RecordDetail extends Pep3RecordSummary {
  const Pep3RecordDetail({
    required super.id,
    required super.studentId,
    required super.studentName,
    required super.assessmentCode,
    required super.assessmentName,
    required super.birthDate,
    required super.assessmentDate,
    required super.examinerName,
    required super.updatedTime,
    required this.input,
  });

  factory Pep3RecordDetail.fromJson(Map<String, dynamic> json) {
    return Pep3RecordDetail(
      id: _intFrom(json['id']),
      studentId: _intFrom(json['studentId']),
      studentName: '${json['studentName'] ?? ''}',
      assessmentCode: '${json['assessmentCode'] ?? ''}',
      assessmentName: '${json['assessmentName'] ?? ''}',
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      examinerName: '${json['examinerName'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
      input: Pep3DraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final Pep3DraftInput input;
}

abstract interface class Pep3AssessmentClient {
  Future<Pep3TemplateSummary> fetchTemplateSummary(String token);

  Future<Pep3AssessmentItem> fetchTemplateItem(String token, int itemNo);

  Future<Pep3DraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  });

  Future<Pep3DraftDetail> fetchDraftDetail(String token, int id);

  Future<Pep3DraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  );

  Future<Pep3DraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  );

  Future<Pep3CaregiverInvite> inviteCaregiverReport(String token, int draftId);

  Future<void> submitDraft(String token, int draftId);

  Future<Pep3RecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentDateEnd = '',
  });

  Future<Pep3RecordDetail> fetchRecordDetail(String token, int id);
}

class ApiPep3AssessmentClient implements Pep3AssessmentClient {
  const ApiPep3AssessmentClient({
    this.educationBaseUrl = defaultEducationApiBaseUrl,
    this.templateSummaryPath = defaultPep3TemplateSummaryPath,
    this.templateItemPath = defaultPep3TemplateItemPath,
    this.draftSavePath = defaultPep3DraftSavePath,
    this.draftItemSavePath = defaultPep3DraftItemSavePath,
    this.draftDetailPath = defaultPep3DraftDetailPath,
    this.draftsPagePath = defaultPep3DraftsPagePath,
    this.draftSubmitPath = defaultPep3DraftSubmitPath,
    this.caregiverInvitePath = defaultPep3CaregiverInvitePath,
    this.recordsPagePath = defaultPep3RecordsPagePath,
    this.recordDetailPath = defaultPep3RecordDetailPath,
  });

  final String educationBaseUrl;
  final String templateSummaryPath;
  final String templateItemPath;
  final String draftSavePath;
  final String draftItemSavePath;
  final String draftDetailPath;
  final String draftsPagePath;
  final String draftSubmitPath;
  final String caregiverInvitePath;
  final String recordsPagePath;
  final String recordDetailPath;

  @override
  Future<Pep3TemplateSummary> fetchTemplateSummary(String token) async {
    final Object? data = await _getJson(_uri(templateSummaryPath), token);
    if (data is! Map) {
      return Pep3TemplateSummary.empty;
    }
    return Pep3TemplateSummary.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3AssessmentItem> fetchTemplateItem(
    String token,
    int itemNo,
  ) async {
    final Uri uri =
        _uri(templateItemPath).replace(queryParameters: <String, String>{
      'itemNo': '$itemNo',
    });
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      return Pep3AssessmentItem.empty;
    }
    return Pep3AssessmentItem.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3DraftPage> fetchDraftsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 1,
    int studentId = 0,
    bool latestOnly = true,
  }) async {
    final Object? data = await _postJson(
      _uri(draftsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'PEP3',
          if (studentId > 0) 'studentId': studentId,
          if (latestOnly) 'latestOnly': true,
        },
        if (latestOnly) 'latestOnly': true,
      },
    );
    if (data is! Map) {
      return const Pep3DraftPage(
        items: <Pep3DraftSummary>[],
        total: 0,
        current: 1,
        size: 0,
      );
    }
    return Pep3DraftPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3DraftDetail> fetchDraftDetail(String token, int id) async {
    final Uri uri =
        _uri(draftDetailPath).replace(queryParameters: <String, String>{
      'id': '$id',
    });
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      throw const Pep3ApiException('草稿详情返回格式不正确');
    }
    return Pep3DraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3DraftDetail> saveDraft(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data = await _postJson(_uri(draftSavePath), token, payload);
    if (data is! Map) {
      throw const Pep3ApiException('草稿保存返回格式不正确');
    }
    return Pep3DraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3DraftDetail> saveDraftItem(
    String token,
    Map<String, dynamic> payload,
  ) async {
    final Object? data =
        await _postJson(_uri(draftItemSavePath), token, payload);
    if (data is! Map) {
      throw const Pep3ApiException('单题保存返回格式不正确');
    }
    return Pep3DraftDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3CaregiverInvite> inviteCaregiverReport(
    String token,
    int draftId,
  ) async {
    final Object? data = await _postJson(
      _uri(caregiverInvitePath),
      token,
      <String, int>{'draftId': draftId},
    );
    if (data is! Map) {
      throw const Pep3ApiException('照护者报告入口返回格式不正确');
    }
    return Pep3CaregiverInvite.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<void> submitDraft(String token, int draftId) async {
    await _postJson(_uri(draftSubmitPath), token, <String, int>{'id': draftId});
  }

  @override
  Future<Pep3RecordPage> fetchRecordsPage(
    String token, {
    int pageIndex = 1,
    int pageSize = 5,
    int studentId = 0,
    String assessmentDateEnd = '',
  }) async {
    final Object? data = await _postJson(
      _uri(recordsPagePath),
      token,
      <String, dynamic>{
        'pageRequestModel': <String, int>{
          'pageIndex': pageIndex,
          'pageSize': pageSize,
        },
        'queryModel': <String, dynamic>{
          'assessmentCode': 'PEP3',
          if (studentId > 0) 'studentId': studentId,
          if (assessmentDateEnd.trim().isNotEmpty)
            'assessmentDateEnd': assessmentDateEnd.trim(),
        },
      },
    );
    if (data is! Map) {
      return const Pep3RecordPage(
        items: <Pep3RecordSummary>[],
        total: 0,
        current: 1,
        size: 0,
      );
    }
    return Pep3RecordPage.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Pep3RecordDetail> fetchRecordDetail(String token, int id) async {
    final Uri uri =
        _uri(recordDetailPath).replace(queryParameters: <String, String>{
      'id': '$id',
    });
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      throw const Pep3ApiException('测评记录详情返回格式不正确');
    }
    return Pep3RecordDetail.fromJson(Map<String, dynamic>.from(data));
  }

  Uri _uri(String path) {
    final String trimmedBase =
        educationBaseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
    final String normalizedPath = path.startsWith('/') ? path : '/$path';
    return Uri.parse('$trimmedBase$normalizedPath');
  }

  Future<Object?> _getJson(Uri uri, String token) async {
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 12),
          );
    } on TimeoutException {
      throw const Pep3ApiException('PEP-3接口响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接PEP-3接口：$error');
    }
    return _handleResponse(response);
  }

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            uri,
            headers: _headers(token),
            body: jsonEncode(payload),
          )
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const Pep3ApiException('PEP-3接口响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接PEP-3接口：$error');
    }
    return _handleResponse(response);
  }

  Map<String, String> _headers(String token) {
    return <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json; charset=utf-8',
      if (token.trim().isNotEmpty) 'Authorization': 'Bearer ${token.trim()}',
      if (token.trim().isNotEmpty) 'X-Access-Token': token.trim(),
    };
  }

  Object? _handleResponse(http.Response response) {
    final Object? decoded = _decodeResponse(response.body);
    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(_messageFromPayload(decoded) ?? 'PEP-3请求失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw Pep3ApiException(
          _messageFromPayload(envelope) ?? 'PEP-3请求失败',
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
      if (envelope.containsKey('result')) {
        return envelope['result'];
      }
    }
    return decoded;
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
  return null;
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
    return value.round();
  }
  return int.tryParse('${value ?? ''}') ?? 0;
}

double _doubleFrom(Object? value) {
  if (value is num) {
    return value.toDouble();
  }
  return double.tryParse('${value ?? ''}') ?? 0;
}

bool _isValidScore(int score) => score == 0 || score == 1 || score == 2;

String _displayItemTitle(String itemTitle, String testItem) {
  final String title = (itemTitle.trim().isNotEmpty ? itemTitle : testItem)
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();
  final String normalized =
      title.replaceFirst(RegExp(r'^[（(]?\d+[）)]?\s*[、.．-]?\s*'), '');
  return normalized.trim().isNotEmpty ? normalized.trim() : title;
}
