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
    required this.costAmountCny,
  });

  final IepPlan plan;
  final IepPlanSaved? savedPlan;
  final double costAmountCny;
}

class _ExecutionPlanGenerationResult<T> {
  const _ExecutionPlanGenerationResult(
    this.plan,
    this.costAmountCny, {
    this.savedExecutionPlans,
  });

  final T plan;
  final double costAmountCny;
  final IepExecutionPlansSaved? savedExecutionPlans;
}

class _IepGenerationSessionSnapshot {
  const _IepGenerationSessionSnapshot({
    required this.taskId,
    required this.durationMonths,
    required this.previewMode,
    required this.monthLabel,
    required this.weekIndex,
    required this.restWeekdays,
    required this.status,
    required this.streamText,
    required this.progress,
    required this.costAmountCny,
  });

  final String taskId;
  final int durationMonths;
  final _IepPreviewMode previewMode;
  final String monthLabel;
  final int weekIndex;
  final List<int> restWeekdays;
  final String status;
  final String streamText;
  final double progress;
  final double costAmountCny;
}

class _IepWorkspace extends StatefulWidget {
  const _IepWorkspace({
    super.key,
    required this.record,
    required this.planClient,
    required this.queueBootstrapLoading,
    required this.onConfirmAvailabilityChanged,
    required this.onRecordStatusChanged,
    required this.onMessage,
  });

  final IepAssessmentRecordSummary? record;
  final IepPlanClient planClient;
  final bool queueBootstrapLoading;
  final ValueChanged<bool> onConfirmAvailabilityChanged;
  final void Function(IepAssessmentRecordSummary record, String status)
      onRecordStatusChanged;
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
  int _periodMonthCount = 6;
  _GoalEditRequest? _selectedGoal;
  List<_DocDomainData> _totalPlanDomains = <_DocDomainData>[];
  IepPlanSaved? _savedPlan;
  IepExecutionPlansSaved? _executionPlans;
  bool _loadingPlan = false;
  bool _hasCompletedInitialPlanLoad = false;
  bool _syncingPeriod = false;
  bool _confirmingPlan = false;
  bool _generatingPlan = false;
  String _generationStatus = '';
  String _aiStreamText = '';
  double _generationProgress = 0;
  double _generationCostAmountCny = 0;
  String _planError = '';
  String _activeGenerationTaskId = '';
  String _activeGenerationRecordKey = '';
  int _activeGenerationDurationMonths = 6;
  _IepPreviewMode _activeGenerationPreviewMode = _IepPreviewMode.total;
  String _activeGenerationMonthLabel = '';
  int _activeGenerationWeekIndex = 0;
  List<int> _activeWeeklyRestWeekdays = <int>[DateTime.sunday];
  final Map<int, List<int>> _monthlyRestWeekdayOverrides = <int, List<int>>{};
  final Map<String, _IepGenerationSessionSnapshot> _generationSessionsByRecord =
      <String, _IepGenerationSessionSnapshot>{};
  IepLessonSessionWeekState? _lessonSessionState;
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
      if (_previewMode == _IepPreviewMode.week) {
        unawaited(_refreshLessonSessionState());
      }
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
      _lessonSessionState = null;
      _planError = '';
      _generationStatus = '';
      _aiStreamText = '';
      _generationProgress = 0;
      _generationCostAmountCny = 0;
      _generatingPlan = false;
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = '';
      _activeGenerationPreviewMode = _IepPreviewMode.total;
      _activeGenerationMonthLabel = '';
      _activeGenerationWeekIndex = 0;
      _activeWeeklyRestWeekdays = <int>[DateTime.sunday];
      _monthlyRestWeekdayOverrides.clear();
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
        _lessonSessionState = null;
        _totalPlanDomains = <_DocDomainData>[];
        _generationStatus = '';
        _aiStreamText = '';
        _generationProgress = 0;
        _generationCostAmountCny = 0;
        _generatingPlan = false;
        _activeGenerationTaskId = '';
        _activeGenerationRecordKey = '';
        _activeGenerationPreviewMode = _IepPreviewMode.total;
        _activeGenerationMonthLabel = '';
        _activeGenerationWeekIndex = 0;
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
      final IepPlanGenerationTask? activeTask =
          await widget.planClient.fetchActiveIepPlanGenerationTask(
        token,
        record: record,
      );
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      if (activeTask != null &&
          !activeTask.isDone &&
          !activeTask.isFailed &&
          activeTask.taskId.trim().isNotEmpty) {
        final _IepGenerationSessionSnapshot session =
            _IepGenerationSessionSnapshot(
          taskId: activeTask.taskId,
          durationMonths: activeTask.durationMonths == 6 ? 6 : 3,
          previewMode: _IepPreviewMode.total,
          monthLabel: '',
          weekIndex: 0,
          restWeekdays: const <int>[DateTime.sunday],
          status: activeTask.message.trim().isEmpty
              ? 'AI正在生成IEP计划'
              : activeTask.message.trim(),
          streamText: activeTask.streamText,
          progress: _streamGenerationProgress(activeTask.streamText),
          costAmountCny: activeTask.costAmountCny,
        );
        _generationSessionsByRecord[_recordGenerationKey(record)] = session;
        if (_restoreGenerationSessionFor(record)) {
          return;
        }
      }
      final IepExecutionPlanGenerationTask? activeExecutionTask =
          await _fetchActiveExecutionGenerationTask(token, record);
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      if (activeExecutionTask != null &&
          !activeExecutionTask.isDone &&
          !activeExecutionTask.isFailed &&
          activeExecutionTask.taskId.trim().isNotEmpty) {
        final _IepPreviewMode previewMode =
            activeExecutionTask.planType == 'weekly'
                ? _IepPreviewMode.week
                : _IepPreviewMode.month;
        final List<String> months = _periodMonths;
        final String monthLabel = activeExecutionTask.targetMonthIndex > 0 &&
                activeExecutionTask.targetMonthIndex <= months.length
            ? months[activeExecutionTask.targetMonthIndex - 1]
            : _previewMonth;
        final _IepGenerationSessionSnapshot session =
            _IepGenerationSessionSnapshot(
          taskId: activeExecutionTask.taskId,
          durationMonths: activeExecutionTask.durationMonths == 6 ? 6 : 3,
          previewMode: previewMode,
          monthLabel: monthLabel,
          weekIndex: activeExecutionTask.targetWeekIndex,
          restWeekdays: activeExecutionTask.restWeekdays,
          status: activeExecutionTask.message.trim().isEmpty
              ? (previewMode == _IepPreviewMode.week
                  ? 'AI正在生成周计划'
                  : 'AI正在生成月计划')
              : activeExecutionTask.message.trim(),
          streamText: activeExecutionTask.streamText,
          progress: _streamGenerationProgress(activeExecutionTask.streamText),
          costAmountCny: activeExecutionTask.costAmountCny,
        );
        _generationSessionsByRecord[_recordGenerationKey(record)] = session;
        if (_restoreGenerationSessionFor(record)) {
          return;
        }
      }
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
      _syncMonthlyRestWeekdayOverridesFromExecutionPlans(executionPlans);
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
      if (_previewMode == _IepPreviewMode.week) {
        unawaited(_refreshLessonSessionState());
      }
      _notifyRecordStatus(record, savedPlan.status);
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

