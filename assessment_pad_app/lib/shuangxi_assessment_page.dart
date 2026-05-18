import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_draft_resume_dialog.dart';
import 'assessment_age_formatter.dart';
import 'assessment_scale_client.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';
import 'shuangxi_assessment_client.dart';

class ShuangxiAssessmentLaunchArgs {
  const ShuangxiAssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = '双溪课程评量表A',
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

class ShuangxiAssessmentPage extends StatefulWidget {
  const ShuangxiAssessmentPage({
    required this.onBack,
    this.args = const ShuangxiAssessmentLaunchArgs(),
    this.client = const ApiShuangxiAssessmentClient(),
    super.key,
  });

  final VoidCallback onBack;
  final ShuangxiAssessmentLaunchArgs args;
  final ShuangxiAssessmentClient client;

  @override
  State<ShuangxiAssessmentPage> createState() => _ShuangxiAssessmentPageState();
}

class _ShuangxiAssessmentPageState extends State<ShuangxiAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';

  ShuangxiTemplateSummary _template = ShuangxiTemplateSummary.empty;
  ShuangxiAssessmentItem _currentItem = ShuangxiAssessmentItem.empty;
  final Map<int, int> _itemScores = <int, int>{};
  final TextEditingController _remarkController = TextEditingController();
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  Future<ShuangxiDraftDetail?>? _saveDraftFuture;
  Future<void>? _currentItemSaveFuture;
  AssessmentDraftSummary? _detectedDraft;
  Future<ShuangxiDraftDetail>? _detectedDraftDetailRequest;
  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _errorMessage = '';
  String _autoSaveText = '等待作答';
  int _selectedItemNo = 0;
  int _studentId = 0;
  int _draftId = 0;
  int _detectedDraftDetailDraftId = 0;
  int _draftDetectionSerial = 0;
  Timer? _autoAdvanceTimer;
  Timer? _remarkAutoSaveTimer;
  bool _loading = true;
  bool _itemLoading = false;
  bool _autoNext = true;
  bool _draftDialogShown = false;
  bool _savingDraft = false;
  bool _submitting = false;
  bool _saveDraftFutureSilent = false;
  bool _saveDraftJoinedByManual = false;

  @override
  void initState() {
    super.initState();
    _draftId = widget.args.draftId;
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isNotEmpty
        ? _dateOnlyText(widget.args.assessmentDate)
        : _todayIsoDate();
    _examinerName = widget.args.examinerName;
    unawaited(_initialize());
  }

  @override
  void dispose() {
    _autoAdvanceTimer?.cancel();
    _remarkAutoSaveTimer?.cancel();
    _remarkController.dispose();
    _messageController.dispose();
    super.dispose();
  }

