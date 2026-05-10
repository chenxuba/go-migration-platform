import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'assessment_scale_client.dart';
import 'iep_assessment_record_client.dart';

const String defaultIepPep3PlanDetailPath = String.fromEnvironment(
  'IEP_PEP3_PLAN_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/detail',
);
const String defaultIepErxinPlanDetailPath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/detail',
);
const String defaultIepPep3ExecutionDetailPath = String.fromEnvironment(
  'IEP_PEP3_EXECUTION_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/execution/detail',
);
const String defaultIepErxinExecutionDetailPath = String.fromEnvironment(
  'IEP_ERXIN_EXECUTION_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/execution/detail',
);
const String defaultIepPep3PeriodSyncPath = String.fromEnvironment(
  'IEP_PEP3_PERIOD_SYNC_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/period/sync',
);
const String defaultIepErxinPeriodSyncPath = String.fromEnvironment(
  'IEP_ERXIN_PERIOD_SYNC_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/period/sync',
);
const String defaultIepPep3PlanSavePath = String.fromEnvironment(
  'IEP_PEP3_PLAN_SAVE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/save',
);
const String defaultIepErxinPlanSavePath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_SAVE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/save',
);
const String defaultIepPep3PlanAiStreamPath = String.fromEnvironment(
  'IEP_PEP3_PLAN_AI_STREAM_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/ai/stream',
);
const String defaultIepErxinPlanAiStreamPath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_AI_STREAM_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/ai/stream',
);
const String defaultIepPep3PlanAiTaskPath = String.fromEnvironment(
  'IEP_PEP3_PLAN_AI_TASK_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/ai/tasks',
);
const String defaultIepErxinPlanAiTaskPath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_AI_TASK_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/ai/tasks',
);
const String defaultIepPep3PlanAiTaskStreamPath = String.fromEnvironment(
  'IEP_PEP3_PLAN_AI_TASK_STREAM_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/ai/tasks/stream',
);
const String defaultIepErxinPlanAiTaskStreamPath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_AI_TASK_STREAM_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/ai/tasks/stream',
);
const String defaultIepPep3PlanAiTaskDetailPath = String.fromEnvironment(
  'IEP_PEP3_PLAN_AI_TASK_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/ai/tasks/detail',
);
const String defaultIepErxinPlanAiTaskDetailPath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_AI_TASK_DETAIL_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/ai/tasks/detail',
);
const String defaultIepPep3PlanAiTaskActivePath = String.fromEnvironment(
  'IEP_PEP3_PLAN_AI_TASK_ACTIVE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/ai/tasks/active',
);
const String defaultIepErxinPlanAiTaskActivePath = String.fromEnvironment(
  'IEP_ERXIN_PLAN_AI_TASK_ACTIVE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/ai/tasks/active',
);
const String defaultIepPep3ExecutionAiStreamPath = String.fromEnvironment(
  'IEP_PEP3_EXECUTION_AI_STREAM_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/execution/ai/stream',
);
const String defaultIepErxinExecutionAiStreamPath = String.fromEnvironment(
  'IEP_ERXIN_EXECUTION_AI_STREAM_PATH',
  defaultValue:
      '/api/v1/assessments/erxin/records/iep-plan/execution/ai/stream',
);
const String defaultIepPep3ExecutionSavePath = String.fromEnvironment(
  'IEP_PEP3_EXECUTION_SAVE_PATH',
  defaultValue: '/api/v1/assessments/pep3/records/iep-plan/execution/save',
);
const String defaultIepErxinExecutionSavePath = String.fromEnvironment(
  'IEP_ERXIN_EXECUTION_SAVE_PATH',
  defaultValue: '/api/v1/assessments/erxin/records/iep-plan/execution/save',
);

class IepPlanApiException implements Exception {
  const IepPlanApiException(this.message, {this.unauthorized = false});

  final String message;
  final bool unauthorized;

  @override
  String toString() => message;
}

abstract interface class IepPlanClient {
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  });

  Future<IepExecutionPlansSaved> fetchExecutionPlans(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  });

  Future<IepPlanPeriodSyncResult> syncIepPlanPeriod(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int sourceDurationMonths,
    required DateTime startDate,
  });

  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  });

  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  });

  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  });

  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  });

  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  });

  Future<IepPlanSaved> saveIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String status,
    required IepPlan plan,
  });

  Stream<IepExecutionPlanGenerationEvent<IepMonthlyPlan>>
      generateMonthlyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepPlan sourcePlan,
  });

  Stream<IepExecutionPlanGenerationEvent<IepWeeklyPlan>> generateWeeklyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
  });

  Future<IepExecutionPlansSaved> saveMonthlyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepMonthlyPlan plan,
    bool preserveWeeklyPlans = false,
  });

  Future<IepExecutionPlansSaved> saveWeeklyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepWeeklyPlan plan,
  });
}

class ApiIepPlanClient implements IepPlanClient {
  const ApiIepPlanClient({
    this.educationBaseUrl = defaultAssessmentEducationApiBaseUrl,
    this.pep3PlanDetailPath = defaultIepPep3PlanDetailPath,
    this.erxinPlanDetailPath = defaultIepErxinPlanDetailPath,
    this.pep3ExecutionDetailPath = defaultIepPep3ExecutionDetailPath,
    this.erxinExecutionDetailPath = defaultIepErxinExecutionDetailPath,
    this.pep3PeriodSyncPath = defaultIepPep3PeriodSyncPath,
    this.erxinPeriodSyncPath = defaultIepErxinPeriodSyncPath,
    this.pep3PlanSavePath = defaultIepPep3PlanSavePath,
    this.erxinPlanSavePath = defaultIepErxinPlanSavePath,
    this.pep3PlanAiStreamPath = defaultIepPep3PlanAiStreamPath,
    this.erxinPlanAiStreamPath = defaultIepErxinPlanAiStreamPath,
    this.pep3PlanAiTaskPath = defaultIepPep3PlanAiTaskPath,
    this.erxinPlanAiTaskPath = defaultIepErxinPlanAiTaskPath,
    this.pep3PlanAiTaskStreamPath = defaultIepPep3PlanAiTaskStreamPath,
    this.erxinPlanAiTaskStreamPath = defaultIepErxinPlanAiTaskStreamPath,
    this.pep3PlanAiTaskDetailPath = defaultIepPep3PlanAiTaskDetailPath,
    this.erxinPlanAiTaskDetailPath = defaultIepErxinPlanAiTaskDetailPath,
    this.pep3PlanAiTaskActivePath = defaultIepPep3PlanAiTaskActivePath,
    this.erxinPlanAiTaskActivePath = defaultIepErxinPlanAiTaskActivePath,
    this.pep3ExecutionAiStreamPath = defaultIepPep3ExecutionAiStreamPath,
    this.erxinExecutionAiStreamPath = defaultIepErxinExecutionAiStreamPath,
    this.pep3ExecutionSavePath = defaultIepPep3ExecutionSavePath,
    this.erxinExecutionSavePath = defaultIepErxinExecutionSavePath,
    this.httpClient,
  });

