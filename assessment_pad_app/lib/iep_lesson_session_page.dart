part of 'iep_center_page.dart';

class _IepLessonSessionDraft {
  const _IepLessonSessionDraft({
    required this.record,
    required this.durationMonths,
    required this.targetMonthIndex,
    required this.targetWeekIndex,
    required this.weeklyPlan,
    required this.studentName,
    required this.gender,
    required this.ageLabel,
    required this.teacherName,
    required this.courseName,
    required this.planTitle,
    required this.stageLabel,
    required this.periodLabel,
    required this.weekLabel,
    required this.initialSelectedDateIndex,
    required this.lessonDate,
    required this.lessonSession,
    String? trainingDateLabel,
    List<String>? weekDateOptions,
    List<String>? completionColumnLabels,
    required this.preparation,
    required this.goals,
    required this.tasks,
  })  : _trainingDateLabel = trainingDateLabel,
        _weekDateOptions = weekDateOptions,
        _completionColumnLabels = completionColumnLabels;

  final IepAssessmentRecordSummary record;
  final int durationMonths;
  final int targetMonthIndex;
  final int targetWeekIndex;
  final IepWeeklyPlan weeklyPlan;
  final String studentName;
  final String gender;
  final String ageLabel;
  final String teacherName;
  final String courseName;
  final String planTitle;
  final String stageLabel;
  final String periodLabel;
  final String weekLabel;
  final int initialSelectedDateIndex;
  final String lessonDate;
  final IepLessonSession lessonSession;
  final String? _trainingDateLabel;
  final List<String>? _weekDateOptions;
  final List<String>? _completionColumnLabels;
  final String preparation;
  final List<String> goals;
  final List<_IepLessonTaskDraft> tasks;

  String get trainingDateLabel => _trainingDateLabel ?? weekLabel;

  List<String> get weekDateOptions => _weekDateOptions ?? const <String>[];

  List<String> get completionColumnLabels =>
      _completionColumnLabels ?? const <String>[];
}

enum _IepLessonSessionExitAction { paused, completed, back }

class _IepLessonSessionExitResult {
  const _IepLessonSessionExitResult(this.action);

  final _IepLessonSessionExitAction action;
}

class _IepLessonFullscreenViewport extends StatelessWidget {
  const _IepLessonFullscreenViewport({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final Size screenSize = MediaQuery.sizeOf(context);
        final double viewportWidth = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : screenSize.width;
        final double viewportHeight = constraints.maxHeight.isFinite
            ? constraints.maxHeight
            : screenSize.height;
        final double designWidth = padDesignWidthForViewport(
          viewportWidth,
          viewportHeight,
        );
        return ColoredBox(
          color: _IepColors.page,
          child: Center(
            child: FittedBox(
              fit: BoxFit.contain,
              child: SizedBox(
                width: designWidth,
                height: padDesignHeight,
                child: child,
              ),
            ),
          ),
        );
      },
    );
  }
}

class _IepLessonTaskDraft {
  const _IepLessonTaskDraft({
    required this.sourceRowIndex,
    required this.title,
    required this.subtitle,
    required this.domain,
    required this.goal,
    required this.materials,
    required this.steps,
    required this.tips,
    List<String>? completionCodes,
  }) : _completionCodes = completionCodes;

  final int sourceRowIndex;
  final String title;
  final String subtitle;
  final String domain;
  final String goal;
  final String materials;
  final List<String> steps;
  final List<String> tips;
  final List<String>? _completionCodes;

  List<String> get completionCodes => _completionCodes ?? const <String>[];
}

class _IepLessonSessionPage extends StatefulWidget {
  const _IepLessonSessionPage({
    required this.draft,
    required this.planClient,
    this.onPlansSaved,
  });

  final _IepLessonSessionDraft draft;
  final IepPlanClient planClient;
  final ValueChanged<IepExecutionPlansSaved>? onPlansSaved;

  @override
  State<_IepLessonSessionPage> createState() => _IepLessonSessionPageState();
}

