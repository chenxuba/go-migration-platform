import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'autismdev_assessment_client.dart';
import 'pad_top_message.dart';

class AutismDevAssessmentPage extends StatefulWidget {
  const AutismDevAssessmentPage({
    required this.onBack,
    this.args = const AutismDevAssessmentLaunchArgs(),
    this.client = const ApiAutismDevAssessmentClient(),
    super.key,
  });

  final VoidCallback onBack;
  final AutismDevAssessmentLaunchArgs args;
  final AutismDevAssessmentClient client;

  @override
  State<AutismDevAssessmentPage> createState() =>
      _AutismDevAssessmentPageState();
}

class _AutismDevAssessmentPageState extends State<AutismDevAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';

  final Map<int, String> _itemScores = <int, String>{};
  final Map<int, String> _itemRemarks = <int, String>{};
  final Map<int, AutismDevAssessmentItem> _itemDetailCache =
      <int, AutismDevAssessmentItem>{};
  final Map<int, Future<AutismDevAssessmentItem>> _itemDetailFetches =
      <int, Future<AutismDevAssessmentItem>>{};
  final TextEditingController _remarkController = TextEditingController();
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  AutismDevTemplateSummary _template = AutismDevTemplateSummary.empty;
  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _selectedDomainCode = '';
  String _selectedRangeFilter = '';
  int _selectedItemNo = 0;
  int _studentId = 0;
  int _draftId = 0;
  bool _loading = true;
  bool _saving = false;
  bool _submitting = false;
  bool _autoNext = true;
  String _errorMessage = '';
  String _autoSaveText = '等待作答';

  @override
  void initState() {
    super.initState();
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isNotEmpty
        ? _dateOnlyText(widget.args.assessmentDate)
        : _todayIsoDate();
    _examinerName = widget.args.examinerName;
    _draftId = widget.args.draftId;
    _initialize();
  }

  @override
  void dispose() {
    _remarkController.dispose();
    _messageController.dispose();
    super.dispose();
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
      final AutismDevTemplateSummary template =
          await widget.client.fetchTemplateSummary(token);
      if (!mounted) {
        return;
      }
      _token = token;
      _template = template;
      _selectedDomainCode = template.domainGroups.isNotEmpty
          ? template.domainGroups.first.domainCode
          : '';
      _selectedItemNo = _firstItemNoInDomain(_selectedDomainCode);
      if (_draftId > 0) {
        final AutismDevDraftDetail detail =
            await widget.client.fetchDraftDetail(token, _draftId);
        _applyDraftDetail(detail);
      } else {
        final AssessmentDraftSummary? latestDraft =
            await _findLatestDraft(token);
        if (latestDraft != null) {
          final AutismDevDraftDetail detail =
              await widget.client.fetchDraftDetail(token, latestDraft.id);
          _applyDraftDetail(detail);
        }
      }
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstUnansweredItemNo() > 0
            ? _firstUnansweredItemNo()
            : _firstItemNoInDomain(_selectedDomainCode);
      }
      _selectedRangeFilter = '';
      _syncRemarkController();
      setState(() {
        _loading = false;
        _autoSaveText = _draftId > 0 ? '已载入草稿' : '已准备';
      });
      _prefetchSelectedItem();
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
        _errorMessage = '孤独症发展评估表加载失败：$error';
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

  void _applyDraftDetail(AutismDevDraftDetail detail) {
    _draftId = detail.id;
    _studentId = detail.studentId > 0 ? detail.studentId : _studentId;
    _studentName = detail.studentName.trim().isNotEmpty
        ? detail.studentName.trim()
        : _studentName;
    _birthDate = detail.birthDate.trim().isNotEmpty
        ? detail.birthDate.trim()
        : _birthDate;
    _assessmentDate = detail.assessmentDate.trim().isNotEmpty
        ? detail.assessmentDate.trim()
        : _assessmentDate;
    _examinerName = detail.examinerName.trim().isNotEmpty
        ? detail.examinerName.trim()
        : _examinerName;
    _itemScores
      ..clear()
      ..addAll(detail.input.itemScores);
    _itemRemarks
      ..clear()
      ..addAll(detail.input.itemRemarks);
    final int firstUnanswered = _firstUnansweredItemNo();
    if (firstUnanswered > 0) {
      _selectedItemNo = firstUnanswered;
      _selectedDomainCode =
          _summaryByNo(firstUnanswered)?.domainCode ?? _selectedDomainCode;
    }
  }

  AutismDevDomainGroup? get _selectedGroup {
    for (final AutismDevDomainGroup group in _template.domainGroups) {
      if (group.domainCode == _selectedDomainCode) {
        return group;
      }
    }
    return _template.domainGroups.isNotEmpty
        ? _template.domainGroups.first
        : null;
  }

  AutismDevItemSummary? get _selectedSummary => _summaryByNo(_selectedItemNo);

  AutismDevAssessmentItem? get _selectedDetail =>
      _itemDetailCache[_selectedItemNo];

  int get _answeredCount => _itemScores.length;

  int get _missingCount => math.max(_totalCount - _answeredCount, 0);

  int get _totalCount {
    if (_template.itemCount > 0) {
      return _template.itemCount;
    }
    return _template.domainGroups.fold<int>(
      0,
      (int total, AutismDevDomainGroup group) => total + group.items.length,
    );
  }

  int get _currentIndex {
    final int index = _allItems.indexWhere(
      (AutismDevItemSummary item) => item.itemNo == _selectedItemNo,
    );
    return index < 0 ? 0 : index;
  }

  bool get _hasPreviousItem => _currentIndex > 0;

  bool get _hasNextItem => _currentIndex < _allItems.length - 1;

  List<AutismDevScoreOption> get _currentScoreOptions {
    final AutismDevAssessmentItem? detail = _selectedDetail;
    if (detail != null && detail.scoreOptions.isNotEmpty) {
      return detail.scoreOptions;
    }
    final String scoreType =
        detail?.scoreType ?? _selectedSummary?.scoreType ?? '';
    return _template.scoreOptions
        .where((AutismDevScoreOption option) =>
            option.scoreType.toUpperCase() == scoreType.toUpperCase())
        .toList();
  }

  void _selectItem(AutismDevItemSummary item) {
    setState(() {
      if (_selectedRangeFilter.isNotEmpty &&
          _assessmentRangeBucket(item) != _selectedRangeFilter) {
        _selectedRangeFilter = '';
      }
      _selectedItemNo = item.itemNo;
      _selectedDomainCode = item.domainCode;
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  void _selectRangeFilter(String rangeFilter) {
    final AutismDevDomainGroup? group = _selectedGroup;
    if (group == null) {
      return;
    }
    final String nextFilter = rangeFilter.trim();
    if (nextFilter.isEmpty) {
      setState(() => _selectedRangeFilter = '');
      return;
    }
    final List<AutismDevItemSummary> filteredItems = group.items
        .where((AutismDevItemSummary item) =>
            _assessmentRangeBucket(item) == nextFilter)
        .toList(growable: false);
    if (filteredItems.isEmpty) {
      setState(() => _selectedRangeFilter = nextFilter);
      return;
    }
    final AutismDevItemSummary target = filteredItems.firstWhere(
      (AutismDevItemSummary item) => !_itemScores.containsKey(item.itemNo),
      orElse: () => filteredItems.first,
    );
    setState(() {
      _selectedRangeFilter = nextFilter;
      _selectedItemNo = target.itemNo;
      _selectedDomainCode = target.domainCode;
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  void _syncRemarkController() {
    final String remark = _itemRemarks[_selectedItemNo] ?? '';
    if (_remarkController.text != remark) {
      _remarkController.text = remark;
    }
  }

  void _syncCurrentRemarkToState() {
    final int itemNo = _selectedItemNo;
    if (itemNo <= 0) {
      return;
    }
    final String remark = _remarkController.text.trim();
    if (remark.isNotEmpty || _itemScores.containsKey(itemNo)) {
      _itemRemarks[itemNo] = remark;
    }
  }

  void _prefetchSelectedItem() {
    final int itemNo = _selectedItemNo;
    if (itemNo <= 0 || _itemDetailCache.containsKey(itemNo)) {
      return;
    }
    final Future<AutismDevAssessmentItem> future = _itemDetailFetches[itemNo] ??
        widget.client.fetchTemplateItem(_token, itemNo: itemNo);
    _itemDetailFetches[itemNo] = future;
    unawaited(
      future.then<void>((AutismDevAssessmentItem item) {
        if (!mounted) {
          return;
        }
        setState(() => _itemDetailCache[itemNo] = item);
      }, onError: (Object _) {}),
    );
  }

  Future<void> _selectScore(String score, {bool moveNext = false}) async {
    if (_selectedItemNo <= 0) {
      return;
    }
    setState(() {
      _itemScores[_selectedItemNo] = score;
      _itemRemarks[_selectedItemNo] = _remarkController.text.trim();
    });
    await _saveCurrentItem(moveNext: moveNext, silent: true);
  }

  Future<void> _saveCurrentItem({
    bool moveNext = false,
    bool silent = false,
  }) async {
    final int itemNo = _selectedItemNo;
    final String? score = _itemScores[itemNo];
    if (itemNo <= 0 || score == null || score.trim().isEmpty) {
      _showMessage('请先选择本题评分');
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存');
      return;
    }
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
      _itemRemarks[itemNo] = _remarkController.text.trim();
    });
    try {
      final AutismDevDraftDetail detail;
      if (_draftId <= 0) {
        detail = await widget.client.saveDraft(_token, _draftPayload());
      } else {
        detail = await widget.client.saveDraftItem(
          _token,
          <String, dynamic>{
            'draftId': _draftId,
            'itemNo': itemNo,
            'score': score,
            'remark': _remarkController.text.trim(),
          },
        );
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _itemScores
          ..clear()
          ..addAll(detail.input.itemScores);
        _itemRemarks
          ..clear()
          ..addAll(detail.input.itemRemarks);
        _saving = false;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      _syncRemarkController();
      if (moveNext) {
        _goNextItem();
      } else if (!silent) {
        _showMessage('已保存本题', tone: PadMessageTone.success);
      }
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage('保存失败：$error');
    }
  }

  Future<void> _saveDraft() async {
    if (_saving) {
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存');
      return;
    }
    _syncCurrentRemarkToState();
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
    });
    try {
      final AutismDevDraftDetail detail =
          await widget.client.saveDraft(_token, _draftPayload());
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _itemScores
          ..clear()
          ..addAll(detail.input.itemScores);
        _itemRemarks
          ..clear()
          ..addAll(detail.input.itemRemarks);
        _saving = false;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      _syncRemarkController();
      _showMessage('草稿已保存', tone: PadMessageTone.success);
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage('保存失败：$error');
    }
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    _syncCurrentRemarkToState();
    if (_totalCount > 0 && _answeredCount < _totalCount) {
      _showMessage('还有 ${_totalCount - _answeredCount} 道题未评分，完成后再提交');
      return;
    }
    setState(() => _submitting = true);
    try {
      if (_draftId <= 0) {
        final AutismDevDraftDetail detail =
            await widget.client.saveDraft(_token, _draftPayload());
        _draftId = detail.id;
      }
      await widget.client.submitDraft(_token, _draftId);
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('正式记录已提交', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 500));
      if (mounted) {
        widget.onBack();
      }
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('提交失败：$error');
    }
  }

  Map<String, dynamic> _draftPayload() {
    return <String, dynamic>{
      if (_draftId > 0) 'id': _draftId,
      if (_studentId > 0) 'studentId': _studentId,
      'studentName': _studentName.trim(),
      'examinerName': _examinerName.trim(),
      'birthDate': _birthDate.trim(),
      'assessmentDate': _assessmentDate.trim(),
      'itemScoreList': _itemScoreList(),
    };
  }

  List<Map<String, dynamic>> _itemScoreList() {
    final List<int> itemNos = _itemScores.keys.toList()..sort();
    return itemNos.map((int itemNo) {
      return <String, dynamic>{
        'itemNo': itemNo,
        'score': _itemScores[itemNo],
        if ((_itemRemarks[itemNo] ?? '').trim().isNotEmpty)
          'remark': _itemRemarks[itemNo]!.trim(),
      };
    }).toList();
  }

  void _goNextItem() {
    final List<AutismDevItemSummary> allItems = _allItems;
    final int currentIndex = allItems.indexWhere(
        (AutismDevItemSummary item) => item.itemNo == _selectedItemNo);
    if (currentIndex >= 0 && currentIndex < allItems.length - 1) {
      final AutismDevItemSummary next = allItems[currentIndex + 1];
      _selectItem(next);
    }
  }

  void _goPreviousItem() {
    final List<AutismDevItemSummary> allItems = _allItems;
    final int currentIndex = allItems.indexWhere(
        (AutismDevItemSummary item) => item.itemNo == _selectedItemNo);
    if (currentIndex > 0) {
      _selectItem(allItems[currentIndex - 1]);
    }
  }

  void _jumpToMissing() {
    final int itemNo = _firstUnansweredItemNo();
    if (itemNo <= 0) {
      _showMessage('当前没有缺题', tone: PadMessageTone.success);
      return;
    }
    final AutismDevItemSummary? item = _summaryByNo(itemNo);
    if (item != null) {
      _selectItem(item);
    }
  }

  List<AutismDevItemSummary> get _allItems {
    return _template.domainGroups
        .expand((AutismDevDomainGroup group) => group.items)
        .toList();
  }

  int _firstItemNoInDomain(String domainCode) {
    for (final AutismDevDomainGroup group in _template.domainGroups) {
      if (group.domainCode == domainCode && group.items.isNotEmpty) {
        return group.items.first.itemNo;
      }
    }
    return _template.domainGroups.isNotEmpty &&
            _template.domainGroups.first.items.isNotEmpty
        ? _template.domainGroups.first.items.first.itemNo
        : 0;
  }

  int _firstUnansweredItemNo() {
    for (final AutismDevItemSummary item in _allItems) {
      if (!_itemScores.containsKey(item.itemNo)) {
        return item.itemNo;
      }
    }
    return 0;
  }

  AutismDevItemSummary? _summaryByNo(int itemNo) {
    for (final AutismDevItemSummary item in _allItems) {
      if (item.itemNo == itemNo) {
        return item;
      }
    }
    return null;
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
      key: 'autismdev-top-message',
    );
  }

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _AutismDevColors.page,
      child: Column(
        children: <Widget>[
          _AutismDevTopBar(
            title: widget.args.scaleName.trim().isEmpty
                ? '孤独症儿童发展评估'
                : widget.args.scaleName.trim(),
            studentName: _studentName.trim().isEmpty ? '-' : _studentName,
            studentAge: _studentAge.trim().isEmpty ? '未知' : _studentAge,
            assessmentDate: _assessmentDate.isEmpty ? '未设置日期' : _assessmentDate,
            examinerName: _examinerName.trim().isEmpty ? '当前老师' : _examinerName,
            autoSaveText: _autoSaveText,
            saving: _saving,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: _saveDraft,
            onSubmit: _submitDraft,
          ),
          Expanded(child: _buildBody()),
        ],
      ),
    );
  }

  Widget _buildBody() {
    if (_loading) {
      return const _AutismDevLoadingBody();
    }
    if (_errorMessage.isNotEmpty) {
      return _AutismDevStateBody(
        title: '加载失败',
        message: _errorMessage,
        actionText: '重新加载',
        onAction: _initialize,
      );
    }
    final AutismDevDomainGroup? group = _selectedGroup;
    final AutismDevItemSummary? summary = _selectedSummary;
    if (group == null || summary == null) {
      return _AutismDevStateBody(
        title: '暂无题目',
        message: '当前量表模板没有可用题目',
        actionText: '返回',
        onAction: widget.onBack,
      );
    }
    return Column(
      children: <Widget>[
        Expanded(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                SizedBox(
                  width: 226,
                  child: _AutismDevDomainPanel(
                    groups: _template.domainGroups,
                    selectedDomainCode: _selectedDomainCode,
                    selectedRangeFilter: _selectedRangeFilter,
                    selectedItemNo: _selectedItemNo,
                    itemScores: _itemScores,
                    onSelectItem: _selectItem,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _AutismDevWorkspacePanel(
                    group: group,
                    item: summary,
                    detail: _selectedDetail,
                    selectedScore: _itemScores[_selectedItemNo],
                    scoreOptions: _currentScoreOptions,
                    onScore: (String score) =>
                        _selectScore(score, moveNext: _autoNext),
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 238,
                  child: _AutismDevRightRail(
                    group: group,
                    item: summary,
                    remarkController: _remarkController,
                    selectedRangeFilter: _selectedRangeFilter,
                    itemScores: _itemScores,
                    answeredCount: _answeredCount,
                    totalCount: _totalCount,
                    missingCount: _missingCount,
                    onSelectRangeFilter: _selectRangeFilter,
                  ),
                ),
              ],
            ),
          ),
        ),
        _AutismDevFooter(
          current: _currentIndex + 1,
          total: _totalCount,
          hasPrevious: _hasPreviousItem,
          hasNext: _hasNextItem,
          autoNext: _autoNext,
          onPrevious: _goPreviousItem,
          onNext: _goNextItem,
          onJumpMissing: _jumpToMissing,
          onToggleAutoNext: (bool value) => setState(() => _autoNext = value),
        ),
      ],
    );
  }
}