  final String educationBaseUrl;
  final String pep3PlanDetailPath;
  final String erxinPlanDetailPath;
  final String pep3ExecutionDetailPath;
  final String erxinExecutionDetailPath;
  final String pep3PeriodSyncPath;
  final String erxinPeriodSyncPath;
  final String pep3PlanSavePath;
  final String erxinPlanSavePath;
  final String pep3PlanAiStreamPath;
  final String erxinPlanAiStreamPath;
  final String pep3PlanAiTaskPath;
  final String erxinPlanAiTaskPath;
  final String pep3PlanAiTaskStreamPath;
  final String erxinPlanAiTaskStreamPath;
  final String pep3PlanAiTaskDetailPath;
  final String erxinPlanAiTaskDetailPath;
  final String pep3PlanAiTaskActivePath;
  final String erxinPlanAiTaskActivePath;
  final String pep3ExecutionAiStreamPath;
  final String erxinExecutionAiStreamPath;
  final String pep3ExecutionSavePath;
  final String erxinExecutionSavePath;
  final http.Client? httpClient;

  @override
  Future<IepPlanSaved> fetchIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    final Object? data = await _getJson(
      _uri(_isErxinRecord(record) ? erxinPlanDetailPath : pep3PlanDetailPath, {
        'id': '${record.id}',
        'durationMonths': '${_normalizeDuration(durationMonths)}',
      }),
      token,
    );
    if (data is! Map) {
      return IepPlanSaved.empty(_normalizeDuration(durationMonths));
    }
    return IepPlanSaved.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<IepExecutionPlansSaved> fetchExecutionPlans(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    final Object? data = await _getJson(
      _uri(
        _isErxinRecord(record)
            ? erxinExecutionDetailPath
            : pep3ExecutionDetailPath,
        {
          'id': '${record.id}',
          'durationMonths': '${_normalizeDuration(durationMonths)}',
        },
      ),
      token,
    );
    if (data is! Map) {
      return IepExecutionPlansSaved.empty(_normalizeDuration(durationMonths));
    }
    return IepExecutionPlansSaved.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<IepPlanPeriodSyncResult> syncIepPlanPeriod(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int sourceDurationMonths,
    required DateTime startDate,
  }) async {
    final Object? data = await _postJson(
      _uri(_isErxinRecord(record) ? erxinPeriodSyncPath : pep3PeriodSyncPath),
      token,
      <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'sourceDurationMonths': _normalizeDuration(sourceDurationMonths),
        'startDate': _formatDateDash(startDate),
      },
    );
    if (data is! Map) {
      return IepPlanPeriodSyncResult.empty(_normalizeDuration(durationMonths));
    }
    return IepPlanPeriodSyncResult.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Stream<IepPlanGenerationEvent> generateIepPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async* {
    final IepPlanGenerationTask task = await createIepPlanGenerationTask(
      token,
      record: record,
      durationMonths: durationMonths,
    );
    yield* watchIepPlanGenerationTask(
      token,
      record: record,
      taskId: task.taskId,
    );
  }

