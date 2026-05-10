part of 'iep_center_page.dart';

enum _IepPreviewMode { total, month, week }

typedef _IepMessageHandler = void Function(
  String message, {
  PadMessageTone tone,
});

class _IepGenerationResult {
  const _IepGenerationResult({
    required this.plan,
    required this.savedPlan,
  });

  final IepPlan plan;
  final IepPlanSaved? savedPlan;
}

class _IepGenerationSessionSnapshot {
  const _IepGenerationSessionSnapshot({
    required this.taskId,
    required this.durationMonths,
    required this.status,
    required this.streamText,
    required this.progress,
  });

  final String taskId;
  final int durationMonths;
  final String status;
  final String streamText;
  final double progress;
}

class _IepWorkspace extends StatefulWidget {
  const _IepWorkspace({
    required this.record,
    required this.planClient,
    required this.queueBootstrapLoading,
    required this.onConfirmAvailabilityChanged,
    required this.onMessage,
  });

  final IepAssessmentRecordSummary? record;
  final IepPlanClient planClient;
  final bool queueBootstrapLoading;
  final ValueChanged<bool> onConfirmAvailabilityChanged;
  final _IepMessageHandler onMessage;

  @override
  State<_IepWorkspace> createState() => _IepWorkspaceState();
}

class _IepWorkspaceState extends State<_IepWorkspace>
    with WidgetsBindingObserver {
  static const String _authTokenStorageKey = 'auth_token';

  _IepPreviewMode _previewMode = _IepPreviewMode.total;
  String _previewMonth = '5月';
  int _previewWeek = 2;
  DateTime _periodStart = DateTime(2026, 5);
  DateTime? _periodEndOverride;
  int _periodMonthCount = 3;
  _GoalEditRequest? _selectedGoal;
  List<_DocDomainData> _totalPlanDomains = <_DocDomainData>[];
  IepPlanSaved? _savedPlan;
  IepExecutionPlansSaved? _executionPlans;
  bool _loadingPlan = false;
  bool _hasCompletedInitialPlanLoad = false;
  bool _syncingPeriod = false;
  bool _generatingPlan = false;
  String _generationStatus = '';
  String _aiStreamText = '';
  double _generationProgress = 0;
  String _planError = '';
  String _activeGenerationTaskId = '';
  String _activeGenerationRecordKey = '';
  int _activeGenerationDurationMonths = 3;
  final Map<String, _IepGenerationSessionSnapshot> _generationSessionsByRecord =
      <String, _IepGenerationSessionSnapshot>{};
  int _loadTicket = 0;
  int _generationTicket = 0;

  DateTime get _periodEnd =>
      _periodEndOverride ?? _periodEndFor(_periodStart, _periodMonthCount);

  List<String> get _periodMonths =>
      _periodMonthLabels(_periodStart, _periodMonthCount);

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _syncPreviewMonthToPeriod();
    if (widget.record != null) {
      runAfterRouteEntrance(context, _loadPlanBundle);
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state != AppLifecycleState.resumed) {
      return;
    }
    Future<void>.delayed(const Duration(milliseconds: 250), () {
      if (!mounted) {
        return;
      }
      _resumeOrRestoreGenerationTask(showMessage: false);
    });
  }

  @override
  void didUpdateWidget(covariant _IepWorkspace oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.record?.id != widget.record?.id ||
        oldWidget.record?.source != widget.record?.source) {
      _storeCurrentGenerationSession(oldWidget.record);
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _totalPlanDomains = <_DocDomainData>[];
      _savedPlan = null;
      _executionPlans = null;
      _planError = '';
      _generationStatus = '';
      _aiStreamText = '';
      _generationProgress = 0;
      _generatingPlan = false;
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = '';
      ++_loadTicket;
      ++_generationTicket;
      widget.onConfirmAvailabilityChanged(false);
      _initPeriodFromRecord(widget.record);
      _syncPreviewMonthToPeriod();
      if (_restoreGenerationSessionFor(widget.record)) {
        return;
      }
      _loadPlanBundle();
    }
  }

  Future<void> _loadPlanBundle() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      setState(() {
        _loadingPlan = false;
        _planError = '';
        _savedPlan = null;
        _executionPlans = null;
        _totalPlanDomains = <_DocDomainData>[];
        _generationStatus = '';
        _aiStreamText = '';
        _generationProgress = 0;
        _generatingPlan = false;
        _activeGenerationTaskId = '';
        _activeGenerationRecordKey = '';
        _hasCompletedInitialPlanLoad = false;
      });
      widget.onConfirmAvailabilityChanged(false);
      return;
    }
    final int ticket = ++_loadTicket;
    setState(() {
      _loadingPlan = true;
      _planError = '';
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanSaved savedPlan = await widget.planClient.fetchIepPlan(
        token,
        record: record,
        durationMonths: _periodMonthCount,
      );
      IepExecutionPlansSaved executionPlans =
          IepExecutionPlansSaved.empty(_periodMonthCount);
      if (savedPlan.hasContent) {
        executionPlans = await widget.planClient.fetchExecutionPlans(
          token,
          record: record,
          durationMonths: savedPlan.durationMonths,
        );
      }
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _hasCompletedInitialPlanLoad = true;
        _savedPlan = savedPlan;
        _executionPlans = executionPlans;
        _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
        _applyPeriodFromPlan(savedPlan.plan, record);
        _totalPlanDomains = savedPlan.plan == null
            ? <_DocDomainData>[]
            : _docDomainsFromPlan(savedPlan.plan!);
        _syncPreviewMonthToPeriod();
      });
      _syncConfirmAvailability(savedPlan);
    } on IepPlanApiException catch (error) {
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _hasCompletedInitialPlanLoad = true;
        _planError = error.message;
      });
      widget.onConfirmAvailabilityChanged(false);
    } on Object catch (error) {
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _hasCompletedInitialPlanLoad = true;
        _planError = 'IEP计划加载失败：$error';
      });
      widget.onConfirmAvailabilityChanged(false);
    }
  }

  void _syncConfirmAvailability(IepPlanSaved? savedPlan) {
    final bool canConfirm = savedPlan?.hasContent == true &&
        savedPlan?.status.trim() != 'confirmed';
    widget.onConfirmAvailabilityChanged(canConfirm);
  }

  IepPlanSaved _draftSavedPlan(IepPlan plan) {
    return IepPlanSaved(
      exists: true,
      status: 'draft',
      durationMonths: _periodMonthCount,
      plan: plan,
      updatedTime: _formatDateDash(DateTime.now()),
    );
  }

  Future<void> _appendDeltaWithTypewriter(
    String delta, {
    required int ticket,
  }) async {
    for (final int codePoint in delta.runes) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      setState(() {
        _aiStreamText += String.fromCharCode(codePoint);
        _generationProgress = math.min(
          .88,
          math.max(_generationProgress, _aiStreamText.length / 1800),
        );
      });
      await Future<void>.delayed(const Duration(milliseconds: 4));
    }
  }

  bool _hasGenerationTaskFor(IepAssessmentRecordSummary? record) {
    return record != null &&
        _activeGenerationTaskId.trim().isNotEmpty &&
        _activeGenerationRecordKey == _recordGenerationKey(record);
  }

  bool get _hasResumableGenerationTask {
    return _hasGenerationTaskFor(widget.record);
  }

  void _storeCurrentGenerationSession(IepAssessmentRecordSummary? record) {
    if (record == null) {
      return;
    }
    final String recordKey = _recordGenerationKey(record);
    if (!_hasGenerationTaskFor(record)) {
      _generationSessionsByRecord.remove(recordKey);
      return;
    }
    _generationSessionsByRecord[recordKey] = _IepGenerationSessionSnapshot(
      taskId: _activeGenerationTaskId,
      durationMonths: _activeGenerationDurationMonths,
      status: _generationStatus,
      streamText: _aiStreamText,
      progress: _generationProgress,
    );
  }

  _IepGenerationSessionSnapshot? _generationSessionFor(
    IepAssessmentRecordSummary? record,
  ) {
    if (record == null) {
      return null;
    }
    return _generationSessionsByRecord[_recordGenerationKey(record)];
  }

  bool _restoreGenerationSessionFor(IepAssessmentRecordSummary? record) {
    if (record == null) {
      return false;
    }
    final _IepGenerationSessionSnapshot? session = _generationSessionFor(record);
    if (session == null) {
      return false;
    }
    setState(() {
      _periodMonthCount = session.durationMonths == 6 ? 6 : 3;
      _periodEndOverride = null;
      _activeGenerationTaskId = session.taskId;
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = session.durationMonths;
      _generationStatus = session.status;
      _aiStreamText = session.streamText;
      _generationProgress = session.progress;
      _generatingPlan = true;
      _planError = '';
      _hasCompletedInitialPlanLoad = true;
      _loadingPlan = false;
    });
    Future<void>.microtask(() {
      if (!mounted) {
        return;
      }
      _resumeIepPlanGenerationTask(showMessage: false);
    });
    return true;
  }

  Future<void> _resumeOrRestoreGenerationTask({
    bool showMessage = false,
  }) async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      return;
    }
    if (_hasResumableGenerationTask) {
      await _resumeIepPlanGenerationTask(showMessage: showMessage);
      return;
    }
    _restoreGenerationSessionFor(record);
  }

  Future<void> _handleGeneratePlanRequest() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (_generatingPlan) {
      return;
    }
    if (_hasResumableGenerationTask) {
      await _resumeIepPlanGenerationTask(showMessage: false);
      return;
    }
    if (_restoreGenerationSessionFor(record)) {
      return;
    }
    await _generateIepPlan();
  }

  Future<void> _handleRetryRequest() async {
    if (_hasResumableGenerationTask || _generationSessionFor(widget.record) != null) {
      await _resumeOrRestoreGenerationTask(showMessage: true);
      return;
    }
    await _loadPlanBundle();
  }

  Future<_IepGenerationResult?> _handleGenerationEvent(
    IepPlanGenerationEvent event, {
    required int ticket,
  }) async {
    switch (event.type) {
      case IepPlanGenerationEventType.status:
        setState(() {
          _generationStatus =
              event.message.trim().isEmpty ? '正在生成IEP计划' : event.message.trim();
        });
      case IepPlanGenerationEventType.delta:
        final String delta = _generationDeltaAfterExistingText(event.text);
        if (delta.isEmpty) {
          break;
        }
        await _appendDeltaWithTypewriter(
          delta,
          ticket: ticket,
        );
        if (!mounted || ticket != _generationTicket) {
          return null;
        }
        setState(() {
          _generationStatus = 'AI正在生成IEP计划';
        });
      case IepPlanGenerationEventType.done:
        final IepPlan? plan = event.plan;
        if (plan == null) {
          throw const IepPlanApiException('AI生成未返回计划数据');
        }
        setState(() {
          _generationProgress = .94;
          _generationStatus = '生成完成，正在自动保存草稿';
        });
        return _IepGenerationResult(plan: plan, savedPlan: event.savedPlan);
      case IepPlanGenerationEventType.error:
        throw IepPlanApiException(
          event.message.trim().isEmpty ? 'AI生成失败' : event.message.trim(),
        );
    }
    return null;
  }

  String _generationDeltaAfterExistingText(String incoming) {
    if (incoming.isEmpty || _aiStreamText.isEmpty) {
      return incoming;
    }
    if (incoming.startsWith(_aiStreamText)) {
      return incoming.substring(_aiStreamText.length);
    }
    if (_aiStreamText.endsWith(incoming)) {
      return '';
    }
    final int maxOverlap = math.min(_aiStreamText.length, incoming.length);
    for (int length = maxOverlap; length > 0; length -= 1) {
      if (_aiStreamText.endsWith(incoming.substring(0, length))) {
        return incoming.substring(length);
      }
    }
    return incoming;
  }

  Future<void> _generateIepPlan({bool forceRegenerate = false}) async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (_generatingPlan) {
      return;
    }
    final int ticket = ++_generationTicket;
    ++_loadTicket;
    _generationSessionsByRecord.remove(_recordGenerationKey(record));
    IepPlan? finalPlan;
    IepPlanSaved? savedPlanFromTask;
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _loadingPlan = false;
      _generatingPlan = true;
      _generationStatus = forceRegenerate ? '正在重新生成IEP计划' : '正在准备AI生成';
      _aiStreamText = '';
      _generationProgress = .08;
      _planError = '';
      _executionPlans = IepExecutionPlansSaved.empty(_periodMonthCount);
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = _periodMonthCount;
      if (forceRegenerate) {
        _savedPlan = null;
        _totalPlanDomains = <_DocDomainData>[];
      }
    });
    widget.onConfirmAvailabilityChanged(false);

    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanGenerationTask task =
          await widget.planClient.createIepPlanGenerationTask(
        token,
        record: record,
        durationMonths: _periodMonthCount,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      setState(() {
        _activeGenerationTaskId = task.taskId;
        _activeGenerationDurationMonths =
            task.durationMonths == 6 ? 6 : _periodMonthCount;
        _generationStatus =
            task.message.trim().isEmpty ? '正在生成IEP计划' : task.message.trim();
      });
      await for (final IepPlanGenerationEvent event
          in widget.planClient.watchIepPlanGenerationTask(
        token,
        record: record,
        taskId: task.taskId,
      )) {
        if (!mounted || ticket != _generationTicket) {
          return;
        }
        final _IepGenerationResult? result =
            await _handleGenerationEvent(event, ticket: ticket);
        finalPlan = result?.plan ?? finalPlan;
        savedPlanFromTask = result?.savedPlan ?? savedPlanFromTask;
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      if (finalPlan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      _applyGeneratedPlan(
        record: record,
        plan: finalPlan,
        savedPlanFromTask: savedPlanFromTask,
      );
      _showMessage('AI生成成功，已自动保存草稿', tone: PadMessageTone.success);
    } on IepPlanApiException catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成失败';
        _planError = _savedPlan?.hasContent == true ? '' : error.message;
      });
      _syncConfirmAvailability(_savedPlan);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final String message = 'AI生成失败：$error';
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成失败';
        _planError = _savedPlan?.hasContent == true ? '' : message;
      });
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
  }

  Future<void> _resumeIepPlanGenerationTask({bool showMessage = true}) async {
    final IepAssessmentRecordSummary? record = widget.record;
    final String taskId = _activeGenerationTaskId.trim();
    if (record == null || taskId.isEmpty) {
      return;
    }
    if (_activeGenerationRecordKey != _recordGenerationKey(record)) {
      return;
    }
    final int ticket = ++_generationTicket;
    ++_loadTicket;
    IepPlan? finalPlan;
    IepPlanSaved? savedPlanFromTask;
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _loadingPlan = false;
      _generatingPlan = true;
      _planError = '';
      _generationStatus = '正在重新连接AI生成任务';
      if (_generationProgress < .12) {
        _generationProgress = .12;
      }
    });
    widget.onConfirmAvailabilityChanged(false);
    if (showMessage) {
      _showMessage('正在重新连接AI生成任务');
    }

    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanGenerationTask latest =
          await widget.planClient.fetchIepPlanGenerationTask(
        token,
        record: record,
        taskId: taskId,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      _activeGenerationDurationMonths =
          latest.durationMonths == 6 ? 6 : _activeGenerationDurationMonths;
      final _IepGenerationResult? latestResult =
          await _applyGenerationTaskSnapshot(latest, ticket: ticket);
      finalPlan = latestResult?.plan ?? finalPlan;
      savedPlanFromTask = latestResult?.savedPlan ?? savedPlanFromTask;
      if (latestResult == null && !latest.isDone) {
        await for (final IepPlanGenerationEvent event
            in widget.planClient.watchIepPlanGenerationTask(
          token,
          record: record,
          taskId: taskId,
        )) {
          if (!mounted || ticket != _generationTicket) {
            return;
          }
          final _IepGenerationResult? result =
              await _handleGenerationEvent(event, ticket: ticket);
          finalPlan = result?.plan ?? finalPlan;
          savedPlanFromTask = result?.savedPlan ?? savedPlanFromTask;
        }
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      if (finalPlan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      _applyGeneratedPlan(
        record: record,
        plan: finalPlan,
        savedPlanFromTask: savedPlanFromTask,
      );
      _showMessage('AI生成成功，已自动保存草稿', tone: PadMessageTone.success);
    } on IepPlanApiException catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成连接已断开';
        _planError = _savedPlan?.hasContent == true ? '' : error.message;
      });
      _syncConfirmAvailability(_savedPlan);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final String message = 'AI生成重连失败：$error';
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成连接已断开';
        _planError = _savedPlan?.hasContent == true ? '' : message;
      });
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
  }

  Future<_IepGenerationResult?> _applyGenerationTaskSnapshot(
    IepPlanGenerationTask task, {
    required int ticket,
  }) async {
    if (task.streamText.isNotEmpty && task.streamText != _aiStreamText) {
      setState(() {
        _aiStreamText = task.streamText;
        _generationProgress = math.min(
          .88,
          math.max(_generationProgress, _aiStreamText.length / 1800),
        );
      });
    }
    if (task.isFailed) {
      throw IepPlanApiException(task.error.isEmpty ? 'AI生成失败' : task.error);
    }
    if (task.isDone) {
      final IepPlan? plan = task.savedPlan?.plan ?? task.plan;
      if (plan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      setState(() {
        _generationProgress = .94;
        _generationStatus =
            task.message.trim().isEmpty ? '生成完成，正在自动保存草稿' : task.message.trim();
      });
      return _IepGenerationResult(plan: plan, savedPlan: task.savedPlan);
    }
    setState(() {
      _generationStatus =
          task.message.trim().isEmpty ? 'AI正在生成IEP计划' : task.message.trim();
    });
    return null;
  }

  void _applyGeneratedPlan({
    required IepAssessmentRecordSummary record,
    required IepPlan plan,
    required IepPlanSaved? savedPlanFromTask,
  }) {
    _generationSessionsByRecord.remove(_recordGenerationKey(record));
    final IepPlanSaved savedPlan = savedPlanFromTask ?? _draftSavedPlan(plan);
    setState(() {
      _generatingPlan = false;
      _generationStatus = '';
      _aiStreamText = '';
      _generationProgress = 1;
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = '';
      _savedPlan = savedPlan;
      _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
      _applyPeriodFromPlan(savedPlan.plan, record);
      _totalPlanDomains = savedPlan.plan == null
          ? <_DocDomainData>[]
          : _docDomainsFromPlan(savedPlan.plan!);
      _syncPreviewMonthToPeriod();
    });
    _syncConfirmAvailability(savedPlan);
  }

  String _recordGenerationKey(IepAssessmentRecordSummary record) {
    return '${record.source.trim().toUpperCase()}-${record.id}';
  }

  void _initPeriodFromRecord(IepAssessmentRecordSummary? record) {
    final DateTime? assessmentDate =
        DateTime.tryParse(record?.assessmentDate.trim() ?? '');
    if (assessmentDate == null) {
      return;
    }
    _periodStart = DateTime(assessmentDate.year, assessmentDate.month);
    _periodEndOverride = null;
  }

  void _applyPeriodFromPlan(IepPlan? plan, IepAssessmentRecordSummary record) {
    final DateTime? planStart = DateTime.tryParse(plan?.meta.startDate ?? '');
    final DateTime? planEnd = DateTime.tryParse(plan?.meta.endDate ?? '');
    if (planStart != null) {
      _periodStart = _dateOnly(planStart);
    } else {
      _initPeriodFromRecord(record);
    }
    _periodEndOverride = planEnd == null ? null : _dateOnly(planEnd);
  }

  void _syncPreviewMonthToPeriod() {
    final List<String> months = _periodMonths;
    if (months.isEmpty) {
      return;
    }
    if (!months.contains(_previewMonth)) {
      _previewMonth = months.first;
      _previewWeek = 1;
    }
  }

  int _previewMonthIndex() {
    final int index = _periodMonths.indexOf(_previewMonth);
    return index < 0 ? 1 : index + 1;
  }

  Future<void> _showEditPeriodDialog() async {
    if (_syncingPeriod || _generatingPlan) {
      return;
    }
    final _IepPeriodDraft? draft = await showDialog<_IepPeriodDraft>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepPeriodEditDialog(
            initialStart: _periodStart,
            monthCount: _periodMonthCount,
          ),
        );
      },
    );
    if (draft == null || !mounted) {
      return;
    }
    await _syncPeriodStart(draft.start);
  }

  Future<void> _showRegeneratePlanConfirmDialog() async {
    if (_generatingPlan) {
      return;
    }
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return const PadDialogViewport(child: _IepRegenerateConfirmDialog());
      },
    );
    if (confirmed != true || !mounted) {
      return;
    }
    await _generateIepPlan(forceRegenerate: true);
  }

  Future<void> _syncPeriodStart(DateTime start) async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (_savedPlan?.hasContent != true) {
      _showMessage('请先生成IEP计划后再调整周期');
      return;
    }
    final DateTime nextStart = _dateOnly(start);
    if (_dateOnly(_periodStart) == nextStart) {
      return;
    }
    ++_loadTicket;
    final int sourceDurationMonths = _savedPlan?.durationMonths == 6 ? 6 : 3;
    setState(() {
      _syncingPeriod = true;
      _planError = '';
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanPeriodSyncResult result =
          await widget.planClient.syncIepPlanPeriod(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        sourceDurationMonths: sourceDurationMonths,
        startDate: nextStart,
      );
      if (!mounted) {
        return;
      }
      _applySyncedPlanBundle(result.iepPlan, result.executionPlans, record);
      _showMessage('计划周期和关联月/周计划日期已同步保存', tone: PadMessageTone.success);
    } on IepPlanApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _syncingPeriod = false;
        _planError = error.message;
      });
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      final String message = '同步计划周期失败：$error';
      setState(() {
        _syncingPeriod = false;
        _planError = message;
      });
      _showMessage(message);
    }
  }

  void _applySyncedPlanBundle(
    IepPlanSaved savedPlan,
    IepExecutionPlansSaved executionPlans,
    IepAssessmentRecordSummary record,
  ) {
    setState(() {
      _syncingPeriod = false;
      _loadingPlan = false;
      _hasCompletedInitialPlanLoad = true;
      _savedPlan = savedPlan;
      _executionPlans = executionPlans;
      _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
      _applyPeriodFromPlan(savedPlan.plan, record);
      _totalPlanDomains = savedPlan.plan == null
          ? <_DocDomainData>[]
          : _docDomainsFromPlan(savedPlan.plan!);
      _syncPreviewMonthToPeriod();
      _ensurePreviewWeekInRange();
    });
    _syncConfirmAvailability(savedPlan);
  }

  void _ensurePreviewWeekInRange() {
    if (_previewMode != _IepPreviewMode.week) {
      return;
    }
    final DateTime monthDate = _monthDateFromLabel(
      _periodStart,
      _periodMonthCount,
      _previewMonth,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: monthDate,
    );
    if (_weekDatesInMonthRange(monthRange, _previewWeek).isEmpty) {
      _previewWeek = 1;
    }
  }

  void _showMessage(String message,
      {PadMessageTone tone = PadMessageTone.info}) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    widget.onMessage(message, tone: tone);
  }

  void _showTotalPlan() {
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
    });
  }

  void _showMonthPlan(String month) {
    setState(() {
      _previewMode = _IepPreviewMode.month;
      _previewMonth = month;
      _selectedGoal = null;
    });
  }

  void _showWeekPlan(String month, int weekNumber) {
    final DateTime monthDate = _monthDateFromLabel(
      _periodStart,
      _periodMonthCount,
      month,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: monthDate,
    );
    final int week = _weekDatesInMonthRange(monthRange, weekNumber).isEmpty
        ? _lastAvailableWeekInMonthRange(monthRange)
        : weekNumber;
    setState(() {
      _previewMode = _IepPreviewMode.week;
      _previewMonth = month;
      _previewWeek = week;
      _selectedGoal = null;
    });
  }

  void _changePeriodMonthCount(int monthCount) {
    if (_periodMonthCount == monthCount || _generatingPlan) {
      return;
    }
    setState(() {
      _periodMonthCount = monthCount;
      _periodEndOverride = null;
      final List<String> months = _periodMonths;
      if (!months.contains(_previewMonth)) {
        _previewMonth = months.first;
        _previewWeek = 1;
      }
      _ensurePreviewWeekInRange();
    });
    _loadPlanBundle();
  }

  Future<void> _showGoalEditDialog(_GoalEditRequest request) async {
    setState(() {
      _selectedGoal = request;
    });
    final _DocDomainData domain = _totalPlanDomains[request.domainIndex];
    final _GoalEditResult? result = await showDialog<_GoalEditResult>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepGoalEditDialog(
            domain: domain,
            request: request,
          ),
        );
      },
    );
    if (result == null || !mounted) {
      return;
    }
    setState(() {
      final List<_DocDomainData> nextDomains =
          List<_DocDomainData>.from(_totalPlanDomains);
      if (result.longGoals != null) {
        nextDomains[request.domainIndex] =
            domain.copyWith(longGoals: result.longGoals);
      }
      if (result.shortGoals != null) {
        nextDomains[request.domainIndex] =
            domain.copyWith(shortGoals: result.shortGoals);
      }
      _totalPlanDomains = nextDomains;
    });
  }

  void _clearSelectedGoal() {
    if (_selectedGoal == null) {
      return;
    }
    setState(() {
      _selectedGoal = null;
    });
  }

  void _handleGoalTap(_GoalEditRequest request) {
    if (_selectedGoal == request) {
      _showGoalEditDialog(request);
      return;
    }
    setState(() {
      _selectedGoal = request;
    });
  }

  @override
  Widget build(BuildContext context) {
    final IepAssessmentRecordSummary? record = widget.record;
    final IepPlan? plan = _savedPlan?.plan;
    final IepMonthlyPlan? monthPlan =
        _executionPlans?.monthPlan(_previewMonthIndex());
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final String title = _workspaceTitle(record, plan);
    final String statusText = _planStatusText(_savedPlan?.status);
    return Container(
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: <Widget>[
          _WorkspaceHeader(
            title: title,
            statusText: statusText,
            periodText: _formatDotRange(_periodStart, _periodEnd),
          ),
          const SizedBox(height: 10),
          _PlanToolbar(
            onShowTotalPlan: _showTotalPlan,
            onShowMonthPlan: _showMonthPlan,
            onShowWeekPlan: _showWeekPlan,
            onEditPeriod: _showEditPeriodDialog,
            onRegeneratePlan: _showRegeneratePlanConfirmDialog,
            monthLabels: _periodMonths,
            periodMonthCount: _periodMonthCount,
            periodStart: _periodStart,
            onPeriodMonthCountChanged: _changePeriodMonthCount,
            syncingPeriod: _syncingPeriod,
            generatingPlan: _generatingPlan,
          ),
          const SizedBox(height: 10),
          Expanded(
            child: _IepTablePreview(
              previewMode: _previewMode,
              month: _previewMonth,
              weekNumber: _previewWeek,
              periodStart: _periodStart,
              periodMonthCount: _periodMonthCount,
              record: record,
              plan: plan,
              monthPlan: monthPlan,
              weekPlan: weekPlan,
              loading: _loadingPlan,
              bootstrapLoading: !_hasCompletedInitialPlanLoad,
              generatingPlan: _generatingPlan,
              generationStatus: _generationStatus,
              generationText: _aiStreamText,
              generationProgress: _generationProgress,
              error: _planError,
              onRetry: _handleRetryRequest,
              onGeneratePlan: _handleGeneratePlanRequest,
              totalPlanDomains: _totalPlanDomains,
              selectedGoal: _selectedGoal,
              onGoalTap: _handleGoalTap,
              onClearSelectedGoal: _clearSelectedGoal,
            ),
          ),
        ],
      ),
    );
  }
}

