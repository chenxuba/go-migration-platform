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
part 'vbmapp_assessment_mand_12_panel.dart';
part 'vbmapp_assessment_timed_mand_panel.dart';
part 'vbmapp_assessment_mand_3_panel.dart';
part 'vbmapp_assessment_mand_5_panel.dart';
part 'vbmapp_assessment_mand_6_panel.dart';
part 'vbmapp_assessment_mand_late_panel.dart';
part 'vbmapp_assessment_mand_widgets.dart';
part 'vbmapp_assessment_mand_dialogs.dart';
part 'vbmapp_assessment_support.dart';
part 'vbmapp_assessment_selectors.dart';
part 'vbmapp_assessment_draft_actions.dart';
part 'vbmapp_assessment_mand_actions.dart';
part 'vbmapp_assessment_navigation_actions.dart';

const String _vbmappAuthTokenStorageKey = 'auth_token';
const String _vbmappScaleVersion = 'VBMAPP_CN_2ND_DRAFT_2026_05';
const int _vbmappTotalItemCount = 212;
const String _vbmappSharedTimedMandStorageKey = '__MAND_TIMED_SHARED__';
const String _vbmappMandLevel2TimedGroupId = 'mand_timed_shared_v1';
const String _vbmappMandLevel3TimedGroupId = 'mand_level3_timed_shared_v1';
const Set<String> _vbmappMandLevel2TimedItemCodes = <String>{
  'MAND_04M',
  'MAND_08M',
  'MAND_09M',
};
const Set<String> _vbmappMandLevel3TimedItemCodes = <String>{
  'MAND_11M',
  'MAND_13M',
};