  @override
  Future<IepPlanGenerationTask> createIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
  }) async {
    final Object? data = await _postJson(
      _uri(_isErxinRecord(record) ? erxinPlanAiTaskPath : pep3PlanAiTaskPath),
      token,
      <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
      },
    );
    if (data is! Map) {
      throw const IepPlanApiException('AI生成任务创建失败：接口返回异常');
    }
    return IepPlanGenerationTask.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<IepPlanGenerationTask> fetchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async {
    final Object? data = await _getJson(
      _uri(
        _isErxinRecord(record)
            ? erxinPlanAiTaskDetailPath
            : pep3PlanAiTaskDetailPath,
        <String, String>{'taskId': taskId},
      ),
      token,
    );
    if (data is! Map) {
      throw const IepPlanApiException('AI生成任务查询失败：接口返回异常');
    }
    return IepPlanGenerationTask.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<IepPlanGenerationTask?> fetchActiveIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
  }) async {
    final Object? data = await _getJson(
      _uri(
        _isErxinRecord(record)
            ? erxinPlanAiTaskActivePath
            : pep3PlanAiTaskActivePath,
        <String, String>{'id': '${record.id}'},
      ),
      token,
    );
    if (data is! Map) {
      return null;
    }
    final IepPlanGenerationTask task = IepPlanGenerationTask.fromJson(
      Map<String, dynamic>.from(data),
    );
    return task.exists ? task : null;
  }

  @override
  Stream<IepPlanGenerationEvent> watchIepPlanGenerationTask(
    String token, {
    required IepAssessmentRecordSummary record,
    required String taskId,
  }) async* {
    final http.Client client = httpClient ?? http.Client();
    final bool shouldCloseClient = httpClient == null;
    bool hasDone = false;
    String lastStreamText = '';
    try {
      final http.Request request = http.Request(
        'GET',
        _uri(
            _isErxinRecord(record)
                ? erxinPlanAiTaskStreamPath
                : pep3PlanAiTaskStreamPath,
            <String, String>{
              'taskId': taskId,
            }),
      )..headers.addAll(_headers(token, accept: 'text/event-stream'));

      final http.StreamedResponse response =
          await client.send(request).timeout(const Duration(seconds: 12));
      if (response.statusCode == 401 || response.statusCode == 403) {
        final String body = await response.stream.bytesToString();
        throw IepPlanApiException(
          _messageFromPayload(_tryDecodeJson(body)) ?? '登录已失效，请重新登录',
          unauthorized: true,
        );
      }
      if (response.statusCode < 200 || response.statusCode >= 300) {
        final String body = await response.stream.bytesToString();
        throw IepPlanApiException(
          _messageFromPayload(_tryDecodeJson(body)) ??
              (body.trim().isEmpty ? 'AI生成失败' : body.trim()),
        );
      }

      String buffer = '';
      await for (final String chunk
          in response.stream.transform(utf8.decoder)) {
        buffer += chunk;
        final List<String> frames = buffer.split(RegExp(r'\r?\n\r?\n'));
        buffer = frames.removeLast();
        for (final String frame in frames) {
          final List<IepPlanGenerationEvent> events =
              _parseTaskSseFrame(frame, lastStreamText);
          if (events.isEmpty) {
            continue;
          }
          final IepPlanGenerationTask? task = _taskFromSseFrame(frame);
          if (task != null) {
            lastStreamText = task.streamText;
          }
          for (final IepPlanGenerationEvent event in events) {
            if (event.type == IepPlanGenerationEventType.done) {
              hasDone = true;
            }
            yield event;
          }
        }
      }
      if (buffer.trim().isNotEmpty) {
        final List<IepPlanGenerationEvent> events =
            _parseTaskSseFrame(buffer, lastStreamText);
        final IepPlanGenerationTask? task = _taskFromSseFrame(buffer);
        if (task != null) {
          lastStreamText = task.streamText;
        }
        for (final IepPlanGenerationEvent event in events) {
          if (event.type == IepPlanGenerationEventType.done) {
            hasDone = true;
          }
          yield event;
        }
      }
      if (!hasDone) {
        final IepPlanGenerationTask latest = await fetchIepPlanGenerationTask(
          token,
          record: record,
          taskId: taskId,
        );
        final IepPlan? restoredPlan = latest.savedPlan?.plan ?? latest.plan;
        if (latest.isDone && restoredPlan != null) {
          yield IepPlanGenerationEvent.done(
            restoredPlan,
            savedPlan: latest.savedPlan,
          );
          return;
        }
        if (latest.isFailed) {
          throw IepPlanApiException(
            latest.error.isEmpty ? 'AI生成失败' : latest.error,
          );
        }
        throw const IepPlanApiException('AI生成连接已断开，请稍后返回页面查看生成结果');
      }
    } on TimeoutException {
      throw const IepPlanApiException('AI生成接口响应超时，请检查网络');
    } on IepPlanApiException {
      rethrow;
    } on Object catch (error) {
      throw IepPlanApiException('无法连接AI生成接口：$error');
    } finally {
      if (shouldCloseClient) {
        client.close();
      }
    }
  }

  @override
  Future<IepPlanSaved> saveIepPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required String status,
    required IepPlan plan,
  }) async {
    final Object? data = await _postJson(
      _uri(_isErxinRecord(record) ? erxinPlanSavePath : pep3PlanSavePath),
      token,
      <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'status': status,
        'plan': plan.toJson(),
      },
    );
    if (data is! Map) {
      throw const IepPlanApiException('IEP计划保存失败：接口返回异常');
    }
    final IepPlanSaved saved = IepPlanSaved.fromJson(
      Map<String, dynamic>.from(data),
    );
    if (!saved.hasContent) {
      throw const IepPlanApiException('IEP计划保存失败：接口未返回已保存内容');
    }
    return saved;
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepMonthlyPlan>>
      generateMonthlyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepPlan sourcePlan,
  }) {
    return _generateExecutionPlanStream<IepMonthlyPlan>(
      token,
      record: record,
      payload: <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'planType': 'monthly',
        'targetMonthIndex': targetMonthIndex,
        'sourcePlan': sourcePlan.toJson(),
      },
      parser: (Map<String, dynamic> json) => IepMonthlyPlan.fromJson(json),
    );
  }

  @override
  Stream<IepExecutionPlanGenerationEvent<IepWeeklyPlan>> generateWeeklyPlanStream(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepPlan sourcePlan,
    IepMonthlyPlan? monthlyPlan,
  }) {
    return _generateExecutionPlanStream<IepWeeklyPlan>(
      token,
      record: record,
      payload: <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'planType': 'weekly',
        'targetMonthIndex': targetMonthIndex,
        'targetWeekIndex': targetWeekIndex,
        'sourcePlan': sourcePlan.toJson(),
        if (monthlyPlan != null) 'monthlyPlan': monthlyPlan.toJson(),
      },
      parser: (Map<String, dynamic> json) => IepWeeklyPlan.fromJson(json),
    );
  }

  @override
  Future<IepExecutionPlansSaved> saveMonthlyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required IepMonthlyPlan plan,
    bool preserveWeeklyPlans = false,
  }) async {
    final Object? data = await _postJson(
      _uri(
        _isErxinRecord(record) ? erxinExecutionSavePath : pep3ExecutionSavePath,
      ),
      token,
      <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'planType': 'monthly',
        'targetMonthIndex': targetMonthIndex,
        'monthlyPlan': plan.toJson(),
        'preserveWeeklyPlans': preserveWeeklyPlans,
      },
    );
    if (data is! Map) {
      throw const IepPlanApiException('月计划保存失败：接口返回异常');
    }
    return IepExecutionPlansSaved.fromJson(Map<String, dynamic>.from(data));
  }

  @override
  Future<IepExecutionPlansSaved> saveWeeklyPlan(
    String token, {
    required IepAssessmentRecordSummary record,
    required int durationMonths,
    required int targetMonthIndex,
    required int targetWeekIndex,
    required IepWeeklyPlan plan,
  }) async {
    final Object? data = await _postJson(
      _uri(
        _isErxinRecord(record) ? erxinExecutionSavePath : pep3ExecutionSavePath,
      ),
      token,
      <String, dynamic>{
        'id': record.id,
        'durationMonths': _normalizeDuration(durationMonths),
        'planType': 'weekly',
        'targetMonthIndex': targetMonthIndex,
        'targetWeekIndex': targetWeekIndex,
        'weeklyPlan': plan.toJson(),
      },
    );
    if (data is! Map) {
      throw const IepPlanApiException('周计划保存失败：接口返回异常');
    }
    return IepExecutionPlansSaved.fromJson(Map<String, dynamic>.from(data));
  }

  Stream<IepExecutionPlanGenerationEvent<T>> _generateExecutionPlanStream<T>(
    String token, {
    required IepAssessmentRecordSummary record,
    required Map<String, dynamic> payload,
    required T Function(Map<String, dynamic>) parser,
  }) async* {
    final http.Client client = httpClient ?? http.Client();
    final bool shouldCloseClient = httpClient == null;
    bool hasDone = false;
    try {
      final http.Request request = http.Request(
        'POST',
        _uri(
          _isErxinRecord(record)
              ? erxinExecutionAiStreamPath
              : pep3ExecutionAiStreamPath,
        ),
      )
        ..headers.addAll(_headers(token, accept: 'text/event-stream'))
        ..body = jsonEncode(payload);

      final http.StreamedResponse response =
          await client.send(request).timeout(const Duration(seconds: 12));
      if (response.statusCode == 401 || response.statusCode == 403) {
        final String body = await response.stream.bytesToString();
        throw IepPlanApiException(
          _messageFromPayload(_tryDecodeJson(body)) ?? '登录已失效，请重新登录',
          unauthorized: true,
        );
      }
      if (response.statusCode < 200 || response.statusCode >= 300) {
        final String body = await response.stream.bytesToString();
        throw IepPlanApiException(
          _messageFromPayload(_tryDecodeJson(body)) ??
              (body.trim().isEmpty ? 'AI生成失败' : body.trim()),
        );
      }

      String buffer = '';
      await for (final String chunk
          in response.stream.transform(utf8.decoder)) {
        buffer += chunk;
        final List<String> frames = buffer.split(RegExp(r'\r?\n\r?\n'));
        buffer = frames.removeLast();
        for (final String frame in frames) {
          final IepExecutionPlanGenerationEvent<T>? event =
              _parseExecutionSseFrame<T>(frame, parser);
          if (event == null) {
            continue;
          }
          if (event.type == IepExecutionPlanGenerationEventType.done) {
            hasDone = true;
          }
          yield event;
        }
      }
      if (buffer.trim().isNotEmpty) {
        final IepExecutionPlanGenerationEvent<T>? event =
            _parseExecutionSseFrame<T>(buffer, parser);
        if (event != null) {
          if (event.type == IepExecutionPlanGenerationEventType.done) {
            hasDone = true;
          }
          yield event;
        }
      }
      if (!hasDone) {
        throw const IepPlanApiException('AI生成连接已断开，请稍后重试');
      }
    } on TimeoutException {
      throw const IepPlanApiException('AI生成接口响应超时，请检查网络');
    } on IepPlanApiException {
      rethrow;
    } on Object catch (error) {
      throw IepPlanApiException('无法连接AI生成接口：$error');
    } finally {
      if (shouldCloseClient) {
        client.close();
      }
    }
  }

  Future<Object?> _getJson(Uri uri, String token) async {
    final http.Client client = httpClient ?? http.Client();
    final bool shouldCloseClient = httpClient == null;
    final http.Response response;
    try {
      response = await client
          .get(uri, headers: _headers(token))
          .timeout(const Duration(seconds: 12));
    } on TimeoutException {
      throw const IepPlanApiException('IEP计划接口响应超时，请检查网络');
    } on Object catch (error) {
      throw IepPlanApiException('无法连接IEP计划接口：$error');
    } finally {
      if (shouldCloseClient) {
        client.close();
      }
    }
    return _handleResponse(response);
  }

  Future<Object?> _postJson(
    Uri uri,
    String token,
    Map<String, dynamic> payload,
  ) async {
    final http.Client client = httpClient ?? http.Client();
    final bool shouldCloseClient = httpClient == null;
    final http.Response response;
    try {
      response = await client
          .post(uri, headers: _headers(token), body: jsonEncode(payload))
          .timeout(const Duration(seconds: 60));
    } on TimeoutException {
      throw const IepPlanApiException('IEP计划接口响应超时，请检查网络');
    } on Object catch (error) {
      throw IepPlanApiException('无法连接IEP计划接口：$error');
    } finally {
      if (shouldCloseClient) {
        client.close();
      }
    }
    return _handleResponse(response);
  }

  Uri _uri(String path, [Map<String, String>? query]) {
    final String trimmedBase =
        educationBaseUrl.trim().replaceFirst(RegExp(r'/+$'), '');
    final String normalizedPath = path.startsWith('/') ? path : '/$path';
    return Uri.parse('$trimmedBase$normalizedPath').replace(
      queryParameters: query,
    );
  }

  Map<String, String> _headers(
    String token, {
    String accept = 'application/json',
  }) {
    return <String, String>{
      'Accept': accept,
      'Content-Type': 'application/json; charset=utf-8',
      'Accept-Language': 'zh-CN',
      if (token.trim().isNotEmpty) 'Authorization': 'Bearer ${token.trim()}',
      if (token.trim().isNotEmpty) 'X-Access-Token': token.trim(),
    };
  }

  Object? _handleResponse(http.Response response) {
    final String body = utf8.decode(response.bodyBytes).trim();
    final Object? decoded = body.isEmpty ? null : _tryDecodeJson(body);
    if (response.statusCode == 401 || response.statusCode == 403) {
      throw IepPlanApiException(
        _messageFromPayload(decoded) ?? '登录已失效，请重新登录',
        unauthorized: true,
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw IepPlanApiException(_messageFromPayload(decoded) ?? 'IEP计划加载失败');
    }
    if (decoded is Map) {
      final Map<String, dynamic> envelope = Map<String, dynamic>.from(decoded);
      if (envelope['success'] == false) {
        throw IepPlanApiException(
          _messageFromPayload(envelope) ?? 'IEP计划加载失败',
        );
      }
      if (envelope.containsKey('data')) {
        return envelope['data'];
      }
    }
    return decoded;
  }
}

