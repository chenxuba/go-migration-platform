import 'dart:async';
import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_draft_resume_dialog.dart';
import 'assessment_age_formatter.dart';
import 'home_client.dart';
import 'pad_top_message.dart';
import 'pad_responsive.dart';
import 'pep3_assessment_client.dart';

part 'pep3_assessment_chrome.dart';
part 'pep3_assessment_navigation.dart';
part 'pep3_assessment_workspace.dart';
part 'pep3_assessment_right_rail.dart';
part 'pep3_assessment_footer.dart';
part 'pep3_assessment_widgets.dart';
part 'pep3_assessment_loading.dart';
part 'pep3_assessment_state_actions.dart';
part 'pep3_assessment_support.dart';

class Pep3AssessmentPage extends StatefulWidget {
  const Pep3AssessmentPage({
    required this.onBack,
    this.args = const Pep3AssessmentLaunchArgs(),
    this.client = const ApiPep3AssessmentClient(),
    this.homeClient = const ApiHomeClient(),
    super.key,
  });

  final VoidCallback onBack;
  final Pep3AssessmentLaunchArgs args;
  final Pep3AssessmentClient client;
  final HomeClient homeClient;

  @override
  State<Pep3AssessmentPage> createState() => _Pep3AssessmentPageState();
}

class _Pep3AssessmentPageState extends State<Pep3AssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';

  Pep3TemplateSummary _template = Pep3TemplateSummary.empty;
  final Map<int, Pep3AssessmentItem> _itemCache = <int, Pep3AssessmentItem>{};
  final Map<int, Future<Pep3AssessmentItem>> _itemFetches =
      <int, Future<Pep3AssessmentItem>>{};
  final Map<int, int> _itemScores = <int, int>{};
  final Map<int, Map<String, dynamic>> _recordValues =
      <int, Map<String, dynamic>>{};
  final Set<int> _savingItems = <int>{};
  final Set<int> _savedItems = <int>{};
  final ScrollController _leftScrollController = ScrollController();
  final ScrollController _questionScrollController = ScrollController();
  final GlobalKey _activeNavItemKey = GlobalKey();
  final Map<String, GlobalKey> _pageGroupKeys = <String, GlobalKey>{};
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  Pep3DraftDetail? _draft;
  Pep3DraftSummary? _detectedDraft;
  Pep3CaregiverInvite? _caregiverInvite;
  int _caregiverInviteDraftId = 0;
  Future<void>? _caregiverInviteRequest;
  Future<Pep3DraftDetail>? _detectedDraftDetailRequest;
  int _detectedDraftDetailDraftId = 0;
  HomeSession _session = HomeSession.fallback;
  final Map<int, int> _previousItemScores = <int, int>{};
  int _currentItemNo = 0;
  String _expandedGroupKey = '';
  String _errorMessage = '';
  String _autoSaveText = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _previousAssessmentDate = '';
  String _previousAssessmentLookupKey = '';
  bool _draftDialogShown = false;
  bool _loading = true;
  bool _itemLoading = false;
  bool _savingDraft = false;
  bool _submitting = false;
  bool _caregiverLoading = false;
  bool _autoNext = true;

  @override
  void initState() {
    super.initState();
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _normalizeDate(widget.args.birthDate) ?? '';
    _assessmentDate =
        _normalizeDate(widget.args.assessmentDate) ?? _todayIsoDate();
    _examinerName = widget.args.examinerName;
    _initialize();
  }

  @override
  void dispose() {
    _messageController.dispose();
    _leftScrollController.dispose();
    _questionScrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return _Pep3LoadingShell(
        title: _scaleTitle,
        studentName: _studentName,
        age: _studentAgeText,
        assessmentDate: _assessmentDate,
        examinerName: _examinerName,
        onBack: widget.onBack,
      );
    }
    if (_errorMessage.isNotEmpty) {
      return _Pep3ErrorShell(message: _errorMessage, onBack: widget.onBack);
    }
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Column(
        children: <Widget>[
          _Pep3Header(
            title: _scaleTitle,
            studentName: _studentName.trim().isEmpty ? '-' : _studentName,
            age: _studentAgeText,
            assessmentDate: _assessmentDate,
            examinerName: _examinerName.trim().isEmpty ? '当前老师' : _examinerName,
            autoSaveText: _autoSaveText,
            saving: _savingDraft,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: () => _saveDraft(),
            onSubmit: _submitDraft,
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  SizedBox(
                    width: 226,
                    child: _Pep3PageSidebar(
                      groups: _template.itemGroups,
                      expandedGroupKey: _expandedGroupKey,
                      itemScores: _itemScores,
                      currentItemNo: _currentItemNo,
                      controller: _leftScrollController,
                      activeItemKey: _activeNavItemKey,
                      groupKeys: _pageGroupKeys,
                      onCollapseAll: () {
                        if (_expandedGroupKey.isEmpty) {
                          return;
                        }
                        setState(() => _expandedGroupKey = '');
                      },
                      onToggleGroup: (String key) {
                        final bool opening = _expandedGroupKey != key;
                        setState(() {
                          _expandedGroupKey = opening ? key : '';
                        });
                        if (opening) {
                          _keepPageGroupVisible(key);
                        }
                      },
                      onTapItem: _goToItem,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _Pep3QuestionPanel(
                      controller: _questionScrollController,
                      loading: _itemLoading,
                      item: _currentItem,
                      summary: _currentSummary,
                      scoreOptions: _currentScoreOptions,
                      selectedScore: _itemScores[_currentItemNo],
                      previousScore: _previousItemScores[_currentItemNo],
                      previousAssessmentDate: _previousAssessmentDate,
                      saving: _savingItems.contains(_currentItemNo),
                      saved: _savedItems.contains(_currentItemNo),
                      recordValues:
                          _recordValues[_currentItemNo] ?? <String, dynamic>{},
                      onScore: _setScore,
                      onRecordValue: _setRecordValue,
                    ),
                  ),
                  const SizedBox(width: 10),
                  SizedBox(
                    width: 238,
                    child: _Pep3RightRail(
                      progressPercent: _progressPercent,
                      answered: _answeredCount,
                      total: _totalCount,
                      missing: _missingCount,
                      currentItemNo: _currentItemNo,
                      recordFields:
                          _currentItem?.recordFields ?? <Pep3RecordField>[],
                      recordValues:
                          _recordValues[_currentItemNo] ?? <String, dynamic>{},
                      caregiverInvite: _caregiverInvite,
                      caregiverLoading: _caregiverLoading,
                      onRecordValue: _setRecordValueForItem,
                      onSmsTap: () => _showMessage('短信发送功能暂未开放'),
                      onWechatTap: () => _showMessage('微信推送功能暂未开放'),
                    ),
                  ),
                ],
              ),
            ),
          ),
          _Pep3Footer(
            current: _currentIndex + 1,
            total: _totalCount,
            hasPrevious: _hasPreviousItem,
            hasNext: _hasNextItem,
            autoNext: _autoNext,
            onPrevious: _goPrevious,
            onNext: _goNext,
            onJumpMissing: _jumpToMissing,
            onToggleAutoNext: (bool value) => setState(() => _autoNext = value),
          ),
        ],
      ),
    );
  }
}