List<double> _activeObservationBarWidths(double available, int count) {
  if (count <= 0) {
    return const <double>[];
  }
  if (count == 1) {
    return <double>[available];
  }
  const double gap = 8;
  final double usable = available - gap * (count - 1);
  if (count == 2 && usable >= 1120) {
    return <double>[usable * .53, usable * .47];
  }
  final double evenWidth = (usable / count).clamp(560, 960);
  return List<double>.filled(count, evenWidth);
}

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
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  final ValueNotifier<int> _selectionRevision = ValueNotifier<int>(0);
  final ValueNotifier<int> _moduleRevision = ValueNotifier<int>(0);
  final ValueNotifier<String> _selectedItemCode = ValueNotifier<String>('');
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
  int _cachedAnsweredCount = 0;
  Map<String, int> _cachedAnsweredCountByModule = const <String, int>{
    'milestones': 0,
    'barriers': 0,
    'transition': 0,
  };
  double _cachedProgressPercent = 0;
  _VbmappScoreSnapshot _cachedScoreSnapshot = const _VbmappScoreSnapshot(
    milestoneTotal: 0,
    milestoneMax: 170,
    barrierTotal: 0,
    barrierMax: 96,
    transitionTotal: 0,
    transitionMax: 90,
    milestoneDomains: <_VbmappDomainScoreSummary>[],
  );

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
    _selectedItemCode.value = _selectedItem.itemCode;
    unawaited(_initialize());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _selectionRevision.dispose();
    _moduleRevision.dispose();
    _selectedItemCode.dispose();
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

  @override
  Widget build(BuildContext context) {
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
            Expanded(
              child: _loading
                  ? const _VbmappLoadingState()
                  : Column(
                      children: <Widget>[
                        ValueListenableBuilder<int>(
                          valueListenable: _selectionRevision,
                          builder:
                              (BuildContext context, int _, Widget? child) {
                            final _VbmappItem item = _selectedItem;
                            final List<Widget> activeBars = <Widget>[];
                            final String? currentSharedGroupId =
                                _sharedTimedMandGroupIdFor(item);
                            for (final _VbmappItem activeObservationItem
                                in _activeMandObservationItems()) {
                              final _VbmappObservationTimerState?
                                  activeObservation = _mandObservationFor(
                                activeObservationItem,
                              );
                              final String? activeSharedGroupId =
                                  _sharedTimedMandGroupIdFor(
                                activeObservationItem,
                              );
                              final bool activeTimedMandShared =
                                  activeSharedGroupId != null;
                              final bool showActiveObservationBar =
                                  activeObservation != null &&
                                      !(activeTimedMandShared &&
                                          activeSharedGroupId ==
                                              currentSharedGroupId) &&
                                      activeObservationItem.itemCode !=
                                          item.itemCode;
                              if (!showActiveObservationBar) {
                                continue;
                              }
                              final _VbmappObservationTimerState
                                  activeBarObservation = activeTimedMandShared
                                      ? activeObservation.withPlannedMinutes(
                                          _sharedTimedMandMaxPlannedMinutesForGroup(
                                            activeSharedGroupId,
                                          ),
                                        )
                                      : activeObservation;
                              activeBars.add(
                                Padding(
                                  padding:
                                      const EdgeInsets.fromLTRB(12, 8, 12, 0),
                                  child: _VbmappActiveObservationBar(
                                    tone: activeObservationItem.color,
                                    observation: activeBarObservation,
                                    statusLabel: activeTimedMandShared
                                        ? _sharedTimedMandStatusLabelForGroup(
                                            activeSharedGroupId,
                                          )
                                        : '${activeObservationItem.navCode}观察中',
                                    sharedSummaryMode: activeTimedMandShared,
                                    summaries: activeTimedMandShared
                                        ? _sharedTimedMandSummariesForGroup(
                                            activeSharedGroupId,
                                          )
                                        : const <_VbmappActiveObservationSummary>[],
                                    recordCount: activeTimedMandShared
                                        ? _sharedTimedMandRecordCountForGroup(
                                            activeSharedGroupId,
                                          )
                                        : _mandStoredEventsFor(
                                                activeObservationItem)
                                            .length,
                                    qualifiedCount:
                                        _activeMandObservationQualifiedCount(
                                      activeObservationItem,
                                    ),
                                    onePointTarget: _scoreCountThreshold(
                                            activeObservationItem, 1) ??
                                        5,
                                    onJump: () {
                                      final String targetCode =
                                          _sharedTimedMandPrimaryItemCodeFor(
                                                activeObservationItem,
                                              ) ??
                                              activeObservationItem.itemCode;
                                      _selectItem(
                                        _milestoneItems.firstWhere(
                                          (_VbmappItem candidate) =>
                                              candidate.itemCode == targetCode,
                                          orElse: () => activeObservationItem,
                                        ),
                                      );
                                    },
                                    onQuickRecord: () => unawaited(
                                      _openActiveMandQuickRecord(
                                        activeObservationItem,
                                      ),
                                    ),
                                    onPrimaryAction: () {
                                      final DateTime now = DateTime.now();
                                      if (activeBarObservation.isRunning) {
                                        unawaited(
                                          _updateMandObservation(
                                            activeObservationItem,
                                            activeBarObservation.pause(now),
                                          ),
                                        );
                                        return;
                                      }
                                      unawaited(
                                        _updateMandObservation(
                                          activeObservationItem,
                                          activeBarObservation.resume(now),
                                        ),
                                      );
                                    },
                                    onFinish: () => unawaited(
                                      _confirmFinishActiveMandObservation(
                                        activeObservationItem,
                                      ),
                                    ),
                                    onAutoFinish: () {
                                      unawaited(
                                        _updateMandObservation(
                                          activeObservationItem,
                                          activeBarObservation
                                              .finishAtPlannedEnd(
                                            DateTime.now(),
                                          ),
                                        ),
                                      );
                                    },
                                  ),
                                ),
                              );
                            }
                            if (activeBars.isEmpty) {
                              return const SizedBox.shrink();
                            }
                            return LayoutBuilder(
                              builder: (
                                BuildContext context,
                                BoxConstraints constraints,
                              ) {
                                final double available =
                                    constraints.maxWidth.clamp(
                                  0,
                                  double.infinity,
                                );
                                final List<double> barWidths =
                                    _activeObservationBarWidths(
                                  available,
                                  activeBars.length,
                                );
                                return SingleChildScrollView(
                                  scrollDirection: Axis.horizontal,
                                  physics: const ClampingScrollPhysics(),
                                  child: Row(
                                    children: <Widget>[
                                      for (int index = 0;
                                          index < activeBars.length;
                                          index++) ...<Widget>[
                                        if (index > 0) const SizedBox(width: 8),
                                        SizedBox(
                                          width: barWidths[index],
                                          child: activeBars[index],
                                        ),
                                      ],
                                    ],
                                  ),
                                );
                              },
                            );
                          },
                        ),
                        Expanded(
                          child: Padding(
                            padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
                            child: Row(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: <Widget>[
                                SizedBox(
                                  width: 250,
                                  child: ValueListenableBuilder<int>(
                                    valueListenable: _moduleRevision,
                                    builder: (
                                      BuildContext context,
                                      int _,
                                      Widget? child,
                                    ) {
                                      final List<_VbmappItem> selectedItems =
                                          _selectedItems;
                                      return _VbmappModuleRail(
                                        modules: _vbmappModules,
                                        selectedCode: _selectedModuleCode,
                                        selectedItemCodeListenable:
                                            _selectedItemCode,
                                        items: selectedItems,
                                        answeredCount: _answeredCountByModule,
                                        isAnswered: (_VbmappItem item) =>
                                            _scoreFor(item) != null,
                                        onSelectModule: _selectModule,
                                        onSelectItem: _selectItem,
                                      );
                                    },
                                  ),
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: ValueListenableBuilder<int>(
                                    valueListenable: _selectionRevision,
                                    builder: (
                                      BuildContext context,
                                      int _,
                                      Widget? child,
                                    ) {
                                      final _VbmappItem item = _selectedItem;
                                      final VbmappItemResponseSchema?
                                          itemSchema = _schemaFor(item);
                                      final VbmappMaterialProfile?
                                          materialProfile =
                                          _materialProfileFor(item, itemSchema);
                                      return _VbmappWorkspace(
                                        item: item,
                                        score: _scoreFor(item),
                                        responseSchema: itemSchema,
                                        materialProfile: materialProfile,
                                        mandEvents: _mandEventsFor(item),
                                        mandObservation: _mandObservationFor(
                                          item,
                                        ),
                                        onAddMandEvent: () => unawaited(
                                          _openMandEventDialog(item),
                                        ),
                                        onSubmitMandEvent:
                                            (_VbmappMandEvent event) =>
                                                unawaited(
                                          _addMandEvent(item, event),
                                        ),
                                        onDeleteMandEvent: (int index) =>
                                            unawaited(
                                          _deleteMandEvent(item, index),
                                        ),
                                        onChangeMandObservation:
                                            (_VbmappObservationTimerState
                                                    observation) =>
                                                unawaited(
                                          _updateMandObservation(
                                            item,
                                            observation,
                                          ),
                                        ),
                                        onSelectScore: _selectScore,
                                      );
                                    },
                                  ),
                                ),
                                const SizedBox(width: 10),
                                SizedBox(
                                  width: 278,
                                  child: ValueListenableBuilder<int>(
                                    valueListenable: _moduleRevision,
                                    child: _VbmappRightRailScoreDetails(
                                      snapshot: _scoreSnapshot,
                                    ),
                                    builder: (
                                      BuildContext context,
                                      int _,
                                      Widget? scoreDetails,
                                    ) {
                                      return _VbmappRightRail(
                                        progressPercent: _progressPercent,
                                        answered: _answeredCount,
                                        total: _vbmappTotalItemCount,
                                        selectedModule:
                                            _moduleByCode(_selectedModuleCode),
                                        scoreDetails: scoreDetails ??
                                            _VbmappRightRailScoreDetails(
                                              snapshot: _scoreSnapshot,
                                            ),
                                      );
                                    },
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                        ValueListenableBuilder<int>(
                          valueListenable: _selectionRevision,
                          builder:
                              (BuildContext context, int _, Widget? child) {
                            final _VbmappItem item = _selectedItem;
                            return _VbmappFooterDock(
                              current: item.sequenceNo,
                              total: _vbmappTotalItemCount,
                              hasPrevious: item.sequenceNo > 1,
                              hasNext: item.sequenceNo < _vbmappTotalItemCount,
                              hasMissing:
                                  _answeredCount < _vbmappTotalItemCount,
                              autoNext: _autoNext,
                              onPrevious: _goPrevious,
                              onNext: _goNext,
                              onJumpMissing: _jumpFirstMissing,
                              onToggleAutoNext: (bool value) =>
                                  setState(() => _autoNext = value),
                            );
                          },
                        ),
                      ],
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