enum IepPlanGenerationEventType { status, delta, done, error }

class IepPlanGenerationEvent {
  const IepPlanGenerationEvent._({
    required this.type,
    this.message = '',
    this.text = '',
    this.costAmountCny = 0,
    this.plan,
    this.savedPlan,
  });

  factory IepPlanGenerationEvent.status(String message) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.status,
      message: message,
    );
  }

  factory IepPlanGenerationEvent.statusWithCost(
    String message,
    double costAmountCny,
  ) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.status,
      message: message,
      costAmountCny: costAmountCny,
    );
  }

  factory IepPlanGenerationEvent.delta(String text) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.delta,
      text: text,
    );
  }

  factory IepPlanGenerationEvent.deltaWithCost(
    String text,
    double costAmountCny,
  ) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.delta,
      text: text,
      costAmountCny: costAmountCny,
    );
  }

  factory IepPlanGenerationEvent.done(
    IepPlan plan, {
    IepPlanSaved? savedPlan,
    double costAmountCny = 0,
  }) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.done,
      costAmountCny: costAmountCny,
      plan: plan,
      savedPlan: savedPlan,
    );
  }

  factory IepPlanGenerationEvent.error(String message) {
    return IepPlanGenerationEvent._(
      type: IepPlanGenerationEventType.error,
      message: message,
    );
  }

  final IepPlanGenerationEventType type;
  final String message;
  final String text;
  final double costAmountCny;
  final IepPlan? plan;
  final IepPlanSaved? savedPlan;
}

