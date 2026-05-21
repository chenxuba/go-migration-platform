import 'dart:async';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_age_formatter.dart';
import 'assessment_scale_client.dart';
import 'home_client.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';
import 'vbmapp_assessment_client.dart';

part 'vbmapp_assessment_data.dart';
part 'vbmapp_assessment_chrome.dart';
part 'vbmapp_assessment_navigation.dart';
part 'vbmapp_assessment_workspace.dart';
part 'vbmapp_assessment_mand_panels.dart';

class VbmappAssessmentLaunchArgs {
  const VbmappAssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = 'VB-MAPP语言行为里程碑评估及安置计划',
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

class VbmappAssessmentPage extends StatefulWidget {
  const VbmappAssessmentPage({
    required this.onBack,
    this.args = const VbmappAssessmentLaunchArgs(),
    this.client = const ApiVbmappAssessmentClient(),
    this.homeClient = const ApiHomeClient(),
    super.key,
  });

  final VoidCallback onBack;
  final VbmappAssessmentLaunchArgs args;
  final VbmappAssessmentClient client;
  final HomeClient homeClient;

  @override
  State<VbmappAssessmentPage> createState() => _VbmappAssessmentPageState();
}

class _VbmappAssessmentPageState extends State<VbmappAssessmentPage>
    with WidgetsBindingObserver {
  static const String _authTokenStorageKey = 'auth_token';
  static const String _scaleVersion = 'VBMAPP_CN_2ND_DRAFT_2026_05';
  static const int _totalItemCount = 212;
  static const String _sharedTimedMandStorageKey = '__MAND_TIMED_SHARED__';
  static const Set<String> _sharedTimedMandItemCodes = <String>{
    'MAND_04M',
    'MAND_08M',
    'MAND_09M',
  };

  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  final Map<String, double> _milestoneScores = <String, double>{};
  final Map<String, int> _barrierScores = <String, int>{};
  final Map<String, int> _transitionScores = <String, int>{};
  final Map<String, VbmappItemResponseSchema> _itemSchemas =
      <String, VbmappItemResponseSchema>{};
  final Map<String, VbmappMaterialProfile> _materialProfiles =
      <String, VbmappMaterialProfile>{};
  final Map<String, VbmappMaterialProfile> _itemMaterialProfiles =
      <String, VbmappMaterialProfile>{};
  final Map<String, List<_VbmappMandEvent>> _mandEventsByItem =
      <String, List<_VbmappMandEvent>>{};
  final Map<String, _VbmappObservationTimerState> _mandObservationByItem =
      <String, _VbmappObservationTimerState>{};
  Timer? _observationTicker;

  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _examinerName = '';
  String _assessmentDate = '';
  String _selectedModuleCode = _vbmappModules.first.code;
  String _autoSaveText = '等待作答';
  int _draftId = 0;
  int _studentId = 0;
  int _selectedItemIndex = 0;
  bool _loading = true;
  bool _saving = false;
  bool _submitting = false;
  bool _autoNext = false;
  bool _keyboardVisible = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _draftId = widget.args.draftId;
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName.trim();
    _studentAge = widget.args.studentAge.trim();
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isEmpty
        ? _todayIsoDate()
        : _dateOnlyText(widget.args.assessmentDate);
    _examinerName = widget.args.examinerName.trim();
    unawaited(_initialize());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _observationTicker?.cancel();
    _messageController.dispose();
    super.dispose();
  }

  @override
  void didChangeMetrics() {
    super.didChangeMetrics();
    final bool keyboardVisible = WidgetsBinding
        .instance.platformDispatcher.views
        .any((ui.FlutterView view) => view.viewInsets.bottom > 0);
    if (_keyboardVisible && !keyboardVisible) {
      _dismissEditingFocus();
    }
    _keyboardVisible = keyboardVisible;
  }

