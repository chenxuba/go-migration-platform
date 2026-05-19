import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import 'erxin_assessment_client.dart';
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
const String defaultPep3RecordCategoryStatsPath = String.fromEnvironment(
  'PEP3_RECORD_CATEGORY_STATS_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/category-stats',
);
const String defaultPep3RecordDetailPath = String.fromEnvironment(
  'PEP3_RECORD_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/detail',
);
const String defaultPep3RecordConfigUpdatePath = String.fromEnvironment(
  'PEP3_RECORD_CONFIG_UPDATE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/config/update',
);
const String defaultPep3RecordBookletPdfPath = String.fromEnvironment(
  'PEP3_RECORD_BOOKLET_PDF_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/booklet/pdf',
);
const String defaultPep3RecordReportInterpretationPdfPath =
    String.fromEnvironment(
  'PEP3_RECORD_REPORT_INTERPRETATION_PDF_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/report/interpretation/pdf',
);
const String defaultPep3RecordReportInterpretationPath = String.fromEnvironment(
  'PEP3_RECORD_REPORT_INTERPRETATION_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/report/interpretation',
);
const String defaultPep3RecordReportInterpretationAiPath =
    String.fromEnvironment(
  'PEP3_RECORD_REPORT_INTERPRETATION_AI_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/report/interpretation/ai',
);
const String defaultPep3RecordReportInterpretationAiStreamPath =
    String.fromEnvironment(
  'PEP3_RECORD_REPORT_INTERPRETATION_AI_STREAM_PATH',
  defaultValue:
      '/api/v1/assessments/pep3/records/report/interpretation/ai/stream',
);
const String defaultAutismDevRecordProfilePdfPath = String.fromEnvironment(
  'AUTISMDEV_RECORD_PROFILE_PDF_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/profile/pdf',
);
const String defaultAutismDevRecordAssessmentInfoPdfPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_ASSESSMENT_INFO_PDF_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/assessment-info/pdf',
);
const String defaultAutismDevRecordResultAnalysisPath = String.fromEnvironment(
  'AUTISMDEV_RECORD_RESULT_ANALYSIS_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/result-analysis',
);
const String defaultAutismDevRecordResultAnalysisPdfPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_RESULT_ANALYSIS_PDF_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/result-analysis/pdf',
);
const String defaultAutismDevRecordSelectedReportPdfPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_SELECTED_REPORT_PDF_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/selected-report/pdf',
);
const String defaultAutismDevRecordResultAnalysisAiStreamPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_RESULT_ANALYSIS_AI_STREAM_PATH',
  defaultValue:
      '/api/v1/assessments/autismdev/records/result-analysis/ai/stream',
);
const String defaultAutismDevRecordReportInterpretationPdfPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_REPORT_INTERPRETATION_PDF_PATH',
  defaultValue:
      '/api/v1/assessments/autismdev/records/report/interpretation/pdf',
);
const String defaultAutismDevRecordReportInterpretationPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_REPORT_INTERPRETATION_PATH',
  defaultValue: '/api/v1/assessments/autismdev/records/report/interpretation',
);
const String defaultAutismDevRecordReportInterpretationAiPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_REPORT_INTERPRETATION_AI_PATH',
  defaultValue:
      '/api/v1/assessments/autismdev/records/report/interpretation/ai',
);
const String defaultAutismDevRecordReportInterpretationAiStreamPath =
    String.fromEnvironment(
  'AUTISMDEV_RECORD_REPORT_INTERPRETATION_AI_STREAM_PATH',
  defaultValue:
      '/api/v1/assessments/autismdev/records/report/interpretation/ai/stream',
);
const String defaultShuangxiRecordDevelopmentProfilePdfPath =
    String.fromEnvironment(
  'SHUANGXI_RECORD_DEVELOPMENT_PROFILE_PDF_PATH',
  defaultValue:
      '/api/v1/assessments/shuangxi-a/records/development-profile/pdf',
);
const String defaultShuangxiRecordResultAnalysisPath = String.fromEnvironment(
  'SHUANGXI_RECORD_RESULT_ANALYSIS_PATH',
  defaultValue: '/api/v1/assessments/shuangxi-a/records/result-analysis',
);
const String defaultShuangxiRecordResultAnalysisAiStreamPath =
    String.fromEnvironment(
  'SHUANGXI_RECORD_RESULT_ANALYSIS_AI_STREAM_PATH',
  defaultValue:
      '/api/v1/assessments/shuangxi-a/records/result-analysis/ai/stream',
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

class ShuangxiDevelopmentProfilePdfConfig {
  const ShuangxiDevelopmentProfilePdfConfig({
    this.showCompare = true,
    this.showScore = true,
    this.compareRecordIds = const <int>[],
  });

  final bool showCompare;
  final bool showScore;
  final List<int> compareRecordIds;
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
    required this.itemScoreLabels,
    required this.itemRecordValues,
  });

  factory Pep3DraftInput.fromJson(Map<String, dynamic> json) {
    final Map<int, int> itemScores = <int, int>{};
    final Map<int, String> itemScoreLabels = <int, String>{};
    _mapFrom(json['itemScores']).forEach((String key, dynamic value) {
      final int itemNo = int.tryParse(key) ?? 0;
      final String scoreLabel = '${value ?? ''}'.trim().toUpperCase();
      if (itemNo > 0 && scoreLabel.isNotEmpty) {
        itemScoreLabels[itemNo] = scoreLabel;
      }
      final int score = _intFrom(value);
      if (itemNo > 0 && _isValidScore(score)) {
        itemScores[itemNo] = score;
      }
    });
    for (final Map<String, dynamic> item in _listFrom(json['itemScoreList'])) {
      final int itemNo = _intFrom(item['itemNo']);
      final String scoreLabel = '${item['score'] ?? ''}'.trim().toUpperCase();
      if (itemNo > 0 && scoreLabel.isNotEmpty) {
        itemScoreLabels[itemNo] = scoreLabel;
      }
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
      itemScoreLabels: itemScoreLabels,
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
    itemScoreLabels: <int, String>{},
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
  final Map<int, String> itemScoreLabels;
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

class Pep3RecordCategoryStats {
  const Pep3RecordCategoryStats({
    required this.total,
    required this.categoryCounts,
  });

  factory Pep3RecordCategoryStats.fromJson(Map<String, dynamic> json) {
    return Pep3RecordCategoryStats(
      total: _intFrom(json['total']),
      categoryCounts: _stringIntMapFrom(json['categoryCounts']),
    );
  }

  static const Pep3RecordCategoryStats empty = Pep3RecordCategoryStats(
    total: 0,
    categoryCounts: <String, int>{},
  );

  final int total;
  final Map<String, int> categoryCounts;
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
  });

  factory Pep3RecordSummary.fromJson(Map<String, dynamic> json) {
    return Pep3RecordSummary(
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
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      ageYears: _intFrom(json['ageYears']),
      ageMonths: _intFrom(json['ageMonths']),
      ageDays: _intFrom(json['ageDays']),
      normAgeMonths: _intFrom(json['normAgeMonths']),
      assessmentSequence: _intFrom(json['assessmentSequence']),
      examinerName: '${json['examinerName'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
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
    super.studentGender,
    super.studentAvatar,
    super.studentPhone,
    super.scaleCategory,
    super.scaleVersion,
    super.ageYears,
    super.ageMonths,
    super.ageDays,
    super.normAgeMonths,
    super.assessmentSequence,
    super.createdTime,
    required this.input,
  });

  factory Pep3RecordDetail.fromJson(Map<String, dynamic> json) {
    return Pep3RecordDetail(
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
      birthDate: '${json['birthDate'] ?? ''}',
      assessmentDate: '${json['assessmentDate'] ?? ''}',
      ageYears: _intFrom(json['ageYears']),
      ageMonths: _intFrom(json['ageMonths']),
      ageDays: _intFrom(json['ageDays']),
      normAgeMonths: _intFrom(json['normAgeMonths']),
      assessmentSequence: _intFrom(json['assessmentSequence']),
      examinerName: '${json['examinerName'] ?? ''}',
      createdTime: '${json['createdTime'] ?? ''}',
      updatedTime: '${json['updatedTime'] ?? ''}',
      input: Pep3DraftInput.fromJson(_mapFrom(json['input'])),
    );
  }

  final Pep3DraftInput input;
}