class IepPlanGenerationTask {
  const IepPlanGenerationTask({
    this.exists = true,
    required this.taskId,
    required this.status,
    required this.durationMonths,
    this.message = '',
    this.streamText = '',
    this.costAmountCny = 0,
    this.plan,
    this.savedPlan,
    this.error = '',
    this.updatedTime = '',
  });

  factory IepPlanGenerationTask.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> planJson = _mapFrom(json['plan']);
    final Map<String, dynamic> savedPlanJson = _mapFrom(json['savedPlan']);
    return IepPlanGenerationTask(
      exists: json.containsKey('exists') ? _boolFrom(json['exists']) : true,
      taskId: _stringFrom(json['taskId']),
      status: _stringFrom(json['status']),
      message: _stringFrom(json['message']),
      streamText: _stringFrom(json['streamText']),
      costAmountCny: _doubleFrom(json['costAmountCny']),
      durationMonths: _intFrom(json['durationMonths']),
      plan: planJson.isEmpty ? null : IepPlan.fromJson(planJson),
      savedPlan:
          savedPlanJson.isEmpty ? null : IepPlanSaved.fromJson(savedPlanJson),
      error: _stringFrom(json['error']),
      updatedTime: _stringFrom(json['updatedTime']),
    );
  }

  final bool exists;
  final String taskId;
  final String status;
  final String message;
  final String streamText;
  final double costAmountCny;
  final int durationMonths;
  final IepPlan? plan;
  final IepPlanSaved? savedPlan;
  final String error;
  final String updatedTime;

  bool get isDone => status == 'done';
  bool get isFailed => status == 'failed';
}

enum IepExecutionPlanGenerationEventType { status, delta, done, error }

class IepExecutionPlanGenerationEvent<T> {
  const IepExecutionPlanGenerationEvent._({
    required this.type,
    this.message = '',
    this.text = '',
    this.costAmountCny = 0,
    this.data,
  });

  factory IepExecutionPlanGenerationEvent.status(String message) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.status,
      message: message,
    );
  }

  factory IepExecutionPlanGenerationEvent.statusWithCost(
    String message,
    double costAmountCny,
  ) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.status,
      message: message,
      costAmountCny: costAmountCny,
    );
  }

  factory IepExecutionPlanGenerationEvent.delta(String text) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.delta,
      text: text,
    );
  }

  factory IepExecutionPlanGenerationEvent.deltaWithCost(
    String text,
    double costAmountCny,
  ) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.delta,
      text: text,
      costAmountCny: costAmountCny,
    );
  }

  factory IepExecutionPlanGenerationEvent.done(T data, {double costAmountCny = 0}) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.done,
      costAmountCny: costAmountCny,
      data: data,
    );
  }

  factory IepExecutionPlanGenerationEvent.error(String message) {
    return IepExecutionPlanGenerationEvent<T>._(
      type: IepExecutionPlanGenerationEventType.error,
      message: message,
    );
  }

  final IepExecutionPlanGenerationEventType type;
  final String message;
  final String text;
  final double costAmountCny;
  final T? data;
}

class IepPlanPeriodSyncResult {
  const IepPlanPeriodSyncResult({
    required this.iepPlan,
    required this.executionPlans,
  });

  factory IepPlanPeriodSyncResult.fromJson(Map<String, dynamic> json) {
    return IepPlanPeriodSyncResult(
      iepPlan: IepPlanSaved.fromJson(_mapFrom(json['iepPlan'])),
      executionPlans:
          IepExecutionPlansSaved.fromJson(_mapFrom(json['executionPlans'])),
    );
  }

  factory IepPlanPeriodSyncResult.empty(int durationMonths) {
    return IepPlanPeriodSyncResult(
      iepPlan: IepPlanSaved.empty(durationMonths),
      executionPlans: IepExecutionPlansSaved.empty(durationMonths),
    );
  }

  final IepPlanSaved iepPlan;
  final IepExecutionPlansSaved executionPlans;
}

class IepPlanSaved {
  const IepPlanSaved({
    required this.exists,
    required this.durationMonths,
    this.status = '',
    this.plan,
    this.updatedTime = '',
  });

  factory IepPlanSaved.fromJson(Map<String, dynamic> json) {
    final Map<String, dynamic> planJson = _mapFrom(json['plan']);
    return IepPlanSaved(
      exists: _boolFrom(json['exists']),
      status: _stringFrom(json['status']),
      durationMonths: _intFrom(json['durationMonths']),
      plan: planJson.isEmpty ? null : IepPlan.fromJson(planJson),
      updatedTime: _stringFrom(json['updatedTime']),
    );
  }

  factory IepPlanSaved.empty(int durationMonths) {
    return IepPlanSaved(
      exists: false,
      durationMonths: _normalizeDuration(durationMonths),
    );
  }

  final bool exists;
  final String status;
  final int durationMonths;
  final IepPlan? plan;
  final String updatedTime;

  bool get hasContent {
    final IepPlan? currentPlan = plan;
    return exists &&
        currentPlan != null &&
        currentPlan.rows.any((IepPlanRow row) => row.shortGoal.isNotEmpty);
  }
}

class IepExecutionPlansSaved {
  const IepExecutionPlansSaved({
    required this.exists,
    required this.durationMonths,
    required this.monthlyPlans,
    required this.weeklyPlans,
  });

  factory IepExecutionPlansSaved.fromJson(Map<String, dynamic> json) {
    return IepExecutionPlansSaved(
      exists: _boolFrom(json['exists']),
      durationMonths: _intFrom(json['durationMonths']),
      monthlyPlans: _listFrom(json['monthlyPlans'])
          .map(IepMonthlyPlanSaved.fromJson)
          .toList(),
      weeklyPlans: _listFrom(json['weeklyPlans'])
          .map(IepWeeklyPlanSaved.fromJson)
          .toList(),
    );
  }

  factory IepExecutionPlansSaved.empty(int durationMonths) {
    return IepExecutionPlansSaved(
      exists: false,
      durationMonths: _normalizeDuration(durationMonths),
      monthlyPlans: const <IepMonthlyPlanSaved>[],
      weeklyPlans: const <IepWeeklyPlanSaved>[],
    );
  }