class _IepLessonSessionPageState extends State<_IepLessonSessionPage>
    with WidgetsBindingObserver {
  static const String _authTokenStorageKey = 'auth_token';

  int _selectedTaskIndex = 0;
  int _selectedDateIndex = 0;
  List<List<String>> _taskCompletionCodes = <List<String>>[];
  String _courseName = '康复教学';
  late IepWeeklyPlan _weeklyPlan;
  late IepLessonSession _lessonSession;
  int _displayElapsedSeconds = 0;
  bool _weeklyPlanDirty = false;
  bool _needsResave = false;
  int _weeklyPlanRevision = 0;
  Future<bool>? _activeSaveFuture;
  Timer? _lessonClockTimer;
  Timer? _lessonHeartbeatTimer;
  bool _sessionPausedByLifecycle = false;
  bool _closingPage = false;
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _weeklyPlan = widget.draft.weeklyPlan;
    _taskCompletionCodes = _normalizedCompletionCodes(widget.draft);
    _selectedDateIndex = _initialSelectedDateIndexFor(widget.draft);
    _courseName = widget.draft.courseName.trim().isEmpty
        ? '康复教学'
        : widget.draft.courseName.trim();
    _lessonSession = widget.draft.lessonSession;
    _displayElapsedSeconds = _lessonSession.elapsedSeconds;
    _syncLessonRuntime();
  }

  @override
  void didUpdateWidget(covariant _IepLessonSessionPage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!identical(oldWidget.draft, widget.draft)) {
      _weeklyPlan = widget.draft.weeklyPlan;
      _taskCompletionCodes = _normalizedCompletionCodes(widget.draft);
      _selectedTaskIndex = 0;
      _selectedDateIndex = _initialSelectedDateIndexFor(widget.draft);
      _courseName = widget.draft.courseName.trim().isEmpty
          ? '康复教学'
          : widget.draft.courseName.trim();
      _lessonSession = widget.draft.lessonSession;
      _displayElapsedSeconds = _lessonSession.elapsedSeconds;
      _weeklyPlanDirty = false;
      _needsResave = false;
      _weeklyPlanRevision = 0;
      _activeSaveFuture = null;
      _sessionPausedByLifecycle = false;
      _closingPage = false;
      _syncLessonRuntime();
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (_closingPage) {
      return;
    }
    switch (state) {
      case AppLifecycleState.resumed:
        if (_sessionPausedByLifecycle) {
          unawaited(_resumeLessonSessionAfterLifecycle());
        }
        break;
      case AppLifecycleState.inactive:
        break;
      case AppLifecycleState.hidden:
      case AppLifecycleState.paused:
      case AppLifecycleState.detached:
        unawaited(_pauseLessonSessionForLifecycle());
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _stopLessonRuntime();
    super.dispose();
  }

  int _initialSelectedDateIndexFor(_IepLessonSessionDraft draft) {
    if (draft.weekDateOptions.isEmpty) {
      return 0;
    }
    return draft.initialSelectedDateIndex.clamp(
      0,
      math.max(0, draft.weekDateOptions.length - 1),
    );
  }

  List<List<String>> _normalizedCompletionCodes(_IepLessonSessionDraft draft) {
    final int columnCount = math.max(
      1,
      draft.completionColumnLabels.isNotEmpty
          ? draft.completionColumnLabels.length
          : draft.weekDateOptions.length,
    );
    return draft.tasks.map((_IepLessonTaskDraft task) {
      return List<String>.generate(columnCount, (int index) {
        if (index < task.completionCodes.length) {
          return task.completionCodes[index].trim();
        }
        return '';
      });
    }).toList(growable: false);
  }

  void _syncLessonRuntime() {
    _stopLessonRuntime();
    if (!_lessonSession.isInProgress) {
      return;
    }
    _lessonClockTimer = Timer.periodic(const Duration(seconds: 1), (
      Timer timer,
    ) {
      if (!mounted || !_lessonSession.isInProgress) {
        return;
      }
      setState(() {
        _displayElapsedSeconds += 1;
      });
    });
    _lessonHeartbeatTimer = Timer.periodic(const Duration(seconds: 20), (
      Timer timer,
    ) {
      if (!mounted || !_lessonSession.isInProgress || _closingPage) {
        return;
      }
      unawaited(_sendLessonHeartbeat());
    });
  }

  void _stopLessonRuntime() {
    _lessonClockTimer?.cancel();
    _lessonClockTimer = null;
    _lessonHeartbeatTimer?.cancel();
    _lessonHeartbeatTimer = null;
  }

  void _applyLessonSession(IepLessonSession session) {
    _lessonSession = session;
    _displayElapsedSeconds = session.elapsedSeconds;
    _syncLessonRuntime();
    if (!mounted) {
      return;
    }
    setState(() {});
  }

  String _lessonDateValue() {
    final String lessonDate = widget.draft.lessonDate.trim();
    if (lessonDate.isNotEmpty) {
      return lessonDate;
    }
    if (_lessonSession.lessonDate.trim().isNotEmpty) {
      return _lessonSession.lessonDate.trim();
    }
    final List<String> weekDates = _weeklyPlan.weekDates;
    if (_selectedDateIndex >= 0 && _selectedDateIndex < weekDates.length) {
      return weekDates[_selectedDateIndex].trim();
    }
    return '';
  }

  IepLessonSession _fallbackLessonSession(String status) {
    return IepLessonSession(
      lessonDate: _lessonDateValue(),
      weekDateIndex: _lessonSession.weekDateIndex > 0
          ? _lessonSession.weekDateIndex
          : _selectedDateIndex + 1,
      status: status,
      elapsedSeconds: _displayElapsedSeconds,
      startedAt: _lessonSession.startedAt,
      lastResumedAt: _lessonSession.lastResumedAt,
      lastHeartbeatAt: _lessonSession.lastHeartbeatAt,
      pausedAt: _lessonSession.pausedAt,
      endedAt: _lessonSession.endedAt,
      updatedTime: _lessonSession.updatedTime,
    );
  }

  IepLessonSession _resolveLessonSessionFromState(
    IepLessonSessionWeekState state,
    String lessonDate,
  ) {
    return state.sessionForDate(lessonDate) ??
        state.currentSession ??
        _fallbackLessonSession(_lessonSession.status);
  }

  Future<String> _readAuthToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_authTokenStorageKey) ?? '';
  }

  Future<bool> _startLessonSessionRemote({bool silent = false}) async {
    final String lessonDate = _lessonDateValue();
    if (lessonDate.isEmpty) {
      return false;
    }
    try {
      final String token = await _readAuthToken();
      if (token.trim().isEmpty) {
        throw const IepPlanApiException('请先登录后再开始上课');
      }
      final IepLessonSessionWeekState state =
          await widget.planClient.startLessonSession(
        token,
        record: widget.draft.record,
        durationMonths: widget.draft.durationMonths,
        targetMonthIndex: widget.draft.targetMonthIndex,
        targetWeekIndex: widget.draft.targetWeekIndex,
        lessonDate: lessonDate,
      );
      final IepLessonSession session = state.sessionForDate(lessonDate) ??
          state.currentSession ??
          _fallbackLessonSession('in_progress');
      _applyLessonSession(session);
      return true;
    } on IepPlanApiException catch (error) {
      if (!silent) {
        _showMessage(error.message);
      }
      return false;
    } on Object catch (error) {
      if (!silent) {
        _showMessage('恢复上课状态失败：$error');
      }
      return false;
    }
  }

  Future<bool> _pauseLessonSessionRemote({bool silent = false}) async {
    final String lessonDate = _lessonDateValue();
    if (lessonDate.isEmpty) {
      return false;
    }
    try {
      final String token = await _readAuthToken();
      if (token.trim().isEmpty) {
        throw const IepPlanApiException('请先登录后再暂停上课');
      }
      final IepLessonSessionWeekState state =
          await widget.planClient.pauseLessonSession(
        token,
        record: widget.draft.record,
        durationMonths: widget.draft.durationMonths,
        targetMonthIndex: widget.draft.targetMonthIndex,
        targetWeekIndex: widget.draft.targetWeekIndex,
        lessonDate: lessonDate,
      );
      final IepLessonSession session = state.sessionForDate(lessonDate) ??
          state.currentSession ??
          _fallbackLessonSession('paused');
      _applyLessonSession(
        session.status.trim().isEmpty
            ? _fallbackLessonSession('paused')
            : session,
      );
      return true;
    } on IepPlanApiException catch (error) {
      if (!silent) {
        _showMessage(error.message);
      }
      return false;
    } on Object catch (error) {
      if (!silent) {
        _showMessage('暂停上课失败：$error');
      }
      return false;
    }
  }

  Future<bool> _completeLessonSessionRemote({bool silent = false}) async {
    final String lessonDate = _lessonDateValue();
    if (lessonDate.isEmpty) {
      return false;
    }
    try {
      final String token = await _readAuthToken();
      if (token.trim().isEmpty) {
        throw const IepPlanApiException('请先登录后再结束上课');
      }
      final IepLessonSessionWeekState state =
          await widget.planClient.completeLessonSession(
        token,
        record: widget.draft.record,
        durationMonths: widget.draft.durationMonths,
        targetMonthIndex: widget.draft.targetMonthIndex,
        targetWeekIndex: widget.draft.targetWeekIndex,
        lessonDate: lessonDate,
      );
      final IepLessonSession session = state.sessionForDate(lessonDate) ??
          state.currentSession ??
          _fallbackLessonSession('completed');
      _applyLessonSession(
        session.status.trim().isEmpty
            ? _fallbackLessonSession('completed')
            : session,
      );
      return true;
    } on IepPlanApiException catch (error) {
      if (!silent) {
        _showMessage(error.message);
      }
      return false;
    } on Object catch (error) {
      if (!silent) {
        _showMessage('结束上课失败：$error');
      }
      return false;
    }
  }

  Future<void> _sendLessonHeartbeat() async {
    final String lessonDate = _lessonDateValue();
    if (lessonDate.isEmpty || !_lessonSession.isInProgress) {
      return;
    }
    try {
      final String token = await _readAuthToken();
      if (token.trim().isEmpty) {
        return;
      }
      final IepLessonSessionWeekState state =
          await widget.planClient.heartbeatLessonSession(
        token,
        record: widget.draft.record,
        durationMonths: widget.draft.durationMonths,
        targetMonthIndex: widget.draft.targetMonthIndex,
        targetWeekIndex: widget.draft.targetWeekIndex,
        lessonDate: lessonDate,
      );
      final IepLessonSession session =
          _resolveLessonSessionFromState(state, lessonDate);
      if (!_lessonSession.isInProgress && !session.isInProgress) {
        return;
      }
      _applyLessonSession(session);
    } on Object {
      // 心跳失败先静默，后端会在超时后自动转为暂停。
    }
  }

  Future<void> _pauseLessonSessionForLifecycle() async {
    if (_closingPage || !_lessonSession.isInProgress) {
      return;
    }
    final bool saved = await _flushWeeklyPlanSave(silent: true);
    if (!saved) {
      return;
    }
    final bool paused = await _pauseLessonSessionRemote(silent: true);
    if (paused) {
      _sessionPausedByLifecycle = true;
    }
  }

  Future<void> _resumeLessonSessionAfterLifecycle() async {
    if (_closingPage) {
      return;
    }
    final bool resumed = await _startLessonSessionRemote(silent: true);
    if (resumed) {
      _sessionPausedByLifecycle = false;
    }
  }

  String _lessonStatusText() {
    final String duration = _formatLessonElapsed(_displayElapsedSeconds);
    if (_lessonSession.isCompleted) {
      return '已结束 $duration';
    }
    if (_lessonSession.isPaused) {
      return '已暂停 $duration';
    }
    return '进行中 $duration';
  }

  String _formatLessonElapsed(int totalSeconds) {
    final int safeSeconds = math.max(0, totalSeconds);
    final int hours = safeSeconds ~/ 3600;
    final int minutes = (safeSeconds % 3600) ~/ 60;
    final int seconds = safeSeconds % 60;
    if (hours > 0) {
      return '$hours:${minutes.toString().padLeft(2, '0')}:${seconds.toString().padLeft(2, '0')}';
    }
    return '${minutes.toString().padLeft(2, '0')}:${seconds.toString().padLeft(2, '0')}';
  }

  void _showMessage(
    String message, {
    PadMessageTone tone = PadMessageTone.info,
  }) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _messageController.show(
      context,
      message,
      tone: tone,
      topMargin: 12,
      key: 'iep-lesson-message',
    );
  }

  IepWeeklyPlan _weeklyPlanWithCurrentState() {
    final List<IepWeeklyPlanRow> rows = List<IepWeeklyPlanRow>.generate(
      _weeklyPlan.rows.length,
      (int rowIndex) {
        final IepWeeklyPlanRow row = _weeklyPlan.rows[rowIndex];
        List<String>? completionCodes;
        for (int taskIndex = 0;
            taskIndex < widget.draft.tasks.length;
            taskIndex++) {
          if (widget.draft.tasks[taskIndex].sourceRowIndex == rowIndex &&
              taskIndex < _taskCompletionCodes.length) {
            completionCodes = _taskCompletionCodes[taskIndex];
            break;
          }
        }
        return IepWeeklyPlanRow(
          project: row.project,
          content: row.content,
          completion: completionCodes == null
              ? List<String>.from(row.completion)
              : List<String>.from(completionCodes),
        );
      },
      growable: false,
    );
    return IepWeeklyPlan(
      title: _weeklyPlan.title,
      student: _weeklyPlan.student,
      teacherName: _weeklyPlan.teacherName,
      courseName: _courseName.trim().isEmpty ? '康复教学' : _courseName.trim(),
      trainingDate: _weeklyPlan.trainingDate,
      preparation: _weeklyPlan.preparation,
      weekDates: List<String>.from(_weeklyPlan.weekDates),
      rows: rows,
    );
  }

  Future<bool> _saveWeeklyPlanNow({bool silent = false}) async {
    final int revision = _weeklyPlanRevision;
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      if (token.trim().isEmpty) {
        throw const IepPlanApiException('请先登录后再保存上课记录');
      }
      final IepExecutionPlansSaved saved =
          await widget.planClient.saveWeeklyPlan(
        token,
        record: widget.draft.record,
        durationMonths: widget.draft.durationMonths,
        targetMonthIndex: widget.draft.targetMonthIndex,
        targetWeekIndex: widget.draft.targetWeekIndex,
        plan: _weeklyPlanWithCurrentState(),
      );
      if (!mounted) {
        return true;
      }
      final IepWeeklyPlan? savedWeekPlan = saved.weekPlan(
        widget.draft.targetMonthIndex,
        widget.draft.targetWeekIndex,
      );
      setState(() {
        if (revision == _weeklyPlanRevision) {
          _weeklyPlanDirty = false;
        }
        if (savedWeekPlan != null) {
          _weeklyPlan = savedWeekPlan;
        }
      });
      widget.onPlansSaved?.call(saved);
      if (!silent) {
        _showMessage('上课记录已保存', tone: PadMessageTone.success);
      }
      if (_needsResave && mounted) {
        _needsResave = false;
        await _queueWeeklyPlanSave(silent: true);
      }
      return true;
    } on IepPlanApiException catch (error) {
      if (!mounted) {
        return false;
      }
      if (!silent) {
        _showMessage(error.message);
      }
      return false;
    } on Object catch (error) {
      if (!mounted) {
        return false;
      }
      if (!silent) {
        _showMessage('保存上课记录失败：$error');
      }
      return false;
    }
  }

  Future<bool> _queueWeeklyPlanSave({bool silent = true}) {
    final Future<bool>? active = _activeSaveFuture;
    if (active != null) {
      return active;
    }
    final Future<bool> future = _saveWeeklyPlanNow(silent: silent);
    _activeSaveFuture = future;
    future.whenComplete(() {
      if (identical(_activeSaveFuture, future)) {
        _activeSaveFuture = null;
      }
    });
    return future;
  }

  Future<bool> _flushWeeklyPlanSave({bool silent = false}) async {
    final Future<bool>? active = _activeSaveFuture;
    if (active != null) {
      return active;
    }
    if (!_weeklyPlanDirty) {
      return true;
    }
    return _saveWeeklyPlanNow(silent: silent);
  }

  void _markWeeklyPlanDirty() {
    _weeklyPlanRevision += 1;
    _weeklyPlanDirty = true;
    if (_activeSaveFuture != null) {
      _needsResave = true;
    }
  }

  void _updateCompletionCode(String code) {
    if (_taskCompletionCodes.isEmpty) {
      return;
    }
    final int taskIndex = _selectedTaskIndex.clamp(
      0,
      math.max(0, _taskCompletionCodes.length - 1),
    );
    final List<String> taskCodes = _taskCompletionCodes[taskIndex];
    if (taskCodes.isEmpty) {
      return;
    }
    final int dateIndex = _selectedDateIndex.clamp(
      0,
      math.max(0, taskCodes.length - 1),
    );
    final List<List<String>> next = _taskCompletionCodes
        .map((List<String> item) => List<String>.from(item))
        .toList(growable: false);
    next[taskIndex][dateIndex] = next[taskIndex][dateIndex] == code ? '' : code;
    setState(() {
      _taskCompletionCodes = next;
      _markWeeklyPlanDirty();
    });
    _queueWeeklyPlanSave();
  }

  Future<void> _editCourseName() async {
    final String? value = await showDialog<String>(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepLessonCourseNameDialog(initialValue: _courseName),
        );
      },
    );
    if (!mounted || value == null || value == _courseName) {
      return;
    }
    setState(() {
      _courseName = value;
      _markWeeklyPlanDirty();
    });
    _queueWeeklyPlanSave();
  }

  String _selectedDateLabelFor(_IepLessonSessionDraft draft) {
    final List<String> weekDateOptions = draft.weekDateOptions;
    if (weekDateOptions.isEmpty) {
      return draft.trainingDateLabel;
    }
    final int selectedDateIndex = _selectedDateIndex.clamp(
      0,
      math.max(0, weekDateOptions.length - 1),
    );
    return weekDateOptions[selectedDateIndex];
  }

  List<int> _missingTaskIndexesForDate({
    required _IepLessonSessionDraft draft,
    required int selectedDateIndex,
  }) {
    final List<_IepLessonTaskDraft> tasks = draft.tasks;
    if (_taskCompletionCodes.length != tasks.length) {
      _taskCompletionCodes = _normalizedCompletionCodes(draft);
    }
    final List<int> indexes = <int>[];
    for (int index = 0; index < _taskCompletionCodes.length; index++) {
      final List<String> taskCodes = _taskCompletionCodes[index];
      if (selectedDateIndex >= taskCodes.length) {
        indexes.add(index);
        continue;
      }
      if (taskCodes[selectedDateIndex].trim().isEmpty) {
        indexes.add(index);
      }
    }
    return indexes;
  }

  void _fillMissingTaskCodes({
    required int selectedDateIndex,
    required List<int> taskIndexes,
    required String code,
  }) {
    final List<List<String>> next = _taskCompletionCodes
        .map((List<String> item) => List<String>.from(item))
        .toList(growable: false);
    for (final int taskIndex in taskIndexes) {
      if (taskIndex < 0 || taskIndex >= next.length) {
        continue;
      }
      final List<String> taskCodes = next[taskIndex];
      if (selectedDateIndex < 0 || selectedDateIndex >= taskCodes.length) {
        continue;
      }
      if (taskCodes[selectedDateIndex].trim().isEmpty) {
        taskCodes[selectedDateIndex] = code;
      }
    }
    setState(() {
      _taskCompletionCodes = next;
      _markWeeklyPlanDirty();
    });
  }

  Future<void> _handleEndLesson() async {
    final _IepLessonSessionDraft draft = widget.draft;
    final List<String> weekDateOptions = draft.weekDateOptions;
    final int selectedDateIndex = weekDateOptions.isEmpty
        ? 0
        : _selectedDateIndex.clamp(0, math.max(0, weekDateOptions.length - 1));
    final List<int> missingTaskIndexes = _missingTaskIndexesForDate(
      draft: draft,
      selectedDateIndex: selectedDateIndex,
    );
    if (missingTaskIndexes.isEmpty) {
      if (_closingPage) {
        return;
      }
      _closingPage = true;
      final bool saved = await _flushWeeklyPlanSave();
      if (!mounted || !saved) {
        _closingPage = false;
        return;
      }
      final bool completed = await _completeLessonSessionRemote();
      if (!mounted || !completed) {
        _closingPage = false;
        return;
      }
      Navigator.of(context).pop(
        const _IepLessonSessionExitResult(
            _IepLessonSessionExitAction.completed),
      );
      return;
    }
    final String? code = await showDialog<String>(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepLessonMissingRecordDialog(
            selectedDateLabel: _selectedDateLabelFor(draft),
            missingTasks: missingTaskIndexes
                .where((int index) => index >= 0 && index < draft.tasks.length)
                .map((int index) => draft.tasks[index])
                .toList(growable: false),
          ),
        );
      },
    );
    if (!mounted || code == null || code.trim().isEmpty) {
      return;
    }
    _fillMissingTaskCodes(
      selectedDateIndex: selectedDateIndex,
      taskIndexes: missingTaskIndexes,
      code: code,
    );
    if (_closingPage) {
      return;
    }
    _closingPage = true;
    final bool saved = await _flushWeeklyPlanSave();
    if (!mounted || !saved) {
      _closingPage = false;
      return;
    }
    final bool completed = await _completeLessonSessionRemote();
    if (!mounted || !completed) {
      _closingPage = false;
      return;
    }
    Navigator.of(context).pop(const _IepLessonSessionExitResult(
        _IepLessonSessionExitAction.completed));
  }

  Future<void> _handleBackPressed() async {
    if (_closingPage) {
      return;
    }
    _closingPage = true;
    final bool saved = await _flushWeeklyPlanSave();
    if (!mounted || !saved) {
      _closingPage = false;
      return;
    }
    final bool paused = await _pauseLessonSessionRemote();
    if (!mounted || !paused) {
      _closingPage = false;
      return;
    }
    Navigator.of(context).pop(
        const _IepLessonSessionExitResult(_IepLessonSessionExitAction.paused));
  }

  Future<void> _handlePauseLesson() async {
    await _handleBackPressed();
  }

  @override
  Widget build(BuildContext context) {
    final _IepLessonSessionDraft draft = widget.draft;
    final List<_IepLessonTaskDraft> tasks = draft.tasks;
    final List<String> weekDateOptions = draft.weekDateOptions;
    if (_taskCompletionCodes.length != tasks.length) {
      _taskCompletionCodes = _normalizedCompletionCodes(draft);
    }
    final int selectedIndex = tasks.isEmpty
        ? 0
        : _selectedTaskIndex.clamp(0, math.max(0, tasks.length - 1));
    final int selectedDateIndex = weekDateOptions.isEmpty
        ? 0
        : _selectedDateIndex.clamp(0, math.max(0, weekDateOptions.length - 1));
    final _IepLessonTaskDraft? selectedTask =
        tasks.isEmpty ? null : tasks[selectedIndex];
    final String selectedDateLabel = weekDateOptions.isEmpty
        ? draft.trainingDateLabel
        : weekDateOptions[selectedDateIndex];
    final int recordedCount = _taskCompletionCodes.where((List<String> item) {
      return selectedDateIndex < item.length &&
          item[selectedDateIndex].isNotEmpty;
    }).length;

    return PopScope(
      canPop: false,
      onPopInvoked: (bool didPop) async {
        if (didPop) {
          return;
        }
        await _handleBackPressed();
      },
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final double width =
              constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
          final bool compact = width < 1180;
          final double outer = compact ? 14 : 20;
          final double gap = compact ? 10 : 14;
          final double leftWidth = compact ? 248 : 272;
          final double rightWidth = compact ? 278 : 320;
          final double centerWidth =
              width - outer * 2 - leftWidth - rightWidth - gap * 2;

          return ColoredBox(
            color: const Color(0xFFFFF7EE),
            child: Stack(
              children: <Widget>[
                Positioned.fill(
                  child: CustomPaint(painter: _IepLessonBackgroundPainter()),
                ),
                Positioned(
                  left: 0,
                  right: 0,
                  top: 0,
                  child: _IepLessonTopBar(
                    onBack: _handleBackPressed,
                    onPause: _handlePauseLesson,
                    draft: draft,
                    selectedDateLabel: selectedDateLabel,
                    statusText: _lessonStatusText(),
                    onEndLesson: _handleEndLesson,
                  ),
                ),
                Positioned(
                  left: outer,
                  top: 84,
                  width: leftWidth,
                  height: 660,
                  child: _IepLessonTaskRail(
                    draft: draft,
                    selectedIndex: selectedIndex,
                    selectedDateIndex: selectedDateIndex,
                    recordedCount: recordedCount,
                    completionCodes: _taskCompletionCodes,
                    onTaskSelected: (int index) {
                      setState(() {
                        _selectedTaskIndex = index;
                      });
                    },
                  ),
                ),
                Positioned(
                  left: outer + leftWidth + gap,
                  top: 84,
                  width: centerWidth,
                  height: 660,
                  child: _IepLessonMainPanel(
                    draft: draft,
                    task: selectedTask,
                    taskIndex: selectedIndex,
                    selectedDateLabel: selectedDateLabel,
                    courseName: _courseName,
                    onEditCourseName: _editCourseName,
                    hasPreviousTask: tasks.isNotEmpty && selectedIndex > 0,
                    hasNextTask:
                        tasks.isNotEmpty && selectedIndex < tasks.length - 1,
                    onPreviousTask: tasks.isNotEmpty && selectedIndex > 0
                        ? () {
                            setState(() {
                              _selectedTaskIndex = selectedIndex - 1;
                            });
                          }
                        : null,
                    onNextTask:
                        tasks.isNotEmpty && selectedIndex < tasks.length - 1
                            ? () {
                                setState(() {
                                  _selectedTaskIndex = selectedIndex + 1;
                                });
                              }
                            : null,
                  ),
                ),
                Positioned(
                  right: outer,
                  top: 84,
                  width: rightWidth,
                  height: 660,
                  child: _IepLessonRecordPanel(
                    draft: draft,
                    task: selectedTask,
                    currentCodes: selectedIndex < _taskCompletionCodes.length
                        ? _taskCompletionCodes[selectedIndex]
                        : const <String>[],
                    weekDateOptions: weekDateOptions,
                    selectedDateIndex: selectedDateIndex,
                    onCodeSelected: _updateCompletionCode,
                    selectedDateLabel: selectedDateLabel,
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _IepLessonBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint wash = Paint()
      ..shader = const LinearGradient(
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
        colors: <Color>[Color(0xFFFFFBF7), Color(0xFFFFF3E7)],
      ).createShader(Offset.zero & size);
    canvas.drawRect(Offset.zero & size, wash);

    final Paint circleA = Paint()..color = const Color(0x22F3C39D);
    final Paint circleB = Paint()..color = const Color(0x14E9854E);
    canvas.drawCircle(
        Offset(size.width * .14, size.height * .08), 120, circleA);
    canvas.drawCircle(
        Offset(size.width * .84, size.height * .18), 140, circleB);
    canvas.drawCircle(
        Offset(size.width * .74, size.height * .82), 180, circleA);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _IepLessonTopBar extends StatelessWidget {
  const _IepLessonTopBar({
    required this.onBack,
    required this.onPause,
    required this.draft,
    required this.selectedDateLabel,
    required this.statusText,
    required this.onEndLesson,
  });

  final VoidCallback onBack;
  final VoidCallback onPause;
  final _IepLessonSessionDraft draft;
  final String selectedDateLabel;
  final String statusText;
  final VoidCallback onEndLesson;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border(
          bottom: BorderSide(color: _IepColors.line.withOpacity(.78)),
        ),
      ),
      child: Row(
        children: <Widget>[
          _IepLessonBackButton(onTap: onBack),
          const SizedBox(width: 16),
          _IepLessonStudentCard(draft: draft),
          const SizedBox(width: 12),
          _IepLessonMetaBadge(
            icon: Icons.assignment_rounded,
            text: draft.weekLabel,
          ),
          const SizedBox(width: 8),
          _IepLessonMetaBadge(
            icon: Icons.today_rounded,
            text: selectedDateLabel,
          ),
          const SizedBox(width: 8),
          _IepLessonMetaBadge(
            icon: Icons.schedule_rounded,
            text: statusText,
            tone: _IepLessonBadgeTone.orange,
          ),
          const Spacer(),
          _IepLessonActionButton(
            label: '暂停',
            icon: Icons.pause_circle_outline_rounded,
            filled: false,
            onTap: onPause,
          ),
          const SizedBox(width: 10),
          _IepLessonActionButton(
            label: '结束上课',
            icon: Icons.stop_circle_rounded,
            filled: true,
            onTap: onEndLesson,
          ),
        ],
      ),
    );
  }
}

class _IepLessonBackButton extends StatelessWidget {
  const _IepLessonBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.line),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: _IepColors.text,
            size: 28,
          ),
        ),
      ),
    );
  }
}