  Future<void> _initialize() async {
    final int draftDetectionSerial = _draftDetectionSerial + 1;
    setState(() {
      _loading = true;
      _errorMessage = '';
      _detectedDraft = null;
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _draftDialogShown = false;
      _draftDetectionSerial = draftDetectionSerial;
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
      final Future<AssessmentDraftSummary?>? detectedDraftRequest =
          _draftId > 0 || _studentId <= 0
              ? null
              : _findLatestDraft(token).then<AssessmentDraftSummary?>(
                  (AssessmentDraftSummary? draft) => draft,
                  onError: (Object _, StackTrace __) => null,
                );
      final ShuangxiTemplateSummary template;
      ShuangxiDraftDetail? launchDraftDetail;
      if (_draftId > 0) {
        final List<Object> result = await Future.wait<Object>(
          <Future<Object>>[
            widget.client.fetchTemplateSummary(token),
            widget.client.fetchDraftDetail(token, _draftId),
          ],
        );
        template = result[0] as ShuangxiTemplateSummary;
        launchDraftDetail = result[1] as ShuangxiDraftDetail;
      } else {
        template = await widget.client.fetchTemplateSummary(token);
      }
      if (!mounted) {
        return;
      }
      if (draftDetectionSerial != _draftDetectionSerial) {
        return;
      }
      setState(() {
        _token = token;
        _template = template;
        if (launchDraftDetail != null) {
          _applyDraftDetail(launchDraftDetail);
        }
        if (_selectedItemNo <= 0) {
          _selectedItemNo = _firstMissingItemNo() > 0
              ? _firstMissingItemNo()
              : (template.allItems.isNotEmpty
                  ? template.allItems.first.itemNo
                  : 0);
        }
      });
      await _loadSelectedItem(_selectedItemNo);
      if (!mounted || draftDetectionSerial != _draftDetectionSerial) {
        return;
      }
      setState(() {
        _loading = false;
        _autoSaveText = _draftId > 0 ? '已载入草稿' : '已准备';
      });
      if (detectedDraftRequest != null) {
        unawaited(
          _completeDetectedDraftLookup(
            token,
            detectedDraftRequest,
            draftDetectionSerial,
          ),
        );
      }
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
        _errorMessage = '双溪课程评量表A加载失败：$error';
      });
    }
  }

  Future<void> _completeDetectedDraftLookup(
    String token,
    Future<AssessmentDraftSummary?> request,
    int serial,
  ) async {
    final AssessmentDraftSummary? draft = await request;
    if (!mounted || serial != _draftDetectionSerial || _draftId > 0) {
      return;
    }
    setState(() {
      _detectedDraft = draft;
    });
    _prefetchDetectedDraftDetail(token, draft);
    _showDetectedDraftDialogIfNeeded();
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
    final Future<ShuangxiDraftDetail> request =
        widget.client.fetchDraftDetail(token, draft.id);
    _detectedDraftDetailDraftId = draft.id;
    _detectedDraftDetailRequest = request;
    unawaited(
      request.then<void>(
        (ShuangxiDraftDetail _) {},
        onError: (Object _, StackTrace __) {},
      ),
    );
  }

  Future<ShuangxiDraftDetail> _resolveDetectedDraftDetail(
    AssessmentDraftSummary draft,
  ) async {
    final Future<ShuangxiDraftDetail>? prefetched = _detectedDraftDetailRequest;
    if (_detectedDraftDetailDraftId == draft.id && prefetched != null) {
      try {
        return await prefetched;
      } on Object {
        // Retry below.
      }
    }
    return widget.client.fetchDraftDetail(_token, draft.id);
  }

  void _showDetectedDraftDialogIfNeeded() {
    final AssessmentDraftSummary? draft = _detectedDraft;
    if (!mounted || _draftDialogShown || draft == null || draft.id <= 0) {
      return;
    }
    final int answered = draft.answeredItemCount;
    final int total = math.max(
      draft.progressItemCount > 0 ? draft.progressItemCount : _totalCount,
      answered,
    );
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
              child: AssessmentDraftResumeDialog(
                message: '当前儿童存在一份未提交的双溪课程评量表A草稿。',
                metaRows: <AssessmentDraftResumeMetaRow>[
                  AssessmentDraftResumeMetaRow(
                    label: '已完成',
                    value: '$answered / $total 题',
                  ),
                  AssessmentDraftResumeMetaRow(
                    label: '更新时间',
                    value: _formatDateTime(draft.updatedTime),
                  ),
                ],
                accentColor: _ShuangxiColors.orange,
                inkColor: _ShuangxiColors.ink,
                bodyColor: _ShuangxiColors.body,
                lineColor: _ShuangxiColors.line,
                lineSoftColor: _ShuangxiColors.lineSoft,
                onRestart: _restartWithoutDetectedDraft,
                onContinue: () => _continueDetectedDraft(draft),
              ),
            ),
          );
        },
      );
    });
  }

  Future<void> _restartWithoutDetectedDraft() async {
    if (!mounted) {
      return;
    }
    setState(() {
      _draftDetectionSerial += 1;
      _detectedDraft = null;
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _draftId = 0;
      _itemScores.clear();
      _remarkController.clear();
      _selectedItemNo =
          _template.allItems.isNotEmpty ? _template.allItems.first.itemNo : 0;
      _autoSaveText = '已重新开始';
      _draftDialogShown = false;
    });
    await _loadSelectedItem(_selectedItemNo);
  }

  Future<bool> _continueDetectedDraft(AssessmentDraftSummary draft) async {
    if (draft.id <= 0) {
      return false;
    }
    try {
      final ShuangxiDraftDetail detail =
          await _resolveDetectedDraftDetail(draft);
      if (!mounted) {
        return false;
      }
      setState(() {
        _applyDraftDetail(detail);
        _selectedItemNo = _firstMissingItemNo() > 0
            ? _firstMissingItemNo()
            : (_template.allItems.isNotEmpty
                ? _template.allItems.first.itemNo
                : 0);
        _detectedDraft = null;
        _draftId = detail.id;
        _autoSaveText = '已恢复最新草稿';
        _draftDialogShown = false;
      });
      await _loadSelectedItem(_selectedItemNo);
      return true;
    } on Object catch (error) {
      if (mounted) {
        _showMessage('恢复草稿失败：$error');
      }
      return false;
    }
  }

  void _applyDraftDetail(ShuangxiDraftDetail detail) {
    _draftId = detail.id;
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
    _setRemarkText(detail.input.remark);
    _itemScores
      ..clear()
      ..addAll(detail.input.itemScores);
    if (_selectedItemNo <= 0) {
      _selectedItemNo = _firstMissingItemNo() > 0
          ? _firstMissingItemNo()
          : (_template.allItems.isNotEmpty
              ? _template.allItems.first.itemNo
              : 0);
    }
  }

  Future<void> _loadSelectedItem(int itemNo) async {
    if (itemNo <= 0) {
      if (mounted) {
        setState(() {
          _currentItem = ShuangxiAssessmentItem.empty;
        });
      }
      return;
    }
    setState(() => _itemLoading = true);
    try {
      final ShuangxiAssessmentItem item =
          await widget.client.fetchTemplateItem(_token, itemNo: itemNo);
      if (!mounted || _selectedItemNo != itemNo) {
        return;
      }
      setState(() {
        _currentItem = item;
        _itemLoading = false;
      });
    } on Object catch (error) {
      if (!mounted || _selectedItemNo != itemNo) {
        return;
      }
      setState(() {
        _itemLoading = false;
        _autoSaveText = '题目加载失败';
      });
      debugPrint('Shuangxi item load failed: $error');
    }
  }

  Future<void> _selectItem(int itemNo) async {
    _cancelPendingAutoAdvance();
    if (itemNo <= 0 || itemNo == _selectedItemNo && !_itemLoading) {
      return;
    }
    setState(() => _selectedItemNo = itemNo);
    await _loadSelectedItem(itemNo);
  }

  void _selectDomain(String domainCode) {
    final ShuangxiDomainSummary? domain = _domainByCode(domainCode);
    final int firstItemNo = domain == null
        ? 0
        : domain.skills
            .expand((ShuangxiSkillSummary skill) => skill.items)
            .map((ShuangxiItemSummary item) => item.itemNo)
            .firstWhere((int itemNo) => itemNo > 0, orElse: () => 0);
    if (firstItemNo > 0) {
      unawaited(_selectItem(firstItemNo));
    }
  }

  void _selectScore(int score) {
    if (_selectedItemNo <= 0 || _submitting) {
      return;
    }
    final int currentItemNo = _selectedItemNo;
    _cancelPendingAutoAdvance();
    setState(() {
      _itemScores[currentItemNo] = score;
      _autoSaveText = '自动保存中...';
    });
    _queueItemSave(itemNo: currentItemNo, score: score);
    if (_autoNext) {
      _scheduleAutoAdvance(currentItemNo);
    }
  }

  void _queueItemSave({required int itemNo, required int score}) {
    final Future<void>? previousSave = _currentItemSaveFuture;
    late final Future<void> trackedSave;
    trackedSave = (() async {
      if (previousSave != null) {
        await previousSave;
      }
      final Future<ShuangxiDraftDetail?>? draftSave = _saveDraftFuture;
      if (draftSave != null) {
        await draftSave;
      }
      await _saveCurrentItem(itemNo: itemNo, score: score, silent: true);
    })()
        .whenComplete(() {
      if (identical(_currentItemSaveFuture, trackedSave)) {
        _currentItemSaveFuture = null;
      }
    });
    _currentItemSaveFuture = trackedSave;
    unawaited(trackedSave);
  }

  Future<void> _saveCurrentItem({
    required int itemNo,
    required int score,
    bool silent = false,
  }) async {
    if (itemNo <= 0 || !_itemScores.containsKey(itemNo)) {
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存');
      return;
    }
    try {
      final ShuangxiDraftDetail detail;
      if (_draftId <= 0) {
        final ShuangxiDraftDetail? created = await _saveDraft(silent: true);
        if (created == null || created.id <= 0) {
          return;
        }
        detail = created;
      } else {
        setState(() {
          _savingDraft = true;
          _autoSaveText = '自动保存中...';
        });
        detail = await widget.client.saveDraftItem(
          _token,
          <String, dynamic>{
            'draftId': _draftId,
            'itemNo': itemNo,
            'score': score,
          },
        );
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _mergeDraftDetailInput(detail);
        _savingDraft = false;
        _autoSaveText = '已自动保存';
      });
      if (!silent) {
        _showMessage('已保存本题', tone: PadMessageTone.success);
      }
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _savingDraft = false;
        _autoSaveText = '保存失败';
      });
      _showMessage('第$itemNo题自动保存失败：$error');
    }
  }

  Future<ShuangxiDraftDetail?> _saveDraft({bool silent = false}) async {
    final Future<ShuangxiDraftDetail?>? inFlight = _saveDraftFuture;
    if (inFlight != null) {
      if (!silent && _saveDraftFutureSilent && !_saveDraftJoinedByManual) {
        _saveDraftJoinedByManual = true;
        return _joinSilentDraftSave(inFlight);
      }
      return inFlight;
    }
    late final Future<ShuangxiDraftDetail?> trackedFuture;
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

  Future<ShuangxiDraftDetail?> _joinSilentDraftSave(
    Future<ShuangxiDraftDetail?> inFlight,
  ) async {
    if (mounted) {
      setState(() => _autoSaveText = '草稿保存中...');
    }
    final ShuangxiDraftDetail? detail = await inFlight;
    if (!mounted) {
      return detail;
    }
    if (detail == null) {
      _showMessage('保存草稿失败，请稍后重试');
      return null;
    }
    _showMessage('草稿已保存', tone: PadMessageTone.success);
    return detail;
  }

  Future<ShuangxiDraftDetail?> _performSaveDraft({
    required bool silent,
  }) async {
    if (_savingDraft) {
      return null;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿');
      return null;
    }
    if (_studentId <= 0) {
      _showMessage('缺少学员信息，无法保存草稿');
      return null;
    }
    setState(() {
      _savingDraft = true;
      _autoSaveText = '保存中...';
    });
    try {
      final ShuangxiDraftDetail detail =
          await widget.client.saveDraft(_token, _buildDraftPayload());
      if (!mounted) {
        return detail;
      }
      setState(() {
        _draftId = detail.id;
        _mergeDraftDetailInput(detail);
        _savingDraft = false;
        _autoSaveText = '草稿已保存';
      });
      if (!silent) {
        _showMessage('草稿已保存', tone: PadMessageTone.success);
      }
      return detail;
    } on Object catch (error) {
      if (!mounted) {
        return null;
      }
      setState(() {
        _savingDraft = false;
        _autoSaveText = '保存失败';
      });
      if (!silent) {
        _showMessage('保存草稿失败：$error');
      }
      return null;
    }
  }

  void _mergeDraftDetailInput(ShuangxiDraftDetail detail) {
    final Map<int, int> localScores = Map<int, int>.from(_itemScores);
    _itemScores
      ..clear()
      ..addAll(detail.input.itemScores)
      ..addAll(localScores);
  }

  void _setRemarkText(String remark) {
    if (_remarkController.text == remark) {
      return;
    }
    _remarkController.text = remark;
    _remarkController.selection = TextSelection.collapsed(
      offset: _remarkController.text.length,
    );
  }

  void _handleRemarkChanged(String value) {
    _remarkAutoSaveTimer?.cancel();
    if (mounted) {
      setState(() => _autoSaveText = '备注待保存');
    }
    if (_token.trim().isEmpty || _studentId <= 0 || _loading || _submitting) {
      return;
    }
    _remarkAutoSaveTimer = Timer(const Duration(milliseconds: 900), () {
      _remarkAutoSaveTimer = null;
      if (!mounted || _submitting) {
        return;
      }
      unawaited(_saveDraft(silent: true));
    });
  }

  void _finishRemarkEditing() {
    _remarkAutoSaveTimer?.cancel();
    _remarkAutoSaveTimer = null;
    FocusManager.instance.primaryFocus?.unfocus();
    if (_token.trim().isEmpty || _studentId <= 0 || _loading || _submitting) {
      return;
    }
    unawaited(_saveDraft(silent: true));
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再提交记录');
      return;
    }
    try {
      final Future<void>? itemSave = _currentItemSaveFuture;
      if (itemSave != null) {
        await itemSave;
      }
      final Future<ShuangxiDraftDetail?>? draftSave = _saveDraftFuture;
      if (draftSave != null) {
        await draftSave;
      }
      if (!mounted) {
        return;
      }
      final int missing = math.max(0, _totalCount - _answeredCount);
      if (missing > 0) {
        final bool confirmed = await _confirmSubmitWithZeroScores(missing);
        if (!mounted || !confirmed) {
          return;
        }
        setState(() {
          _fillMissingScoresWithZero();
        });
      }
      setState(() => _submitting = true);
      final ShuangxiDraftDetail? detail = await _saveDraft(silent: true);
      if (!mounted) {
        return;
      }
      if (detail == null) {
        setState(() => _submitting = false);
        _showMessage('保存草稿失败，请稍后重试');
        return;
      }
      final int draftId = detail.id > 0 ? detail.id : _draftId;
      if (draftId <= 0) {
        setState(() => _submitting = false);
        _showMessage('请先保存草稿，再提交正式记录');
        return;
      }
      await widget.client.submitDraft(_token, draftId);
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('正式记录已提交', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 650));
      if (mounted) {
        widget.onBack();
      }
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('提交记录失败：$error');
    }
  }

  Future<bool> _confirmSubmitWithZeroScores(int missingCount) async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      barrierColor: Colors.black.withOpacity(.32),
      builder: (BuildContext dialogContext) {
        return PopScope(
          canPop: false,
          child: PadDialogViewport(
            child: _SubmitZeroScoreDialog(
              missingCount: missingCount,
              totalCount: _totalCount,
              answeredCount: _answeredCount,
              onCancel: () => Navigator.of(dialogContext).pop(false),
              onConfirm: () => Navigator.of(dialogContext).pop(true),
            ),
          ),
        );
      },
    );
    return confirmed ?? false;
  }

  void _fillMissingScoresWithZero() {
    final Set<int> itemNos = <int>{};
    for (final ShuangxiItemSummary item in _template.allItems) {
      if (item.itemNo > 0) {
        itemNos.add(item.itemNo);
      }
    }
    if (itemNos.length < _totalCount) {
      for (int itemNo = 1; itemNo <= _totalCount; itemNo += 1) {
        itemNos.add(itemNo);
      }
    }
    for (final int itemNo in itemNos) {
      _itemScores.putIfAbsent(itemNo, () => 0);
    }
  }

  Map<String, dynamic> _buildDraftPayload() {
    return <String, dynamic>{
      if (_draftId > 0) 'id': _draftId,
      'studentId': _studentId,
      'studentName': _studentName.trim(),
      'examinerName': _examinerName.trim(),
      'remark': _remarkController.text.trim(),
      'birthDate': _dateOnlyText(_birthDate),
      'assessmentDate': _dateOnlyText(_assessmentDate),
      'itemScoreList': _itemScoreList(),
    };
  }

  List<Map<String, dynamic>> _itemScoreList() {
    final List<int> itemNos = _itemScores.keys.toList()..sort();
    return <Map<String, dynamic>>[
      for (final int itemNo in itemNos)
        <String, dynamic>{
          'itemNo': itemNo,
          'score': _itemScores[itemNo],
        },
    ];
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
      key: 'shuangxi-assessment-top-message',
    );
  }

  void _scheduleAutoAdvance(int currentItemNo) {
    final int next = _nextItemNo(afterItemNo: currentItemNo);
    if (next <= 0 || next == currentItemNo) {
      return;
    }
    _autoAdvanceTimer = Timer(const Duration(milliseconds: 650), () {
      _autoAdvanceTimer = null;
      if (!mounted || !_autoNext || _selectedItemNo != currentItemNo) {
        return;
      }
      unawaited(_selectItem(next));
    });
  }

  void _cancelPendingAutoAdvance() {
    _autoAdvanceTimer?.cancel();
    _autoAdvanceTimer = null;
  }

  void _goPrevious() {
    final List<ShuangxiItemSummary> items = _template.allItems;
    final int index = items.indexWhere(
        (ShuangxiItemSummary item) => item.itemNo == _selectedItemNo);
    if (index > 0) {
      unawaited(_selectItem(items[index - 1].itemNo));
    }
  }

  void _goNext() {
    final int next = _nextItemNo(afterItemNo: _selectedItemNo);
    if (next > 0) {
      unawaited(_selectItem(next));
    }
  }

  void _goFirstMissing() {
    final int next = _firstMissingItemNo();
    if (next > 0) {
      unawaited(_selectItem(next));
    }
  }

  int _nextItemNo({required int afterItemNo}) {
    final List<ShuangxiItemSummary> items = _template.allItems;
    if (items.isEmpty) {
      return 0;
    }
    final int index = items
        .indexWhere((ShuangxiItemSummary item) => item.itemNo == afterItemNo);
    if (index >= 0 && index < items.length - 1) {
      return items[index + 1].itemNo;
    }
    return 0;
  }

  int _firstMissingItemNo() {
    for (final ShuangxiItemSummary item in _template.allItems) {
      if (!_itemScores.containsKey(item.itemNo)) {
        return item.itemNo;
      }
    }
    return 0;
  }

  ShuangxiItemSummary get _currentItemSummary {
    for (final ShuangxiItemSummary item in _template.allItems) {
      if (item.itemNo == _selectedItemNo) {
        return item;
      }
    }
    return ShuangxiItemSummary.empty;
  }

  ShuangxiDomainSummary? _domainByCode(String domainCode) {
    final String normalized = domainCode.trim();
    for (final ShuangxiDomainSummary domain in _template.domains) {
      if (domain.domainCode == normalized) {
        return domain;
      }
    }
    return null;
  }

  String get _selectedDomainCode {
    final String itemDomain = _currentItem.domainCode.trim().isNotEmpty
        ? _currentItem.domainCode.trim()
        : _currentItemSummary.domainCode.trim();
    if (itemDomain.isNotEmpty) {
      return itemDomain;
    }
    return _template.domains.isEmpty ? '' : _template.domains.first.domainCode;
  }

  String get _selectedSkillCode {
    final String itemSkill = _currentItem.skillCode.trim().isNotEmpty
        ? _currentItem.skillCode.trim()
        : _currentItemSummary.skillCode.trim();
    return itemSkill;
  }

  int get _answeredCount => _itemScores.length;

  int get _totalCount {
    if (_template.itemCount > 0) {
      return _template.itemCount;
    }
    return _template.allItems.length;
  }

  int get _progressPercent {
    final int total = _totalCount;
    if (total <= 0) {
      return 0;
    }
    return ((_answeredCount / total) * 100).round().clamp(0, 100);
  }

  String get _studentAgeText {
    final String fallback =
        _studentAge.trim().isEmpty ? '未知' : _studentAge.trim();
    return formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
      fallback: fallback,
    );
  }

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _ShuangxiColors.page,
      child: Column(
        children: <Widget>[
          _TopBar(
            scaleName: widget.args.scaleName,
            studentName: _studentName,
            studentAge: _studentAgeText,
            assessmentDate: _assessmentDate,
            examinerName: _examinerName,
            autoSaveText: _autoSaveText,
            saving: _savingDraft,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: () => unawaited(_saveDraft()),
            onSubmit: () => unawaited(_submitDraft()),
          ),
          if (_loading)
            const Expanded(child: _LoadingState())
          else if (_errorMessage.isNotEmpty)
            Expanded(
              child: _ErrorState(message: _errorMessage, onRetry: _initialize),
            )
          else ...<Widget>[
            _DimensionOverview(
              template: _template,
              itemScores: _itemScores,
              selectedDomainCode: _selectedDomainCode,
              onSelectDomain: _selectDomain,
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    SizedBox(
                      width: 252,
                      child: _SkillNavigator(
                        template: _template,
                        selectedDomainCode: _selectedDomainCode,
                        selectedSkillCode: _selectedSkillCode,
                        selectedItemNo: _selectedItemNo,
                        itemScores: _itemScores,
                        onSelectItem: (int itemNo) =>
                            unawaited(_selectItem(itemNo)),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _QuestionWorkspace(
                        item: _currentItem,
                        summary: _currentItemSummary,
                        fallbackScoreOptions: _template.scoreOptions,
                        itemLoading: _itemLoading,
                        selectedScore: _itemScores[_selectedItemNo] ?? -1,
                        onSelectScore: _selectScore,
                      ),
                    ),
                    const SizedBox(width: 10),
                    SizedBox(
                      width: 270,
                      child: _RightRail(
                        progressPercent: _progressPercent,
                        answered: _answeredCount,
                        total: _totalCount,
                        remarkController: _remarkController,
                        currentDomain: _domainByCode(_selectedDomainCode),
                        currentSkillCode: _selectedSkillCode,
                        firstMissingTitle: _firstMissingTitle(),
                        autoNext: _autoNext,
                        onRemarkChanged: _handleRemarkChanged,
                        onRemarkEditingComplete: _finishRemarkEditing,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            _FooterDock(
              current: _selectedItemNo,
              total: _totalCount,
              hasPrevious: _hasPreviousItem,
              hasNext: _hasNextItem,
              hasMissing: _firstMissingItemNo() > 0,
              autoNext: _autoNext,
              onPrevious: _goPrevious,
              onNext: _goNext,
              onJumpMissing: _goFirstMissing,
              onToggleAutoNext: (bool value) {
                if (!value) {
                  _cancelPendingAutoAdvance();
                }
                setState(() => _autoNext = value);
              },
            ),
          ],
        ],
      ),
    );
  }

  bool get _hasPreviousItem {
    final List<ShuangxiItemSummary> items = _template.allItems;
    return items.indexWhere(
          (ShuangxiItemSummary item) => item.itemNo == _selectedItemNo,
        ) >
        0;
  }

  bool get _hasNextItem {
    final List<ShuangxiItemSummary> items = _template.allItems;
    final int index = items.indexWhere(
      (ShuangxiItemSummary item) => item.itemNo == _selectedItemNo,
    );
    return index >= 0 && index < items.length - 1;
  }

  String _firstMissingTitle() {
    final int next = _firstMissingItemNo();
    if (next <= 0) {
      return '已无缺题';
    }
    for (final ShuangxiItemSummary item in _template.allItems) {
      if (item.itemNo == next) {
        return item.itemTitle.trim().isEmpty
            ? '${item.itemCode} ${item.testItem}'.trim()
            : item.itemTitle.trim();
      }
    }
    return '第 $next 题';
  }
}

class _ShuangxiColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color green = Color(0xFF5B9E68);
}