  final bool exists;
  final int durationMonths;
  final List<IepMonthlyPlanSaved> monthlyPlans;
  final List<IepWeeklyPlanSaved> weeklyPlans;

  IepMonthlyPlan? monthPlan(int monthIndex) {
    for (final IepMonthlyPlanSaved item in monthlyPlans) {
      if (item.targetMonthIndex == monthIndex) {
        return item.plan;
      }
    }
    return null;
  }

  IepWeeklyPlan? weekPlan(int monthIndex, int weekIndex) {
    for (final IepWeeklyPlanSaved item in weeklyPlans) {
      if (item.targetMonthIndex == monthIndex &&
          item.targetWeekIndex == weekIndex) {
        return item.plan;
      }
    }
    return null;
  }
}

class IepMonthlyPlanSaved {
  const IepMonthlyPlanSaved({
    required this.targetMonthIndex,
    required this.plan,
    this.updatedTime = '',
  });

  factory IepMonthlyPlanSaved.fromJson(Map<String, dynamic> json) {
    return IepMonthlyPlanSaved(
      targetMonthIndex: _intFrom(json['targetMonthIndex']),
      plan: IepMonthlyPlan.fromJson(_mapFrom(json['plan'])),
      updatedTime: _stringFrom(json['updatedTime']),
    );
  }

  final int targetMonthIndex;
  final IepMonthlyPlan plan;
  final String updatedTime;
}

class IepWeeklyPlanSaved {
  const IepWeeklyPlanSaved({
    required this.targetMonthIndex,
    required this.targetWeekIndex,
    required this.plan,
    this.updatedTime = '',
  });

  factory IepWeeklyPlanSaved.fromJson(Map<String, dynamic> json) {
    return IepWeeklyPlanSaved(
      targetMonthIndex: _intFrom(json['targetMonthIndex']),
      targetWeekIndex: _intFrom(json['targetWeekIndex']),
      plan: IepWeeklyPlan.fromJson(_mapFrom(json['plan'])),
      updatedTime: _stringFrom(json['updatedTime']),
    );
  }

  final int targetMonthIndex;
  final int targetWeekIndex;
  final IepWeeklyPlan plan;
  final String updatedTime;
}

class IepPlan {
  const IepPlan({
    required this.title,
    required this.student,
    required this.meta,
    required this.rows,
  });

  factory IepPlan.fromJson(Map<String, dynamic> json) {
    return IepPlan(
      title: _stringFrom(json['title']),
      student: IepPlanStudent.fromJson(_mapFrom(json['student'])),
      meta: IepPlanMeta.fromJson(_mapFrom(json['meta'])),
      rows: _listFrom(json['rows']).map(IepPlanRow.fromJson).toList(),
    );
  }

  final String title;
  final IepPlanStudent student;
  final IepPlanMeta meta;
  final List<IepPlanRow> rows;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      'student': student.toJson(),
      'meta': meta.toJson(),
      'rows': rows.map((IepPlanRow row) => row.toJson()).toList(),
    };
  }
}

class IepPlanStudent {
  const IepPlanStudent({
    required this.name,
    required this.gender,
    required this.birthDate,
  });

  factory IepPlanStudent.fromJson(Map<String, dynamic> json) {
    return IepPlanStudent(
      name: _stringFrom(json['name']),
      gender: _stringFrom(json['gender']),
      birthDate: _dateStringFrom(json['birthDate']),
    );
  }

  final String name;
  final String gender;
  final String birthDate;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'name': name,
      'gender': gender,
      'birthDate': birthDate,
    };
  }
}

class IepPlanMeta {
  const IepPlanMeta({
    required this.planDate,
    required this.participant,
    required this.implementer,
    required this.startDate,
    required this.endDate,
  });

  factory IepPlanMeta.fromJson(Map<String, dynamic> json) {
    return IepPlanMeta(
      planDate: _dateStringFrom(json['planDate']),
      participant: _stringFrom(json['participant']),
      implementer: _stringFrom(json['implementer']),
      startDate: _dateStringFrom(json['startDate']),
      endDate: _dateStringFrom(json['endDate']),
    );
  }

  final String planDate;
  final String participant;
  final String implementer;
  final String startDate;
  final String endDate;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'planDate': planDate,
      'participant': participant,
      'implementer': implementer,
      'startDate': startDate,
      'endDate': endDate,
    };
  }
}

class IepPlanRow {
  const IepPlanRow({
    required this.domain,
    required this.longGoal,
    required this.shortGoal,
    required this.courseForm,
    required this.startEndDate,
  });

  factory IepPlanRow.fromJson(Map<String, dynamic> json) {
    return IepPlanRow(
      domain: _stringFrom(json['domain']),
      longGoal: _stringFrom(json['longGoal']),
      shortGoal: _stringFrom(json['shortGoal']),
      courseForm: _stringFrom(json['courseForm']),
      startEndDate: _stringFrom(json['startEndDate']),
    );
  }

  final String domain;
  final String longGoal;
  final String shortGoal;
  final String courseForm;
  final String startEndDate;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'domain': domain,
      'longGoal': longGoal,
      'shortGoal': shortGoal,
      'courseForm': courseForm,
      'startEndDate': startEndDate,
    };
  }
}

class IepMonthlyPlan {
  const IepMonthlyPlan({
    required this.title,
    required this.student,
    required this.meta,
    required this.rows,
  });

  factory IepMonthlyPlan.fromJson(Map<String, dynamic> json) {
    return IepMonthlyPlan(
      title: _stringFrom(json['title']),
      student: IepPlanStudent.fromJson(_mapFrom(json['student'])),
      meta: IepMonthlyPlanMeta.fromJson(_mapFrom(json['meta'])),
      rows: _listFrom(json['rows']).map(IepMonthlyPlanRow.fromJson).toList(),
    );
  }

  final String title;
  final IepPlanStudent student;
  final IepMonthlyPlanMeta meta;
  final List<IepMonthlyPlanRow> rows;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      'student': student.toJson(),
      'meta': meta.toJson(),
      'rows': rows.map((IepMonthlyPlanRow row) => row.toJson()).toList(),
    };
  }
}

class IepMonthlyPlanMeta extends IepPlanMeta {
  const IepMonthlyPlanMeta({
    required super.planDate,
    required super.participant,
    required super.implementer,
    required super.startDate,
    required super.endDate,
    required this.monthLabel,
    required this.sourceTitle,
  });

