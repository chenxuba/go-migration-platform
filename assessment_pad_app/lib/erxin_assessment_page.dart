import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_draft_resume_dialog.dart';
import 'assessment_age_formatter.dart';
import 'assessment_scale_client.dart';
import 'erxin_assessment_client.dart';
import 'home_client.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';

part 'erxin_assessment_overview.dart';
part 'erxin_assessment_chrome.dart';
part 'erxin_assessment_loading.dart';
part 'erxin_assessment_navigation.dart';
part 'erxin_assessment_items.dart';
part 'erxin_assessment_remark.dart';
part 'erxin_assessment_rules.dart';
part 'erxin_assessment_misc.dart';
part 'erxin_assessment_state_bootstrap.dart';
part 'erxin_assessment_state_actions.dart';
part 'erxin_assessment_support.dart';

const double _erxinDetailPanelHeight = 156;
const double _erxinDetailPanelTopPadding = 8;
const double _erxinDetailPanelBottomPadding = 12;
const double _erxinDetailHeaderHeight = 28;
const double _erxinDetailHeaderGap = 6;
const double _erxinDetailContentHeight = _erxinDetailPanelHeight -
    _erxinDetailPanelTopPadding -
    _erxinDetailPanelBottomPadding -
    _erxinDetailHeaderHeight -
    _erxinDetailHeaderGap;
const double _erxinRulePanelBottomPadding = 12;
const double _erxinRightRemarkSectionHeight =
    _erxinDetailPanelHeight - _erxinRulePanelBottomPadding;
const double _erxinSidebarBottomPadding = 6;
const double _erxinProgressSummaryHeight =
    _erxinDetailPanelHeight - _erxinDetailPanelBottomPadding;
const double _erxinProgressSummaryBottomGap =
    _erxinDetailPanelBottomPadding - _erxinSidebarBottomPadding;

class ErxinAssessmentPage extends StatefulWidget {
  const ErxinAssessmentPage({
    required this.onBack,
    this.args = const ErxinAssessmentLaunchArgs(),
    this.client = const ApiErxinAssessmentClient(),
    this.homeClient,
    super.key,
  });

  final VoidCallback onBack;
  final ErxinAssessmentLaunchArgs args;
  final ErxinAssessmentClient client;
  final HomeClient? homeClient;

  @override
  State<ErxinAssessmentPage> createState() => _ErxinAssessmentPageState();
}

class _ErxinAssessmentPageState extends State<ErxinAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';
  static const List<int> _standardAgeMonths = <int>[
    1,
    2,
    3,
    4,
    5,
    6,
    7,
    8,
    9,
    10,
    11,
    12,
    15,
    18,
    21,
    24,
    27,
    30,
    33,
    36,
    42,
    48,
    54,
    60,
    66,
    72,
    78,
    84,
  ];

  final Map<int, bool> _itemPasses = <int, bool>{};
  final Map<int, String> _itemRemarks = <int, String>{};
  final Map<int, ErxinAssessmentItem> _itemDetailCache =
      <int, ErxinAssessmentItem>{};
  final Map<int, Future<ErxinAssessmentItem>> _itemDetailFetches =
      <int, Future<ErxinAssessmentItem>>{};
  final Map<int, GlobalKey> _itemRowKeys = <int, GlobalKey>{};
  final Map<String, GlobalKey> _monthSectionKeys = <String, GlobalKey>{};
  final ScrollController _workspaceScrollController = ScrollController();
  final Map<String, int> _previousStartIndexByDomain = <String, int>{};
  final Map<String, int> _futureEndIndexByDomain = <String, int>{};
  final Set<String> _futureVisibleDomains = <String>{};
  final Map<String, int> _reviewMonthByDomain = <String, int>{};
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  Future<ErxinDraftDetail?>? _saveDraftFuture;
  List<int> _recordRevealMonths = const <int>[];
  List<int> _workspaceFlashMonths = const <int>[];
  int _recordRevealSerial = 0;
  int _workspaceFlashSerial = 0;
  Timer? _workspaceFlashTimer;
  bool _saveDraftFutureSilent = false;
  bool _saveDraftJoinedByManual = false;

  ErxinTemplateSummary _template = ErxinTemplateSummary.empty;
  ErxinDraftProgress _draftProgress = ErxinDraftProgress.empty;
  AssessmentDraftSummary? _detectedDraft;
  Future<ErxinDraftDetail>? _detectedDraftDetailRequest;
  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _selectedDomainCode = '';
  int _selectedItemNo = 0;
  int _studentId = 0;
  int _draftId = 0;
  int _detectedDraftDetailDraftId = 0;
  bool _draftDialogShown = false;
  bool _loading = true;
  bool _savingDraft = false;
  bool _submitting = false;
  String _errorMessage = '';
  String _autoSaveText = '等待作答';

  @override
  void initState() {
    super.initState();
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate);
    _examinerName = widget.args.examinerName;
    _initialize();
  }

  @override
  void dispose() {
    _workspaceFlashTimer?.cancel();
    _workspaceScrollController.dispose();
    _messageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return _buildLoadingShell();
    }
    if (_errorMessage.trim().isNotEmpty) {
      return _ErrorView(message: _errorMessage, onBack: widget.onBack);
    }
    return ColoredBox(
      color: _ErxinColors.page,
      child: Column(
        children: <Widget>[
          _Header(
            args: _headerArgs(),
            autoSaveText: _autoSaveText,
            saving: _savingDraft,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: _saveDraft,
            onSubmit: _submitDraft,
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _DomainSidebar(
                    domains: _template.domains,
                    selectedCode: _selectedDomainCode,
                    progressForDomain: _domainProgress,
                    domainCompleteForDomain: _domainStopRuleComplete,
                    completedDomainCount: _completedDomainCount(),
                    savedItemCount: _draftProgress.answeredItemCount,
                    onSelect: _selectDomain,
                    onShowAllItems: _showAllItemsOverview,
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Container(
                      clipBehavior: Clip.antiAlias,
                      decoration: _erxinPanelDecoration(),
                      child: Column(
                        children: <Widget>[
                          Expanded(child: _buildWorkspace()),
                          _buildDetailPanel(),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  _buildRulePanel(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