  Future<void> _initialize() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _autoSaveText = '请先登录';
      });
      _showMessage('请先登录后再进行VB-MAPP测评', tone: PadMessageTone.error);
      return;
    }
    HomeSession session = HomeSession.fallback;
    try {
      session = await widget.homeClient.fetchCurrentSession(token);
    } on Object {
      session = HomeSession.fallback;
    }
    VbmappAssessmentSchema? smartSchema;
    VbmappMaterialCatalog? materialCatalog;
    try {
      smartSchema = await widget.client.fetchAssessmentSchema(token);
    } on Object catch (error) {
      if (mounted) {
        _showMessage('VB-MAPP智能题库载入失败，先使用基础题库：$error');
      }
    }
    try {
      materialCatalog = await widget.client.fetchMaterialCatalog(
        token,
        moduleCode: 'milestones',
      );
    } on Object catch (error) {
      if (mounted) {
        _showMessage('VB-MAPP素材目录载入失败，先使用基础素材：$error');
      }
    }
    VbmappDraftDetail? launchDraft;
    if (_draftId > 0) {
      try {
        launchDraft = await widget.client.fetchDraftDetail(token, _draftId);
      } on Object catch (error) {
        if (mounted) {
          _showMessage('VB-MAPP草稿载入失败：$error', tone: PadMessageTone.error);
        }
      }
    }
    if (!mounted) {
      return;
    }
    setState(() {
      _token = token;
      if (smartSchema != null) {
        _itemSchemas
          ..clear()
          ..addAll(smartSchema.itemSchemas);
        _materialProfiles
          ..clear()
          ..addAll(smartSchema.materialProfiles);
      }
      _itemMaterialProfiles
        ..clear()
        ..addAll(_itemMaterialProfileMapFromCatalog(materialCatalog));
      if (launchDraft != null) {
        _applyDraftDetail(launchDraft);
      }
      if (_examinerName.isEmpty) {
        _examinerName = _sessionExaminerName(session);
      }
      _autoSaveText = _draftId > 0 ? '草稿已载入' : '等待作答';
      _loading = false;
    });
    _syncObservationTicker();
  }

  void _applyDraftDetail(VbmappDraftDetail detail) {
    _draftId = detail.id > 0 ? detail.id : _draftId;
    _studentId = detail.studentId > 0 ? detail.studentId : _studentId;
    if (detail.studentName.trim().isNotEmpty) {
      _studentName = detail.studentName.trim();
    }
    if (detail.birthDate.trim().isNotEmpty) {
      _birthDate = _dateOnlyText(detail.birthDate);
    }
    if (detail.assessmentDate.trim().isNotEmpty) {
      _assessmentDate = _dateOnlyText(detail.assessmentDate);
    }
    if (detail.examinerName.trim().isNotEmpty) {
      _examinerName = detail.examinerName.trim();
    }
    _milestoneScores
      ..clear()
      ..addAll(detail.milestoneScores);
    _barrierScores
      ..clear()
      ..addAll(detail.barrierScores);
    _transitionScores
      ..clear()
      ..addAll(detail.transitionScores);
    _restoreMandEvents(detail.itemResponses);
    _restoreMandObservations(detail.itemResponses);
    if (_hasSharedTimedMandEvidence()) {
      _syncSharedTimedMandScores();
    } else {
      _clearBuggedSharedTimedMandScores();
    }
    final _VbmappItem? firstMissing = _firstMissingItem();
    if (firstMissing != null) {
      _selectedModuleCode = firstMissing.moduleCode;
      final int missingIndex = _itemsForModule(firstMissing.moduleCode)
          .indexWhere(
              (_VbmappItem item) => item.itemCode == firstMissing.itemCode);
      _selectedItemIndex = missingIndex < 0 ? 0 : missingIndex;
    }
  }

  List<_VbmappItem> get _selectedItems {
    return _itemsForModule(_selectedModuleCode);
  }

  _VbmappItem get _selectedItem {
    final List<_VbmappItem> items = _selectedItems;
    return items[_selectedItemIndex.clamp(0, items.length - 1)];
  }

  int get _answeredCount {
    return _milestoneScores.length +
        _barrierScores.length +
        _transitionScores.length;
  }

  double get _progressPercent {
    return _answeredCount / _totalItemCount;
  }

  _VbmappScoreSnapshot get _scoreSnapshot {
    return _VbmappScoreSnapshot(
      milestoneTotal: _milestoneScores.values.fold<double>(
        0,
        (double total, double score) => total + score,
      ),
      milestoneMax: _milestoneItems.length,
      barrierTotal: _barrierScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      barrierMax: _barrierItems.length * 4,
      transitionTotal: _transitionScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      transitionMax: _transitionItems.length * 5,
      milestoneDomains: _milestoneDomainSummaries,
    );
  }

  List<_VbmappDomainScoreSummary> get _milestoneDomainSummaries {
    final Map<String, List<_VbmappItem>> groupedItems =
        <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in _milestoneItems) {
      groupedItems
          .putIfAbsent(item.domainName, () => <_VbmappItem>[])
          .add(item);
    }
    return groupedItems.entries
        .map((MapEntry<String, List<_VbmappItem>> entry) {
      final List<_VbmappItem> items = entry.value;
      final int answered = items
          .where(
              (_VbmappItem item) => _milestoneScores.containsKey(item.itemCode))
          .length;
      final double score = items.fold<double>(
        0,
        (double total, _VbmappItem item) =>
            total + (_milestoneScores[item.itemCode] ?? 0),
      );
      return _VbmappDomainScoreSummary(
        name: entry.key,
        score: score,
        maxScore: items.length,
        answered: answered,
        total: items.length,
      );
    }).toList(growable: false);
  }

  String get _studentAgeText {
    final String fallback = _studentAge.trim().isEmpty ? '未知' : _studentAge;
    return formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
      fallback: fallback,
    );
  }

  void _selectModule(String code) {
    if (_selectedModuleCode == code) {
      return;
    }
    setState(() {
      _selectedModuleCode = code;
      _selectedItemIndex = 0;
    });
  }

  void _selectScore(num score) {
    final _VbmappItem item = _selectedItem;
    setState(() {
      switch (item.moduleCode) {
        case 'milestones':
          _milestoneScores[item.itemCode] = score.toDouble();
          break;
        case 'barriers':
          _barrierScores[item.itemCode] = score.toInt();
          break;
        case 'transition':
          _transitionScores[item.itemCode] = score.toInt();
          break;
      }
      _autoSaveText = '已记录';
    });
    if (_autoNext) {
      Future<void>.delayed(const Duration(milliseconds: 220), () {
        if (mounted) {
          _goNext();
        }
      });
    }
  }

  Future<void> _openMandEventDialog(_VbmappItem item) async {
    final _VbmappMandEvent? event = await showDialog<_VbmappMandEvent>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappMandEventDialog(
            generalizationMode: item.itemCode == 'MAND_03M',
          ),
        );
      },
    );
    if (event == null) {
      return;
    }
    await _addMandEvent(item, event);
  }

  Future<void> _addMandEvent(_VbmappItem item, _VbmappMandEvent event) async {
    final DateTime now = DateTime.now();
    final List<_VbmappMandEvent> events =
        List<_VbmappMandEvent>.from(_mandStoredEventsFor(item))
          ..add(
            event.recordedAtIso.trim().isEmpty
                ? event.copyWith(
                    recordedAtIso: now.toIso8601String(),
                    sourceItemCode: item.itemCode,
                  )
                : event,
          );
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
    );
    setState(() {
      _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] = events;
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_sharedTimedMandItemCodes.contains(item.itemCode)) {
        _syncSharedTimedMandScores();
      }
      _autoSaveText = '保存中...';
    });
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
    );
  }

  Future<void> _deleteMandEvent(_VbmappItem item, int index) async {
    final List<_VbmappMandEvent> events =
        List<_VbmappMandEvent>.from(_mandStoredEventsFor(item));
    if (index < 0 || index >= events.length) {
      return;
    }
    events.removeAt(index);
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
    );
    setState(() {
      if (events.isEmpty) {
        _mandEventsByItem.remove(_mandStorageKeyFor(item.itemCode));
      } else {
        _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] = events;
      }
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_sharedTimedMandItemCodes.contains(item.itemCode)) {
        _syncSharedTimedMandScores();
      }
      _autoSaveText = '保存中...';
    });
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
    );
  }

  Future<void> _updateMandObservation(
    _VbmappItem item,
    _VbmappObservationTimerState observation,
  ) async {
    final List<_VbmappMandEvent> events = _mandStoredEventsFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
    );
    setState(() {
      _mandObservationByItem[_mandStorageKeyFor(item.itemCode)] = observation;
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_sharedTimedMandItemCodes.contains(item.itemCode)) {
        _syncSharedTimedMandScores();
      }
      _autoSaveText = '保存中...';
    });
    _syncObservationTicker();
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
    );
  }

  Future<void> _saveMandEvidence(
    _VbmappItem item,
    List<_VbmappMandEvent> events,
    double suggestedScore, {
    _VbmappObservationTimerState? observation,
  }) async {
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存证据', tone: PadMessageTone.error);
      return;
    }
    final int draftId = await _saveDraft(silent: true);
    if (draftId <= 0) {
      return;
    }
    final int qualifiedCount = _qualifiedMandCountForItem(
      item,
      events,
      observation: observation,
    );
    final _VbmappObservationTimerState? timerState = observation;
    final int actualObservationSeconds =
        timerState?.elapsedSecondsAt(DateTime.now()) ?? 0;
    final int effectiveObservationSeconds = _effectiveObservationSecondsForItem(
      item,
      timerState,
    );
    final int multiWordCount = item.itemCode == 'MAND_08M'
        ? _mandPhraseQualifiedCountForItem(
            item,
            events,
            observation: observation,
          )
        : 0;
    try {
      final VbmappDraftDetail detail = await widget.client.saveDraftItem(
        _token,
        <String, dynamic>{
          'draftId': draftId,
          'moduleCode': item.moduleCode,
          'itemCode': item.itemCode,
          'score': suggestedScore,
          'suggestedScore': suggestedScore,
          'teacherConfirmed': false,
          'recordStatus': 'auto_suggested',
          'evidence': <String, dynamic>{
            'mandEvents': events
                .map((_VbmappMandEvent event) => event.toJson())
                .toList(growable: false),
            'qualifiedCount': qualifiedCount,
            'uniqueTargetCount': qualifiedCount,
            if (timerState != null) 'timer': timerState.toJson(),
            if (timerState != null)
              'actualObservationMinutes':
                  actualObservationSeconds / Duration.secondsPerMinute,
            if (timerState != null)
              'actualObservationSeconds': actualObservationSeconds,
            if (timerState != null)
              'effectiveObservationSeconds': effectiveObservationSeconds,
            if (timerState != null)
              'effectiveObservationMinutes':
                  effectiveObservationSeconds / Duration.secondsPerMinute,
            if (item.itemCode == 'MAND_08M')
              'multiWordQualifiedCount': multiWordCount,
            if (item.itemCode == 'MAND_03M')
              'generalizationCounts': _mandGeneralizationCounts(events),
            'scoreBasis': item.itemCode == 'MAND_03M'
                ? '系统按互动对象、环境、不同例子的泛化记录建议${_formatScore(suggestedScore)}分，老师可在下方评分区覆盖。'
                : _sharedTimedMandItemCodes.contains(item.itemCode)
                    ? _mandTimedScoreBasisText(
                        item,
                        suggestedScore,
                        qualifiedCount,
                        actualObservationSeconds,
                        multiWordCount,
                      )
                    : '系统按有效要求数量建议${_formatScore(suggestedScore)}分，老师可在下方评分区覆盖。',
          },
        },
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id > 0 ? detail.id : _draftId;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _autoSaveText = '保存失败');
      _showMessage('VB-MAPP单题证据保存失败：$error', tone: PadMessageTone.error);
    }
  }

  void _restoreMandEvents(
    Map<String, Map<String, Map<String, dynamic>>> itemResponses,
  ) {
    _mandEventsByItem.clear();
    final Map<String, Map<String, dynamic>> milestoneResponses =
        itemResponses['milestones'] ?? const <String, Map<String, dynamic>>{};
    milestoneResponses.forEach((String itemCode, Map<String, dynamic> value) {
      final Object? evidenceRaw = value['evidence'];
      if (evidenceRaw is! Map) {
        return;
      }
      final Object? eventsRaw = evidenceRaw['mandEvents'];
      if (eventsRaw is! List) {
        return;
      }
      final List<_VbmappMandEvent> events = eventsRaw
          .map((Object? raw) => _VbmappMandEvent.fromJson(_dynamicMap(raw)))
          .where((_VbmappMandEvent event) => event.isNotEmpty)
          .toList(growable: false);
      if (events.isNotEmpty) {
        _mandEventsByItem[_mandStorageKeyFor(itemCode)] = events;
      }
    });
  }

  void _restoreMandObservations(
    Map<String, Map<String, Map<String, dynamic>>> itemResponses,
  ) {
    _mandObservationByItem.clear();
    final Map<String, Map<String, dynamic>> milestoneResponses =
        itemResponses['milestones'] ?? const <String, Map<String, dynamic>>{};
    milestoneResponses.forEach((String itemCode, Map<String, dynamic> value) {
      final Object? evidenceRaw = value['evidence'];
      if (evidenceRaw is! Map) {
        return;
      }
      final Object? timerRaw = evidenceRaw['timer'];
      if (timerRaw is! Map) {
        return;
      }
      final _VbmappObservationTimerState observation =
          _VbmappObservationTimerState.fromJson(_dynamicMap(timerRaw));
      if (!observation.isEmpty) {
        _mandObservationByItem[_mandStorageKeyFor(itemCode)] = observation;
      }
    });
  }

  num? _scoreFor(_VbmappItem item) {
    switch (item.moduleCode) {
      case 'milestones':
        return _milestoneScores[item.itemCode];
      case 'barriers':
        return _barrierScores[item.itemCode];
      case 'transition':
        return _transitionScores[item.itemCode];
    }
    return null;
  }

  VbmappItemResponseSchema? _schemaFor(_VbmappItem item) {
    return _itemSchemas[_schemaKey(item.moduleCode, item.itemCode)];
  }

  VbmappMaterialProfile? _materialProfileFor(
    _VbmappItem item,
    VbmappItemResponseSchema? schema,
  ) {
    final VbmappMaterialProfile? itemProfile =
        _itemMaterialProfiles[item.itemCode];
    if (itemProfile != null) {
      return itemProfile;
    }
    if (schema == null || schema.materialProfileId.isEmpty) {
      return null;
    }
    return _materialProfiles[schema.materialProfileId];
  }

  List<_VbmappMandEvent> _mandEventsFor(_VbmappItem item) {
    return _mandStoredEventsFor(item);
  }

  _VbmappObservationTimerState? _mandObservationFor(_VbmappItem item) {
    return _mandObservationByItem[_mandStorageKeyFor(item.itemCode)];
  }

  _VbmappItem? _activeMandObservationItem() {
    final _VbmappObservationTimerState? sharedObservation =
        _mandObservationByItem[_sharedTimedMandStorageKey];
    if (sharedObservation != null &&
        sharedObservation.hasStarted &&
        !sharedObservation.ended) {
      return _milestoneItems.firstWhere(
        (_VbmappItem item) => item.itemCode == 'MAND_04M',
        orElse: () => _selectedItem,
      );
    }
    for (final _VbmappItem item in _milestoneItems) {
      final _VbmappObservationTimerState? observation =
          _mandObservationByItem[item.itemCode];
      if (observation != null && observation.hasStarted && !observation.ended) {
        return item;
      }
    }
    return null;
  }

  int _activeMandObservationQualifiedCount(_VbmappItem item) {
    return _qualifiedMandCountForItem(
      item,
      _mandStoredEventsFor(item),
      observation: _mandObservationFor(item),
    );
  }

  List<_VbmappMandEvent> _mandStoredEventsFor(_VbmappItem item) {
    return _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] ??
        const <_VbmappMandEvent>[];
  }

  bool _hasSharedTimedMandEvidence() {
    final List<_VbmappMandEvent> sharedEvents =
        _mandEventsByItem[_sharedTimedMandStorageKey] ??
            const <_VbmappMandEvent>[];
    if (sharedEvents.isNotEmpty) {
      return true;
    }
    final _VbmappObservationTimerState? sharedObservation =
        _mandObservationByItem[_sharedTimedMandStorageKey];
    if (sharedObservation == null) {
      return false;
    }
    return sharedObservation.hasStarted ||
        sharedObservation.accumulatedSeconds > 0 ||
        sharedObservation.ended;
  }

  void _clearBuggedSharedTimedMandScores() {
    const Set<String> codes = _sharedTimedMandItemCodes;
    final bool allZero = codes.every(
      (String code) => (_milestoneScores[code] ?? -1) == 0,
    );
    if (!allZero) {
      return;
    }
    for (final String code in codes) {
      _milestoneScores.remove(code);
    }
  }

  String _mandStorageKeyFor(String itemCode) {
    return _sharedTimedMandItemCodes.contains(itemCode)
        ? _sharedTimedMandStorageKey
        : itemCode;
  }

  void _syncSharedTimedMandScores() {
    if (!_hasSharedTimedMandEvidence()) {
      return;
    }
    for (final _VbmappItem item in _milestoneItems) {
      if (!_sharedTimedMandItemCodes.contains(item.itemCode)) {
        continue;
      }
      final List<_VbmappMandEvent> events = _mandStoredEventsFor(item);
      final _VbmappObservationTimerState? observation =
          _mandObservationFor(item);
      _milestoneScores[item.itemCode] = _suggestMandScore(
        item,
        events,
        observation: observation,
      );
    }
  }

  void _syncObservationTicker() {
    _observationTicker?.cancel();
    _observationTicker = null;
    final _VbmappItem? activeItem = _activeMandObservationItem();
    final _VbmappObservationTimerState? observation =
        activeItem == null ? null : _mandObservationFor(activeItem);
    if (observation == null || !observation.isRunning) {
      return;
    }
    _observationTicker = Timer.periodic(const Duration(seconds: 1), (
      Timer timer,
    ) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {});
    });
  }

  Future<void> _openActiveMandQuickRecord() async {
    final _VbmappItem? item = _activeMandObservationItem();
    if (item == null) {
      return;
    }
    final VbmappMaterialProfile? profile =
        _materialProfileFor(item, _schemaFor(item));
    final _VbmappMandEvent? event = await showDialog<_VbmappMandEvent>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappMand4QuickRecordDialog(materialProfile: profile),
        );
      },
    );
    if (event == null) {
      return;
    }
    await _addMandEvent(item, event);
  }

  Future<void> _confirmFinishActiveMandObservation() async {
    final _VbmappItem? item = _activeMandObservationItem();
    if (item == null) {
      return;
    }
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return const PadDialogViewport(
          child: _VbmappObservationFinishConfirmDialog(),
        );
      },
    );
    if (confirmed != true) {
      return;
    }
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    if (observation == null) {
      return;
    }
    await _updateMandObservation(item, observation.finish(DateTime.now()));
  }

  void _goPrevious() {
    if (_selectedItemIndex > 0) {
      setState(() => _selectedItemIndex -= 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex <= 0) {
      return;
    }
    final String previousCode = _vbmappModules[moduleIndex - 1].code;
    setState(() {
      _selectedModuleCode = previousCode;
      _selectedItemIndex = _itemsForModule(previousCode).length - 1;
    });
  }

  void _goNext() {
    final List<_VbmappItem> items = _selectedItems;
    if (_selectedItemIndex < items.length - 1) {
      setState(() => _selectedItemIndex += 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex < 0 || moduleIndex >= _vbmappModules.length - 1) {
      return;
    }
    setState(() {
      _selectedModuleCode = _vbmappModules[moduleIndex + 1].code;
      _selectedItemIndex = 0;
    });
  }

  void _jumpFirstMissing() {
    final _VbmappItem? missing = _firstMissingItem();
    if (missing == null) {
      return;
    }
    setState(() {
      _selectedModuleCode = missing.moduleCode;
      _selectedItemIndex = _itemsForModule(missing.moduleCode)
          .indexWhere((_VbmappItem item) => item.itemCode == missing.itemCode);
    });
  }

  _VbmappItem? _firstMissingItem() {
    for (final _VbmappModule module in _vbmappModules) {
      final List<_VbmappItem> items = _itemsForModule(module.code);
      final int index =
          items.indexWhere((_VbmappItem item) => _scoreFor(item) == null);
      if (index >= 0) {
        return items[index];
      }
    }
    return null;
  }

  void _selectItem(_VbmappItem item) {
    final List<_VbmappItem> items = _itemsForModule(item.moduleCode);
    final int index = items.indexWhere(
      (_VbmappItem candidate) => candidate.itemCode == item.itemCode,
    );
    if (index < 0) {
      return;
    }
    setState(() {
      _selectedModuleCode = item.moduleCode;
      _selectedItemIndex = index;
    });
  }

  Future<int> _saveDraft({bool silent = false}) async {
    if (_saving) {
      return _draftId;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿', tone: PadMessageTone.error);
      return 0;
    }
    if (_studentId <= 0) {
      _showMessage('请先从开始测评页选择学员', tone: PadMessageTone.error);
      return 0;
    }
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
    });
    try {
      final VbmappDraftSaveResult result = await widget.client.saveDraft(
        _token,
        <String, dynamic>{
          if (_draftId > 0) 'id': _draftId,
          'studentId': _studentId,
          'studentName': _studentName,
          'examinerName': _examinerName,
          'birthDate': _birthDate,
          'assessmentDate': _assessmentDate,
          'scaleVersion': _scaleVersion,
          'milestoneScores': _milestoneScores,
          'barrierScores': _barrierScores,
          'transitionScores': _transitionScores,
        },
      );
      if (!mounted) {
        return 0;
      }
      setState(() {
        _draftId = result.id > 0 ? result.id : _draftId;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
        _saving = false;
      });
      if (!silent) {
        _showMessage('VB-MAPP草稿已保存', tone: PadMessageTone.success);
      }
      return _draftId;
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return 0;
      }
      setState(() {
        _autoSaveText = '保存失败';
        _saving = false;
      });
      _showMessage(error.message, tone: PadMessageTone.error);
      return 0;
    } on Object catch (error) {
      if (!mounted) {
        return 0;
      }
      setState(() {
        _autoSaveText = '保存失败';
        _saving = false;
      });
      _showMessage('VB-MAPP草稿保存失败：$error', tone: PadMessageTone.error);
      return 0;
    }
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    final int missingCount = _totalItemCount - _answeredCount;
    if (missingCount > 0) {
      _showMessage(
        'VB-MAPP还有 $missingCount 个项目未评分，完成后才能提交正式记录',
        tone: PadMessageTone.error,
      );
      _jumpFirstMissing();
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再提交记录', tone: PadMessageTone.error);
      return;
    }
    setState(() {
      _submitting = true;
      _autoSaveText = '提交中...';
    });
    try {
      final int draftId = await _saveDraft(silent: true);
      if (draftId <= 0) {
        if (mounted) {
          setState(() => _submitting = false);
        }
        return;
      }
      await widget.client.submitDraft(_token, draftId);
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '已提交';
        _submitting = false;
      });
      _showMessage('VB-MAPP正式记录已提交', tone: PadMessageTone.success);
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '提交失败';
        _submitting = false;
      });
      _showMessage(error.message, tone: PadMessageTone.error);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '提交失败';
        _submitting = false;
      });
      _showMessage('VB-MAPP记录提交失败：$error', tone: PadMessageTone.error);
    }
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
      key: 'vbmapp-assessment-top-message',
    );
  }

  void _dismissEditingFocus() {
    final FocusManager manager = FocusManager.instance;
    if (manager.primaryFocus != null) {
      manager.primaryFocus?.unfocus();
    }
  }

  @override
  Widget build(BuildContext context) {
    final _VbmappItem item = _selectedItem;
    final _VbmappScoreSnapshot scoreSnapshot = _scoreSnapshot;
    final _VbmappItem? activeObservationItem =
        _loading ? null : _activeMandObservationItem();
    final _VbmappObservationTimerState? activeObservation =
        activeObservationItem == null
            ? null
            : _mandObservationFor(activeObservationItem);
    final bool activeTimedMandShared = activeObservationItem != null &&
        _mandStorageKeyFor(activeObservationItem.itemCode) ==
            _sharedTimedMandStorageKey;
    final bool showActiveObservationBar = activeObservationItem != null &&
        activeObservation != null &&
        !(activeTimedMandShared &&
            _sharedTimedMandItemCodes.contains(item.itemCode)) &&
        activeObservationItem.itemCode != item.itemCode;
    return GestureDetector(
      behavior: HitTestBehavior.translucent,
      onTap: _dismissEditingFocus,
      child: ColoredBox(
        color: _VbmappColors.page,
        child: Column(
          children: <Widget>[
            _VbmappTopBar(
              scaleName: widget.args.scaleName,
              studentName: _studentName,
              studentAge: _studentAgeText,
              assessmentDate: _assessmentDate,
              examinerName: _examinerName,
              autoSaveText: _autoSaveText,
              saving: _saving,
              submitting: _submitting,
              onBack: widget.onBack,
              onSave: () => unawaited(_saveDraft()),
              onSubmit: () => unawaited(_submitDraft()),
            ),
            if (showActiveObservationBar)
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 8, 12, 0),
                child: _VbmappActiveObservationBar(
                  tone: activeObservationItem.color,
                  observation: activeObservation,
                  statusLabel: activeTimedMandShared
                      ? '提要求观察中'
                      : '${activeObservationItem.navCode}观察中',
                  summaryText: activeTimedMandShared
                      ? '${_vbmappDurationText(activeObservation.elapsedSecondsAt(DateTime.now()))} · 已记录 ${_mandStoredEventsFor(activeObservationItem).length} 条'
                      : '${_vbmappDurationText(activeObservation.elapsedSecondsAt(DateTime.now()))} · ${_activeMandObservationQualifiedCount(activeObservationItem)}/${_scoreCountThreshold(activeObservationItem, 1) ?? 5}',
                  onJump: () => _selectItem(
                    _milestoneItems.firstWhere(
                      (_VbmappItem candidate) =>
                          candidate.itemCode == 'MAND_04M',
                      orElse: () => activeObservationItem,
                    ),
                  ),
                  onQuickRecord: () => unawaited(_openActiveMandQuickRecord()),
                  onPrimaryAction: () {
                    final DateTime now = DateTime.now();
                    if (activeObservation.isRunning) {
                      unawaited(
                        _updateMandObservation(
                          activeObservationItem,
                          activeObservation.pause(now),
                        ),
                      );
                      return;
                    }
                    unawaited(
                      _updateMandObservation(
                        activeObservationItem,
                        activeObservation.resume(now),
                      ),
                    );
                  },
                  onFinish: () =>
                      unawaited(_confirmFinishActiveMandObservation()),
                ),
              ),
            if (_loading)
              const Expanded(child: _VbmappLoadingState())
            else ...<Widget>[
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: <Widget>[
                      SizedBox(
                        width: 250,
                        child: _VbmappModuleRail(
                          modules: _vbmappModules,
                          selectedCode: _selectedModuleCode,
                          selectedItemCode: item.itemCode,
                          items: _selectedItems,
                          answeredCount: _answeredCountByModule,
                          isAnswered: (_VbmappItem item) =>
                              _scoreFor(item) != null,
                          onSelectModule: _selectModule,
                          onSelectItem: _selectItem,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: _VbmappWorkspace(
                          item: item,
                          score: _scoreFor(item),
                          responseSchema: _schemaFor(item),
                          materialProfile: _materialProfileFor(
                            item,
                            _schemaFor(item),
                          ),
                          mandEvents: _mandEventsFor(item),
                          mandObservation: _mandObservationFor(item),
                          onAddMandEvent: () => unawaited(
                            _openMandEventDialog(item),
                          ),
                          onSubmitMandEvent: (_VbmappMandEvent event) =>
                              unawaited(_addMandEvent(item, event)),
                          onDeleteMandEvent: (int index) =>
                              unawaited(_deleteMandEvent(item, index)),
                          onChangeMandObservation:
                              (_VbmappObservationTimerState observation) =>
                                  unawaited(
                            _updateMandObservation(item, observation),
                          ),
                          onSelectScore: _selectScore,
                        ),
                      ),
                      const SizedBox(width: 10),
                      SizedBox(
                        width: 278,
                        child: _VbmappRightRail(
                          progressPercent: _progressPercent,
                          answered: _answeredCount,
                          total: _totalItemCount,
                          selectedModule: _moduleByCode(_selectedModuleCode),
                          scoreSnapshot: scoreSnapshot,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              _VbmappFooterDock(
                current: item.sequenceNo,
                total: _totalItemCount,
                hasPrevious: item.sequenceNo > 1,
                hasNext: item.sequenceNo < _totalItemCount,
                hasMissing: _answeredCount < _totalItemCount,
                autoNext: _autoNext,
                onPrevious: _goPrevious,
                onNext: _goNext,
                onJumpMissing: _jumpFirstMissing,
                onToggleAutoNext: (bool value) =>
                    setState(() => _autoNext = value),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Map<String, int> get _answeredCountByModule {
    return <String, int>{
      'milestones': _milestoneScores.length,
      'barriers': _barrierScores.length,
      'transition': _transitionScores.length,
    };
  }
}

Map<String, VbmappMaterialProfile> _itemMaterialProfileMapFromCatalog(
  VbmappMaterialCatalog? catalog,
) {
  if (catalog == null || catalog.items.isEmpty) {
    return const <String, VbmappMaterialProfile>{};
  }
  final Map<String, VbmappMaterialProfile> out =
      <String, VbmappMaterialProfile>{};
  for (final VbmappMaterialCatalogItem item in catalog.items) {
    if (item.itemCode.isEmpty) {
      continue;
    }
    out[item.itemCode] = item.toMaterialProfile();
  }
  return out;
}

class _VbmappScoreSnapshot {
  const _VbmappScoreSnapshot({
    required this.milestoneTotal,
    required this.milestoneMax,
    required this.barrierTotal,
    required this.barrierMax,
    required this.transitionTotal,
    required this.transitionMax,
    required this.milestoneDomains,
  });

  final double milestoneTotal;
  final int milestoneMax;
  final int barrierTotal;
  final int barrierMax;
  final int transitionTotal;
  final int transitionMax;
  final List<_VbmappDomainScoreSummary> milestoneDomains;

  String get milestoneScoreText {
    return milestoneTotal.toStringAsFixed(1);
  }
}

class _VbmappDomainScoreSummary {
  const _VbmappDomainScoreSummary({
    required this.name,
    required this.score,
    required this.maxScore,
    required this.answered,
    required this.total,
  });

  final String name;
  final double score;
  final int maxScore;
  final int answered;
  final int total;

  double get percent {
    if (maxScore <= 0) {
      return 0;
    }
    return (score / maxScore).clamp(0, 1).toDouble();
  }

  String get scoreText {
    return score.toStringAsFixed(1);
  }
}

class _VbmappModule {
  const _VbmappModule({
    required this.code,
    required this.title,
    required this.subtitle,
    required this.itemCount,
    required this.icon,
    required this.color,
  });

  final String code;
  final String title;
  final String subtitle;
  final int itemCount;
  final IconData icon;
  final Color color;
}

class _VbmappItem {
  const _VbmappItem({
    required this.sequenceNo,
    required this.moduleCode,
    required this.itemCode,
    required this.label,
    required this.domainName,
    required this.ageBand,
    required this.assessmentMode,
    required this.title,
    required this.scoreTitle,
    required this.scoreOptions,
    required this.materialHint,
    required this.color,
  });

  final int sequenceNo;
  final String moduleCode;
  final String itemCode;
  final String label;
  final String domainName;
  final String ageBand;
  final String assessmentMode;
  final String title;
  final String scoreTitle;
  final List<_VbmappScoreOption> scoreOptions;
  final String materialHint;
  final Color color;

  int get localNo {
    switch (moduleCode) {
      case 'barriers':
        return sequenceNo - 170;
      case 'transition':
        return sequenceNo - 194;
      case 'milestones':
      default:
        return sequenceNo;
    }
  }

  String get navCode {
    if (moduleCode == 'milestones') {
      final RegExpMatch? labelMatch =
          RegExp(r'(\d+)\s*-\s*M').firstMatch(label);
      if (labelMatch != null) {
        return '${labelMatch.group(1)}M';
      }
      final RegExpMatch? codeMatch = RegExp(r'_(\d+)M$').firstMatch(itemCode);
      if (codeMatch != null) {
        return '${int.parse(codeMatch.group(1)!)}M';
      }
      return '${localNo}M';
    }
    return itemCode;
  }
}

class _VbmappMandEvent {
  const _VbmappMandEvent({
    required this.utterance,
    required this.target,
    required this.motivationContext,
    this.environment = '',
    this.targetKind = '',
    required this.person,
    required this.setting,
    required this.example,
    required this.responseMode,
    required this.promptLevel,
    this.phraseLevel = '',
    this.recordedAtIso = '',
    this.sourceItemCode = '',
    required this.functional,
  });

  factory _VbmappMandEvent.fromJson(Map<String, dynamic> json) {
    return _VbmappMandEvent(
      utterance: _safeText(json['utterance']),
      target: _safeText(json['target']),
      motivationContext: _safeText(json['motivationContext']),
      environment: _safeText(json['environment']),
      targetKind: _safeText(json['targetKind']),
      person: _safeText(json['person']),
      setting: _safeText(json['setting']),
      example: _safeText(json['example']),
      responseMode: _safeText(json['responseMode']),
      promptLevel: _safeText(json['promptLevel']),
      phraseLevel: _safeText(
        json['phraseLevel'] ?? json['languageLevel'] ?? json['phrase_level'],
      ),
      recordedAtIso: _safeText(
        json['recordedAtIso'] ?? json['recorded_at'] ?? '',
      ),
      sourceItemCode: _safeText(
        json['sourceItemCode'] ?? json['source_item_code'] ?? '',
      ),
      functional: json['functional'] != false,
    );
  }

  final String utterance;
  final String target;
  final String motivationContext;
  final String environment;
  final String targetKind;
  final String person;
  final String setting;
  final String example;
  final String responseMode;
  final String promptLevel;
  final String phraseLevel;
  final String recordedAtIso;
  final String sourceItemCode;
  final bool functional;

  bool get isNotEmpty => utterance.isNotEmpty || target.isNotEmpty;

  DateTime? get recordedAt => _safeDateTimeParse(recordedAtIso);

  bool get hasPhysicalPrompt => promptLevel == '肢体辅助';

  bool get hasDisallowedPrompt =>
      hasPhysicalPrompt ||
      promptLevel == '口头辅助' ||
      promptLevel == '有口头辅助' ||
      promptLevel == '有额外辅助' ||
      promptLevel == '额外辅助' ||
      promptLevel == '其他辅助';

  bool get isQualified => functional && isNotEmpty && !hasDisallowedPrompt;

  String get uniqueKey {
    final String text = target.trim().isNotEmpty ? target : utterance;
    return text.trim().toLowerCase();
  }

  String get summary {
    final String spoken = utterance.trim().isEmpty ? '未记录表达' : utterance;
    final String targetText = target.trim().isEmpty ? '未记录目标' : target;
    final List<String> dimensions = <String>[
      if (person.trim().isNotEmpty) person.trim(),
      if (setting.trim().isNotEmpty) setting.trim(),
      if (example.trim().isNotEmpty) example.trim(),
      if (environment.trim().isNotEmpty) environment.trim(),
      if (targetKind.trim().isNotEmpty) targetKind.trim(),
      if (phraseLevel.trim().isNotEmpty) phraseLevel.trim(),
    ];
    final String dimensionText =
        dimensions.isEmpty ? '' : ' · ${dimensions.join('/')}';
    return '$spoken -> $targetText · $responseMode$dimensionText';
  }

  _VbmappMandEvent copyWith({
    String? utterance,
    String? target,
    String? motivationContext,
    String? environment,
    String? targetKind,
    String? person,
    String? setting,
    String? example,
    String? responseMode,
    String? promptLevel,
    String? phraseLevel,
    String? recordedAtIso,
    String? sourceItemCode,
    bool? functional,
  }) {
    return _VbmappMandEvent(
      utterance: utterance ?? this.utterance,
      target: target ?? this.target,
      motivationContext: motivationContext ?? this.motivationContext,
      environment: environment ?? this.environment,
      targetKind: targetKind ?? this.targetKind,
      person: person ?? this.person,
      setting: setting ?? this.setting,
      example: example ?? this.example,
      responseMode: responseMode ?? this.responseMode,
      promptLevel: promptLevel ?? this.promptLevel,
      phraseLevel: phraseLevel ?? this.phraseLevel,
      recordedAtIso: recordedAtIso ?? this.recordedAtIso,
      sourceItemCode: sourceItemCode ?? this.sourceItemCode,
      functional: functional ?? this.functional,
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'utterance': utterance,
      'target': target,
      'motivationContext': motivationContext,
      'environment': environment,
      'targetKind': targetKind,
      'person': person,
      'setting': setting,
      'example': example,
      'responseMode': responseMode,
      'promptLevel': promptLevel,
      'phraseLevel': phraseLevel,
      'recordedAtIso': recordedAtIso,
      'sourceItemCode': sourceItemCode,
      'functional': functional,
    };
  }
}

class _VbmappObservationTimerState {
  const _VbmappObservationTimerState({
    this.plannedMinutes = 60,
    this.accumulatedSeconds = 0,
    this.runningSinceIso = '',
    this.startedAtIso = '',
    this.ended = false,
  });

  factory _VbmappObservationTimerState.fromJson(Map<String, dynamic> json) {
    int intFrom(Object? value) {
      if (value is num) {
        return value.toInt();
      }
      return int.tryParse(_safeText(value)) ?? 0;
    }

    final int plannedMinutes =
        intFrom(json['plannedMinutes'] ?? json['planned_minutes']);
    final int accumulatedSeconds = intFrom(
      json['accumulatedSeconds'] ??
          json['accumulated_seconds'] ??
          json['actualObservationSeconds'],
    );
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes <= 0 ? 60 : plannedMinutes,
      accumulatedSeconds: accumulatedSeconds < 0 ? 0 : accumulatedSeconds,
      runningSinceIso: _safeText(
        json['runningSinceIso'] ?? json['running_since'] ?? '',
      ),
      startedAtIso: _safeText(
        json['startedAtIso'] ?? json['startTime'] ?? json['start_time'] ?? '',
      ),
      ended: json['ended'] == true || _safeText(json['status']) == 'ended',
    );
  }

  final int plannedMinutes;
  final int accumulatedSeconds;
  final String runningSinceIso;
  final String startedAtIso;
  final bool ended;

  bool get isRunning => runningSinceIso.trim().isNotEmpty;

  bool get hasStarted =>
      startedAtIso.trim().isNotEmpty || accumulatedSeconds > 0 || isRunning;

  bool get isEmpty => !hasStarted && !ended;

  int get plannedSeconds => plannedMinutes * Duration.secondsPerMinute;

  DateTime? get runningSince => _safeDateTimeParse(runningSinceIso);

  DateTime? get startedAt => _safeDateTimeParse(startedAtIso);

  int elapsedSecondsAt(DateTime now) {
    if (!isRunning) {
      return accumulatedSeconds;
    }
    final DateTime? since = runningSince;
    if (since == null) {
      return accumulatedSeconds;
    }
    final int delta = now.difference(since).inSeconds;
    return accumulatedSeconds + (delta > 0 ? delta : 0);
  }

  _VbmappObservationTimerState start(DateTime now) {
    if (isRunning) {
      return this;
    }
    final String iso = now.toIso8601String();
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: accumulatedSeconds,
      runningSinceIso: iso,
      startedAtIso: startedAtIso.trim().isEmpty ? iso : startedAtIso,
      ended: false,
    );
  }

  _VbmappObservationTimerState resume(DateTime now) {
    return start(now);
  }

  _VbmappObservationTimerState pause(DateTime now) {
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: elapsedSecondsAt(now),
      runningSinceIso: '',
      startedAtIso: startedAtIso,
      ended: false,
    );
  }

  _VbmappObservationTimerState finish(DateTime now) {
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: elapsedSecondsAt(now),
      runningSinceIso: '',
      startedAtIso: startedAtIso,
      ended: true,
    );
  }

  _VbmappObservationTimerState restart(DateTime now) {
    final String iso = now.toIso8601String();
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: 0,
      runningSinceIso: iso,
      startedAtIso: iso,
      ended: false,
    );
  }

  _VbmappObservationTimerState withPlannedMinutes(int value) {
    if (value <= 0 || value == plannedMinutes) {
      return this;
    }
    return _VbmappObservationTimerState(
      plannedMinutes: value,
      accumulatedSeconds: accumulatedSeconds,
      runningSinceIso: runningSinceIso,
      startedAtIso: startedAtIso,
      ended: ended,
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'plannedMinutes': plannedMinutes,
      'planned_seconds': plannedSeconds,
      'accumulatedSeconds': accumulatedSeconds,
      'runningSinceIso': runningSinceIso,
      'startTime': startedAtIso,
      'status': ended
          ? 'ended'
          : isRunning
              ? 'running'
              : hasStarted
                  ? 'paused'
                  : 'idle',
      'ended': ended,
    };
  }
}

class _VbmappScoreOption {
  const _VbmappScoreOption({
    required this.score,
    required this.label,
  });

  final num score;
  final String label;

  String get displayScore {
    if (score is int || score == score.roundToDouble()) {
      return score.toInt().toString();
    }
    return score.toString();
  }
}

String _sessionExaminerName(HomeSession session) {
  if (session.nickName.trim().isNotEmpty) {
    return session.nickName.trim();
  }
  if (session.username.trim().isNotEmpty) {
    return session.username.trim();
  }
  return '';
}

String _todayIsoDate() {
  return _dateOnlyText(DateTime.now());
}

String _formatClock(DateTime value) {
  return '${value.hour.toString().padLeft(2, '0')}:'
      '${value.minute.toString().padLeft(2, '0')}';
}

String _vbmappDurationText(int totalSeconds) {
  final int seconds = totalSeconds < 0 ? 0 : totalSeconds;
  final int minutesPart = seconds ~/ Duration.secondsPerMinute;
  final int secondsPart = seconds % Duration.secondsPerMinute;
  return '${minutesPart.toString().padLeft(2, '0')}:'
      '${secondsPart.toString().padLeft(2, '0')}';
}

bool _isSimpleMandRecorder(
  _VbmappItem item,
  VbmappItemResponseSchema? schema,
) {
  if (schema?.uiPattern != 'mand_event_recorder') {
    return false;
  }
  return item.itemCode == 'MAND_01M' ||
      item.itemCode == 'MAND_02M' ||
      item.itemCode == 'MAND_03M' ||
      item.itemCode == 'MAND_04M';
}

bool _isTimedMandItemCode(String itemCode) {
  return itemCode == 'MAND_04M' ||
      itemCode == 'MAND_08M' ||
      itemCode == 'MAND_09M';
}

int _plannedMinutesForMandItem(_VbmappItem item) {
  switch (item.itemCode) {
    case 'MAND_09M':
      return 30;
    case 'MAND_04M':
    case 'MAND_08M':
    default:
      return 60;
  }
}

String _mandInitiationText(_VbmappMandEvent event) {
  final String prompt = event.promptLevel.trim();
  if (prompt == '提问下' || prompt == '自发地') {
    return prompt;
  }
  if (event.responseMode.contains('自发')) {
    return '自发地';
  }
  return '';
}

bool _mandEventWithinWindow(
  _VbmappMandEvent event,
  _VbmappObservationTimerState? observation, {
  required int plannedMinutes,
}) {
  final DateTime? startedAt = observation?.startedAt;
  final DateTime? recordedAt = event.recordedAt;
  if (startedAt == null || recordedAt == null) {
    return true;
  }
  if (recordedAt.isBefore(startedAt)) {
    return false;
  }
  return !recordedAt.isAfter(
    startedAt.add(Duration(minutes: plannedMinutes)),
  );
}

bool _mandEventCountsForItem(
  _VbmappItem item,
  _VbmappMandEvent event, {
  _VbmappObservationTimerState? observation,
}) {
  if (!event.isQualified) {
    return false;
  }
  if (_isTimedMandItemCode(item.itemCode) &&
      !_mandEventWithinWindow(
        event,
        observation,
        plannedMinutes: _plannedMinutesForMandItem(item),
      )) {
    return false;
  }
  switch (item.itemCode) {
    case 'MAND_05M':
      return event.environment.trim() == '呈现物品' &&
          _mandInitiationText(event) != '提问下';
    case 'MAND_04M':
      return event.environment.trim() == '呈现物品' &&
          _mandInitiationText(event) != '提问下';
    case 'MAND_08M':
      return true;
    case 'MAND_09M':
      return _mandInitiationText(event) != '提问下';
    default:
      return event.isQualified;
  }
}

int _qualifiedMandCount(List<_VbmappMandEvent> events) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (event.isQualified && event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _qualifiedMandCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (_mandEventCountsForItem(item, event, observation: observation) &&
        event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _mandPhraseQualifiedCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (!_mandEventCountsForItem(item, event, observation: observation)) {
      continue;
    }
    if (!_isLikelyMultiWordMand(event)) {
      continue;
    }
    if (event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

bool _isLikelyMultiWordMand(_VbmappMandEvent event) {
  final String explicitLevel = event.phraseLevel.trim();
  if (explicitLevel == '双词+') {
    return true;
  }
  if (explicitLevel == '单词') {
    return false;
  }

  String text = _mandRequestText(event)
      .replaceAll(RegExp(r'[，。！？、,.!?;；:/\\]+'), ' ')
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();
  if (text.isEmpty) {
    return false;
  }

  final List<String> spacedTokens = text
      .split(' ')
      .map((String part) => part.trim())
      .where((String part) => part.isNotEmpty)
      .toList(growable: false);
  if (spacedTokens.length >= 2) {
    return true;
  }

  text = text.replaceFirst(RegExp(r'^我想要'), '').trim();
  if (text.isEmpty) {
    return false;
  }

  if (RegExp(r'(我|你|他|她|它|我们|我要|给我|帮我)').hasMatch(text) &&
      text.runes.length >= 3) {
    return true;
  }
  if (RegExp(r'(快点|一下|一会|给我|帮我|让我|一起|该我了|不要|好了|开门)$').hasMatch(text)) {
    return true;
  }
  if (text.runes.length >= 3 &&
      RegExp(r'^(打开|帮我|给我|让我|带我|一起|还要|再来|我要|我们|该我|跑|倒|推|拿|去|来|开)')
          .hasMatch(text)) {
    return true;
  }

  return false;
}

String _mandRequestText(_VbmappMandEvent event) {
  final String utterance = event.utterance.trim();
  if (utterance.isNotEmpty) {
    return utterance;
  }
  final String target = event.target.trim();
  if (target.isNotEmpty) {
    return target;
  }
  return '未记录要求';
}

String _mandRecordMetaText(
  _VbmappMandEvent event, {
  _VbmappItem? item,
}) {
  final List<String> values = <String>[];

  void addMeta(String raw) {
    final String value = raw.trim();
    if (value.isEmpty || values.contains(value)) {
      return;
    }
    values.add(value);
  }

  addMeta(_mandInitiationText(event));
  addMeta(event.environment);
  addMeta(event.targetKind);
  addMeta(event.phraseLevel);
  addMeta(event.promptLevel);
  if (item != null && _isTimedMandItemCode(item.itemCode)) {
    final DateTime? recordedAt = event.recordedAt;
    if (recordedAt != null) {
      values.add(_formatClock(recordedAt));
    }
  }
  return values.isEmpty ? '未记录条件' : values.join(' · ');
}

int _effectiveObservationSecondsForItem(
  _VbmappItem item,
  _VbmappObservationTimerState? observation,
) {
  if (observation == null) {
    return 0;
  }
  final int elapsed = observation.elapsedSecondsAt(DateTime.now());
  final int maxSeconds =
      _plannedMinutesForMandItem(item) * Duration.secondsPerMinute;
  return elapsed > maxSeconds ? maxSeconds : elapsed;
}

String _mandTimedScoreBasisText(
  _VbmappItem item,
  double suggestedScore,
  int qualifiedCount,
  int actualObservationSeconds,
  int multiWordCount,
) {
  final String baseDuration = '${_plannedMinutesForMandItem(item)}分钟观察窗';
  switch (item.itemCode) {
    case 'MAND_08M':
      return '系统按$baseDuration内的不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，其中双词+$multiWordCount条，'
          '已观察${_vbmappDurationText(actualObservationSeconds)}，老师可在下方评分区覆盖。';
    case 'MAND_09M':
      return '系统按$baseDuration内的自发不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
          '老师可在下方评分区覆盖。';
    case 'MAND_04M':
    default:
      return '系统按$baseDuration内的有效自发要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
          '老师可在下方评分区覆盖。';
  }
}

double _suggestMandScore(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  if (item.itemCode == 'MAND_03M') {
    final Map<String, int> counts = _mandGeneralizationCounts(events);
    final bool onePoint = (counts['people'] ?? 0) >= 2 &&
        (counts['settings'] ?? 0) >= 2 &&
        (counts['examples'] ?? 0) >= 2;
    if (onePoint) {
      return 1;
    }
    final bool halfPoint = (counts['people'] ?? 0) >= 1 &&
        (counts['settings'] ?? 0) >= 1 &&
        (counts['examples'] ?? 0) >= 1;
    return halfPoint ? .5 : 0;
  }
  final int count = _qualifiedMandCountForItem(
    item,
    events,
    observation: observation,
  );
  if (item.itemCode == 'MAND_08M') {
    final int multiWordCount = _mandPhraseQualifiedCountForItem(
      item,
      events,
      observation: observation,
    );
    if (count >= 5 && multiWordCount >= 2) {
      return 1;
    }
    if (count >= 2) {
      return .5;
    }
    return 0;
  }
  final int onePointCount = _scoreCountThreshold(item, 1) ?? 1;
  final int halfPointCount = _scoreCountThreshold(item, .5) ?? onePointCount;
  if (count >= onePointCount) {
    return 1;
  }
  if (count >= halfPointCount) {
    return .5;
  }
  return 0;
}

Map<String, int> _mandGeneralizationCounts(List<_VbmappMandEvent> events) {
  final Map<String, List<String>> values = _mandGeneralizationValues(events);
  return <String, int>{
    'people': values['people']?.length ?? 0,
    'settings': values['settings']?.length ?? 0,
    'examples': values['examples']?.length ?? 0,
  };
}

Map<String, List<String>> _mandGeneralizationValues(
  List<_VbmappMandEvent> events,
) {
  final Set<String> people = <String>{};
  final Set<String> settings = <String>{};
  final Set<String> examples = <String>{};
  final List<String> peopleValues = <String>[];
  final List<String> settingValues = <String>[];
  final List<String> exampleValues = <String>[];
  for (final _VbmappMandEvent event in events) {
    if (!event.isQualified) {
      continue;
    }
    if (event.person.trim().isNotEmpty) {
      final String value = event.person.trim();
      final String normalized = value.toLowerCase();
      if (people.add(normalized)) {
        peopleValues.add(value);
      }
    }
    if (event.setting.trim().isNotEmpty) {
      final String value = event.setting.trim();
      final String normalized = value.toLowerCase();
      if (settings.add(normalized)) {
        settingValues.add(value);
      }
    }
    if (event.example.trim().isNotEmpty) {
      final String value = event.example.trim();
      final String normalized = value.toLowerCase();
      if (examples.add(normalized)) {
        exampleValues.add(value);
      }
    }
  }
  return <String, List<String>>{
    'people': peopleValues,
    'settings': settingValues,
    'examples': exampleValues,
  };
}

String _mand3DimensionText(_VbmappMandEvent event) {
  if (event.person.trim().isNotEmpty) {
    return '人物：${event.person.trim()}';
  }
  if (event.setting.trim().isNotEmpty) {
    return '环境：${event.setting.trim()}';
  }
  if (event.example.trim().isNotEmpty) {
    return '例子：${event.example.trim()}';
  }
  return '未记录';
}

int? _scoreCountThreshold(_VbmappItem item, num score) {
  for (final _VbmappScoreOption option in item.scoreOptions) {
    if (option.score == score) {
      final RegExpMatch? match =
          RegExp(r'[：:]\s*(\d+)').firstMatch(option.label) ??
              RegExp(r'(\d+)').firstMatch(option.label);
      if (match != null) {
        return int.tryParse(match.group(1)!);
      }
    }
  }
  return null;
}

String _schemaKey(String moduleCode, String itemCode) {
  return '${_safeText(moduleCode).toLowerCase()}::${_safeText(itemCode).toUpperCase()}';
}

String _formatScore(num score) {
  if (score == score.roundToDouble()) {
    return score.toInt().toString();
  }
  return score.toString();
}

Map<String, List<String>> _normalizedMaterialQuickPicks(
  Map<String, Object?> raw,
) {
  if (raw.isEmpty) {
    return const <String, List<String>>{};
  }
  final Map<String, List<String>> out = <String, List<String>>{};
  raw.forEach((String key, Object? value) {
    final List<String> values = _materialStringList(value);
    if (key.trim().isNotEmpty && values.isNotEmpty) {
      out[key.trim()] = values;
    }
  });
  return out;
}

List<String> _materialStringList(Object? raw) {
  if (raw is! List) {
    return const <String>[];
  }
  return raw
      .map((Object? value) => _safeText(value))
      .where((String value) => value.isNotEmpty)
      .toList(growable: false);
}

String _materialFieldLabel(String key) {
  if (key.contains('people')) {
    return '人物';
  }
  if (key.contains('settings')) {
    return '环境';
  }
  if (key.contains('examples')) {
    return '例子';
  }
  return '词库';
}

List<String> _deduplicatedTexts(List<String> values) {
  final Set<String> seen = <String>{};
  final List<String> out = <String>[];
  for (final String value in values) {
    final String normalized = value.trim();
    if (normalized.isEmpty || seen.contains(normalized)) {
      continue;
    }
    seen.add(normalized);
    out.add(normalized);
  }
  return out;
}

List<String> _smartMandQuickPicks(
  VbmappMaterialProfile? profile, {
  required String targetKind,
  required List<String> fallback,
  int limit = 8,
}) {
  final List<String> typed = <String>[
    for (final VbmappMaterialSuggestion material
        in profile?.recommendedMaterials ?? const <VbmappMaterialSuggestion>[])
      if (_materialMatchesMandTarget(
        name: material.name,
        type: material.type,
        targetKind: targetKind,
      ))
        material.name,
  ];
  final List<String> untyped = profile?.quickPicks ?? const <String>[];
  final List<String> targetFallback = _mandFallbackQuickPicksForTarget(
    targetKind,
  );
  final List<String> filteredFallback = fallback
      .where((String value) => _materialMatchesMandTarget(
            name: value,
            type: '',
            targetKind: targetKind,
          ))
      .toList(growable: false);
  final List<String> source = typed.isEmpty
      ? <String>[
          ...untyped,
          ...targetFallback,
          ...filteredFallback,
        ]
      : <String>[...typed, ...targetFallback, ...untyped, ...filteredFallback];
  return _deduplicatedTexts(source).take(limit).toList(growable: false);
}

List<String> _mandFallbackQuickPicksForTarget(String targetKind) {
  switch (targetKind.trim()) {
    case '动作':
      return const <String>['打开', '出去', '帮我', '给我', '推', '倒果汁'];
    case '活动':
      return const <String>['音乐', '秋千', '泡泡', '转圈', '一起玩', '出去'];
    case '物品':
    default:
      return const <String>['饼干', '书', '球', '泡泡', '车', '积木', '彩虹弹簧'];
  }
}

bool _materialMatchesMandTarget({
  required String name,
  required String type,
  required String targetKind,
}) {
  final String normalizedName = name.trim();
  final String normalizedType = type.trim();
  if (normalizedName.isEmpty) {
    return false;
  }
  switch (targetKind.trim()) {
    case '动作':
      return normalizedType.contains('动作') ||
          normalizedType.contains('帮助') ||
          _looksLikeActionMand(normalizedName);
    case '活动':
      return normalizedType.contains('活动') ||
          normalizedType.contains('社交游戏') ||
          _looksLikeActivityMand(normalizedName);
    case '物品':
    default:
      return !_looksLikeActionMand(normalizedName) &&
          !_looksLikeActivityMand(normalizedName);
  }
}

bool _looksLikeActionMand(String value) {
  return RegExp(r'(打开|出去|帮我|给我|推|倒|拿|开门|关门|再来|快点|该我了)').hasMatch(value.trim());
}

bool _looksLikeActivityMand(String value) {
  return RegExp(r'(音乐|秋千|泡泡|转圈|一起玩|游戏|出去)').hasMatch(value.trim());
}

Map<String, dynamic> _dynamicMap(Object? raw) {
  if (raw is Map) {
    final Map<String, dynamic> out = <String, dynamic>{};
    raw.forEach((Object? key, Object? value) {
      final String normalizedKey = _safeText(key);
      if (normalizedKey.isNotEmpty) {
        out[normalizedKey] = value;
      }
    });
    return out;
  }
  return <String, dynamic>{};
}

DateTime? _safeDateTimeParse(Object? value) {
  final String text = _safeText(value);
  if (text.isEmpty) {
    return null;
  }
  return DateTime.tryParse(text);
}

String _dateOnlyText(Object? value) {
  if (value is DateTime) {
    return '${value.year.toString().padLeft(4, '0')}-'
        '${value.month.toString().padLeft(2, '0')}-'
        '${value.day.toString().padLeft(2, '0')}';
  }
  final String text = _safeText(value);
  if (text.length >= 10) {
    return text.substring(0, 10);
  }
  return text;
}

String _safeText(Object? value) {
  if (value == null) {
    return '';
  }
  if (value is String) {
    return value.trim();
  }
  return '$value'.trim();
}