  factory IepMonthlyPlanMeta.fromJson(Map<String, dynamic> json) {
    return IepMonthlyPlanMeta(
      planDate: _dateStringFrom(json['planDate']),
      participant: _stringFrom(json['participant']),
      implementer: _stringFrom(json['implementer']),
      startDate: _dateStringFrom(json['startDate']),
      endDate: _dateStringFrom(json['endDate']),
      monthLabel: _stringFrom(json['monthLabel']),
      sourceTitle: _stringFrom(json['sourceTitle']),
    );
  }

  final String monthLabel;
  final String sourceTitle;

  @override
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      ...super.toJson(),
      'monthLabel': monthLabel,
      'sourceTitle': sourceTitle,
    };
  }
}

class IepMonthlyPlanRow {
  const IepMonthlyPlanRow({
    required this.domain,
    required this.longGoal,
    required this.shortGoal,
    required this.trainingItems,
    required this.courseForm,
  });

  factory IepMonthlyPlanRow.fromJson(Map<String, dynamic> json) {
    return IepMonthlyPlanRow(
      domain: _stringFrom(json['domain']),
      longGoal: _stringFrom(json['longGoal']),
      shortGoal: _stringFrom(json['shortGoal']),
      trainingItems: _listFrom(json['trainingItems'])
          .map(IepMonthlyTrainingItem.fromJson)
          .toList(),
      courseForm: _stringFrom(json['courseForm']),
    );
  }

  final String domain;
  final String longGoal;
  final String shortGoal;
  final List<IepMonthlyTrainingItem> trainingItems;
  final String courseForm;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'domain': domain,
      'longGoal': longGoal,
      'shortGoal': shortGoal,
      'trainingItems': trainingItems
          .map((IepMonthlyTrainingItem item) => item.toJson())
          .toList(),
      'courseForm': courseForm,
    };
  }
}

class IepMonthlyTrainingItem {
  const IepMonthlyTrainingItem({
    required this.content,
    required this.startEndDate,
  });

  factory IepMonthlyTrainingItem.fromJson(Map<String, dynamic> json) {
    return IepMonthlyTrainingItem(
      content: _stringFrom(json['content']),
      startEndDate: _stringFrom(json['startEndDate']),
    );
  }

  final String content;
  final String startEndDate;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'content': content,
      'startEndDate': startEndDate,
    };
  }
}

class IepWeeklyPlan {
  const IepWeeklyPlan({
    required this.title,
    required this.student,
    required this.teacherName,
    required this.courseName,
    required this.trainingDate,
    required this.preparation,
    required this.weekDates,
    required this.rows,
  });

  factory IepWeeklyPlan.fromJson(Map<String, dynamic> json) {
    return IepWeeklyPlan(
      title: _stringFrom(json['title']),
      student: IepPlanStudent.fromJson(_mapFrom(json['student'])),
      teacherName: _stringFrom(json['teacherName']),
      courseName: _stringFrom(json['courseName']),
      trainingDate: _stringFrom(json['trainingDate']),
      preparation: _stringFrom(json['preparation']),
      weekDates: _stringListFrom(json['weekDates']),
      rows: _listFrom(json['rows']).map(IepWeeklyPlanRow.fromJson).toList(),
    );
  }

  final String title;
  final IepPlanStudent student;
  final String teacherName;
  final String courseName;
  final String trainingDate;
  final String preparation;
  final List<String> weekDates;
  final List<IepWeeklyPlanRow> rows;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'title': title,
      'student': student.toJson(),
      'teacherName': teacherName,
      'courseName': courseName,
      'trainingDate': trainingDate,
      'preparation': preparation,
      'weekDates': weekDates,
      'rows': rows.map((IepWeeklyPlanRow row) => row.toJson()).toList(),
    };
  }
}

class IepWeeklyPlanRow {
  const IepWeeklyPlanRow({
    required this.project,
    required this.content,
    required this.completion,
  });

  factory IepWeeklyPlanRow.fromJson(Map<String, dynamic> json) {
    return IepWeeklyPlanRow(
      project: _stringFrom(json['project']),
      content: _stringFrom(json['content']),
      completion: _stringListFrom(json['completion']),
    );
  }

  final String project;
  final String content;
  final List<String> completion;

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'project': project,
      'content': content,
      'completion': completion,
    };
  }
}

bool _isErxinRecord(IepAssessmentRecordSummary record) {
  final String source = record.source.trim().toUpperCase();
  final String code = record.assessmentCode.trim().toUpperCase();
  return source == 'ERXIN' || code.startsWith('ERXIN');
}

int _normalizeDuration(int durationMonths) => durationMonths == 6 ? 6 : 3;

String _formatDateDash(DateTime value) {
  return '${value.year.toString().padLeft(4, '0')}-'
      '${value.month.toString().padLeft(2, '0')}-'
      '${value.day.toString().padLeft(2, '0')}';
}

Map<String, dynamic> _mapFrom(Object? value) {
  if (value is Map) {
    return Map<String, dynamic>.from(value);
  }
  return <String, dynamic>{};
}

List<Map<String, dynamic>> _listFrom(Object? value) {
  if (value is! List) {
    return <Map<String, dynamic>>[];
  }
  return value
      .whereType<Map>()
      .map((Map<dynamic, dynamic> item) => Map<String, dynamic>.from(item))
      .toList();
}

List<String> _stringListFrom(Object? value) {
  if (value is! List) {
    return <String>[];
  }
  return value.map((Object? item) => _stringFrom(item)).toList();
}

bool _boolFrom(Object? value) {
  if (value is bool) {
    return value;
  }
  return '${value ?? ''}'.trim().toLowerCase() == 'true';
}

int _intFrom(Object? value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  return int.tryParse('${value ?? ''}') ?? 0;
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
  return double.tryParse('${value ?? ''}') ?? 0;
}

String _stringFrom(Object? value) => '${value ?? ''}'.trim();

