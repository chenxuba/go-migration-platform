import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'erxin_assessment_client.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';

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
    super.key,
  });

  final VoidCallback onBack;
  final ErxinAssessmentLaunchArgs args;
  final ErxinAssessmentClient client;

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
  final Map<String, int> _previousStartIndexByDomain = <String, int>{};
  final Map<String, int> _futureEndIndexByDomain = <String, int>{};
  final Set<String> _futureVisibleDomains = <String>{};
  final Map<String, int> _reviewMonthByDomain = <String, int>{};
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  Future<ErxinDraftDetail?>? _saveDraftFuture;
  List<int> _recordRevealMonths = const <int>[];
  int _recordRevealSerial = 0;
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

  Future<void> _initialize() async {
    setState(() {
      _loading = true;
      _errorMessage = '';
    });
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '请先登录后再进行测评';
      });
      return;
    }
    try {
      final ErxinTemplateSummary template =
          await widget.client.fetchTemplateSummary(token);
      if (!mounted) {
        return;
      }
      final String firstDomain = template.domains.isNotEmpty
          ? template.domains.first.domainCode
          : 'GM';
      _token = token;
      _template = template;
      _selectedDomainCode = firstDomain;
      if (widget.args.draftId > 0) {
        final ErxinDraftDetail detail = await widget.client.fetchDraftDetail(
          token,
          widget.args.draftId,
        );
        _applyDraftDetail(detail);
      } else {
        _detectedDraft = await _findLatestDraft(token);
        _prefetchDetectedDraftDetail(token, _detectedDraft);
      }
      _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
      setState(() {
        _loading = false;
        _autoSaveText = '已准备';
      });
      _prefetchSelectedItem();
      _showDetectedDraftDialogIfNeeded();
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '儿心量表加载失败：$error';
      });
    }
  }

  Future<AssessmentDraftSummary?> _findLatestDraft(String token) async {
    if (_studentId <= 0) {
      return null;
    }
    final AssessmentDraftPage page = await widget.client.fetchDraftsPage(
      token,
      studentId: _studentId,
      pageSize: 1,
      latestOnly: true,
    );
    if (page.items.isEmpty || page.items.first.id <= 0) {
      return null;
    }
    return page.items.first;
  }

  void _prefetchDetectedDraftDetail(
    String token,
    AssessmentDraftSummary? draft,
  ) {
    if (draft == null || draft.id <= 0) {
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      return;
    }
    if (_detectedDraftDetailDraftId == draft.id &&
        _detectedDraftDetailRequest != null) {
      return;
    }
    final Future<ErxinDraftDetail> request =
        widget.client.fetchDraftDetail(token, draft.id);
    _detectedDraftDetailDraftId = draft.id;
    _detectedDraftDetailRequest = request;
    unawaited(
      request.then<void>(
        (ErxinDraftDetail _) {},
        onError: (Object _, StackTrace __) {},
      ),
    );
  }

  Future<ErxinDraftDetail> _resolveDetectedDraftDetail(
    AssessmentDraftSummary draft,
  ) async {
    final Future<ErxinDraftDetail>? prefetched = _detectedDraftDetailRequest;
    if (_detectedDraftDetailDraftId == draft.id && prefetched != null) {
      try {
        return await prefetched;
      } on Object {
        // Retry below if the prefetch failed.
      }
    }
    return widget.client.fetchDraftDetail(_token, draft.id);
  }

  void _showDetectedDraftDialogIfNeeded() {
    final AssessmentDraftSummary? draft = _detectedDraft;
    if (!mounted || _draftDialogShown || draft == null || draft.id <= 0) {
      return;
    }
    _draftDialogShown = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      showDialog<void>(
        context: context,
        barrierDismissible: false,
        barrierColor: Colors.black.withOpacity(.32),
        builder: (BuildContext dialogContext) {
          return PopScope(
            canPop: false,
            child: PadDialogViewport(
              child: _ErxinDraftResumeDialog(
                draft: draft,
                onRestart: _restartWithoutDetectedDraft,
                onContinue: () => _continueDetectedDraft(draft),
              ),
            ),
          );
        },
      );
    });
  }

  void _restartWithoutDetectedDraft() {
    if (!mounted) {
      return;
    }
    setState(() {
      _detectedDraft = null;
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _draftId = 0;
      _draftProgress = ErxinDraftProgress.empty;
      _itemPasses.clear();
      _itemRemarks.clear();
      _previousStartIndexByDomain.clear();
      _futureEndIndexByDomain.clear();
      _futureVisibleDomains.clear();
      _reviewMonthByDomain.clear();
      _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      _autoSaveText = '已开始新的测评';
    });
    _prefetchSelectedItem();
  }

  Future<bool> _continueDetectedDraft(AssessmentDraftSummary draft) async {
    if (draft.id <= 0 || _token.trim().isEmpty) {
      return false;
    }
    try {
      final ErxinDraftDetail detail = await _resolveDetectedDraftDetail(draft);
      if (!mounted) {
        return false;
      }
      setState(() {
        _applyDraftDetail(detail);
        _detectedDraft = null;
        _detectedDraftDetailDraftId = 0;
        _detectedDraftDetailRequest = null;
        _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
        if (_selectedItemNo <= 0) {
          _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
        }
        _autoSaveText = '已恢复最新草稿';
      });
      _prefetchSelectedItem();
      _revealSelectedItem();
      return true;
    } on Object catch (error) {
      if (mounted) {
        _showMessage('恢复草稿失败：$error');
      }
      return false;
    }
  }

  int get _mainAgeMonth {
    final double months = _actualAgeMonths(
      _birthDate,
      _assessmentDate,
    );
    if (months <= 0) {
      return 0;
    }
    int selected = _standardAgeMonths.first;
    double bestDistance = (months - selected).abs();
    for (final int ageMonth in _standardAgeMonths.skip(1)) {
      final double distance = (months - ageMonth).abs();
      if (distance < bestDistance) {
        selected = ageMonth;
        bestDistance = distance;
      }
    }
    return selected;
  }

  int get _mainAgeIndex {
    final int mainAge = _mainAgeMonth;
    return _standardAgeMonths.indexOf(mainAge);
  }

  int get _defaultPreviousStartIndex {
    final int index = _mainAgeIndex;
    return index < 0 ? 0 : math.max(0, index - 2);
  }

  int _previousStartIndexForDomain(String domainCode) {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return 0;
    }
    return (_previousStartIndexByDomain[domainCode] ??
            _defaultPreviousStartIndex)
        .clamp(0, index);
  }

  List<int> _previousMonthsForDomain(String domainCode) {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return <int>[];
    }
    final int start = _previousStartIndexForDomain(domainCode);
    return _standardAgeMonths.sublist(start, index);
  }

  List<int> _visibleMonthsBeforeFutureForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return <int>[
      mainAge,
      ...previous.reversed,
    ];
  }

  List<int> _futureMonthsForDomain(String domainCode) {
    if (!_futureVisibleDomains.contains(domainCode)) {
      return <int>[];
    }
    final int index = _mainAgeIndex;
    if (index < 0) {
      return <int>[];
    }
    final int end = (_futureEndIndexByDomain[domainCode] ??
            math.min(_standardAgeMonths.length - 1, index + 2))
        .clamp(index, _standardAgeMonths.length - 1);
    if (index + 1 > end) {
      return <int>[];
    }
    return _standardAgeMonths.sublist(index + 1, end + 1);
  }

  List<int> _visibleMonthsForDomain(String domainCode) {
    return <int>[
      ..._visibleMonthsBeforeFutureForDomain(domainCode),
      ..._futureMonthsForDomain(domainCode),
    ];
  }

  List<int> get _visibleMonths => _visibleMonthsForDomain(_selectedDomainCode);

  bool get _isReviewingRecord {
    return _reviewMonthByDomain.containsKey(_selectedDomainCode);
  }

  int? get _reviewMonth {
    return _reviewMonthByDomain[_selectedDomainCode];
  }

  List<int> _centerMonthsForDomain(String domainCode) {
    final int? reviewMonth = _reviewMonthByDomain[domainCode];
    if (reviewMonth != null) {
      return <int>[reviewMonth];
    }
    return <int>[
      for (final int month in _visibleMonthsForDomain(domainCode))
        if (_itemsFor(domainCode, month).isNotEmpty) month,
    ];
  }

  List<int> get _centerMonths => _centerMonthsForDomain(_selectedDomainCode);

  List<int> _recordMonthsForDomain(String domainCode) {
    final Set<int> months = <int>{..._visibleMonthsForDomain(domainCode)};
    for (final ErxinAgeGroup group in _template.ageGroups) {
      final bool hasAnsweredItem = group.items.any(
        (ErxinItemSummary item) =>
            item.domainCode == domainCode &&
            _itemPasses.containsKey(item.itemNo),
      );
      if (hasAnsweredItem) {
        months.add(group.ageMonth);
      }
    }

    final int mainAge = _mainAgeMonth;
    final List<int> previous = months
        .where((int month) => month > 0 && month < mainAge)
        .toList()
      ..sort((int left, int right) => right.compareTo(left));
    final List<int> future =
        months.where((int month) => month > mainAge).toList()..sort();
    return <int>[
      if (months.contains(mainAge)) mainAge,
      ...previous,
      ...future,
    ];
  }

  bool get _previousMonthsComplete {
    return _previousMonthsCompleteForDomain(_selectedDomainCode);
  }

  bool _previousMonthsCompleteForDomain(String domainCode) {
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return previous.isNotEmpty &&
        previous.every(
          (int month) => _ageMonthComplete(domainCode, month),
        );
  }

  bool get _hasPreviousBaseline {
    return _hasPreviousBaselineForDomain(_selectedDomainCode);
  }

  bool _hasPreviousBaselineForDomain(String domainCode) {
    return _previousBaselineMonthsForDomain(domainCode).isNotEmpty;
  }

  List<int> _previousBaselineMonthsForDomain(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex <= 1) {
      return const <int>[];
    }
    final int currentStart = _previousStartIndexForDomain(domainCode);
    for (int index = mainIndex - 2; index >= currentStart; index--) {
      final int lowerMonth = _standardAgeMonths[index];
      final int upperMonth = _standardAgeMonths[index + 1];
      if (_ageMonthAllPassed(domainCode, lowerMonth) &&
          _ageMonthAllPassed(domainCode, upperMonth)) {
        return <int>[upperMonth, lowerMonth];
      }
    }
    return const <int>[];
  }

  bool _mainMonthCompleteForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    return mainAge > 0 && _ageMonthComplete(domainCode, mainAge);
  }

  bool get _canContinuePreviousMonths {
    return _canContinuePreviousMonthsForDomain(_selectedDomainCode);
  }

  bool _canContinuePreviousMonthsForDomain(String domainCode) {
    if (_hasPreviousBaselineForDomain(domainCode) ||
        !_previousMonthsCompleteForDomain(domainCode) ||
        !_mainMonthCompleteForDomain(domainCode) ||
        _previousStartIndexForDomain(domainCode) <= 0) {
      return false;
    }
    return true;
  }

  bool get _canEnterFutureMonths {
    return _canEnterFutureMonthsForDomain(_selectedDomainCode);
  }

  bool _canEnterFutureMonthsForDomain(String domainCode) {
    return !_futureVisibleDomains.contains(domainCode) &&
        _mainAgeIndex < _standardAgeMonths.length - 1 &&
        _mainMonthCompleteForDomain(domainCode) &&
        _previousSearchResolvedForDomain(domainCode);
  }

  bool get _futureMonthsComplete {
    return _futureMonthsCompleteForDomain(_selectedDomainCode);
  }

  bool _futureMonthsCompleteForDomain(String domainCode) {
    final List<int> future = _futureMonthsForDomain(domainCode);
    return future.isNotEmpty &&
        future.every(
          (int month) => _ageMonthComplete(domainCode, month),
        );
  }

  bool get _hasFutureCeiling {
    return _hasFutureCeilingForDomain(_selectedDomainCode);
  }

  bool _hasFutureCeilingForDomain(String domainCode) {
    return _futureCeilingMonthsForDomain(domainCode).isNotEmpty;
  }

  List<int> _futureCeilingMonthsForDomain(String domainCode) {
    final List<int> future = _futureMonthsForDomain(domainCode);
    for (int index = 0; index < future.length - 1; index++) {
      final int current = future[index];
      final int next = future[index + 1];
      final int currentIndex = _standardAgeMonths.indexOf(current);
      final int nextIndex = _standardAgeMonths.indexOf(next);
      if (nextIndex - currentIndex != 1) {
        continue;
      }
      if (_ageMonthAllFailed(domainCode, current) &&
          _ageMonthAllFailed(domainCode, next)) {
        return <int>[current, next];
      }
    }
    return const <int>[];
  }

  bool get _canContinueFutureMonths {
    return _canContinueFutureMonthsForDomain(_selectedDomainCode);
  }

  bool _canContinueFutureMonthsForDomain(String domainCode) {
    if (!_futureVisibleDomains.contains(domainCode) ||
        _hasFutureCeilingForDomain(domainCode) ||
        !_futureMonthsCompleteForDomain(domainCode)) {
      return false;
    }
    final int index = _mainAgeIndex;
    final int end = _futureEndIndexByDomain[domainCode] ??
        math.min(_standardAgeMonths.length - 1, index + 2);
    return end < _standardAgeMonths.length - 1;
  }

  bool _previousBoundaryStopForDomain(String domainCode) {
    if (_hasPreviousBaselineForDomain(domainCode) ||
        _previousStartIndexForDomain(domainCode) > 0) {
      return false;
    }
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return previous.isEmpty || _previousMonthsCompleteForDomain(domainCode);
  }

  bool _previousSearchResolvedForDomain(String domainCode) {
    return _hasPreviousBaselineForDomain(domainCode) ||
        _previousBoundaryStopForDomain(domainCode);
  }

  bool _futureBoundaryStopForDomain(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0) {
      return false;
    }
    if (!_futureVisibleDomains.contains(domainCode)) {
      return mainIndex >= _standardAgeMonths.length - 1;
    }
    if (!_futureMonthsCompleteForDomain(domainCode)) {
      return false;
    }
    final int end = _futureEndIndexByDomain[domainCode] ??
        math.min(_standardAgeMonths.length - 1, mainIndex + 2);
    return end >= _standardAgeMonths.length - 1 &&
        !_hasFutureCeilingForDomain(domainCode);
  }

  bool _futureSearchResolvedForDomain(String domainCode) {
    return _hasFutureCeilingForDomain(domainCode) ||
        _futureBoundaryStopForDomain(domainCode);
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage.trim().isNotEmpty) {
      return _ErrorView(message: _errorMessage, onBack: widget.onBack);
    }
    return ColoredBox(
      color: _ErxinColors.page,
      child: Column(
        children: <Widget>[
          _Header(
            args: ErxinAssessmentLaunchArgs(
              studentId: _studentId,
              studentName: _studentName,
              studentAge: _resolvedStudentAgeText(),
              birthDate: _dateOnlyText(_birthDate),
              assessmentDate: _dateOnlyText(_assessmentDate),
              examinerName: _examinerName,
              scaleName: widget.args.scaleName,
            ),
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
                    completedDomainCount: _completedDomainCount(),
                    savedItemCount: _draftProgress.answeredItemCount,
                    onSelect: _selectDomain,
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

  Widget _buildWorkspace() {
    final List<int> months = _centerMonths;
    final bool reviewing = _isReviewingRecord;
    final int? reviewMonth = _reviewMonth;
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 10, 12, 6),
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                reviewing && reviewMonth != null
                    ? '${_domainName(_selectedDomainCode)} · ${reviewMonth}月龄记录'
                    : '${_domainName(_selectedDomainCode)} · 当前测查',
                style: const TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              if (reviewing) ...<Widget>[
                SizedBox(
                  height: 30,
                  child: OutlinedButton.icon(
                    onPressed: _returnToCurrentAssessment,
                    icon: const Icon(Icons.arrow_back_rounded, size: 14),
                    label: const Text('返回当前测查'),
                    style: OutlinedButton.styleFrom(
                      visualDensity: VisualDensity.compact,
                      padding: const EdgeInsets.symmetric(horizontal: 10),
                      textStyle: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
              ],
              _SmallBadge(text: '主测月龄 $_mainAgeMonth月龄', strong: true),
            ],
          ),
          const SizedBox(height: 8),
          Expanded(
            child: months.isEmpty
                ? const _CurrentItemsEmptyState(text: '请按右侧规则继续推进测查')
                : SingleChildScrollView(
                    padding: EdgeInsets.zero,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: <Widget>[
                        for (final int month in months)
                          _AgeMonthSection(
                            month: month,
                            isMainAge: month == _mainAgeMonth,
                            items: _itemsFor(_selectedDomainCode, month),
                            itemPasses: _itemPasses,
                            selectedItemNo: _selectedItemNo,
                            itemKeyFor: _itemRowKeyFor,
                            onSelectItem: _selectItem,
                            onScore: _scoreItem,
                          ),
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  List<_RuleRow> _recordRowsForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    final int? reviewMonth = _reviewMonthByDomain[domainCode];
    final List<int> baselineMonths = _previousBaselineMonthsForDomain(
      domainCode,
    );
    final List<int> ceilingMonths = _futureCeilingMonthsForDomain(domainCode);
    final bool previousBoundaryStop = _previousBoundaryStopForDomain(
      domainCode,
    );
    final bool previousResolved =
        _hasPreviousBaselineForDomain(domainCode) || previousBoundaryStop;
    final bool futureBoundaryStop = _futureBoundaryStopForDomain(domainCode);
    final bool futureResolved =
        _hasFutureCeilingForDomain(domainCode) || futureBoundaryStop;
    final List<_RuleRow> rows = <_RuleRow>[
      for (final int month in _recordMonthsForDomain(domainCode))
        _RuleRow(
          label: month == mainAge
              ? '主测月龄$month月龄'
              : month < mainAge
                  ? '往前$month月龄'
                  : '往后$month月龄',
          value: _recordStatusText(domainCode, month),
          done: _recordDone(domainCode, month),
          month: month,
          selected: reviewMonth == month,
        ),
      _RuleRow(
        label: '前测基线',
        value: _hasPreviousBaselineForDomain(domainCode)
            ? '已建立'
            : previousBoundaryStop
                ? '已到最低月龄'
                : '未形成',
        done: previousResolved,
        targetMonths: baselineMonths,
      ),
    ];
    if (_futureVisibleDomains.contains(domainCode) ||
        _hasAnsweredFutureMonth(domainCode) ||
        futureBoundaryStop) {
      rows.add(
        _RuleRow(
          label: '后测封顶',
          value: _hasFutureCeilingForDomain(domainCode)
              ? '已建立'
              : futureBoundaryStop
                  ? '已到最高月龄'
                  : '未形成',
          done: futureResolved,
          targetMonths: ceilingMonths,
        ),
      );
    }
    return rows;
  }

  Widget _buildRulePanel() {
    final String nextText = _nextActionText();
    final bool futureShown =
        _futureVisibleDomains.contains(_selectedDomainCode);
    final bool canContinuePrevious = _canContinuePreviousMonths;
    final bool canEnterFuture = _canEnterFutureMonths;
    final bool canContinueFuture = _canContinueFutureMonths;
    final bool domainComplete = _domainStopRuleComplete(_selectedDomainCode);
    final bool canLocateCurrentItem = !canContinuePrevious &&
        !canEnterFuture &&
        !canContinueFuture &&
        !domainComplete &&
        _firstPendingAssessmentItemNoForDomain(_selectedDomainCode) > 0;
    final String actionLabel = canContinuePrevious
        ? '继续往前测查'
        : canContinueFuture
            ? '继续往后测查'
            : canLocateCurrentItem
                ? futureShown
                    ? '往后测查已展开'
                    : '往前测查已展开'
                : domainComplete
                    ? '本能区测查完成'
                    : futureShown
                        ? _hasFutureCeiling
                            ? '本能区测查完成'
                            : '往后测查已展开'
                        : '进入往后测查';
    final VoidCallback? actionHandler = canContinuePrevious
        ? _continuePreviousMonths
        : canEnterFuture
            ? _enterFutureMonths
            : canContinueFuture
                ? _continueFutureMonths
                : canLocateCurrentItem
                    ? _locateCurrentAssessmentItem
                    : null;
    return Container(
      width: 296,
      padding: const EdgeInsets.fromLTRB(
        14,
        14,
        14,
        _erxinRulePanelBottomPadding,
      ),
      clipBehavior: Clip.antiAlias,
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '规则判断',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _RuleCard(
                  title: '下一步',
                  body: nextText,
                  icon: canContinuePrevious
                      ? Icons.keyboard_double_arrow_left_rounded
                      : canContinueFuture
                          ? Icons.keyboard_double_arrow_right_rounded
                          : canLocateCurrentItem
                              ? Icons.center_focus_strong_rounded
                              : domainComplete
                                  ? Icons.check_circle_rounded
                                  : Icons.arrow_forward_rounded,
                  color: _ErxinColors.blue,
                ),
                const SizedBox(height: 10),
                Row(
                  children: const <Widget>[
                    Text(
                      '测评记录',
                      style: TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Expanded(
                  child: _RuleChecklist(
                    rows: _recordRowsForDomain(_selectedDomainCode),
                    revealMonths: _recordRevealMonths,
                    revealSerial: _recordRevealSerial,
                    onTapMonth: _openAssessmentRecord,
                  ),
                ),
                const SizedBox(height: 12),
                const Text(
                  '测查推进',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 5),
                Text(
                  _hasFutureCeiling
                      ? '往后测查已形成连续两个标准月龄全不通过，本能区达到停止规则。'
                      : _futureBoundaryStopForDomain(_selectedDomainCode)
                          ? '已测至最高标准月龄，仍未形成后测封顶，按边界规则强行停止。'
                          : _hasPreviousBaseline
                              ? '前测已形成连续两个标准月龄全通过，继续往后寻找连续两个标准月龄全不通过。'
                              : _previousBoundaryStopForDomain(
                                      _selectedDomainCode)
                                  ? '已测至最低标准月龄，仍未形成前测基线，按边界规则进入往后测查。'
                                  : '前测尚未形成连续两个标准月龄全通过，需继续向更低月龄追测。',
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ErxinColors.muted,
                    fontSize: 12,
                    height: 1.28,
                  ),
                ),
                const SizedBox(height: 6),
                SizedBox(
                  width: double.infinity,
                  height: 38,
                  child: FilledButton(
                    onPressed: actionHandler,
                    style: FilledButton.styleFrom(
                      backgroundColor: _ErxinColors.blue,
                      disabledBackgroundColor: const Color(0xFFE2D6CE),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    child: Text(
                      actionLabel,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _RightRemarkSection(
            height: _erxinRightRemarkSectionHeight,
            itemNo: _selectedItemNo,
            remark: _itemRemarks[_selectedItemNo] ?? '',
            onChanged: _updateItemRemark,
            onEditingComplete: _finishItemRemarkEdit,
          ),
        ],
      ),
    );
  }

  Widget _buildDetailPanel() {
    final ErxinItemSummary? summary = _summaryByNo(_selectedItemNo);
    final String fallbackTitle =
        summary == null ? '当前题目说明' : '${summary.itemNo} ${summary.itemTitle}';
    return Container(
      height: _erxinDetailPanelHeight,
      padding: const EdgeInsets.fromLTRB(
        16,
        _erxinDetailPanelTopPadding,
        12,
        _erxinDetailPanelBottomPadding,
      ),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: _ErxinColors.line)),
      ),
      child: FutureBuilder<ErxinAssessmentItem>(
        future: _selectedItemNo > 0 ? _detailFuture(_selectedItemNo) : null,
        builder:
            (BuildContext context, AsyncSnapshot<ErxinAssessmentItem> snap) {
          final ErxinAssessmentItem? item = snap.data;
          final String title = item == null || item.itemNo <= 0
              ? fallbackTitle
              : '${item.itemNo} ${item.itemTitle}'
                  '${item.parentReportAllowed ? '（R）' : ''}';
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      '当前题目说明：$title',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _showFullItemDetail(item, summary),
                      icon: const Icon(Icons.open_in_full_rounded, size: 14),
                      label: const Text('完整说明'),
                      style: OutlinedButton.styleFrom(
                        visualDensity: VisualDensity.compact,
                        padding: const EdgeInsets.symmetric(horizontal: 9),
                        textStyle: const TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: _erxinDetailHeaderGap),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _DetailTextBox(
                        title: '操作方法',
                        text: item?.method ?? '正在加载题目操作方法...',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _DetailTextBox(
                        title: '通过标准',
                        text: item?.passCriteria ?? '正在加载通过标准...',
                      ),
                    ),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<ErxinAssessmentItem> _detailFuture(int itemNo) {
    final ErxinAssessmentItem? cached = _itemDetailCache[itemNo];
    if (cached != null) {
      return Future<ErxinAssessmentItem>.value(cached);
    }
    return _itemDetailFetches.putIfAbsent(itemNo, () async {
      final ErxinAssessmentItem item = await widget.client.fetchTemplateItem(
        _token,
        itemNo: itemNo,
      );
      _itemDetailCache[itemNo] = item;
      return item;
    });
  }

  void _prefetchSelectedItem() {
    if (_selectedItemNo > 0 && _token.trim().isNotEmpty) {
      _detailFuture(_selectedItemNo);
    }
  }

  GlobalKey _itemRowKeyFor(int itemNo) {
    return _itemRowKeys.putIfAbsent(
      itemNo,
      () => GlobalKey(debugLabel: 'erxin-item-row-$itemNo'),
    );
  }

  void _revealSelectedItem() {
    final int itemNo = _selectedItemNo;
    if (itemNo <= 0) {
      return;
    }
    void reveal() {
      if (!mounted || _selectedItemNo != itemNo) {
        return;
      }
      final BuildContext? rowContext = _itemRowKeys[itemNo]?.currentContext;
      if (rowContext == null) {
        return;
      }
      unawaited(
        Scrollable.ensureVisible(
          rowContext,
          alignment: 0.1,
          duration: const Duration(milliseconds: 260),
          curve: Curves.easeOutCubic,
        ),
      );
    }

    WidgetsBinding.instance.addPostFrameCallback((_) {
      reveal();
      unawaited(
        Future<void>.delayed(
          const Duration(milliseconds: 120),
          reveal,
        ),
      );
    });
  }

  void _selectDomain(String domainCode) {
    setState(() {
      _selectedDomainCode = domainCode;
      _reviewMonthByDomain.remove(domainCode);
      _selectedItemNo = _firstCurrentItemNo(domainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(domainCode);
      }
    });
    _prefetchSelectedItem();
  }

  void _selectItem(int itemNo) {
    setState(() => _selectedItemNo = itemNo);
    _prefetchSelectedItem();
  }

  void _updateItemRemark(int itemNo, String remark) {
    if (itemNo <= 0) {
      return;
    }
    final String normalized = remark.trim();
    setState(() {
      if (normalized.isEmpty) {
        _itemRemarks.remove(itemNo);
      } else {
        _itemRemarks[itemNo] = normalized;
      }
    });
  }

  void _finishItemRemarkEdit(int itemNo) {
    if (itemNo <= 0 ||
        !_itemPasses.containsKey(itemNo) ||
        _token.trim().isEmpty) {
      return;
    }
    unawaited(_saveItem(itemNo));
  }

  void _scoreItem(int itemNo, bool passed) {
    unawaited(_scoreItemInternal(itemNo, passed));
  }

  Future<void> _scoreItemInternal(int itemNo, bool passed) async {
    final ErxinItemSummary? summary = _summaryByNo(itemNo);
    final String domainCode = summary?.domainCode ?? _selectedDomainCode;
    final bool historyReview = _reviewMonthByDomain[domainCode] != null &&
        summary?.ageMonth == _reviewMonthByDomain[domainCode];
    final bool changedExisting =
        _itemPasses.containsKey(itemNo) && _itemPasses[itemNo] != passed;
    if (historyReview && changedExisting) {
      final bool confirmed = await _confirmHistoryChange(
        itemNo: itemNo,
        passed: passed,
      );
      if (!confirmed || !mounted) {
        return;
      }
    }
    if (!mounted) {
      return;
    }
    int revealedItemNo = itemNo;
    setState(() {
      _itemPasses[itemNo] = passed;
      _reconcileAfterScore(domainCode);
      _selectedDomainCode = domainCode;
      final int nextSelected = _nextSelectedItemNoForDomain(domainCode, itemNo);
      _selectedItemNo = nextSelected > 0 ? nextSelected : itemNo;
      revealedItemNo = _selectedItemNo;
      _autoSaveText = _token.trim().isEmpty ? '本地已记录' : '自动保存中...';
    });
    _prefetchSelectedItem();
    if (revealedItemNo != itemNo) {
      _revealSelectedItem();
    }
    if (_token.trim().isNotEmpty) {
      unawaited(_saveItem(itemNo));
    }
  }

  void _openAssessmentRecord(int month) {
    setState(() {
      _reviewMonthByDomain[_selectedDomainCode] = month;
      _selectedItemNo = _firstItemNoForMonth(_selectedDomainCode, month);
      _autoSaveText = '正在查看历史记录';
    });
    _prefetchSelectedItem();
  }

  void _returnToCurrentAssessment() {
    setState(() {
      _reviewMonthByDomain.remove(_selectedDomainCode);
      _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
      _autoSaveText = '返回当前测查';
    });
    _prefetchSelectedItem();
  }

  void _locateCurrentAssessmentItem() {
    setState(() {
      _reviewMonthByDomain.remove(_selectedDomainCode);
      final int pendingItemNo =
          _firstPendingAssessmentItemNoForDomain(_selectedDomainCode);
      _selectedItemNo = pendingItemNo > 0
          ? pendingItemNo
          : _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
      _autoSaveText = '已定位当前题';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  Future<bool> _confirmHistoryChange({
    required int itemNo,
    required bool passed,
  }) async {
    final ErxinItemSummary? summary = _summaryByNo(itemNo);
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      barrierColor: Colors.black.withOpacity(.24),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: Center(
            child: _HistoryChangeConfirmDialog(
              itemTitle: summary == null
                  ? '第$itemNo题'
                  : '${summary.itemNo} ${summary.itemTitle}',
              nextStatus: passed ? '通过' : '不通过',
              onCancel: () => Navigator.of(dialogContext).pop(false),
              onConfirm: () => Navigator.of(dialogContext).pop(true),
            ),
          ),
        );
      },
    );
    return confirmed == true;
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
      key: 'erxin-top-message',
    );
  }

  Future<ErxinDraftDetail?> _saveDraft({bool silent = false}) {
    final Future<ErxinDraftDetail?>? inFlight = _saveDraftFuture;
    if (inFlight != null) {
      if (!silent && _saveDraftFutureSilent && !_saveDraftJoinedByManual) {
        _saveDraftJoinedByManual = true;
        return _joinSilentDraftSave(inFlight);
      }
      return inFlight;
    }
    late final Future<ErxinDraftDetail?> trackedFuture;
    _saveDraftFutureSilent = silent;
    _saveDraftJoinedByManual = false;
    trackedFuture = _performSaveDraft(silent: silent).whenComplete(() {
      if (identical(_saveDraftFuture, trackedFuture)) {
        _saveDraftFuture = null;
        _saveDraftFutureSilent = false;
        _saveDraftJoinedByManual = false;
      }
    });
    _saveDraftFuture = trackedFuture;
    return trackedFuture;
  }

  Future<ErxinDraftDetail?> _joinSilentDraftSave(
    Future<ErxinDraftDetail?> inFlight,
  ) async {
    if (mounted) {
      setState(() => _autoSaveText = '草稿保存中...');
    }
    final ErxinDraftDetail? detail = await inFlight;
    if (!mounted) {
      return detail;
    }
    if (detail == null) {
      _showMessage('保存草稿失败，请稍后重试');
    } else {
      _showMessage('草稿已保存', tone: PadMessageTone.success);
    }
    return detail;
  }

  Future<ErxinDraftDetail?> _performSaveDraft({required bool silent}) async {
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿');
      return null;
    }
    if (_studentId <= 0 || _studentName.trim().isEmpty) {
      _showMessage('缺少儿童信息，无法保存草稿');
      return null;
    }
    setState(() => _savingDraft = true);
    try {
      final ErxinDraftDetail detail = await _sendSaveDraftWithRetry();
      if (!mounted) {
        return detail;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '草稿已保存';
      });
      if (!silent) {
        _showMessage('草稿已保存', tone: PadMessageTone.success);
      }
      return detail;
    } on Object catch (error) {
      if (!silent) {
        _showMessage('保存草稿失败：$error');
      } else if (mounted) {
        setState(() => _autoSaveText = '保存失败');
      }
      return null;
    } finally {
      if (mounted) {
        setState(() => _savingDraft = false);
      }
    }
  }

  Future<ErxinDraftDetail> _sendSaveDraftWithRetry() async {
    final int attemptedDraftId = _draftId;
    try {
      return await widget.client.saveDraft(
        _token,
        _buildDraftPayload(),
      );
    } on Object catch (error) {
      if (attemptedDraftId <= 0 || !_isDraftNotFoundError(error)) {
        rethrow;
      }
      if (mounted) {
        setState(() {
          _draftId = 0;
          _autoSaveText = '正在创建新草稿...';
        });
      } else {
        _draftId = 0;
      }
      return widget.client.saveDraft(
        _token,
        _buildDraftPayload(),
      );
    }
  }

  Future<void> _saveItem(int itemNo) async {
    if (itemNo <= 0 || !_itemPasses.containsKey(itemNo)) {
      return;
    }
    int draftId = _draftId;
    if (draftId <= 0) {
      final ErxinDraftDetail? created = await _saveDraft(silent: true);
      draftId = created?.id ?? _draftId;
    }
    if (draftId <= 0 || !mounted) {
      return;
    }
    setState(() {
      _autoSaveText = '自动保存中...';
    });
    try {
      final ErxinDraftDetail detail = await _sendSaveDraftItem(
        itemNo: itemNo,
        draftId: draftId,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '已自动保存';
      });
    } on Object catch (error) {
      if (_isDraftNotFoundError(error)) {
        final bool retried = await _retrySaveItemAfterMissingDraft(itemNo);
        if (retried) {
          return;
        }
      }
      if (mounted) {
        setState(() {
          _autoSaveText = '保存失败';
        });
      }
      _showMessage('第$itemNo题自动保存失败：$error');
    }
  }

  Future<ErxinDraftDetail> _sendSaveDraftItem({
    required int itemNo,
    required int draftId,
  }) {
    return widget.client.saveDraftItem(
      _token,
      <String, dynamic>{
        'draftId': draftId,
        'itemNo': itemNo,
        'passed': _itemPasses[itemNo],
        'remark': _itemRemarks[itemNo]?.trim() ?? '',
      },
    );
  }

  Future<bool> _retrySaveItemAfterMissingDraft(int itemNo) async {
    if (!mounted || itemNo <= 0 || !_itemPasses.containsKey(itemNo)) {
      return false;
    }
    setState(() {
      _draftId = 0;
      _autoSaveText = '正在重建草稿...';
    });
    try {
      final ErxinDraftDetail? created = await _saveDraft(silent: true);
      final int retryDraftId = created?.id ?? _draftId;
      if (retryDraftId <= 0 || !mounted) {
        return false;
      }
      final ErxinDraftDetail detail = await _sendSaveDraftItem(
        itemNo: itemNo,
        draftId: retryDraftId,
      );
      if (!mounted) {
        return true;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '已自动保存';
      });
      return true;
    } on Object {
      return false;
    }
  }

  bool _isDraftNotFoundError(Object error) {
    return error
        .toString()
        .toLowerCase()
        .contains('assessment draft not found');
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    final String? blocker = _localSubmitBlocker();
    if (blocker != null) {
      _showMessage(blocker);
      return;
    }
    setState(() => _submitting = true);
    try {
      final ErxinDraftDetail? detail = await _saveDraft(silent: true);
      final int draftId = detail?.id ?? _draftId;
      if (draftId <= 0) {
        _showMessage('请先保存草稿，再提交正式记录');
        return;
      }
      await widget.client.submitDraft(_token, draftId);
      if (!mounted) {
        return;
      }
      _showMessage('已提交正式测评记录', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 650));
      if (mounted) {
        widget.onBack();
      }
    } on Object catch (error) {
      if (mounted) {
        _showMessage('提交记录失败：$error');
      }
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  Map<String, dynamic> _buildDraftPayload() {
    return <String, dynamic>{
      if (_draftId > 0) 'id': _draftId,
      'studentId': _studentId,
      'studentName': _studentName.trim(),
      'examinerName': _examinerName.trim(),
      'birthDate': _dateOnlyText(_birthDate),
      'assessmentDate': _dateOnlyText(_assessmentDate),
      'itemPassList': _itemPassList(),
      if (_itemRemarks.isNotEmpty) 'itemRemarkList': _itemRemarkList(),
    };
  }

  List<Map<String, dynamic>> _itemPassList() {
    final List<int> itemNos = _itemPasses.keys.toList()..sort();
    return <Map<String, dynamic>>[
      for (final int itemNo in itemNos)
        <String, dynamic>{
          'itemNo': itemNo,
          'passed': _itemPasses[itemNo],
          'remark': _itemRemarks[itemNo]?.trim() ?? '',
        },
    ];
  }

  List<Map<String, dynamic>> _itemRemarkList() {
    final List<int> itemNos = _itemRemarks.keys.toList()..sort();
    return <Map<String, dynamic>>[
      for (final int itemNo in itemNos)
        if ((_itemRemarks[itemNo] ?? '').trim().isNotEmpty)
          <String, dynamic>{
            'itemNo': itemNo,
            'remark': (_itemRemarks[itemNo] ?? '').trim(),
          },
    ];
  }

  void _applyDraftDetail(ErxinDraftDetail detail) {
    _draftId = detail.id;
    _draftProgress = detail.progress;
    if (detail.studentId > 0) {
      _studentId = detail.studentId;
    }
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
    _itemPasses
      ..clear()
      ..addAll(detail.input.itemPasses);
    _itemRemarks
      ..clear()
      ..addAll(detail.input.itemRemarks);
    _restoreAssessmentWindowsFromAnswers();
  }

  void _restoreAssessmentWindowsFromAnswers() {
    _previousStartIndexByDomain.clear();
    _futureEndIndexByDomain.clear();
    _futureVisibleDomains.clear();
    _reviewMonthByDomain.clear();
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0) {
      return;
    }
    for (final ErxinDomain domain in _template.domains) {
      final String domainCode = domain.domainCode;
      int previousStart = _defaultPreviousStartIndex;
      bool hasPreviousAnswer = false;
      int futureEnd = math.min(_standardAgeMonths.length - 1, mainIndex + 2);
      bool hasFutureAnswer = false;
      for (final ErxinAgeGroup group in _template.ageGroups) {
        final int ageIndex = _standardAgeMonths.indexOf(group.ageMonth);
        if (ageIndex < 0) {
          continue;
        }
        final bool hasAnsweredItem = group.items.any(
          (ErxinItemSummary item) =>
              item.domainCode == domainCode &&
              _itemPasses.containsKey(item.itemNo),
        );
        if (!hasAnsweredItem) {
          continue;
        }
        if (ageIndex < mainIndex) {
          hasPreviousAnswer = true;
          previousStart = math.min(previousStart, ageIndex);
        } else if (ageIndex > mainIndex) {
          hasFutureAnswer = true;
          futureEnd = math.max(futureEnd, ageIndex);
        }
      }
      if (hasPreviousAnswer) {
        _previousStartIndexByDomain[domainCode] = previousStart;
      }
      if (hasFutureAnswer) {
        _futureVisibleDomains.add(domainCode);
        _futureEndIndexByDomain[domainCode] =
            math.min(futureEnd, _standardAgeMonths.length - 1);
      }
      _trimPreviousWindowToActiveBaseline(domainCode);
      if (_futureVisibleDomains.contains(domainCode)) {
        _futureEndIndexByDomain[domainCode] = _trimFutureEndToActiveCeiling(
          domainCode,
          _futureEndIndexByDomain[domainCode] ?? futureEnd,
        );
      }
    }
  }

  String? _localSubmitBlocker() {
    if (_token.trim().isEmpty) {
      return '请先登录后再提交正式记录';
    }
    if (_birthDate.trim().isEmpty || _assessmentDate.trim().isEmpty) {
      return '缺少出生日期或测查日期，不能提交正式记录';
    }
    for (final ErxinDomain domain in _template.domains) {
      final String code = domain.domainCode;
      for (final int month in _visibleMonthsForDomain(code)) {
        for (final ErxinItemSummary item in _itemsFor(code, month)) {
          if (!_itemPasses.containsKey(item.itemNo)) {
            _selectDomain(code);
            return '${_domainName(code)}还有当前可见题目未记录，请补全后再提交';
          }
        }
      }
      if (_canContinuePreviousMonthsForDomain(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未形成连续两个往前月龄全通过，请继续往前测查';
      }
      if (_canEnterFutureMonthsForDomain(code)) {
        _selectDomain(code);
        return _hasPreviousBaselineForDomain(code)
            ? '${_domainName(code)}已建立前测基线，请先进入往后测查'
            : '${_domainName(code)}已测至最低月龄，请先进入往后测查';
      }
      if (_canContinueFutureMonthsForDomain(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未形成连续两个往后月龄全不通过，请继续往后测查';
      }
      if (!_domainStopRuleComplete(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未满足儿心量表停止规则，请完成规则提示的测查';
      }
    }
    return null;
  }

  void _continuePreviousMonths() {
    final String domainCode = _selectedDomainCode;
    final int currentStart = _previousStartIndexForDomain(domainCode);
    if (currentStart <= 0) {
      return;
    }
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      int start = currentStart;
      while (start > 0) {
        final int lowestVisibleMonth = _standardAgeMonths[start];
        final int step =
            _ageMonthAllPassed(domainCode, lowestVisibleMonth) ? 1 : 2;
        final int nextStart = math.max(0, start - step);
        if (nextStart == start) {
          break;
        }
        _previousStartIndexByDomain[domainCode] = nextStart;
        start = nextStart;
        if (_firstPendingCurrentItemNo(domainCode) > 0 ||
            _hasPreviousBaselineForDomain(domainCode) ||
            !_previousMonthsCompleteForDomain(domainCode)) {
          break;
        }
      }
      _selectedItemNo = _firstPendingCurrentItemNo(domainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstCurrentItemNo(domainCode);
      }
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(domainCode);
      }
      final List<int> addedMonths = _standardAgeMonths
          .sublist(start, currentStart)
          .reversed
          .toList(growable: false);
      if (addedMonths.isNotEmpty) {
        _recordRevealMonths = addedMonths;
        _recordRevealSerial++;
      }
      _autoSaveText = '已追加往前测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _enterFutureMonths() {
    final int index = _mainAgeIndex;
    final String domainCode = _selectedDomainCode;
    final int endIndex = math.min(_standardAgeMonths.length - 1, index + 2);
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      _futureVisibleDomains.add(domainCode);
      _futureEndIndexByDomain[domainCode] = endIndex;
      _selectedItemNo = _firstItemNoForMonth(
        domainCode,
        _standardAgeMonths[index + 1],
      );
      _recordRevealMonths = _standardAgeMonths
          .sublist(index + 1, endIndex + 1)
          .toList(growable: false);
      if (_recordRevealMonths.isNotEmpty) {
        _recordRevealSerial++;
      }
      _autoSaveText = '已进入往后测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _continueFutureMonths() {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return;
    }
    final String domainCode = _selectedDomainCode;
    final int currentEnd = _futureEndIndexByDomain[domainCode] ??
        math.min(_standardAgeMonths.length - 1, index + 2);
    if (currentEnd >= _standardAgeMonths.length - 1) {
      return;
    }
    final int highestVisibleMonth = _standardAgeMonths[currentEnd];
    final int step = _ageMonthAllFailed(
      domainCode,
      highestVisibleMonth,
    )
        ? 1
        : 2;
    final int nextEnd =
        math.min(_standardAgeMonths.length - 1, currentEnd + step);
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      _futureEndIndexByDomain[domainCode] = nextEnd;
      _selectedItemNo = _firstItemNoForMonth(
        domainCode,
        _standardAgeMonths[currentEnd + 1],
      );
      _recordRevealMonths = _standardAgeMonths
          .sublist(currentEnd + 1, nextEnd + 1)
          .toList(growable: false);
      if (_recordRevealMonths.isNotEmpty) {
        _recordRevealSerial++;
      }
      _autoSaveText = '已追加往后测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _showFullItemDetail(
    ErxinAssessmentItem? item,
    ErxinItemSummary? summary,
  ) {
    final String title = item == null || item.itemNo <= 0
        ? summary == null
            ? '题目说明'
            : '${summary.itemNo} ${summary.itemTitle}'
        : '${item.itemNo} ${item.itemTitle}';
    showDialog<void>(
      context: context,
      barrierColor: Colors.black.withOpacity(.24),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: Center(
            child: AlertDialog(
              title: Text(title),
              content: SizedBox(
                width: 640,
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      _DialogTextBlock(title: '操作方法', text: item?.method ?? ''),
                      const SizedBox(height: 18),
                      _DialogTextBlock(
                        title: '通过标准',
                        text: item?.passCriteria ?? '',
                      ),
                    ],
                  ),
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('关闭'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  _DomainProgress _domainProgress(String domainCode) {
    final List<ErxinItemSummary> items = <ErxinItemSummary>[
      for (final int month in _visibleMonthsForDomain(domainCode))
        ..._itemsFor(domainCode, month),
    ];
    final int answered = items
        .where((ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo))
        .length;
    return _DomainProgress(answered: answered, total: items.length);
  }

  int _completedDomainCount() {
    return _template.domains
        .where(
            (ErxinDomain domain) => _domainStopRuleComplete(domain.domainCode))
        .length
        .clamp(0, 5);
  }

  bool _domainStopRuleComplete(String domainCode) {
    if (!_mainMonthCompleteForDomain(domainCode)) {
      return false;
    }
    if (!_previousSearchResolvedForDomain(domainCode)) {
      return false;
    }
    return _futureSearchResolvedForDomain(domainCode);
  }

  String _recordStatusText(String domainCode, int month) {
    if (month == _mainAgeMonth) {
      return _ageMonthComplete(domainCode, month) ? '已完成' : '未完成';
    }
    if (month < _mainAgeMonth) {
      return _ageMonthAllPassed(domainCode, month)
          ? '全通过'
          : _ageMonthComplete(domainCode, month)
              ? '未全通过'
              : '未完成';
    }
    return _ageMonthAllFailed(domainCode, month)
        ? '全不通过'
        : _ageMonthComplete(domainCode, month)
            ? '未全不通过'
            : '未完成';
  }

  bool _recordDone(String domainCode, int month) {
    if (month == _mainAgeMonth) {
      return _ageMonthComplete(domainCode, month);
    }
    if (month < _mainAgeMonth) {
      return _ageMonthAllPassed(domainCode, month);
    }
    return _ageMonthAllFailed(domainCode, month);
  }

  bool _hasAnsweredFutureMonth(String domainCode) {
    return _highestAnsweredFutureIndex(domainCode) >= 0;
  }

  int _highestAnsweredFutureIndex(String domainCode) {
    int highest = -1;
    for (final ErxinAgeGroup group in _template.ageGroups) {
      final int index = _standardAgeMonths.indexOf(group.ageMonth);
      if (index <= _mainAgeIndex) {
        continue;
      }
      final bool hasAnswered = group.items.any(
        (ErxinItemSummary item) =>
            item.domainCode == domainCode &&
            _itemPasses.containsKey(item.itemNo),
      );
      if (hasAnswered) {
        highest = math.max(highest, index);
      }
    }
    return highest;
  }

  void _reconcileAfterScore(String domainCode) {
    _trimPreviousWindowToActiveBaseline(domainCode);
    final bool previousReady = _mainMonthCompleteForDomain(domainCode) &&
        _previousSearchResolvedForDomain(domainCode);
    if (!previousReady) {
      _futureVisibleDomains.remove(domainCode);
      _futureEndIndexByDomain.remove(domainCode);
      return;
    }

    final int highestAnsweredFuture = _highestAnsweredFutureIndex(domainCode);
    if (!_futureVisibleDomains.contains(domainCode) &&
        highestAnsweredFuture < 0) {
      _futureEndIndexByDomain.remove(domainCode);
      return;
    }

    _futureVisibleDomains.add(domainCode);
    final int defaultEnd =
        math.min(_standardAgeMonths.length - 1, _mainAgeIndex + 2);
    final int existingEnd = _futureEndIndexByDomain[domainCode] ?? defaultEnd;
    final int end = math.max(existingEnd, highestAnsweredFuture);
    final int normalizedEnd =
        math.max(defaultEnd, math.min(end, _standardAgeMonths.length - 1));
    _futureEndIndexByDomain[domainCode] =
        _trimFutureEndToActiveCeiling(domainCode, normalizedEnd);
  }

  void _trimPreviousWindowToActiveBaseline(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex <= 0) {
      return;
    }
    final int currentStart = _previousStartIndexForDomain(domainCode);
    int? bestStart;
    for (int index = mainIndex - 2; index >= currentStart; index--) {
      final int current = _standardAgeMonths[index];
      final int next = _standardAgeMonths[index + 1];
      if (_ageMonthAllPassed(domainCode, current) &&
          _ageMonthAllPassed(domainCode, next)) {
        bestStart = index;
        break;
      }
    }
    if (bestStart != null && bestStart != currentStart) {
      _previousStartIndexByDomain[domainCode] = bestStart;
    }
  }

  int _trimFutureEndToActiveCeiling(String domainCode, int endIndex) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0 || endIndex <= mainIndex + 1) {
      return endIndex;
    }
    for (int index = mainIndex + 1; index < endIndex; index++) {
      final int current = _standardAgeMonths[index];
      final int next = _standardAgeMonths[index + 1];
      if (_ageMonthAllFailed(domainCode, current) &&
          _ageMonthAllFailed(domainCode, next)) {
        return index + 1;
      }
    }
    return endIndex;
  }

  List<ErxinItemSummary> _itemsFor(String domainCode, int ageMonth) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      if (group.ageMonth == ageMonth) {
        return group.items
            .where((ErxinItemSummary item) => item.domainCode == domainCode)
            .toList();
      }
    }
    return <ErxinItemSummary>[];
  }

  bool _ageMonthComplete(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every(
            (ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo));
  }

  bool _ageMonthAllPassed(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items
            .every((ErxinItemSummary item) => _itemPasses[item.itemNo] == true);
  }

  bool _ageMonthAllFailed(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every(
            (ErxinItemSummary item) => _itemPasses[item.itemNo] == false);
  }

  int _firstVisibleItemNo(String domainCode) {
    for (final int month in _visibleMonthsForDomain(domainCode)) {
      final int itemNo = _firstItemNoForMonth(domainCode, month);
      if (itemNo > 0) {
        return itemNo;
      }
    }
    return 0;
  }

  int _firstCurrentItemNo(String domainCode) {
    final int pendingItemNo = _firstPendingCurrentItemNo(domainCode);
    if (pendingItemNo > 0) {
      return pendingItemNo;
    }
    for (final int month in _centerMonthsForDomain(domainCode)) {
      final int itemNo = _firstItemNoForMonth(domainCode, month);
      if (itemNo > 0) {
        return itemNo;
      }
    }
    return 0;
  }

  int _firstPendingCurrentItemNo(String domainCode) {
    for (final int month in _centerMonthsForDomain(domainCode)) {
      for (final ErxinItemSummary item in _itemsFor(domainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return item.itemNo;
        }
      }
    }
    return 0;
  }

  int _firstPendingAssessmentItemNoForDomain(String domainCode) {
    for (final int month in _visibleMonthsForDomain(domainCode)) {
      for (final ErxinItemSummary item in _itemsFor(domainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return item.itemNo;
        }
      }
    }
    return 0;
  }

  int _nextSelectedItemNoForDomain(String domainCode, int fallbackItemNo) {
    if (_reviewMonthByDomain.containsKey(domainCode)) {
      return fallbackItemNo;
    }
    final int pendingItemNo = _firstPendingCurrentItemNo(domainCode);
    return pendingItemNo > 0 ? pendingItemNo : fallbackItemNo;
  }

  int _firstItemNoForMonth(String domainCode, int ageMonth) {
    for (final ErxinItemSummary item in _itemsFor(domainCode, ageMonth)) {
      return item.itemNo;
    }
    return 0;
  }

  ErxinItemSummary? _summaryByNo(int itemNo) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      for (final ErxinItemSummary item in group.items) {
        if (item.itemNo == itemNo) {
          return item;
        }
      }
    }
    return null;
  }

  String _domainName(String domainCode) {
    for (final ErxinDomain domain in _template.domains) {
      if (domain.domainCode == domainCode) {
        return domain.domainName.trim().isEmpty
            ? domain.domainCode
            : domain.domainName;
      }
    }
    return domainCode;
  }

  String _resolvedStudentAgeText() {
    final String explicit = _studentAge.trim();
    if (explicit.isNotEmpty && explicit != '未知') {
      return explicit;
    }
    final double months = _actualAgeMonths(_birthDate, _assessmentDate);
    if (months <= 0) {
      return '未知';
    }
    final int years = (months ~/ 12).clamp(0, 6);
    final int remainingMonths = (months - years * 12).round().clamp(0, 11);
    if (years > 0) {
      return remainingMonths > 0 ? '$years岁$remainingMonths个月' : '$years岁';
    }
    return '$remainingMonths个月';
  }

  String _nextActionText() {
    for (final int month in _visibleMonths) {
      for (final ErxinItemSummary item
          in _itemsFor(_selectedDomainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return '完成$month月龄第${item.itemNo}题';
        }
      }
    }
    if (_canEnterFutureMonths) {
      return _hasPreviousBaseline ? '前测已达标，可以进入往后测查' : '已到最低月龄，前测强行停止，可进入往后测查';
    }
    if (_canContinuePreviousMonths) {
      final int currentStart =
          _previousStartIndexForDomain(_selectedDomainCode);
      final int lowestVisibleMonth = _standardAgeMonths[currentStart];
      final int step =
          _ageMonthAllPassed(_selectedDomainCode, lowestVisibleMonth) ? 1 : 2;
      final int nextStart = math.max(0, currentStart - step);
      final List<int> nextMonths =
          _standardAgeMonths.sublist(nextStart, currentStart);
      return '前测未形成连续全通过，继续追加${nextMonths.join('月、')}月';
    }
    if (_canContinueFutureMonths) {
      final int index = _mainAgeIndex;
      final int currentEnd = _futureEndIndexByDomain[_selectedDomainCode] ??
          math.min(_standardAgeMonths.length - 1, index + 2);
      final int highestVisibleMonth = _standardAgeMonths[currentEnd];
      final int step = _ageMonthAllFailed(
        _selectedDomainCode,
        highestVisibleMonth,
      )
          ? 1
          : 2;
      final int nextEnd =
          math.min(_standardAgeMonths.length - 1, currentEnd + step);
      final List<int> nextMonths =
          _standardAgeMonths.sublist(currentEnd + 1, nextEnd + 1);
      return '后测未形成连续全不通过，继续追加${nextMonths.join('月、')}月';
    }
    if (_hasFutureCeiling) {
      return '当前能区停止规则已满足，可切换下一个能区或提交';
    }
    if (_domainStopRuleComplete(_selectedDomainCode)) {
      return '当前能区边界停止，可切换下一个能区或提交';
    }
    if (_futureVisibleDomains.contains(_selectedDomainCode) &&
        !_hasFutureCeiling) {
      return _futureBoundaryStopForDomain(_selectedDomainCode)
          ? '已到最高月龄，后测强行停止'
          : _futureMonthsComplete
              ? '已到最高可追测月龄，仍未形成连续全不通过'
              : '先完成当前可见的往后测查题目';
    }
    if (!_hasPreviousBaseline) {
      return _previousBoundaryStopForDomain(_selectedDomainCode)
          ? '已到最低可追测月龄，仍未形成连续全通过'
          : _previousMonthsComplete
              ? '已到最低可追测月龄，仍未形成连续全通过'
              : '先完成当前可见的往前测查题目';
    }
    return '当前可见题目已完成';
  }
}

class _ErxinDraftResumeDialog extends StatefulWidget {
  const _ErxinDraftResumeDialog({
    required this.draft,
    required this.onRestart,
    required this.onContinue,
  });

  final AssessmentDraftSummary draft;
  final VoidCallback onRestart;
  final Future<bool> Function() onContinue;

  @override
  State<_ErxinDraftResumeDialog> createState() =>
      _ErxinDraftResumeDialogState();
}

class _ErxinDraftResumeDialogState extends State<_ErxinDraftResumeDialog> {
  bool _continuing = false;

  void _handleRestart() {
    if (_continuing) {
      return;
    }
    Navigator.of(context).pop();
    widget.onRestart();
  }

  Future<void> _handleContinue() async {
    if (_continuing) {
      return;
    }
    setState(() => _continuing = true);
    final bool restored = await widget.onContinue();
    if (!mounted) {
      return;
    }
    if (restored) {
      Navigator.of(context).pop();
      return;
    }
    setState(() => _continuing = false);
  }

  @override
  Widget build(BuildContext context) {
    final int answered = widget.draft.answeredItemCount;
    final int percent = widget.draft.completionPercentInt;
    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 32),
      backgroundColor: Colors.transparent,
      child: Container(
        width: 520,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          boxShadow: _erxinShadow(
            color: const Color(0x33000000),
            blur: 30,
            offset: const Offset(0, 18),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Padding(
              padding: EdgeInsets.fromLTRB(20, 20, 20, 20),
              child: Text(
                '发现未完成草稿',
                style: TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 19,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            const Divider(height: 1, color: _ErxinColors.line),
            Padding(
              padding: const EdgeInsets.fromLTRB(30, 30, 30, 30),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  const Text(
                    '当前儿童存在一份未提交的儿心量表测评草稿。',
                    style: TextStyle(
                      color: _ErxinColors.ink,
                      fontSize: 15,
                      height: 1.2,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 22),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFFBF7),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: _ErxinColors.line),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        _ErxinDraftResumeMeta(
                          label: '已记录',
                          value: '$answered 题 · $percent%',
                        ),
                        const SizedBox(height: 13),
                        _ErxinDraftResumeMeta(
                          label: '更新时间',
                          value: _formatErxinDateTime(widget.draft.updatedTime),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const Divider(height: 1, color: _ErxinColors.line),
            Padding(
              padding: const EdgeInsets.fromLTRB(30, 18, 30, 20),
              child: Align(
                alignment: Alignment.centerRight,
                child: SizedBox(
                  height: 42,
                  child: _continuing
                      ? const SizedBox(
                          width: 112,
                          child: Center(
                            child: SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                          ),
                        )
                      : Row(
                          mainAxisSize: MainAxisSize.min,
                          children: <Widget>[
                            OutlinedButton(
                              onPressed: _handleRestart,
                              child: const Text('重新测评'),
                            ),
                            const SizedBox(width: 12),
                            FilledButton(
                              onPressed: _handleContinue,
                              child: const Text('继续测评'),
                            ),
                          ],
                        ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ErxinDraftResumeMeta extends StatelessWidget {
  const _ErxinDraftResumeMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Text.rich(
      TextSpan(
        children: <InlineSpan>[
          TextSpan(text: '$label：'),
          TextSpan(
            text: value,
            style: const TextStyle(
              color: _ErxinColors.ink,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
      style: const TextStyle(
        color: _ErxinColors.body,
        fontSize: 14,
        height: 1.2,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _Header extends StatelessWidget {
  const _Header({
    required this.args,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final ErxinAssessmentLaunchArgs args;
  final String autoSaveText;
  final bool saving;
  final bool submitting;
  final VoidCallback onBack;
  final VoidCallback onSave;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _ErxinColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _erxinShadow(color: const Color(0x12172033), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1260;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 282 : 306,
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  alignment: Alignment.centerLeft,
                  child: const Text(
                    '儿心量表-II 测评工作台',
                    maxLines: 1,
                    softWrap: false,
                    style: TextStyle(
                      color: _ErxinColors.ink,
                      fontSize: 22,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    SizedBox(
                      width: compact ? 104 : 118,
                      child: _HeaderMeta(
                        label: '儿童',
                        value: args.studentName.trim().isEmpty
                            ? '-'
                            : args.studentName.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '出生日期',
                        value: args.birthDate.trim().isEmpty
                            ? '-'
                            : args.birthDate.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '测查日期',
                        value: args.assessmentDate.trim().isEmpty
                            ? '-'
                            : args.assessmentDate.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '实足年龄',
                        value: args.studentAge.trim().isEmpty
                            ? '-'
                            : args.studentAge.trim(),
                      ),
                    ),
                  ],
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                SizedBox(
                  width: compact ? 86 : 106,
                  child: Text(
                    autoSaveText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: const TextStyle(
                      color: _ErxinColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
              const SizedBox(width: 10),
              _TopActionButton(
                label: '保存草稿',
                icon: Icons.save_outlined,
                loading: saving,
                filled: false,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _TopActionButton(
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

class _HeaderMeta extends StatelessWidget {
  const _HeaderMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(left: 8),
      padding: const EdgeInsets.only(left: 8),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _ErxinColors.line)),
      ),
      child: FittedBox(
        fit: BoxFit.scaleDown,
        alignment: Alignment.centerLeft,
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
            color: _ErxinColors.body,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
        ),
      ),
    );
  }
}

class _HeaderIconButton extends StatelessWidget {
  const _HeaderIconButton({required this.icon, required this.onTap});

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
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Icon(
            icon,
            color: _ErxinColors.body,
            size: 34,
          ),
        ),
      ),
    );
  }
}

class _TopActionButton extends StatelessWidget {
  const _TopActionButton({
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
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: filled ? _ErxinColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _ErxinColors.orange),
            boxShadow: filled
                ? _erxinShadow(
                    color: const Color(0x28E96F43),
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
                    color: filled ? Colors.white : _ErxinColors.orange,
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : _ErxinColors.orange,
                ),
              const SizedBox(width: 7),
              Text(
                label,
                maxLines: 1,
                softWrap: false,
                style: TextStyle(
                  color: filled ? Colors.white : _ErxinColors.orangeDeep,
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

class _DomainSidebar extends StatelessWidget {
  const _DomainSidebar({
    required this.domains,
    required this.selectedCode,
    required this.progressForDomain,
    required this.completedDomainCount,
    required this.savedItemCount,
    required this.onSelect,
  });

  final List<ErxinDomain> domains;
  final String selectedCode;
  final _DomainProgress Function(String domainCode) progressForDomain;
  final int completedDomainCount;
  final int savedItemCount;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    final _DomainProgress selectedProgress = progressForDomain(selectedCode);
    return Container(
      width: 214,
      padding: const EdgeInsets.fromLTRB(
        14,
        16,
        12,
        _erxinSidebarBottomPadding,
      ),
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '能区进度',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 12),
          for (final ErxinDomain domain in domains)
            _DomainRow(
              domain: domain,
              selected: domain.domainCode == selectedCode,
              progress: progressForDomain(domain.domainCode),
              onTap: () => onSelect(domain.domainCode),
            ),
          const SizedBox(height: 2),
          const _AllItemsButton(),
          const Spacer(),
          _ProgressSummary(
            domainStatus: selectedProgress.answered >= selectedProgress.total &&
                    selectedProgress.total > 0
                ? '本能区：当前可见完成'
                : '本能区：测查中',
            scaleStatus: savedItemCount > 0
                ? '$completedDomainCount/5 能区完成\n已保存$savedItemCount题'
                : '$completedDomainCount/5 能区完成',
          ),
          const SizedBox(height: _erxinProgressSummaryBottomGap),
        ],
      ),
    );
  }
}

class _DomainRow extends StatelessWidget {
  const _DomainRow({
    required this.domain,
    required this.selected,
    required this.progress,
    required this.onTap,
  });

  final ErxinDomain domain;
  final bool selected;
  final _DomainProgress progress;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool complete =
        progress.total > 0 && progress.answered >= progress.total;
    final double percent =
        progress.total <= 0 ? 0 : progress.answered / progress.total;
    final String status = complete
        ? '已完成'
        : progress.answered > 0
            ? '测查中'
            : '待测';
    final Color statusColor = complete
        ? _ErxinColors.green
        : selected
            ? _ErxinColors.blue
            : _ErxinColors.muted;
    return Padding(
      padding: const EdgeInsets.only(bottom: 9),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          height: 66,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFEEE5) : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Column(
            children: <Widget>[
              Row(
                children: <Widget>[
                  _DomainIcon(
                    icon: _domainIconFor(domain),
                    selected: selected,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      domain.domainName,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected ? _ErxinColors.blue : _ErxinColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  Text(
                    '${progress.answered}/${progress.total}',
                    style: const TextStyle(
                      color: _ErxinColors.body,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
              const Spacer(),
              Row(
                children: <Widget>[
                  Expanded(
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(2),
                      child: LinearProgressIndicator(
                        value: percent.clamp(0, 1),
                        minHeight: 4,
                        backgroundColor: const Color(0xFFF2E6DC),
                        valueColor: AlwaysStoppedAnimation<Color>(
                          complete ? _ErxinColors.green : _ErxinColors.blue,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 9),
                  Text(
                    status,
                    style: TextStyle(
                      color: statusColor,
                      fontSize: 12,
                      height: 1,
                      fontWeight: FontWeight.w800,
                    ),
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

class _DomainIcon extends StatelessWidget {
  const _DomainIcon({required this.icon, required this.selected});

  final IconData icon;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 24,
      height: 24,
      decoration: BoxDecoration(
        color: selected ? _ErxinColors.blue : const Color(0xFFFFF2EA),
        borderRadius: BorderRadius.circular(7),
      ),
      alignment: Alignment.center,
      child: Icon(
        icon,
        size: 15,
        color: selected ? Colors.white : _ErxinColors.body,
      ),
    );
  }
}

IconData _domainIconFor(ErxinDomain domain) {
  final String code = domain.domainCode.toUpperCase();
  final String name = domain.domainName;
  if (code == 'GM' || name.contains('大运动')) {
    return Icons.directions_run_rounded;
  }
  if (code == 'FM' || name.contains('精细')) {
    return Icons.gesture_rounded;
  }
  if (code == 'AD' || name.contains('适应')) {
    return Icons.psychology_alt_rounded;
  }
  if (code == 'LANG' || name.contains('语言')) {
    return Icons.record_voice_over_rounded;
  }
  if (code == 'SOC' || name.contains('社会') || name.contains('社交')) {
    return Icons.groups_2_rounded;
  }
  return Icons.extension_rounded;
}

class _AllItemsButton extends StatelessWidget {
  const _AllItemsButton();

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: () {},
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF5),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Row(
            children: const <Widget>[
              Icon(
                Icons.list_alt_rounded,
                size: 17,
                color: _ErxinColors.blue,
              ),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  '查看全部题目',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _ErxinColors.blue,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Icon(
                Icons.chevron_right_rounded,
                size: 18,
                color: _ErxinColors.blue,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AgeMonthSection extends StatelessWidget {
  const _AgeMonthSection({
    required this.month,
    required this.isMainAge,
    required this.items,
    required this.itemPasses,
    required this.selectedItemNo,
    required this.itemKeyFor,
    required this.onSelectItem,
    required this.onScore,
  });

  final int month;
  final bool isMainAge;
  final List<ErxinItemSummary> items;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;
  final GlobalKey Function(int itemNo) itemKeyFor;
  final ValueChanged<int> onSelectItem;
  final void Function(int itemNo, bool passed) onScore;

  @override
  Widget build(BuildContext context) {
    final List<ErxinItemSummary> displayItems = items;
    final int answered = items
        .where((ErxinItemSummary item) => itemPasses.containsKey(item.itemNo))
        .length;
    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        border: Border.all(color: _ErxinColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Container(
            height: 34,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color:
                  isMainAge ? const Color(0xFFFFF1E8) : const Color(0xFFFFFAF5),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(8),
              ),
              border: const Border(
                bottom: BorderSide(color: _ErxinColors.line),
              ),
            ),
            child: Row(
              children: <Widget>[
                Text(
                  '$month月龄',
                  style: TextStyle(
                    color: isMainAge ? _ErxinColors.blue : _ErxinColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                if (isMainAge) ...<Widget>[
                  const SizedBox(width: 8),
                  Container(
                    height: 22,
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(color: const Color(0xFFFFC8AD)),
                    ),
                    child: const Center(
                      child: Text(
                        '主测月龄',
                        style: TextStyle(
                          color: _ErxinColors.blue,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ),
                ],
                const Spacer(),
                Text(
                  '已测 $answered/${items.length}',
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          for (final MapEntry<int, ErxinItemSummary> entry
              in displayItems.asMap().entries)
            _ItemScoreRow(
              key: itemKeyFor(entry.value.itemNo),
              item: entry.value,
              selected: entry.value.itemNo == selectedItemNo,
              passed: itemPasses[entry.value.itemNo],
              showBottomDivider: entry.key < displayItems.length - 1,
              onTap: () => onSelectItem(entry.value.itemNo),
              onScore: (bool passed) => onScore(entry.value.itemNo, passed),
            ),
        ],
      ),
    );
  }
}

class _ItemScoreRow extends StatelessWidget {
  const _ItemScoreRow({
    required this.item,
    required this.selected,
    required this.passed,
    required this.showBottomDivider,
    required this.onTap,
    required this.onScore,
    super.key,
  });

  final ErxinItemSummary item;
  final bool selected;
  final bool? passed;
  final bool showBottomDivider;
  final VoidCallback onTap;
  final ValueChanged<bool> onScore;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 48,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFFBEB) : Colors.white,
          border: showBottomDivider
              ? const Border(bottom: BorderSide(color: _ErxinColors.line))
              : null,
        ),
        child: Row(
          children: <Widget>[
            SizedBox(
              width: 42,
              child: Text(
                '${item.itemNo}',
                style: const TextStyle(
                  color: _ErxinColors.body,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            Expanded(
              child: Row(
                children: <Widget>[
                  Flexible(
                    child: Text(
                      item.itemTitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  if (item.parentReportAllowed) ...<Widget>[
                    const SizedBox(width: 8),
                    const _MiniMarker(text: 'R'),
                  ],
                  if (item.attentionIfFailed) ...<Widget>[
                    const SizedBox(width: 6),
                    const _MiniMarker(text: '*', warning: true),
                  ],
                ],
              ),
            ),
            _ScoreButton(
              label: '通过',
              selected: passed == true,
              color: _ErxinColors.green,
              icon: Icons.check_circle_rounded,
              onTap: () => onScore(true),
            ),
            const SizedBox(width: 8),
            _ScoreButton(
              label: '不通过',
              selected: passed == false,
              color: _ErxinColors.red,
              icon: Icons.cancel_rounded,
              onTap: () => onScore(false),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScoreButton extends StatelessWidget {
  const _ScoreButton({
    required this.label,
    required this.selected,
    required this.color,
    required this.icon,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final Color color;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color selectedFill = Color.alphaBlend(
      color.withOpacity(.12),
      Colors.white,
    );
    final Color selectedBorder = Color.alphaBlend(
      color.withOpacity(.48),
      Colors.white,
    );
    final Color contentColor = selected ? color : _ErxinColors.body;
    return SizedBox(
      width: label.length > 2 ? 104 : 88,
      height: 34,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(8),
          child: Ink(
            decoration: BoxDecoration(
              color: selected ? selectedFill : Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: selected ? selectedBorder : _ErxinColors.line,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: 17, color: contentColor),
                const SizedBox(width: 6),
                Text(
                  label,
                  maxLines: 1,
                  softWrap: false,
                  overflow: TextOverflow.visible,
                  style: TextStyle(
                    color: contentColor,
                    fontSize: 13,
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

class _DetailTextBox extends StatelessWidget {
  const _DetailTextBox({required this.title, required this.text});

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    final String normalizedText = _inlineDetailText(text);
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 8, 10, 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: _ErxinColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 4),
          Expanded(
            child: Text(
              normalizedText.isEmpty ? '暂无内容' : normalizedText,
              maxLines: 4,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 13,
                height: 1.28,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

String _inlineDetailText(String value) {
  return value
      .trim()
      .replaceAll(RegExp(r'[\r\n]+\s*'), '')
      .replaceAll(RegExp(r'\s+'), ' ')
      .replaceAllMapped(
        RegExp(r'([，。；、：！？])\s+'),
        (Match match) => match.group(1) ?? '',
      )
      .trim();
}

class _RightRemarkSection extends StatelessWidget {
  const _RightRemarkSection({
    required this.height,
    required this.itemNo,
    required this.remark,
    required this.onChanged,
    required this.onEditingComplete,
  });

  final double height;
  final int itemNo;
  final String remark;
  final void Function(int itemNo, String remark) onChanged;
  final ValueChanged<int> onEditingComplete;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: Padding(
        padding: const EdgeInsets.only(top: _erxinDetailPanelTopPadding),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const SizedBox(
              height: _erxinDetailHeaderHeight,
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  '题目备注',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
            const SizedBox(height: _erxinDetailHeaderGap),
            SizedBox(
              height: _erxinDetailContentHeight,
              child: _RemarkBox(
                itemNo: itemNo,
                remark: remark,
                onChanged: onChanged,
                onEditingComplete: onEditingComplete,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RemarkBox extends StatelessWidget {
  const _RemarkBox({
    required this.itemNo,
    required this.remark,
    required this.onChanged,
    required this.onEditingComplete,
  });

  final int itemNo;
  final String remark;
  final void Function(int itemNo, String remark) onChanged;
  final ValueChanged<int> onEditingComplete;

  @override
  Widget build(BuildContext context) {
    final bool enabled = itemNo > 0;
    final String preview = remark.trim();
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? () => _openRemarkEditor(context) : null,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          padding: const EdgeInsets.all(9),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF5),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Align(
            alignment: Alignment.topLeft,
            child: Text(
              enabled
                  ? preview.isEmpty
                      ? '添加本题备注'
                      : preview
                  : '选择题目后添加备注',
              maxLines: 4,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: preview.isEmpty ? _ErxinColors.muted : _ErxinColors.body,
                fontSize: 13,
                height: 1.25,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _openRemarkEditor(BuildContext context) async {
    bool changed = false;
    await showDialog<void>(
      context: context,
      barrierColor: Colors.black.withOpacity(.18),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          alignment: Alignment.topCenter,
          child: _RemarkEditorDialog(
            initialValue: remark,
            onChanged: (String value) {
              changed = true;
              onChanged(itemNo, value);
            },
            onClear: () {
              changed = true;
              onChanged(itemNo, '');
            },
          ),
        );
      },
    );
    if (changed) {
      onEditingComplete(itemNo);
    }
  }
}

class _RemarkEditorDialog extends StatefulWidget {
  const _RemarkEditorDialog({
    required this.initialValue,
    required this.onChanged,
    required this.onClear,
  });

  final String initialValue;
  final ValueChanged<String> onChanged;
  final VoidCallback onClear;

  @override
  State<_RemarkEditorDialog> createState() => _RemarkEditorDialogState();
}

class _RemarkEditorDialogState extends State<_RemarkEditorDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialValue);
    _controller.selection = TextSelection.collapsed(
      offset: _controller.text.length,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final double keyboardBottom = MediaQuery.viewInsetsOf(context).bottom;
    return AnimatedPadding(
      duration: const Duration(milliseconds: 180),
      curve: Curves.easeOutCubic,
      padding: EdgeInsets.fromLTRB(24, 108, 24, keyboardBottom + 24),
      child: Align(
        alignment: Alignment.topCenter,
        child: Material(
          color: Colors.white,
          borderRadius: BorderRadius.circular(10),
          clipBehavior: Clip.antiAlias,
          child: Container(
            width: 430,
            height: 258,
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
            decoration: BoxDecoration(
              border: Border.all(color: _ErxinColors.line),
              borderRadius: BorderRadius.circular(10),
              boxShadow: _erxinShadow(
                color: const Color(0x22000000),
                blur: 22,
                offset: const Offset(0, 10),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    const Expanded(
                      child: Text(
                        '题目备注',
                        style: TextStyle(
                          color: _ErxinColors.ink,
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close_rounded),
                      iconSize: 20,
                      visualDensity: VisualDensity.compact,
                      tooltip: '关闭',
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: TextField(
                    controller: _controller,
                    autofocus: true,
                    maxLines: null,
                    expands: true,
                    textAlignVertical: TextAlignVertical.top,
                    onChanged: widget.onChanged,
                    onTapOutside: (_) =>
                        FocusManager.instance.primaryFocus?.unfocus(),
                    style: const TextStyle(fontSize: 14, height: 1.35),
                    decoration: InputDecoration(
                      hintText: '添加本题备注',
                      filled: true,
                      fillColor: const Color(0xFFFFFAF5),
                      contentPadding: const EdgeInsets.all(10),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(color: _ErxinColors.line),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(color: _ErxinColors.line),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                Row(
                  children: <Widget>[
                    TextButton(
                      onPressed: () {
                        _controller.clear();
                        widget.onClear();
                      },
                      child: const Text('清空'),
                    ),
                    const Spacer(),
                    FilledButton(
                      onPressed: () => Navigator.of(context).pop(),
                      style: FilledButton.styleFrom(
                        backgroundColor: _ErxinColors.blue,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                      child: const Text('完成'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _RuleCard extends StatelessWidget {
  const _RuleCard({
    required this.title,
    required this.body,
    required this.icon,
    required this.color,
  });

  final String title;
  final String body;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: color, size: 22),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  body,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 12,
                    height: 1.3,
                    fontWeight: FontWeight.w700,
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

class _RuleChecklist extends StatefulWidget {
  const _RuleChecklist({
    required this.rows,
    required this.revealMonths,
    required this.revealSerial,
    required this.onTapMonth,
  });

  final List<_RuleRow> rows;
  final List<int> revealMonths;
  final int revealSerial;
  final ValueChanged<int> onTapMonth;

  @override
  State<_RuleChecklist> createState() => _RuleChecklistState();
}

class _RuleChecklistState extends State<_RuleChecklist> {
  static const double _rowExtent = 41;

  final ScrollController _scrollController = ScrollController();
  final Set<int> _flashingMonths = <int>{};
  Timer? _flashTimer;

  @override
  void dispose() {
    _flashTimer?.cancel();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant _RuleChecklist oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.revealSerial == widget.revealSerial ||
        widget.revealMonths.isEmpty) {
      return;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      _revealMonths(widget.revealMonths);
    });
  }

  void _revealTargetMonths(_RuleRow row) {
    _revealMonths(row.targetMonths);
  }

  void _revealMonths(List<int> months) {
    if (months.isEmpty) {
      return;
    }
    final Set<int> targetMonths = months.toSet();
    final List<_RuleRow> monthRows = widget.rows
        .where((_RuleRow candidate) => candidate.month != null)
        .toList(growable: false);
    final int targetIndex = monthRows.indexWhere(
      (_RuleRow candidate) => targetMonths.contains(candidate.month),
    );
    if (targetIndex < 0) {
      return;
    }

    _flashTimer?.cancel();
    setState(() {
      _flashingMonths
        ..clear()
        ..addAll(targetMonths);
    });
    _flashTimer = Timer(const Duration(milliseconds: 900), () {
      if (!mounted) {
        return;
      }
      setState(_flashingMonths.clear);
    });

    if (!_scrollController.hasClients) {
      return;
    }
    final double maxOffset = _scrollController.position.maxScrollExtent;
    final double targetOffset = math.min(
      maxOffset,
      math.max(0, 6 + targetIndex * _rowExtent),
    );
    unawaited(
      _scrollController.animateTo(
        targetOffset,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOutCubic,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final List<_RuleRow> monthRows = widget.rows
        .where((_RuleRow row) => row.month != null)
        .toList(growable: false);
    final List<_RuleRow> pinnedRows = widget.rows
        .where((_RuleRow row) => row.month == null)
        .toList(growable: false);
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        children: <Widget>[
          Expanded(
            child: ListView.separated(
              controller: _scrollController,
              padding: const EdgeInsets.symmetric(vertical: 6),
              physics: const BouncingScrollPhysics(),
              itemCount: monthRows.length,
              separatorBuilder: (_, __) => const Divider(
                height: 1,
                thickness: 1,
                color: _ErxinColors.line,
              ),
              itemBuilder: (BuildContext context, int index) {
                final _RuleRow row = monthRows[index];
                return _RuleChecklistRow(
                  row: row,
                  highlighted:
                      row.month != null && _flashingMonths.contains(row.month),
                  onTapMonth: widget.onTapMonth,
                );
              },
            ),
          ),
          if (pinnedRows.isNotEmpty) ...<Widget>[
            const Divider(height: 1, thickness: 1, color: _ErxinColors.line),
            for (final MapEntry<int, _RuleRow> entry
                in pinnedRows.asMap().entries) ...<Widget>[
              _RuleChecklistRow(
                row: entry.value,
                onTapMonth: widget.onTapMonth,
                onTapTargets: () => _revealTargetMonths(entry.value),
              ),
              if (entry.key < pinnedRows.length - 1)
                const Divider(
                  height: 1,
                  thickness: 1,
                  color: _ErxinColors.line,
                ),
            ],
          ],
        ],
      ),
    );
  }
}

class _RuleChecklistRow extends StatelessWidget {
  const _RuleChecklistRow({
    required this.row,
    required this.onTapMonth,
    this.highlighted = false,
    this.onTapTargets,
  });

  final _RuleRow row;
  final ValueChanged<int> onTapMonth;
  final bool highlighted;
  final VoidCallback? onTapTargets;

  @override
  Widget build(BuildContext context) {
    final bool targetClickable =
        row.month == null && row.done && row.targetMonths.isNotEmpty;
    final bool clickable = row.month != null || targetClickable;
    final bool unmetResult = _ruleRowHasUnmetResult(row);
    return Material(
      color: highlighted
          ? const Color(0xFFFFF3BF)
          : row.selected
              ? const Color(0xFFFFF1E8)
              : Colors.white,
      child: InkWell(
        onTap: row.month != null
            ? () => onTapMonth(row.month!)
            : targetClickable
                ? onTapTargets
                : null,
        child: SizedBox(
          height: 40,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: <Widget>[
                Icon(
                  row.done
                      ? Icons.check_circle
                      : unmetResult
                          ? Icons.cancel_rounded
                          : Icons.radio_button_unchecked,
                  color: row.done
                      ? _ErxinColors.green
                      : unmetResult
                          ? _ErxinColors.red
                          : _ErxinColors.muted,
                  size: 18,
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: Text(
                    row.label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color:
                          row.selected ? _ErxinColors.blue : _ErxinColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  row.value,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: row.done
                        ? _ErxinColors.green
                        : unmetResult
                            ? _ErxinColors.red
                            : _ErxinColors.body,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                if (clickable) ...<Widget>[
                  const SizedBox(width: 5),
                  Icon(
                    targetClickable
                        ? Icons.center_focus_strong_rounded
                        : row.selected
                            ? Icons.edit_note_rounded
                            : Icons.history_rounded,
                    color: targetClickable || row.selected
                        ? _ErxinColors.blue
                        : _ErxinColors.muted,
                    size: 16,
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

bool _ruleRowHasUnmetResult(_RuleRow row) {
  if (row.done || row.month == null) {
    return false;
  }
  return row.value.contains('未全') || row.value.contains('未通过');
}

class _CurrentItemsEmptyState extends StatelessWidget {
  const _CurrentItemsEmptyState({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 420,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFAF5),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: _ErxinColors.line),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(
              Icons.fact_check_outlined,
              color: _ErxinColors.blue,
              size: 34,
            ),
            const SizedBox(height: 10),
            const Text(
              '当前题目已完成',
              style: TextStyle(
                color: _ErxinColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              text,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 13,
                height: 1.35,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _HistoryChangeConfirmDialog extends StatelessWidget {
  const _HistoryChangeConfirmDialog({
    required this.itemTitle,
    required this.nextStatus,
    required this.onCancel,
    required this.onConfirm,
  });

  final String itemTitle;
  final String nextStatus;
  final VoidCallback onCancel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Container(
        width: 438,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: _ErxinColors.line),
          boxShadow: _erxinShadow(
            color: const Color(0x24172033),
            blur: 28,
            offset: const Offset(0, 14),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Text(
              '确认修改历史记录',
              style: TextStyle(
                color: _ErxinColors.ink,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            Text(
              '$itemTitle 将改为“$nextStatus”。修改后会重新计算当前能区的前测基线、后测封顶和后续需测月龄。',
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 14,
                height: 1.45,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 18),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                SizedBox(
                  height: 36,
                  child: OutlinedButton(
                    onPressed: onCancel,
                    child: const Text('取消'),
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  height: 36,
                  child: FilledButton(
                    onPressed: onConfirm,
                    style: FilledButton.styleFrom(
                      backgroundColor: _ErxinColors.orange,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    child: const Text(
                      '确认修改',
                      style: TextStyle(fontWeight: FontWeight.w900),
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

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.domainStatus,
    required this.scaleStatus,
  });

  final String domainStatus;
  final String scaleStatus;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: _erxinProgressSummaryHeight,
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '完成情况',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            domainStatus,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: _summaryStyle,
          ),
          const SizedBox(height: 4),
          Text(
            '全量表：$scaleStatus',
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: _summaryStyle,
          ),
        ],
      ),
    );
  }

  static const TextStyle _summaryStyle = TextStyle(
    color: _ErxinColors.body,
    fontSize: 13,
    fontWeight: FontWeight.w700,
  );
}

class _SmallBadge extends StatelessWidget {
  const _SmallBadge({required this.text, this.strong = false});

  final String text;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: strong ? const Color(0xFFFFF1E8) : const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: strong ? _ErxinColors.blue : _ErxinColors.line,
        ),
      ),
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(
          color: strong ? _ErxinColors.blue : _ErxinColors.body,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _MiniMarker extends StatelessWidget {
  const _MiniMarker({required this.text, this.warning = false});

  final String text;
  final bool warning;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 20,
      width: 20,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: warning ? const Color(0xFFFFF2E8) : const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: warning ? const Color(0xFFEA580C) : _ErxinColors.blue,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _DialogTextBlock extends StatelessWidget {
  const _DialogTextBlock({required this.title, required this.text});

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w900),
        ),
        const SizedBox(height: 8),
        Text(
          text.trim().isEmpty ? '暂无内容' : text.trim(),
          style: const TextStyle(fontSize: 14, height: 1.55),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onBack});

  final String message;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(message, style: const TextStyle(color: _ErxinColors.red)),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onBack, child: const Text('返回')),
        ],
      ),
    );
  }
}

class _RuleRow {
  const _RuleRow({
    required this.label,
    required this.value,
    required this.done,
    this.month,
    this.targetMonths = const <int>[],
    this.selected = false,
  });

  final String label;
  final String value;
  final bool done;
  final int? month;
  final List<int> targetMonths;
  final bool selected;
}

class _DomainProgress {
  const _DomainProgress({required this.answered, required this.total});

  final int answered;
  final int total;
}

class _ErxinColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color blue = Color(0xFFE96F43);
  static const Color green = Color(0xFF6F9F70);
  static const Color red = Color(0xFFD94A42);
}

BoxDecoration _erxinPanelDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.9),
    borderRadius: BorderRadius.circular(10),
    border: Border.all(color: _ErxinColors.line),
    boxShadow: _erxinShadow(
      color: const Color(0x12B05F32),
      blur: 15,
      offset: const Offset(0, 7),
    ),
  );
}

List<BoxShadow> _erxinShadow({
  Color color = const Color(0x16000000),
  double blur = 16,
  Offset offset = const Offset(0, 8),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

double _actualAgeMonths(String birthDate, String assessmentDate) {
  final DateTime? birth = DateTime.tryParse(birthDate);
  final DateTime? target = DateTime.tryParse(assessmentDate);
  if (birth == null || target == null || birth.isAfter(target)) {
    return 0;
  }
  final int days = target.difference(birth).inDays;
  return days / 30.0;
}

String _dateOnlyText(String value) {
  final String text = value.trim();
  if (text.isEmpty) {
    return '';
  }
  final RegExpMatch? match =
      RegExp(r'^(\d{4})[-/](\d{1,2})[-/](\d{1,2})').firstMatch(text);
  if (match != null) {
    final String year = match.group(1)!;
    final String month = match.group(2)!.padLeft(2, '0');
    final String day = match.group(3)!.padLeft(2, '0');
    return '$year-$month-$day';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    return text;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')}';
}

String _formatErxinDateTime(String value) {
  final String text = value.trim();
  if (text.isEmpty) {
    return '-';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    if (text.length >= 16) {
      return text.substring(0, 16).replaceFirst('T', ' ');
    }
    return text;
  }
  final DateTime local = parsed.toLocal();
  return '${local.year.toString().padLeft(4, '0')}-'
      '${local.month.toString().padLeft(2, '0')}-'
      '${local.day.toString().padLeft(2, '0')} '
      '${local.hour.toString().padLeft(2, '0')}:'
      '${local.minute.toString().padLeft(2, '0')}';
}