class _WorkspaceHeader extends StatelessWidget {
  const _WorkspaceHeader({
    required this.title,
    required this.statusText,
    required this.periodText,
  });

  final String title;
  final String statusText;
  final String periodText;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              title,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 19,
                fontWeight: FontWeight.w900,
                height: 1,
              ),
            ),
          ),
          _HeaderMetaPill(
            icon: statusText == '已确认'
                ? Icons.verified_rounded
                : Icons.pending_actions_rounded,
            text: statusText,
            iconColor:
                statusText == '已确认' ? _IepColors.green : _IepColors.yellow,
          ),
          const SizedBox(width: 10),
          _HeaderMetaPill(
            icon: Icons.date_range_rounded,
            text: periodText,
          ),
          const SizedBox(width: 10),
          const _ClassContextPill(),
          const SizedBox(width: 10),
          const _StartClassButton(),
        ],
      ),
    );
  }
}

class _ClassContextPill extends StatelessWidget {
  const _ClassContextPill();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF6EE),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: const Color(0xFFFFD3BA)),
      ),
      child: const Text(
        '第2月 · 第1周',
        style: TextStyle(
          color: _IepColors.orangeDeep,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _HeaderMetaPill extends StatelessWidget {
  const _HeaderMetaPill({
    required this.icon,
    required this.text,
    this.iconColor = _IepColors.muted,
  });

  final IconData icon;
  final String text;
  final Color iconColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, size: 15, color: iconColor),
          const SizedBox(width: 5),
          Text(
            text,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _StartClassButton extends StatelessWidget {
  const _StartClassButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 15),
      decoration: BoxDecoration(
        color: _IepColors.orange,
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(
          color: const Color(0x32E96F43),
          blur: 12,
          offset: const Offset(0, 5),
        ),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.play_circle_fill_rounded, color: Colors.white, size: 19),
          SizedBox(width: 6),
          Text(
            '开始上课',
            style: TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PlanToolbar extends StatelessWidget {
  const _PlanToolbar({
    required this.onShowTotalPlan,
    required this.onShowMonthPlan,
    required this.onShowWeekPlan,
    required this.onEditPeriod,
    required this.onRegeneratePlan,
    required this.monthLabels,
    required this.periodMonthCount,
    required this.periodStart,
    required this.onPeriodMonthCountChanged,
    required this.syncingPeriod,
    required this.generatingPlan,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final VoidCallback onEditPeriod;
  final VoidCallback onRegeneratePlan;
  final List<String> monthLabels;
  final int periodMonthCount;
  final DateTime periodStart;
  final ValueChanged<int> onPeriodMonthCountChanged;
  final bool syncingPeriod;
  final bool generatingPlan;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          _PeriodSwitch(
            selectedMonthCount: periodMonthCount,
            onChanged: onPeriodMonthCountChanged,
          ),
          const _ToolbarDivider(),
          Expanded(
            child: _ScrollablePlanNav(
              onShowTotalPlan: onShowTotalPlan,
              onShowMonthPlan: onShowMonthPlan,
              onShowWeekPlan: onShowWeekPlan,
              monthLabels: monthLabels,
              periodStart: periodStart,
              periodMonthCount: periodMonthCount,
            ),
          ),
          const _ToolbarDivider(),
          _TableTinyAction(
            icon: syncingPeriod || generatingPlan
                ? Icons.hourglass_top_rounded
                : Icons.edit_calendar_rounded,
            label: syncingPeriod ? '同步中' : '编辑周期',
            onTap: syncingPeriod || generatingPlan ? null : onEditPeriod,
          ),
          const SizedBox(width: 8),
          _TableTinyAction(
            icon: generatingPlan
                ? Icons.hourglass_top_rounded
                : Icons.refresh_rounded,
            label: generatingPlan ? '生成中' : '重新生成',
            onTap: generatingPlan ? null : onRegeneratePlan,
          ),
        ],
      ),
    );
  }
}

class _ScrollablePlanNav extends StatefulWidget {
  const _ScrollablePlanNav({
    required this.onShowTotalPlan,
    required this.onShowMonthPlan,
    required this.onShowWeekPlan,
    required this.monthLabels,
    required this.periodStart,
    required this.periodMonthCount,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final List<String> monthLabels;
  final DateTime periodStart;
  final int periodMonthCount;

  @override
  State<_ScrollablePlanNav> createState() => _ScrollablePlanNavState();
}

class _ScrollablePlanNavState extends State<_ScrollablePlanNav>
    with SingleTickerProviderStateMixin {
  late final ScrollController _scrollController;
  late final AnimationController _hintController;
  late final Animation<double> _hintOffset;
  bool _showLeftHint = false;
  bool _showRightHint = false;
  String _selectedSection = 'iep';
  String _selectedMonth = '5月';
  int? _selectedWeek;

  @override
  void didUpdateWidget(covariant _ScrollablePlanNav oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!widget.monthLabels.contains(_selectedMonth) &&
        widget.monthLabels.isNotEmpty) {
      _selectedMonth = widget.monthLabels.first;
      _selectedWeek = null;
    }
    final int maxWeek = _weekCountForSelectedMonth();
    if (_selectedWeek != null && _selectedWeek! > maxWeek) {
      _selectedWeek = maxWeek;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    _scrollController.addListener(_syncHints);
    _hintController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 760),
    )..repeat(reverse: true);
    _hintOffset = CurvedAnimation(
      parent: _hintController,
      curve: Curves.easeInOutCubic,
    );
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _syncHints() {
    if (!_scrollController.hasClients) {
      return;
    }
    final ScrollPosition position = _scrollController.position;
    final bool canScroll = position.maxScrollExtent > 1;
    final bool nextLeft = canScroll && position.pixels > 1;
    final bool nextRight =
        canScroll && position.pixels < position.maxScrollExtent - 1;
    if (_showLeftHint == nextLeft && _showRightHint == nextRight) {
      return;
    }
    setState(() {
      _showLeftHint = nextLeft;
      _showRightHint = nextRight;
    });
  }

  void _selectPlan(String plan) {
    setState(() {
      _selectedSection = plan;
      _selectedWeek = null;
    });
    if (plan == 'iep') {
      widget.onShowTotalPlan();
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _selectMonth(String month) {
    setState(() {
      _selectedSection = 'month';
      _selectedMonth = month;
      _selectedWeek = null;
    });
    widget.onShowMonthPlan(month);
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _selectWeek(int weekNumber) {
    setState(() {
      _selectedSection = 'week';
      _selectedWeek = weekNumber;
    });
    widget.onShowWeekPlan(_selectedMonth, weekNumber);
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  @override
  void dispose() {
    _scrollController.removeListener(_syncHints);
    _scrollController.dispose();
    _hintController.dispose();
    super.dispose();
  }

  int _weekCountForSelectedMonth() {
    final DateTime selectedMonthDate = _monthDateFromLabel(
      widget.periodStart,
      widget.periodMonthCount,
      _selectedMonth,
    );
    final DateTimeRange selectedMonthRange = _monthRangeInPeriod(
      periodStart: widget.periodStart,
      monthCount: widget.periodMonthCount,
      monthDate: selectedMonthDate,
    );
    return _lastAvailableWeekInMonthRange(selectedMonthRange);
  }

  @override
  Widget build(BuildContext context) {
    final int weekCount = _weekCountForSelectedMonth();
    return Center(
      child: SizedBox(
        height: 34,
        child: Stack(
          alignment: Alignment.center,
          children: <Widget>[
            Center(
              child: SingleChildScrollView(
                controller: _scrollController,
                scrollDirection: Axis.horizontal,
                physics: const BouncingScrollPhysics(
                  parent: ClampingScrollPhysics(),
                ),
                padding: const EdgeInsets.only(right: 2),
                child: SizedBox(
                  height: 34,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    children: <Widget>[
                      _PlanTab(
                        text: 'IEP总计划',
                        active: _selectedSection == 'iep',
                        width: 92,
                        onTap: () => _selectPlan('iep'),
                      ),
                      const _PlanNavLabel(text: '月计划'),
                      ...widget.monthLabels.map((String month) {
                        return _PlanTab(
                          text: month,
                          active: _selectedSection == 'month' &&
                              _selectedMonth == month,
                          width: 54,
                          onTap: () => _selectMonth(month),
                        );
                      }),
                      const _PlanNavLabel(text: '周计划'),
                      ...List<Widget>.generate(weekCount, (int index) {
                        final int weekNumber = index + 1;
                        return _PlanTab(
                          text: '${_selectedMonth} W$weekNumber',
                          width: 68,
                          active: _selectedSection == 'week' &&
                              _selectedWeek == weekNumber,
                          activeTone: _PlanTabTone.week,
                          rightGap: weekNumber == weekCount ? 2 : 6,
                          onTap: () => _selectWeek(weekNumber),
                        );
                      }),
                    ],
                  ),
                ),
              ),
            ),
            Align(
              alignment: Alignment.centerLeft,
              child: _PlanScrollHint(
                visible: _showLeftHint,
                alignment: Alignment.centerLeft,
                direction: AxisDirection.left,
                animation: _hintOffset,
              ),
            ),
            Align(
              alignment: Alignment.centerRight,
              child: _PlanScrollHint(
                visible: _showRightHint,
                alignment: Alignment.centerRight,
                direction: AxisDirection.right,
                animation: _hintOffset,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PlanScrollHint extends StatelessWidget {
  const _PlanScrollHint({
    required this.visible,
    required this.alignment,
    required this.direction,
    required this.animation,
  });

  final bool visible;
  final Alignment alignment;
  final AxisDirection direction;
  final Animation<double> animation;

  @override
  Widget build(BuildContext context) {
    final bool right = direction == AxisDirection.right;
    return IgnorePointer(
      child: AnimatedOpacity(
        opacity: visible ? 1 : 0,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        child: Container(
          width: 62,
          height: 34,
          alignment: alignment,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: right ? Alignment.centerLeft : Alignment.centerRight,
              end: right ? Alignment.centerRight : Alignment.centerLeft,
              colors: const <Color>[
                Color(0x00FFFAF5),
                Color(0xEFFFFAF5),
                Color(0xFFFFFAF5),
              ],
            ),
          ),
          child: AnimatedBuilder(
            animation: animation,
            builder: (BuildContext context, Widget? child) {
              final double dx = (animation.value * 5 + 1) * (right ? 1 : -1);
              return Transform.translate(
                offset: Offset(dx, 0),
                child: child,
              );
            },
            child: Padding(
              padding: EdgeInsets.only(
                left: right ? 0 : 4,
                right: right ? 4 : 0,
              ),
              child: Icon(
                right
                    ? Icons.chevron_right_rounded
                    : Icons.chevron_left_rounded,
                size: 27,
                color: _IepColors.orangeDeep.withOpacity(.86),
                shadows: const <Shadow>[
                  Shadow(
                    color: Color(0x22E96F43),
                    blurRadius: 5,
                    offset: Offset(0, 2),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PeriodSwitch extends StatelessWidget {
  const _PeriodSwitch({
    required this.selectedMonthCount,
    required this.onChanged,
  });

  final int selectedMonthCount;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.all(2),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          _PeriodOption(
            text: '3个月',
            active: selectedMonthCount == 3,
            onTap: () => onChanged(3),
          ),
          _PeriodOption(
            text: '6个月',
            active: selectedMonthCount == 6,
            onTap: () => onChanged(6),
          ),
        ],
      ),
    );
  }
}

class _PeriodOption extends StatelessWidget {
  const _PeriodOption({
    required this.text,
    required this.onTap,
    this.active = false,
  });

  final String text;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 26,
          width: 50,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: active ? _IepColors.orange : Colors.transparent,
            borderRadius: BorderRadius.circular(13),
          ),
          child: Text(
            text,
            style: TextStyle(
              color: active ? Colors.white : _IepColors.text,
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _ToolbarDivider extends StatelessWidget {
  const _ToolbarDivider();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 1,
      height: 24,
      margin: const EdgeInsets.symmetric(horizontal: 8),
      color: _IepColors.lightLine,
    );
  }
}

class _TableTinyAction extends StatelessWidget {
  const _TableTinyAction({
    required this.icon,
    required this.label,
    this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(15),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 9),
          decoration: BoxDecoration(
            color: onTap == null ? const Color(0xFFF8EEE6) : Colors.white,
            borderRadius: BorderRadius.circular(15),
            border: Border.all(color: _IepColors.line),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              Icon(
                icon,
                color: onTap == null ? _IepColors.muted : _IepColors.text,
                size: 16,
              ),
              const SizedBox(width: 5),
              Text(
                label,
                style: TextStyle(
                  color: onTap == null ? _IepColors.muted : _IepColors.text,
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