  Future<IepExecutionPlanGenerationTask?> _fetchActiveExecutionGenerationTask(
    String token,
    IepAssessmentRecordSummary record,
  ) async {
    final int monthIndex = _previewMonthIndex();
    final IepExecutionPlanGenerationTask? weeklyTask =
        await widget.planClient.fetchActiveExecutionPlanGenerationTask(
      token,
      record: record,
      durationMonths: _periodMonthCount,
      planType: 'weekly',
      targetMonthIndex: monthIndex,
      targetWeekIndex: _previewWeek,
    );
    if (weeklyTask != null &&
        !weeklyTask.isDone &&
        !weeklyTask.isFailed &&
        weeklyTask.taskId.trim().isNotEmpty) {
      return weeklyTask;
    }
    final IepExecutionPlanGenerationTask? monthlyTask =
        await widget.planClient.fetchActiveExecutionPlanGenerationTask(
      token,
      record: record,
      durationMonths: _periodMonthCount,
      planType: 'monthly',
      targetMonthIndex: monthIndex,
    );
    if (monthlyTask != null &&
        !monthlyTask.isDone &&
        !monthlyTask.isFailed &&
        monthlyTask.taskId.trim().isNotEmpty) {
      return monthlyTask;
    }
    return null;
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

  List<String> _normalizedGoalLines(List<String> values) {
    return _normalizedNumberedTextLines(values);
  }

  IepPlan _planPayloadForSave() {
    final IepPlan? basePlan = _savedPlan?.plan;
    final IepPlanMeta meta = basePlan?.meta ??
        IepPlanMeta(
          planDate: _formatDateDash(DateTime.now()),
          participant: widget.record?.examinerName.trim() ?? '',
          implementer: widget.record?.examinerName.trim() ?? '',
          startDate: _formatDateDash(_periodStart),
          endDate: _formatDateDash(_periodEnd),
        );
    final IepPlanStudent student = basePlan?.student ??
        IepPlanStudent(
          name: widget.record?.studentName.trim() ?? '',
          gender: widget.record?.studentGender.trim() ?? '',
          birthDate: widget.record?.birthDate.trim() ?? '',
        );
    final String title = basePlan?.title.trim().isNotEmpty == true
        ? basePlan!.title.trim()
        : '康复教学计划';
    final List<IepPlanRow> rows = <IepPlanRow>[];
    for (final _DocDomainData domain in _totalPlanDomains) {
      final String domainName = domain.domain.trim();
      final List<String> longGoals = _normalizedGoalLines(domain.longGoals);
      final String longGoalText = longGoals.join('\n');
      for (final _DocShortGoalData shortGoal in domain.shortGoals) {
        final String goal = shortGoal.goal.trim();
        if (goal.isEmpty) {
          continue;
        }
        rows.add(
          IepPlanRow(
            domain: domainName,
            longGoal: longGoalText,
            shortGoal: goal,
            courseForm: shortGoal.lesson.trim(),
            startEndDate: shortGoal.period.trim(),
          ),
        );
      }
    }
    return IepPlan(
      title: title,
      student: student,
      meta: IepPlanMeta(
        planDate: meta.planDate.trim().isNotEmpty
            ? meta.planDate
            : _formatDateDash(DateTime.now()),
        participant: meta.participant,
        implementer: meta.implementer,
        startDate: _formatDateDash(_periodStart),
        endDate: _formatDateDash(_periodEnd),
      ),
      rows: rows,
    );
  }

  Future<void> requestConfirmIepPlan() async {
    if (!mounted || _confirmingPlan || _generatingPlan || _loadingPlan) {
      return;
    }
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    final IepPlan plan = _planPayloadForSave();
    if (plan.rows.isEmpty) {
      _showMessage('请先生成或填写IEP总计划');
      return;
    }
    setState(() {
      _confirmingPlan = true;
      _planError = '';
    });
    widget.onConfirmAvailabilityChanged(false);
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanSaved savedPlan = await widget.planClient.saveIepPlan(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        status: 'confirmed',
        plan: plan,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _confirmingPlan = false;
        _savedPlan = savedPlan;
        _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
        _applyPeriodFromPlan(savedPlan.plan, record);
        _totalPlanDomains = savedPlan.plan == null
            ? <_DocDomainData>[]
            : _docDomainsFromPlan(savedPlan.plan!);
        _syncPreviewMonthToPeriod();
      });
      _notifyRecordStatus(record, savedPlan.status);
      _syncConfirmAvailability(savedPlan);
      _showMessage('IEP已确认生成', tone: PadMessageTone.success);
    } on IepPlanApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _confirmingPlan = false;
        _planError = error.message;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      final String message = '确认IEP失败：$error';
      setState(() {
        _confirmingPlan = false;
        _planError = message;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
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
        _generationProgress = math.max(
          _generationProgress,
          _streamGenerationProgress(_aiStreamText),
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
      previewMode: _activeGenerationPreviewMode,
      monthLabel: _activeGenerationMonthLabel,
      weekIndex: _activeGenerationWeekIndex,
      restWeekdays: List<int>.from(_activeWeeklyRestWeekdays),
      status: _generationStatus,
      streamText: _aiStreamText,
      progress: _generationProgress,
      costAmountCny: _generationCostAmountCny,
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
    final _IepGenerationSessionSnapshot? session =
        _generationSessionFor(record);
    if (session == null) {
      return false;
    }
    setState(() {
      _periodMonthCount = session.durationMonths == 6 ? 6 : 3;
      _periodEndOverride = null;
      _activeGenerationTaskId = session.taskId;
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = session.durationMonths;
      _activeGenerationPreviewMode = session.previewMode;
      _activeGenerationMonthLabel = session.monthLabel;
      _activeGenerationWeekIndex = session.weekIndex;
      _activeWeeklyRestWeekdays = List<int>.from(session.restWeekdays);
      _previewMode = session.previewMode;
      if (session.monthLabel.trim().isNotEmpty) {
        _previewMonth = session.monthLabel;
      }
      if (session.weekIndex > 0) {
        _previewWeek = session.weekIndex;
      }
      _generationStatus = session.status;
      _aiStreamText = session.streamText;
      _generationProgress = session.progress;
      _generationCostAmountCny = session.costAmountCny;
      _generatingPlan = true;
      _planError = '';
      _hasCompletedInitialPlanLoad = true;
      _loadingPlan = false;
    });
    _notifyRecordStatus(record, 'generating');
    Future<void>.microtask(() {
      if (!mounted) {
        return;
      }
      switch (session.previewMode) {
        case _IepPreviewMode.total:
          _resumeIepPlanGenerationTask(showMessage: false);
        case _IepPreviewMode.month:
        case _IepPreviewMode.week:
          _resumeExecutionPlanGenerationTask(showMessage: false);
      }
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
      switch (_activeGenerationPreviewMode) {
        case _IepPreviewMode.total:
          await _resumeIepPlanGenerationTask(showMessage: showMessage);
        case _IepPreviewMode.month:
        case _IepPreviewMode.week:
          await _resumeExecutionPlanGenerationTask(showMessage: showMessage);
      }
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
      switch (_activeGenerationPreviewMode) {
        case _IepPreviewMode.total:
          await _resumeIepPlanGenerationTask(showMessage: false);
        case _IepPreviewMode.month:
        case _IepPreviewMode.week:
          await _resumeExecutionPlanGenerationTask(showMessage: false);
      }
      return;
    }
    if (_restoreGenerationSessionFor(record)) {
      return;
    }
    if (_previewMode == _IepPreviewMode.week &&
        !_currentWeekCanGenerateDirectly()) {
      await _showWeeklyPlanMissingMonthConfirmDialog();
      return;
    }
    switch (_previewMode) {
      case _IepPreviewMode.total:
        await _generateIepPlan();
      case _IepPreviewMode.month:
        await _generateMonthlyPlan();
      case _IepPreviewMode.week:
        await _generateWeeklyPlan();
    }
  }

  Future<void> _handleRetryRequest() async {
    if (_hasResumableGenerationTask ||
        _generationSessionFor(widget.record) != null) {
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
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
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
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
        });
      case IepPlanGenerationEventType.done:
        final IepPlan? plan = event.plan;
        if (plan == null) {
          throw const IepPlanApiException('AI生成未返回计划数据');
        }
        setState(() {
          _generationProgress = math.max(_generationProgress, .99);
          _generationStatus = '生成完成，正在自动保存草稿';
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
        });
        return _IepGenerationResult(
          plan: plan,
          savedPlan: event.savedPlan,
          costAmountCny: event.costAmountCny,
        );
      case IepPlanGenerationEventType.error:
        throw IepPlanApiException(
          event.message.trim().isEmpty ? 'AI生成失败' : event.message.trim(),
        );
    }
    return null;
  }

  Future<_ExecutionPlanGenerationResult<T>?> _handleExecutionGenerationEvent<T>(
    IepExecutionPlanGenerationEvent<T> event, {
    required int ticket,
    required String statusLabel,
  }) async {
    switch (event.type) {
      case IepExecutionPlanGenerationEventType.status:
        setState(() {
          _generationStatus =
              event.message.trim().isEmpty ? statusLabel : event.message.trim();
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
        });
      case IepExecutionPlanGenerationEventType.delta:
        final String delta = _generationDeltaAfterExistingText(event.text);
        if (delta.isEmpty) {
          break;
        }
        await _appendDeltaWithTypewriter(delta, ticket: ticket);
        if (!mounted || ticket != _generationTicket) {
          return null;
        }
        setState(() {
          _generationStatus = statusLabel;
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
        });
      case IepExecutionPlanGenerationEventType.done:
        final T? plan = event.data;
        if (plan == null) {
          throw const IepPlanApiException('AI生成未返回计划数据');
        }
        setState(() {
          _generationProgress = math.max(_generationProgress, .99);
          _generationStatus = '生成完成，正在自动保存草稿';
          _generationCostAmountCny = math.max(
            _generationCostAmountCny,
            event.costAmountCny,
          );
        });
        return _ExecutionPlanGenerationResult<T>(
          plan,
          event.costAmountCny,
          savedExecutionPlans: switch (plan) {
            final IepExecutionPlansSaved saved => saved,
            _ => null,
          },
        );
      case IepExecutionPlanGenerationEventType.error:
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
    double actualCostAmountCny = 0;
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _loadingPlan = false;
      _generatingPlan = true;
      _generationStatus = forceRegenerate ? '正在重新生成IEP计划' : '正在准备AI生成';
      _aiStreamText = '';
      _generationProgress = .08;
      _generationCostAmountCny = 0;
      _planError = '';
      _executionPlans = IepExecutionPlansSaved.empty(_periodMonthCount);
      _monthlyRestWeekdayOverrides.clear();
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = _periodMonthCount;
      if (forceRegenerate) {
        _savedPlan = null;
        _totalPlanDomains = <_DocDomainData>[];
      }
    });
    _notifyRecordStatus(record, 'generating');
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
        if (result != null) {
          actualCostAmountCny = result.costAmountCny;
        }
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
      await _showGenerationCostDialog(
        planLabel: 'IEP计划',
        costAmountCny: actualCostAmountCny,
      );
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
      _notifyRecordStatus(record, _savedPlan?.status);
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
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
  }