class AutismDevResultAnalysis {
  const AutismDevResultAnalysis({
    required this.title,
    required this.rows,
    this.model = '',
    this.generatedBy = '',
    this.generatedAt = '',
  });

  factory AutismDevResultAnalysis.fromJson(Map<String, dynamic> json) {
    return AutismDevResultAnalysis(
      title: '${json['title'] ?? ''}',
      model: '${json['model'] ?? ''}',
      generatedBy: '${json['generatedBy'] ?? ''}',
      generatedAt: '${json['generatedAt'] ?? ''}',
      rows: _listFrom(json['rows'])
          .map(AutismDevResultAnalysisRow.fromJson)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      if (model.trim().isNotEmpty) 'model': model,
      if (generatedBy.trim().isNotEmpty) 'generatedBy': generatedBy,
      if (generatedAt.trim().isNotEmpty) 'generatedAt': generatedAt,
      'rows':
          rows.map((AutismDevResultAnalysisRow row) => row.toJson()).toList(),
    };
  }

  static const AutismDevResultAnalysis empty = AutismDevResultAnalysis(
    title: '',
    rows: <AutismDevResultAnalysisRow>[],
  );

  final String title;
  final String model;
  final String generatedBy;
  final String generatedAt;
  final List<AutismDevResultAnalysisRow> rows;

  bool get isEmpty => rows.every((AutismDevResultAnalysisRow row) =>
      row.status.trim().isEmpty &&
      row.strengths.trim().isEmpty &&
      row.weaknesses.trim().isEmpty &&
      row.targets.trim().isEmpty);
}

class AutismDevResultAnalysisRow {
  const AutismDevResultAnalysisRow({
    required this.domain,
    this.status = '',
    this.strengths = '',
    this.weaknesses = '',
    this.targets = '',
  });

  factory AutismDevResultAnalysisRow.fromJson(Map<String, dynamic> json) {
    return AutismDevResultAnalysisRow(
      domain: '${json['domain'] ?? ''}',
      status: '${json['status'] ?? ''}',
      strengths: '${json['strengths'] ?? ''}',
      weaknesses: '${json['weaknesses'] ?? ''}',
      targets: '${json['targets'] ?? ''}',
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'domain': domain,
      'status': status,
      'strengths': strengths,
      'weaknesses': weaknesses,
      'targets': targets,
    };
  }

  final String domain;
  final String status;
  final String strengths;
  final String weaknesses;
  final String targets;

  AutismDevResultAnalysisRow copyWith({
    String? domain,
    String? status,
    String? strengths,
    String? weaknesses,
    String? targets,
  }) {
    return AutismDevResultAnalysisRow(
      domain: domain ?? this.domain,
      status: status ?? this.status,
      strengths: strengths ?? this.strengths,
      weaknesses: weaknesses ?? this.weaknesses,
      targets: targets ?? this.targets,
    );
  }
}

class AutismDevResultAnalysisStreamEvent {
  const AutismDevResultAnalysisStreamEvent({
    required this.type,
    this.message = '',
    this.text = '',
    this.data,
  });

  final String type;
  final String message;
  final String text;
  final AutismDevResultAnalysis? data;
}

class ShuangxiResultAnalysis {
  const ShuangxiResultAnalysis({
    required this.title,
    required this.rows,
    this.courseName = '',
    this.model = '',
    this.generatedBy = '',
    this.generatedAt = '',
  });

  factory ShuangxiResultAnalysis.fromJson(Map<String, dynamic> json) {
    return ShuangxiResultAnalysis(
      title: '${json['title'] ?? ''}',
      courseName: '${json['courseName'] ?? ''}',
      model: '${json['model'] ?? ''}',
      generatedBy: '${json['generatedBy'] ?? ''}',
      generatedAt: '${json['generatedAt'] ?? ''}',
      rows: _listFrom(json['rows'])
          .map(ShuangxiResultAnalysisRow.fromJson)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      if (courseName.trim().isNotEmpty) 'courseName': courseName,
      if (model.trim().isNotEmpty) 'model': model,
      if (generatedBy.trim().isNotEmpty) 'generatedBy': generatedBy,
      if (generatedAt.trim().isNotEmpty) 'generatedAt': generatedAt,
      'rows':
          rows.map((ShuangxiResultAnalysisRow row) => row.toJson()).toList(),
    };
  }