class _AutismDevTopBar extends StatelessWidget {
  const _AutismDevTopBar({
    required this.title,
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

  final String title;
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
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _autismDevShadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1120;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 246 : 308,
                child: Text(
                  '$title 测评工作台',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 23,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _HeaderMeta(label: '儿童', value: studentName),
                    ),
                    Expanded(
                      child: _HeaderMeta(label: '年龄', value: studentAge),
                    ),
                    Expanded(
                      flex: 2,
                      child: _HeaderMeta(label: '测评日期', value: assessmentDate),
                    ),
                    Expanded(
                      child: _HeaderMeta(label: '施测者', value: examinerName),
                    ),
                  ],
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                SizedBox(
                  width: compact ? 82 : 112,
                  child: Text(
                    autoSaveText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: const TextStyle(
                      color: _AutismDevColors.muted,
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
      margin: const EdgeInsets.only(left: 10),
      padding: const EdgeInsets.only(left: 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _AutismDevColors.line)),
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
          style: const TextStyle(
            color: _AutismDevColors.body,
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
            border: Border.all(color: _AutismDevColors.line),
          ),
          child: Icon(
            icon,
            color: _AutismDevColors.body,
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
            color: filled ? _AutismDevColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _AutismDevColors.orange),
            boxShadow: filled
                ? _autismDevShadow(
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
                    color: filled ? Colors.white : _AutismDevColors.orange,
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : _AutismDevColors.orange,
                ),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _AutismDevColors.orangeDeep,
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

class _AutismDevDomainPanel extends StatefulWidget {
  const _AutismDevDomainPanel({
    required this.groups,
    required this.selectedDomainCode,
    required this.selectedRangeFilter,
    required this.selectedItemNo,
    required this.itemScores,
    required this.onSelectItem,
  });

  final List<AutismDevDomainGroup> groups;
  final String selectedDomainCode;
  final String selectedRangeFilter;
  final int selectedItemNo;
  final Map<int, String> itemScores;
  final ValueChanged<AutismDevItemSummary> onSelectItem;

  @override
  State<_AutismDevDomainPanel> createState() => _AutismDevDomainPanelState();
}

class _AutismDevDomainPanelState extends State<_AutismDevDomainPanel> {
  final ScrollController _scrollController = ScrollController();
  final GlobalKey _activeNavItemKey = GlobalKey();
  final Map<String, GlobalKey> _domainHeaderKeys = <String, GlobalKey>{};
  String _expandedDomainCode = '';

  @override
  void initState() {
    super.initState();
    _syncExpandedDomain();
    _keepActiveItemVisible();
  }

  @override
  void didUpdateWidget(covariant _AutismDevDomainPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    _domainHeaderKeys
        .removeWhere((String code, GlobalKey key) => !_hasDomainCode(code));
    final bool selectedDomainChanged =
        oldWidget.selectedDomainCode != widget.selectedDomainCode;
    final bool selectedItemChanged =
        oldWidget.selectedItemNo != widget.selectedItemNo;
    if (selectedDomainChanged || !_hasDomainCode(_expandedDomainCode)) {
      _syncExpandedDomain();
    }
    if (selectedDomainChanged || selectedItemChanged) {
      _keepActiveItemVisible();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  bool _hasDomainCode(String domainCode) {
    return widget.groups.any(
      (AutismDevDomainGroup group) => group.domainCode == domainCode,
    );
  }

  void _syncExpandedDomain() {
    _expandedDomainCode = widget.selectedDomainCode.trim();
  }

  void _toggleDomain(AutismDevDomainGroup group) {
    final String domainCode = group.domainCode;
    final bool opening = _expandedDomainCode != domainCode;
    setState(() {
      _expandedDomainCode = opening ? domainCode : '';
    });
    if (opening) {
      _keepDomainHeaderVisible(domainCode);
    }
  }

  void _keepDomainHeaderVisible(String domainCode) {
    WidgetsBinding.instance.addPostFrameCallback((Duration _) {
      if (!mounted ||
          !_scrollController.hasClients ||
          domainCode.trim().isEmpty) {
        return;
      }
      final BuildContext? headerContext =
          _domainHeaderKeys[domainCode]?.currentContext;
      if (headerContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        headerContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .02,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _keepActiveItemVisible() {
    WidgetsBinding.instance.addPostFrameCallback((Duration _) {
      if (!mounted ||
          !_scrollController.hasClients ||
          widget.selectedItemNo <= 0) {
        return;
      }
      final BuildContext? itemContext = _activeNavItemKey.currentContext;
      if (itemContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        itemContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .34,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  List<AutismDevItemSummary> _visibleItemsForGroup(
    AutismDevDomainGroup group,
  ) {
    final String rangeFilter = _rangeFilterForGroup(group);
    if (rangeFilter.isEmpty) {
      return group.items;
    }
    return group.items
        .where((AutismDevItemSummary item) =>
            _assessmentRangeBucket(item) == rangeFilter)
        .toList(growable: false);
  }

  String _rangeFilterForGroup(AutismDevDomainGroup group) {
    if (group.domainCode == widget.selectedDomainCode &&
        widget.selectedRangeFilter.trim().isNotEmpty) {
      return widget.selectedRangeFilter.trim();
    }
    return '';
  }

  Widget _buildDomainTile(AutismDevDomainGroup group) {
    final int done = group.items
        .where((AutismDevItemSummary item) =>
            widget.itemScores.containsKey(item.itemNo))
        .length;
    final List<AutismDevItemSummary> visibleItems =
        _visibleItemsForGroup(group);
    return _AutismDevDomainTile(
      group: group,
      headerKey: _domainHeaderKeys.putIfAbsent(
        group.domainCode,
        () => GlobalKey(
          debugLabel: 'autismdev-domain-header-${group.domainCode}',
        ),
      ),
      visibleItems: visibleItems,
      done: done,
      expanded: _expandedDomainCode == group.domainCode,
      selected: group.domainCode == widget.selectedDomainCode,
      itemScores: widget.itemScores,
      selectedItemNo: widget.selectedItemNo,
      activeItemKey: _activeNavItemKey,
      onTap: () => _toggleDomain(group),
      onTapItem: widget.onSelectItem,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          const _AutismDevSidebarHeader(),
          Expanded(
            child: SingleChildScrollView(
              controller: _scrollController,
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: <Widget>[
                  for (final AutismDevDomainGroup group in widget.groups)
                    _buildDomainTile(group),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevDomainTile extends StatelessWidget {
  const _AutismDevDomainTile({
    required this.group,
    required this.headerKey,
    required this.visibleItems,
    required this.done,
    required this.expanded,
    required this.selected,
    required this.itemScores,
    required this.selectedItemNo,
    required this.activeItemKey,
    required this.onTap,
    required this.onTapItem,
  });

  final AutismDevDomainGroup group;
  final Key headerKey;
  final List<AutismDevItemSummary> visibleItems;
  final int done;
  final bool expanded;
  final bool selected;
  final Map<int, String> itemScores;
  final int selectedItemNo;
  final GlobalKey activeItemKey;
  final VoidCallback onTap;
  final ValueChanged<AutismDevItemSummary> onTapItem;

  @override
  Widget build(BuildContext context) {
    final Color color = _domainColor(group.domainCode);
    final double progress = group.itemCount <= 0
        ? 0
        : (done / group.itemCount).clamp(0, 1).toDouble();
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 9),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFFBF8) : Colors.transparent,
        border: const Border(
          bottom: BorderSide(color: _AutismDevColors.lineSoft),
        ),
      ),
      child: Column(
        children: <Widget>[
          Column(
            key: headerKey,
            children: <Widget>[
              InkWell(
                onTap: onTap,
                borderRadius: BorderRadius.circular(8),
                child: SizedBox(
                  height: 28,
                  child: Row(
                    children: <Widget>[
                      Icon(
                        expanded
                            ? Icons.keyboard_arrow_down_rounded
                            : Icons.chevron_right_rounded,
                        size: 20,
                        color: _AutismDevColors.ink,
                      ),
                      const SizedBox(width: 4),
                      Container(
                        width: 9,
                        height: 9,
                        decoration:
                            BoxDecoration(color: color, shape: BoxShape.circle),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          group.domainName,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: selected
                                ? _AutismDevColors.orangeDeep
                                : _AutismDevColors.ink,
                            fontSize: 13,
                            height: 1,
                            fontWeight:
                                selected ? FontWeight.w900 : FontWeight.w800,
                          ),
                        ),
                      ),
                      Text(
                        '$done/${group.itemCount}',
                        style: const TextStyle(
                          color: _AutismDevColors.muted,
                          fontSize: 12,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 7),
              Row(
                children: <Widget>[
                  const SizedBox(width: 28),
                  Expanded(
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(99),
                      child: LinearProgressIndicator(
                        value: progress,
                        minHeight: 5,
                        color: _AutismDevColors.orange,
                        backgroundColor: const Color(0xFFF2E6DC),
                      ),
                    ),
                  ),
                  const SizedBox(width: 9),
                  SizedBox(
                    width: 34,
                    child: Text(
                      '${(progress * 100).round()}%',
                      textAlign: TextAlign.right,
                      style: const TextStyle(
                        color: _AutismDevColors.muted,
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
          if (expanded) ...<Widget>[
            const SizedBox(height: 7),
            for (final AutismDevItemSummary item in visibleItems)
              _AutismDevQuestionNavItem(
                key: item.itemNo == selectedItemNo
                    ? activeItemKey
                    : ValueKey<int>(item.itemNo),
                item: item,
                active: item.itemNo == selectedItemNo,
                done: itemScores.containsKey(item.itemNo),
                onTap: () => onTapItem(item),
              ),
          ],
        ],
      ),
    );
  }
}

class _AutismDevRangeQuickFilter extends StatelessWidget {
  const _AutismDevRangeQuickFilter({
    required this.group,
    required this.selectedRangeFilter,
    required this.itemScores,
    required this.onSelectRangeFilter,
  });

  final AutismDevDomainGroup? group;
  final String selectedRangeFilter;
  final Map<int, String> itemScores;
  final ValueChanged<String> onSelectRangeFilter;

  @override
  Widget build(BuildContext context) {
    final List<_AutismDevRangeOption> options =
        _rangeOptionsForGroup(group, itemScores);
    final int total = group?.itemCount ?? 0;
    final int done = group?.items
            .where((AutismDevItemSummary item) =>
                itemScores.containsKey(item.itemNo))
            .length ??
        0;
    final String title = (group?.domainName ?? '').trim();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                title.isEmpty ? '分类' : title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '${options.length}类',
              style: const TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Expanded(
          child: options.isEmpty
              ? const Center(
                  child: Text(
                    '暂无分类',
                    style: TextStyle(
                      color: _AutismDevColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                )
              : ClipRect(
                  child: Stack(
                    children: <Widget>[
                      Scrollbar(
                        child: ListView.separated(
                          padding: const EdgeInsets.only(bottom: 26),
                          physics: const BouncingScrollPhysics(),
                          itemBuilder: (BuildContext context, int index) {
                            if (index == 0) {
                              return _AutismDevRangeListItem(
                                label: '全部',
                                total: total,
                                done: done,
                                active: selectedRangeFilter.trim().isEmpty,
                                onTap: () => onSelectRangeFilter(''),
                              );
                            }
                            final _AutismDevRangeOption option =
                                options[index - 1];
                            return _AutismDevRangeListItem(
                              label: option.label,
                              total: option.total,
                              done: option.done,
                              active:
                                  selectedRangeFilter.trim() == option.label,
                              onTap: () => onSelectRangeFilter(option.label),
                            );
                          },
                          separatorBuilder: (BuildContext context, int index) =>
                              const SizedBox(height: 7),
                          itemCount: options.length + 1,
                        ),
                      ),
                      const Positioned(
                        left: 0,
                        right: 0,
                        bottom: 0,
                        child: IgnorePointer(
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              gradient: LinearGradient(
                                begin: Alignment.topCenter,
                                end: Alignment.bottomCenter,
                                colors: <Color>[
                                  Color(0x00FFFFFF),
                                  Color(0xF6FFFFFF),
                                ],
                              ),
                            ),
                            child: SizedBox(height: 30),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
        ),
      ],
    );
  }
}

class _AutismDevRangeListItem extends StatelessWidget {
  const _AutismDevRangeListItem({
    required this.label,
    required this.total,
    required this.done,
    required this.active,
    required this.onTap,
  });

  final String label;
  final int total;
  final int done;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          height: 39,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color:
                active ? _AutismDevColors.orange : _AutismDevColors.softPanel,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: active ? _AutismDevColors.orange : _AutismDevColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 7,
                height: 7,
                decoration: BoxDecoration(
                  color: active ? Colors.white : _AutismDevColors.orange,
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: active ? Colors.white : _AutismDevColors.ink,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                height: 22,
                padding: const EdgeInsets.symmetric(horizontal: 7),
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: active ? Colors.white.withOpacity(.16) : Colors.white,
                  borderRadius: BorderRadius.circular(11),
                  border: Border.all(
                    color: active
                        ? Colors.white.withOpacity(.26)
                        : _AutismDevColors.lineSoft,
                  ),
                ),
                child: Text(
                  '$done/$total',
                  style: TextStyle(
                    color: active ? Colors.white : _AutismDevColors.body,
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

class _AutismDevRangeOption {
  _AutismDevRangeOption(this.label);

  final String label;
  int total = 0;
  int done = 0;
}

class _AutismDevSidebarHeader extends StatelessWidget {
  const _AutismDevSidebarHeader();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _AutismDevColors.lineSoft)),
      ),
      child: const Row(
        children: <Widget>[
          Text(
            '领域任务',
            style: TextStyle(
              color: _AutismDevColors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          Spacer(),
          Icon(Icons.tune_rounded, size: 18, color: _AutismDevColors.muted),
        ],
      ),
    );
  }
}

class _AutismDevQuestionNavItem extends StatelessWidget {
  const _AutismDevQuestionNavItem({
    required this.item,
    required this.active,
    required this.done,
    required this.onTap,
    super.key,
  });

  final AutismDevItemSummary item;
  final bool active;
  final bool done;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 18, top: 3),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(10),
          child: Ink(
            height: 31,
            padding: const EdgeInsets.symmetric(horizontal: 9),
            decoration: BoxDecoration(
              color: active ? const Color(0xFFFFEEE5) : Colors.transparent,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: <Widget>[
                SizedBox(
                  width: 42,
                  child: Text(
                    '第 ${item.itemNo}',
                    maxLines: 1,
                    style: TextStyle(
                      color: active
                          ? _AutismDevColors.orangeDeep
                          : _AutismDevColors.body,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  child: Text(
                    _displayItemTitle(item),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active
                          ? _AutismDevColors.orangeDeep
                          : _AutismDevColors.body,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                const SizedBox(width: 7),
                if (done)
                  const Icon(
                    Icons.check_circle_rounded,
                    size: 17,
                    color: _AutismDevColors.green,
                  )
                else
                  Container(
                    width: 15,
                    height: 15,
                    decoration: BoxDecoration(
                      color: active ? _AutismDevColors.orange : Colors.white,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: active
                            ? _AutismDevColors.orange
                            : const Color(0xFFCAB8AA),
                      ),
                    ),
                    child: active
                        ? const Center(
                            child: DecoratedBox(
                              decoration: BoxDecoration(
                                color: Colors.white,
                                shape: BoxShape.circle,
                              ),
                              child: SizedBox(width: 5, height: 5),
                            ),
                          )
                        : null,
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _AutismDevWorkspacePanel extends StatelessWidget {
  const _AutismDevWorkspacePanel({
    required this.group,
    required this.item,
    required this.detail,
    required this.selectedScore,
    required this.scoreOptions,
    required this.onScore,
  });

  final AutismDevDomainGroup group;
  final AutismDevItemSummary item;
  final AutismDevAssessmentItem? detail;
  final String? selectedScore;
  final List<AutismDevScoreOption> scoreOptions;
  final ValueChanged<String> onScore;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            SizedBox(
              height: 38,
              child: Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      '第 ${item.itemNo} 项  ${_displayItemTitle(item)}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _AutismDevColors.ink,
                        fontSize: 23,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  _DomainChip(
                    code: group.domainCode,
                    name: group.domainName,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                physics: const BouncingScrollPhysics(),
                children: <Widget>[
                  _AutismDevItemDetailCard(item: item, detail: detail),
                ],
              ),
            ),
            const SizedBox(height: 10),
            _AutismDevScoreDock(
              scoreOptions: scoreOptions,
              selectedScore: selectedScore,
              onScore: onScore,
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevRightRail extends StatelessWidget {
  const _AutismDevRightRail({
    required this.group,
    required this.item,
    required this.remarkController,
    required this.selectedRangeFilter,
    required this.itemScores,
    required this.answeredCount,
    required this.totalCount,
    required this.missingCount,
    required this.onSelectRangeFilter,
  });

  final AutismDevDomainGroup group;
  final AutismDevItemSummary item;
  final TextEditingController remarkController;
  final String selectedRangeFilter;
  final Map<int, String> itemScores;
  final int answeredCount;
  final int totalCount;
  final int missingCount;
  final ValueChanged<String> onSelectRangeFilter;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            _ProgressSummary(
              answeredCount: answeredCount,
              totalCount: totalCount,
              missing: missingCount,
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _AutismDevRangeQuickFilter(
                group: group,
                selectedRangeFilter: selectedRangeFilter,
                itemScores: itemScores,
                onSelectRangeFilter: onSelectRangeFilter,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '备注',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            SizedBox(
              height: 66,
              child: TextField(
                controller: remarkController,
                minLines: 2,
                maxLines: 2,
                onTapOutside: (_) {
                  FocusManager.instance.primaryFocus?.unfocus();
                },
                onEditingComplete: () {
                  FocusManager.instance.primaryFocus?.unfocus();
                },
                textAlignVertical: TextAlignVertical.top,
                style: const TextStyle(
                  color: _AutismDevColors.body,
                  fontSize: 15,
                  height: 1.35,
                  fontWeight: FontWeight.w600,
                ),
                decoration: InputDecoration(
                  hintText: item.scoreType.toUpperCase() == 'AMS'
                      ? '可记录情境、反应强度或持续时间'
                      : '可记录提示层级、辅助方式或备注',
                  hintStyle: const TextStyle(
                    color: _AutismDevColors.muted,
                    fontSize: 12,
                    height: 1.25,
                    fontWeight: FontWeight.w600,
                  ),
                  filled: true,
                  fillColor: _AutismDevColors.softPanel,
                  contentPadding: const EdgeInsets.fromLTRB(12, 9, 12, 9),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _AutismDevColors.line),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide:
                        const BorderSide(color: _AutismDevColors.orange),
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

class _AutismDevScoreDock extends StatelessWidget {
  const _AutismDevScoreDock({
    required this.scoreOptions,
    required this.selectedScore,
    required this.onScore,
  });

  final List<AutismDevScoreOption> scoreOptions;
  final String? selectedScore;
  final ValueChanged<String> onScore;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        const Row(
          children: <Widget>[
            Text(
              '评分',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            Spacer(),
          ],
        ),
        const SizedBox(height: 8),
        LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            const double spacing = 12;
            final bool singleColumn = constraints.maxWidth < 460;
            final double cardWidth = scoreOptions.length == 1 || singleColumn
                ? constraints.maxWidth
                : (constraints.maxWidth - spacing) / 2;
            return Wrap(
              spacing: spacing,
              runSpacing: 8,
              children: <Widget>[
                for (final AutismDevScoreOption option in scoreOptions)
                  SizedBox(
                    width: cardWidth,
                    child: _ScoreOptionButton(
                      option: option,
                      selected: selectedScore == option.value,
                      onTap: () => onScore(option.value),
                    ),
                  ),
              ],
            );
          },
        ),
      ],
    );
  }
}

class _ScoreOptionButton extends StatelessWidget {
  const _ScoreOptionButton({
    required this.option,
    required this.selected,
    required this.onTap,
  });

  final AutismDevScoreOption option;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(option.value);
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(10),
      child: InkWell(
        borderRadius: BorderRadius.circular(10),
        onTap: onTap,
        child: Ink(
          height: 78,
          padding: const EdgeInsets.fromLTRB(13, 9, 11, 9),
          decoration: BoxDecoration(
            color: selected ? color.withOpacity(.08) : Colors.white,
            border: Border.all(
              color: selected ? color : _AutismDevColors.line,
              width: selected ? 1.5 : 1,
            ),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 36,
                height: 36,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: color.withOpacity(selected ? .14 : .08),
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: color.withOpacity(.38)),
                ),
                child: Text(
                  option.value,
                  style: TextStyle(
                    color: color,
                    fontSize: 21,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      _optionTitle(option),
                      maxLines: 2,
                      style: TextStyle(
                        color: _AutismDevColors.ink,
                        fontSize: 14,
                        height: 1.12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      option.description,
                      maxLines: 2,
                      style: TextStyle(
                        color: _AutismDevColors.muted,
                        fontSize: 10.5,
                        height: 1.12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Icon(
                selected
                    ? Icons.check_circle_rounded
                    : Icons.radio_button_unchecked_rounded,
                color: selected ? color : const Color(0xFFCAB8AA),
                size: 20,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.answeredCount,
    required this.totalCount,
    required this.missing,
  });

  final int answeredCount;
  final int totalCount;
  final int missing;

  @override
  Widget build(BuildContext context) {
    final double percent = totalCount <= 0 ? 0 : answeredCount / totalCount;
    final double normalizedPercent = percent.clamp(0, 1).toDouble();
    final int percentText = (normalizedPercent * 100).round();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '当前进度',
          style: TextStyle(
            color: _AutismDevColors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Row(
          children: <Widget>[
            SizedBox(
              width: 82,
              height: 82,
              child: CustomPaint(
                painter: _AutismDevDonutPainter(percent: percentText),
                child: Center(
                  child: Text(
                    '$percentText%',
                    style: const TextStyle(
                      color: _AutismDevColors.ink,
                      fontSize: 21,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _ProgressText(
                    label: '已完成',
                    value: '$answeredCount / $totalCount 项',
                  ),
                  const SizedBox(height: 10),
                  _ProgressText(
                    label: '缺题',
                    value: '$missing 项',
                    danger: true,
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _ProgressText extends StatelessWidget {
  const _ProgressText({
    required this.label,
    required this.value,
    this.danger = false,
  });

  final String label;
  final String value;
  final bool danger;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _AutismDevColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 5),
        Text(
          value,
          style: TextStyle(
            color:
                danger ? const Color(0xFFE04438) : _AutismDevColors.orangeDeep,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _AutismDevDonutPainter extends CustomPainter {
  const _AutismDevDonutPainter({required this.percent});

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
      ..color = _AutismDevColors.orange
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
  bool shouldRepaint(covariant _AutismDevDonutPainter oldDelegate) {
    return oldDelegate.percent != percent;
  }
}

class _AutismDevFooter extends StatelessWidget {
  const _AutismDevFooter({
    required this.current,
    required this.total,
    required this.hasPrevious,
    required this.hasNext,
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
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _autismDevShadow(color: const Color(0x14B05F32), blur: 16),
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
                    color: _AutismDevColors.ink,
                    fontSize: 27,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.chevron_right_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.swipe_right_alt_rounded,
            enabled: true,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _AutismDevColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _AutismDevColors.orange,
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
    final Color textColor = filled ? Colors.white : _AutismDevColors.orangeDeep;
    final List<Widget> children = <Widget>[
      Icon(icon, size: 22, color: textColor),
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
            color: filled
                ? _AutismDevColors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  enabled ? _AutismDevColors.orange : const Color(0xFFE2D6CE),
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

class _AutismDevItemDetailCard extends StatelessWidget {
  const _AutismDevItemDetailCard({
    required this.item,
    required this.detail,
  });

  final AutismDevItemSummary item;
  final AutismDevAssessmentItem? detail;

  @override
  Widget build(BuildContext context) {
    final AutismDevAssessmentItem? currentDetail = detail;
    final String rangeText = _assessmentRangeText(item, currentDetail);
    final String ageText =
        _detailText(currentDetail?.ageSegment, item.ageSegment);
    final String methodText =
        _nonEmptyDetailText(currentDetail?.method, item.method);
    final String criteriaText =
        _nonEmptyDetailText(currentDetail?.passCriteria, item.passCriteria);
    final List<_AutismDevDetailField> cards = <_AutismDevDetailField>[
      _AutismDevDetailField(
        icon: Icons.inventory_2_outlined,
        label: '评估材料',
        value: _detailText(currentDetail?.materials, item.materials),
      ),
      _AutismDevDetailField(
        icon: Icons.assignment_outlined,
        label: '评估方法',
        value: methodText.isEmpty ? ' ' : methodText,
      ),
      _AutismDevDetailField(
        icon: Icons.article_outlined,
        label: '评分标准',
        value: criteriaText.isEmpty ? ' ' : _criteriaDisplayText(criteriaText),
      ),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            final bool compact = constraints.maxWidth < 540;
            if (compact) {
              return Column(
                children: <Widget>[
                  _AutismDevMetaCard(
                    icon: Icons.account_tree_outlined,
                    label: '评估范围',
                    value: rangeText,
                  ),
                  const SizedBox(height: 10),
                  _AutismDevMetaCard(
                    icon: Icons.calendar_month_outlined,
                    label: '参考年龄',
                    value: ageText,
                  ),
                ],
              );
            }
            return Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: _AutismDevMetaCard(
                    icon: Icons.account_tree_outlined,
                    label: '评估范围',
                    value: rangeText,
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 158,
                  child: _AutismDevMetaCard(
                    icon: Icons.calendar_month_outlined,
                    label: '参考年龄',
                    value: ageText,
                  ),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 10),
        for (final _AutismDevDetailField card in cards)
          _AutismDevDetailInfoCard(field: card),
      ],
    );
  }
}

class _AutismDevDetailField {
  const _AutismDevDetailField({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;
}

class _AutismDevMetaCard extends StatelessWidget {
  const _AutismDevMetaCard({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 70),
      padding: const EdgeInsets.fromLTRB(13, 12, 13, 12),
      decoration: _autismDevDetailCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
                child: Icon(
                  icon,
                  size: 12,
                  color: _AutismDevColors.orange,
                ),
              ),
              const SizedBox(width: 7),
              Text(
                label,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 13,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(
              color: _AutismDevColors.body,
              fontSize: 14,
              height: 1.28,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevDetailInfoCard extends StatelessWidget {
  const _AutismDevDetailInfoCard({required this.field});

  final _AutismDevDetailField field;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: _autismDevDetailCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
                child: Icon(
                  field.icon,
                  size: 12,
                  color: _AutismDevColors.orange,
                ),
              ),
              const SizedBox(width: 7),
              Text(
                field.label,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 11),
          Text(
            field.value,
            style: const TextStyle(
              color: _AutismDevColors.body,
              fontSize: 14,
              height: 1.55,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

BoxDecoration _autismDevDetailCardDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.94),
    borderRadius: BorderRadius.circular(10),
    border: Border.all(color: _AutismDevColors.line),
    boxShadow: _autismDevShadow(
      color: const Color(0x0FB05F32),
      blur: 12,
      offset: const Offset(0, 6),
    ),
  );
}

class _DomainChip extends StatelessWidget {
  const _DomainChip({required this.code, required this.name});

  final String code;
  final String name;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: const Color(0xFFFFC8AD)),
      ),
      child: Text(
        '${code.trim()} ${name.trim()}'.trim(),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(
          color: _AutismDevColors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _AutismDevLoadingBody extends StatelessWidget {
  const _AutismDevLoadingBody();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: SizedBox(
        width: 34,
        height: 34,
        child: CircularProgressIndicator(strokeWidth: 3),
      ),
    );
  }
}

class _AutismDevStateBody extends StatelessWidget {
  const _AutismDevStateBody({
    required this.title,
    required this.message,
    required this.actionText,
    required this.onAction,
  });

  final String title;
  final String message;
  final String actionText;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 420,
        padding: const EdgeInsets.all(28),
        decoration: _panelDecoration(),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Text(
              title,
              style: const TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 24,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _AutismDevColors.body,
                fontSize: 16,
                height: 1.45,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 22),
            FilledButton(
              onPressed: onAction,
              child: Text(actionText),
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
  static const Color lineSoft = Color(0xFFF6E7DC);
  static const Color softPanel = Color(0xFFFFFBF4);
  static const Color blue = Color(0xFF3F82D2);
  static const Color green = Color(0xFF6F9F70);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color red = Color(0xFFD94A42);
  static const Color teal = Color(0xFF00A7A7);
  static const Color purple = Color(0xFF7C5CFC);
  static const Color pink = Color(0xFFD0578B);
}

List<BoxShadow> _autismDevShadow({
  Color color = const Color(0x18000000),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

BoxDecoration _panelDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.92),
    border: Border.all(color: _AutismDevColors.line),
    borderRadius: BorderRadius.circular(8),
    boxShadow: _autismDevShadow(color: const Color(0x12B05F32), blur: 15),
  );
}

Color _domainColor(String code) {
  switch (code.toUpperCase()) {
    case 'GM':
      return _AutismDevColors.green;
    case 'FM':
      return _AutismDevColors.orange;
    case 'LC':
      return _AutismDevColors.purple;
    case 'COG':
      return _AutismDevColors.teal;
    case 'SOC':
      return _AutismDevColors.pink;
    case 'ADL':
      return const Color(0xFF6A7A20);
    case 'EB':
      return _AutismDevColors.red;
    case 'SP':
    default:
      return _AutismDevColors.blue;
  }
}

Color _scoreColor(String score) {
  switch (score.toUpperCase()) {
    case 'P':
    case 'A':
      return _AutismDevColors.green;
    case 'E':
    case 'M':
      return _AutismDevColors.orange;
    case 'F':
    case 'S':
      return _AutismDevColors.red;
    case 'X':
    default:
      return _AutismDevColors.muted;
  }
}

String _optionTitle(AutismDevScoreOption option) {
  final String label = option.label.trim();
  if (label.isEmpty) {
    return option.value;
  }
  final String prefix = option.value.trim();
  if (prefix.isNotEmpty && label.startsWith(prefix)) {
    return label.substring(prefix.length).trim();
  }
  return label;
}

String _displayItemTitle(AutismDevItemSummary item) {
  return item.itemTitle.trim().isNotEmpty
      ? item.itemTitle.trim()
      : item.testItem.trim();
}

String _assessmentRangeText(
  AutismDevItemSummary item,
  AutismDevAssessmentItem? detail,
) {
  final String range = (detail?.assessmentRange ?? item.assessmentRange).trim();
  return range.isEmpty ? '-' : range;
}

String _assessmentRangeBucket(AutismDevItemSummary item) {
  final List<String> parts = item.assessmentRange
      .split('/')
      .map((String part) => part.trim())
      .where((String part) => part.isNotEmpty)
      .toList(growable: false);
  if (parts.length >= 2) {
    return parts[1];
  }
  if (parts.isNotEmpty) {
    return parts.first;
  }
  return '未分类';
}

List<_AutismDevRangeOption> _rangeOptionsForGroup(
  AutismDevDomainGroup? group,
  Map<int, String> itemScores,
) {
  if (group == null) {
    return const <_AutismDevRangeOption>[];
  }
  final Map<String, _AutismDevRangeOption> optionByLabel =
      <String, _AutismDevRangeOption>{};
  for (final AutismDevItemSummary item in group.items) {
    final String label = _assessmentRangeBucket(item);
    final _AutismDevRangeOption option =
        optionByLabel.putIfAbsent(label, () => _AutismDevRangeOption(label));
    option.total += 1;
    if (itemScores.containsKey(item.itemNo)) {
      option.done += 1;
    }
  }
  return optionByLabel.values.toList(growable: false);
}

String _detailText(String? preferred, [String fallback = '']) {
  final String value = (preferred ?? '').trim();
  if (value.isNotEmpty) {
    return value;
  }
  final String fallbackValue = fallback.trim();
  return fallbackValue.isNotEmpty ? fallbackValue : '-';
}

String _nonEmptyDetailText(String? preferred, [String fallback = '']) {
  final String value = (preferred ?? '').trim();
  if (value.isNotEmpty) {
    return value;
  }
  return fallback.trim();
}

List<String> _criteriaDisplayLines(String? value) {
  final String text = _detailText(value);
  if (text == '-') {
    return const <String>[];
  }
  return text
      .split(RegExp(r'[\r\n]+'))
      .map((String line) => line.trim())
      .where((String line) => line.isNotEmpty)
      .toList();
}

String _criteriaDisplayText(String? value) {
  final List<String> lines = _criteriaDisplayLines(value);
  return lines.isEmpty ? '-' : lines.join('；');
}

String _dateOnlyText(String value) {
  final String text = value.trim();
  if (text.length >= 10) {
    return text.substring(0, 10);
  }
  return text;
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-'
      '${now.month.toString().padLeft(2, '0')}-'
      '${now.day.toString().padLeft(2, '0')}';
}

String _formatClock(DateTime value) {
  return '${value.hour.toString().padLeft(2, '0')}:'
      '${value.minute.toString().padLeft(2, '0')}';
}