  Future<void> _generateMonthlyPlan({bool forceRegenerate = false}) async {
    final IepAssessmentRecordSummary? record = widget.record;
    final IepPlan? sourcePlan = _savedPlan?.plan;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (sourcePlan == null || !_savedPlan!.hasContent) {
      _showMessage('请先生成IEP总计划');
      return;
    }
    if (_generatingPlan) {
      return;
    }
    final List<int>? restWeekdays = await _showWeeklyRestWeekdaysDialog(
      initialSelectedWeekdays: _currentMonthlyRestWeekdays(),
      title: '选择$_previewMonth休息日',
      helperText: '请先选择机构休息日。系统会先据此计算当前月份周数，再让月计划每一条训练内容的日期区间与后续周计划一一对应。',
      confirmLabel: '确认并生成',
    );
    if (restWeekdays == null || !mounted) {
      return;
    }
    final int ticket = ++_generationTicket;
    ++_loadTicket;
    final int monthIndex = _previewMonthIndex();
    IepMonthlyPlan? finalPlan;
    IepExecutionPlansSaved? savedPlansFromTask;
    setState(() {
      _activeWeeklyRestWeekdays = List<int>.from(restWeekdays);
      _monthlyRestWeekdayOverrides[monthIndex] = List<int>.from(restWeekdays);
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
      final int maxWeek = _lastAvailableWeekInMonthRange(
        monthRange,
        restWeekdays: restWeekdays,
      );
      if (_previewWeek > maxWeek) {
        _previewWeek = 1;
      }
      _loadingPlan = false;
      _generatingPlan = true;
      _generationStatus = forceRegenerate ? '正在重新生成月计划' : '正在准备月计划';
      _aiStreamText = '';
      _generationProgress = .08;
      _generationCostAmountCny = 0;
      _planError = '';
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = _periodMonthCount;
      _activeGenerationPreviewMode = _IepPreviewMode.month;
      _activeGenerationMonthLabel = _previewMonth;
      _activeGenerationWeekIndex = 0;
    });
    _notifyRecordStatus(record, 'generating');
    widget.onConfirmAvailabilityChanged(false);

    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepExecutionPlanGenerationTask task =
          await widget.planClient.createExecutionPlanGenerationTask(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        planType: 'monthly',
        targetMonthIndex: monthIndex,
        sourcePlan: sourcePlan,
        restWeekdays: restWeekdays,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      setState(() {
        _activeGenerationTaskId = task.taskId;
      });
      await for (final IepExecutionPlanGenerationEvent<dynamic> rawEvent
          in widget.planClient.watchExecutionPlanGenerationTask(
        token,
        record: record,
        taskId: task.taskId,
      )) {
        if (!mounted || ticket != _generationTicket) {
          return;
        }
        final IepExecutionPlanGenerationEvent<IepMonthlyPlan> event =
            _castExecutionEvent<IepMonthlyPlan>(rawEvent);
        final _ExecutionPlanGenerationResult<IepMonthlyPlan>? result =
            await _handleExecutionGenerationEvent<IepMonthlyPlan>(
          event,
          ticket: ticket,
          statusLabel: 'AI正在生成月计划',
        );
        finalPlan = result?.plan ?? finalPlan;
        savedPlansFromTask = result?.savedExecutionPlans ?? savedPlansFromTask;
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      if (finalPlan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      savedPlansFromTask ??= await widget.planClient.saveMonthlyPlan(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        targetMonthIndex: monthIndex,
        plan: finalPlan,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final double actualCostAmountCny = _generationCostAmountCny;
      final IepExecutionPlansSaved saved = savedPlansFromTask;
      _syncMonthlyRestWeekdayOverridesFromExecutionPlans(saved);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '';
        _aiStreamText = '';
        _generationProgress = 1;
        _activeGenerationTaskId = '';
        _activeGenerationRecordKey = '';
        _executionPlans = saved;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      await _showGenerationCostDialog(
        planLabel: '月计划',
        costAmountCny: actualCostAmountCny,
      );
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
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final String message = '月计划生成失败：$error';
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成失败';
        _planError = _savedPlan?.hasContent == true ? '' : message;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
  }

  Future<void> _generateWeeklyPlan({bool forceRegenerate = false}) async {
    final IepAssessmentRecordSummary? record = widget.record;
    final IepPlan? sourcePlan = _savedPlan?.plan;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (sourcePlan == null || !_savedPlan!.hasContent) {
      _showMessage('请先生成IEP总计划');
      return;
    }
    if (_generatingPlan) {
      return;
    }
    List<int>? restWeekdays;
    final int monthIndex = _previewMonthIndex();
    final IepMonthlyPlan? monthlyPlan = _executionPlans?.monthPlan(monthIndex);
    if (monthlyPlan != null && monthlyPlan.restWeekdays.isNotEmpty) {
      restWeekdays = _normalizeWeeklyRestWeekdays(monthlyPlan.restWeekdays);
    } else {
      restWeekdays = await _showWeeklyRestWeekdaysDialog(
        initialSelectedWeekdays: _currentWeeklyRestWeekdays(),
        title: '选择$_previewMonth第$_previewWeek周休息日',
        helperText: '当前没有可复用的月计划休息日，请先选择机构休息日。系统会自动从周计划完成情况日期里过滤这些星期。',
        confirmLabel: '确认并生成',
      );
      if (restWeekdays == null || !mounted) {
        return;
      }
    }
    final List<int> effectiveRestWeekdays = List<int>.from(restWeekdays);
    final int ticket = ++_generationTicket;
    ++_loadTicket;
    IepWeeklyPlan? finalPlan;
    IepExecutionPlansSaved? savedPlansFromTask;
    setState(() {
      _activeWeeklyRestWeekdays = List<int>.from(effectiveRestWeekdays);
      _monthlyRestWeekdayOverrides[monthIndex] =
          List<int>.from(effectiveRestWeekdays);
      _loadingPlan = false;
      _generatingPlan = true;
      _generationStatus = forceRegenerate ? '正在重新生成周计划' : '正在准备周计划';
      _aiStreamText = '';
      _generationProgress = .08;
      _generationCostAmountCny = 0;
      _planError = '';
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = _recordGenerationKey(record);
      _activeGenerationDurationMonths = _periodMonthCount;
      _activeGenerationPreviewMode = _IepPreviewMode.week;
      _activeGenerationMonthLabel = _previewMonth;
      _activeGenerationWeekIndex = _previewWeek;
    });
    _notifyRecordStatus(record, 'generating');
    widget.onConfirmAvailabilityChanged(false);

    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepExecutionPlanGenerationTask task =
          await widget.planClient.createExecutionPlanGenerationTask(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        planType: 'weekly',
        targetMonthIndex: monthIndex,
        targetWeekIndex: _previewWeek,
        sourcePlan: sourcePlan,
        monthlyPlan: monthlyPlan,
        restWeekdays: effectiveRestWeekdays,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      setState(() {
        _activeGenerationTaskId = task.taskId;
      });
      await for (final IepExecutionPlanGenerationEvent<dynamic> rawEvent
          in widget.planClient.watchExecutionPlanGenerationTask(
        token,
        record: record,
        taskId: task.taskId,
      )) {
        if (!mounted || ticket != _generationTicket) {
          return;
        }
        final IepExecutionPlanGenerationEvent<IepWeeklyPlan> event =
            _castExecutionEvent<IepWeeklyPlan>(rawEvent);
        final _ExecutionPlanGenerationResult<IepWeeklyPlan>? result =
            await _handleExecutionGenerationEvent<IepWeeklyPlan>(
          event,
          ticket: ticket,
          statusLabel: 'AI正在生成周计划',
        );
        finalPlan = result?.plan ?? finalPlan;
        savedPlansFromTask = result?.savedExecutionPlans ?? savedPlansFromTask;
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      if (finalPlan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      final IepExecutionPlansSaved saved =
          savedPlansFromTask ??
              await widget.planClient.saveWeeklyPlan(
                token,
                record: record,
                durationMonths: _periodMonthCount,
                targetMonthIndex: monthIndex,
                targetWeekIndex: _previewWeek,
                plan: finalPlan,
              );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final double actualCostAmountCny = _generationCostAmountCny;
      _syncMonthlyRestWeekdayOverridesFromExecutionPlans(saved);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '';
        _aiStreamText = '';
        _generationProgress = 1;
        _activeGenerationTaskId = '';
        _activeGenerationRecordKey = '';
        _executionPlans = saved;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      await _showGenerationCostDialog(
        planLabel: '周计划',
        costAmountCny: actualCostAmountCny,
      );
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
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final String message = '周计划生成失败：$error';
      _storeCurrentGenerationSession(record);
      setState(() {
        _generatingPlan = false;
        _generationStatus = '生成失败';
        _planError = _savedPlan?.hasContent == true ? '' : message;
      });
      _notifyRecordStatus(record, _savedPlan?.status);
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
    double actualCostAmountCny = 0;
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
    _notifyRecordStatus(record, 'generating');
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
          if (result != null) {
            actualCostAmountCny = result.costAmountCny;
          }
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
      await _showGenerationCostDialog(
        planLabel: 'IEP计划',
        costAmountCny: actualCostAmountCny > 0
            ? actualCostAmountCny
            : _generationCostAmountCny,
      );
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
      _notifyRecordStatus(record, _savedPlan?.status);
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
      _notifyRecordStatus(record, _savedPlan?.status);
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
        _generationProgress = math.max(
          _generationProgress,
          _streamGenerationProgress(_aiStreamText),
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
        _generationProgress = math.max(_generationProgress, .99);
        _generationStatus =
            task.message.trim().isEmpty ? '生成完成，正在自动保存草稿' : task.message.trim();
        _generationCostAmountCny = math.max(
          _generationCostAmountCny,
          task.costAmountCny,
        );
      });
      return _IepGenerationResult(
        plan: plan,
        savedPlan: task.savedPlan,
        costAmountCny: task.costAmountCny,
      );
    }
    setState(() {
      _generationStatus =
          task.message.trim().isEmpty ? 'AI正在生成IEP计划' : task.message.trim();
      _generationCostAmountCny = math.max(
        _generationCostAmountCny,
        task.costAmountCny,
      );
    });
    return null;
  }

  Future<_ExecutionPlanGenerationResult<dynamic>?>
      _applyExecutionGenerationTaskSnapshot(
    IepExecutionPlanGenerationTask task, {
    required int ticket,
  }) async {
    if (task.streamText.isNotEmpty && task.streamText != _aiStreamText) {
      setState(() {
        _aiStreamText = task.streamText;
        _generationProgress = math.max(
          _generationProgress,
          _streamGenerationProgress(_aiStreamText),
        );
      });
    }
    if (task.isFailed) {
      throw IepPlanApiException(task.error.isEmpty ? 'AI生成失败' : task.error);
    }
    final String statusLabel = task.planType == 'weekly' ? 'AI正在生成周计划' : 'AI正在生成月计划';
    if (task.isDone) {
      final dynamic plan = task.savedPlan ?? task.plan;
      if (plan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      setState(() {
        _generationProgress = math.max(_generationProgress, .99);
        _generationStatus =
            task.message.trim().isEmpty ? '生成完成，正在自动保存草稿' : task.message.trim();
        _generationCostAmountCny = math.max(
          _generationCostAmountCny,
          task.costAmountCny,
        );
      });
      return _ExecutionPlanGenerationResult<dynamic>(
        plan,
        task.costAmountCny,
        savedExecutionPlans: task.savedExecutionPlans,
      );
    }
    setState(() {
      _generationStatus =
          task.message.trim().isEmpty ? statusLabel : task.message.trim();
      _generationCostAmountCny = math.max(
        _generationCostAmountCny,
        task.costAmountCny,
      );
    });
    return null;
  }

  Future<void> _resumeExecutionPlanGenerationTask({
    bool showMessage = true,
  }) async {
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
    dynamic finalPlan;
    IepExecutionPlansSaved? savedPlans;
    double actualCostAmountCny = 0;
    final bool isWeekly = _activeGenerationPreviewMode == _IepPreviewMode.week;
    setState(() {
      _loadingPlan = false;
      _generatingPlan = true;
      _planError = '';
      _generationStatus = '正在重新连接AI生成任务';
      if (_generationProgress < .12) {
        _generationProgress = .12;
      }
    });
    _notifyRecordStatus(record, 'generating');
    widget.onConfirmAvailabilityChanged(false);
    if (showMessage) {
      _showMessage('正在重新连接AI生成任务');
    }
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepExecutionPlanGenerationTask latest =
          await widget.planClient.fetchExecutionPlanGenerationTask(
        token,
        record: record,
        taskId: taskId,
      );
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      final _ExecutionPlanGenerationResult<dynamic>? latestResult =
          await _applyExecutionGenerationTaskSnapshot(latest, ticket: ticket);
      finalPlan = latestResult?.plan ?? finalPlan;
      savedPlans = latestResult?.savedExecutionPlans ?? savedPlans;
      if (latestResult == null && !latest.isDone) {
        await for (final IepExecutionPlanGenerationEvent<dynamic> rawEvent
            in widget.planClient.watchExecutionPlanGenerationTask(
          token,
          record: record,
          taskId: taskId,
        )) {
          if (!mounted || ticket != _generationTicket) {
            return;
          }
          final _ExecutionPlanGenerationResult<dynamic>? result =
              await _handleExecutionGenerationEvent<dynamic>(
            rawEvent,
            ticket: ticket,
            statusLabel: isWeekly ? 'AI正在生成周计划' : 'AI正在生成月计划',
          );
          finalPlan = result?.plan ?? finalPlan;
          savedPlans = result?.savedExecutionPlans ?? savedPlans;
          if (result != null) {
            actualCostAmountCny = result.costAmountCny;
          }
        }
      }
      if (!mounted || ticket != _generationTicket) {
        return;
      }
      if (finalPlan == null) {
        throw const IepPlanApiException('AI生成未返回计划数据');
      }
      if (savedPlans == null) {
        throw const IepPlanApiException('AI生成已完成，但未返回已保存的计划数据');
      }
      _applyGeneratedExecutionPlans(record: record, saved: savedPlans);
      await _showGenerationCostDialog(
        planLabel: isWeekly ? '周计划' : '月计划',
        costAmountCny:
            actualCostAmountCny > 0 ? actualCostAmountCny : _generationCostAmountCny,
      );
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
      _notifyRecordStatus(record, _savedPlan?.status);
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
      _notifyRecordStatus(record, _savedPlan?.status);
      _syncConfirmAvailability(_savedPlan);
      _showMessage(message);
    }
  }

  void _applyGeneratedExecutionPlans({
    required IepAssessmentRecordSummary record,
    required IepExecutionPlansSaved saved,
  }) {
    _generationSessionsByRecord.remove(_recordGenerationKey(record));
    _syncMonthlyRestWeekdayOverridesFromExecutionPlans(saved);
    setState(() {
      _generatingPlan = false;
      _generationStatus = '';
      _aiStreamText = '';
      _generationProgress = 1;
      _activeGenerationTaskId = '';
      _activeGenerationRecordKey = '';
      _activeGenerationPreviewMode = _IepPreviewMode.total;
      _activeGenerationMonthLabel = '';
      _activeGenerationWeekIndex = 0;
      _executionPlans = saved;
    });
    _notifyRecordStatus(record, _savedPlan?.status);
    _syncConfirmAvailability(_savedPlan);
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
      _activeGenerationPreviewMode = _IepPreviewMode.total;
      _activeGenerationMonthLabel = '';
      _activeGenerationWeekIndex = 0;
      _savedPlan = savedPlan;
      _executionPlans = IepExecutionPlansSaved.empty(
        savedPlan.durationMonths == 6 ? 6 : 3,
      );
      _monthlyRestWeekdayOverrides.clear();
      _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
      _applyPeriodFromPlan(savedPlan.plan, record);
      _totalPlanDomains = savedPlan.plan == null
          ? <_DocDomainData>[]
          : _docDomainsFromPlan(savedPlan.plan!);
      _syncPreviewMonthToPeriod();
    });
    _notifyRecordStatus(record, savedPlan.status);
    _syncConfirmAvailability(savedPlan);
  }

  void _notifyRecordStatus(
    IepAssessmentRecordSummary record,
    String? savedStatus,
  ) {
    final String normalized = savedStatus?.trim() ?? '';
    if (_generatingPlan || _activeGenerationTaskId.trim().isNotEmpty) {
      widget.onRecordStatusChanged(record, 'generating');
      return;
    }
    widget.onRecordStatusChanged(record, normalized);
  }

  double _streamGenerationProgress(String text) {
    final String trimmed = text.trim();
    if (trimmed.isEmpty) {
      return .12;
    }
    final double length = trimmed.runes.length.toDouble();
    const double floor = .12;
    const double ceiling = .975;

    if (length <= 520) {
      return _lerpProgress(floor, .26, length / 520);
    }
    if (length <= 2000) {
      return _lerpProgress(.26, .88, (length - 520) / 1480);
    }
    final double tail = 1 - math.exp(-(length - 2000) / 900);
    return _lerpProgress(.88, ceiling, tail.clamp(0, 1));
  }

  double _lerpProgress(double start, double end, double t) {
    final double normalized = t.clamp(0, 1).toDouble();
    return start + (end - start) * normalized;
  }

  String _recordGenerationKey(IepAssessmentRecordSummary record) {
    return '${record.source.trim().toUpperCase()}-${record.id}';
  }

  List<int> _normalizeWeeklyRestWeekdays(List<int> value) {
    final List<int> normalized = <int>[];
    for (final int weekday in value) {
      if (weekday < DateTime.monday || weekday > DateTime.sunday) {
        continue;
      }
      if (normalized.contains(weekday)) {
        continue;
      }
      normalized.add(weekday);
      if (normalized.length >= 2) {
        break;
      }
    }
    return normalized.isEmpty ? <int>[DateTime.sunday] : normalized;
  }

  List<int>? _weeklyPlanRestWeekdaysForMonth(int monthIndex) {
    final IepExecutionPlansSaved? executionPlans = _executionPlans;
    if (executionPlans == null) {
      return null;
    }
    for (final IepWeeklyPlanSaved item in executionPlans.weeklyPlans) {
      if (item.targetMonthIndex != monthIndex ||
          item.plan.restWeekdays.isEmpty) {
        continue;
      }
      return _normalizeWeeklyRestWeekdays(item.plan.restWeekdays);
    }
    return null;
  }

  List<int> _restWeekdaysForMonth(int monthIndex) {
    final List<int>? override = _monthlyRestWeekdayOverrides[monthIndex];
    if (override != null && override.isNotEmpty) {
      return _normalizeWeeklyRestWeekdays(override);
    }
    final List<int>? weeklyRestWeekdays =
        _weeklyPlanRestWeekdaysForMonth(monthIndex);
    if (weeklyRestWeekdays != null && weeklyRestWeekdays.isNotEmpty) {
      return weeklyRestWeekdays;
    }
    final IepMonthlyPlan? monthPlan = _executionPlans?.monthPlan(monthIndex);
    if (monthPlan != null && monthPlan.restWeekdays.isNotEmpty) {
      return _normalizeWeeklyRestWeekdays(monthPlan.restWeekdays);
    }
    return _normalizeWeeklyRestWeekdays(_activeWeeklyRestWeekdays);
  }

  List<int> _currentMonthlyRestWeekdays() {
    return _restWeekdaysForMonth(_previewMonthIndex());
  }

  List<int> _currentWeeklyRestWeekdays() {
    final IepWeeklyPlan? plan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    if (plan != null && plan.restWeekdays.isNotEmpty) {
      return _normalizeWeeklyRestWeekdays(plan.restWeekdays);
    }
    return _currentMonthlyRestWeekdays();
  }

  void _syncMonthlyRestWeekdayOverridesFromExecutionPlans(
    IepExecutionPlansSaved? executionPlans,
  ) {
    _monthlyRestWeekdayOverrides.clear();
    if (executionPlans == null) {
      return;
    }
    for (final IepMonthlyPlanSaved item in executionPlans.monthlyPlans) {
      if (item.plan.restWeekdays.isEmpty) {
        continue;
      }
      _monthlyRestWeekdayOverrides[item.targetMonthIndex] =
          _normalizeWeeklyRestWeekdays(item.plan.restWeekdays);
    }
    for (final IepWeeklyPlanSaved item in executionPlans.weeklyPlans) {
      if (_monthlyRestWeekdayOverrides.containsKey(item.targetMonthIndex) ||
          item.plan.restWeekdays.isEmpty) {
        continue;
      }
      _monthlyRestWeekdayOverrides[item.targetMonthIndex] =
          _normalizeWeeklyRestWeekdays(item.plan.restWeekdays);
    }
  }

  void _initPeriodFromRecord(IepAssessmentRecordSummary? record) {
    final DateTime? assessmentDate =
        DateTime.tryParse(record?.assessmentDate.trim() ?? '');
    if (assessmentDate == null) {
      return;
    }
    _periodStart = _dateOnly(assessmentDate);
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
    if (_hasAnyExecutionRecord()) {
      _showMessage('该记录卡已记录，不支持编辑周期');
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
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final bool blocked = switch (_previewMode) {
      _IepPreviewMode.total => _totalPlanHasAnyWeeklyExecutionRecord(),
      _IepPreviewMode.month => _monthHasAnyWeeklyExecutionRecord(
          _previewMonthIndex(),
        ),
      _IepPreviewMode.week => _selectedWeekHasRecordedEntry(weekPlan),
    };
    if (blocked) {
      final String message = switch (_previewMode) {
        _IepPreviewMode.total => '已有月计划下的周计划被记录，不支持重新生成总计划',
        _IepPreviewMode.month => '当前月份下已有周计划被记录，不支持重新生成月计划',
        _IepPreviewMode.week => '该记录卡已记录，不支持重新生成',
      };
      _showMessage(message);
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
    switch (_previewMode) {
      case _IepPreviewMode.total:
        await _generateIepPlan(forceRegenerate: true);
      case _IepPreviewMode.month:
        await _generateMonthlyPlan(forceRegenerate: true);
      case _IepPreviewMode.week:
        if (!_currentWeekCanGenerateDirectly()) {
          await _showWeeklyPlanMissingMonthConfirmDialog(
            forceRegenerate: true,
          );
          return;
        }
        await _generateWeeklyPlan(forceRegenerate: true);
    }
  }

  bool _currentWeekCanGenerateDirectly() {
    if (_previewMode != _IepPreviewMode.week) {
      return true;
    }
    final int monthIndex = _previewMonthIndex();
    final IepMonthlyPlan? monthlyPlan = _executionPlans?.monthPlan(monthIndex);
    return monthlyPlan != null && monthlyPlan.rows.isNotEmpty;
  }

  Future<void> _showWeeklyPlanMissingMonthConfirmDialog({
    bool forceRegenerate = false,
  }) async {
    if (_generatingPlan) {
      return;
    }
    final bool? directWeekly = await showDialog<bool>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepWeeklyPlanMissingMonthConfirmDialog(
            monthLabel: _previewMonth,
            weekNumber: _previewWeek,
            planTitle: _savedPlan?.plan?.title.trim().isNotEmpty == true
                ? _savedPlan!.plan!.title.trim()
                : '当前IEP总计划',
          ),
        );
      },
    );
    if (!mounted) {
      return;
    }
    if (directWeekly == true) {
      await _generateWeeklyPlan(forceRegenerate: forceRegenerate);
      return;
    }
    if (directWeekly == false) {
      setState(() {
        _previewMode = _IepPreviewMode.month;
        _selectedGoal = null;
      });
      await _generateMonthlyPlan(forceRegenerate: forceRegenerate);
    }
  }

  Future<List<int>?> _showWeeklyRestWeekdaysDialog({
    List<int>? initialSelectedWeekdays,
    String title = '选择机构休息日',
    String helperText = '系统会自动从周计划完成情况日期里过滤这些星期，保留当前 6 列布局。',
    String confirmLabel = '开始生成',
  }) async {
    final List<int>? result = await showDialog<List<int>>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepWeeklyRestWeekdayDialog(
            initialSelectedWeekdays:
                initialSelectedWeekdays ?? _currentWeeklyRestWeekdays(),
            title: title,
            helperText: helperText,
            confirmLabel: confirmLabel,
          ),
        );
      },
    );
    if (result == null) {
      return null;
    }
    return _normalizeWeeklyRestWeekdays(result);
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
      _showMessage('计划周期已同步，月计划和周计划日期已重算', tone: PadMessageTone.success);
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
      _syncMonthlyRestWeekdayOverridesFromExecutionPlans(executionPlans);
      _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
      _applyPeriodFromPlan(savedPlan.plan, record);
      _totalPlanDomains = savedPlan.plan == null
          ? <_DocDomainData>[]
          : _docDomainsFromPlan(savedPlan.plan!);
      _syncPreviewMonthToPeriod();
      _ensurePreviewWeekInRange();
    });
    _notifyRecordStatus(record, savedPlan.status);
    _syncConfirmAvailability(savedPlan);
  }

  void _ensurePreviewWeekInRange() {
    if (_previewMode != _IepPreviewMode.week) {
      return;
    }
    final int monthIndex = _previewMonthIndex();
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
    if (_weekDatesInMonthRange(
      monthRange,
      _previewWeek,
      restWeekdays: _restWeekdaysForMonth(monthIndex),
    ).isEmpty) {
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

  Future<void> exportCurrentPlanWord() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    try {
      final IepWordFile file = await _exportWordFileForCurrentPreview(record);
      final bool saved = await _saveWordFile(file);
      if (saved) {
        _showMessage('导出成功', tone: PadMessageTone.success);
      }
    } on IepPlanApiException catch (error) {
      _showMessage(error.message);
    } on Object catch (error) {
      _showMessage('导出失败：$error');
    }
  }

  Future<void> printCurrentPlan() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    try {
      final Uint8List bytes = await _downloadCurrentPlanPrintPdf(record);
      await Printing.layoutPdf(
        name: _currentPrintFileName(record),
        onLayout: (_) async => bytes,
      );
    } on IepPlanApiException catch (error) {
      _showMessage(error.message);
    } on Object catch (error) {
      _showMessage('打印失败：$error');
    }
  }

  Future<IepWordFile> _exportWordFileForCurrentPreview(
    IepAssessmentRecordSummary record,
  ) async {
    final IepPlan? totalPlan = _savedPlan?.plan;
    final IepMonthlyPlan? monthPlan =
        _executionPlans?.monthPlan(_previewMonthIndex());
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    return switch (_previewMode) {
      _IepPreviewMode.total => totalPlan == null
          ? throw const IepPlanApiException('请先生成IEP总计划')
          : widget.planClient.downloadIepPlanWord(
              token,
              record: record,
              durationMonths: _periodMonthCount,
              plan: _planPayloadForSave(),
            ),
      _IepPreviewMode.month => monthPlan == null
          ? throw const IepPlanApiException('请先生成月计划')
          : widget.planClient.downloadMonthlyPlanWord(
              token,
              record: record,
              plan: monthPlan,
            ),
      _IepPreviewMode.week => weekPlan == null
          ? throw const IepPlanApiException('请先生成周计划')
          : widget.planClient.downloadWeeklyPlanWord(
              token,
              record: record,
              plan: weekPlan,
            ),
    };
  }

  Future<bool> _saveWordFile(IepWordFile file) async {
    return DownloadedFileSaver.save(file);
  }

  String _currentPrintFileName(IepAssessmentRecordSummary record) {
    final String studentName =
        record.studentName.trim().isEmpty ? '学员' : record.studentName.trim();
    final String suffix = switch (_previewMode) {
      _IepPreviewMode.total => 'IEP总计划',
      _IepPreviewMode.month => '${_previewMonth}月计划',
      _IepPreviewMode.week => '${_previewMonth}第$_previewWeek周计划',
    };
    return '$studentName-$suffix.pdf';
  }

  Future<Uint8List> _downloadCurrentPlanPrintPdf(
    IepAssessmentRecordSummary record,
  ) async {
    final IepPlan? totalPlan = _savedPlan?.plan;
    final IepMonthlyPlan? monthPlan =
        _executionPlans?.monthPlan(_previewMonthIndex());
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    return switch (_previewMode) {
      _IepPreviewMode.total => totalPlan == null
          ? throw const IepPlanApiException('请先生成IEP总计划')
          : widget.planClient.downloadIepPlanPdf(
              token,
              record: record,
              durationMonths: _periodMonthCount,
              plan: _planPayloadForSave(),
            ),
      _IepPreviewMode.month => monthPlan == null
          ? throw const IepPlanApiException('请先生成月计划')
          : widget.planClient.downloadMonthlyPlanPdf(
              token,
              record: record,
              plan: monthPlan,
            ),
      _IepPreviewMode.week => weekPlan == null
          ? throw const IepPlanApiException('请先生成周计划')
          : widget.planClient.downloadWeeklyPlanPdf(
              token,
              record: record,
              plan: weekPlan,
            ),
    };
  }

  // ignore: unused_element
  Future<Uint8List> _buildCurrentPlanPrintPdf(PdfPageFormat format) async {
    final IepPlan? totalPlan = _savedPlan?.plan;
    final int monthIndex = _previewMonthIndex();
    final IepMonthlyPlan? monthPlan = _executionPlans?.monthPlan(monthIndex);
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(monthIndex, _previewWeek);
    if (_previewMode == _IepPreviewMode.total && totalPlan == null) {
      throw const IepPlanApiException('请先生成IEP总计划');
    }
    if (_previewMode == _IepPreviewMode.month && monthPlan == null) {
      throw const IepPlanApiException('请先生成月计划');
    }
    if (_previewMode == _IepPreviewMode.week && weekPlan == null) {
      throw const IepPlanApiException('请先生成周计划');
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
    final List<int> restWeekdays = weekPlan?.restWeekdays.isNotEmpty == true
        ? _normalizeWeeklyRestWeekdays(weekPlan!.restWeekdays)
        : _restWeekdaysForMonth(monthIndex);
    final List<DateTime> weekDates =
        _dateListFromStrings(weekPlan?.weekDates) ??
            _weekDatesInMonthRange(
              monthRange,
              _previewWeek,
              restWeekdays: restWeekdays,
            );
    return _buildIepPlanPrintPdf(
      format: format,
      mode: _previewMode,
      totalPlan: totalPlan,
      totalDomains: _totalPlanDomains,
      monthPlan: monthPlan,
      weekPlan: weekPlan,
      periodText: _formatDotRange(_periodStart, _periodEnd),
      monthLabel: _previewMonth,
      weekNumber: _previewWeek,
      monthRange: monthRange,
      weekDates: weekDates,
    );
  }

  Future<void> _showGenerationCostDialog({
    required String planLabel,
    required double costAmountCny,
  }) async {
    if (!mounted) {
      return;
    }
    await showDialog<void>(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepGenerationCostDialog(
            planLabel: planLabel,
            costAmountCny: costAmountCny,
          ),
        );
      },
    );
  }

  IepExecutionPlanGenerationEvent<T> _castExecutionEvent<T>(
    IepExecutionPlanGenerationEvent<dynamic> event,
  ) {
    switch (event.type) {
      case IepExecutionPlanGenerationEventType.status:
        return IepExecutionPlanGenerationEvent<T>.statusWithCost(
          event.message,
          event.costAmountCny,
        );
      case IepExecutionPlanGenerationEventType.delta:
        return IepExecutionPlanGenerationEvent<T>.deltaWithCost(
          event.text,
          event.costAmountCny,
        );
      case IepExecutionPlanGenerationEventType.done:
        return IepExecutionPlanGenerationEvent<T>.done(
          event.data as T,
          costAmountCny: event.costAmountCny,
        );
      case IepExecutionPlanGenerationEventType.error:
        return IepExecutionPlanGenerationEvent<T>.error(event.message);
    }
  }

  void _showTotalPlan() {
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _lessonSessionState = null;
    });
  }

  void _showMonthPlan(String month) {
    setState(() {
      _previewMode = _IepPreviewMode.month;
      _previewMonth = month;
      _selectedGoal = null;
      _lessonSessionState = null;
    });
  }

  void _showWeekPlan(String month, int weekNumber) {
    final int monthIndex =
        _periodMonths.indexOf(month) < 0 ? 1 : _periodMonths.indexOf(month) + 1;
    final List<int> restWeekdays = _restWeekdaysForMonth(monthIndex);
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
    final int week = _weekDatesInMonthRange(
      monthRange,
      weekNumber,
      restWeekdays: restWeekdays,
    ).isEmpty
        ? _lastAvailableWeekInMonthRange(
            monthRange,
            restWeekdays: restWeekdays,
          )
        : weekNumber;
    setState(() {
      _previewMode = _IepPreviewMode.week;
      _previewMonth = month;
      _previewWeek = week;
      _selectedGoal = null;
    });
    unawaited(_refreshLessonSessionState());
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

  Future<void> _openLessonSession() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (_previewMode != _IepPreviewMode.week) {
      _showMessage('请先切换到周计划后再开始上课');
      return;
    }
    final IepPlan? totalPlan = _savedPlan?.plan;
    final int monthIndex = _previewMonthIndex();
    final IepMonthlyPlan? monthPlan = _executionPlans?.monthPlan(monthIndex);
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(monthIndex, _previewWeek);
    if (weekPlan == null) {
      _showMessage('请先生成当前周计划后再开始上课');
      return;
    }
    final List<DateTime> weekDates = _selectedWeekDates(weekPlan);
    if (weekDates.isEmpty) {
      _showMessage('当前周计划缺少可记录日期');
      return;
    }
    if (_selectedWeekIsFuture(weekPlan)) {
      final String weekRangeText = _weekRangeText(weekDates);
      _showMessage('当前周计划尚未到开始记录时间，当前选中：$weekRangeText');
      return;
    }
    final IepLessonSession? resumeSession = _currentLessonSession();
    final bool editOnly = _lessonEntryUsesEditMode(weekPlan);
    final int? selectedDateIndex = resumeSession != null
        ? _selectedDateIndexForLessonSession(resumeSession, weekDates)
        : await _resolveLessonEntryDateIndex(weekPlan, weekDates);
    if (!mounted || selectedDateIndex == null) {
      return;
    }
    final String lessonDate = _formatDateDash(weekDates[selectedDateIndex]);
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      if (token.trim().isEmpty) {
        throw const IepPlanApiException('请先登录后再开始上课');
      }
      IepLessonSession lessonSession = IepLessonSession(
        lessonDate: lessonDate,
        weekDateIndex: selectedDateIndex + 1,
      );
      if (!editOnly) {
        final IepLessonSessionWeekState lessonSessionState =
            await widget.planClient.startLessonSession(
          token,
          record: record,
          durationMonths: _periodMonthCount,
          targetMonthIndex: monthIndex,
          targetWeekIndex: _previewWeek,
          lessonDate: lessonDate,
        );
        if (!mounted) {
          return;
        }
        lessonSession = lessonSessionState.sessionForDate(lessonDate) ??
            lessonSessionState.currentSession ??
            IepLessonSession(
              lessonDate: lessonDate,
              weekDateIndex: selectedDateIndex + 1,
              status: 'in_progress',
            );
        setState(() {
          _lessonSessionState = lessonSessionState;
        });
      }
      final _IepLessonSessionDraft draft = _buildLessonSessionDraft(
        record: record,
        totalPlan: totalPlan,
        monthPlan: monthPlan,
        weekPlan: weekPlan,
        initialSelectedDateIndex: selectedDateIndex,
        lessonDate: lessonDate,
        lessonSession: lessonSession,
        entryMode:
            editOnly ? _IepLessonEntryMode.edit : _IepLessonEntryMode.session,
      );
      await Navigator.of(context).push<_IepLessonSessionExitResult>(
        MaterialPageRoute<_IepLessonSessionExitResult>(
          builder: (BuildContext routeContext) => Scaffold(
            body: _IepLessonFullscreenViewport(
              child: _IepLessonSessionPage(
                draft: draft,
                planClient: widget.planClient,
                onPlansSaved: (IepExecutionPlansSaved saved) {
                  if (!mounted) {
                    return;
                  }
                  _syncMonthlyRestWeekdayOverridesFromExecutionPlans(saved);
                  setState(() {
                    _executionPlans = saved;
                  });
                },
              ),
            ),
          ),
        ),
      );
      if (!mounted) {
        return;
      }
      await _refreshLessonSessionState();
    } on IepPlanApiException catch (error) {
      if (!mounted) {
        return;
      }
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      _showMessage('开始上课失败：$error');
    }
  }

  Future<int?> _resolveLessonEntryDateIndex(
    IepWeeklyPlan weekPlan,
    List<DateTime> weekDates,
  ) async {
    final int todayIndex = _selectedWeekTodayIndex(weekPlan);
    if (_selectedWeekIsPast(weekPlan)) {
      return _showLessonDateSelectionDialog(
        weekPlan: weekPlan,
        weekDates: weekDates,
        selectableIndexes: List<int>.generate(
          weekDates.length,
          (int index) => index,
          growable: false,
        ),
        title: '选择修/补日期',
        helperText: '当前周已过去，请选择需要补记或修改记录的日期。',
      );
    }
    if (todayIndex < 0) {
      return null;
    }
    if (todayIndex == 0) {
      return todayIndex;
    }
    return _showLessonDateSelectionDialog(
      weekPlan: weekPlan,
      weekDates: weekDates,
      selectableIndexes: List<int>.generate(
        todayIndex + 1,
        (int index) => index,
        growable: false,
      ),
      title: '选择记录日期',
      helperText: '可选择今天直接修改，也可选择本周此前日期进行补记或修改。',
      preferredIndex: todayIndex,
    );
  }

  Future<int?> _showLessonDateSelectionDialog({
    required IepWeeklyPlan weekPlan,
    required List<DateTime> weekDates,
    required List<int> selectableIndexes,
    required String title,
    required String helperText,
    int? preferredIndex,
  }) async {
    final List<_IepLessonDateOption> options =
        selectableIndexes.map((int index) {
      final DateTime date = weekDates[index];
      final bool recorded = _selectedWeekDateHasRecordedEntry(weekPlan, index);
      final int recordedTaskCount =
          _selectedWeekDateRecordedTaskCount(weekPlan, index);
      final bool today = _dateOnly(date) == _dateOnly(DateTime.now());
      final String stateLabel = recorded
          ? '已记录${recordedTaskCount > 0 ? ' $recordedTaskCount 项' : ''}'
          : (today ? '今日待记录' : '待修/补');
      return _IepLessonDateOption(
        index: index,
        label: _weekDateOptionLabel(date),
        stateLabel: stateLabel,
        recorded: recorded,
        isToday: today,
      );
    }).toList(growable: false);
    if (options.isEmpty) {
      return null;
    }
    final int preferred = preferredIndex ?? options.first.index;
    final int initialIndex = options.indexWhere(
      (_IepLessonDateOption item) => item.index == preferred,
    );
    return showDialog<int>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepLessonDateSelectionDialog(
            title: title,
            helperText: helperText,
            options: options,
            initialSelectedIndex: initialIndex < 0 ? 0 : initialIndex,
          ),
        );
      },
    );
  }

  _IepLessonSessionDraft _buildLessonSessionDraft({
    required IepAssessmentRecordSummary record,
    required IepPlan? totalPlan,
    required IepMonthlyPlan? monthPlan,
    required IepWeeklyPlan? weekPlan,
    int initialSelectedDateIndex = 0,
    required String lessonDate,
    required IepLessonSession lessonSession,
    required _IepLessonEntryMode entryMode,
  }) {
    final String studentName = totalPlan?.student.name.trim().isNotEmpty == true
        ? totalPlan!.student.name.trim()
        : record.studentName.trim();
    final String gender = totalPlan?.student.gender.trim().isNotEmpty == true
        ? totalPlan!.student.gender.trim()
        : record.studentGender.trim();
    final String teacherName = weekPlan?.teacherName.trim().isNotEmpty == true
        ? weekPlan!.teacherName.trim()
        : record.examinerName.trim();
    final String courseName = weekPlan?.courseName.trim().isNotEmpty == true
        ? weekPlan!.courseName.trim()
        : '康复教学';
    final String planTitle = switch (_previewMode) {
      _IepPreviewMode.total => totalPlan?.title.trim().isNotEmpty == true
          ? totalPlan!.title.trim()
          : '康复教学计划',
      _IepPreviewMode.month => monthPlan?.title.trim().isNotEmpty == true
          ? monthPlan!.title.trim()
          : '康复教学$_previewMonth计划',
      _IepPreviewMode.week => weekPlan?.title.trim().isNotEmpty == true
          ? weekPlan!.title.trim()
          : '康复教学周计划日记录卡$_previewMonth第$_previewWeek周',
    };
    final String periodLabel = _formatDotRange(_periodStart, _periodEnd);
    final String weekLabel = switch (_previewMode) {
      _IepPreviewMode.total => 'IEP总计划',
      _IepPreviewMode.month => '$_previewMonth 月计划',
      _IepPreviewMode.week => '$_previewMonth 第$_previewWeek周',
    };
    final int monthIndex = _previewMonthIndex();
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: _monthDateFromLabel(
        _periodStart,
        _periodMonthCount,
        _previewMonth,
      ),
    );
    final List<int> restWeekdays = weekPlan?.restWeekdays.isNotEmpty == true
        ? _normalizeWeeklyRestWeekdays(weekPlan!.restWeekdays)
        : _restWeekdaysForMonth(monthIndex);
    final List<DateTime> weekDates =
        _dateListFromStrings(weekPlan?.weekDates) ??
            _weekDatesInMonthRange(
              monthRange,
              _previewWeek,
              restWeekdays: restWeekdays,
            );
    final List<String> weekDateOptions = weekDates
        .map((DateTime item) => _weekDateOptionLabel(item))
        .toList(growable: false);
    final String trainingDateLabel = weekDates.isEmpty
        ? (weekPlan?.trainingDate.trim().isNotEmpty == true
            ? weekPlan!.trainingDate.trim()
            : _weekRangeText(weekDates))
        : weekDateOptions.first;
    final String ageLabel = _ageSummaryText(record);
    final String preparation = weekPlan?.preparation.trim().isNotEmpty == true
        ? weekPlan!.preparation.trim()
        : '训练材料、视觉提示卡、强化物、记录表';

    final List<String> goals = <String>{}
        .followedBy(_goalLinesFromWeekPlan(weekPlan))
        .followedBy(_goalLinesFromMonthPlan(monthPlan))
        .followedBy(_goalLinesFromTotalPlan(totalPlan))
        .where((String item) => item.trim().isNotEmpty)
        .toList();

    final List<_IepLessonTaskDraft> tasks = <_IepLessonTaskDraft>[];
    if (weekPlan != null) {
      for (int rowIndex = 0; rowIndex < weekPlan.rows.length; rowIndex++) {
        final IepWeeklyPlanRow row = weekPlan.rows[rowIndex];
        final String project = row.project.trim();
        final String content = row.content.trim();
        if (project.isEmpty && content.isEmpty) {
          continue;
        }
        tasks.add(
          _IepLessonTaskDraft(
            sourceRowIndex: rowIndex,
            title: project.isEmpty ? '训练项目' : project,
            subtitle: content.isEmpty ? '待补充训练内容' : content,
            domain: _domainFromWeeklyProject(project),
            goal: goals.isEmpty ? '' : goals.first,
            materials: preparation,
            steps: _defaultLessonSteps(content.isEmpty ? project : content),
            tips: _defaultLessonTips(project, content),
            completionCodes: row.completion,
          ),
        );
      }
    }
    final List<_IepLessonTaskDraft> effectiveTasks = tasks.isEmpty
        ? <_IepLessonTaskDraft>[
            _IepLessonTaskDraft(
              sourceRowIndex: -1,
              title: '待执行训练任务',
              subtitle: '当前周计划暂无可执行训练项，请先补充本周训练内容。',
              domain: '周计划',
              goal: goals.isEmpty ? '完成本周课堂训练记录' : goals.first,
              materials: preparation,
              steps: _defaultLessonSteps(
                '根据本周计划补充具体训练项目后再进入正式上课。',
              ),
              tips: _defaultLessonTips(
                '周计划任务待补充',
                '根据本周计划补充具体训练项目后再进入正式上课。',
              ),
              completionCodes: List<String>.filled(weekDateOptions.length, ''),
            ),
          ]
        : tasks;

    return _IepLessonSessionDraft(
      record: record,
      durationMonths: _periodMonthCount,
      targetMonthIndex: monthIndex,
      targetWeekIndex: _previewWeek,
      weeklyPlan: weekPlan ??
          IepWeeklyPlan(
            title: planTitle,
            student: totalPlan?.student ??
                IepPlanStudent(
                  name: studentName,
                  gender: gender,
                  birthDate: record.birthDate.trim(),
                ),
            teacherName: teacherName.isEmpty ? '未设置老师' : teacherName,
            courseName: courseName,
            trainingDate: trainingDateLabel,
            preparation: preparation,
            weekDates: weekDates.map(_formatDateDash).toList(growable: false),
            restWeekdays: const <int>[DateTime.sunday],
            rows: const <IepWeeklyPlanRow>[],
          ),
      studentName: studentName.isEmpty ? '未选择学员' : studentName,
      gender: gender,
      ageLabel: ageLabel,
      teacherName: teacherName.isEmpty ? '未设置老师' : teacherName,
      courseName: courseName,
      planTitle: planTitle,
      stageLabel: '课堂执行中',
      periodLabel: periodLabel,
      weekLabel: weekLabel,
      initialSelectedDateIndex: initialSelectedDateIndex,
      lessonDate: lessonDate,
      lessonSession: lessonSession,
      entryMode: entryMode,
      trainingDateLabel: trainingDateLabel,
      weekDateOptions: weekDateOptions,
      completionColumnLabels: weekDateOptions,
      preparation: preparation,
      goals: goals.isEmpty ? <String>['提升课堂参与度与任务完成度'] : goals,
      tasks: effectiveTasks,
    );
  }

  List<String> _goalLinesFromWeekPlan(IepWeeklyPlan? plan) {
    if (plan == null) {
      return const <String>[];
    }
    return plan.rows
        .map((IepWeeklyPlanRow row) => row.content.trim())
        .where((String item) => item.isNotEmpty)
        .take(3)
        .toList();
  }

  List<String> _goalLinesFromMonthPlan(IepMonthlyPlan? plan) {
    if (plan == null) {
      return const <String>[];
    }
    return plan.rows
        .map((IepMonthlyPlanRow row) => row.shortGoal.trim())
        .where((String item) => item.isNotEmpty)
        .take(3)
        .toList();
  }

  List<String> _goalLinesFromTotalPlan(IepPlan? plan) {
    if (plan == null) {
      return const <String>[];
    }
    return plan.rows
        .map((IepPlanRow row) => row.shortGoal.trim())
        .where((String item) => item.isNotEmpty)
        .take(3)
        .toList();
  }

  String _ageSummaryText(IepAssessmentRecordSummary record) {
    final List<String> parts = <String>[];
    if (record.ageYears > 0) {
      parts.add('${record.ageYears}岁');
    }
    if (record.ageMonths > 0 && parts.length < 2) {
      parts.add('${record.ageMonths}个月');
    }
    if (parts.isEmpty && record.birthDate.trim().isNotEmpty) {
      return record.birthDate.trim();
    }
    return parts.isEmpty ? '年龄待补充' : parts.join();
  }

  String _domainFromWeeklyProject(String project) {
    if (project.contains('情绪') || project.contains('表达')) {
      return '情绪表达';
    }
    if (project.contains('社交')) {
      return '社交互动';
    }
    if (project.contains('动作') || project.contains('协调')) {
      return '动作协调';
    }
    if (project.contains('语言')) {
      return '语言表达';
    }
    return '认知理解';
  }

  List<String> _defaultLessonSteps(String content) {
    final String core = content.trim().isEmpty ? '完成当前训练任务' : content.trim();
    return <String>[
      '先用口头提示和示范建立本轮任务规则，确认学员理解 $core。',
      '进入正式训练，记录独立完成、辅助完成和错误反应的次数。',
      '完成后立即给予反馈和强化，并根据表现决定是否进入下一轮。',
    ];
  }

  List<String> _defaultLessonTips(String project, String content) {
    final String anchor =
        project.trim().isNotEmpty ? project.trim() : content.trim();
    return <String>[
      '先说短句指令，再给视觉提示，避免一次性输入过多信息。',
      '当学员完成 $anchor 时，立即给予明确强化和表扬。',
      '如果出现分心或抗拒，先降低难度，再逐步回到原任务。',
    ];
  }

  String _weekDateOptionLabel(DateTime value) {
    final List<String> weekdays = <String>[
      '周一',
      '周二',
      '周三',
      '周四',
      '周五',
      '周六',
      '周日',
    ];
    final int weekdayIndex = value.weekday.clamp(1, 7) - 1;
    return '${_weekDateLabel(value)} ${weekdays[weekdayIndex]}';
  }

  List<DateTime> _selectedWeekDates(IepWeeklyPlan? weekPlan) {
    if (weekPlan == null || _previewMode != _IepPreviewMode.week) {
      return const <DateTime>[];
    }
    final int monthIndex = _previewMonthIndex();
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: _monthDateFromLabel(
        _periodStart,
        _periodMonthCount,
        _previewMonth,
      ),
    );
    return _dateListFromStrings(weekPlan.weekDates) ??
        _weekDatesInMonthRange(
          monthRange,
          _previewWeek,
          restWeekdays: _restWeekdaysForMonth(monthIndex),
        );
  }

  int _selectedWeekTodayIndex(IepWeeklyPlan? weekPlan) {
    final List<DateTime> weekDates = _selectedWeekDates(weekPlan);
    if (weekDates.isEmpty) {
      return -1;
    }
    final DateTime today = _dateOnly(DateTime.now());
    return weekDates.indexWhere((DateTime item) => _dateOnly(item) == today);
  }

  bool _selectedWeekIsCurrent(IepWeeklyPlan? weekPlan) {
    return _selectedWeekTodayIndex(weekPlan) >= 0;
  }

  bool _selectedWeekIsPast(IepWeeklyPlan? weekPlan) {
    final List<DateTime> weekDates = _selectedWeekDates(weekPlan);
    if (weekDates.isEmpty) {
      return false;
    }
    final DateTime today = _dateOnly(DateTime.now());
    return _dateOnly(weekDates.last).isBefore(today);
  }

  bool _selectedWeekIsFuture(IepWeeklyPlan? weekPlan) {
    final List<DateTime> weekDates = _selectedWeekDates(weekPlan);
    if (weekDates.isEmpty) {
      return false;
    }
    final DateTime today = _dateOnly(DateTime.now());
    return _dateOnly(weekDates.first).isAfter(today);
  }

  Future<void> _refreshLessonSessionState() async {
    final IepAssessmentRecordSummary? record = widget.record;
    final int monthIndex = _previewMonthIndex();
    final int weekIndex = _previewWeek;
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(monthIndex, weekIndex);
    if (record == null ||
        _previewMode != _IepPreviewMode.week ||
        weekPlan == null) {
      if (_lessonSessionState != null && mounted) {
        setState(() {
          _lessonSessionState = null;
        });
      }
      return;
    }
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      if (token.trim().isEmpty) {
        return;
      }
      final IepLessonSessionWeekState state =
          await widget.planClient.fetchLessonSessionWeekState(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        targetMonthIndex: monthIndex,
        targetWeekIndex: weekIndex,
      );
      if (!mounted) {
        return;
      }
      final IepAssessmentRecordSummary? currentRecord = widget.record;
      if (currentRecord == null ||
          !_sameRecord(currentRecord, record) ||
          _previewMode != _IepPreviewMode.week ||
          _previewMonthIndex() != monthIndex ||
          _previewWeek != weekIndex) {
        return;
      }
      setState(() {
        _lessonSessionState = state;
      });
    } on Object {
      if (!mounted) {
        return;
      }
      final IepAssessmentRecordSummary? currentRecord = widget.record;
      if (currentRecord == null ||
          !_sameRecord(currentRecord, record) ||
          _previewMode != _IepPreviewMode.week ||
          _previewMonthIndex() != monthIndex ||
          _previewWeek != weekIndex) {
        return;
      }
      setState(() {
        _lessonSessionState = null;
      });
    }
  }

  IepLessonSession? _currentLessonSession() {
    final IepLessonSessionWeekState? state = _lessonSessionState;
    if (state == null) {
      return null;
    }
    final IepLessonSession? current = state.currentSession;
    if (current != null && current.isInProgress) {
      return current;
    }
    IepLessonSession? paused;
    for (final IepLessonSession item in state.sessions) {
      if (item.isInProgress) {
        return item;
      }
      if (!item.isPaused) {
        continue;
      }
      if (paused == null ||
          _lessonSessionSortKey(item).compareTo(_lessonSessionSortKey(paused)) >
              0) {
        paused = item;
      }
    }
    return paused;
  }

  String _lessonSessionSortKey(IepLessonSession session) {
    final List<String> candidates = <String>[
      session.updatedTime,
      session.lastHeartbeatAt,
      session.lastResumedAt,
      session.pausedAt,
      session.endedAt,
      session.startedAt,
      session.lessonDate,
    ];
    for (final String item in candidates) {
      final String normalized = item.trim();
      if (normalized.isNotEmpty) {
        return normalized;
      }
    }
    return '';
  }

  int _selectedDateIndexForLessonSession(
    IepLessonSession session,
    List<DateTime> weekDates,
  ) {
    if (session.weekDateIndex > 0 &&
        session.weekDateIndex <= weekDates.length) {
      return session.weekDateIndex - 1;
    }
    final int matchedIndex = weekDates.indexWhere(
      (DateTime item) => _formatDateDash(item) == session.lessonDate,
    );
    return matchedIndex < 0 ? 0 : matchedIndex;
  }

  bool _lessonEntryUsesEditMode(IepWeeklyPlan? weekPlan) {
    final String label = _lessonEntryButtonLabel(weekPlan);
    return label == '修改记录' || label == '修/补记录';
  }

  bool _canOpenLessonForSelectedWeek(IepWeeklyPlan? weekPlan) {
    return weekPlan != null &&
        _previewMode == _IepPreviewMode.week &&
        _selectedWeekDates(weekPlan).isNotEmpty &&
        !_selectedWeekIsFuture(weekPlan);
  }

  bool _selectedWeekDateHasRecordedEntry(
    IepWeeklyPlan? weekPlan,
    int dateIndex,
  ) {
    if (weekPlan == null || dateIndex < 0) {
      return false;
    }
    return weekPlan.rows.any((IepWeeklyPlanRow row) {
      if (dateIndex >= row.completion.length) {
        return false;
      }
      return row.completion[dateIndex].trim().isNotEmpty;
    });
  }

  int _selectedWeekDateRecordedTaskCount(
      IepWeeklyPlan? weekPlan, int dateIndex) {
    if (weekPlan == null || dateIndex < 0) {
      return 0;
    }
    int count = 0;
    for (final IepWeeklyPlanRow row in weekPlan.rows) {
      if (dateIndex < row.completion.length &&
          row.completion[dateIndex].trim().isNotEmpty) {
        count += 1;
      }
    }
    return count;
  }

  String _lessonEntryButtonLabel(IepWeeklyPlan? weekPlan) {
    if (_previewMode != _IepPreviewMode.week || weekPlan == null) {
      return '开始上课';
    }
    final IepLessonSession? session = _currentLessonSession();
    if (session != null && (session.isInProgress || session.isPaused)) {
      return '继续上课';
    }
    if (_selectedWeekIsCurrent(weekPlan)) {
      return _selectedWeekHasRecordedEntry(weekPlan) ? '修改记录' : '开始上课';
    }
    if (_selectedWeekIsPast(weekPlan)) {
      return _selectedWeekHasRecordedEntry(weekPlan) ? '修改记录' : '修/补记录';
    }
    return '开始上课';
  }

  bool _selectedWeekHasRecordedEntry(IepWeeklyPlan? weekPlan) {
    if (weekPlan == null) {
      return false;
    }
    return weekPlan.rows.any(
      (IepWeeklyPlanRow row) => row.completion.any(
        (String code) => code.trim().isNotEmpty,
      ),
    );
  }

  bool _hasAnyExecutionRecord() {
    final IepExecutionPlansSaved? plans = _executionPlans;
    if (plans == null) {
      return false;
    }
    return plans.weeklyPlans.any(
      (IepWeeklyPlanSaved item) => _selectedWeekHasRecordedEntry(item.plan),
    );
  }

  bool _monthHasAnyWeeklyExecutionRecord(int monthIndex) {
    final IepExecutionPlansSaved? plans = _executionPlans;
    if (plans == null) {
      return false;
    }
    return plans.weeklyPlans.any(
      (IepWeeklyPlanSaved item) =>
          item.targetMonthIndex == monthIndex &&
          _selectedWeekHasRecordedEntry(item.plan),
    );
  }

  bool _totalPlanHasAnyWeeklyExecutionRecord() {
    return _hasAnyExecutionRecord();
  }

  @override
  Widget build(BuildContext context) {
    final IepAssessmentRecordSummary? record = widget.record;
    final IepPlan? plan = _savedPlan?.plan;
    final IepMonthlyPlan? monthPlan =
        _executionPlans?.monthPlan(_previewMonthIndex());
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final bool canOpenLesson = _canOpenLessonForSelectedWeek(weekPlan);
    final bool weekHasAnyRecord = _selectedWeekHasRecordedEntry(weekPlan);
    final bool hasAnyExecutionRecord = _hasAnyExecutionRecord();
    final bool totalPlanRegenerateDisabled =
        _totalPlanHasAnyWeeklyExecutionRecord();
    final bool monthPlanRegenerateDisabled =
        _monthHasAnyWeeklyExecutionRecord(_previewMonthIndex());
    final bool regenerateDisabledWithHint = switch (_previewMode) {
      _IepPreviewMode.total => totalPlanRegenerateDisabled,
      _IepPreviewMode.month => monthPlanRegenerateDisabled,
      _IepPreviewMode.week => weekHasAnyRecord,
    };
    final bool generatePlanEnabled = switch (_previewMode) {
      _IepPreviewMode.total => true,
      _IepPreviewMode.month => _savedPlan?.hasContent == true,
      _IepPreviewMode.week => _savedPlan?.hasContent == true,
    };
    final bool currentPlanGenerated = switch (_previewMode) {
      _IepPreviewMode.total => _savedPlan?.hasContent == true,
      _IepPreviewMode.month => monthPlan != null,
      _IepPreviewMode.week => weekPlan != null,
    };
    final bool startClassEnabled = record != null &&
        _previewMode == _IepPreviewMode.week &&
        canOpenLesson &&
        !_loadingPlan &&
        !_generatingPlan;
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
            previewMode: _previewMode,
            previewMonthIndex: _previewMonthIndex(),
            previewWeek: _previewWeek,
            onStartClass: startClassEnabled
                ? _openLessonSession
                : () {
                    if (_previewMode != _IepPreviewMode.week) {
                      _showMessage('请选中周计划');
                      return;
                    }
                    _openLessonSession();
                  },
            startClassEnabled: startClassEnabled,
            startClassLabel: _lessonEntryButtonLabel(weekPlan),
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
            currentPlanGenerated: currentPlanGenerated,
            onGeneratePlan: _handleGeneratePlanRequest,
            editPeriodDisabledWithHint: hasAnyExecutionRecord,
            onDisabledEditPeriodTap: () {
              _showMessage('该记录卡已记录，不支持编辑周期');
            },
            regenerateDisabledWithHint: regenerateDisabledWithHint,
            onDisabledRegenerateTap: () {
              final String message = switch (_previewMode) {
                _IepPreviewMode.total => '已有月计划下的周计划被记录，不支持重新生成总计划',
                _IepPreviewMode.month => '当前月份下已有周计划被记录，不支持重新生成月计划',
                _IepPreviewMode.week => '该记录卡已记录，不支持重新生成',
              };
              _showMessage(message);
            },
            generateEnabled: generatePlanEnabled,
            onDisabledGenerateTap: () {
              _showMessage('请先生成IEP总计划，再生成月计划或周计划');
            },
            previewMode: _previewMode,
            previewMonth: _previewMonth,
            previewWeek: _previewWeek,
            totalPlanGenerated: _savedPlan?.hasContent == true,
            generatedMonthIndexes: _executionPlans?.monthlyPlans
                    .map((IepMonthlyPlanSaved item) => item.targetMonthIndex)
                    .toSet() ??
                <int>{},
            generatedWeekKeys: _executionPlans?.weeklyPlans
                    .map(
                      (IepWeeklyPlanSaved item) =>
                          '${item.targetMonthIndex}-${item.targetWeekIndex}',
                    )
                    .toSet() ??
                <String>{},
            resolveMonthRestWeekdays: _restWeekdaysForMonth,
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
              weekFallbackRestWeekdays: _previewMode == _IepPreviewMode.week
                  ? (weekPlan?.restWeekdays.isNotEmpty == true
                      ? _normalizeWeeklyRestWeekdays(weekPlan!.restWeekdays)
                      : _restWeekdaysForMonth(_previewMonthIndex()))
                  : _restWeekdaysForMonth(_previewMonthIndex()),
              loading: _loadingPlan,
              bootstrapLoading: !_hasCompletedInitialPlanLoad,
              generatingPlan: _generatingPlan,
              generationStatus: _generationStatus,
              generationText: _aiStreamText,
              generationProgress: _generationProgress,
              generationCostAmountCny: _generationCostAmountCny,
              error: _planError,
              onRetry: _handleRetryRequest,
              onGeneratePlan: _handleGeneratePlanRequest,
              generatePlanEnabled: generatePlanEnabled,
              onDisabledGeneratePlanTap: () {
                _showMessage('请先生成IEP总计划，再生成月计划或周计划');
              },
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
    required this.previewMode,
    required this.previewMonthIndex,
    required this.previewWeek,
    required this.onStartClass,
    required this.startClassEnabled,
    required this.startClassLabel,
  });

  final String title;
  final String statusText;
  final String periodText;
  final _IepPreviewMode previewMode;
  final int previewMonthIndex;
  final int previewWeek;
  final VoidCallback onStartClass;
  final bool startClassEnabled;
  final String startClassLabel;

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
          _ClassContextPill(
            previewMode: previewMode,
            previewMonthIndex: previewMonthIndex,
            previewWeek: previewWeek,
          ),
          const SizedBox(width: 10),
          _StartClassButton(
            onTap: onStartClass,
            enabled: startClassEnabled,
            label: startClassLabel,
          ),
        ],
      ),
    );
  }
}

