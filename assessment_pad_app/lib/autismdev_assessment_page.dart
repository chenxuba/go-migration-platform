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
  AutismDevDraftProgress _draftProgress = AutismDevDraftProgress.empty;
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
  bool _loading = true;
  bool _saving = false;
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
    _draftProgress = detail.progress;
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

  int get _totalCount {
    if (_template.itemCount > 0) {
      return _template.itemCount;
    }
    return _template.domainGroups.fold<int>(
      0,
      (int total, AutismDevDomainGroup group) => total + group.items.length,
    );
  }

  double get _progressPercent {
    if (_totalCount <= 0) {
      return 0;
    }
    return (_answeredCount / _totalCount).clamp(0, 1).toDouble();
  }

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

  void _selectDomain(String domainCode) {
    AutismDevDomainGroup? group;
    for (final AutismDevDomainGroup item in _template.domainGroups) {
      if (item.domainCode == domainCode) {
        group = item;
        break;
      }
    }
    if (group == null) {
      return;
    }
    final AutismDevDomainGroup selectedGroup = group;
    final int nextItemNo = selectedGroup.items
        .map((AutismDevItemSummary item) => item.itemNo)
        .firstWhere(
          (int itemNo) => !_itemScores.containsKey(itemNo),
          orElse: () => selectedGroup.items.isNotEmpty
              ? selectedGroup.items.first.itemNo
              : 0,
        );
    setState(() {
      _selectedDomainCode = domainCode;
      _selectedItemNo = nextItemNo;
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  void _selectItem(AutismDevItemSummary item) {
    setState(() {
      _selectedItemNo = item.itemNo;
      _selectedDomainCode = item.domainCode;
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
        _draftProgress = detail.progress;
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

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
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
        _draftProgress = detail.progress;
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
      child: Row(
        children: <Widget>[
          _AutismDevRail(onBack: widget.onBack),
          Expanded(
            child: Column(
              children: <Widget>[
                _AutismDevTopBar(
                  title: widget.args.scaleName.trim().isEmpty
                      ? '孤独症儿童发展评估'
                      : widget.args.scaleName.trim(),
                  subtitle:
                      '现场测评工作台 / ${_assessmentDate.isEmpty ? '未设置日期' : _assessmentDate}',
                  studentName:
                      _studentName.trim().isEmpty ? '未选择儿童' : _studentName,
                  studentAge: _studentAge.trim().isEmpty ? '年龄未知' : _studentAge,
                  draftId: _draftId,
                ),
                Expanded(child: _buildBody()),
              ],
            ),
          ),
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
    return Padding(
      padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 298,
            child: _AutismDevDomainPanel(
              groups: _template.domainGroups,
              selectedDomainCode: _selectedDomainCode,
              itemScores: _itemScores,
              onSelectDomain: _selectDomain,
              answeredCount: _answeredCount,
              totalCount: _totalCount,
              progressPercent: _progressPercent,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: _AutismDevWorkspacePanel(
              group: group,
              item: summary,
              detail: _selectedDetail,
              selectedScore: _itemScores[_selectedItemNo],
              itemScores: _itemScores,
              onSelectItem: _selectItem,
              onPrevious: _goPreviousItem,
              onNext: _goNextItem,
            ),
          ),
          const SizedBox(width: 10),
          SizedBox(
            width: 296,
            child: _AutismDevScorePanel(
              item: summary,
              options: _currentScoreOptions,
              selectedScore: _itemScores[_selectedItemNo],
              remarkController: _remarkController,
              saving: _saving,
              submitting: _submitting,
              autoSaveText: _autoSaveText,
              progress: _draftProgress,
              answeredCount: _answeredCount,
              totalCount: _totalCount,
              onScore: (String score) => _selectScore(score),
              onSave: () => _saveCurrentItem(moveNext: false),
              onSaveNext: () => _saveCurrentItem(moveNext: true),
              onSubmit: _submitDraft,
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevRail extends StatelessWidget {
  const _AutismDevRail({required this.onBack});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 84,
      decoration: const BoxDecoration(
        color: _AutismDevColors.rail,
        border: Border(
          right: BorderSide(color: _AutismDevColors.line),
        ),
      ),
      child: Column(
        children: <Widget>[
          const SizedBox(height: 24),
          Material(
            color: _AutismDevColors.railActive,
            borderRadius: BorderRadius.circular(8),
            child: InkWell(
              borderRadius: BorderRadius.circular(8),
              onTap: onBack,
              child: const SizedBox(
                width: 44,
                height: 44,
                child: Icon(Icons.arrow_back, color: Colors.white, size: 22),
              ),
            ),
          ),
          const SizedBox(height: 52),
          const _RailItem(
            icon: Icons.fact_check_outlined,
            label: '任务',
            active: true,
          ),
          const _RailItem(icon: Icons.history, label: '记录'),
          const _RailItem(icon: Icons.analytics_outlined, label: '报告'),
          const _RailItem(icon: Icons.folder_shared_outlined, label: '档案'),
        ],
      ),
    );
  }
}

class _RailItem extends StatelessWidget {
  const _RailItem({
    required this.icon,
    required this.label,
    this.active = false,
  });

  final IconData icon;
  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    final Color contentColor =
        active ? _AutismDevColors.orangeDeep : _AutismDevColors.body;
    return Container(
      width: 58,
      margin: const EdgeInsets.only(bottom: 22),
      padding: const EdgeInsets.symmetric(vertical: 10),
      decoration: BoxDecoration(
        color: active ? _AutismDevColors.railItem : Colors.transparent,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Icon(icon, size: 23, color: contentColor),
          const SizedBox(height: 7),
          Text(
            label,
            style: TextStyle(
              color: contentColor,
              fontSize: 13,
              fontWeight: active ? FontWeight.w900 : FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevTopBar extends StatelessWidget {
  const _AutismDevTopBar({
    required this.title,
    required this.subtitle,
    required this.studentName,
    required this.studentAge,
    required this.draftId,
  });

  final String title;
  final String subtitle;
  final String studentName;
  final String studentAge;
  final int draftId;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 74,
      padding: const EdgeInsets.symmetric(horizontal: 28),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(
          bottom: Radius.circular(12),
        ),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x16B05F32),
            blurRadius: 16,
            offset: Offset(0, 9),
          ),
        ],
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 25,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  subtitle,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _AutismDevColors.muted,
                    fontSize: 14,
                    height: 1,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
          Container(
            width: 306,
            height: 44,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            decoration: BoxDecoration(
              color: _AutismDevColors.softPanel,
              border: Border.all(color: _AutismDevColors.line),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: <Widget>[
                Expanded(
                  child: Text(
                    '$studentName  $studentAge',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _AutismDevColors.ink,
                      fontSize: 15,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                Text(
                  draftId > 0 ? '草稿 #$draftId' : '新测评',
                  style: const TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 14,
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

class _AutismDevDomainPanel extends StatelessWidget {
  const _AutismDevDomainPanel({
    required this.groups,
    required this.selectedDomainCode,
    required this.itemScores,
    required this.onSelectDomain,
    required this.answeredCount,
    required this.totalCount,
    required this.progressPercent,
  });

  final List<AutismDevDomainGroup> groups;
  final String selectedDomainCode;
  final Map<int, String> itemScores;
  final ValueChanged<String> onSelectDomain;
  final int answeredCount;
  final int totalCount;
  final double progressPercent;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 18, 18, 14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                const Text(
                  '领域任务',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 22,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 14),
                ClipRRect(
                  borderRadius: BorderRadius.circular(5),
                  child: LinearProgressIndicator(
                    value: progressPercent,
                    minHeight: 10,
                    backgroundColor: const Color(0xFFE7EBF3),
                    color: _AutismDevColors.blue,
                  ),
                ),
                const SizedBox(height: 10),
                Text(
                  '已完成 $answeredCount/$totalCount',
                  style: const TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1, color: _AutismDevColors.line),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.all(14),
              itemBuilder: (BuildContext context, int index) {
                final AutismDevDomainGroup group = groups[index];
                final int done = group.items
                    .where((AutismDevItemSummary item) =>
                        itemScores.containsKey(item.itemNo))
                    .length;
                return _AutismDevDomainTile(
                  group: group,
                  done: done,
                  selected: group.domainCode == selectedDomainCode,
                  onTap: () => onSelectDomain(group.domainCode),
                );
              },
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemCount: groups.length,
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
    required this.done,
    required this.selected,
    required this.onTap,
  });

  final AutismDevDomainGroup group;
  final int done;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = _domainColor(group.domainCode);
    final double progress = group.itemCount <= 0
        ? 0
        : (done / group.itemCount).clamp(0, 1).toDouble();
    return Material(
      color: selected ? color.withOpacity(.08) : Colors.white,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            border: Border.all(
              color: selected ? color.withOpacity(.7) : _AutismDevColors.line,
              width: selected ? 1.5 : 1,
            ),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Container(
                    width: 10,
                    height: 10,
                    decoration: BoxDecoration(
                      color: color,
                      borderRadius: BorderRadius.circular(5),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      group.domainName,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected ? color : _AutismDevColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  Text(
                    '$done/${group.itemCount}',
                    style: const TextStyle(
                      color: _AutismDevColors.body,
                      fontSize: 13,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: LinearProgressIndicator(
                  value: progress,
                  minHeight: 7,
                  color: color,
                  backgroundColor: const Color(0xFFE9EDF5),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                group.scoreType.toUpperCase() == 'AMS'
                    ? 'A/M/S 临床判断评分'
                    : 'P/E/F/X 发展项目评分',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _AutismDevColors.muted,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
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
    required this.itemScores,
    required this.onSelectItem,
    required this.onPrevious,
    required this.onNext,
  });

  final AutismDevDomainGroup group;
  final AutismDevItemSummary item;
  final AutismDevAssessmentItem? detail;
  final String? selectedScore;
  final Map<int, String> itemScores;
  final ValueChanged<AutismDevItemSummary> onSelectItem;
  final VoidCallback onPrevious;
  final VoidCallback onNext;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.fromLTRB(28, 24, 28, 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Expanded(
                        child: Text(
                          '#${item.itemNo} ${_displayItemTitle(item)}',
                          style: const TextStyle(
                            color: _AutismDevColors.ink,
                            fontSize: 28,
                            height: 1.15,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                      if (selectedScore != null)
                        _ScoreBadge(score: selectedScore!),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Wrap(
                    spacing: 10,
                    runSpacing: 8,
                    children: <Widget>[
                      _Chip(
                          label: group.domainName,
                          color: _domainColor(group.domainCode)),
                      _Chip(
                        label: item.scoreType.toUpperCase() == 'AMS'
                            ? '情绪行为'
                            : '发展项目',
                        color: item.scoreType.toUpperCase() == 'AMS'
                            ? _AutismDevColors.red
                            : _AutismDevColors.blue,
                      ),
                      if (item.ageSegment.trim().isNotEmpty)
                        _Chip(
                            label: item.ageSegment,
                            color: _AutismDevColors.body),
                    ],
                  ),
                  const SizedBox(height: 30),
                  _SectionBlock(
                    title:
                        item.scoreType.toUpperCase() == 'AMS' ? '观察问题' : '评估项目',
                    body: _detailText(detail?.testItem, item.testItem),
                    boxed: true,
                  ),
                  const SizedBox(height: 24),
                  _SectionBlock(
                    title: '评估方法',
                    body: _detailText(
                      detail?.method,
                      item.scoreType.toUpperCase() == 'AMS'
                          ? '结合现场观察、访谈和活动中的自然反应进行临床判断。'
                          : '按题目要求进行结构化或自然情境观察，并记录儿童的独立完成程度。',
                    ),
                  ),
                  const SizedBox(height: 24),
                  _SectionBlock(
                    title: '评分标准',
                    body: _detailText(
                      detail?.passCriteria,
                      item.scoreType.toUpperCase() == 'AMS'
                          ? 'A=没有异常；M=轻度异常；S=重度异常。'
                          : 'P=通过，记1分；E=中间反应，不计分，可作为训练目标；F=不通过，记0分；X=不适用，不计分。',
                    ),
                  ),
                ],
              ),
            ),
          ),
          Container(
            height: 154,
            padding: const EdgeInsets.fromLTRB(20, 14, 20, 16),
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: _AutismDevColors.line)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    Expanded(
                      child: Text(
                        '${group.domainName}项目',
                        style: const TextStyle(
                          color: _AutismDevColors.ink,
                          fontSize: 15,
                          height: 1,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    IconButton(
                      tooltip: '上一题',
                      onPressed: onPrevious,
                      icon: const Icon(Icons.chevron_left),
                      color: _AutismDevColors.body,
                    ),
                    IconButton(
                      tooltip: '下一题',
                      onPressed: onNext,
                      icon: const Icon(Icons.chevron_right),
                      color: _AutismDevColors.body,
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: ListView.separated(
                    scrollDirection: Axis.horizontal,
                    itemBuilder: (BuildContext context, int index) {
                      final AutismDevItemSummary row = group.items[index];
                      return _BottomItemCard(
                        item: row,
                        selected: row.itemNo == item.itemNo,
                        score: itemScores[row.itemNo],
                        onTap: () => onSelectItem(row),
                      );
                    },
                    separatorBuilder: (_, __) => const SizedBox(width: 10),
                    itemCount: group.items.length,
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

class _AutismDevScorePanel extends StatelessWidget {
  const _AutismDevScorePanel({
    required this.item,
    required this.options,
    required this.selectedScore,
    required this.remarkController,
    required this.saving,
    required this.submitting,
    required this.autoSaveText,
    required this.progress,
    required this.answeredCount,
    required this.totalCount,
    required this.onScore,
    required this.onSave,
    required this.onSaveNext,
    required this.onSubmit,
  });

  final AutismDevItemSummary item;
  final List<AutismDevScoreOption> options;
  final String? selectedScore;
  final TextEditingController remarkController;
  final bool saving;
  final bool submitting;
  final String autoSaveText;
  final AutismDevDraftProgress progress;
  final int answeredCount;
  final int totalCount;
  final ValueChanged<String> onScore;
  final VoidCallback onSave;
  final VoidCallback onSaveNext;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final int missing = math.max(totalCount - answeredCount, 0);
    return Container(
      decoration: _panelDecoration(),
      padding: const EdgeInsets.fromLTRB(20, 22, 20, 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '本项评分',
            style: TextStyle(
              color: _AutismDevColors.ink,
              fontSize: 22,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 18),
          for (final AutismDevScoreOption option in options)
            Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: _ScoreOptionButton(
                option: option,
                selected: selectedScore == option.value,
                onTap: () => onScore(option.value),
              ),
            ),
          const SizedBox(height: 8),
          const Text(
            '观察记录',
            style: TextStyle(
              color: _AutismDevColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          SizedBox(
            height: 112,
            child: TextField(
              controller: remarkController,
              maxLines: null,
              expands: true,
              textAlignVertical: TextAlignVertical.top,
              style: const TextStyle(
                color: _AutismDevColors.body,
                fontSize: 15,
                height: 1.35,
                fontWeight: FontWeight.w600,
              ),
              decoration: InputDecoration(
                hintText: item.scoreType.toUpperCase() == 'AMS'
                    ? '记录触发情境、反应强度和持续时间'
                    : '记录提示层级、辅助方式或训练目标候选',
                hintStyle: const TextStyle(color: _AutismDevColors.muted),
                filled: true,
                fillColor: _AutismDevColors.softPanel,
                contentPadding: const EdgeInsets.all(14),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: _AutismDevColors.line),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: _AutismDevColors.blue),
                ),
              ),
            ),
          ),
          const SizedBox(height: 18),
          _ProgressSummary(
            answeredCount: answeredCount,
            totalCount: totalCount,
            missing: missing,
            backendPercent: progress.completionPercent,
          ),
          const Spacer(),
          Text(
            autoSaveText,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _AutismDevColors.muted,
              fontSize: 13,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 10),
          SizedBox(
            width: double.infinity,
            height: 48,
            child: FilledButton.icon(
              onPressed: saving ? null : onSaveNext,
              icon: saving
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.save_outlined, size: 19),
              label: Text(saving ? '保存中' : '保存并下一题'),
              style: FilledButton.styleFrom(
                backgroundColor: _AutismDevColors.ink,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
            ),
          ),
          const SizedBox(height: 10),
          Row(
            children: <Widget>[
              Expanded(
                child: OutlinedButton(
                  onPressed: saving ? null : onSave,
                  style: OutlinedButton.styleFrom(
                    foregroundColor: _AutismDevColors.body,
                    side: const BorderSide(color: _AutismDevColors.line),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: const Text('保存'),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: OutlinedButton(
                  onPressed: submitting ? null : onSubmit,
                  style: OutlinedButton.styleFrom(
                    foregroundColor: _AutismDevColors.blue,
                    side: const BorderSide(color: _AutismDevColors.blue),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: Text(submitting ? '提交中' : '提交'),
                ),
              ),
            ],
          ),
        ],
      ),
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
      color: selected ? color : Colors.white,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          height: 64,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            border: Border.all(color: color, width: selected ? 0 : 1.4),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            children: <Widget>[
              Text(
                option.value,
                style: TextStyle(
                  color: selected ? Colors.white : color,
                  fontSize: 26,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      _optionTitle(option),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected ? Colors.white : _AutismDevColors.ink,
                        fontSize: 17,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 5),
                    Text(
                      option.description,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected
                            ? Colors.white.withOpacity(.82)
                            : _AutismDevColors.muted,
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
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

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.answeredCount,
    required this.totalCount,
    required this.missing,
    required this.backendPercent,
  });

  final int answeredCount;
  final int totalCount;
  final int missing;
  final double backendPercent;

  @override
  Widget build(BuildContext context) {
    final double percent = totalCount <= 0 ? 0 : answeredCount / totalCount;
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _AutismDevColors.softPanel,
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              const Expanded(
                child: Text(
                  '全量进度',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Text(
                '${(percent * 100).round()}%',
                style: const TextStyle(
                  color: _AutismDevColors.blue,
                  fontSize: 15,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(5),
            child: LinearProgressIndicator(
              value: percent.clamp(0, 1).toDouble(),
              minHeight: 9,
              color: _AutismDevColors.blue,
              backgroundColor: const Color(0xFFE1E6EF),
            ),
          ),
          const SizedBox(height: 10),
          Row(
            children: <Widget>[
              Expanded(
                  child: _SmallMetric(label: '完成', value: '$answeredCount')),
              Expanded(child: _SmallMetric(label: '未评', value: '$missing')),
              Expanded(
                child: _SmallMetric(
                  label: '服务端',
                  value: '${backendPercent.round()}%',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _SmallMetric extends StatelessWidget {
  const _SmallMetric({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          value,
          style: const TextStyle(
            color: _AutismDevColors.ink,
            fontSize: 17,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 3),
        Text(
          label,
          style: const TextStyle(
            color: _AutismDevColors.muted,
            fontSize: 11,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _SectionBlock extends StatelessWidget {
  const _SectionBlock({
    required this.title,
    required this.body,
    this.boxed = false,
  });

  final String title;
  final String body;
  final bool boxed;

  @override
  Widget build(BuildContext context) {
    final TextStyle bodyStyle = const TextStyle(
      color: _AutismDevColors.body,
      fontSize: 18,
      height: 1.58,
      fontWeight: FontWeight.w600,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(
            color: _AutismDevColors.ink,
            fontSize: 20,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 12),
        if (boxed)
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: _AutismDevColors.softPanel,
              border: Border.all(color: _AutismDevColors.line),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(body, style: bodyStyle),
          )
        else
          Text(body, style: bodyStyle),
      ],
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: color.withOpacity(.08),
        border: Border.all(color: color.withOpacity(.25)),
        borderRadius: BorderRadius.circular(15),
      ),
      child: Center(
        child: Text(
          label,
          style: TextStyle(
            color: color,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _BottomItemCard extends StatelessWidget {
  const _BottomItemCard({
    required this.item,
    required this.selected,
    required this.score,
    required this.onTap,
  });

  final AutismDevItemSummary item;
  final bool selected;
  final String? score;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? _AutismDevColors.blue.withOpacity(.08) : Colors.white,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          width: 178,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            border: Border.all(
              color: selected ? _AutismDevColors.blue : _AutismDevColors.line,
            ),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Text(
                    '${item.itemNo}',
                    style: const TextStyle(
                      color: _AutismDevColors.muted,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const Spacer(),
                  if (score != null) _ScoreBadge(score: score!, compact: true),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                _displayItemTitle(item),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 14,
                  height: 1.2,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ScoreBadge extends StatelessWidget {
  const _ScoreBadge({
    required this.score,
    this.compact = false,
  });

  final String score;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(score);
    return Container(
      width: compact ? 26 : 42,
      height: compact ? 24 : 34,
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(compact ? 12 : 17),
      ),
      child: Center(
        child: Text(
          score,
          style: TextStyle(
            color: Colors.white,
            fontSize: compact ? 13 : 17,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
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
  static const Color rail = Color(0xFFFFFBF4);
  static const Color railItem = Color(0xFFFFE9DB);
  static const Color railActive = Color(0xFFE96F43);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
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

BoxDecoration _panelDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.92),
    border: Border.all(color: _AutismDevColors.line),
    borderRadius: BorderRadius.circular(8),
    boxShadow: <BoxShadow>[
      BoxShadow(
        color: const Color(0x16B05F32),
        blurRadius: 16,
        offset: const Offset(0, 9),
      ),
    ],
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

String _detailText(String? preferred, String fallback) {
  final String value = (preferred ?? '').trim();
  return value.isNotEmpty ? value : fallback.trim();
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
