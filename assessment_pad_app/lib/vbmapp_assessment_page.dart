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
const Set<String> _vbmappSharedTimedMandItemCodes = <String>{
  'MAND_04M',
  'MAND_08M',
  'MAND_09M',
};

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
            _vbmappSharedTimedMandStorageKey;
    final bool showActiveObservationBar = activeObservationItem != null &&
        activeObservation != null &&
        !(activeTimedMandShared &&
            _vbmappSharedTimedMandItemCodes.contains(item.itemCode)) &&
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
                          total: _vbmappTotalItemCount,
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
                total: _vbmappTotalItemCount,
                hasPrevious: item.sequenceNo > 1,
                hasNext: item.sequenceNo < _vbmappTotalItemCount,
                hasMissing: _answeredCount < _vbmappTotalItemCount,
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
}