class _ClassContextPill extends StatelessWidget {
  const _ClassContextPill({
    required this.previewMode,
    required this.previewMonthIndex,
    required this.previewWeek,
  });

  final _IepPreviewMode previewMode;
  final int previewMonthIndex;
  final int previewWeek;

  @override
  Widget build(BuildContext context) {
    final String text = switch (previewMode) {
      _IepPreviewMode.total => 'IEP总计划',
      _IepPreviewMode.month => '第$previewMonthIndex月',
      _IepPreviewMode.week => '第$previewMonthIndex月 · 第$previewWeek周',
    };
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF6EE),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: const Color(0xFFFFD3BA)),
      ),
      child: Text(
        text,
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
  const _StartClassButton({
    required this.onTap,
    required this.enabled,
    required this.label,
  });

  final VoidCallback onTap;
  final bool enabled;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 15),
          decoration: BoxDecoration(
            color: enabled ? _IepColors.orange : const Color(0xFFF3E5DA),
            borderRadius: BorderRadius.circular(18),
            boxShadow: enabled
                ? _iepShadow(
                    color: const Color(0x32E96F43),
                    blur: 12,
                    offset: const Offset(0, 5),
                  )
                : const <BoxShadow>[],
          ),
          child: Row(
            children: <Widget>[
              Icon(
                Icons.play_circle_fill_rounded,
                color: enabled ? Colors.white : _IepColors.muted,
                size: 19,
              ),
              const SizedBox(width: 6),
              Text(
                label,
                style: TextStyle(
                  color: enabled ? Colors.white : _IepColors.muted,
                  fontSize: 13,
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

class _PlanToolbar extends StatelessWidget {
  const _PlanToolbar({
    required this.onShowTotalPlan,
    required this.onShowMonthPlan,
    required this.onShowWeekPlan,
    required this.onEditPeriod,
    required this.onGeneratePlan,
    required this.onRegeneratePlan,
    required this.monthLabels,
    required this.periodMonthCount,
    required this.periodStart,
    required this.onPeriodMonthCountChanged,
    required this.syncingPeriod,
    required this.generatingPlan,
    required this.currentPlanGenerated,
    required this.editPeriodDisabledWithHint,
    required this.onDisabledEditPeriodTap,
    required this.generateEnabled,
    required this.onDisabledGenerateTap,
    required this.regenerateDisabledWithHint,
    required this.onDisabledRegenerateTap,
    required this.previewMode,
    required this.previewMonth,
    required this.previewWeek,
    required this.totalPlanGenerated,
    required this.generatedMonthIndexes,
    required this.generatedWeekKeys,
    required this.resolveMonthRestWeekdays,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final VoidCallback onEditPeriod;
  final VoidCallback onGeneratePlan;
  final VoidCallback onRegeneratePlan;
  final List<String> monthLabels;
  final int periodMonthCount;
  final DateTime periodStart;
  final ValueChanged<int> onPeriodMonthCountChanged;
  final bool syncingPeriod;
  final bool generatingPlan;
  final bool currentPlanGenerated;
  final bool editPeriodDisabledWithHint;
  final VoidCallback onDisabledEditPeriodTap;
  final bool generateEnabled;
  final VoidCallback onDisabledGenerateTap;
  final bool regenerateDisabledWithHint;
  final VoidCallback onDisabledRegenerateTap;
  final _IepPreviewMode previewMode;
  final String previewMonth;
  final int previewWeek;
  final bool totalPlanGenerated;
  final Set<int> generatedMonthIndexes;
  final Set<String> generatedWeekKeys;
  final List<int> Function(int monthIndex) resolveMonthRestWeekdays;

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
              previewMode: previewMode,
              previewMonth: previewMonth,
              previewWeek: previewWeek,
              totalPlanGenerated: totalPlanGenerated,
              generatedMonthIndexes: generatedMonthIndexes,
              generatedWeekKeys: generatedWeekKeys,
              resolveMonthRestWeekdays: resolveMonthRestWeekdays,
            ),
          ),
          const _ToolbarDivider(),
          _TableTinyAction(
            icon: syncingPeriod || generatingPlan
                ? Icons.hourglass_top_rounded
                : Icons.edit_calendar_rounded,
            label: syncingPeriod ? '同步中' : '编辑周期',
            enabled: !syncingPeriod &&
                !generatingPlan &&
                !editPeriodDisabledWithHint,
            onTap: syncingPeriod || generatingPlan
                ? null
                : (editPeriodDisabledWithHint
                    ? onDisabledEditPeriodTap
                    : onEditPeriod),
          ),
          const SizedBox(width: 8),
          _TableTinyAction(
            icon: generatingPlan
                ? Icons.hourglass_top_rounded
                : (currentPlanGenerated
                    ? Icons.refresh_rounded
                    : Icons.auto_awesome_rounded),
            label: generatingPlan
                ? '生成中'
                : (currentPlanGenerated ? '重新生成' : 'AI生成'),
            primary: true,
            width: 96,
            enabled: !generatingPlan &&
                (currentPlanGenerated
                    ? !regenerateDisabledWithHint
                    : generateEnabled),
            onTap: generatingPlan
                ? null
                : ((currentPlanGenerated
                        ? regenerateDisabledWithHint
                        : !generateEnabled)
                    ? (currentPlanGenerated
                        ? onDisabledRegenerateTap
                        : onDisabledGenerateTap)
                    : (currentPlanGenerated
                        ? onRegeneratePlan
                        : onGeneratePlan)),
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
    required this.previewMode,
    required this.previewMonth,
    required this.previewWeek,
    required this.totalPlanGenerated,
    required this.generatedMonthIndexes,
    required this.generatedWeekKeys,
    required this.resolveMonthRestWeekdays,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final List<String> monthLabels;
  final DateTime periodStart;
  final int periodMonthCount;
  final _IepPreviewMode previewMode;
  final String previewMonth;
  final int previewWeek;
  final bool totalPlanGenerated;
  final Set<int> generatedMonthIndexes;
  final Set<String> generatedWeekKeys;
  final List<int> Function(int monthIndex) resolveMonthRestWeekdays;

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

  void _syncSelectionFromWidget() {
    final String nextSection = switch (widget.previewMode) {
      _IepPreviewMode.total => 'iep',
      _IepPreviewMode.month => 'month',
      _IepPreviewMode.week => 'week',
    };
    final String nextMonth = widget.monthLabels.contains(widget.previewMonth)
        ? widget.previewMonth
        : (widget.monthLabels.isNotEmpty ? widget.monthLabels.first : '5月');
    final int? nextWeek =
        widget.previewMode == _IepPreviewMode.week ? widget.previewWeek : null;
    _selectedSection = nextSection;
    _selectedMonth = nextMonth;
    _selectedWeek = nextWeek;
  }

  @override
  void didUpdateWidget(covariant _ScrollablePlanNav oldWidget) {
    super.didUpdateWidget(oldWidget);
    _syncSelectionFromWidget();
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
    _syncSelectionFromWidget();
    _scrollController = ScrollController();
    _scrollController.addListener(_syncHints);
    _hintController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 760),
    );
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
    _syncHintAnimation();
  }

  void _syncHintAnimation() {
    final bool shouldAnimate = _showLeftHint || _showRightHint;
    if (shouldAnimate && !_hintController.isAnimating) {
      _hintController.repeat(reverse: true);
    } else if (!shouldAnimate && _hintController.isAnimating) {
      _hintController.stop();
      _hintController.value = 0;
    }
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
    final int monthIndex = widget.monthLabels.indexOf(_selectedMonth) + 1;
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
    return _lastAvailableWeekInMonthRange(
      selectedMonthRange,
      restWeekdays: widget.resolveMonthRestWeekdays(
        monthIndex < 1 ? 1 : monthIndex,
      ),
    );
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
                        generated: widget.totalPlanGenerated,
                        width: 92,
                        onTap: () => _selectPlan('iep'),
                      ),
                      const _PlanNavLabel(text: '月计划'),
                      ...widget.monthLabels.map((String month) {
                        final int monthIndex =
                            widget.monthLabels.indexOf(month) + 1;
                        return _PlanTab(
                          text: month,
                          active: _selectedSection == 'month' &&
                              _selectedMonth == month,
                          generated:
                              widget.generatedMonthIndexes.contains(monthIndex),
                          width: 54,
                          onTap: () => _selectMonth(month),
                        );
                      }),
                      const _PlanNavLabel(text: '周计划'),
                      ...List<Widget>.generate(weekCount, (int index) {
                        final int weekNumber = index + 1;
                        final int monthIndex =
                            widget.monthLabels.indexOf(_selectedMonth) + 1;
                        return _PlanTab(
                          text: '${_selectedMonth} W$weekNumber',
                          width: 68,
                          active: _selectedSection == 'week' &&
                              _selectedWeek == weekNumber,
                          generated: widget.generatedWeekKeys
                              .contains('$monthIndex-$weekNumber'),
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
    this.enabled = true,
    this.primary = false,
    this.width,
    this.onTap,
  });

  final IconData icon;
  final String label;
  final bool enabled;
  final bool primary;
  final double? width;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool filled = primary;
    final Color backgroundColor = !enabled
        ? (filled ? const Color(0xFFF3E1D6) : const Color(0xFFF8EEE6))
        : (filled ? _IepColors.orange : Colors.white);
    final Color borderColor = filled
        ? (enabled ? _IepColors.orange : const Color(0xFFF3E1D6))
        : _IepColors.line;
    final Color foregroundColor =
        enabled ? (filled ? Colors.white : _IepColors.text) : _IepColors.muted;
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(15),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Container(
          height: 30,
          width: width,
          padding: EdgeInsets.symmetric(horizontal: filled ? 11 : 9),
          decoration: BoxDecoration(
            color: backgroundColor,
            borderRadius: BorderRadius.circular(15),
            border: Border.all(color: borderColor),
            boxShadow: filled && enabled
                ? _iepShadow(
                    color: const Color(0x2EE96F43),
                    blur: 10,
                    offset: const Offset(0, 4),
                  )
                : const <BoxShadow>[],
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              Icon(
                icon,
                color: foregroundColor,
                size: 16,
              ),
              const SizedBox(width: 5),
              Text(
                label,
                style: TextStyle(
                  color: foregroundColor,
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