  static const ShuangxiResultAnalysis empty = ShuangxiResultAnalysis(
    title: '',
    rows: <ShuangxiResultAnalysisRow>[],
  );

  final String title;
  final String courseName;
  final String model;
  final String generatedBy;
  final String generatedAt;
  final List<ShuangxiResultAnalysisRow> rows;

  bool get isEmpty => rows.every((ShuangxiResultAnalysisRow row) =>
      row.strengths.trim().isEmpty &&
      row.weaknesses.trim().isEmpty &&
      row.reason.trim().isEmpty &&
      row.strategy.trim().isEmpty);

  ShuangxiResultAnalysis copyWith({
    String? title,
    String? courseName,
    String? model,
    String? generatedBy,
    String? generatedAt,
    List<ShuangxiResultAnalysisRow>? rows,
  }) {
    return ShuangxiResultAnalysis(
      title: title ?? this.title,
      courseName: courseName ?? this.courseName,
      model: model ?? this.model,
      generatedBy: generatedBy ?? this.generatedBy,
      generatedAt: generatedAt ?? this.generatedAt,
      rows: rows ?? this.rows,
    );
  }
}

class ShuangxiResultAnalysisRow {
  const ShuangxiResultAnalysisRow({
    required this.domain,
    this.domainCode = '',
    this.strengths = '',
    this.weaknesses = '',
    this.reason = '',
    this.strategy = '',
  });

  factory ShuangxiResultAnalysisRow.fromJson(Map<String, dynamic> json) {
    return ShuangxiResultAnalysisRow(
      domainCode: '${json['domainCode'] ?? ''}',
      domain: '${json['domain'] ?? ''}',
      strengths: '${json['strengths'] ?? ''}',
      weaknesses: '${json['weaknesses'] ?? ''}',
      reason: '${json['reason'] ?? ''}',
      strategy: '${json['strategy'] ?? ''}',
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      if (domainCode.trim().isNotEmpty) 'domainCode': domainCode,
      'domain': domain,
      'strengths': strengths,
      'weaknesses': weaknesses,
      'reason': reason,
      'strategy': strategy,
    };
  }

  final String domainCode;
  final String domain;
  final String strengths;
  final String weaknesses;
  final String reason;
  final String strategy;

  ShuangxiResultAnalysisRow copyWith({
    String? domainCode,
    String? domain,
    String? strengths,
    String? weaknesses,
    String? reason,
    String? strategy,
  }) {
    return ShuangxiResultAnalysisRow(
      domainCode: domainCode ?? this.domainCode,
      domain: domain ?? this.domain,
      strengths: strengths ?? this.strengths,
      weaknesses: weaknesses ?? this.weaknesses,
      reason: reason ?? this.reason,
      strategy: strategy ?? this.strategy,
    );
  }
}

class ShuangxiResultAnalysisStreamEvent {
  const ShuangxiResultAnalysisStreamEvent({
    required this.type,
    this.message = '',
    this.text = '',
    this.data,
  });