List<BoxShadow> _shuangxiShadow({
  Color color = const Color(0x12B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 8),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

BoxDecoration _panelDecoration({double radius = 8}) {
  return BoxDecoration(
    color: Colors.white.withOpacity(.92),
    borderRadius: BorderRadius.circular(radius),
    border: Border.all(color: _ShuangxiColors.line),
    boxShadow: _shuangxiShadow(),
  );
}

class _TopBar extends StatelessWidget {
  const _TopBar({
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
        scaleName.trim().isEmpty ? '双溪课程评量表A' : scaleName.trim();
    final String student =
        studentName.trim().isEmpty ? '-' : studentName.trim();
    final String date =
        assessmentDate.trim().isEmpty ? _todayIsoDate() : assessmentDate.trim();
    final String age = studentAge.trim().isEmpty ? '未知' : studentAge.trim();
    final String examiner =
        examinerName.trim().isEmpty ? '-' : examinerName.trim();

    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.98),
        border: Border.all(color: _ShuangxiColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _shuangxiShadow(color: const Color(0x14B05F32), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1120;
          final List<Widget> headerChildren = <Widget>[
            Text(
              '$title 测评工作台',
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _ShuangxiColors.ink,
                fontSize: 23,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            _HeaderMeta(
              label: '儿童',
              value: student,
              compact: compact,
            ),
            _HeaderMeta(
              label: '年龄',
              value: age,
              compact: compact,
            ),
            _HeaderMeta(
              label: compact ? '日期' : '测评日期',
              value: date,
              compact: compact,
            ),
            _HeaderMeta(
              label: '施测者',
              value: examiner,
              compact: compact,
            ),
          ];
          return Row(
            children: <Widget>[
              _IconButtonBox(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: compact
                    ? SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        physics: const ClampingScrollPhysics(),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: headerChildren,
                        ),
                      )
                    : Row(children: headerChildren),
              ),
              if (!compact) ...<Widget>[
                const Icon(
                  Icons.check_circle_outline_rounded,
                  color: _ShuangxiColors.green,
                  size: 18,
                ),
                const SizedBox(width: 5),
                Text(
                  autoSaveText.trim().isEmpty ? '已自动保存' : autoSaveText.trim(),
                  maxLines: 1,
                  softWrap: false,
                  style: const TextStyle(
                    color: _ShuangxiColors.body,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(width: 8),
              ],
              _TopActionButton(
                label: saving ? '保存中' : '保存草稿',
                icon: Icons.save_outlined,
                filled: false,
                onTap: saving ? null : onSave,
              ),
              const SizedBox(width: 6),
              _TopActionButton(
                label: submitting ? '提交中' : '提交记录',
                icon: Icons.fact_check_outlined,
                filled: true,
                onTap: submitting ? null : onSubmit,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _HeaderMeta extends StatelessWidget {
  const _HeaderMeta({
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
        border: Border(left: BorderSide(color: _ShuangxiColors.line)),
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
          color: _ShuangxiColors.body,
          fontSize: 13,
          height: 1,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _IconButtonBox extends StatelessWidget {
  const _IconButtonBox({required this.icon, required this.onTap});

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
            border: Border.all(color: _ShuangxiColors.line),
          ),
          child: Icon(icon, color: _ShuangxiColors.body, size: 34),
        ),
      ),
    );
  }
}

class _TopActionButton extends StatelessWidget {
  const _TopActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 34,
          padding: const EdgeInsets.symmetric(horizontal: 11),
          decoration: BoxDecoration(
            color: filled ? _ShuangxiColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _ShuangxiColors.orange),
            boxShadow: filled
                ? _shuangxiShadow(
                    color: const Color(0x2AE96F43),
                    blur: 12,
                    offset: const Offset(0, 5),
                  )
                : null,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                icon,
                size: 18,
                color: filled ? Colors.white : _ShuangxiColors.orangeDeep,
              ),
              const SizedBox(width: 6),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _ShuangxiColors.orangeDeep,
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

class _SubmitZeroScoreDialog extends StatelessWidget {
  const _SubmitZeroScoreDialog({
    required this.missingCount,
    required this.totalCount,
    required this.answeredCount,
    required this.onCancel,
    required this.onConfirm,
  });

  final int missingCount;
  final int totalCount;
  final int answeredCount;
  final VoidCallback onCancel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 32),
      backgroundColor: Colors.transparent,
      child: Container(
        width: 548,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x33000000),
              blurRadius: 30,
              offset: Offset(0, 18),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
              child: Row(
                children: <Widget>[
                  Container(
                    width: 38,
                    height: 38,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFEEE3),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(
                      Icons.priority_high_rounded,
                      color: _ShuangxiColors.orange,
                      size: 26,
                    ),
                  ),
                  const SizedBox(width: 12),
                  const Expanded(
                    child: Text(
                      '未评分题将按 0 分提交',
                      maxLines: 1,
                      softWrap: false,
                      style: TextStyle(
                        color: _ShuangxiColors.ink,
                        fontSize: 19,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const Divider(height: 1, color: _ShuangxiColors.lineSoft),
            Padding(
              padding: const EdgeInsets.fromLTRB(30, 26, 30, 26),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    '当前还有 $missingCount 道题未评分。确认提交后，系统会把这些题目默认记录为 0 分，并生成正式测评记录。',
                    style: const TextStyle(
                      color: _ShuangxiColors.ink,
                      fontSize: 15,
                      height: 1.35,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 18),
                  Container(
                    padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFFBF7),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(color: _ShuangxiColors.line),
                    ),
                    child: Column(
                      children: <Widget>[
                        _SubmitZeroScoreMetaRow(
                          label: '已评分题',
                          value: '$answeredCount / $totalCount',
                        ),
                        const SizedBox(height: 13),
                        _SubmitZeroScoreMetaRow(
                          label: '未评分题',
                          value: '$missingCount 道',
                        ),
                        const SizedBox(height: 13),
                        const _SubmitZeroScoreMetaRow(
                          label: '默认评分',
                          value: '0 分',
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const Divider(height: 1, color: _ShuangxiColors.lineSoft),
            Padding(
              padding: const EdgeInsets.fromLTRB(30, 18, 30, 20),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  OutlinedButton.icon(
                    onPressed: onCancel,
                    icon: const Icon(Icons.edit_note_rounded, size: 18),
                    label: const Text('继续作答'),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: _ShuangxiColors.body,
                      side: const BorderSide(color: _ShuangxiColors.line),
                      minimumSize: const Size(118, 40),
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      textStyle: const TextStyle(
                        fontSize: 14,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  FilledButton.icon(
                    key: const ValueKey<String>(
                      'shuangxi-submit-zero-confirm-button',
                    ),
                    onPressed: onConfirm,
                    icon: const Icon(Icons.check_circle_outline_rounded,
                        size: 18),
                    label: const Text('按 0 分提交'),
                    style: FilledButton.styleFrom(
                      backgroundColor: _ShuangxiColors.orange,
                      foregroundColor: Colors.white,
                      minimumSize: const Size(136, 40),
                      padding: const EdgeInsets.symmetric(horizontal: 18),
                      textStyle: const TextStyle(
                        fontSize: 14,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
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

class _SubmitZeroScoreMetaRow extends StatelessWidget {
  const _SubmitZeroScoreMetaRow({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _ShuangxiColors.body,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w800,
          ),
        ),
        const Spacer(),
        Text(
          value,
          style: const TextStyle(
            color: _ShuangxiColors.ink,
            fontSize: 14,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _DimensionOverview extends StatelessWidget {
  const _DimensionOverview({
    required this.template,
    required this.itemScores,
    required this.selectedDomainCode,
    required this.onSelectDomain,
  });

  final ShuangxiTemplateSummary template;
  final Map<int, int> itemScores;
  final String selectedDomainCode;
  final ValueChanged<String> onSelectDomain;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 84,
      margin: const EdgeInsets.fromLTRB(12, 8, 12, 0),
      padding: const EdgeInsets.all(10),
      decoration: _panelDecoration(radius: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          for (int index = 0;
              index < template.domains.length;
              index++) ...<Widget>[
            Expanded(
              child: _DimensionCard(
                domain: template.domains[index],
                answered: _answeredInDomain(template.domains[index]),
                active:
                    template.domains[index].domainCode == selectedDomainCode,
                onTap: () => onSelectDomain(template.domains[index].domainCode),
              ),
            ),
            if (index != template.domains.length - 1) const SizedBox(width: 8),
          ],
        ],
      ),
    );
  }

  int _answeredInDomain(ShuangxiDomainSummary domain) {
    int count = 0;
    for (final ShuangxiSkillSummary skill in domain.skills) {
      for (final ShuangxiItemSummary item in skill.items) {
        if (itemScores.containsKey(item.itemNo)) {
          count++;
        }
      }
    }
    return count;
  }
}

class _DimensionCard extends StatelessWidget {
  const _DimensionCard({
    required this.domain,
    required this.answered,
    required this.active,
    required this.onTap,
  });

  final ShuangxiDomainSummary domain;
  final int answered;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final int total = domain.itemCount <= 0
        ? domain.skills.fold<int>(
            0,
            (int sum, ShuangxiSkillSummary skill) => sum + skill.items.length,
          )
        : domain.itemCount;
    final double progress = total <= 0 ? 0 : answered / total;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          padding: const EdgeInsets.fromLTRB(7, 7, 7, 6),
          decoration: BoxDecoration(
            color: active ? const Color(0xFFFFF7EF) : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: active ? _ShuangxiColors.orange : _ShuangxiColors.line,
              width: active ? 1.2 : 1,
            ),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    Row(
                      children: <Widget>[
                        Container(
                          width: 30,
                          height: 30,
                          decoration: BoxDecoration(
                            color: active
                                ? const Color(0xFFFFE5D4)
                                : const Color(0xFFFFF8F1),
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            _domainIcon(domain.domainCode, domain.domainName),
                            color: _ShuangxiColors.ink,
                            size: 18,
                          ),
                        ),
                        const SizedBox(width: 7),
                        Expanded(
                          child: Text.rich(
                            TextSpan(
                              children: <InlineSpan>[
                                TextSpan(text: domain.domainName),
                                TextSpan(
                                  text: '  $answered/$total',
                                  style: TextStyle(
                                    color: active
                                        ? _ShuangxiColors.orange
                                        : _ShuangxiColors.ink,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                              ],
                            ),
                            maxLines: 1,
                            softWrap: false,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _ShuangxiColors.ink,
                              fontSize: 12,
                              height: 1,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        minHeight: 4,
                        value: progress.clamp(0, 1),
                        backgroundColor: const Color(0xFFE6E1DE),
                        valueColor: const AlwaysStoppedAnimation<Color>(
                          _ShuangxiColors.orange,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SkillNavigator extends StatelessWidget {
  const _SkillNavigator({
    required this.template,
    required this.selectedDomainCode,
    required this.selectedSkillCode,
    required this.selectedItemNo,
    required this.itemScores,
    required this.onSelectItem,
  });

  final ShuangxiTemplateSummary template;
  final String selectedDomainCode;
  final String selectedSkillCode;
  final int selectedItemNo;
  final Map<int, int> itemScores;
  final ValueChanged<int> onSelectItem;

  @override
  Widget build(BuildContext context) {
    final ShuangxiDomainSummary domain = template.domains.firstWhere(
      (ShuangxiDomainSummary item) => item.domainCode == selectedDomainCode,
      orElse: () => template.domains.isEmpty
          ? const ShuangxiDomainSummary(
              domainCode: '',
              domainName: '维度',
              sortNo: 0,
              itemCount: 0,
              maxRawScore: 0,
              skills: <ShuangxiSkillSummary>[],
            )
          : template.domains.first,
    );
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 12, 10, 10),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Icon(
                _domainIcon(domain.domainCode, domain.domainName),
                color: _ShuangxiColors.orange,
                size: 22,
              ),
              const SizedBox(width: 7),
              Text(
                domain.domainName,
                style: const TextStyle(
                  color: _ShuangxiColors.ink,
                  fontSize: 18,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 9),
          Expanded(
            child: ListView(
              padding: EdgeInsets.zero,
              physics: const BouncingScrollPhysics(),
              children: <Widget>[
                for (final ShuangxiSkillSummary skill in domain.skills)
                  _SkillSection(
                    skill: skill,
                    answered: _answeredInSkill(skill),
                    expanded: skill.skillCode == selectedSkillCode,
                    selectedItemNo: selectedItemNo,
                    itemScores: itemScores,
                    onSelectItem: onSelectItem,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  int _answeredInSkill(ShuangxiSkillSummary skill) {
    int count = 0;
    for (final ShuangxiItemSummary item in skill.items) {
      if (itemScores.containsKey(item.itemNo)) {
        count++;
      }
    }
    return count;
  }
}

class _SkillSection extends StatelessWidget {
  const _SkillSection({
    required this.skill,
    required this.answered,
    required this.expanded,
    required this.selectedItemNo,
    required this.itemScores,
    required this.onSelectItem,
  });

  final ShuangxiSkillSummary skill;
  final int answered;
  final bool expanded;
  final int selectedItemNo;
  final Map<int, int> itemScores;
  final ValueChanged<int> onSelectItem;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      decoration: BoxDecoration(
        color: expanded ? const Color(0xFFFFF8F1) : Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: expanded ? const Color(0xFFFFC7A7) : _ShuangxiColors.lineSoft,
        ),
      ),
      child: Column(
        children: <Widget>[
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: skill.items.isEmpty
                  ? null
                  : () => onSelectItem(skill.items.first.itemNo),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(8),
              ),
              child: SizedBox(
                height: 36,
                child: Row(
                  children: <Widget>[
                    const SizedBox(width: 12),
                    Container(
                      width: 14,
                      height: 14,
                      decoration: BoxDecoration(
                        color: expanded ? _ShuangxiColors.orange : Colors.white,
                        shape: BoxShape.circle,
                        border: Border.all(
                          color: expanded
                              ? _ShuangxiColors.orange
                              : const Color(0xFFCFC7C2),
                        ),
                      ),
                    ),
                    const SizedBox(width: 9),
                    Expanded(
                      child: Text(
                        '${skill.skillCode}  ${skill.skillName}',
                        maxLines: 1,
                        style: const TextStyle(
                          color: _ShuangxiColors.ink,
                          fontSize: 14,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    Text(
                      '$answered/${skill.itemCount}',
                      style: TextStyle(
                        color: expanded
                            ? _ShuangxiColors.orange
                            : _ShuangxiColors.muted,
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Icon(
                      expanded
                          ? Icons.keyboard_arrow_up_rounded
                          : Icons.keyboard_arrow_down_rounded,
                      color: _ShuangxiColors.muted,
                      size: 20,
                    ),
                    const SizedBox(width: 6),
                  ],
                ),
              ),
            ),
          ),
          if (expanded)
            Padding(
              padding: const EdgeInsets.fromLTRB(32, 0, 8, 8),
              child: Column(
                children: <Widget>[
                  for (final ShuangxiItemSummary item in skill.items)
                    _QuestionNavRow(
                      item: item,
                      active: item.itemNo == selectedItemNo,
                      answered: itemScores.containsKey(item.itemNo),
                      onTap: () => onSelectItem(item.itemNo),
                    ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _QuestionNavRow extends StatelessWidget {
  const _QuestionNavRow({
    required this.item,
    required this.active,
    required this.answered,
    required this.onTap,
  });

  final ShuangxiItemSummary item;
  final bool active;
  final bool answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final String title = item.itemTitle.trim().isEmpty
        ? '${item.itemCode} ${item.testItem}'.trim()
        : item.itemTitle.trim();
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(6),
        child: SizedBox(
          height: 31,
          child: Row(
            children: <Widget>[
              Container(
                width: 10,
                height: 10,
                decoration: BoxDecoration(
                  color: answered ? _ShuangxiColors.green : Colors.white,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: active
                        ? _ShuangxiColors.orange
                        : answered
                            ? _ShuangxiColors.green
                            : const Color(0xFFCFC7C2),
                    width: active ? 2 : 1.2,
                  ),
                ),
              ),
              const SizedBox(width: 9),
              Expanded(
                child: Text(
                  title,
                  maxLines: 1,
                  style: TextStyle(
                    color:
                        active ? _ShuangxiColors.orange : _ShuangxiColors.ink,
                    fontSize: 12,
                    fontWeight: active ? FontWeight.w900 : FontWeight.w700,
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

class _QuestionWorkspace extends StatelessWidget {
  const _QuestionWorkspace({
    required this.item,
    required this.summary,
    required this.fallbackScoreOptions,
    required this.itemLoading,
    required this.selectedScore,
    required this.onSelectScore,
  });

  final ShuangxiAssessmentItem item;
  final ShuangxiItemSummary summary;
  final List<ShuangxiScoreOption> fallbackScoreOptions;
  final bool itemLoading;
  final int selectedScore;
  final ValueChanged<int> onSelectScore;

  @override
  Widget build(BuildContext context) {
    final String skillName = item.skillName.trim().isNotEmpty
        ? item.skillName.trim()
        : summary.skillName.trim();
    final String skillCode = item.skillCode.trim().isNotEmpty
        ? item.skillCode.trim()
        : summary.skillCode.trim();
    final String itemTitle = item.itemTitle.trim().isNotEmpty
        ? item.itemTitle.trim()
        : summary.itemTitle.trim().isNotEmpty
            ? summary.itemTitle.trim()
            : '${summary.itemCode} ${summary.testItem}'.trim();
    final int itemNo = item.itemNo > 0 ? item.itemNo : summary.itemNo;
    final List<ShuangxiScoreOption> scoreOptions = item.scoreOptions.isNotEmpty
        ? item.scoreOptions
        : fallbackScoreOptions.isNotEmpty
            ? fallbackScoreOptions
            : const <ShuangxiScoreOption>[
                ShuangxiScoreOption(
                    value: 0, label: '0分', description: '尚未出现或无法完成'),
                ShuangxiScoreOption(
                    value: 1, label: '1分', description: '大量协助下可完成'),
                ShuangxiScoreOption(
                    value: 2, label: '2分', description: '少量协助或提示下可完成'),
                ShuangxiScoreOption(
                    value: 3, label: '3分', description: '能独立稳定完成'),
              ];
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 12),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          SizedBox(
            height: 28,
            child: Row(
              children: <Widget>[
                Container(
                  width: 4,
                  height: 18,
                  decoration: BoxDecoration(
                    color: _ShuangxiColors.orange,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(width: 9),
                Icon(
                  Icons.account_tree_outlined,
                  color: _ShuangxiColors.orange,
                  size: 18,
                ),
                const SizedBox(width: 7),
                Expanded(
                  child: Text(
                    '$skillCode $skillName'.trim(),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _ShuangxiColors.orangeDeep,
                      fontSize: 14,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                if (itemNo > 0) ...<Widget>[
                  const SizedBox(width: 10),
                  Text(
                    '第 $itemNo 题',
                    maxLines: 1,
                    style: const TextStyle(
                      color: _ShuangxiColors.muted,
                      fontSize: 13,
                      height: 1,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 12),
          Text(
            itemTitle.isEmpty ? '题目加载中' : itemTitle,
            maxLines: 1,
            style: const TextStyle(
              color: _ShuangxiColors.ink,
              fontSize: 23,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 20),
          if (itemLoading)
            const LinearProgressIndicator(
              minHeight: 2,
              color: _ShuangxiColors.orange,
              backgroundColor: Color(0xFFFFE6D7),
            )
          else
            for (int index = 0;
                index < scoreOptions.length;
                index++) ...<Widget>[
              _ScoreChoice(
                option: scoreOptions[index],
                selected: selectedScore == scoreOptions[index].value,
                onTap: () => onSelectScore(scoreOptions[index].value),
              ),
              if (index != scoreOptions.length - 1) const SizedBox(height: 12),
            ],
        ],
      ),
    );
  }
}

class _ScoreChoice extends StatelessWidget {
  const _ScoreChoice({
    required this.option,
    required this.selected,
    required this.onTap,
  });

  final ShuangxiScoreOption option;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          height: 54,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF7EF) : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: selected ? _ShuangxiColors.orange : _ShuangxiColors.line,
              width: selected ? 1.3 : 1,
            ),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 20,
                height: 20,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: selected
                        ? _ShuangxiColors.orange
                        : const Color(0xFFD6CEC9),
                    width: selected ? 2.2 : 1.3,
                  ),
                ),
                child: selected
                    ? Center(
                        child: Container(
                          width: 12,
                          height: 12,
                          decoration: const BoxDecoration(
                            color: _ShuangxiColors.orange,
                            shape: BoxShape.circle,
                          ),
                        ),
                      )
                    : null,
              ),
              const SizedBox(width: 12),
              SizedBox(
                width: 42,
                child: Text(
                  option.label.trim().isEmpty
                      ? '${option.value}分'
                      : option.label,
                  style: const TextStyle(
                    color: _ShuangxiColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              Expanded(
                child: Text(
                  option.description,
                  maxLines: 1,
                  style: const TextStyle(
                    color: _ShuangxiColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
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

class _RightRail extends StatelessWidget {
  const _RightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.remarkController,
    required this.currentDomain,
    required this.currentSkillCode,
    required this.firstMissingTitle,
    required this.autoNext,
    required this.onRemarkChanged,
    required this.onRemarkEditingComplete,
  });

  final int progressPercent;
  final int answered;
  final int total;
  final TextEditingController remarkController;
  final ShuangxiDomainSummary? currentDomain;
  final String currentSkillCode;
  final String firstMissingTitle;
  final bool autoNext;
  final ValueChanged<String> onRemarkChanged;
  final VoidCallback onRemarkEditingComplete;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _ProgressPanel(
          progressPercent: progressPercent,
          answered: answered,
          total: total,
        ),
        const SizedBox(height: 8),
        _RemarkPanel(
          controller: remarkController,
          onChanged: onRemarkChanged,
          onEditingComplete: onRemarkEditingComplete,
        ),
        const SizedBox(height: 8),
        Expanded(
          child: _MissingNavigationPanel(
            currentDomain: currentDomain,
            currentSkillCode: currentSkillCode,
            firstMissingTitle: firstMissingTitle,
            autoNext: autoNext,
          ),
        ),
      ],
    );
  }
}

class _ProgressPanel extends StatelessWidget {
  const _ProgressPanel({
    required this.progressPercent,
    required this.answered,
    required this.total,
  });

  final int progressPercent;
  final int answered;
  final int total;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 148,
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 12),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const _PanelTitle(icon: Icons.bar_chart_rounded, title: '测评进度'),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              _ProgressRing(percent: progressPercent),
              const SizedBox(width: 9),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    _SummaryRow(label: '已完成', value: '$answered / $total 题'),
                    const SizedBox(height: 9),
                    _SummaryRow(
                      label: '缺题',
                      value: '${math.max(0, total - answered)} 题',
                      danger: true,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _PanelTitle extends StatelessWidget {
  const _PanelTitle({required this.icon, required this.title});

  final IconData icon;
  final String title;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Icon(icon, color: _ShuangxiColors.orange, size: 20),
        const SizedBox(width: 7),
        Text(
          title,
          style: const TextStyle(
            color: _ShuangxiColors.ink,
            fontSize: 16,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _ProgressRing extends StatelessWidget {
  const _ProgressRing({required this.percent});

  final int percent;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 74,
      height: 74,
      child: Stack(
        alignment: Alignment.center,
        children: <Widget>[
          CustomPaint(
            painter: _ShuangxiDonutPainter(percent: percent),
            size: const Size(74, 74),
          ),
          Text(
            '$percent%',
            style: TextStyle(
              color: _ShuangxiColors.ink,
              fontSize: 19,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _SummaryRow extends StatelessWidget {
  const _SummaryRow({
    required this.label,
    required this.value,
    this.danger = false,
  });

  final String label;
  final String value;
  final bool danger;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              label,
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _ShuangxiColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          Text(
            value,
            maxLines: 1,
            softWrap: false,
            style: TextStyle(
              color: danger ? const Color(0xFFE04438) : _ShuangxiColors.orange,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _RemarkPanel extends StatelessWidget {
  const _RemarkPanel({
    required this.controller,
    required this.onChanged,
    required this.onEditingComplete,
  });

  final TextEditingController controller;
  final ValueChanged<String> onChanged;
  final VoidCallback onEditingComplete;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 142,
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 10),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const _PanelTitle(icon: Icons.assignment_outlined, title: '观察备注'),
          const SizedBox(height: 8),
          Expanded(
            child: TextField(
              key: const ValueKey<String>('shuangxi-observation-remark-field'),
              controller: controller,
              expands: true,
              minLines: null,
              maxLines: null,
              maxLength: 300,
              inputFormatters: <TextInputFormatter>[
                LengthLimitingTextInputFormatter(300),
              ],
              keyboardType: TextInputType.multiline,
              textInputAction: TextInputAction.newline,
              textAlignVertical: TextAlignVertical.top,
              onChanged: onChanged,
              onEditingComplete: onEditingComplete,
              onTapOutside: (_) => onEditingComplete(),
              style: const TextStyle(
                color: _ShuangxiColors.ink,
                fontSize: 13,
                height: 1.35,
                fontWeight: FontWeight.w700,
              ),
              decoration: InputDecoration(
                hintText: '记录学生反应、协助程度或环境因素',
                hintStyle: const TextStyle(
                  color: _ShuangxiColors.muted,
                  fontSize: 13,
                  height: 1.25,
                  fontWeight: FontWeight.w600,
                ),
                filled: true,
                fillColor: Colors.white,
                counterText: '',
                contentPadding: const EdgeInsets.fromLTRB(11, 10, 11, 10),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: _ShuangxiColors.line),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(
                    color: _ShuangxiColors.orange,
                    width: 1.2,
                  ),
                ),
              ),
            ),
          ),
          ValueListenableBuilder<TextEditingValue>(
            valueListenable: controller,
            builder: (
              BuildContext context,
              TextEditingValue value,
              Widget? child,
            ) {
              return Align(
                alignment: Alignment.bottomRight,
                child: Text(
                  '${value.text.characters.length}/300',
                  style: const TextStyle(
                    color: _ShuangxiColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

class _MissingNavigationPanel extends StatelessWidget {
  const _MissingNavigationPanel({
    required this.currentDomain,
    required this.currentSkillCode,
    required this.firstMissingTitle,
    required this.autoNext,
  });

  final ShuangxiDomainSummary? currentDomain;
  final String currentSkillCode;
  final String firstMissingTitle;
  final bool autoNext;

  @override
  Widget build(BuildContext context) {
    final ShuangxiDomainSummary? domain = currentDomain;
    final ShuangxiSkillSummary? skill = domain == null
        ? null
        : domain.skills.cast<ShuangxiSkillSummary?>().firstWhere(
              (ShuangxiSkillSummary? item) =>
                  item?.skillCode == currentSkillCode,
              orElse: () => null,
            );
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 10),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const _PanelTitle(
              icon: Icons.lightbulb_outline_rounded, title: '缺题导航'),
          const SizedBox(height: 9),
          _MissingInfoRow(
            icon: Icons.auto_awesome_rounded,
            text: '自动下一题：${autoNext ? '已开启' : '已关闭'}',
          ),
          _MissingInfoRow(
            icon: Icons.grid_view_rounded,
            text: '当前维度：${domain?.domainName ?? '-'}',
          ),
          _MissingInfoRow(
            icon: Icons.account_tree_outlined,
            text:
                '当前技能：${skill == null ? '-' : '${skill.skillCode} ${skill.skillName}'}',
          ),
          _MissingInfoRow(
            icon: Icons.arrow_forward_rounded,
            text: '第一缺题：$firstMissingTitle',
          ),
        ],
      ),
    );
  }
}

class _MissingInfoRow extends StatelessWidget {
  const _MissingInfoRow({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      margin: const EdgeInsets.only(bottom: 6),
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: _ShuangxiColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _ShuangxiColors.orange, size: 17),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              maxLines: 1,
              style: const TextStyle(
                color: _ShuangxiColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _FooterDock extends StatelessWidget {
  const _FooterDock({
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
        border: Border.all(color: _ShuangxiColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
        boxShadow: _shuangxiShadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _FooterButton(
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
                    color: _ShuangxiColors.ink,
                    fontSize: 26,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _ShuangxiColors.body,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.arrow_forward_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.format_list_bulleted_rounded,
            enabled: hasMissing,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _ShuangxiColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _ShuangxiColors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
      ),
    );
  }
}

class _FooterButton extends StatelessWidget {
  const _FooterButton({
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
            : _ShuangxiColors.orangeDeep
        : _ShuangxiColors.muted;
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
                ? _ShuangxiColors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _ShuangxiColors.orange : const Color(0xFFE2D6CE),
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

class _LoadingState extends StatelessWidget {
  const _LoadingState();

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: <Widget>[
        _LoadingDimensionOverview(),
        Expanded(
          child: Padding(
            padding: EdgeInsets.fromLTRB(12, 8, 12, 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                SizedBox(width: 252, child: _LoadingSkillNavigator()),
                SizedBox(width: 10),
                Expanded(child: _LoadingQuestionWorkspace()),
                SizedBox(width: 10),
                SizedBox(width: 270, child: _LoadingRightRail()),
              ],
            ),
          ),
        ),
        _LoadingFooterDock(),
      ],
    );
  }
}

class _LoadingDimensionOverview extends StatelessWidget {
  const _LoadingDimensionOverview();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 84,
      margin: const EdgeInsets.fromLTRB(12, 8, 12, 0),
      padding: const EdgeInsets.all(10),
      decoration: _panelDecoration(radius: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          for (int index = 0; index < 7; index++) ...<Widget>[
            Expanded(child: _LoadingDimensionCard(active: index == 0)),
            if (index != 6) const SizedBox(width: 8),
          ],
        ],
      ),
    );
  }
}

class _LoadingDimensionCard extends StatelessWidget {
  const _LoadingDimensionCard({required this.active});

  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(7, 7, 7, 6),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF7EF) : Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: active ? _ShuangxiColors.orange : _ShuangxiColors.line,
          width: active ? 1.2 : 1,
        ),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              _ShuangxiSkeletonBlock(
                width: 30,
                height: 30,
                radius: 15,
                highlight: active,
              ),
              const SizedBox(width: 7),
              const Expanded(
                child: _ShuangxiSkeletonBlock(height: 12, widthFactor: .82),
              ),
            ],
          ),
          const SizedBox(height: 8),
          _ShuangxiSkeletonBlock(height: 4, radius: 4, highlight: active),
        ],
      ),
    );
  }
}

class _LoadingSkillNavigator extends StatelessWidget {
  const _LoadingSkillNavigator();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 12, 10, 10),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: const <Widget>[
              Icon(
                Icons.account_tree_outlined,
                color: _ShuangxiColors.orange,
                size: 22,
              ),
              SizedBox(width: 7),
              Expanded(
                child: _ShuangxiSkeletonBlock(height: 17, widthFactor: .62),
              ),
            ],
          ),
          const SizedBox(height: 9),
          Expanded(
            child: SingleChildScrollView(
              physics: const NeverScrollableScrollPhysics(),
              child: Column(
                children: const <Widget>[
                  _LoadingSkillSection(expanded: true),
                  _LoadingSkillSection(expanded: false),
                  _LoadingSkillSection(expanded: false),
                  _LoadingSkillSection(expanded: false),
                  _LoadingSkillSection(expanded: false),
                  _LoadingSkillSection(expanded: false),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _LoadingSkillSection extends StatelessWidget {
  const _LoadingSkillSection({required this.expanded});

  final bool expanded;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      decoration: BoxDecoration(
        color: expanded ? const Color(0xFFFFF8F1) : Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: expanded ? const Color(0xFFFFC7A7) : _ShuangxiColors.lineSoft,
        ),
      ),
      child: Column(
        children: <Widget>[
          SizedBox(
            height: 36,
            child: Row(
              children: <Widget>[
                const SizedBox(width: 12),
                _ShuangxiSkeletonBlock(
                  width: 14,
                  height: 14,
                  radius: 7,
                  highlight: expanded,
                ),
                const SizedBox(width: 9),
                const Expanded(
                  child: _ShuangxiSkeletonBlock(height: 13, widthFactor: .72),
                ),
                const SizedBox(width: 9),
                const _ShuangxiSkeletonBlock(width: 34, height: 12, radius: 6),
                const SizedBox(width: 12),
                Icon(
                  expanded
                      ? Icons.keyboard_arrow_up_rounded
                      : Icons.keyboard_arrow_down_rounded,
                  color: _ShuangxiColors.muted,
                  size: 20,
                ),
                const SizedBox(width: 6),
              ],
            ),
          ),
          if (expanded)
            Padding(
              padding: const EdgeInsets.fromLTRB(32, 0, 8, 8),
              child: Column(
                children: const <Widget>[
                  _LoadingQuestionNavRow(active: true),
                  _LoadingQuestionNavRow(active: false),
                  _LoadingQuestionNavRow(active: false),
                  _LoadingQuestionNavRow(active: false),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _LoadingQuestionNavRow extends StatelessWidget {
  const _LoadingQuestionNavRow({required this.active});

  final bool active;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 31,
      child: Row(
        children: <Widget>[
          _ShuangxiSkeletonBlock(
            width: 10,
            height: 10,
            radius: 5,
            highlight: active,
          ),
          const SizedBox(width: 9),
          const Expanded(
            child: _ShuangxiSkeletonBlock(height: 12, widthFactor: .78),
          ),
        ],
      ),
    );
  }
}

class _LoadingQuestionWorkspace extends StatelessWidget {
  const _LoadingQuestionWorkspace();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 12),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          SizedBox(
            height: 28,
            child: Row(
              children: <Widget>[
                Container(
                  width: 4,
                  height: 18,
                  decoration: BoxDecoration(
                    color: _ShuangxiColors.orange,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(width: 9),
                const Icon(
                  Icons.account_tree_outlined,
                  color: _ShuangxiColors.orange,
                  size: 18,
                ),
                const SizedBox(width: 7),
                const Expanded(
                  child: _ShuangxiSkeletonBlock(height: 13, widthFactor: .44),
                ),
                const SizedBox(width: 10),
                const _ShuangxiSkeletonBlock(width: 48, height: 12, radius: 6),
              ],
            ),
          ),
          const SizedBox(height: 12),
          const _ShuangxiSkeletonBlock(height: 24, widthFactor: .66),
          const SizedBox(height: 20),
          const _LoadingScoreChoice(active: true),
          const SizedBox(height: 12),
          const _LoadingScoreChoice(active: false),
          const SizedBox(height: 12),
          const _LoadingScoreChoice(active: false),
          const SizedBox(height: 12),
          const _LoadingScoreChoice(active: false),
        ],
      ),
    );
  }
}

class _LoadingScoreChoice extends StatelessWidget {
  const _LoadingScoreChoice({required this.active});

  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 54,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF7EF) : Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: active ? _ShuangxiColors.orange : _ShuangxiColors.line,
          width: active ? 1.3 : 1,
        ),
      ),
      child: Row(
        children: <Widget>[
          _ShuangxiSkeletonBlock(
            width: 20,
            height: 20,
            radius: 10,
            highlight: active,
          ),
          const SizedBox(width: 12),
          const SizedBox(
            width: 42,
            child: _ShuangxiSkeletonBlock(height: 15, widthFactor: .78),
          ),
          const SizedBox(width: 10),
          const Expanded(
            child: _ShuangxiSkeletonBlock(height: 14, widthFactor: .72),
          ),
        ],
      ),
    );
  }
}

class _LoadingRightRail extends StatelessWidget {
  const _LoadingRightRail();

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: <Widget>[
        _LoadingProgressPanel(),
        SizedBox(height: 8),
        _LoadingRemarkPanel(),
        SizedBox(height: 8),
        Expanded(child: _LoadingMissingPanel()),
      ],
    );
  }
}

class _LoadingProgressPanel extends StatelessWidget {
  const _LoadingProgressPanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 148,
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 12),
      decoration: _panelDecoration(radius: 8),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          _PanelTitle(icon: Icons.bar_chart_rounded, title: '测评进度'),
          SizedBox(height: 12),
          Row(
            children: <Widget>[
              _ProgressRing(percent: 0),
              SizedBox(width: 9),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    _SummaryRow(label: '已完成', value: '0 / 0 题'),
                    SizedBox(height: 9),
                    _SummaryRow(label: '缺题', value: '0 题', danger: true),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _LoadingRemarkPanel extends StatelessWidget {
  const _LoadingRemarkPanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 142,
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 10),
      decoration: _panelDecoration(radius: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: const <Widget>[
          _PanelTitle(icon: Icons.assignment_outlined, title: '观察备注'),
          SizedBox(height: 8),
          Expanded(child: _ShuangxiSkeletonBlock(height: 82, radius: 8)),
          SizedBox(height: 5),
          Align(
            alignment: Alignment.bottomRight,
            child: _ShuangxiSkeletonBlock(width: 42, height: 12, radius: 6),
          ),
        ],
      ),
    );
  }
}

class _LoadingMissingPanel extends StatelessWidget {
  const _LoadingMissingPanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 10),
      decoration: _panelDecoration(radius: 8),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          _PanelTitle(icon: Icons.lightbulb_outline_rounded, title: '缺题导航'),
          SizedBox(height: 9),
          _LoadingMissingInfoRow(
            icon: Icons.auto_awesome_rounded,
            label: '自动下一题：已开启',
          ),
          _LoadingMissingInfoRow(
            icon: Icons.grid_view_rounded,
            label: '当前维度：加载中',
          ),
          _LoadingMissingInfoRow(
            icon: Icons.account_tree_outlined,
            label: '当前技能：加载中',
          ),
          _LoadingMissingInfoRow(
            icon: Icons.arrow_forward_rounded,
            label: '第一缺题：加载中',
          ),
        ],
      ),
    );
  }
}

class _LoadingMissingInfoRow extends StatelessWidget {
  const _LoadingMissingInfoRow({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      margin: const EdgeInsets.only(bottom: 6),
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: _ShuangxiColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _ShuangxiColors.orange, size: 17),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              label,
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _ShuangxiColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _LoadingFooterDock extends StatelessWidget {
  const _LoadingFooterDock();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _ShuangxiColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
        boxShadow: _shuangxiShadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _FooterButton(
            label: '上一题',
            icon: Icons.chevron_left_rounded,
            enabled: false,
            onTap: () {},
          ),
          const Spacer(),
          Text.rich(
            const TextSpan(
              children: <InlineSpan>[
                TextSpan(
                  text: '0',
                  style: TextStyle(
                    color: _ShuangxiColors.ink,
                    fontSize: 26,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / 0',
                  style: TextStyle(
                    color: _ShuangxiColors.body,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.arrow_forward_rounded,
            enabled: false,
            filled: true,
            reverseIcon: true,
            onTap: () {},
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.format_list_bulleted_rounded,
            enabled: false,
            onTap: () {},
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _ShuangxiColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 8),
          const Switch(
            value: true,
            activeColor: _ShuangxiColors.orange,
            onChanged: null,
          ),
        ],
      ),
    );
  }
}

class _ShuangxiSkeletonBlock extends StatelessWidget {
  const _ShuangxiSkeletonBlock({
    this.width,
    this.widthFactor,
    required this.height,
    this.radius = 6,
    this.highlight = false,
  });

  final double? width;
  final double? widthFactor;
  final double height;
  final double radius;
  final bool highlight;

  @override
  Widget build(BuildContext context) {
    final Widget block = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: highlight ? const Color(0xFFFFE2D1) : const Color(0xFFF3E3D8),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
    final double? factor = widthFactor;
    if (factor != null) {
      return FractionallySizedBox(
        widthFactor: factor,
        alignment: Alignment.centerLeft,
        child: block,
      );
    }
    return block;
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 360,
        padding: const EdgeInsets.all(18),
        decoration: _panelDecoration(radius: 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(
              Icons.error_outline_rounded,
              color: _ShuangxiColors.orange,
              size: 34,
            ),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _ShuangxiColors.body,
                fontSize: 14,
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 14),
            _TopActionButton(
              label: '重新加载',
              icon: Icons.refresh_rounded,
              filled: true,
              onTap: onRetry,
            ),
          ],
        ),
      ),
    );
  }
}

class _ShuangxiDonutPainter extends CustomPainter {
  const _ShuangxiDonutPainter({required this.percent});

  final int percent;

  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2);
    final double radius = math.min(size.width, size.height) / 2 - 6;
    final Paint track = Paint()
      ..color = const Color(0xFFE9DDD3)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    final Paint progress = Paint()
      ..color = _ShuangxiColors.orange
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    canvas.drawCircle(center, radius, track);
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi / 2,
      math.pi * 2 * percent.clamp(0, 100) / 100,
      false,
      progress,
    );
  }

  @override
  bool shouldRepaint(covariant _ShuangxiDonutPainter oldDelegate) {
    return oldDelegate.percent != percent;
  }
}

IconData _domainIcon(String code, String name) {
  final String target = '${code.toLowerCase()} $name';
  if (target.contains('sensory') || target.contains('感官')) {
    return Icons.visibility_outlined;
  }
  if (target.contains('gross') || target.contains('粗大')) {
    return Icons.directions_run_rounded;
  }
  if (target.contains('fine') || target.contains('精细')) {
    return Icons.back_hand_outlined;
  }
  if (target.contains('self') || target.contains('生活')) {
    return Icons.checkroom_outlined;
  }
  if (target.contains('communication') || target.contains('沟通')) {
    return Icons.chat_bubble_outline_rounded;
  }
  if (target.contains('cognition') || target.contains('认知')) {
    return Icons.psychology_alt_outlined;
  }
  if (target.contains('social') || target.contains('社会')) {
    return Icons.groups_2_outlined;
  }
  return Icons.grid_view_rounded;
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-'
      '${now.month.toString().padLeft(2, '0')}-'
      '${now.day.toString().padLeft(2, '0')}';
}

String _dateOnlyText(String raw) {
  final String value = raw.trim();
  if (value.length >= 10) {
    return value.substring(0, 10);
  }
  return value;
}

String _formatDateTime(String raw) {
  final String value = raw.trim();
  if (value.isEmpty) {
    return '-';
  }
  final DateTime? parsed = DateTime.tryParse(value);
  if (parsed == null) {
    return value.length > 16 ? value.substring(0, 16) : value;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')} '
      '${parsed.hour.toString().padLeft(2, '0')}:'
      '${parsed.minute.toString().padLeft(2, '0')}';
}