String _dateStringFrom(Object? value) {
  final String raw = _stringFrom(value);
  if (raw.isEmpty) {
    return '';
  }
  final DateTime? parsed = DateTime.tryParse(raw);
  if (parsed == null) {
    return raw.contains('T') ? raw.split('T').first : raw;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')}';
}

String? _messageFromPayload(Object? payload) {
  if (payload is Map) {
    for (final String key in <String>['message', 'msg', 'error']) {
      final Object? value = payload[key];
      if (value != null && '$value'.trim().isNotEmpty) {
        return '$value';
      }
    }
  }
  return null;
}

IepPlanGenerationEvent? _parseSseFrame(String frame) {
  final List<String> dataLines = frame
      .split(RegExp(r'\r?\n'))
      .where((String line) => line.startsWith('data:'))
      .map((String line) => line.substring(5).trim())
      .toList();
  if (dataLines.isEmpty) {
    return null;
  }
  final Object? decoded = _tryDecodeJson(dataLines.join('\n'));
  if (decoded is! Map) {
    return null;
  }
  final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
  final String type = _stringFrom(payload['type']);
  final Map<String, dynamic> taskJson = _mapFrom(payload['data']);
  if (taskJson.isNotEmpty) {
    final IepPlanGenerationTask task = IepPlanGenerationTask.fromJson(taskJson);
    if (task.isDone) {
      final IepPlan? plan = task.savedPlan?.plan ?? task.plan;
      if (plan == null) {
        return IepPlanGenerationEvent.error('AI生成未返回计划数据');
      }
      return IepPlanGenerationEvent.done(
        plan,
        savedPlan: task.savedPlan,
        costAmountCny: task.costAmountCny,
      );
    }
    if (task.isFailed) {
      return IepPlanGenerationEvent.error(
        task.error.isEmpty ? 'AI生成失败' : task.error,
      );
    }
    return IepPlanGenerationEvent.statusWithCost(
      task.message,
      task.costAmountCny,
    );
  }
  return switch (type) {
    'status' => IepPlanGenerationEvent.status(
        _stringFrom(payload['message']),
      ),
    'delta' => IepPlanGenerationEvent.delta(
        _stringFrom(payload['text']),
      ),
    'done' => IepPlanGenerationEvent.done(
        IepPlan.fromJson(_mapFrom(payload['data'])),
        costAmountCny: _doubleFrom(payload['costAmountCny']),
      ),
    'error' => IepPlanGenerationEvent.error(
        _stringFrom(payload['message']).isEmpty
            ? 'AI生成失败'
            : _stringFrom(payload['message']),
      ),
    _ => null,
  };
}

List<IepPlanGenerationEvent> _parseTaskSseFrame(
  String frame,
  String lastStreamText,
) {
  final IepPlanGenerationTask? task = _taskFromSseFrame(frame);
  if (task == null) {
    final IepPlanGenerationEvent? event = _parseSseFrame(frame);
    return event == null
        ? const <IepPlanGenerationEvent>[]
        : <IepPlanGenerationEvent>[event];
  }
  if (task.isDone) {
    final IepPlan? plan = task.savedPlan?.plan ?? task.plan;
    if (plan == null) {
      return <IepPlanGenerationEvent>[
        IepPlanGenerationEvent.error('AI生成未返回计划数据'),
      ];
    }
    return <IepPlanGenerationEvent>[
      IepPlanGenerationEvent.status(task.message),
      IepPlanGenerationEvent.done(
        plan,
        savedPlan: task.savedPlan,
        costAmountCny: task.costAmountCny,
      ),
    ];
  }
  if (task.isFailed) {
    return <IepPlanGenerationEvent>[
      IepPlanGenerationEvent.error(
        task.error.isEmpty ? 'AI生成失败' : task.error,
      ),
    ];
  }
  final List<IepPlanGenerationEvent> events = <IepPlanGenerationEvent>[];
  if (task.message.isNotEmpty) {
    events.add(
      IepPlanGenerationEvent.statusWithCost(task.message, task.costAmountCny),
    );
  }
  if (task.streamText.length > lastStreamText.length &&
      task.streamText.startsWith(lastStreamText)) {
    events.add(
      IepPlanGenerationEvent.deltaWithCost(
        task.streamText.substring(lastStreamText.length),
        task.costAmountCny,
      ),
    );
  } else if (task.streamText.isNotEmpty && task.streamText != lastStreamText) {
    events.add(
      IepPlanGenerationEvent.deltaWithCost(
        task.streamText,
        task.costAmountCny,
      ),
    );
  }
  return events;
}

IepPlanGenerationTask? _taskFromSseFrame(String frame) {
  final List<String> dataLines = frame
      .split(RegExp(r'\r?\n'))
      .where((String line) => line.startsWith('data:'))
      .map((String line) => line.substring(5).trim())
      .toList();
  if (dataLines.isEmpty) {
    return null;
  }
  final Object? decoded = _tryDecodeJson(dataLines.join('\n'));
  if (decoded is! Map) {
    return null;
  }
  final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
  final Map<String, dynamic> taskJson = _mapFrom(payload['data']);
  if (taskJson.isEmpty || !taskJson.containsKey('taskId')) {
    return null;
  }
  return IepPlanGenerationTask.fromJson(taskJson);
}

IepExecutionPlanGenerationEvent<T>? _parseExecutionSseFrame<T>(
  String frame,
  T Function(Map<String, dynamic>) parser,
) {
  final List<String> dataLines = frame
      .split(RegExp(r'\r?\n'))
      .where((String line) => line.startsWith('data:'))
      .map((String line) => line.substring(5).trim())
      .toList();
  if (dataLines.isEmpty) {
    return null;
  }
  final Object? decoded = _tryDecodeJson(dataLines.join('\n'));
  if (decoded is! Map) {
    return null;
  }
  final Map<String, dynamic> payload = Map<String, dynamic>.from(decoded);
  final String type = _stringFrom(payload['type']);
  return switch (type) {
    'status' => IepExecutionPlanGenerationEvent<T>.status(
        _stringFrom(payload['message']),
      ),
    'delta' => IepExecutionPlanGenerationEvent<T>.delta(
        _stringFrom(payload['text']),
      ),
    'done' => IepExecutionPlanGenerationEvent<T>.done(
        parser(_mapFrom(payload['data'])),
        costAmountCny: _doubleFrom(payload['costAmountCny']),
      ),
    'error' => IepExecutionPlanGenerationEvent<T>.error(
        _stringFrom(payload['message']).isEmpty
            ? 'AI生成失败'
            : _stringFrom(payload['message']),
      ),
    _ => null,
  };
}

Object? _tryDecodeJson(String text) {
  final String content = text.trim();
  if (content.isEmpty) {
    return null;
  }
  try {
    return jsonDecode(content);
  } on Object {
    return content;
  }
}