class _IepLessonStudentCard extends StatelessWidget {
  const _IepLessonStudentCard({required this.draft});

  final _IepLessonSessionDraft draft;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      padding: const EdgeInsets.fromLTRB(6, 6, 14, 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          _IepLessonAvatar(name: draft.studentName, size: 32),
          const SizedBox(width: 10),
          Text(
            '${draft.studentName} · ${draft.ageLabel}',
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

enum _IepLessonBadgeTone { neutral, orange }

class _IepLessonMetaBadge extends StatelessWidget {
  const _IepLessonMetaBadge({
    required this.icon,
    required this.text,
    this.tone = _IepLessonBadgeTone.neutral,
  });

  final IconData icon;
  final String text;
  final _IepLessonBadgeTone tone;

  @override
  Widget build(BuildContext context) {
    final bool orange = tone == _IepLessonBadgeTone.orange;
    return Container(
      height: 34,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: orange ? const Color(0xFFFFEFE4) : const Color(0xFFFFFBF7),
        borderRadius: BorderRadius.circular(17),
        border: Border.all(
          color: orange ? const Color(0xFFF4D0B6) : _IepColors.lightLine,
        ),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            icon,
            size: 16,
            color: orange ? _IepColors.orangeDeep : _IepColors.muted,
          ),
          const SizedBox(width: 6),
          Text(
            text,
            style: TextStyle(
              color: orange ? _IepColors.orangeDeep : _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonActionButton extends StatelessWidget {
  const _IepLessonActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    this.onTap,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: onTap == null ? .58 : 1,
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(19),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(19),
          child: Container(
            height: 38,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: BoxDecoration(
              color: filled ? _IepColors.orange : Colors.white,
              borderRadius: BorderRadius.circular(19),
              border: Border.all(
                color: filled ? _IepColors.orange : _IepColors.line,
              ),
              boxShadow: filled
                  ? _iepShadow(
                      color: const Color(0x26E96F43),
                      blur: 14,
                      offset: const Offset(0, 5),
                    )
                  : const <BoxShadow>[],
            ),
            child: Row(
              children: <Widget>[
                Icon(
                  icon,
                  size: 18,
                  color: filled ? Colors.white : _IepColors.text,
                ),
                const SizedBox(width: 6),
                Text(
                  label,
                  style: TextStyle(
                    color: filled ? Colors.white : _IepColors.text,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _IepLessonTaskRail extends StatelessWidget {
  const _IepLessonTaskRail({
    required this.draft,
    required this.selectedIndex,
    required this.selectedDateIndex,
    required this.recordedCount,
    required this.completionCodes,
    required this.onTaskSelected,
  });

  final _IepLessonSessionDraft draft;
  final int selectedIndex;
  final int selectedDateIndex;
  final int recordedCount;
  final List<List<String>> completionCodes;
  final ValueChanged<int> onTaskSelected;

  @override
  Widget build(BuildContext context) {
    final List<_IepLessonTaskDraft> tasks = draft.tasks;
    final String dayLabel = draft.weekDateOptions.isEmpty
        ? draft.trainingDateLabel
        : draft.weekDateOptions[selectedDateIndex];
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.95),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          _IepLessonTaskRailHeader(
            dayLabel: dayLabel,
            recordedCount: recordedCount,
            totalCount: tasks.length,
          ),
          const SizedBox(height: 12),
          Expanded(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: List<Widget>.generate(tasks.length, (int index) {
                  final _IepLessonTaskDraft task = tasks[index];
                  final bool selected = index == selectedIndex;
                  final List<String> taskCodes = index < completionCodes.length
                      ? completionCodes[index]
                      : const <String>[];
                  final String currentCode =
                      selectedDateIndex < taskCodes.length
                          ? taskCodes[selectedDateIndex].trim()
                          : '';
                  return Padding(
                    padding: EdgeInsets.only(
                      bottom: index == tasks.length - 1 ? 0 : 6,
                    ),
                    child: _IepLessonTaskCard(
                      index: index,
                      task: task,
                      selected: selected,
                      currentCode: currentCode,
                      onTap: () => onTaskSelected(index),
                    ),
                  );
                }),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonTaskCard extends StatelessWidget {
  const _IepLessonTaskCard({
    required this.index,
    required this.task,
    required this.selected,
    required this.currentCode,
    required this.onTap,
  });

  final int index;
  final _IepLessonTaskDraft task;
  final bool selected;
  final String currentCode;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color borderColor =
        selected ? _IepColors.orange.withOpacity(.55) : const Color(0xFFF1E6DC);
    final Color fillColor =
        selected ? const Color(0xFFFFF7F0) : const Color(0xFFFFFDFC);
    final bool recorded = currentCode.isNotEmpty;
    final String stateLabel = recorded ? _lessonCodeLabel(currentCode) : '待记录';
    final Color stateColor =
        recorded ? _lessonCodeColor(currentCode) : _IepColors.muted;
    final Color leadingColor =
        recorded ? _lessonCodeColor(currentCode) : _IepColors.orangeDeep;
    final double inactiveOpacity = selected ? 1 : .74;

    return Opacity(
      opacity: inactiveOpacity,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
            decoration: BoxDecoration(
              color: fillColor,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: borderColor),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Padding(
                  padding: const EdgeInsets.only(top: 2),
                  child: Container(
                    width: 28,
                    height: 28,
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                      color: leadingColor.withOpacity(selected ? .12 : .08),
                      borderRadius: BorderRadius.circular(9),
                    ),
                    child: recorded
                        ? Text(
                            currentCode,
                            style: TextStyle(
                              color: leadingColor,
                              fontSize: 12,
                              fontWeight: FontWeight.w900,
                            ),
                          )
                        : Text(
                            '${index + 1}',
                            style: TextStyle(
                              color: leadingColor,
                              fontSize: 12,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        task.title,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _IepColors.ink,
                          fontSize: 13,
                          fontWeight: FontWeight.w900,
                          height: 1.28,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Row(
                        children: <Widget>[
                          Text(
                            '第${index + 1}项',
                            style: const TextStyle(
                              color: _IepColors.muted,
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          const Spacer(),
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 7,
                              vertical: 3,
                            ),
                            decoration: BoxDecoration(
                              color: stateColor.withOpacity(.10),
                              borderRadius: BorderRadius.circular(999),
                            ),
                            child: Text(
                              stateLabel,
                              style: TextStyle(
                                color: stateColor,
                                fontSize: 10,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _IepLessonTaskRailHeader extends StatelessWidget {
  const _IepLessonTaskRailHeader({
    required this.dayLabel,
    required this.recordedCount,
    required this.totalCount,
  });

  final String dayLabel;
  final int recordedCount;
  final int totalCount;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Row(
          children: <Widget>[
            const Text(
              '训练项目',
              style: TextStyle(
                color: _IepColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            Text(
              '$recordedCount/$totalCount',
              style: const TextStyle(
                color: _IepColors.orangeDeep,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFBF7),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.lightLine),
          ),
          child: Row(
            children: <Widget>[
              Text(
                dayLabel,
                style: const TextStyle(
                  color: _IepColors.orangeDeep,
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              Text(
                '已记录 $recordedCount 项',
                style: const TextStyle(
                  color: _IepColors.muted,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _IepLessonMainPanel extends StatelessWidget {
  const _IepLessonMainPanel({
    required this.draft,
    required this.task,
    required this.taskIndex,
    required this.selectedDateLabel,
    required this.courseName,
    required this.onEditCourseName,
    required this.hasPreviousTask,
    required this.hasNextTask,
    required this.onPreviousTask,
    required this.onNextTask,
  });

  final _IepLessonSessionDraft draft;
  final _IepLessonTaskDraft? task;
  final int taskIndex;
  final String selectedDateLabel;
  final String courseName;
  final VoidCallback onEditCourseName;
  final bool hasPreviousTask;
  final bool hasNextTask;
  final VoidCallback? onPreviousTask;
  final VoidCallback? onNextTask;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
      child: task == null
          ? const Center(
              child: Text(
                '暂无训练任务',
                style: TextStyle(
                  color: _IepColors.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w700,
                ),
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _IepLessonHeroCard(
                  eyebrow: '当前训练',
                  title: task!.title,
                  content: task!.subtitle,
                  indexLabel: '第${taskIndex + 1}项',
                ),
                const SizedBox(height: 10),
                _IepLessonMetaStrip(
                  teacherName: draft.teacherName,
                  courseName: courseName,
                  onEditCourseName: onEditCourseName,
                  selectedDateLabel: selectedDateLabel,
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: SingleChildScrollView(
                    physics: const BouncingScrollPhysics(),
                    child: _IepLessonPreparationCard(
                      preparation: draft.preparation,
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _IepLessonNavButton(
                        label: '上一项',
                        icon: Icons.west_rounded,
                        enabled: hasPreviousTask,
                        primary: false,
                        onTap: onPreviousTask,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _IepLessonNavButton(
                        label: '下一项',
                        icon: Icons.east_rounded,
                        enabled: hasNextTask,
                        primary: true,
                        onTap: onNextTask,
                      ),
                    ),
                  ],
                ),
              ],
            ),
    );
  }
}

class _IepLessonMetaStrip extends StatelessWidget {
  const _IepLessonMetaStrip({
    required this.teacherName,
    required this.courseName,
    required this.onEditCourseName,
    required this.selectedDateLabel,
  });

  final String teacherName;
  final String courseName;
  final VoidCallback onEditCourseName;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBF7),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: _IepLessonMetaItem(
              label: '任教老师',
              value: teacherName,
            ),
          ),
          const _IepLessonMetaDivider(),
          Expanded(
            child: _IepLessonEditableMetaItem(
              label: '课程名称',
              value: courseName,
              onTap: onEditCourseName,
            ),
          ),
          const _IepLessonMetaDivider(),
          Expanded(
            child: _IepLessonMetaItem(
              label: '训练日期',
              value: selectedDateLabel,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonMetaItem extends StatelessWidget {
  const _IepLessonMetaItem({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 5),
          Text(
            value.trim().isEmpty ? '-' : value.trim(),
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
              height: 1.35,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonEditableMetaItem extends StatelessWidget {
  const _IepLessonEditableMetaItem({
    required this.label,
    required this.value,
    required this.onTap,
  });

  final String label;
  final String value;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                label,
                style: const TextStyle(
                  color: _IepColors.muted,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 5),
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      value.trim().isEmpty ? '-' : value.trim(),
                      style: const TextStyle(
                        color: _IepColors.ink,
                        fontSize: 13,
                        fontWeight: FontWeight.w900,
                        height: 1.35,
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  const Icon(
                    Icons.edit_outlined,
                    size: 14,
                    color: _IepColors.muted,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _IepLessonMetaDivider extends StatelessWidget {
  const _IepLessonMetaDivider();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 1,
      height: 34,
      color: _IepColors.lightLine,
    );
  }
}

class _IepLessonCourseNameDialog extends StatefulWidget {
  const _IepLessonCourseNameDialog({required this.initialValue});

  final String initialValue;

  @override
  State<_IepLessonCourseNameDialog> createState() =>
      _IepLessonCourseNameDialogState();
}

class _IepLessonCourseNameDialogState
    extends State<_IepLessonCourseNameDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(
      text: widget.initialValue.trim().isEmpty ? '康复教学' : widget.initialValue,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      alignment: const Alignment(0, -0.18),
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 420,
        padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Text(
                  '编辑课程名称',
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 18,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 18),
            const Text(
              '课程名称',
              style: TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _controller,
              autofocus: true,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 14,
                fontWeight: FontWeight.w700,
                height: 1.35,
              ),
              decoration: InputDecoration(
                isDense: true,
                filled: true,
                fillColor: Colors.white,
                hintText: '请输入课程名称',
                hintStyle: const TextStyle(
                  color: _IepColors.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w700,
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 11,
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: const BorderSide(color: _IepColors.line),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: const BorderSide(
                    color: _IepColors.orange,
                    width: 1.2,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Row(
              children: <Widget>[
                Expanded(
                  child: _IepLessonDialogButton(
                    label: '取消',
                    primary: false,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _IepLessonDialogButton(
                    label: '确定',
                    primary: true,
                    onTap: () => Navigator.of(context).pop(
                      _controller.text.trim().isEmpty
                          ? '康复教学'
                          : _controller.text.trim(),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _IepLessonMissingRecordDialog extends StatefulWidget {
  const _IepLessonMissingRecordDialog({
    required this.selectedDateLabel,
    required this.missingTasks,
  });

  final String selectedDateLabel;
  final List<_IepLessonTaskDraft> missingTasks;

  @override
  State<_IepLessonMissingRecordDialog> createState() =>
      _IepLessonMissingRecordDialogState();
}

class _IepLessonMissingRecordDialogState
    extends State<_IepLessonMissingRecordDialog> {
  String _selectedCode = '';

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      alignment: const Alignment(0, -0.08),
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 560,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 20),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        '还有 ${widget.missingTasks.length} 项未记录',
                        style: const TextStyle(
                          color: _IepColors.ink,
                          fontSize: 18,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '${widget.selectedDateLabel} 仍有训练项目未填写完成情况，选择一个结果后可一键补记并结束上课。',
                        style: const TextStyle(
                          color: _IepColors.text,
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          height: 1.45,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 12),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFBF7),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Container(
                    width: 28,
                    height: 28,
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFF1E6),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: const Icon(
                      Icons.info_outline_rounded,
                      size: 16,
                      color: _IepColors.orangeDeep,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      '当前还有 ${widget.missingTasks.length} 项未填写完成情况。如果本次训练结果一致，可直接选择一个记录结果，一键补记后结束上课。',
                      style: const TextStyle(
                        color: _IepColors.text,
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        height: 1.45,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '一键补记为',
              style: TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 8),
            _IepLessonCodeGrid(
              codes: _lessonRecordCodeOptions,
              currentCode: _selectedCode,
              enabled: true,
              onCodeSelected: (String code) {
                setState(() {
                  _selectedCode = code;
                });
              },
            ),
            const SizedBox(height: 18),
            Row(
              children: <Widget>[
                Expanded(
                  child: _IepLessonDialogButton(
                    label: '继续记录',
                    primary: false,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _IepLessonDialogButton(
                    label: '补记并结束',
                    primary: true,
                    onTap: _selectedCode.isEmpty
                        ? null
                        : () => Navigator.of(context).pop(_selectedCode),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _IepLessonDialogButton extends StatelessWidget {
  const _IepLessonDialogButton({
    required this.label,
    required this.primary,
    required this.onTap,
  });

  final String label;
  final bool primary;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 44,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: !enabled
                ? const Color(0xFFF3F4F6)
                : (primary ? _IepColors.orange : Colors.white),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: !enabled
                  ? const Color(0xFFE0E3E8)
                  : (primary ? _IepColors.orange : _IepColors.line),
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: !enabled
                  ? const Color(0xFFB4BAC4)
                  : (primary ? Colors.white : _IepColors.text),
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _IepLessonHeroCard extends StatelessWidget {
  const _IepLessonHeroCard({
    required this.eyebrow,
    required this.title,
    required this.content,
    required this.indexLabel,
  });

  final String eyebrow;
  final String title;
  final String content;
  final String indexLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(18, 18, 18, 18),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: <Color>[Color(0xFFFFFCF8), Color(0xFFFFF4EA)],
        ),
        border: Border.all(color: const Color(0xFFF2D9C6)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(.72),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(color: const Color(0xFFF0D7C5)),
                ),
                child: Text(
                  eyebrow,
                  style: const TextStyle(
                    color: _IepColors.orangeDeep,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const Spacer(),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF8F1),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: const Color(0xFFF0D7C5)),
                ),
                child: Text(
                  indexLabel,
                  style: const TextStyle(
                    color: _IepColors.orangeDeep,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Text(
            title.trim().isEmpty ? '-' : title.trim(),
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 26,
              fontWeight: FontWeight.w900,
              height: 1.22,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            content.trim().isEmpty ? '-' : content.trim(),
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 16,
              fontWeight: FontWeight.w800,
              height: 1.65,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonPreparationCard extends StatelessWidget {
  const _IepLessonPreparationCard({
    required this.preparation,
  });

  final String preparation;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _IepColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '训练前准备',
            style: TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 12),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFAF6),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: _IepColors.lightLine),
            ),
            child: Text(
              preparation.trim().isEmpty ? '-' : preparation.trim(),
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 14,
                fontWeight: FontWeight.w800,
                height: 1.72,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonNavButton extends StatelessWidget {
  const _IepLessonNavButton({
    required this.label,
    required this.icon,
    required this.enabled,
    required this.primary,
    this.onTap,
  });

  final String label;
  final IconData icon;
  final bool enabled;
  final bool primary;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final Color background = !enabled
        ? const Color(0xFFF6EEE7)
        : primary
            ? _IepColors.orange
            : Colors.white;
    final Color borderColor = !enabled
        ? const Color(0xFFEADCCF)
        : primary
            ? _IepColors.orange
            : _IepColors.line;
    final Color textColor = !enabled
        ? _IepColors.muted
        : primary
            ? Colors.white
            : _IepColors.text;
    return Opacity(
      opacity: enabled ? 1 : .72,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: enabled ? onTap : null,
          borderRadius: BorderRadius.circular(16),
          child: Container(
            height: 48,
            decoration: BoxDecoration(
              color: background,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: borderColor),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: 18, color: textColor),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: TextStyle(
                    color: textColor,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _IepLessonRecordPanel extends StatelessWidget {
  const _IepLessonRecordPanel({
    required this.draft,
    required this.task,
    required this.currentCodes,
    required this.weekDateOptions,
    required this.selectedDateIndex,
    required this.onCodeSelected,
    required this.selectedDateLabel,
  });

  final _IepLessonSessionDraft draft;
  final _IepLessonTaskDraft? task;
  final List<String> currentCodes;
  final List<String> weekDateOptions;
  final int selectedDateIndex;
  final ValueChanged<String> onCodeSelected;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    final List<String> dateOptions = weekDateOptions.isEmpty
        ? draft.completionColumnLabels
        : weekDateOptions;
    final int safeDateIndex = dateOptions.isEmpty
        ? 0
        : selectedDateIndex.clamp(0, math.max(0, dateOptions.length - 1));
    final String currentCode = safeDateIndex < currentCodes.length
        ? currentCodes[safeDateIndex].trim()
        : '';
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.95),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '记录区域',
            style: TextStyle(
              color: _IepColors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 14),
          Expanded(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _IepLessonRecordSection(
                    title: '训练日期',
                    child: _IepLessonDateChip(
                      label: selectedDateLabel,
                      selected: true,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _IepLessonRecordSection(
                    title: '当前记录',
                    child: Container(
                      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: _IepColors.lightLine),
                      ),
                      child: Row(
                        children: <Widget>[
                          Container(
                            width: 42,
                            height: 42,
                            alignment: Alignment.center,
                            decoration: BoxDecoration(
                              color: _lessonCodeColor(currentCode)
                                  .withOpacity(.12),
                              shape: BoxShape.circle,
                            ),
                            child: Text(
                              currentCode.isEmpty ? '·' : currentCode,
                              style: TextStyle(
                                color: _lessonCodeColor(currentCode),
                                fontSize: currentCode.isEmpty ? 22 : 20,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: <Widget>[
                                Text(
                                  currentCode.isEmpty
                                      ? '待记录'
                                      : _lessonCodeLabel(currentCode),
                                  style: TextStyle(
                                    color: _lessonCodeColor(currentCode),
                                    fontSize: 16,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  currentCode.isEmpty
                                      ? '请选择一项数据记录'
                                      : '当前记录为 $currentCode',
                                  style: const TextStyle(
                                    color: _IepColors.text,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    height: 1.35,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  _IepLessonRecordSection(
                    title: '数据记录',
                    child: _IepLessonCodeGrid(
                      codes: _lessonRecordCodeOptions,
                      currentCode: currentCode,
                      enabled: task != null,
                      onCodeSelected: onCodeSelected,
                    ),
                  ),
                ],
              ),
            ),
          ),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFBF7),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: _IepColors.lightLine),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                const Icon(
                  Icons.lightbulb_outline_rounded,
                  size: 18,
                  color: _IepColors.orangeDeep,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '结束上课后，可将本次记录回写到 ${draft.weekLabel} 中 $selectedDateLabel 的完成情况。',
                    style: const TextStyle(
                      color: _IepColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      height: 1.45,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonRecordSection extends StatelessWidget {
  const _IepLessonRecordSection({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(
            color: _IepColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 8),
        child,
      ],
    );
  }
}

class _IepLessonDateChip extends StatelessWidget {
  const _IepLessonDateChip({
    required this.label,
    required this.selected,
  });

  final String label;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFF1E6) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: selected ? _IepColors.orange : _IepColors.line,
        ),
      ),
      child: Text(
        label,
        textAlign: TextAlign.center,
        style: TextStyle(
          color: selected ? _IepColors.orangeDeep : _IepColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w800,
          height: 1.35,
        ),
      ),
    );
  }
}

class _IepLessonCodeGrid extends StatelessWidget {
  const _IepLessonCodeGrid({
    required this.codes,
    required this.currentCode,
    required this.enabled,
    required this.onCodeSelected,
  });

  final List<String> codes;
  final String currentCode;
  final bool enabled;
  final ValueChanged<String> onCodeSelected;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double itemWidth = (constraints.maxWidth - 12) / 2;
        return Wrap(
          spacing: 12,
          runSpacing: 12,
          children: List<Widget>.generate(codes.length, (int index) {
            final String code = codes[index];
            final bool isLast = index == codes.length - 1;
            return _IepLessonCodeCard(
              code: code,
              label: _lessonCodeLabel(code),
              width: isLast ? constraints.maxWidth : itemWidth,
              selected: currentCode == code,
              enabled: enabled,
              onTap: () => onCodeSelected(code),
            );
          }),
        );
      },
    );
  }
}

class _IepLessonCodeCard extends StatelessWidget {
  const _IepLessonCodeCard({
    required this.code,
    required this.label,
    required this.width,
    required this.selected,
    required this.enabled,
    required this.onTap,
  });

  final String code;
  final String label;
  final double width;
  final bool selected;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: enabled ? 1 : .45,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: enabled ? onTap : null,
          borderRadius: BorderRadius.circular(16),
          child: Container(
            width: width,
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
            decoration: BoxDecoration(
              color: selected
                  ? _lessonCodeColor(code).withOpacity(.12)
                  : Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: selected ? _lessonCodeColor(code) : _IepColors.line,
                width: selected ? 1.4 : 1,
              ),
            ),
            child: Row(
              children: <Widget>[
                Container(
                  width: 36,
                  height: 36,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: _lessonCodeColor(code).withOpacity(.12),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    code,
                    style: TextStyle(
                      color: _lessonCodeColor(code),
                      fontSize: 18,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    label,
                    maxLines: 2,
                    overflow: TextOverflow.visible,
                    style: TextStyle(
                      color:
                          selected ? _lessonCodeColor(code) : _IepColors.text,
                      fontSize: 14,
                      fontWeight: FontWeight.w800,
                      height: 1.2,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

const List<String> _lessonRecordCodeOptions = <String>[
  '√',
  '✗',
  'S',
  'G',
  'M',
  'V',
  'P',
];

String _lessonCodeLabel(String code) {
  switch (code) {
    case '√':
      return '独立完成';
    case '✗':
      return '未完成';
    case 'S':
      return '语言提示';
    case 'G':
      return '手势提示';
    case 'M':
      return '示范辅助';
    case 'V':
      return '视觉提示';
    case 'P':
      return '肢体辅助';
    default:
      return '待记录';
  }
}

Color _lessonCodeColor(String code) {
  switch (code) {
    case '√':
      return _IepColors.green;
    case '✗':
      return const Color(0xFFD2573F);
    case 'S':
      return const Color(0xFFE0A339);
    case 'G':
      return const Color(0xFFCE7F3B);
    case 'M':
      return const Color(0xFFB77BCE);
    case 'V':
      return const Color(0xFF5E98C9);
    case 'P':
      return const Color(0xFF8D6E63);
    default:
      return _IepColors.muted;
  }
}

class _IepLessonAvatar extends StatelessWidget {
  const _IepLessonAvatar({
    required this.name,
    required this.size,
  });

  final String name;
  final double size;

  @override
  Widget build(BuildContext context) {
    final int seed = name.runes.fold<int>(0, (int sum, int item) => sum + item);
    final List<List<Color>> palettes = <List<Color>>[
      const <Color>[Color(0xFFFFD8C2), Color(0xFFFFA36F)],
      const <Color>[Color(0xFFFFE0B7), Color(0xFFFFB067)],
      const <Color>[Color(0xFFFFD2C8), Color(0xFFFF8E75)],
      const <Color>[Color(0xFFFFE5BF), Color(0xFFFFB74E)],
    ];
    final List<Color> colors = palettes[seed % palettes.length];
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: colors,
        ),
      ),
      child: Icon(
        Icons.face_rounded,
        size: size * .72,
        color: const Color(0xFF6B4336),
      ),
    );
  }
}
