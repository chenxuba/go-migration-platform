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
    _syncSharedTimedMandScores();
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
      _syncSharedTimedMandScores();
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
      _syncSharedTimedMandScores();
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
      _syncSharedTimedMandScores();
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

  String _mandStorageKeyFor(String itemCode) {
    return _sharedTimedMandItemCodes.contains(itemCode)
        ? _sharedTimedMandStorageKey
        : itemCode;
  }

  void _syncSharedTimedMandScores() {
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

class _VbmappColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF3F2B22);
  static const Color body = Color(0xFF705B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color green = Color(0xFF7FA874);
  static const Color red = Color(0xFFD85F4A);
  static const Color blue = Color(0xFF5D7F9F);
}

List<BoxShadow> _vbmappShadow({
  Color color = const Color(0x12B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 10),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

class _VbmappTopBar extends StatelessWidget {
  const _VbmappTopBar({
    required this.scaleName,
    required this.studentName,
    required this.studentAge,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final String scaleName;
  final String studentName;
  final String studentAge;
  final String assessmentDate;
  final String examinerName;
  final String autoSaveText;
  final bool saving;
  final bool submitting;
  final VoidCallback onBack;
  final VoidCallback onSave;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final String title =
        scaleName.trim().isEmpty || scaleName.contains('VB-MAPP')
            ? 'VB-MAPP语言行为评估'
            : scaleName.trim();
    final String student =
        studentName.trim().isEmpty ? '-' : studentName.trim();
    final String age = studentAge.trim().isEmpty ? '未知' : studentAge.trim();
    final String date =
        assessmentDate.trim().isEmpty ? _todayIsoDate() : assessmentDate;
    final String examiner =
        examinerName.trim().isEmpty ? '-' : examinerName.trim();
    final String status =
        autoSaveText.trim().isEmpty ? '等待作答' : autoSaveText.trim();

    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.98),
        border: Border.all(color: _VbmappColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1280;
          final List<Widget> headerChildren = <Widget>[
            Text(
              '$title 测评工作台',
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 23,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            _VbmappHeaderMeta(label: '儿童', value: student, compact: compact),
            _VbmappHeaderMeta(label: '年龄', value: age, compact: compact),
            _VbmappHeaderMeta(
              label: compact ? '日期' : '测评日期',
              value: date,
              compact: compact,
            ),
            _VbmappHeaderMeta(
              label: '施测者',
              value: examiner,
              compact: compact,
            ),
          ];
          return Row(
            children: <Widget>[
              _VbmappIconButtonBox(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  physics: const ClampingScrollPhysics(),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: headerChildren,
                  ),
                ),
              ),
              _VbmappSaveStatusLabel(text: status, saving: saving),
              const SizedBox(width: 10),
              _VbmappTopActionButton(
                label: '保存草稿',
                icon: Icons.save_outlined,
                loading: saving,
                filled: false,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _VbmappTopActionButton(
                label: '提交记录',
                icon: Icons.fact_check_outlined,
                loading: submitting,
                filled: true,
                onTap: onSubmit,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _VbmappHeaderMeta extends StatelessWidget {
  const _VbmappHeaderMeta({
    required this.label,
    required this.value,
    required this.compact,
  });

  final String label;
  final String value;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: EdgeInsets.only(left: compact ? 6 : 10),
      padding: EdgeInsets.only(left: compact ? 6 : 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _VbmappColors.line)),
      ),
      child: Text.rich(
        TextSpan(
          children: <InlineSpan>[
            TextSpan(text: '$label：'),
            TextSpan(
              text: value,
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ],
        ),
        maxLines: 1,
        softWrap: false,
        style: const TextStyle(
          color: _VbmappColors.body,
          fontSize: 13,
          height: 1,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _VbmappSaveStatusLabel extends StatelessWidget {
  const _VbmappSaveStatusLabel({required this.text, required this.saving});

  final String text;
  final bool saving;

  bool get _activeSaving {
    return saving ||
        text.contains('保存中') ||
        text.contains('提交中') ||
        text.contains('草稿保存中') ||
        text.startsWith('正在');
  }

  bool get _failed {
    return text.contains('失败');
  }

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 116),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.end,
        children: <Widget>[
          Icon(
            _failed
                ? Icons.error_outline_rounded
                : (_activeSaving
                    ? Icons.sync_rounded
                    : Icons.check_circle_outline_rounded),
            color: _failed
                ? _VbmappColors.red
                : (_activeSaving
                    ? _VbmappColors.orangeDeep
                    : _VbmappColors.green),
            size: _activeSaving ? 17 : 18,
          ),
          const SizedBox(width: 5),
          Text(
            text,
            maxLines: 1,
            softWrap: false,
            textAlign: TextAlign.right,
            style: TextStyle(
              color: _failed
                  ? _VbmappColors.red
                  : (_activeSaving
                      ? _VbmappColors.orangeDeep
                      : _VbmappColors.body),
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappIconButtonBox extends StatelessWidget {
  const _VbmappIconButtonBox({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          width: 46,
          height: 46,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _VbmappColors.line),
          ),
          child: Icon(icon, color: _VbmappColors.body, size: 34),
        ),
      ),
    );
  }
}

class _VbmappTopActionButton extends StatelessWidget {
  const _VbmappTopActionButton({
    required this.label,
    required this.icon,
    required this.loading,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool loading;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final bool interactive = enabled && !loading;
    final Color foreground = filled
        ? Colors.white
        : enabled
            ? _VbmappColors.orangeDeep
            : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: interactive ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: filled
                ? enabled
                    ? _VbmappColors.orange
                    : const Color(0xFFE7DDD6)
                : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
            boxShadow: filled && enabled
                ? _vbmappShadow(
                    color: const Color(0x2AE96F43),
                    blur: 12,
                    offset: const Offset(0, 5),
                  )
                : null,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (loading)
                SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: filled ? Colors.white : _VbmappColors.orange,
                  ),
                )
              else
                Icon(icon, size: 17, color: foreground),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: foreground,
                  fontSize: 14,
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

class _VbmappModuleRail extends StatelessWidget {
  const _VbmappModuleRail({
    required this.modules,
    required this.selectedCode,
    required this.selectedItemCode,
    required this.items,
    required this.answeredCount,
    required this.isAnswered,
    required this.onSelectModule,
    required this.onSelectItem,
  });

  final List<_VbmappModule> modules;
  final String selectedCode;
  final String selectedItemCode;
  final List<_VbmappItem> items;
  final Map<String, int> answeredCount;
  final bool Function(_VbmappItem item) isAnswered;
  final ValueChanged<String> onSelectModule;
  final ValueChanged<_VbmappItem> onSelectItem;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          for (final _VbmappModule module in modules) ...<Widget>[
            _VbmappModuleTile(
              module: module,
              selected: module.code == selectedCode,
              answered: answeredCount[module.code] ?? 0,
              onTap: () => onSelectModule(module.code),
            ),
            const SizedBox(height: 6),
          ],
          const SizedBox(height: 2),
          const Divider(height: 1, color: _VbmappColors.lineSoft),
          const SizedBox(height: 8),
          Expanded(
            child: ListView.builder(
              padding: EdgeInsets.zero,
              itemCount: items.length,
              itemBuilder: (BuildContext context, int index) {
                final _VbmappItem item = items[index];
                final _VbmappItem? previous =
                    index > 0 ? items[index - 1] : null;
                final bool showHeader = previous == null ||
                    previous.domainName != item.domainName ||
                    previous.ageBand != item.ageBand;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    if (showHeader)
                      Padding(
                        padding: EdgeInsets.only(
                          top: index == 0 ? 0 : 10,
                          bottom: 6,
                        ),
                        child: Text(
                          '${item.domainName} · ${item.ageBand}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _VbmappColors.muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                    _VbmappItemNavTile(
                      item: item,
                      selected: item.itemCode == selectedItemCode,
                      answered: isAnswered(item),
                      onTap: () => onSelectItem(item),
                    ),
                    const SizedBox(height: 6),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappModuleTile extends StatelessWidget {
  const _VbmappModuleTile({
    required this.module,
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappModule module;
  final bool selected;
  final int answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent = selected ? module.color : _VbmappColors.body;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? module.color.withOpacity(.12) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  selected ? module.color.withOpacity(.55) : _VbmappColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Icon(module.icon, size: 19, color: accent),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  module.title,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: selected
                      ? module.color.withOpacity(.14)
                      : const Color(0xFFFFF6EF),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: selected
                        ? module.color.withOpacity(.3)
                        : _VbmappColors.lineSoft,
                  ),
                ),
                child: Text(
                  '$answered/${module.itemCount}',
                  style: TextStyle(
                    color: selected ? module.color : _VbmappColors.body,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappItemNavTile extends StatelessWidget {
  const _VbmappItemNavTile({
    required this.item,
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappItem item;
  final bool selected;
  final bool answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent = item.color;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? accent.withOpacity(.12) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  selected ? accent.withOpacity(.55) : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  item.navCode,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              if (answered)
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: accent,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.check, color: Colors.white, size: 14),
                )
              else
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFF6EF),
                    shape: BoxShape.circle,
                    border: Border.all(color: _VbmappColors.line),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappWorkspace extends StatelessWidget {
  const _VbmappWorkspace({
    required this.item,
    required this.score,
    required this.responseSchema,
    required this.materialProfile,
    required this.mandEvents,
    required this.mandObservation,
    required this.onAddMandEvent,
    required this.onSubmitMandEvent,
    required this.onDeleteMandEvent,
    required this.onChangeMandObservation,
    required this.onSelectScore,
  });

  final _VbmappItem item;
  final num? score;
  final VbmappItemResponseSchema? responseSchema;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> mandEvents;
  final _VbmappObservationTimerState? mandObservation;
  final VoidCallback onAddMandEvent;
  final ValueChanged<_VbmappMandEvent> onSubmitMandEvent;
  final ValueChanged<int> onDeleteMandEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeMandObservation;
  final ValueChanged<num> onSelectScore;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _VbmappQuestionHeader(
                    item: item,
                    schema: responseSchema,
                    materialProfile: materialProfile,
                  ),
                  const SizedBox(height: 12),
                  _VbmappSmartEvidencePanel(
                    item: item,
                    schema: responseSchema,
                    materialProfile: materialProfile,
                    mandEvents: mandEvents,
                    mandObservation: mandObservation,
                    onAddMandEvent: onAddMandEvent,
                    onSubmitMandEvent: onSubmitMandEvent,
                    onDeleteMandEvent: onDeleteMandEvent,
                    onChangeMandObservation: onChangeMandObservation,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          const Divider(height: 1, color: _VbmappColors.lineSoft),
          const SizedBox(height: 14),
          _VbmappScoreDock(
            item: item,
            score: score,
            onSelectScore: onSelectScore,
          ),
        ],
      ),
    );
  }
}

bool _hasPreparationData(
  VbmappItemResponseSchema? schema,
  VbmappMaterialProfile? materialProfile,
) {
  return (materialProfile != null &&
          (materialProfile.quickPickLabels.isNotEmpty ||
              materialProfile.preparationChecks.isNotEmpty ||
              materialProfile.quickPicksByField.isNotEmpty)) ||
      (schema != null && schema.qualityChecks.isNotEmpty);
}

bool _shouldShowPreparationEntry(
  VbmappItemResponseSchema? schema,
  VbmappMaterialProfile? materialProfile,
) {
  return schema?.showPreparationEntry == true &&
      _hasPreparationData(schema, materialProfile);
}

class _VbmappPreparationDialog extends StatelessWidget {
  const _VbmappPreparationDialog({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 860,
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(
                    Icons.inventory_2_outlined,
                    color: item.color,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  const Expanded(
                    child: Text(
                      '题前准备',
                      style: TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappIconButtonBox(
                    icon: Icons.close_rounded,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Flexible(
                fit: FlexFit.loose,
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxHeight: 560),
                  child: SingleChildScrollView(
                    padding: EdgeInsets.zero,
                    child: _VbmappPreparationContent(
                      item: item,
                      schema: schema,
                      materialProfile: materialProfile,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappPreparationContent extends StatelessWidget {
  const _VbmappPreparationContent({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final VbmappMaterialProfile? profile = materialProfile;
    final List<String> quickPicks =
        profile?.quickPickLabels ?? const <String>[];
    final List<String> checks = _deduplicatedTexts(<String>[
      ...?profile?.preparationChecks,
      ...?schema?.qualityChecks,
    ]);
    final Map<String, List<String>> fieldQuickPicks =
        _normalizedMaterialQuickPicks(
      profile?.quickPicksByField ?? const <String, Object?>{},
    );
    final List<MapEntry<String, List<String>>> visibleFieldQuickPicks =
        fieldQuickPicks.entries.take(6).toList(growable: false);
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          if (profile != null && profile.label.trim().isNotEmpty) ...<Widget>[
            Text(
              profile.label.trim(),
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
          ],
          if (quickPicks.isNotEmpty) ...<Widget>[
            _VbmappPreparationSection(
              title: '推荐素材',
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  for (final String label in quickPicks.take(16))
                    _VbmappPreparationChip(label: label),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          if (fieldQuickPicks.isNotEmpty) ...<Widget>[
            _VbmappPreparationSection(
              title: '快捷词',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  for (int index = 0;
                      index < visibleFieldQuickPicks.length;
                      index++) ...<Widget>[
                    if (index > 0) const SizedBox(height: 6),
                    Text(
                      '${_materialFieldLabel(visibleFieldQuickPicks[index].key)}：${visibleFieldQuickPicks[index].value.join('、')}',
                      style: const TextStyle(
                        color: _VbmappColors.body,
                        fontSize: 12.5,
                        height: 1.4,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          if (checks.isNotEmpty)
            _VbmappPreparationSection(
              title: '准备检查',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  for (final String check in checks)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 6),
                      child: Text(
                        '- $check',
                        style: const TextStyle(
                          color: _VbmappColors.body,
                          fontSize: 12.5,
                          height: 1.4,
                          fontWeight: FontWeight.w700,
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

class _VbmappPreparationSection extends StatelessWidget {
  const _VbmappPreparationSection({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(
            color: _VbmappColors.ink,
            fontSize: 12.5,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 6),
        child,
      ],
    );
  }
}

class _VbmappPreparationChip extends StatelessWidget {
  const _VbmappPreparationChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.line),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: _VbmappColors.body,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _VbmappQuestionHeader extends StatelessWidget {
  const _VbmappQuestionHeader({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final Color accent = item.color;
    final _VbmappQuestionHeaderCopy headerCopy =
        _vbmappQuestionHeaderCopy(item);
    final Widget badge = Container(
      padding: const EdgeInsets.fromLTRB(9, 6, 10, 6),
      decoration: BoxDecoration(
        color: accent.withOpacity(.08),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: accent.withOpacity(.22)),
      ),
      child: Wrap(
        spacing: 7,
        runSpacing: 5,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: <Widget>[
          Container(
            width: 3,
            height: 18,
            decoration: BoxDecoration(
              color: accent,
              borderRadius: BorderRadius.circular(999),
            ),
          ),
          Text(
            _vbmappQuestionDomainLabel(item),
            style: TextStyle(
              color: accent,
              fontSize: 14,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Text(
            _vbmappQuestionStepLabel(item),
            style: TextStyle(
              color: accent,
              fontSize: 16,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            width: 1,
            height: 14,
            color: accent.withOpacity(.28),
          ),
          Text(
            '${_vbmappQuestionStageLabel(item)} ${item.ageBand}',
            style: TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            width: 1,
            height: 14,
            color: accent.withOpacity(.28),
          ),
          Text(
            _vbmappQuestionModuleLabel(item.moduleCode),
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
            decoration: BoxDecoration(
              color: accent.withOpacity(.11),
              borderRadius: BorderRadius.circular(999),
              border: Border.all(color: accent.withOpacity(.18)),
            ),
            child: Text(
              item.assessmentMode,
              style: TextStyle(
                color: accent,
                fontSize: 11.5,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );

    final Widget title = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        for (int index = 0;
            index < headerCopy.lines.length;
            index++) ...<Widget>[
          if (index > 0) const SizedBox(height: 3),
          Text(
            headerCopy.lines[index],
            style: const TextStyle(
              color: _VbmappColors.ink,
              fontSize: 20,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ],
    );

    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool hasPreparationContent =
            _shouldShowPreparationEntry(schema, materialProfile);
        final Widget? preparationAction = hasPreparationContent
            ? _VbmappPreparationEntryButton(
                item: item,
                schema: schema,
                materialProfile: materialProfile,
              )
            : null;
        final Widget badgeSlot = Expanded(
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const ClampingScrollPhysics(),
            child: Align(
              alignment: Alignment.centerLeft,
              child: badge,
            ),
          ),
        );
        if (constraints.maxWidth < 760) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: <Widget>[
                  badgeSlot,
                  if (preparationAction != null) ...<Widget>[
                    const SizedBox(width: 8),
                    preparationAction,
                  ],
                ],
              ),
              const SizedBox(height: 7),
              title,
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                badgeSlot,
                if (preparationAction != null) ...<Widget>[
                  const SizedBox(width: 10),
                  preparationAction,
                ],
              ],
            ),
            const SizedBox(height: 8),
            title,
          ],
        );
      },
    );
  }
}

class _VbmappPreparationEntryButton extends StatelessWidget {
  const _VbmappPreparationEntryButton({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final VbmappMaterialProfile? profile = materialProfile;
    final List<String> quickPicks =
        profile?.quickPickLabels ?? const <String>[];
    final List<String> checks = _deduplicatedTexts(<String>[
      ...?profile?.preparationChecks,
      ...?schema?.qualityChecks,
    ]);
    final Color accent = item.color;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        key: const ValueKey<String>('vbmapp-preparation-entry'),
        onTap: () => _openPreparationDialog(context),
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _VbmappColors.lineSoft),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(Icons.inventory_2_outlined, color: accent, size: 16),
              const SizedBox(width: 6),
              const Text(
                '题前准备',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 12.5,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              if (quickPicks.isNotEmpty || checks.isNotEmpty) ...<Widget>[
                const SizedBox(width: 7),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                  decoration: BoxDecoration(
                    color: accent.withOpacity(.08),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: Text(
                    '${quickPicks.length}/${checks.length}',
                    style: TextStyle(
                      color: accent,
                      fontSize: 10.5,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _openPreparationDialog(BuildContext context) async {
    await showDialog<void>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappPreparationDialog(
            item: item,
            schema: schema,
            materialProfile: materialProfile,
          ),
        );
      },
    );
  }
}

class _VbmappQuestionHeaderCopy {
  const _VbmappQuestionHeaderCopy({
    required this.lines,
  });

  final List<String> lines;
}

String _vbmappQuestionDomainLabel(_VbmappItem item) {
  final String domain = item.domainName.trim();
  if (domain.isNotEmpty) {
    return domain;
  }
  return _vbmappQuestionModuleLabel(item.moduleCode);
}

String _vbmappQuestionStepLabel(_VbmappItem item) {
  final String label = item.label.trim();
  final String domain = item.domainName.trim();
  if (domain.isNotEmpty && label.startsWith(domain)) {
    final String step = label.substring(domain.length).trim();
    if (step.isNotEmpty) {
      return step;
    }
  }
  if (item.moduleCode == 'milestones') {
    final RegExpMatch? match = RegExp(r'^(\d+)M$').firstMatch(item.navCode);
    if (match != null) {
      return '${match.group(1)}-M';
    }
    return item.navCode;
  }
  return item.itemCode;
}

_VbmappQuestionHeaderCopy _vbmappQuestionHeaderCopy(_VbmappItem item) {
  switch (item.itemCode) {
    case 'MAND_08M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '能提出5个不同的要求，其中至少要包含2个或2个以上的单词（不包括“我想要”）（如：跑快点、该我了、倒果汁）（TO：60分钟）',
        ],
      );
    case 'MAND_04M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '自发性地提出（没有口头辅助）5项要求，所要的物件可在眼前（TO：60分钟）',
        ],
      );
    case 'MAND_09M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '自发性地提出15个不同的要求，（如我们一起玩、打开、我想要书）（TO：30分钟）',
        ],
      );
  }
  return _VbmappQuestionHeaderCopy(
    lines: <String>[_vbmappQuestionTitleWithoutMode(item)],
  );
}

String _vbmappQuestionTitleWithoutMode(_VbmappItem item) {
  final String title = item.title.trim();
  if (title.isEmpty) {
    return '';
  }
  final String assessmentMode = item.assessmentMode.trim();
  if (assessmentMode.isEmpty) {
    return title;
  }
  return title
      .replaceFirst('（$assessmentMode）', '')
      .replaceFirst('($assessmentMode)', '')
      .trim();
}

String _vbmappQuestionStageLabel(_VbmappItem item) {
  switch (item.ageBand.trim()) {
    case '0-18个月':
      return '第一阶段';
    case '18-30个月':
      return '第二阶段';
    case '30-48个月':
      return '第三阶段';
    default:
      return '阶段';
  }
}

String _vbmappQuestionModuleLabel(String moduleCode) {
  switch (moduleCode) {
    case 'barriers':
      return '障碍评估';
    case 'transition':
      return '转衔评估';
    case 'milestones':
    default:
      return '里程碑评估';
  }
}

class _VbmappSmartEvidencePanel extends StatelessWidget {
  const _VbmappSmartEvidencePanel({
    required this.item,
    required this.schema,
    required this.materialProfile,
    required this.mandEvents,
    required this.mandObservation,
    required this.onAddMandEvent,
    required this.onSubmitMandEvent,
    required this.onDeleteMandEvent,
    required this.onChangeMandObservation,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> mandEvents;
  final _VbmappObservationTimerState? mandObservation;
  final VoidCallback onAddMandEvent;
  final ValueChanged<_VbmappMandEvent> onSubmitMandEvent;
  final ValueChanged<int> onDeleteMandEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeMandObservation;

  @override
  Widget build(BuildContext context) {
    if (item.itemCode == 'MAND_01M' || item.itemCode == 'MAND_02M') {
      return _VbmappMand1InlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (item.itemCode == 'MAND_03M') {
      return _VbmappMand3InlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (_isTimedMandItemCode(item.itemCode)) {
      return _VbmappTimedMandInlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        observation: mandObservation,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
        onChangeObservation: onChangeMandObservation,
      );
    }
    if (_isSimpleMandRecorder(item, schema)) {
      return _VbmappMandRecorderPanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onAddEvent: onAddMandEvent,
      );
    }
    return const SizedBox.shrink();
  }
}

class _VbmappMand1InlinePanel extends StatefulWidget {
  const _VbmappMand1InlinePanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final ValueChanged<_VbmappMandEvent> onSubmitEvent;
  final ValueChanged<int> onDeleteEvent;

  @override
  State<_VbmappMand1InlinePanel> createState() =>
      _VbmappMand1InlinePanelState();
}

class _VbmappMand1InlinePanelState extends State<_VbmappMand1InlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '饼干',
    '书',
    '球',
    '泡泡',
    '音乐',
    '车',
    '秋千',
    '积木',
  ];

  final TextEditingController _requestController = TextEditingController();

  String _environment = '呈现物品';
  String _targetKind = '物品';
  String _promptChoice = '否';
  int? _selectedRecordIndex;

  bool get _usesExtraPromptRule => widget.item.itemCode == 'MAND_02M';

  int get _onePointRequestCount =>
      _scoreCountThreshold(widget.item, 1) ?? (_usesExtraPromptRule ? 4 : 2);

  int get _halfPointRequestCount =>
      _scoreCountThreshold(widget.item, .5) ?? (_usesExtraPromptRule ? 3 : 1);

  String get _recordTitle =>
      '${_vbmappQuestionDomainLabel(widget.item)}${widget.item.navCode}现场记录';

  String get _promptLabel => _usesExtraPromptRule ? '辅助' : '肢体辅助';

  List<String> get _promptValues => _usesExtraPromptRule
      ? const <String>['提问下', '自发地']
      : const <String>['否', '是'];

  String get _currentPromptChoice {
    final List<String> values = _promptValues;
    if (values.contains(_promptChoice)) {
      return _promptChoice;
    }
    return values.first;
  }

  String get _promptLevel {
    if (_usesExtraPromptRule) {
      return _currentPromptChoice;
    }
    return _currentPromptChoice == '是' ? '肢体辅助' : '无肢体辅助';
  }

  String get _requestHint => _usesExtraPromptRule ? '如：音乐、彩虹弹簧、球' : '如：饼干、书、打开';

  String get _scoreReference {
    final String promptRule =
        _usesExtraPromptRule ? '提问下仅限“你想要什么？”，自发地也计入有效要求。' : '肢体辅助不计入有效要求。';
    return '参考：0个计0分，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；$promptRule';
  }

  @override
  void didUpdateWidget(covariant _VbmappMand1InlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
  }

  @override
  void dispose() {
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCount(widget.events);
    final double suggestedScore = _suggestMandScore(widget.item, widget.events);
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool narrow = constraints.maxWidth < 780;
          final Widget form = _buildRecordForm();
          final Widget records = _buildRecordSummary();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(Icons.record_voice_over_outlined,
                      color: widget.item.color, size: 19),
                  const SizedBox(width: 7),
                  Expanded(
                    child: Text(
                      _recordTitle,
                      style: const TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: '有效',
                    value: '$qualifiedCount/$_onePointRequestCount',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '建议',
                    value: '${_formatScore(suggestedScore)}分',
                    color: widget.item.color,
                  ),
                ],
              ),
              const SizedBox(height: 9),
              if (narrow)
                Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    form,
                    const SizedBox(height: 10),
                    records,
                  ],
                )
              else
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(flex: 5, child: form),
                    const SizedBox(width: 12),
                    Expanded(flex: 4, child: records),
                  ],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildRecordForm() {
    final List<String> materials = _mandMaterialQuickPicks(
      widget.materialProfile,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _buildChoiceRow(),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: '孩子要求内容',
                hintText: _requestHint,
              ),
            ),
            const SizedBox(width: 10),
            _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: '记录本次要求',
              onTap: _submit,
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 6,
          children: <Widget>[
            for (final String material in materials)
              _VbmappMandMaterialChip(
                label: material,
                selected: _requestController.text.trim() == material,
                onTap: () => _selectMaterial(material),
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildChoiceRow() {
    final Widget environmentChoice = Expanded(
      flex: 5,
      child: _VbmappMandInlineChoiceGroup(
        label: '环境',
        value: _environment,
        values: const <String>['呈现物品', '未呈现物品'],
        onChanged: (String value) => setState(() {
          _environment = value;
        }),
      ),
    );
    final Widget targetChoice = Expanded(
      flex: 4,
      child: _VbmappMandInlineChoiceGroup(
        label: '对象',
        value: _targetKind,
        values: const <String>['物品', '动作'],
        onChanged: (String value) => setState(() {
          _targetKind = value;
        }),
      ),
    );
    final Widget promptChoice = Expanded(
      flex: _usesExtraPromptRule ? 5 : 4,
      child: _VbmappMandInlineChoiceGroup(
        label: _promptLabel,
        value: _currentPromptChoice,
        values: _promptValues,
        onChanged: (String value) => setState(() {
          _promptChoice = value;
        }),
      ),
    );
    final List<Widget> choices = _usesExtraPromptRule
        ? <Widget>[promptChoice, environmentChoice, targetChoice]
        : <Widget>[environmentChoice, targetChoice, promptChoice];
    return Row(
      children: <Widget>[
        for (int index = 0; index < choices.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 10),
          choices[index],
        ],
      ],
    );
  }

  List<String> _mandMaterialQuickPicks(VbmappMaterialProfile? profile) {
    return _smartMandQuickPicks(
      profile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
    );
  }

  Widget _buildRecordSummary() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        if (widget.materialProfile != null) ...<Widget>[
          _VbmappInlineInfo(
            icon: Icons.inventory_2_outlined,
            text: widget.materialProfile!.label,
          ),
          const SizedBox(height: 10),
        ],
        const Text(
          '有效要求记录',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        _VbmappMand1RecordGrid(
          events: widget.events,
          minSlots: _onePointRequestCount,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
        ),
        const SizedBox(height: 10),
        Text(
          _scoreReference,
          style: const TextStyle(
            color: _VbmappColors.body,
            fontSize: 12,
            height: 1.35,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  void _selectMaterial(String material) {
    setState(() {
      _requestController.text = material;
    });
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _environment,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode: '要求',
        promptLevel: _promptLevel,
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _promptChoice = _promptValues.first;
      _selectedRecordIndex = null;
    });
  }

  void _selectRecord(int index) {
    setState(() {
      _selectedRecordIndex = _selectedRecordIndex == index ? null : index;
    });
  }

  void _deleteRecord(int index) {
    setState(() {
      _selectedRecordIndex = null;
    });
    widget.onDeleteEvent(index);
  }
}

class _VbmappTimedMandInlinePanel extends StatefulWidget {
  const _VbmappTimedMandInlinePanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.observation,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
    required this.onChangeObservation,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final _VbmappObservationTimerState? observation;
  final ValueChanged<_VbmappMandEvent> onSubmitEvent;
  final ValueChanged<int> onDeleteEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeObservation;

  @override
  State<_VbmappTimedMandInlinePanel> createState() =>
      _VbmappTimedMandInlinePanelState();
}

class _VbmappTimedMandInlinePanelState
    extends State<_VbmappTimedMandInlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '球',
    '音乐',
    '彩虹弹簧',
    '出去',
    '打开',
    '秋千',
    '车',
  ];

  final TextEditingController _requestController = TextEditingController();

  Timer? _clockTimer;
  String _presentation = '呈现物品';
  String _targetKind = '物品';
  String _promptMode = '自发地';
  int? _selectedRecordIndex;

  _VbmappObservationTimerState get _observation =>
      (widget.observation ?? const _VbmappObservationTimerState())
          .withPlannedMinutes(60);

  bool get _isMand8 => widget.item.itemCode == 'MAND_08M';

  bool get _isMand9 => widget.item.itemCode == 'MAND_09M';

  int get _onePointRequestCount => _scoreCountThreshold(widget.item, 1) ?? 5;

  int get _halfPointRequestCount => _scoreCountThreshold(widget.item, .5) ?? 2;

  int get _plannedMinutes => _plannedMinutesForMandItem(widget.item);

  int get _plannedSeconds => _plannedMinutes * Duration.secondsPerMinute;

  int get _elapsedSeconds => _observation.elapsedSecondsAt(DateTime.now());

  int get _remainingSeconds {
    final int remaining = _plannedSeconds - _elapsedSeconds;
    return remaining > 0 ? remaining : 0;
  }

  bool get _observationMet => _elapsedSeconds >= _plannedSeconds;

  int get _displayMinSlots => _isMand9 ? 6 : _onePointRequestCount;

  String get _recordTitle => '提要求${widget.item.navCode}观察记录';

  String get _scoreReference {
    if (_isMand8) {
      return '参考：60分钟观察窗内，$_halfPointRequestCount个不同要求计0.5分，'
          '$_onePointRequestCount个不同要求且系统判定至少2条为双词+计1分。';
    }
    if (_isMand9) {
      return '参考：30分钟观察窗内，$_halfPointRequestCount个自发不同要求计0.5分，'
          '$_onePointRequestCount个自发不同要求计1分。';
    }
    return '参考：60分钟观察窗内，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；仅统计呈现物品条件下的自发要求。';
  }

  String get _observationHint {
    final int qualifiedCount = _qualifiedMandCountForItem(
      widget.item,
      widget.events,
      observation: widget.observation,
    );
    if (!_observation.hasStarted) {
      return '先开启${_plannedMinutes}分钟观察窗，再连续记录孩子的自然要求。';
    }
    if (!_observationMet && qualifiedCount >= _onePointRequestCount) {
      return _isMand8
          ? '已达到数量条件，但观察未满${_plannedMinutes}分钟；双词结构由系统根据原话自动判定。'
          : '已达到1分数量条件，但观察未满${_plannedMinutes}分钟，可继续观察或由老师确认结束。';
    }
    if (!_observationMet && qualifiedCount >= _halfPointRequestCount) {
      return '已达到0.5分数量条件，但观察未满${_plannedMinutes}分钟，建议继续观察。';
    }
    if (_observationMet) {
      return '${_plannedMinutes}分钟观察窗已完成，系统会继续保留新增记录供老师判断。';
    }
    if (_isMand8) {
      return '记录自然情境下的不同要求；双词结构由系统根据原话自动判定。';
    }
    if (_isMand9) {
      return '记录30分钟内的自发不同要求，系统自动去重统计。';
    }
    return '记录自然情境下的自发要求，系统自动统计不同要求数量。';
  }

  @override
  void initState() {
    super.initState();
    _syncClockTicker();
  }

  @override
  void didUpdateWidget(covariant _VbmappTimedMandInlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
    if (oldWidget.observation != widget.observation) {
      _syncClockTicker();
    }
  }

  @override
  void dispose() {
    _clockTimer?.cancel();
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCountForItem(
      widget.item,
      widget.events,
      observation: widget.observation,
    );
    final double suggestedScore = _suggestMandScore(
      widget.item,
      widget.events,
      observation: widget.observation,
    );
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool narrow = constraints.maxWidth < 860;
          final Widget form = _buildRecordForm();
          final Widget summary = _buildRecordSummary();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(Icons.timer_outlined,
                      color: widget.item.color, size: 19),
                  const SizedBox(width: 7),
                  Expanded(
                    child: Text(
                      _recordTitle,
                      style: const TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: '观察',
                    value: _vbmappDurationText(_elapsedSeconds),
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '不同',
                    value: '$qualifiedCount/$_onePointRequestCount',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '建议',
                    value: '${_formatScore(suggestedScore)}分',
                    color: widget.item.color,
                  ),
                ],
              ),
              const SizedBox(height: 10),
              _buildObservationBar(),
              const SizedBox(height: 10),
              if (narrow)
                Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    form,
                    const SizedBox(height: 10),
                    summary,
                  ],
                )
              else
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(flex: 5, child: form),
                    const SizedBox(width: 12),
                    Expanded(flex: 4, child: summary),
                  ],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildObservationBar() {
    final bool running = _observation.isRunning;
    final bool ended = _observation.ended;
    final String primaryActionLabel = !_observation.hasStarted
        ? '开始观察'
        : ended
            ? '重新观察'
            : running
                ? '暂停'
                : '继续观察';
    final IconData primaryActionIcon = !_observation.hasStarted
        ? Icons.play_arrow_rounded
        : ended
            ? Icons.replay_rounded
            : running
                ? Icons.pause_rounded
                : Icons.play_arrow_rounded;
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(11),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: <Widget>[
                    _VbmappMand4TimerMetricChip(
                      label: '观察窗',
                      value: '${_plannedMinutes}分钟',
                      tone: widget.item.color,
                    ),
                    _VbmappMand4TimerMetricChip(
                      label: '已观察',
                      value: _vbmappDurationText(_elapsedSeconds),
                      tone: widget.item.color,
                    ),
                    _VbmappMand4TimerMetricChip(
                      label: '剩余',
                      value: _vbmappDurationText(_remainingSeconds),
                      tone: _remainingSeconds == 0
                          ? _VbmappColors.green
                          : _VbmappColors.orangeDeep,
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-mand4-primary-timer'),
                    icon: primaryActionIcon,
                    label: primaryActionLabel,
                    filled: true,
                    onTap: _handlePrimaryTimerAction,
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-mand4-stop-timer'),
                    icon: Icons.stop_rounded,
                    label: '结束',
                    onTap: _observation.hasStarted ? _finishObservation : null,
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            _observationHint,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1.35,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRecordForm() {
    final List<String> materials =
        _mandMaterialQuickPicks(widget.materialProfile);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _buildChoiceRow(),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: _isMand8 ? '孩子要求原话' : '孩子要求内容',
                hintText: _isMand8 ? '如：跑快点、该我了、倒果汁' : '如：泡泡、出去、打开',
              ),
            ),
            const SizedBox(width: 10),
            _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: '记录本次要求',
              onTap: _submit,
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 6,
          children: <Widget>[
            for (final String material in materials)
              _VbmappMandMaterialChip(
                label: material,
                selected: _requestController.text.trim() == material,
                onTap: () => _selectMaterial(material),
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildChoiceRow() {
    final Widget presentationChoice = _VbmappMandInlineChoiceGroup(
      label: '目标呈现',
      value: _presentation,
      values: const <String>['呈现物品', '未呈现物品'],
      onChanged: (String value) => setState(() {
        _presentation = value;
      }),
    );
    final Widget targetChoice = _VbmappMandInlineChoiceGroup(
      label: '目标',
      value: _targetKind,
      values: const <String>['物品', '动作', '活动'],
      onChanged: (String value) => setState(() {
        _targetKind = value;
      }),
    );
    if (_isMand8) {
      return Row(
        children: <Widget>[
          Expanded(
            flex: 4,
            child: _VbmappMandInlineChoiceGroup(
              label: '诱发',
              value: _promptMode,
              values: const <String>['自发地', '提问下'],
              onChanged: (String value) => setState(() {
                _promptMode = value;
              }),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(flex: 6, child: presentationChoice),
          const SizedBox(width: 10),
          Expanded(flex: 5, child: targetChoice),
        ],
      );
    }
    return Row(
      children: <Widget>[
        Expanded(flex: 5, child: presentationChoice),
        const SizedBox(width: 10),
        Expanded(flex: 5, child: targetChoice),
      ],
    );
  }

  List<String> _mandMaterialQuickPicks(VbmappMaterialProfile? profile) {
    return _smartMandQuickPicks(
      profile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
    );
  }

  Widget _buildRecordSummary() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        if (widget.materialProfile != null) ...<Widget>[
          _VbmappInlineInfo(
            icon: Icons.inventory_2_outlined,
            text: widget.materialProfile!.label,
          ),
          const SizedBox(height: 10),
        ],
        const Text(
          '观察记录',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        _VbmappMand1RecordGrid(
          events: widget.events,
          minSlots: _displayMinSlots,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
          isQualified: (_VbmappMandEvent event) => _mandEventCountsForItem(
            widget.item,
            event,
            observation: widget.observation,
          ),
          metaTextBuilder: (_VbmappMandEvent event) =>
              _mandRecordMetaText(event, item: widget.item),
        ),
        const SizedBox(height: 10),
        Text(
          _scoreReference,
          style: const TextStyle(
            color: _VbmappColors.body,
            fontSize: 12,
            height: 1.35,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  void _syncClockTicker() {
    _clockTimer?.cancel();
    _clockTimer = null;
    if (!_observation.isRunning) {
      return;
    }
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (Timer timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {});
    });
  }

  void _handlePrimaryTimerAction() {
    final DateTime now = DateTime.now();
    if (!_observation.hasStarted || _observation.ended) {
      widget.onChangeObservation(_observation.restart(now));
      return;
    }
    if (_observation.isRunning) {
      widget.onChangeObservation(_observation.pause(now));
      return;
    }
    widget.onChangeObservation(_observation.resume(now));
  }

  void _finishObservation() {
    _confirmFinishObservation();
  }

  Future<void> _confirmFinishObservation() async {
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
    widget.onChangeObservation(_observation.finish(DateTime.now()));
  }

  void _selectMaterial(String material) {
    setState(() {
      _requestController.text = material;
    });
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    if (!_observation.hasStarted) {
      widget.onChangeObservation(_observation.start(DateTime.now()));
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _presentation,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode:
            _isMand8 ? (_promptMode == '提问下' ? '提问下要求' : '自发要求') : '自发要求',
        promptLevel: _isMand8 ? _promptMode : '',
        phraseLevel: '',
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      if (_isMand8) {
        _promptMode = '自发地';
      }
    });
  }

  void _selectRecord(int index) {
    setState(() {
      _selectedRecordIndex = _selectedRecordIndex == index ? null : index;
    });
  }

  void _deleteRecord(int index) {
    setState(() {
      _selectedRecordIndex = null;
    });
    widget.onDeleteEvent(index);
  }
}

class _VbmappMand3InlinePanel extends StatefulWidget {
  const _VbmappMand3InlinePanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final ValueChanged<_VbmappMandEvent> onSubmitEvent;
  final ValueChanged<int> onDeleteEvent;

  @override
  State<_VbmappMand3InlinePanel> createState() =>
      _VbmappMand3InlinePanelState();
}

class _VbmappMand3InlinePanelState extends State<_VbmappMand3InlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '音乐',
    '球',
    '彩虹弹簧',
    '秋千',
    '车',
    '饼干',
    '书',
  ];
  static const List<String> _fallbackPeople = <String>[
    '爸爸',
    '妈妈',
    '老师',
    '治疗师',
    '爷爷',
    '奶奶',
  ];
  static const List<String> _fallbackSettings = <String>[
    '屋里',
    '屋外',
    '教室',
    '游戏区',
    '桌面',
    '走廊',
  ];
  static const List<String> _fallbackExamples = <String>[
    '红瓶泡泡',
    '蓝瓶泡泡',
    '大球',
    '小球',
    '红车',
    '蓝车',
  ];

  final TextEditingController _requestController = TextEditingController();
  final TextEditingController _dimensionController = TextEditingController();

  String _dimension = '人物';
  String _promptChoice = '提问下';
  int? _selectedRecordIndex;

  String get _dimensionHint {
    switch (_dimension) {
      case '环境':
        return '如：屋里、屋外';
      case '例子':
        return '如：红瓶泡泡、蓝瓶泡泡';
      case '人物':
      default:
        return '如：爸爸、妈妈';
    }
  }

  @override
  void initState() {
    super.initState();
    _requestController.addListener(_handleRequestTextChanged);
  }

  @override
  void didUpdateWidget(covariant _VbmappMand3InlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
  }

  @override
  void dispose() {
    _requestController.removeListener(_handleRequestTextChanged);
    _requestController.dispose();
    _dimensionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final Map<String, int> counts = _mandGeneralizationCounts(widget.events);
    final double suggestedScore = _suggestMandScore(widget.item, widget.events);
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool narrow = constraints.maxWidth < 780;
          final Widget form = _buildRecordForm();
          final Widget summary = _buildGeneralizationSummary(counts);
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(Icons.hub_outlined, color: widget.item.color, size: 19),
                  const SizedBox(width: 7),
                  const Expanded(
                    child: Text(
                      '提要求3M泛化记录',
                      style: TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: '人物',
                    value: '${counts['people'] ?? 0}/2',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '环境',
                    value: '${counts['settings'] ?? 0}/2',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '例子',
                    value: '${counts['examples'] ?? 0}/2',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '建议',
                    value: '${_formatScore(suggestedScore)}分',
                    color: widget.item.color,
                  ),
                ],
              ),
              const SizedBox(height: 9),
              if (narrow)
                Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    form,
                    const SizedBox(height: 10),
                    summary,
                  ],
                )
              else
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(flex: 5, child: form),
                    const SizedBox(width: 12),
                    Expanded(flex: 4, child: summary),
                  ],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildRecordForm() {
    final List<String> materials = _mand3MaterialQuickPicks(
      widget.materialProfile,
    );
    final List<String> dimensionQuickPicks = _dimensionQuickPicks();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              flex: 6,
              child: _VbmappMandInlineChoiceGroup(
                label: '本次记录',
                value: _dimension,
                values: const <String>['人物', '环境', '例子'],
                onChanged: (String value) => setState(() {
                  _dimension = value;
                  _dimensionController.clear();
                }),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              flex: 4,
              child: _VbmappMandInlineChoiceGroup(
                label: '辅助',
                value: _promptChoice,
                values: const <String>['提问下', '自发地'],
                onChanged: (String value) => setState(() {
                  _promptChoice = value;
                }),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _dimensionController,
                label: _dimension,
                hintText: _dimensionHint,
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: '要求内容',
                hintText: '如：泡泡、音乐、球',
              ),
            ),
            const SizedBox(width: 10),
            _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: '记录本次要求',
              onTap: _submit,
            ),
          ],
        ),
        const SizedBox(height: 8),
        _buildQuickPickBar(
          dimensionValues: dimensionQuickPicks,
          materialValues: materials,
        ),
      ],
    );
  }

  Widget _buildQuickPickBar({
    required List<String> dimensionValues,
    required List<String> materialValues,
  }) {
    if (dimensionValues.isEmpty && materialValues.isEmpty) {
      return const SizedBox.shrink();
    }
    return SizedBox(
      height: 30,
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: <Widget>[
            if (dimensionValues.isNotEmpty) ...<Widget>[
              _VbmappMandQuickPickLabel(label: _dimension),
              const SizedBox(width: 5),
              for (final String value in dimensionValues) ...<Widget>[
                _VbmappMandCompactChip(
                  label: value,
                  selected: _dimensionController.text.trim() == value,
                  onTap: () => _selectDimensionValue(value),
                ),
                const SizedBox(width: 5),
              ],
            ],
            if (dimensionValues.isNotEmpty &&
                materialValues.isNotEmpty) ...<Widget>[
              Container(
                width: 1,
                height: 18,
                margin: const EdgeInsets.symmetric(horizontal: 4),
                color: _VbmappColors.lineSoft,
              ),
              const SizedBox(width: 2),
            ],
            if (materialValues.isNotEmpty) ...<Widget>[
              const _VbmappMandQuickPickLabel(label: '要求'),
              const SizedBox(width: 5),
              for (final String material in materialValues) ...<Widget>[
                _VbmappMandCompactChip(
                  label: material,
                  selected: _requestController.text.trim() == material,
                  onTap: () => _selectMaterial(material),
                ),
                const SizedBox(width: 5),
              ],
            ],
          ],
        ),
      ),
    );
  }

  List<String> _mand3MaterialQuickPicks(VbmappMaterialProfile? profile) {
    final List<String> configured =
        profile?.quickPickLabels ?? const <String>[];
    if (configured.isEmpty) {
      return _fallbackMaterials;
    }
    final List<String> preferred = <String>[
      '泡泡',
      '音乐',
      '球',
      '彩虹弹簧',
      '秋千',
      ...configured,
    ];
    final Set<String> seen = <String>{};
    return preferred
        .where((String value) {
          final String normalized = value.trim();
          if (normalized.isEmpty || seen.contains(normalized)) {
            return false;
          }
          seen.add(normalized);
          return true;
        })
        .take(8)
        .toList(growable: false);
  }

  List<String> _dimensionQuickPicks() {
    switch (_dimension) {
      case '环境':
        return _configuredDimensionQuickPicks(
          'mand3_settings',
          _fallbackSettings,
        );
      case '例子':
        final String key = _exampleQuickPickFieldKey();
        final List<String> configured = _configuredDimensionQuickPicks(
          key,
          const <String>[],
        );
        if (configured.isNotEmpty) {
          return configured;
        }
        return _configuredDimensionQuickPicks(
          'mand3_examples_default',
          _fallbackExamples,
        );
      case '人物':
      default:
        return _configuredDimensionQuickPicks('mand3_people', _fallbackPeople);
    }
  }

  List<String> _configuredDimensionQuickPicks(
    String fieldKey,
    List<String> fallback,
  ) {
    final List<String> configured =
        widget.materialProfile?.quickPicksFor(fieldKey) ?? const <String>[];
    final List<String> source = configured.isEmpty ? fallback : configured;
    return source.take(6).toList(growable: false);
  }

  String _exampleQuickPickFieldKey() {
    final String request = _requestController.text.trim();
    if (request.contains('泡泡')) {
      return 'mand3_examples_bubbles';
    }
    if (request.contains('球')) {
      return 'mand3_examples_ball';
    }
    if (request.contains('车')) {
      return 'mand3_examples_car';
    }
    if (request.contains('音乐') || request.contains('歌')) {
      return 'mand3_examples_music';
    }
    if (request.contains('彩虹') || request.contains('弹簧')) {
      return 'mand3_examples_slinky';
    }
    if (request.contains('秋千')) {
      return 'mand3_examples_swing';
    }
    return 'mand3_examples_default';
  }

  Widget _buildGeneralizationSummary(Map<String, int> counts) {
    final Map<String, List<String>> values =
        _mandGeneralizationValues(widget.events);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        if (widget.materialProfile != null) ...<Widget>[
          _VbmappInlineInfo(
            icon: Icons.inventory_2_outlined,
            text: widget.materialProfile!.label,
          ),
          const SizedBox(height: 10),
        ],
        const Text(
          '达标进度',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        Row(
          children: <Widget>[
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '人物',
                values: values['people'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '环境',
                values: values['settings'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '例子',
                values: values['examples'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        _VbmappMand3RecordList(
          events: widget.events,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
        ),
        const SizedBox(height: 10),
        Text(
          '参考：资料要求同一强化物覆盖2个人、2个环境、2个不同例子；各1个计0.5分，各2个计1分。',
          style: const TextStyle(
            color: _VbmappColors.body,
            fontSize: 12,
            height: 1.35,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  void _selectMaterial(String material) {
    setState(() {
      _requestController.text = material;
    });
  }

  void _selectDimensionValue(String value) {
    setState(() {
      _dimensionController.text = value;
    });
  }

  void _handleRequestTextChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  void _submit() {
    final String request = _requestController.text.trim();
    final String dimensionValue = _dimensionController.text.trim();
    if (request.isEmpty || dimensionValue.isEmpty) {
      return;
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        person: _dimension == '人物' ? dimensionValue : '',
        setting: _dimension == '环境' ? dimensionValue : '',
        example: _dimension == '例子' ? dimensionValue : '',
        responseMode: '要求',
        promptLevel: _promptChoice,
        functional: true,
      ),
    );
    _requestController.clear();
    _dimensionController.clear();
    setState(() {
      _selectedRecordIndex = null;
    });
  }

  void _selectRecord(int index) {
    setState(() {
      _selectedRecordIndex = _selectedRecordIndex == index ? null : index;
    });
  }

  void _deleteRecord(int index) {
    setState(() {
      _selectedRecordIndex = null;
    });
    widget.onDeleteEvent(index);
  }
}

class _VbmappMandQuickPickLabel extends StatelessWidget {
  const _VbmappMandQuickPickLabel({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Text(
      label,
      style: const TextStyle(
        color: _VbmappColors.muted,
        fontSize: 11,
        fontWeight: FontWeight.w900,
      ),
    );
  }
}

class _VbmappMandCompactChip extends StatelessWidget {
  const _VbmappMandCompactChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          height: 28,
          padding: const EdgeInsets.symmetric(horizontal: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : const Color(0xFFFFFCFA),
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
                fontSize: 11.5,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand3CoverageColumn extends StatelessWidget {
  const _VbmappMand3CoverageColumn({
    required this.label,
    required this.values,
    required this.color,
  });

  final String label;
  final List<String> values;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final int filledCount = values.length > 2 ? 2 : values.length;
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Text(
            '$label $filledCount/2',
            style: TextStyle(
              color: filledCount >= 2 ? color : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 7),
          for (int index = 0; index < 2; index++) ...<Widget>[
            _VbmappMand3CoverageSlot(
              text: index < values.length ? values[index] : '待记录',
              filled: index < values.length,
              color: color,
            ),
            if (index == 0) const SizedBox(height: 6),
          ],
        ],
      ),
    );
  }
}

class _VbmappMand3CoverageSlot extends StatelessWidget {
  const _VbmappMand3CoverageSlot({
    required this.text,
    required this.filled,
    required this.color,
  });

  final String text;
  final bool filled;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      alignment: Alignment.center,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      decoration: BoxDecoration(
        color: filled ? color.withOpacity(.09) : const Color(0xFFFFF6EF),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: filled ? color.withOpacity(.24) : _VbmappColors.lineSoft,
        ),
      ),
      child: Text(
        text,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: filled ? color : _VbmappColors.muted,
          fontSize: 11.5,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappMand3RecordList extends StatelessWidget {
  const _VbmappMand3RecordList({
    required this.events,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
  });

  final List<_VbmappMandEvent> events;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;

  @override
  Widget build(BuildContext context) {
    final int itemCount = events.length < 6 ? 6 : events.length;
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool twoColumns = constraints.maxWidth >= 340;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: 7),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              Row(
                children: <Widget>[
                  Expanded(child: _buildSlot(index)),
                  if (index + 1 < itemCount) ...<Widget>[
                    const SizedBox(width: 8),
                    Expanded(child: _buildSlot(index + 1)),
                  ],
                ],
              ),
              if (index + 2 < itemCount) const SizedBox(height: 7),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand3RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
    );
  }
}

class _VbmappMand3RecordSlot extends StatelessWidget {
  const _VbmappMand3RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row?.isQualified ?? false;
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String dimensionText = row == null ? '' : _mand3DimensionText(row);
    return Material(
      key: ValueKey<String>('vbmapp-mand3-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 22,
                child: Text(
                  '${index + 1}',
                  style: TextStyle(
                    color: filled ? accent : _VbmappColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Expanded(
                child: filled
                    ? Row(
                        children: <Widget>[
                          Flexible(
                            child: Text(
                              _mandRequestText(row),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                          Container(
                            width: 1,
                            height: 14,
                            margin: const EdgeInsets.symmetric(horizontal: 7),
                            color: _VbmappColors.lineSoft,
                          ),
                          Flexible(
                            child: Text(
                              dimensionText,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                        ],
                      )
                    : const Text(
                        '等待记录',
                        style: TextStyle(
                          color: _VbmappColors.muted,
                          fontSize: 12.5,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
              ),
              const SizedBox(width: 8),
              _VbmappMandRecordActions(
                filled: filled,
                qualified: qualified,
                selected: selected,
                accent: accent,
                onDelete: onDelete,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineChoiceGroup extends StatelessWidget {
  const _VbmappMandInlineChoiceGroup({
    required this.label,
    required this.value,
    required this.values,
    required this.onChanged,
  });

  final String label;
  final String value;
  final List<String> values;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.max,
      children: <Widget>[
        Text(
          label,
          maxLines: 1,
          softWrap: false,
          style: const TextStyle(
            color: _VbmappColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 7),
        Expanded(
          child: Row(
            children: <Widget>[
              for (int index = 0; index < values.length; index++) ...<Widget>[
                if (index > 0) const SizedBox(width: 5),
                Expanded(
                  child: _VbmappMandChoiceButton(
                    label: values[index],
                    selected: values[index] == value,
                    onTap: () => onChanged(values[index]),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _VbmappMandChoiceButton extends StatelessWidget {
  const _VbmappMandChoiceButton({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: selected ? _VbmappColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              softWrap: false,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: selected ? Colors.white : _VbmappColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineTextField extends StatelessWidget {
  const _VbmappMandInlineTextField({
    required this.controller,
    required this.label,
    required this.hintText,
  });

  final TextEditingController controller;
  final String label;
  final String hintText;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      onTapOutside: (_) => FocusManager.instance.primaryFocus?.unfocus(),
      style: const TextStyle(
        color: _VbmappColors.ink,
        fontSize: 14,
        fontWeight: FontWeight.w800,
      ),
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        floatingLabelBehavior: FloatingLabelBehavior.auto,
        labelStyle: const TextStyle(
          color: Color(0xFFB8A79E),
          fontSize: 13,
          fontWeight: FontWeight.w600,
        ),
        floatingLabelStyle: const TextStyle(
          color: _VbmappColors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
        hintStyle: const TextStyle(
          color: Color(0xFFC7B9B1),
          fontSize: 13,
          fontWeight: FontWeight.w500,
        ),
        isDense: true,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        filled: true,
        fillColor: Colors.white,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.orange),
        ),
      ),
    );
  }
}

class _VbmappMandMaterialChip extends StatelessWidget {
  const _VbmappMandMaterialChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : Colors.white,
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand1RecordGrid extends StatelessWidget {
  const _VbmappMand1RecordGrid({
    required this.events,
    required this.minSlots,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
    this.isQualified,
    this.metaTextBuilder,
  });

  final List<_VbmappMandEvent> events;
  final int minSlots;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;
  final bool Function(_VbmappMandEvent event)? isQualified;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        const double spacing = 8;
        final int itemCount =
            events.length < minSlots ? minSlots : events.length;
        final bool twoColumns = constraints.maxWidth >= 360;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: spacing),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              if (index + 1 < itemCount)
                IntrinsicHeight(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: <Widget>[
                      Expanded(child: _buildSlot(index)),
                      const SizedBox(width: spacing),
                      Expanded(child: _buildSlot(index + 1)),
                    ],
                  ),
                )
              else
                _buildSlot(index),
              if (index + 2 < itemCount) const SizedBox(height: spacing),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand1RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
      qualificationResolver: isQualified,
      metaTextBuilder: metaTextBuilder,
    );
  }
}

class _VbmappMand1RecordSlot extends StatelessWidget {
  const _VbmappMand1RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
    this.qualificationResolver,
    this.metaTextBuilder,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;
  final bool Function(_VbmappMandEvent event)? qualificationResolver;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row == null
        ? false
        : (qualificationResolver?.call(row) ?? row.isQualified);
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String requestText = row == null ? '' : _mandRequestText(row);
    final BorderRadius radius = BorderRadius.circular(10);
    return Material(
      key: ValueKey<String>('vbmapp-mand-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: radius,
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: radius,
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: ConstrainedBox(
            constraints: const BoxConstraints(minHeight: 34),
            child: Row(
              children: <Widget>[
                Container(
                  width: 28,
                  height: 28,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: filled
                        ? accent.withOpacity(.12)
                        : const Color(0xFFFFF6EF),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    '${index + 1}',
                    style: TextStyle(
                      color: filled ? accent : _VbmappColors.muted,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: filled
                      ? Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            Text(
                              requestText,
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 13,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                            const SizedBox(height: 3),
                            Text(
                              metaTextBuilder?.call(row) ??
                                  _mandRecordMetaText(row),
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.body,
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ],
                        )
                      : const Text(
                          '等待记录一条有效要求',
                          style: TextStyle(
                            color: _VbmappColors.muted,
                            fontSize: 13,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                ),
                const SizedBox(width: 8),
                _VbmappMandRecordActions(
                  filled: filled,
                  qualified: qualified,
                  selected: selected,
                  accent: accent,
                  onDelete: onDelete,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandRecordActions extends StatelessWidget {
  const _VbmappMandRecordActions({
    required this.filled,
    required this.qualified,
    required this.selected,
    required this.accent,
    required this.onDelete,
  });

  final bool filled;
  final bool qualified;
  final bool selected;
  final Color accent;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    if (!selected || !filled) {
      return Text(
        filled ? (qualified ? '计入' : '不计') : '-',
        style: TextStyle(
          color: accent,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      );
    }
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Text(
          qualified ? '计入' : '不计',
          style: TextStyle(
            color: accent,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 6),
        Material(
          color: Colors.transparent,
          child: InkWell(
            key: const ValueKey<String>('vbmapp-mand-delete-record'),
            onTap: onDelete,
            borderRadius: BorderRadius.circular(999),
            child: Ink(
              height: 22,
              padding: const EdgeInsets.symmetric(horizontal: 7),
              decoration: BoxDecoration(
                color: const Color(0xFFFFE6D9),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(color: _VbmappColors.orange),
              ),
              child: const Center(
                child: Text(
                  '删除',
                  style: TextStyle(
                    color: _VbmappColors.orangeDeep,
                    fontSize: 11,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappMandRecorderPanel extends StatelessWidget {
  const _VbmappMandRecorderPanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onAddEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final VoidCallback onAddEvent;

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCount(events);
    final double suggestedScore = _suggestMandScore(item, events);
    final bool generalizationMode = item.itemCode == 'MAND_03M';
    final Map<String, int> generalizationCounts =
        _mandGeneralizationCounts(events);
    final List<_VbmappMandEvent> visibleEvents =
        events.length <= 4 ? events : events.sublist(events.length - 4);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Icon(Icons.record_voice_over_outlined,
                  color: item.color, size: 21),
              const SizedBox(width: 8),
              const Expanded(
                child: Text(
                  '提要求事件记录',
                  style: TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _VbmappEvidenceMetric(
                label: '有效',
                value: '$qualifiedCount',
                color: item.color,
              ),
              const SizedBox(width: 8),
              if (generalizationMode) ...<Widget>[
                _VbmappEvidenceMetric(
                  label: '人',
                  value: '${generalizationCounts['people'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '环境',
                  value: '${generalizationCounts['settings'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '例子',
                  value: '${generalizationCounts['examples'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
              ],
              _VbmappEvidenceMetric(
                label: '建议',
                value: _formatScore(suggestedScore),
                color: item.color,
              ),
            ],
          ),
          if (materialProfile != null) ...<Widget>[
            const SizedBox(height: 10),
            _VbmappInlineInfo(
              icon: Icons.inventory_2_outlined,
              text: materialProfile!.label,
            ),
          ],
          const SizedBox(height: 12),
          if (visibleEvents.isEmpty)
            const _VbmappEmptyEvidence(text: '还没有记录孩子实际发出的要求')
          else
            Column(
              children: <Widget>[
                for (final _VbmappMandEvent event in visibleEvents)
                  _VbmappMandEventRow(event: event),
              ],
            ),
          const SizedBox(height: 12),
          Align(
            alignment: Alignment.centerLeft,
            child: _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: generalizationMode ? '记录一次泛化要求' : '记录一次要求',
              onTap: onAddEvent,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappMandEventRow extends StatelessWidget {
  const _VbmappMandEventRow({required this.event});

  final _VbmappMandEvent event;

  @override
  Widget build(BuildContext context) {
    final Color color =
        event.isQualified ? _VbmappColors.green : _VbmappColors.muted;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            event.isQualified
                ? Icons.check_circle_outline_rounded
                : Icons.radio_button_unchecked_rounded,
            color: color,
            size: 18,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              event.summary,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Text(
            event.promptLevel,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappEvidenceMetric extends StatelessWidget {
  const _VbmappEvidenceMetric({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappInlineInfo extends StatelessWidget {
  const _VbmappInlineInfo({
    required this.icon,
    required this.text,
  });

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Icon(icon, color: _VbmappColors.orange, size: 17),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappEmptyEvidence extends StatelessWidget {
  const _VbmappEmptyEvidence({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 13),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _VbmappColors.muted,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _VbmappSmallActionButton extends StatelessWidget {
  const _VbmappSmallActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: _VbmappColors.orange,
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(icon, size: 18, color: Colors.white),
              const SizedBox(width: 6),
              Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
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

class _VbmappMand4TimerButton extends StatelessWidget {
  const _VbmappMand4TimerButton({
    required this.icon,
    required this.label,
    required this.onTap,
    this.filled = false,
    this.compact = false,
    super.key,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onTap;
  final bool filled;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _VbmappColors.orangeDeep
        : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: compact ? 28 : 32,
          padding: EdgeInsets.symmetric(horizontal: compact ? 8 : 10),
          decoration: BoxDecoration(
            color: filled && enabled
                ? _VbmappColors.orange
                : enabled
                    ? const Color(0xFFFFFCFA)
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Center(
            child: Row(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: compact ? 14 : 16, color: textColor),
                SizedBox(width: compact ? 4 : 5),
                Text(
                  label,
                  style: TextStyle(
                    color: textColor,
                    fontSize: compact ? 11.5 : 12.5,
                    height: 1,
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

class _VbmappMand4TimerMetricChip extends StatelessWidget {
  const _VbmappMand4TimerMetricChip({
    required this.label,
    required this.value,
    required this.tone,
  });

  final String label;
  final String value;
  final Color tone;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: tone.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: tone.withOpacity(.18)),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: tone,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappMand4QuickRecordDialog extends StatefulWidget {
  const _VbmappMand4QuickRecordDialog({required this.materialProfile});

  final VbmappMaterialProfile? materialProfile;

  @override
  State<_VbmappMand4QuickRecordDialog> createState() =>
      _VbmappMand4QuickRecordDialogState();
}

class _VbmappMand4QuickRecordDialogState
    extends State<_VbmappMand4QuickRecordDialog> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '球',
    '音乐',
    '彩虹弹簧',
    '出去',
    '打开',
    '秋千',
    '车',
  ];

  final TextEditingController _requestController = TextEditingController();
  String _promptMode = '自发地';
  String _presentation = '呈现物品';
  String _targetKind = '物品';

  @override
  void dispose() {
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final List<String> materials = _smartMandQuickPicks(
      widget.materialProfile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
    );
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 760,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              const Text(
                '补记一条提要求观察记录',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 14),
              Row(
                children: <Widget>[
                  Expanded(
                    flex: 4,
                    child: _VbmappMandInlineChoiceGroup(
                      label: '诱发',
                      value: _promptMode,
                      values: const <String>['自发地', '提问下'],
                      onChanged: (String value) => setState(() {
                        _promptMode = value;
                      }),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    flex: 6,
                    child: _VbmappMandInlineChoiceGroup(
                      label: '目标呈现',
                      value: _presentation,
                      values: const <String>['呈现物品', '未呈现物品'],
                      onChanged: (String value) => setState(() {
                        _presentation = value;
                      }),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    flex: 5,
                    child: _VbmappMandInlineChoiceGroup(
                      label: '目标',
                      value: _targetKind,
                      values: const <String>['物品', '动作', '活动'],
                      onChanged: (String value) => setState(() {
                        _targetKind = value;
                      }),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              _VbmappMandInlineTextField(
                controller: _requestController,
                label: '孩子要求原话',
                hintText: '如：泡泡、出去、该我了',
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 8,
                runSpacing: 6,
                children: <Widget>[
                  for (final String material in materials)
                    _VbmappMandMaterialChip(
                      label: material,
                      selected: _requestController.text.trim() == material,
                      onTap: () => setState(() {
                        _requestController.text = material;
                      }),
                    ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    icon: Icons.close_rounded,
                    label: '取消',
                    onTap: () => Navigator.of(context).pop(),
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-global-mand4-submit'),
                    icon: Icons.check_rounded,
                    label: '保存记录',
                    filled: true,
                    onTap: _submit,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    Navigator.of(context).pop(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _presentation,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode: _promptMode == '提问下' ? '提问下要求' : '自发要求',
        promptLevel: _promptMode,
        phraseLevel: '',
        functional: true,
      ),
    );
  }
}

class _VbmappObservationFinishConfirmDialog extends StatelessWidget {
  const _VbmappObservationFinishConfirmDialog();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 480,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              const Text(
                '确认结束观察？',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                '结束后会保留当前计时和记录。如只是暂时离开，建议点暂停。',
                style: TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 13,
                  height: 1.45,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    icon: Icons.close_rounded,
                    label: '取消',
                    onTap: () => Navigator.of(context).pop(false),
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>(
                        'vbmapp-confirm-finish-observation'),
                    icon: Icons.stop_rounded,
                    label: '确认结束',
                    filled: true,
                    onTap: () => Navigator.of(context).pop(true),
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

class _VbmappMandEventDialog extends StatefulWidget {
  const _VbmappMandEventDialog({required this.generalizationMode});

  final bool generalizationMode;

  @override
  State<_VbmappMandEventDialog> createState() => _VbmappMandEventDialogState();
}

class _VbmappMandEventDialogState extends State<_VbmappMandEventDialog> {
  final TextEditingController _utteranceController = TextEditingController();
  final TextEditingController _targetController = TextEditingController();
  final TextEditingController _contextController = TextEditingController();
  final TextEditingController _personController = TextEditingController();
  final TextEditingController _settingController = TextEditingController();
  final TextEditingController _exampleController = TextEditingController();

  String _responseMode = '口语';
  String _promptLevel = '无辅助';
  bool _functional = true;

  @override
  void dispose() {
    _utteranceController.dispose();
    _targetController.dispose();
    _contextController.dispose();
    _personController.dispose();
    _settingController.dispose();
    _exampleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: widget.generalizationMode ? 720 : 560,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Text(
                widget.generalizationMode ? '记录一次泛化要求' : '记录一次提要求',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 14),
              _VbmappDialogTextField(
                controller: _utteranceController,
                label: '孩子发出的词语 / 手语 / 图片交换',
                hintText: '例如：饼干、球、打开',
              ),
              const SizedBox(height: 10),
              _VbmappDialogTextField(
                controller: _targetController,
                label: '要求的物品或活动',
                hintText: '例如：饼干、泡泡、秋千',
              ),
              const SizedBox(height: 10),
              _VbmappDialogTextField(
                controller: _contextController,
                label: '动机情境',
                hintText: '例如：看到饼干但拿不到',
              ),
              if (widget.generalizationMode) ...<Widget>[
                const SizedBox(height: 10),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _personController,
                        label: '互动对象',
                        hintText: '例如：妈妈、老师',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _settingController,
                        label: '环境',
                        hintText: '例如：教室、户外',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _exampleController,
                        label: '不同例子',
                        hintText: '例如：红泡泡、蓝泡泡',
                      ),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 12),
              Row(
                children: <Widget>[
                  Expanded(
                    child: _VbmappDialogDropdown(
                      label: '沟通形式',
                      value: _responseMode,
                      values: const <String>['口语', '手语', '图片交换', '手势', '其他'],
                      onChanged: (String value) {
                        setState(() => _responseMode = value);
                      },
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _VbmappDialogDropdown(
                      label: '辅助水平',
                      value: _promptLevel,
                      values: const <String>[
                        '无辅助',
                        '口头提示',
                        '仿说/模仿',
                        '其他辅助',
                        '肢体辅助',
                      ],
                      onChanged: (String value) {
                        setState(() => _promptLevel = value);
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                value: _functional,
                activeColor: _VbmappColors.orange,
                onChanged: (bool? value) {
                  setState(() => _functional = value ?? true);
                },
                title: const Text(
                  '本次反应是功能性要求，并获得或指向目标物',
                  style: TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: const Text('取消'),
                  ),
                  const SizedBox(width: 10),
                  ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _VbmappColors.orange,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 18,
                        vertical: 12,
                      ),
                    ),
                    onPressed: _submit,
                    child: const Text('保存事件'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _submit() {
    final String utterance = _utteranceController.text.trim();
    final String target = _targetController.text.trim();
    if (utterance.isEmpty && target.isEmpty) {
      return;
    }
    Navigator.of(context).pop(
      _VbmappMandEvent(
        utterance: utterance,
        target: target,
        motivationContext: _contextController.text.trim(),
        person: _personController.text.trim(),
        setting: _settingController.text.trim(),
        example: _exampleController.text.trim(),
        responseMode: _responseMode,
        promptLevel: _promptLevel,
        functional: _functional,
      ),
    );
  }
}

class _VbmappDialogTextField extends StatelessWidget {
  const _VbmappDialogTextField({
    required this.controller,
    required this.label,
    required this.hintText,
  });

  final TextEditingController controller;
  final String label;
  final String hintText;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        filled: true,
        fillColor: const Color(0xFFFFFAF5),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
      ),
    );
  }
}

class _VbmappDialogDropdown extends StatelessWidget {
  const _VbmappDialogDropdown({
    required this.label,
    required this.value,
    required this.values,
    required this.onChanged,
  });

  final String label;
  final String value;
  final List<String> values;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      value: value,
      decoration: InputDecoration(
        labelText: label,
        filled: true,
        fillColor: const Color(0xFFFFFAF5),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
      ),
      items: <DropdownMenuItem<String>>[
        for (final String option in values)
          DropdownMenuItem<String>(
            value: option,
            child: Text(option),
          ),
      ],
      onChanged: (String? value) {
        if (value != null) {
          onChanged(value);
        }
      },
    );
  }
}

class _VbmappScoreDock extends StatelessWidget {
  const _VbmappScoreDock({
    required this.item,
    required this.score,
    required this.onSelectScore,
  });

  final _VbmappItem item;
  final num? score;
  final ValueChanged<num> onSelectScore;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        _VbmappScoreOptionGrid(
          options: item.scoreOptions,
          selectedScore: score,
          accent: item.color,
          onSelectScore: onSelectScore,
        ),
      ],
    );
  }
}

class _VbmappScoreOptionGrid extends StatelessWidget {
  const _VbmappScoreOptionGrid({
    required this.options,
    required this.selectedScore,
    required this.accent,
    required this.onSelectScore,
  });

  final List<_VbmappScoreOption> options;
  final num? selectedScore;
  final Color accent;
  final ValueChanged<num> onSelectScore;

  int _columnCount(double width) {
    if (options.length <= 3) {
      return options.length.clamp(1, 3);
    }
    return width < 560 ? 2 : 3;
  }

  List<List<_VbmappScoreOption>> _optionRows(int columns) {
    final List<List<_VbmappScoreOption>> rows = <List<_VbmappScoreOption>>[];
    for (int start = 0; start < options.length; start += columns) {
      final int end = (start + columns).clamp(0, options.length);
      rows.add(options.sublist(start, end));
    }
    return rows;
  }

  @override
  Widget build(BuildContext context) {
    const double spacing = 10;
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final int columns = _columnCount(constraints.maxWidth);
        final List<List<_VbmappScoreOption>> rows = _optionRows(columns);
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int rowIndex = 0;
                rowIndex < rows.length;
                rowIndex++) ...<Widget>[
              if (rowIndex > 0) const SizedBox(height: spacing),
              Row(
                children: <Widget>[
                  for (int index = 0;
                      index < rows[rowIndex].length;
                      index++) ...<Widget>[
                    if (index > 0) const SizedBox(width: spacing),
                    Expanded(
                      child: _VbmappScoreOptionButton(
                        option: rows[rowIndex][index],
                        selected: selectedScore == rows[rowIndex][index].score,
                        accent: accent,
                        onTap: () => onSelectScore(rows[rowIndex][index].score),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ],
        );
      },
    );
  }
}

class _VbmappScoreOptionButton extends StatelessWidget {
  const _VbmappScoreOptionButton({
    required this.option,
    required this.selected,
    required this.accent,
    required this.onTap,
  });

  final _VbmappScoreOption option;
  final bool selected;
  final Color accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          width: double.infinity,
          height: 58,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? accent : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: selected ? accent : _VbmappColors.line),
            boxShadow: selected
                ? _vbmappShadow(color: accent.withOpacity(.16), blur: 14)
                : null,
          ),
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 42,
                child: Text(
                  option.displayScore,
                  style: TextStyle(
                    color: selected ? Colors.white : _VbmappColors.ink,
                    fontSize: 20,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Container(
                width: 1,
                height: 28,
                color: selected
                    ? Colors.white.withOpacity(.28)
                    : _VbmappColors.lineSoft,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  option.label,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: selected ? Colors.white : _VbmappColors.body,
                    fontSize: 11.5,
                    height: 1.18,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappRightRail extends StatelessWidget {
  const _VbmappRightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.selectedModule,
    required this.scoreSnapshot,
  });

  final double progressPercent;
  final int answered;
  final int total;
  final _VbmappModule selectedModule;
  final _VbmappScoreSnapshot scoreSnapshot;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(15),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Text(
            '测评进度',
            style: TextStyle(
              color: _VbmappColors.ink,
              fontSize: 18,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 14),
          ClipRRect(
            borderRadius: BorderRadius.circular(999),
            child: LinearProgressIndicator(
              value: progressPercent.clamp(0, 1).toDouble(),
              minHeight: 10,
              color: selectedModule.color,
              backgroundColor: _VbmappColors.lineSoft,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            '$answered / $total 项',
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 18),
          _VbmappSummaryStrip(
            label: selectedModule.title,
            value: '${selectedModule.itemCount}',
            subValue: selectedModule.subtitle,
            icon: selectedModule.icon,
            color: selectedModule.color,
          ),
          const SizedBox(height: 12),
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _VbmappCurrentScoreCard(snapshot: scoreSnapshot),
                  const SizedBox(height: 12),
                  _VbmappMilestoneDomainScoreCard(
                    domains: scoreSnapshot.milestoneDomains,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          const _VbmappLegend(),
        ],
      ),
    );
  }
}

class _VbmappSummaryStrip extends StatelessWidget {
  const _VbmappSummaryStrip({
    required this.label,
    required this.value,
    required this.subValue,
    required this.icon,
    this.color = _VbmappColors.orange,
  });

  final String label;
  final String value;
  final String subValue;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(.24)),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: color, size: 24),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  label,
                  maxLines: 1,
                  softWrap: false,
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subValue,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
          Text(
            value,
            style: const TextStyle(
              color: _VbmappColors.ink,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappCurrentScoreCard extends StatelessWidget {
  const _VbmappCurrentScoreCard({required this.snapshot});

  final _VbmappScoreSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Row(
            children: <Widget>[
              Icon(
                Icons.insights_rounded,
                color: _VbmappColors.orange,
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                '当前得分',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 15,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          _VbmappTinyMetric(
            label: '里程碑',
            value: '${snapshot.milestoneScoreText} / ${snapshot.milestoneMax}',
            color: _VbmappColors.orange,
          ),
          const SizedBox(height: 8),
          _VbmappTinyMetric(
            label: '障碍',
            value: '${snapshot.barrierTotal} / ${snapshot.barrierMax}',
            color: _VbmappColors.blue,
          ),
          const SizedBox(height: 8),
          _VbmappTinyMetric(
            label: '转衔',
            value: '${snapshot.transitionTotal} / ${snapshot.transitionMax}',
            color: _VbmappColors.green,
          ),
        ],
      ),
    );
  }
}

class _VbmappMilestoneDomainScoreCard extends StatelessWidget {
  const _VbmappMilestoneDomainScoreCard({required this.domains});

  final List<_VbmappDomainScoreSummary> domains;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Text(
            '里程碑领域',
            style: TextStyle(
              color: _VbmappColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          for (final _VbmappDomainScoreSummary domain in domains) ...<Widget>[
            _VbmappDomainScoreRow(domain: domain),
            if (domain != domains.last) const SizedBox(height: 10),
          ],
        ],
      ),
    );
  }
}

class _VbmappDomainScoreRow extends StatelessWidget {
  const _VbmappDomainScoreRow({required this.domain});

  final _VbmappDomainScoreSummary domain;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                domain.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            Text(
              '${domain.scoreText}/${domain.maxScore}',
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(width: 6),
            Text(
              '${domain.answered}/${domain.total}项',
              style: const TextStyle(
                color: _VbmappColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: domain.percent,
            minHeight: 7,
            color: _VbmappColors.orange,
            backgroundColor: _VbmappColors.lineSoft,
          ),
        ),
      ],
    );
  }
}

class _VbmappTinyMetric extends StatelessWidget {
  const _VbmappTinyMetric({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _VbmappLegend extends StatelessWidget {
  const _VbmappLegend();

  @override
  Widget build(BuildContext context) {
    return const Row(
      children: <Widget>[
        _VbmappLegendItem(color: _VbmappColors.orange, text: '0 / .5 / 1'),
        SizedBox(width: 8),
        _VbmappLegendItem(color: _VbmappColors.blue, text: '0-4'),
        SizedBox(width: 8),
        _VbmappLegendItem(color: _VbmappColors.green, text: '1-5'),
      ],
    );
  }
}

class _VbmappLegendItem extends StatelessWidget {
  const _VbmappLegendItem({required this.color, required this.text});

  final Color color;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Row(
        children: <Widget>[
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 4),
          Expanded(
            child: Text(
              text,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _VbmappColors.body,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappFooterDock extends StatelessWidget {
  const _VbmappFooterDock({
    required this.current,
    required this.total,
    required this.hasPrevious,
    required this.hasNext,
    required this.hasMissing,
    required this.autoNext,
    required this.onPrevious,
    required this.onNext,
    required this.onJumpMissing,
    required this.onToggleAutoNext,
  });

  final int current;
  final int total;
  final bool hasPrevious;
  final bool hasNext;
  final bool hasMissing;
  final bool autoNext;
  final VoidCallback onPrevious;
  final VoidCallback onNext;
  final VoidCallback onJumpMissing;
  final ValueChanged<bool> onToggleAutoNext;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _VbmappColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _VbmappFooterButton(
            label: '上一题',
            icon: Icons.chevron_left_rounded,
            enabled: hasPrevious,
            onTap: onPrevious,
          ),
          const Spacer(),
          Text.rich(
            TextSpan(
              children: <InlineSpan>[
                TextSpan(
                  text: '$current',
                  style: const TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 26,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _VbmappFooterButton(
            label: '下一题',
            icon: Icons.arrow_forward_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _VbmappFooterButton(
            label: '跳到缺题',
            icon: Icons.format_list_bulleted_rounded,
            enabled: hasMissing,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _VbmappColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _VbmappColors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
      ),
    );
  }
}

class _VbmappFooterButton extends StatelessWidget {
  const _VbmappFooterButton({
    required this.label,
    required this.icon,
    required this.enabled,
    required this.onTap,
    this.filled = false,
    this.reverseIcon = false,
  });

  final String label;
  final IconData icon;
  final bool enabled;
  final VoidCallback onTap;
  final bool filled;
  final bool reverseIcon;

  @override
  Widget build(BuildContext context) {
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _VbmappColors.orangeDeep
        : _VbmappColors.muted;
    final List<Widget> children = <Widget>[
      Icon(icon, size: 21, color: textColor),
      const SizedBox(width: 8),
      Text(
        label,
        style: TextStyle(
          color: textColor,
          fontSize: 15,
          fontWeight: FontWeight.w900,
        ),
      ),
    ];
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          width: filled ? 140 : 128,
          height: 38,
          decoration: BoxDecoration(
            color: filled && enabled
                ? _VbmappColors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: reverseIcon ? children.reversed.toList() : children,
          ),
        ),
      ),
    );
  }
}

class _VbmappActiveObservationBar extends StatefulWidget {
  const _VbmappActiveObservationBar({
    required this.tone,
    required this.observation,
    required this.statusLabel,
    required this.summaryText,
    required this.onJump,
    required this.onQuickRecord,
    required this.onPrimaryAction,
    required this.onFinish,
  });

  final Color tone;
  final _VbmappObservationTimerState observation;
  final String statusLabel;
  final String summaryText;
  final VoidCallback onJump;
  final VoidCallback onQuickRecord;
  final VoidCallback onPrimaryAction;
  final VoidCallback onFinish;

  @override
  State<_VbmappActiveObservationBar> createState() =>
      _VbmappActiveObservationBarState();
}

class _VbmappActiveObservationBarState
    extends State<_VbmappActiveObservationBar> {
  Timer? _ticker;

  @override
  void initState() {
    super.initState();
    _syncTicker();
  }

  @override
  void didUpdateWidget(covariant _VbmappActiveObservationBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.observation != widget.observation) {
      _syncTicker();
    }
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  void _syncTicker() {
    _ticker?.cancel();
    _ticker = null;
    if (!widget.observation.isRunning) {
      return;
    }
    _ticker = Timer.periodic(const Duration(seconds: 1), (Timer timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {});
    });
  }

  @override
  Widget build(BuildContext context) {
    final bool running = widget.observation.isRunning;
    final String statusLabel = running
        ? widget.statusLabel
        : widget.statusLabel.replaceFirst('观察中', '观察暂停');
    return Container(
      key: const ValueKey<String>('vbmapp-active-observation-bar'),
      height: 40,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 12),
      ),
      child: Row(
        children: <Widget>[
          InkWell(
            onTap: widget.onJump,
            borderRadius: BorderRadius.circular(999),
            child: Ink(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: widget.tone.withOpacity(.1),
                borderRadius: BorderRadius.circular(999),
              ),
              child: Text(
                statusLabel,
                style: TextStyle(
                  color: widget.tone,
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
          ),
          const SizedBox(width: 7),
          Text(
            widget.summaryText,
            style: const TextStyle(
              color: _VbmappColors.ink,
              fontSize: 11.5,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _VbmappMand4TimerButton(
            key: const ValueKey<String>(
                'vbmapp-active-observation-quick-record'),
            icon: Icons.add_rounded,
            label: '记一条',
            filled: true,
            compact: true,
            onTap: widget.onQuickRecord,
          ),
          const SizedBox(width: 5),
          _VbmappObservationMiniIconButton(
            key: const ValueKey<String>('vbmapp-active-observation-primary'),
            icon: running ? Icons.pause_rounded : Icons.play_arrow_rounded,
            tooltip: running ? '暂停观察' : '继续观察',
            onTap: widget.onPrimaryAction,
          ),
          const SizedBox(width: 5),
          _VbmappObservationMiniIconButton(
            key: const ValueKey<String>('vbmapp-active-observation-finish'),
            icon: Icons.stop_rounded,
            tooltip: '结束观察',
            onTap: widget.onFinish,
          ),
        ],
      ),
    );
  }
}

class _VbmappObservationMiniIconButton extends StatelessWidget {
  const _VbmappObservationMiniIconButton({
    required this.icon,
    required this.tooltip,
    required this.onTap,
    super.key,
  });

  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(8),
          child: Ink(
            width: 28,
            height: 28,
            decoration: BoxDecoration(
              color: const Color(0xFFFFFCFA),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: _VbmappColors.lineSoft),
            ),
            child: Icon(icon, size: 16, color: _VbmappColors.orangeDeep),
          ),
        ),
      ),
    );
  }
}

class _VbmappLoadingState extends StatelessWidget {
  const _VbmappLoadingState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Text(
        '正在载入VB-MAPP测评',
        style: TextStyle(
          color: _VbmappColors.body,
          fontSize: 18,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

BoxDecoration _vbmappCardDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.94),
    borderRadius: BorderRadius.circular(16),
    border: Border.all(color: _VbmappColors.line),
    boxShadow: _vbmappShadow(),
  );
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
      item.itemCode == 'MAND_04M' ||
      item.itemCode == 'MAND_05M';
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
  final List<String> values = <String>[
    if (_mandInitiationText(event).isNotEmpty) _mandInitiationText(event),
    if (event.environment.trim().isNotEmpty) event.environment.trim(),
    if (event.targetKind.trim().isNotEmpty) event.targetKind.trim(),
    if (event.phraseLevel.trim().isNotEmpty) event.phraseLevel.trim(),
    if (event.promptLevel.trim().isNotEmpty) event.promptLevel.trim(),
  ];
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