  final String type;
  final String message;
  final String text;
  final ShuangxiResultAnalysis? data;
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
    String assessmentCode = 'PEP3',
    String scaleCategory = '',
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  });

  Future<Pep3RecordCategoryStats> fetchRecordCategoryStats(
    String token, {
    int studentId = 0,
    String assessmentCode = 'PEP3',
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  });

  Future<Pep3RecordDetail> fetchRecordDetail(String token, int id);

  Future<Pep3RecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  });

  Future<Uint8List> downloadRecordBookletPdf(
    String token,
    int id, {
    String dimension = 'score_and_profile',
  });

  Future<Uint8List> downloadRecordReportInterpretationPdf(
    String token,
    int id,
  );

  Future<Uint8List> downloadAutismDevRecordProfilePdf(
    String token,
    int id, {
    required String profile,
  });

  Future<Uint8List> downloadAutismDevAssessmentInfoPdf(
    String token,
    int id,
  );

  Future<AutismDevResultAnalysis> fetchAutismDevResultAnalysis(
    String token,
    int id,
  );

  Future<AutismDevResultAnalysis> saveAutismDevResultAnalysis(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  );

  Future<Uint8List> downloadAutismDevResultAnalysisPdf(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  );

  Future<Uint8List> downloadAutismDevSelectedReportPdf(
    String token,
    int id, {
    required List<String> sections,
    AutismDevResultAnalysis? analysis,
  });

  Future<Uint8List> downloadAutismDevRecordReportInterpretationPdf(
    String token,
    int id,
  );

  Future<Uint8List> downloadShuangxiDevelopmentProfilePdf(
    String token,
    int id, {
    ShuangxiDevelopmentProfilePdfConfig config =
        const ShuangxiDevelopmentProfilePdfConfig(),
  });

  Future<ShuangxiResultAnalysis> fetchShuangxiResultAnalysis(
    String token,
    int id,
  );

  Future<ShuangxiResultAnalysis> saveShuangxiResultAnalysis(
    String token,
    int id,
    ShuangxiResultAnalysis analysis,
  );

  Stream<ShuangxiResultAnalysisStreamEvent>
      generateShuangxiResultAnalysisStream(
    String token,
    int id,
  );

  Stream<AutismDevResultAnalysisStreamEvent>
      generateAutismDevResultAnalysisStream(
    String token,
    int id,
  );

  Future<ErxinReportInterpretation> fetchAutismDevRecordReportInterpretation(
    String token,
    int id,
  );

  Future<ErxinReportInterpretation> generateAutismDevRecordReportInterpretation(
    String token,
    int id,
  );

  Stream<ErxinReportInterpretationStreamEvent>
      generateAutismDevRecordReportInterpretationStream(
    String token,
    int id,
  );

  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  );

  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  );

  Stream<ErxinReportInterpretationStreamEvent>
      generateRecordReportInterpretationStream(
    String token,
    int id,
  );
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
    this.recordCategoryStatsPath = defaultPep3RecordCategoryStatsPath,
    this.recordDetailPath = defaultPep3RecordDetailPath,
    this.recordConfigUpdatePath = defaultPep3RecordConfigUpdatePath,
    this.recordBookletPdfPath = defaultPep3RecordBookletPdfPath,
    this.recordReportInterpretationPdfPath =
        defaultPep3RecordReportInterpretationPdfPath,
    this.recordReportInterpretationPath =
        defaultPep3RecordReportInterpretationPath,
    this.recordReportInterpretationAiPath =
        defaultPep3RecordReportInterpretationAiPath,
    this.recordReportInterpretationAiStreamPath =
        defaultPep3RecordReportInterpretationAiStreamPath,
    this.autismDevRecordProfilePdfPath = defaultAutismDevRecordProfilePdfPath,
    this.autismDevRecordAssessmentInfoPdfPath =
        defaultAutismDevRecordAssessmentInfoPdfPath,
    this.autismDevRecordResultAnalysisPath =
        defaultAutismDevRecordResultAnalysisPath,
    this.autismDevRecordResultAnalysisPdfPath =
        defaultAutismDevRecordResultAnalysisPdfPath,
    this.autismDevRecordSelectedReportPdfPath =
        defaultAutismDevRecordSelectedReportPdfPath,
    this.autismDevRecordResultAnalysisAiStreamPath =
        defaultAutismDevRecordResultAnalysisAiStreamPath,
    this.autismDevRecordReportInterpretationPdfPath =
        defaultAutismDevRecordReportInterpretationPdfPath,
    this.autismDevRecordReportInterpretationPath =
        defaultAutismDevRecordReportInterpretationPath,
    this.autismDevRecordReportInterpretationAiPath =
        defaultAutismDevRecordReportInterpretationAiPath,
    this.autismDevRecordReportInterpretationAiStreamPath =
        defaultAutismDevRecordReportInterpretationAiStreamPath,
    this.shuangxiRecordDevelopmentProfilePdfPath =
        defaultShuangxiRecordDevelopmentProfilePdfPath,
    this.shuangxiRecordResultAnalysisPath =
        defaultShuangxiRecordResultAnalysisPath,
    this.shuangxiRecordResultAnalysisAiStreamPath =
        defaultShuangxiRecordResultAnalysisAiStreamPath,
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
  final String recordCategoryStatsPath;
  final String recordDetailPath;
  final String recordConfigUpdatePath;
  final String recordBookletPdfPath;
  final String recordReportInterpretationPdfPath;
  final String recordReportInterpretationPath;
  final String recordReportInterpretationAiPath;
  final String recordReportInterpretationAiStreamPath;
  final String autismDevRecordProfilePdfPath;
  final String autismDevRecordAssessmentInfoPdfPath;
  final String autismDevRecordResultAnalysisPath;
  final String autismDevRecordResultAnalysisPdfPath;
  final String autismDevRecordSelectedReportPdfPath;
  final String autismDevRecordResultAnalysisAiStreamPath;
  final String autismDevRecordReportInterpretationPdfPath;
  final String autismDevRecordReportInterpretationPath;
  final String autismDevRecordReportInterpretationAiPath;
  final String autismDevRecordReportInterpretationAiStreamPath;
  final String shuangxiRecordDevelopmentProfilePdfPath;
  final String shuangxiRecordResultAnalysisPath;
  final String shuangxiRecordResultAnalysisAiStreamPath;

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
    String assessmentCode = 'PEP3',
    String scaleCategory = '',
    String searchKey = '',
    String assessmentDateBegin = '',
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
          if (assessmentCode.trim().isNotEmpty)
            'assessmentCode': assessmentCode.trim(),
          if (scaleCategory.trim().isNotEmpty)
            'scaleCategory': scaleCategory.trim(),
          if (studentId > 0) 'studentId': studentId,
          if (searchKey.trim().isNotEmpty) 'searchKey': searchKey.trim(),
          if (assessmentDateBegin.trim().isNotEmpty)
            'assessmentDateBegin': assessmentDateBegin.trim(),
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
  Future<Pep3RecordCategoryStats> fetchRecordCategoryStats(
    String token, {
    int studentId = 0,
    String assessmentCode = 'PEP3',
    String searchKey = '',
    String assessmentDateBegin = '',
    String assessmentDateEnd = '',
  }) async {
    final Object? data = await _postJson(
      _uri(recordCategoryStatsPath),
      token,
      <String, dynamic>{
        'pageRequestModel': const <String, int>{
          'pageIndex': 1,
          'pageSize': 1,
        },
        'queryModel': <String, dynamic>{
          if (assessmentCode.trim().isNotEmpty)
            'assessmentCode': assessmentCode.trim(),
          if (studentId > 0) 'studentId': studentId,
          if (searchKey.trim().isNotEmpty) 'searchKey': searchKey.trim(),
          if (assessmentDateBegin.trim().isNotEmpty)
            'assessmentDateBegin': assessmentDateBegin.trim(),
          if (assessmentDateEnd.trim().isNotEmpty)
            'assessmentDateEnd': assessmentDateEnd.trim(),
        },
      },
    );
    if (data is! Map) {
      return Pep3RecordCategoryStats.empty;
    }
    return Pep3RecordCategoryStats.fromJson(Map<String, dynamic>.from(data));
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

  @override
  Future<Pep3RecordDetail> updateRecordConfig(
    String token,
    int id, {
    required String examinerName,
    required String assessmentDate,
  }) async {
    final Object? data = await _postJson(
      _uri(recordConfigUpdatePath),
      token,
      <String, dynamic>{
        'id': id,
        'examinerName': examinerName.trim(),
        'assessmentDate': assessmentDate.trim(),
      },
    );
    if (data is! Map) {
      throw const Pep3ApiException('评估配置返回格式不正确');
    }
    return Pep3RecordDetail.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Uint8List> downloadRecordBookletPdf(
    String token,
    int id, {
    String dimension = 'score_and_profile',
  }) async {
    final Uri uri = _uri(recordBookletPdfPath).replace(
      queryParameters: <String, String>{
        'id': '$id',
        'dimension':
            dimension.trim().isEmpty ? 'score_and_profile' : dimension.trim(),
      },
    );
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 20),
          );
    } on TimeoutException {
      throw const Pep3ApiException('评估报告PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接评估报告PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '评估报告PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadRecordReportInterpretationPdf(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(recordReportInterpretationPdfPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 20),
          );
    } on TimeoutException {
      throw const Pep3ApiException('报告解读PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '报告解读PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadAutismDevRecordProfilePdf(
    String token,
    int id, {
    required String profile,
  }) async {
    final String normalizedProfile =
        profile.trim().isEmpty ? 'development' : profile.trim().toLowerCase();
    final Uri uri = _uri(autismDevRecordProfilePdfPath).replace(
      queryParameters: <String, String>{
        'id': '$id',
        'profile': normalizedProfile,
      },
    );
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 20),
          );
    } on TimeoutException {
      throw const Pep3ApiException('孤独症报告PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接孤独症报告PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '孤独症报告PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadAutismDevAssessmentInfoPdf(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(autismDevRecordAssessmentInfoPdfPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final http.Response response;
    try {
      response = await http.get(
        uri,
        headers: <String, String>{
          ..._headers(token),
          'Accept': 'application/pdf, application/octet-stream, */*',
        },
      ).timeout(const Duration(seconds: 60));
    } on TimeoutException {
      throw const Pep3ApiException('评估情况PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接评估情况PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '评估情况PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<AutismDevResultAnalysis> fetchAutismDevResultAnalysis(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(autismDevRecordResultAnalysisPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      return AutismDevResultAnalysis.empty;
    }
    return AutismDevResultAnalysis.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<AutismDevResultAnalysis> saveAutismDevResultAnalysis(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  ) async {
    final Object? data = await _postJson(
      _uri(autismDevRecordResultAnalysisPath),
      token,
      <String, dynamic>{
        'id': id,
        'analysis': analysis.toJson(),
      },
    );
    if (data is! Map) {
      throw const Pep3ApiException('评估结果分析保存返回格式不正确');
    }
    return AutismDevResultAnalysis.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<Uint8List> downloadAutismDevResultAnalysisPdf(
    String token,
    int id,
    AutismDevResultAnalysis analysis,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            _uri(autismDevRecordResultAnalysisPdfPath),
            headers: <String, String>{
              ..._headers(token),
              'Accept': 'application/pdf, application/octet-stream, */*',
            },
            body: jsonEncode(<String, dynamic>{
              'id': id,
              'analysis': analysis.toJson(),
            }),
          )
          .timeout(const Duration(seconds: 60));
    } on TimeoutException {
      throw const Pep3ApiException('评估结果分析PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接评估结果分析PDF接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = utf8.decode(response.bodyBytes).trim();
      throw Pep3ApiException(
        _messageFromPayload(
                body.isEmpty ? null : await _decodeResponse(body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = utf8.decode(response.bodyBytes).trim();
      throw Pep3ApiException(
        _messageFromPayload(
                body.isEmpty ? null : await _decodeResponse(body)) ??
            '评估结果分析PDF生成失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadAutismDevSelectedReportPdf(
    String token,
    int id, {
    required List<String> sections,
    AutismDevResultAnalysis? analysis,
  }) async {
    final http.Response response;
    try {
      response = await http
          .post(
            _uri(autismDevRecordSelectedReportPdfPath),
            headers: <String, String>{
              ..._headers(token),
              'Accept': 'application/pdf, application/octet-stream, */*',
            },
            body: jsonEncode(<String, dynamic>{
              'id': id,
              'sections': sections,
              if (analysis != null) 'analysis': analysis.toJson(),
            }),
          )
          .timeout(const Duration(seconds: 90));
    } on TimeoutException {
      throw const Pep3ApiException('评估报告PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接评估报告PDF接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = utf8.decode(response.bodyBytes).trim();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = utf8.decode(response.bodyBytes).trim();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '评估报告PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadAutismDevRecordReportInterpretationPdf(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(autismDevRecordReportInterpretationPdfPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 20),
          );
    } on TimeoutException {
      throw const Pep3ApiException('报告解读PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '报告解读PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<Uint8List> downloadShuangxiDevelopmentProfilePdf(
    String token,
    int id, {
    ShuangxiDevelopmentProfilePdfConfig config =
        const ShuangxiDevelopmentProfilePdfConfig(),
  }) async {
    final List<int> compareRecordIds = config.compareRecordIds
        .where((int value) => value > 0)
        .toSet()
        .toList(growable: false);
    final Uri uri = _uri(shuangxiRecordDevelopmentProfilePdfPath).replace(
      queryParameters: <String, String>{
        'id': '$id',
        'showCompare': '${config.showCompare}',
        'showScore': '${config.showScore}',
        if (compareRecordIds.isNotEmpty)
          'compareRecordIds': compareRecordIds.join(','),
      },
    );
    final http.Response response;
    try {
      response = await http.get(uri, headers: _headers(token)).timeout(
            const Duration(seconds: 20),
          );
    } on TimeoutException {
      throw const Pep3ApiException('双溪发展侧面图PDF响应超时，请检查网络');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接双溪发展侧面图PDF接口：$error');
    }

    if (response.statusCode == 401 || response.statusCode == 403) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(response.body)) ??
            '双溪发展侧面图PDF加载失败',
      );
    }
    return _normalizeReportPdfBytes(response.bodyBytes);
  }

  @override
  Future<ShuangxiResultAnalysis> fetchShuangxiResultAnalysis(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(shuangxiRecordResultAnalysisPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      return ShuangxiResultAnalysis.empty;
    }
    return ShuangxiResultAnalysis.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ShuangxiResultAnalysis> saveShuangxiResultAnalysis(
    String token,
    int id,
    ShuangxiResultAnalysis analysis,
  ) async {
    final Object? data = await _postJson(
      _uri(shuangxiRecordResultAnalysisPath),
      token,
      <String, dynamic>{
        'id': id,
        'analysis': analysis.toJson(),
      },
    );
    if (data is! Map) {
      throw const Pep3ApiException('双溪评量结果分析保存返回格式不正确');
    }
    return ShuangxiResultAnalysis.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Stream<ShuangxiResultAnalysisStreamEvent>
      generateShuangxiResultAnalysisStream(
    String token,
    int id,
  ) async* {
    final http.Request request = http.Request(
      'POST',
      _uri(shuangxiRecordResultAnalysisAiStreamPath),
    )
      ..headers.addAll(<String, String>{
        ..._headers(token),
        'Accept': 'text/event-stream',
      })
      ..body = jsonEncode(<String, int>{'id': id});
    final http.StreamedResponse response;
    try {
      response = await request.send().timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const Pep3ApiException('双溪评量结果分析生成连接超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接双溪评量结果分析流式接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '双溪评量结果分析生成失败',
      );
    }

    await for (final ShuangxiResultAnalysisStreamEvent event
        in _decodeShuangxiResultAnalysisSse(response.stream)) {
      yield event;
    }
  }

  @override
  Stream<AutismDevResultAnalysisStreamEvent>
      generateAutismDevResultAnalysisStream(String token, int id) async* {
    final http.Request request = http.Request(
      'POST',
      _uri(autismDevRecordResultAnalysisAiStreamPath),
    )
      ..headers.addAll(<String, String>{
        ..._headers(token),
        'Accept': 'text/event-stream',
      })
      ..body = jsonEncode(<String, int>{'id': id});
    final http.StreamedResponse response;
    try {
      response = await request.send().timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const Pep3ApiException('评估结果分析生成连接超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接评估结果分析流式接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '评估结果分析生成失败',
      );
    }

    await for (final AutismDevResultAnalysisStreamEvent event
        in _decodeAutismDevResultAnalysisSse(response.stream)) {
      yield event;
    }
  }

  @override
  Future<ErxinReportInterpretation> fetchAutismDevRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(autismDevRecordReportInterpretationPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      throw const Pep3ApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinReportInterpretation> generateAutismDevRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            _uri(autismDevRecordReportInterpretationAiPath),
            headers: _headers(token),
            body: jsonEncode(<String, int>{'id': id}),
          )
          .timeout(const Duration(seconds: 190));
    } on TimeoutException {
      throw const Pep3ApiException('报告解读生成超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读接口：$error');
    }
    final Object? data = await _handleResponse(response);
    if (data is! Map) {
      throw const Pep3ApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Stream<ErxinReportInterpretationStreamEvent>
      generateAutismDevRecordReportInterpretationStream(
    String token,
    int id,
  ) async* {
    final http.Request request = http.Request(
      'POST',
      _uri(autismDevRecordReportInterpretationAiStreamPath),
    )
      ..headers.addAll(<String, String>{
        ..._headers(token),
        'Accept': 'text/event-stream',
      })
      ..body = jsonEncode(<String, int>{'id': id});
    final http.StreamedResponse response;
    try {
      response = await request.send().timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const Pep3ApiException('报告解读生成连接超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读流式接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '报告解读生成失败',
      );
    }

    await for (final ErxinReportInterpretationStreamEvent event
        in _decodePep3ReportInterpretationSse(response.stream)) {
      yield event;
    }
  }

  @override
  Future<ErxinReportInterpretation> fetchRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final Uri uri = _uri(recordReportInterpretationPath).replace(
      queryParameters: <String, String>{'id': '$id'},
    );
    final Object? data = await _getJson(uri, token);
    if (data is! Map) {
      throw const Pep3ApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<ErxinReportInterpretation> generateRecordReportInterpretation(
    String token,
    int id,
  ) async {
    final http.Response response;
    try {
      response = await http
          .post(
            _uri(recordReportInterpretationAiPath),
            headers: _headers(token),
            body: jsonEncode(<String, int>{'id': id}),
          )
          .timeout(const Duration(seconds: 190));
    } on TimeoutException {
      throw const Pep3ApiException('报告解读生成超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读接口：$error');
    }
    final Object? data = await _handleResponse(response);
    if (data is! Map) {
      throw const Pep3ApiException('报告解读返回格式不正确');
    }
    return ErxinReportInterpretation.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Stream<ErxinReportInterpretationStreamEvent>
      generateRecordReportInterpretationStream(String token, int id) async* {
    final http.Request request = http.Request(
      'POST',
      _uri(recordReportInterpretationAiStreamPath),
    )
      ..headers.addAll(<String, String>{
        ..._headers(token),
        'Accept': 'text/event-stream',
      })
      ..body = jsonEncode(<String, int>{'id': id});
    final http.StreamedResponse response;
    try {
      response = await request.send().timeout(const Duration(seconds: 20));
    } on TimeoutException {
      throw const Pep3ApiException('报告解读生成连接超时，请稍后重试');
    } on Object catch (error) {
      throw Pep3ApiException('无法连接报告解读流式接口：$error');
    }
    if (response.statusCode == 401 || response.statusCode == 403) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final String body = await response.stream.bytesToString();
      throw Pep3ApiException(
        _messageFromPayload(await _decodeResponse(body)) ?? '报告解读生成失败',
      );
    }

    await for (final ErxinReportInterpretationStreamEvent event
        in _decodePep3ReportInterpretationSse(response.stream)) {
      yield event;
    }
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

  Future<Object?> _handleResponse(http.Response response) async {
    final Object? decoded = await _decodeResponse(response.body);
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

Stream<ErxinReportInterpretationStreamEvent> _decodePep3ReportInterpretationSse(
  Stream<List<int>> byteStream,
) async* {
  final Stream<String> lines =
      byteStream.transform(utf8.decoder).transform(const LineSplitter());
  String eventName = 'message';
  final StringBuffer dataBuffer = StringBuffer();

  ErxinReportInterpretationStreamEvent? parseEvent(String event, String data) {
    final String trimmed = data.trim();
    if (trimmed.isEmpty) {
      return null;
    }
    final Object? decoded = jsonDecode(trimmed);
    if (decoded is! Map) {
      return null;
    }
    final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
    final String type = '${payload['type'] ?? event}'.trim();
    switch (type) {
      case 'status':
        return ErxinReportInterpretationStreamEvent(
          type: 'status',
          message: '${payload['message'] ?? ''}',
        );
      case 'delta':
        return ErxinReportInterpretationStreamEvent(
          type: 'delta',
          text: '${payload['text'] ?? ''}',
        );
      case 'done':
        final Object? data = payload['data'];
        return ErxinReportInterpretationStreamEvent(
          type: 'done',
          data: data is Map
              ? ErxinReportInterpretation.fromJson(
                  Map<String, dynamic>.from(data),
                )
              : ErxinReportInterpretation.empty,
        );
      case 'error':
        return ErxinReportInterpretationStreamEvent(
          type: 'error',
          message: '${payload['message'] ?? '报告解读生成失败'}',
        );
      default:
        return ErxinReportInterpretationStreamEvent(
          type: type.isEmpty ? event : type,
          message: '${payload['message'] ?? ''}',
          text: '${payload['text'] ?? ''}',
        );
    }
  }

  await for (final String rawLine in lines) {
    final String line = rawLine.trimRight();
    if (line.isEmpty) {
      final ErxinReportInterpretationStreamEvent? event =
          parseEvent(eventName, dataBuffer.toString());
      if (event != null) {
        yield event;
      }
      eventName = 'message';
      dataBuffer.clear();
      continue;
    }
    if (line.startsWith(':')) {
      continue;
    }
    if (line.startsWith('event:')) {
      eventName = line.substring(6).trim();
      continue;
    }
    if (line.startsWith('data:')) {
      if (dataBuffer.isNotEmpty) {
        dataBuffer.write('\n');
      }
      dataBuffer.write(line.substring(5).trimLeft());
    }
  }
  final ErxinReportInterpretationStreamEvent? event =
      parseEvent(eventName, dataBuffer.toString());
  if (event != null) {
    yield event;
  }
}

Stream<AutismDevResultAnalysisStreamEvent> _decodeAutismDevResultAnalysisSse(
  Stream<List<int>> byteStream,
) async* {
  final Stream<String> lines =
      byteStream.transform(utf8.decoder).transform(const LineSplitter());
  String eventName = 'message';
  final StringBuffer dataBuffer = StringBuffer();

  AutismDevResultAnalysisStreamEvent? parseEvent(String event, String data) {
    final String trimmed = data.trim();
    if (trimmed.isEmpty) {
      return null;
    }
    final Object? decoded = jsonDecode(trimmed);
    if (decoded is! Map) {
      return null;
    }
    final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
    final String type = '${payload['type'] ?? event}'.trim();
    switch (type) {
      case 'status':
        return AutismDevResultAnalysisStreamEvent(
          type: 'status',
          message: '${payload['message'] ?? ''}',
        );
      case 'delta':
        return AutismDevResultAnalysisStreamEvent(
          type: 'delta',
          text: '${payload['text'] ?? ''}',
        );
      case 'done':
        final Object? data = payload['data'];
        return AutismDevResultAnalysisStreamEvent(
          type: 'done',
          data: data is Map
              ? AutismDevResultAnalysis.fromJson(
                  Map<String, dynamic>.from(data),
                )
              : AutismDevResultAnalysis.empty,
        );
      case 'error':
        return AutismDevResultAnalysisStreamEvent(
          type: 'error',
          message: '${payload['message'] ?? '评估结果分析生成失败'}',
        );
      default:
        return AutismDevResultAnalysisStreamEvent(
          type: type.isEmpty ? event : type,
          message: '${payload['message'] ?? ''}',
          text: '${payload['text'] ?? ''}',
        );
    }
  }

  await for (final String rawLine in lines) {
    final String line = rawLine.trimRight();
    if (line.isEmpty) {
      final AutismDevResultAnalysisStreamEvent? event =
          parseEvent(eventName, dataBuffer.toString());
      if (event != null) {
        yield event;
      }
      eventName = 'message';
      dataBuffer.clear();
      continue;
    }
    if (line.startsWith(':')) {
      continue;
    }
    if (line.startsWith('event:')) {
      eventName = line.substring(6).trim();
      continue;
    }
    if (line.startsWith('data:')) {
      if (dataBuffer.isNotEmpty) {
        dataBuffer.write('\n');
      }
      dataBuffer.write(line.substring(5).trimLeft());
    }
  }
  final AutismDevResultAnalysisStreamEvent? event =
      parseEvent(eventName, dataBuffer.toString());
  if (event != null) {
    yield event;
  }
}

Stream<ShuangxiResultAnalysisStreamEvent> _decodeShuangxiResultAnalysisSse(
  Stream<List<int>> byteStream,
) async* {
  final Stream<String> lines =
      byteStream.transform(utf8.decoder).transform(const LineSplitter());
  String eventName = 'message';
  final StringBuffer dataBuffer = StringBuffer();

  ShuangxiResultAnalysisStreamEvent? parseEvent(String event, String data) {
    final String trimmed = data.trim();
    if (trimmed.isEmpty) {
      return null;
    }
    final Object? decoded = jsonDecode(trimmed);
    if (decoded is! Map) {
      return null;
    }
    final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
    final String type = '${payload['type'] ?? event}'.trim();
    switch (type) {
      case 'status':
        return ShuangxiResultAnalysisStreamEvent(
          type: 'status',
          message: '${payload['message'] ?? ''}',
        );
      case 'delta':
        return ShuangxiResultAnalysisStreamEvent(
          type: 'delta',
          text: '${payload['text'] ?? ''}',
        );
      case 'done':
        final Object? data = payload['data'];
        return ShuangxiResultAnalysisStreamEvent(
          type: 'done',
          data: data is Map
              ? ShuangxiResultAnalysis.fromJson(
                  Map<String, dynamic>.from(data),
                )
              : ShuangxiResultAnalysis.empty,
        );
      case 'error':
        return ShuangxiResultAnalysisStreamEvent(
          type: 'error',
          message: '${payload['message'] ?? '双溪评量结果分析生成失败'}',
        );
      default:
        return ShuangxiResultAnalysisStreamEvent(
          type: type.isEmpty ? event : type,
          message: '${payload['message'] ?? ''}',
          text: '${payload['text'] ?? ''}',
        );
    }
  }

  await for (final String rawLine in lines) {
    final String line = rawLine.trimRight();
    if (line.isEmpty) {
      final ShuangxiResultAnalysisStreamEvent? event =
          parseEvent(eventName, dataBuffer.toString());
      if (event != null) {
        yield event;
      }
      eventName = 'message';
      dataBuffer.clear();
      continue;
    }
    if (line.startsWith(':')) {
      continue;
    }
    if (line.startsWith('event:')) {
      eventName = line.substring(6).trim();
      continue;
    }
    if (line.startsWith('data:')) {
      if (dataBuffer.isNotEmpty) {
        dataBuffer.write('\n');
      }
      dataBuffer.write(line.substring(5).trimLeft());
    }
  }
  final ShuangxiResultAnalysisStreamEvent? event =
      parseEvent(eventName, dataBuffer.toString());
  if (event != null) {
    yield event;
  }
}

const int _pep3BackgroundDecodeThreshold = 24 * 1024;

Future<Object?> _decodeResponse(String body) async {
  if (body.trim().isEmpty) {
    return null;
  }
  try {
    if (body.length >= _pep3BackgroundDecodeThreshold) {
      return await compute(_decodePep3JsonPayload, body);
    }
    return _decodePep3JsonPayload(body);
  } on FormatException {
    return body;
  }
}

Object? _decodePep3JsonPayload(String body) => jsonDecode(body);

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    final Object? message = payload['message'] ?? payload['msg'];
    if (message != null && '$message'.trim().isNotEmpty) {
      return '$message';
    }
  }
  return null;
}

const List<int> _pep3InvalidPdfBinaryMarker = <int>[
  0x25,
  0x50,
  0x44,
  0x46,
  0x2d,
  0x31,
  0x2e,
  0x37,
  0x0a,
  0x25,
  0xef,
  0xbf,
  0xbd,
  0xef,
  0xbf,
  0xbd,
  0xef,
  0xbf,
  0xbd,
  0xef,
  0xbf,
  0xbd,
  0x0a,
  0x0a,
];

const List<int> _pep3ValidPdfBinaryMarker = <int>[
  0x25,
  0x50,
  0x44,
  0x46,
  0x2d,
  0x31,
  0x2e,
  0x37,
  0x0a,
  0x25,
  0xe2,
  0xe3,
  0xcf,
  0xd3,
  0x0a,
  0x0a,
];

Uint8List _normalizeReportPdfBytes(Uint8List bytes) {
  if (!_startsWithBytes(bytes, _pep3InvalidPdfBinaryMarker)) {
    return bytes;
  }
  final int delta =
      _pep3InvalidPdfBinaryMarker.length - _pep3ValidPdfBinaryMarker.length;
  final Uint8List repaired = Uint8List.fromList(<int>[
    ..._pep3ValidPdfBinaryMarker,
    ...bytes.skip(_pep3InvalidPdfBinaryMarker.length),
  ]);
  return _repairPdfCrossReferenceTable(repaired, delta);
}

bool _startsWithBytes(Uint8List bytes, List<int> prefix) {
  if (bytes.length < prefix.length) {
    return false;
  }
  for (int index = 0; index < prefix.length; index++) {
    if (bytes[index] != prefix[index]) {
      return false;
    }
  }
  return true;
}

Uint8List _repairPdfCrossReferenceTable(Uint8List bytes, int delta) {
  final String source = latin1.decode(bytes, allowInvalid: true);
  final int startxrefLabelIndex = source.lastIndexOf('startxref');
  if (startxrefLabelIndex < 0) {
    return bytes;
  }

  final int startxrefValueStart =
      _skipPdfWhitespace(source, startxrefLabelIndex + 'startxref'.length);
  final int startxrefValueEnd = _scanPdfDigits(source, startxrefValueStart);
  if (startxrefValueEnd <= startxrefValueStart) {
    return bytes;
  }

  final int? oldStartxref = int.tryParse(
    source.substring(startxrefValueStart, startxrefValueEnd),
  );
  if (oldStartxref == null || oldStartxref < delta) {
    return bytes;
  }

  final int xrefIndex = oldStartxref - delta;
  if (xrefIndex < 0 || xrefIndex >= source.length) {
    return bytes;
  }
  if (!source.startsWith('xref', xrefIndex)) {
    return bytes;
  }

  final int trailerIndex = source.indexOf('trailer', xrefIndex);
  if (trailerIndex <= xrefIndex) {
    return bytes;
  }

  final String xrefSection = source.substring(xrefIndex, trailerIndex);
  final String repairedXrefSection = xrefSection.replaceAllMapped(
    RegExp(r'^(\d{10}) (\d{5}) n(\s*)$', multiLine: true),
    (Match match) {
      final int offset = int.parse(match.group(1)!);
      final int repairedOffset = offset >= delta ? offset - delta : offset;
      return '${repairedOffset.toString().padLeft(10, '0')} '
          '${match.group(2)} n${match.group(3)}';
    },
  );

  final String repairedSource = source
      .replaceRange(xrefIndex, trailerIndex, repairedXrefSection)
      .replaceRange(
        startxrefValueStart,
        startxrefValueEnd,
        '${oldStartxref - delta}',
      );

  return Uint8List.fromList(latin1.encode(repairedSource));
}

int _skipPdfWhitespace(String source, int index) {
  int cursor = index;
  while (cursor < source.length) {
    final int codeUnit = source.codeUnitAt(cursor);
    if (codeUnit != 0x20 &&
        codeUnit != 0x09 &&
        codeUnit != 0x0a &&
        codeUnit != 0x0d) {
      break;
    }
    cursor += 1;
  }
  return cursor;
}

int _scanPdfDigits(String source, int index) {
  int cursor = index;
  while (cursor < source.length) {
    final int codeUnit = source.codeUnitAt(cursor);
    if (codeUnit < 0x30 || codeUnit > 0x39) {
      break;
    }
    cursor += 1;
  }
  return cursor;
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

Map<String, int> _stringIntMapFrom(Object? value) {
  final Map<String, dynamic> source = _mapFrom(value);
  final Map<String, int> result = <String, int>{};
  source.forEach((String key, dynamic item) {
    final String normalizedKey = key.trim();
    if (normalizedKey.isEmpty) {
      return;
    }
    result[normalizedKey] = _intFrom(item);
  });
  return result;
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
